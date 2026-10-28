package issues

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	fieldKeyRe  = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	errorCodeRe = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)
	stepRe      = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
)

func accepted(ev Event) bool {
	switch ev.Type {
	case TypeClick, TypeClickOutcome:
		return ev.ElementSignature != ""
	case TypeFormOutcome:
		return ev.FormKey != ""
	case TypeNavigation, TypeError, TypeBusiness, TypeFlowStep:
		return true
	default:
		return false
	}
}

type clickRec struct {
	EventID      string
	AnonymousID  string
	SessionID    string
	PageSeq      int
	MonoMS       int64
	Timestamp    int64
	Route        string
	Signature    string
	FlowName     string
	Dead         bool
	AppVersion   string
	AppBuild     string
	AppNamespace string
	ReleaseID    string
}

type outcomeRec struct {
	Result string
	At     int64
}

type formGroup struct {
	Session string
	Page    int
	Form    string
}

type formRec struct {
	EventID      string
	Timestamp    int64
	MonoMS       int64
	AnonymousID  string
	SessionID    string
	PageSeq      int
	Route        string
	FlowName     string
	FormKey      string
	Result       string
	FieldKeys    []string
	FieldSet     string
	ErrorCode    string
	Truncated    bool
	AppVersion   string
	AppBuild     string
	AppNamespace string
	ReleaseID    string
}

type occurrence struct {
	ID               string
	Type             string
	Route            string
	Signature        string
	FlowName         string
	AnonymousID      string
	SessionID        string
	FirstMS          int64
	LastMS           int64
	Count            int
	DeadCorroborated bool
	Truncated        bool
	FormKey          string
	FieldKeys        []string
	ErrorCode        string
	Kind             string
	Method           string
	Status           int
	Path             string
	Name             string
	File             string
	Function         string
	Step             string
	EndReason        string
	PageSeq          int
	FirstAppVersion  string
	AppVersion       string
	LatestAppVersion string
	AppBuild         string
	AppNamespace     string
	ReleaseID        string
}

func (s *Store) addClick(ev Event) {
	c := clickRec{
		EventID:      ev.EventID,
		AnonymousID:  ev.AnonymousID,
		SessionID:    ev.SessionID,
		PageSeq:      ev.PageSeq,
		MonoMS:       ev.MonoMS,
		Timestamp:    ev.Timestamp,
		Route:        ev.Route,
		Signature:    ev.ElementSignature,
		FlowName:     ev.FlowName,
		AppVersion:   ev.AppVersion,
		AppBuild:     ev.AppBuild,
		AppNamespace: ev.AppNamespace,
		ReleaseID:    ev.ReleaseID,
	}
	if o, ok := s.outcomes[ev.EventID]; ok && o.Result == "dead" {
		c.Dead = true
	}
	s.clicks[ev.SessionID] = append(s.clicks[ev.SessionID], c)
	s.clickLoc[ev.EventID] = ev.SessionID
}

// addOutcome records the first outcome for a click_id. A dead result creates
// dead_click even when the original click never arrived. The returned session
// is set when a stored click was marked and bursts should be rebuilt.
func (s *Store) addOutcome(ev Event) (string, bool) {
	if ev.ClickID == "" {
		return "", false
	}
	if _, exists := s.outcomes[ev.ClickID]; exists {
		return "", false
	}
	s.outcomes[ev.ClickID] = outcomeRec{Result: ev.Result, At: ev.Timestamp}
	touched := ""
	if sess, ok := s.clickLoc[ev.ClickID]; ok && ev.Result == "dead" {
		s.markClickDead(sess, ev.ClickID)
		touched = sess
	}
	if ev.Result == "dead" {
		occ := baseOcc(ev)
		occ.ID = "dead:" + ev.ClickID
		occ.Type = IssueDeadClick
		occ.Signature = ev.ElementSignature
		s.occs[occ.ID] = occ
	}
	return touched, touched != ""
}

func (s *Store) markClickDead(sessionID, clickID string) {
	clicks := s.clicks[sessionID]
	for i := range clicks {
		if clicks[i].EventID == clickID {
			clicks[i].Dead = true
		}
	}
	s.clicks[sessionID] = clicks
}

func (s *Store) addForm(ev Event) (formGroup, bool) {
	switch ev.Result {
	case "success", "validation_error", "submit_error":
	default:
		return formGroup{}, false
	}
	keys, set := normalizeFieldKeys(ev.FieldKeys)
	g := formGroup{Session: ev.SessionID, Page: ev.PageSeq, Form: ev.FormKey}
	s.forms[g] = append(s.forms[g], formRec{
		EventID:      ev.EventID,
		Timestamp:    ev.Timestamp,
		MonoMS:       ev.MonoMS,
		AnonymousID:  ev.AnonymousID,
		SessionID:    ev.SessionID,
		PageSeq:      ev.PageSeq,
		Route:        ev.Route,
		FlowName:     ev.FlowName,
		FormKey:      ev.FormKey,
		Result:       ev.Result,
		FieldKeys:    keys,
		FieldSet:     set,
		ErrorCode:    cleanErrorCode(ev.ErrorCode),
		Truncated:    ev.Truncated,
		AppVersion:   ev.AppVersion,
		AppBuild:     ev.AppBuild,
		AppNamespace: ev.AppNamespace,
		ReleaseID:    ev.ReleaseID,
	})
	return g, true
}

func (s *Store) addError(ev Event) {
	switch ev.Kind {
	case KindHTTP:
		if ev.HTTPStatus != 0 && ev.HTTPStatus < 400 {
			return
		}
		method := strings.ToUpper(ev.Method)
		occ := baseOcc(ev)
		occ.ID = "fail:" + ev.EventID
		occ.Type = IssueFailedAction
		occ.Signature = fmt.Sprintf("http:%s:%d:%s", method, ev.HTTPStatus, ev.Path)
		occ.Kind = KindHTTP
		occ.Method = method
		occ.Status = ev.HTTPStatus
		occ.Path = ev.Path
		s.occs[occ.ID] = occ
	case KindJS:
		occ := baseOcc(ev)
		occ.ID = "fail:" + ev.EventID
		occ.Type = IssueFailedAction
		occ.Signature = fmt.Sprintf("js:%s:%s:%s:%s", ev.Name, ev.MessageSlug, ev.File, ev.Function)
		occ.Kind = KindJS
		occ.Name = ev.Name
		occ.File = ev.File
		occ.Function = ev.Function
		s.occs[occ.ID] = occ
	}
}

func (s *Store) rebuildBursts(sessionID string) {
	for id, occ := range s.occs {
		if occ.Type == IssueRepeatedClick && occ.SessionID == sessionID {
			delete(s.occs, id)
		}
	}
	clicks := append([]clickRec(nil), s.clicks[sessionID]...)
	sort.Slice(clicks, func(i, j int) bool {
		if clicks[i].PageSeq != clicks[j].PageSeq {
			return clicks[i].PageSeq < clicks[j].PageSeq
		}
		if clicks[i].MonoMS != clicks[j].MonoMS {
			return clicks[i].MonoMS < clicks[j].MonoMS
		}
		if clicks[i].Timestamp != clicks[j].Timestamp {
			return clicks[i].Timestamp < clicks[j].Timestamp
		}
		return clicks[i].EventID < clicks[j].EventID
	})
	for i := 0; i < len(clicks); {
		run := []clickRec{clicks[i]}
		j := i + 1
		for j < len(clicks) {
			prev := run[len(run)-1]
			cur := clicks[j]
			if cur.PageSeq != prev.PageSeq || cur.Signature != prev.Signature || cur.Route != prev.Route {
				break
			}
			if cur.MonoMS-prev.MonoMS >= RepeatedClickGapMS {
				break
			}
			run = append(run, cur)
			j++
		}
		if len(run) >= RepeatedClickMin {
			s.occs[burstID(run[0])] = burstOcc(run)
		}
		i = j
	}
}

func burstID(first clickRec) string {
	return "burst:" + first.SessionID + strconv.Itoa(first.PageSeq) + first.Signature + first.EventID
}

func burstOcc(run []clickRec) occurrence {
	first := run[0]
	last := run[len(run)-1]
	var info relInfo
	dead := false
	for _, c := range run {
		info.add(c.Timestamp, c.AppVersion, c.AppBuild, c.AppNamespace, c.ReleaseID)
		if c.Dead {
			dead = true
		}
	}
	return occurrence{
		ID:               burstID(first),
		Type:             IssueRepeatedClick,
		Route:            first.Route,
		Signature:        first.Signature,
		FlowName:         first.FlowName,
		AnonymousID:      first.AnonymousID,
		SessionID:        first.SessionID,
		FirstMS:          first.Timestamp,
		LastMS:           last.Timestamp,
		Count:            len(run),
		DeadCorroborated: dead,
		PageSeq:          first.PageSeq,
		FirstAppVersion:  first.AppVersion,
		AppVersion:       last.AppVersion,
		LatestAppVersion: info.latestVersion,
		AppBuild:         info.build,
		AppNamespace:     info.namespace,
		ReleaseID:        info.releaseID,
	}
}

func (s *Store) rebuildForms(g formGroup) {
	for id, occ := range s.occs {
		if occ.SessionID == g.Session && occ.PageSeq == g.Page && occ.FormKey == g.Form &&
			(occ.Type == IssueValidationLoop || (occ.Type == IssueFailedAction && occ.Kind == KindSubmit)) {
			delete(s.occs, id)
		}
	}
	for _, occ := range evalForms(s.forms[g]) {
		s.occs[occ.ID] = occ
	}
}

func evalForms(events []formRec) []occurrence {
	evs := append([]formRec(nil), events...)
	sort.SliceStable(evs, func(i, j int) bool {
		if evs[i].MonoMS != evs[j].MonoMS {
			return evs[i].MonoMS < evs[j].MonoMS
		}
		if evs[i].Timestamp != evs[j].Timestamp {
			return evs[i].Timestamp < evs[j].Timestamp
		}
		return evs[i].EventID < evs[j].EventID
	})

	type run struct {
		recs []formRec
	}
	current := map[string]*run{}
	var done []run
	finish := func(r *run) {
		if r == nil || len(r.recs) == 0 {
			return
		}
		cp := make([]formRec, len(r.recs))
		copy(cp, r.recs)
		done = append(done, run{recs: cp})
	}
	finishAll := func() {
		for k, r := range current {
			finish(r)
			delete(current, k)
		}
	}

	var out []occurrence
	for _, ev := range evs {
		switch ev.Result {
		case "success", "submit_error":
			finishAll()
			if ev.Result == "submit_error" {
				out = append(out, submitOcc(ev))
			}
		case "validation_error":
			r := current[ev.FieldSet]
			if r != nil && len(r.recs) > 0 {
				prev := r.recs[len(r.recs)-1]
				if ev.MonoMS-prev.MonoMS >= ValidationLoopGapMS {
					finish(r)
					delete(current, ev.FieldSet)
					r = nil
				}
			}
			if r == nil {
				r = &run{}
				current[ev.FieldSet] = r
			}
			r.recs = append(r.recs, ev)
		}
	}
	finishAll()
	for _, r := range done {
		if len(r.recs) >= ValidationLoopMin {
			out = append(out, loopOcc(r.recs))
		}
	}
	return out
}

func loopOcc(recs []formRec) occurrence {
	first := recs[0]
	last := recs[len(recs)-1]
	sigSet := first.FieldSet
	if sigSet == "" {
		sigSet = "-"
	}
	var info relInfo
	truncated := false
	for _, rec := range recs {
		info.add(rec.Timestamp, rec.AppVersion, rec.AppBuild, rec.AppNamespace, rec.ReleaseID)
		if rec.Truncated {
			truncated = true
		}
	}
	occ := occurrence{
		ID:               "vloop:" + first.SessionID + strconv.Itoa(first.PageSeq) + first.FormKey + first.FieldSet + first.EventID,
		Type:             IssueValidationLoop,
		Route:            last.Route,
		Signature:        "form:" + first.FormKey + ":fields:" + sigSet,
		FlowName:         last.FlowName,
		AnonymousID:      first.AnonymousID,
		SessionID:        first.SessionID,
		FirstMS:          first.Timestamp,
		LastMS:           last.Timestamp,
		Count:            len(recs),
		Truncated:        truncated,
		FormKey:          first.FormKey,
		FieldKeys:        append([]string(nil), first.FieldKeys...),
		ErrorCode:        last.ErrorCode,
		PageSeq:          first.PageSeq,
		FirstAppVersion:  first.AppVersion,
		AppVersion:       last.AppVersion,
		LatestAppVersion: info.latestVersion,
		AppBuild:         info.build,
		AppNamespace:     info.namespace,
		ReleaseID:        info.releaseID,
	}
	return occ
}

func submitOcc(ev formRec) occurrence {
	occ := occurrence{
		ID:               "fail:" + ev.EventID,
		Type:             IssueFailedAction,
		Route:            ev.Route,
		Signature:        "form:" + ev.FormKey + ":submit_error",
		FlowName:         ev.FlowName,
		AnonymousID:      ev.AnonymousID,
		SessionID:        ev.SessionID,
		FirstMS:          ev.Timestamp,
		LastMS:           ev.Timestamp,
		Count:            1,
		Truncated:        ev.Truncated,
		FormKey:          ev.FormKey,
		ErrorCode:        ev.ErrorCode,
		Kind:             KindSubmit,
		PageSeq:          ev.PageSeq,
		FirstAppVersion:  ev.AppVersion,
		AppVersion:       ev.AppVersion,
		LatestAppVersion: ev.AppVersion,
		AppBuild:         ev.AppBuild,
		AppNamespace:     ev.AppNamespace,
		ReleaseID:        ev.ReleaseID,
	}
	return occ
}

func baseOcc(ev Event) occurrence {
	return occurrence{
		Route:            ev.Route,
		FlowName:         ev.FlowName,
		AnonymousID:      ev.AnonymousID,
		SessionID:        ev.SessionID,
		FirstMS:          ev.Timestamp,
		LastMS:           ev.Timestamp,
		Count:            1,
		Truncated:        ev.Truncated,
		PageSeq:          ev.PageSeq,
		FirstAppVersion:  ev.AppVersion,
		AppVersion:       ev.AppVersion,
		LatestAppVersion: ev.AppVersion,
		AppBuild:         ev.AppBuild,
		AppNamespace:     ev.AppNamespace,
		ReleaseID:        ev.ReleaseID,
	}
}

func normalizeFieldKeys(keys []string) ([]string, string) {
	valid := make([]string, 0, len(keys))
	for _, key := range keys {
		if fieldKeyRe.MatchString(key) {
			valid = append(valid, key)
			if len(valid) == MaxFieldKeys {
				break
			}
		}
	}
	sort.Strings(valid)
	return valid, strings.Join(valid, ",")
}

func cleanErrorCode(code string) string {
	if errorCodeRe.MatchString(code) {
		return code
	}
	return ""
}

func normalizeSteps(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, step := range in {
		if !stepRe.MatchString(step) || seen[step] {
			return nil
		}
		seen[step] = true
		out = append(out, step)
	}
	return out
}

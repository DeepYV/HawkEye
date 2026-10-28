package issues

import "strings"

// ImpactScore is users × weight + min(occurrences, 999). Weight is 1000 when
// the flow is important and 100 otherwise. The occurrence cap applies only here.
func ImpactScore(users, occurrences int, important bool) int {
	weight := RankWeightDefault
	if important {
		weight = RankWeightImportant
	}
	extra := occurrences
	if extra > OccurrenceScoreCap {
		extra = OccurrenceScoreCap
	}
	if extra < 0 {
		extra = 0
	}
	if users < 0 {
		users = 0
	}
	return users*weight + extra
}

func (s *Store) collect(nowMS, cutoff int64, includeLow bool) []Issue {
	groups := map[string][]occurrence{}
	for _, occ := range s.occs {
		if occ.LastMS < cutoff || occ.LastMS > nowMS {
			continue
		}
		id := IssueID(occ.Type, occ.Route, occ.Signature, occ.FlowName)
		groups[id] = append(groups[id], occ)
	}
	issues := make([]Issue, 0, len(groups))
	for id, occs := range groups {
		iss := s.buildIssue(id, occs, cutoff, nowMS)
		if !includeLow && iss.Confidence == ConfidenceLow {
			continue
		}
		issues = append(issues, iss)
	}
	sortIssues(issues)
	return issues
}

func sortIssues(issues []Issue) {
	// score desc, lastSeen desc, id asc.
	for i := 1; i < len(issues); i++ {
		j := i
		for j > 0 && issueLess(issues[j], issues[j-1]) {
			issues[j], issues[j-1] = issues[j-1], issues[j]
			j--
		}
	}
}

func issueLess(a, b Issue) bool {
	if a.ImpactScore != b.ImpactScore {
		return a.ImpactScore > b.ImpactScore
	}
	if a.LastSeen != b.LastSeen {
		return a.LastSeen > b.LastSeen
	}
	return a.ID < b.ID
}

func (s *Store) buildIssue(id string, occs []occurrence, cutoff, nowMS int64) Issue {
	users := map[string]struct{}{}
	sessions := map[string]struct{}{}
	var corroborated, truncated bool
	var maxRun int
	var info relInfo
	first := occs[0]
	last := occs[0]
	for _, occ := range occs {
		if occ.AnonymousID != "" {
			users[occ.AnonymousID] = struct{}{}
		}
		if occ.SessionID != "" {
			sessions[occ.SessionID] = struct{}{}
		}
		if occ.DeadCorroborated {
			corroborated = true
		}
		if occ.Truncated {
			truncated = true
		}
		if occ.Count > maxRun {
			maxRun = occ.Count
		}
		if occ.FirstMS < first.FirstMS || (occ.FirstMS == first.FirstMS && occ.ID < first.ID) {
			first = occ
		}
		if occ.LastMS > last.LastMS || (occ.LastMS == last.LastMS && occ.ID > last.ID) {
			last = occ
		}
		info.add(occ.LastMS, occ.LatestAppVersion, occ.AppBuild, occ.AppNamespace, occ.ReleaseID)
	}

	important := last.FlowName != "" && s.important[last.FlowName]
	iss := Issue{
		ID:               id,
		Type:             last.Type,
		Route:            last.Route,
		ElementSignature: last.Signature,
		FlowName:         last.FlowName,
		Users:            len(users),
		Occurrences:      len(occs),
		FirstSeen:        first.FirstMS,
		LastSeen:         last.LastMS,
		Release:          info.releaseOrVersion(),
		Important:        important,
		maxRun:           maxRun,
	}
	iss.ImpactScore = ImpactScore(iss.Users, iss.Occurrences, important)

	ev := Evidence{
		FormKey:             last.FormKey,
		FieldKeys:           append([]string(nil), last.FieldKeys...),
		ErrorCode:           last.ErrorCode,
		Kind:                last.Kind,
		Method:              last.Method,
		Path:                last.Path,
		Name:                last.Name,
		File:                last.File,
		Function:            last.Function,
		Step:                last.Step,
		EndReason:           last.EndReason,
		SessionCount:        len(sessions),
		FirstSeenAppVersion: first.FirstAppVersion,
		LastSeenAppVersion:  last.AppVersion,
		AppBuild:            info.build,
		AppNamespace:        info.namespace,
		ReleaseID:           info.releaseID,
		Truncated:           truncated,
	}
	if last.Type == IssueRepeatedClick {
		ev.MaxBurst = maxRun
	}
	if last.Kind == KindHTTP {
		status := last.Status
		ev.Status = &status
	}
	if last.Type == IssueFlowAbandoned {
		ev.AttemptsReached = s.attemptsReached(last.FlowName, last.Step, cutoff, nowMS)
	}
	iss.Evidence = ev
	iss.Confidence, iss.Evidence.ConfidenceRule = confidence(iss, corroborated)
	iss.Observed, iss.Explanation, iss.Suggestion = suggest(iss)
	return iss
}

func confidence(iss Issue, corroborated bool) (string, string) {
	switch iss.Type {
	case IssueRepeatedClick:
		if strings.HasPrefix(iss.ElementSignature, "he:") && iss.Users >= 3 && corroborated {
			return ConfidenceHigh, "repeated_click.host_token.dead_corroborated.users_3"
		}
		if namedControl(iss.ElementSignature) {
			return ConfidenceMedium, "repeated_click.named_control"
		}
		return ConfidenceLow, "repeated_click.coarse_control"
	case IssueDeadClick:
		if strings.HasPrefix(iss.ElementSignature, "he:") && iss.Users >= 3 {
			return ConfidenceHigh, "dead_click.host_token.users_3"
		}
		if namedControl(iss.ElementSignature) {
			return ConfidenceMedium, "dead_click.named_control"
		}
		return ConfidenceLow, "dead_click.coarse_control"
	case IssueValidationLoop:
		fields := len(iss.Evidence.FieldKeys) > 0
		code := iss.Evidence.ErrorCode != ""
		if fields && code && iss.Users >= 2 {
			return ConfidenceHigh, "validation_loop.fields_and_code.users_2"
		}
		if fields {
			return ConfidenceMedium, "validation_loop.fields"
		}
		return ConfidenceLow, "validation_loop.form_only"
	case IssueFailedAction:
		return failedConfidence(iss)
	case IssueFlowAbandoned:
		end := iss.Evidence.EndReason
		reached := iss.Evidence.AttemptsReached
		if end == EndHostFailed || end == EndHostAbandoned ||
			((end == EndWindow || end == EndRestart) && reached >= 10 && majority(iss.Occurrences, reached)) {
			return ConfidenceHigh, "flow_abandoned.explicit_or_majority"
		}
		if reached >= 3 {
			return ConfidenceMedium, "flow_abandoned.window_attempts_3"
		}
		return ConfidenceLow, "flow_abandoned.window_sparse"
	default:
		return ConfidenceLow, ""
	}
}

func failedConfidence(iss Issue) (string, string) {
	ev := iss.Evidence
	status := 0
	if ev.Status != nil {
		status = *ev.Status
	}
	confirmed := iss.Users >= 2 && ((ev.Kind == KindHTTP && (status >= 500 || status == 0)) ||
		(ev.Kind == KindJS && realFile(ev.File)) ||
		(ev.Kind == KindSubmit && ev.ErrorCode != ""))
	if confirmed {
		return ConfidenceHigh, "failed_action.confirmed.users_2"
	}
	client := ev.Kind == KindHTTP && status >= 400 && status <= 499
	single := iss.Users == 1 && (realFile(ev.File) || ev.Kind == KindHTTP || ev.ErrorCode != "")
	if client || single {
		return ConfidenceMedium, "failed_action.single_or_client_error"
	}
	if ev.Kind == KindJS && ev.File == "-" {
		return ConfidenceLow, "failed_action.thin_exception"
	}
	return ConfidenceLow, "failed_action.thin_exception"
}

func namedControl(sig string) bool {
	if strings.HasPrefix(sig, "he:") {
		return true
	}
	// Label slug is ctrl:tag:role:slug. The id form contains '#' and is low.
	return strings.HasPrefix(sig, "ctrl:") && !strings.Contains(sig, "#") && strings.Count(sig, ":") >= 3
}

func realFile(file string) bool {
	return file != "" && file != "-"
}

func majority(occurrences, reached int) bool {
	if reached <= 0 {
		return false
	}
	return float64(occurrences)/float64(reached) >= 0.5
}

// relInfo tracks the latest non-empty release fields inside one window.
type relInfo struct {
	saw           bool
	firstVersion  string
	lastVersion   string
	latestVersion string
	haveVersion   bool
	versionAt     int64
	build         string
	haveBuild     bool
	buildAt       int64
	namespace     string
	haveNS        bool
	nsAt          int64
	releaseID     string
	haveRelease   bool
	releaseAt     int64
}

func (r *relInfo) add(ts int64, version, build, namespace, releaseID string) {
	if !r.saw {
		r.firstVersion = version
		r.saw = true
	}
	r.lastVersion = version
	if version != "" && (!r.haveVersion || ts >= r.versionAt) {
		r.latestVersion = version
		r.versionAt = ts
		r.haveVersion = true
	}
	if build != "" && (!r.haveBuild || ts >= r.buildAt) {
		r.build = build
		r.buildAt = ts
		r.haveBuild = true
	}
	if namespace != "" && (!r.haveNS || ts >= r.nsAt) {
		r.namespace = namespace
		r.nsAt = ts
		r.haveNS = true
	}
	if releaseID != "" && (!r.haveRelease || ts >= r.releaseAt) {
		r.releaseID = releaseID
		r.releaseAt = ts
		r.haveRelease = true
	}
}

func (r relInfo) releaseOrVersion() string {
	if r.releaseID != "" {
		return r.releaseID
	}
	return r.latestVersion
}

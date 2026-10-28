package issues

import "sort"

// attempt is one host flow entrance. Derived fields are rebuilt from Events
// so a late event cannot depend on arrival order.
type attempt struct {
	ID            string
	FlowName      string
	SessionID     string
	AnonymousID   string
	Steps         []string
	WindowMS      int64
	Important     bool
	Events        []flowEvent
	Externals     []extEnd
	Open          bool
	Void          bool
	Completed     bool
	EndReason     string
	StartedAt     int64
	Deadline      int64
	EndedAt       int64
	Furthest      int
	StepRoutes    []string
	AppVersion    string
	FirstVersion  string
	LastVersion   string
	LatestVersion string
	AppBuild      string
	AppNamespace  string
	ReleaseID     string
}

type flowEvent struct {
	Timestamp    int64
	MonoMS       int64
	PageSeq      int
	Step         string
	Status       string
	Route        string
	AnonymousID  string
	SessionID    string
	AppVersion   string
	AppBuild     string
	AppNamespace string
	ReleaseID    string
}

// extEnd closes an open attempt because another flow started in the same tab.
type extEnd struct {
	At int64
}

func clampWindow(ms int64) int64 {
	if ms >= FlowWindowMinMS && ms <= FlowWindowMaxMS {
		return ms
	}
	return FlowWindowDefaultMS
}

func (s *Store) foldFlow(ev Event, nowMS int64) {
	if ev.AttemptID == "" || ev.FlowName == "" {
		return
	}
	a := s.attempts[ev.AttemptID]
	if a == nil {
		steps := normalizeSteps(ev.Steps)
		if len(steps) == 0 {
			return
		}
		a = &attempt{
			ID:          ev.AttemptID,
			FlowName:    ev.FlowName,
			SessionID:   ev.SessionID,
			AnonymousID: ev.AnonymousID,
			Steps:       steps,
			WindowMS:    clampWindow(ev.WindowMS),
			Important:   ev.Important,
			Furthest:    -1,
		}
		s.attempts[a.ID] = a
		// Only the first event of the attempt carries important. Once true for
		// the flow name, later attempts cannot clear it.
		if ev.Important {
			s.important[ev.FlowName] = true
		}
	}
	a.Events = append(a.Events, flowEvent{
		Timestamp:    ev.Timestamp,
		MonoMS:       ev.MonoMS,
		PageSeq:      ev.PageSeq,
		Step:         ev.Step,
		Status:       ev.Status,
		Route:        ev.Route,
		AnonymousID:  ev.AnonymousID,
		SessionID:    ev.SessionID,
		AppVersion:   ev.AppVersion,
		AppBuild:     ev.AppBuild,
		AppNamespace: ev.AppNamespace,
		ReleaseID:    ev.ReleaseID,
	})
	prevStarted := a.StartedAt
	s.recompute(a, nowMS)
	if prevStarted == 0 && a.StartedAt > 0 {
		s.closeOthers(a, nowMS)
	}
}

func (s *Store) closeOthers(a *attempt, nowMS int64) {
	if a.SessionID == "" {
		return
	}
	for _, other := range s.attempts {
		if other.ID == a.ID || other.SessionID != a.SessionID || !other.Open {
			continue
		}
		other.Externals = append(other.Externals, extEnd{At: a.StartedAt})
		s.recompute(other, nowMS)
	}
}

func (s *Store) recompute(a *attempt, nowMS int64) {
	events := append([]flowEvent(nil), a.Events...)
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Timestamp != events[j].Timestamp {
			return events[i].Timestamp < events[j].Timestamp
		}
		if events[i].PageSeq != events[j].PageSeq {
			return events[i].PageSeq < events[j].PageSeq
		}
		return events[i].MonoMS < events[j].MonoMS
	})
	exts := append([]extEnd(nil), a.Externals...)
	sort.SliceStable(exts, func(i, j int) bool { return exts[i].At < exts[j].At })

	a.Open = false
	a.Void = false
	a.Completed = false
	a.EndReason = ""
	a.StartedAt = 0
	a.Deadline = 0
	a.EndedAt = 0
	a.Furthest = -1
	a.StepRoutes = make([]string, len(a.Steps))
	a.FirstVersion = ""
	a.LastVersion = ""
	a.LatestVersion = ""
	a.AppVersion = ""
	a.AppBuild = ""
	a.AppNamespace = ""
	a.ReleaseID = ""

	var info relInfo
	for _, ev := range events {
		info.add(ev.Timestamp, ev.AppVersion, ev.AppBuild, ev.AppNamespace, ev.ReleaseID)
		if a.AnonymousID == "" && ev.AnonymousID != "" {
			a.AnonymousID = ev.AnonymousID
		}
	}
	if len(events) > 0 {
		a.FirstVersion = events[0].AppVersion
		a.LastVersion = events[len(events)-1].AppVersion
	}
	a.LatestVersion = info.latestVersion
	a.AppVersion = info.latestVersion
	a.AppBuild = info.build
	a.AppNamespace = info.namespace
	a.ReleaseID = info.releaseID

	ei, xi := 0, 0
	for ei < len(events) || xi < len(exts) {
		if xi >= len(exts) || (ei < len(events) && events[ei].Timestamp <= exts[xi].At) {
			applyFlowEvent(a, events[ei])
			ei++
		} else {
			applyExternal(a, exts[xi])
			xi++
		}
	}
	if a.Open && a.Deadline > 0 && nowMS > a.Deadline {
		a.Open = false
		a.EndReason = EndWindow
		a.EndedAt = a.Deadline
	}
	s.syncFlowOcc(a)
}

func applyFlowEvent(a *attempt, ev flowEvent) {
	if a.Completed || a.EndReason != "" || a.Void {
		return
	}
	if a.Open && ev.Timestamp > a.Deadline {
		return
	}
	idx := stepIndex(a.Steps, ev.Step)
	if idx < 0 {
		return
	}
	switch ev.Status {
	case "entered":
		if !a.Open {
			if idx == 0 {
				openAttempt(a, ev)
			}
			return
		}
		if idx == 0 && a.Furthest == 0 {
			a.StartedAt = ev.Timestamp
			a.Deadline = ev.Timestamp + a.WindowMS
			a.StepRoutes[0] = ev.Route
			return
		}
		if idx == 0 && a.Furthest > 0 {
			a.Open = false
			a.EndReason = EndRestart
			a.EndedAt = ev.Timestamp
			return
		}
		if idx == a.Furthest || idx == a.Furthest+1 {
			if idx > a.Furthest {
				a.Furthest = idx
			}
			a.StepRoutes[idx] = ev.Route
		}
	case "completed":
		if !a.Open {
			return
		}
		finalIdx := len(a.Steps) - 1
		if idx != finalIdx {
			return
		}
		if idx == a.Furthest || idx == a.Furthest+1 {
			if idx > a.Furthest {
				a.Furthest = idx
				a.StepRoutes[idx] = ev.Route
			}
			a.Completed = true
			a.Open = false
			a.EndReason = ""
			a.EndedAt = ev.Timestamp
		}
	case "failed", "abandoned":
		if !a.Open {
			return
		}
		if ev.Status == "failed" {
			a.EndReason = EndHostFailed
		} else {
			a.EndReason = EndHostAbandoned
		}
		a.Open = false
		a.EndedAt = ev.Timestamp
	}
}

func openAttempt(a *attempt, ev flowEvent) {
	a.Open = true
	a.StartedAt = ev.Timestamp
	a.Deadline = ev.Timestamp + a.WindowMS
	a.Furthest = 0
	a.StepRoutes[0] = ev.Route
	if ev.AnonymousID != "" {
		a.AnonymousID = ev.AnonymousID
	}
}

func applyExternal(a *attempt, e extEnd) {
	if !a.Open || a.EndReason != "" || a.Completed || a.Void {
		return
	}
	if a.Deadline > 0 && e.At > a.Deadline {
		return
	}
	a.Open = false
	a.EndedAt = e.At
	if a.Furthest > 0 {
		a.EndReason = EndRestart
		return
	}
	a.Void = true
	a.EndReason = ""
}

func (s *Store) syncFlowOcc(a *attempt) {
	id := "flow:" + a.ID
	if a.EndReason == "" || a.Furthest < 0 || a.Furthest >= len(a.Steps) {
		delete(s.occs, id)
		return
	}
	step := a.Steps[a.Furthest]
	route := ""
	if a.Furthest < len(a.StepRoutes) {
		route = a.StepRoutes[a.Furthest]
	}
	s.occs[id] = occurrence{
		ID:               id,
		Type:             IssueFlowAbandoned,
		Route:            route,
		Signature:        "step:" + step + ":" + a.EndReason,
		FlowName:         a.FlowName,
		AnonymousID:      a.AnonymousID,
		SessionID:        a.SessionID,
		FirstMS:          a.StartedAt,
		LastMS:           a.EndedAt,
		Count:            1,
		Step:             step,
		EndReason:        a.EndReason,
		FirstAppVersion:  a.FirstVersion,
		AppVersion:       a.LastVersion,
		LatestAppVersion: a.LatestVersion,
		AppBuild:         a.AppBuild,
		AppNamespace:     a.AppNamespace,
		ReleaseID:        a.ReleaseID,
	}
}

func (s *Store) closeExpired(nowMS int64) {
	for _, a := range s.attempts {
		if a.Open {
			s.recompute(a, nowMS)
		}
	}
}

func (s *Store) attemptsReached(flowName, step string, cutoff, nowMS int64) int {
	n := 0
	for _, a := range s.attempts {
		if a.FlowName != flowName || a.StartedAt == 0 {
			continue
		}
		if a.StartedAt < cutoff || a.StartedAt > nowMS {
			continue
		}
		idx := stepIndex(a.Steps, step)
		if idx >= 0 && a.Furthest >= idx {
			n++
		}
	}
	return n
}

func stepIndex(steps []string, step string) int {
	for i, name := range steps {
		if name == step {
			return i
		}
	}
	return -1
}

func (s *Store) flowReport(nowMS, cutoff int64) []FlowSummary {
	byName := map[string][]*attempt{}
	for _, a := range s.attempts {
		if a.StartedAt == 0 || a.StartedAt < cutoff || a.StartedAt > nowMS {
			continue
		}
		byName[a.FlowName] = append(byName[a.FlowName], a)
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]FlowSummary, 0, len(names))
	for _, name := range names {
		group := byName[name]
		var steps []string
		var best int64 = -1
		for _, a := range group {
			if a.StartedAt >= best && len(a.Steps) > 0 {
				best = a.StartedAt
				steps = a.Steps
			}
		}
		sum := FlowSummary{
			Name:      name,
			Important: s.important[name],
			Attempts:  len(group),
			Steps:     make([]FlowStepSummary, 0, len(steps)),
		}
		for _, a := range group {
			switch {
			case a.Completed && a.EndReason == "":
				sum.Completed++
			case a.EndReason != "":
				sum.Abandoned++
			case a.Open:
				sum.Open++
			}
		}
		for _, step := range steps {
			reached, abandoned := 0, 0
			for _, a := range group {
				idx := stepIndex(a.Steps, step)
				if idx < 0 {
					continue
				}
				if a.Furthest >= idx {
					reached++
				}
				if a.EndReason != "" && a.Furthest == idx {
					abandoned++
				}
			}
			sum.Steps = append(sum.Steps, FlowStepSummary{Step: step, Reached: reached, Abandoned: abandoned})
		}
		out = append(out, sum)
	}
	return out
}

package issues

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Store is the in-memory issue engine. Every method uses the same mutex.
// Apply and Sweep take now so tests do not sleep.
type Store struct {
	mu        sync.Mutex
	seen      map[string]int64
	clicks    map[string][]clickRec
	clickLoc  map[string]string
	outcomes  map[string]outcomeRec
	forms     map[formGroup][]formRec
	occs      map[string]occurrence
	attempts  map[string]*attempt
	important map[string]bool
}

// NewStore returns an empty engine.
func NewStore() *Store {
	return &Store{
		seen:      map[string]int64{},
		clicks:    map[string][]clickRec{},
		clickLoc:  map[string]string{},
		outcomes:  map[string]outcomeRec{},
		forms:     map[formGroup][]formRec{},
		occs:      map[string]occurrence{},
		attempts:  map[string]*attempt{},
		important: map[string]bool{},
	}
}

// Apply folds events as of now and returns how many events were kept.
func (s *Store) Apply(now time.Time, events []Event) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.keep(events)
	touchedSess := map[string]struct{}{}
	touchedForms := map[formGroup]struct{}{}
	nowMS := now.UnixMilli()
	for _, ev := range kept {
		switch ev.Type {
		case TypeClick:
			s.addClick(ev)
			touchedSess[ev.SessionID] = struct{}{}
		case TypeClickOutcome:
			if sess, ok := s.addOutcome(ev); ok {
				touchedSess[sess] = struct{}{}
			}
		case TypeFormOutcome:
			if g, ok := s.addForm(ev); ok {
				touchedForms[g] = struct{}{}
			}
		case TypeError:
			s.addError(ev)
		case TypeFlowStep:
			s.foldFlow(ev, nowMS)
		}
	}
	for sess := range touchedSess {
		s.rebuildBursts(sess)
	}
	for g := range touchedForms {
		s.rebuildForms(g)
	}
	s.closeExpired(nowMS)
	return len(kept)
}

// Sweep closes expired attempts and drops raw events, occurrences, and seen
// ids older than 7 days. Open attempts stay until they end.
func (s *Store) Sweep(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nowMS := now.UnixMilli()
	s.closeExpired(nowMS)
	cutoff := now.Add(-RetentionWindow).UnixMilli()

	for id, ts := range s.seen {
		if ts < cutoff {
			delete(s.seen, id)
		}
	}
	for id, o := range s.outcomes {
		if o.At < cutoff {
			delete(s.outcomes, id)
		}
	}

	sessSet := map[string]struct{}{}
	for sess, clicks := range s.clicks {
		kept := make([]clickRec, 0, len(clicks))
		for _, c := range clicks {
			if c.Timestamp >= cutoff {
				kept = append(kept, c)
			} else {
				delete(s.clickLoc, c.EventID)
			}
		}
		if len(kept) == 0 {
			delete(s.clicks, sess)
		} else {
			s.clicks[sess] = kept
		}
		sessSet[sess] = struct{}{}
	}
	for _, occ := range s.occs {
		if occ.Type == IssueRepeatedClick {
			sessSet[occ.SessionID] = struct{}{}
		}
	}
	for sess := range sessSet {
		s.rebuildBursts(sess)
	}

	for g, evs := range s.forms {
		kept := make([]formRec, 0, len(evs))
		for _, ev := range evs {
			if ev.Timestamp >= cutoff {
				kept = append(kept, ev)
			}
		}
		if len(kept) == 0 {
			delete(s.forms, g)
		} else {
			s.forms[g] = kept
		}
		s.rebuildForms(g)
	}

	var drop []string
	for id, a := range s.attempts {
		if a.Open {
			continue
		}
		ts := a.EndedAt
		if ts == 0 {
			ts = a.StartedAt
		}
		if ts < cutoff {
			drop = append(drop, id)
		}
	}
	for _, id := range drop {
		delete(s.attempts, id)
	}
	for id, occ := range s.occs {
		if occ.LastMS < cutoff {
			delete(s.occs, id)
		}
	}
}

// List returns ranked issues in the window.
func (s *Store) List(q Query) ListResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	now, window, cutoff, nowMS := normalizeWindow(q.Now, q.Window)
	_ = now
	issues := s.collect(nowMS, cutoff, q.IncludeLow)
	limit := q.Limit
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	if limit > 0 && len(issues) > limit {
		issues = issues[:limit]
	}
	if issues == nil {
		issues = []Issue{}
	}
	return ListResult{WindowMs: window.Milliseconds(), Issues: issues}
}

// Get returns one issue in the window, including low confidence.
func (s *Store) Get(now time.Time, id string, window time.Duration) (Issue, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, window, cutoff, nowMS := normalizeWindow(now, window)
	_ = window
	id = strings.ToLower(id)
	for _, iss := range s.collect(nowMS, cutoff, true) {
		if iss.ID == id {
			return iss, true
		}
	}
	return Issue{}, false
}

// Flows summarizes attempts whose entrance is inside the window.
func (s *Store) Flows(now time.Time, window time.Duration) FlowReport {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, window, cutoff, nowMS := normalizeWindow(now, window)
	flows := s.flowReport(nowMS, cutoff)
	if flows == nil {
		flows = []FlowSummary{}
	}
	return FlowReport{WindowMs: window.Milliseconds(), Flows: flows}
}

func normalizeWindow(now time.Time, window time.Duration) (time.Time, time.Duration, int64, int64) {
	if now.IsZero() {
		now = time.Now()
	}
	if window <= 0 {
		window = DefaultWindow
	}
	return now, window, now.Add(-window).UnixMilli(), now.UnixMilli()
}

func (s *Store) keep(events []Event) []Event {
	kept := make([]Event, 0, len(events))
	batch := map[string]struct{}{}
	for _, ev := range events {
		if ev.EventID == "" || ev.AnonymousID == "" || ev.SessionID == "" || ev.Type == "" {
			continue
		}
		if _, ok := s.seen[ev.EventID]; ok {
			continue
		}
		if _, ok := batch[ev.EventID]; ok {
			continue
		}
		if !accepted(ev) {
			continue
		}
		batch[ev.EventID] = struct{}{}
		s.seen[ev.EventID] = ev.Timestamp
		kept = append(kept, ev)
	}
	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].Timestamp != kept[j].Timestamp {
			return kept[i].Timestamp < kept[j].Timestamp
		}
		if kept[i].PageSeq != kept[j].PageSeq {
			return kept[i].PageSeq < kept[j].PageSeq
		}
		return kept[i].MonoMS < kept[j].MonoMS
	})
	return kept
}

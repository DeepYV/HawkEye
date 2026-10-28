package issues

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const baseMS int64 = 1_700_000_000_000

func at(ms int64) time.Time { return time.UnixMilli(ms) }

func listAt(st *Store, now time.Time, low bool) ListResult {
	return st.List(Query{Now: now, Window: DefaultWindow, IncludeLow: low, Limit: MaxListLimit})
}

func findIssue(issues []Issue, typ, flow string, matchFlow bool) (Issue, bool) {
	for _, iss := range issues {
		if iss.Type != typ {
			continue
		}
		if matchFlow && iss.FlowName != flow {
			continue
		}
		return iss, true
	}
	return Issue{}, false
}

func shell() Event {
	return Event{
		Timestamp:   baseMS,
		PageSeq:     1,
		AnonymousID: "user-1",
		SessionID:   "sess-1",
		Route:       "/signup",
	}
}

func TestIssueIDVectors(t *testing.T) {
	got := IssueID(IssueDeadClick, "/signup", "he:signup-submit", "signup")
	if got != "ec03db7020bd5be8a2198e7515819ae1a11a09e65dde9f1287aa4b0f28b10938" {
		t.Fatalf("signup flow id = %s", got)
	}
	got = IssueID(IssueDeadClick, "/signup", "he:signup-submit", "")
	if got != "1b2f7b8c6474e9fed045900f3ffa03dfae552143d0d4c9e6386aa9c52f963894" {
		t.Fatalf("empty flow id = %s", got)
	}
}

func TestImpactScoresFromEvents(t *testing.T) {
	cases := []struct {
		name      string
		users     int
		occ       int
		flow      string
		important bool
		want      int
	}{
		{name: "12 users 40 important", users: 12, occ: 40, flow: "signup", important: true, want: 12040},
		{name: "12 users 40 unmarked", users: 12, occ: 40, want: 1240},
		{name: "3 users 200 important", users: 3, occ: 200, flow: "signup", important: true, want: 3200},
		{name: "50 users 50 unmarked", users: 50, occ: 50, want: 5050},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			iss := applyDead(t, tc.users, tc.occ, tc.flow, tc.important, "/r", "he:x")
			if iss.ImpactScore != tc.want {
				t.Fatalf("score = %d, want %d", iss.ImpactScore, tc.want)
			}
			if iss.ImpactScore != ImpactScore(tc.users, tc.occ, tc.important) {
				t.Fatalf("score %d disagrees with ImpactScore", iss.ImpactScore)
			}
			if iss.Users != tc.users || iss.Occurrences != tc.occ {
				t.Fatalf("users=%d occ=%d", iss.Users, iss.Occurrences)
			}
			if iss.Important != tc.important {
				t.Fatalf("important = %v", iss.Important)
			}
		})
	}
}

func TestThousandOccurrencesCap(t *testing.T) {
	iss := applyDead(t, 1, 1000, "", false, "/r", "he:x")
	if iss.ImpactScore != 1099 {
		t.Fatalf("score = %d, want 1099 (100 + min(1000, 999))", iss.ImpactScore)
	}
	want := "On /r, he:x was clicked and no response arrived within 2500 ms. This happened 1000 times, across 1 users."
	if iss.Observed != want {
		t.Fatalf("observed = %q", iss.Observed)
	}
	if !strings.Contains(iss.Observed, "1000") || strings.Contains(iss.Observed, "999 times") {
		t.Fatalf("sentence should show 1000, got %q", iss.Observed)
	}
}

func TestGoldenDeadClick(t *testing.T) {
	st := NewStore()
	now := at(baseMS)
	events := []Event{flowOpen("signup", true)}
	events = append(events, deadClicks(3, 4, "signup", "/signup", "he:signup-submit")...)
	events = append(events, deadClicks(3, 4, "", "/signup", "he:signup-submit")...)
	// The unmarked batch reuses click ids from the marked batch. Give them a prefix.
	for i := range events {
		if events[i].Type == TypeClickOutcome && events[i].FlowName == "" {
			events[i].EventID = "u-" + events[i].EventID
			events[i].ClickID = "u-" + events[i].ClickID
			events[i].AnonymousID = "plain-" + events[i].AnonymousID
			events[i].SessionID = "plain-" + events[i].SessionID
		}
	}
	st.Apply(now, events)
	res := listAt(st, now, false)

	marked, ok := findIssue(res.Issues, IssueDeadClick, "signup", true)
	if !ok {
		t.Fatal("missing important dead click")
	}
	plain, ok := findIssue(res.Issues, IssueDeadClick, "", true)
	if !ok {
		t.Fatal("missing unmarked dead click")
	}
	const sentence = "On /signup, he:signup-submit was clicked and no response arrived within 2500 ms. This happened 4 times, across 3 users."
	const explanation = "The control presents as clickable, and the page produced no navigation, host outcome, or disabled, aria-busy, or aria-expanded change on that element."
	const suggestion = "Attach the missing action, or stop presenting the element as clickable. When the action starts, set disabled or aria-busy on this element."
	for _, iss := range []Issue{marked, plain} {
		if iss.Observed != sentence {
			t.Errorf("%s observed = %q", iss.FlowName, iss.Observed)
		}
		if iss.Explanation != explanation {
			t.Errorf("explanation = %q", iss.Explanation)
		}
		if iss.Suggestion != suggestion {
			t.Errorf("suggestion = %q", iss.Suggestion)
		}
		if iss.Confidence != ConfidenceHigh || iss.Evidence.ConfidenceRule != "dead_click.host_token.users_3" {
			t.Errorf("confidence %s %s", iss.Confidence, iss.Evidence.ConfidenceRule)
		}
		if iss.Users != 3 || iss.Occurrences != 4 {
			t.Errorf("users %d occ %d", iss.Users, iss.Occurrences)
		}
	}
	if marked.ImpactScore != 3004 || !marked.Important {
		t.Fatalf("important score = %d important=%v", marked.ImpactScore, marked.Important)
	}
	if plain.ImpactScore != 304 || plain.Important {
		t.Fatalf("unmarked score = %d important=%v", plain.ImpactScore, plain.Important)
	}
	if marked.ID != IssueID(IssueDeadClick, "/signup", "he:signup-submit", "signup") {
		t.Fatalf("marked id %s", marked.ID)
	}
	if plain.ID != IssueID(IssueDeadClick, "/signup", "he:signup-submit", "") {
		t.Fatalf("plain id %s", plain.ID)
	}
	if res.WindowMs != 604800000 {
		t.Fatalf("windowMs = %d", res.WindowMs)
	}
}

func TestRepeatedClickBursts(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	// Arrival order is reversed. Gaps of 999 and 999 are one burst of length 3.
	clicks := []Event{
		clickAt("c3", "he:save", "/home", "sess", "user", 1, baseMS, 1998),
		clickAt("c1", "he:save", "/home", "sess", "user", 1, baseMS, 0),
		clickAt("c2", "he:save", "/home", "sess", "user", 1, baseMS, 999),
	}
	st.Apply(now, clicks)
	res := listAt(st, now, true)
	iss, ok := findIssue(res.Issues, IssueRepeatedClick, "", true)
	if !ok {
		t.Fatal("expected one repeated click")
	}
	if iss.Occurrences != 1 || iss.Evidence.MaxBurst != 3 || iss.maxRun != 3 {
		t.Fatalf("occ=%d max=%d", iss.Occurrences, iss.Evidence.MaxBurst)
	}
	if iss.Observed != "On /home, he:save was clicked repeatedly in 1 bursts by 1 users. The largest burst was 3 clicks." {
		t.Fatalf("observed = %q", iss.Observed)
	}

	st = NewStore()
	st.Apply(now, []Event{
		clickAt("a", "he:save", "/home", "sess", "user", 1, baseMS, 0),
		clickAt("b", "he:save", "/home", "sess", "user", 1, baseMS, 1000),
		clickAt("c", "he:save", "/home", "sess", "user", 1, baseMS, 2000),
	})
	if _, ok := findIssue(listAt(st, now, true).Issues, IssueRepeatedClick, "", false); ok {
		t.Fatal("gap of 1000 must not be a burst")
	}

	st = NewStore()
	st.Apply(now, []Event{
		clickAt("a", "he:save", "/home", "sess", "user", 1, baseMS, 0),
		clickAt("b", "he:save", "/home", "sess", "user", 1, baseMS, 500),
	})
	if _, ok := findIssue(listAt(st, now, true).Issues, IssueRepeatedClick, "", false); ok {
		t.Fatal("two clicks are not an issue")
	}
}

func TestDeadClickRules(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	click := clickAt("c1", "he:save", "/signup", "sess", "user-a", 1, baseMS, 0)
	st.Apply(now, []Event{click})
	if _, ok := findIssue(listAt(st, now, true).Issues, IssueDeadClick, "", false); ok {
		t.Fatal("a click with no outcome is not dead")
	}

	st = NewStore()
	dead := deadOutcome("o1", "c-missing", "user-a", "sess", "he:save", "/signup", "")
	st.Apply(now, []Event{dead})
	iss, ok := findIssue(listAt(st, now, true).Issues, IssueDeadClick, "", false)
	if !ok {
		t.Fatal("dead outcome without the click must create the issue")
	}
	if iss.Users != 1 || iss.Occurrences != 1 {
		t.Fatalf("users=%d occ=%d", iss.Users, iss.Occurrences)
	}

	second := deadOutcome("o2", "c-missing", "user-b", "sess-b", "he:other", "/elsewhere", "")
	st.Apply(now, []Event{second})
	iss, _ = findIssue(listAt(st, now, true).Issues, IssueDeadClick, "", false)
	if iss.Users != 1 || iss.Occurrences != 1 {
		t.Fatalf("duplicate click_id added a user: users=%d occ=%d", iss.Users, iss.Occurrences)
	}

	st = NewStore()
	responded := dead
	responded.EventID = "r1"
	responded.Result = "responded"
	responded.Reason = "navigation"
	later := dead
	later.EventID = "d1"
	st.Apply(now, []Event{responded, later})
	if _, ok := findIssue(listAt(st, now, true).Issues, IssueDeadClick, "", false); ok {
		t.Fatal("responded outcome must not become dead, and a later dead for the same click_id is ignored")
	}
}

func TestDeadCorroboratesBurst(t *testing.T) {
	now := at(baseMS)
	var clicks []Event
	for u := 0; u < 3; u++ {
		user := fmt.Sprintf("u%d", u)
		sess := fmt.Sprintf("s%d", u)
		for c := 0; c < 3; c++ {
			clicks = append(clicks, clickAt(
				fmt.Sprintf("%s-%d", user, c),
				"he:save", "/home", sess, user, 1, baseMS, int64(c*999),
			))
		}
	}
	st := NewStore()
	st.Apply(now, clicks)
	iss, ok := findIssue(listAt(st, now, true).Issues, IssueRepeatedClick, "", false)
	if !ok {
		t.Fatal("expected repeated click")
	}
	if iss.Confidence != ConfidenceMedium || iss.Evidence.ConfidenceRule != "repeated_click.named_control" {
		t.Fatalf("before dead: %s %s", iss.Confidence, iss.Evidence.ConfidenceRule)
	}
	st.Apply(now, []Event{deadOutcome("out", "u0-0", "u0", "s0", "he:save", "/home", "")})
	iss, _ = findIssue(listAt(st, now, true).Issues, IssueRepeatedClick, "", false)
	if iss.Users != 3 || iss.Occurrences != 3 {
		t.Fatalf("users=%d occ=%d", iss.Users, iss.Occurrences)
	}
	if iss.Confidence != ConfidenceHigh || iss.Evidence.ConfidenceRule != "repeated_click.host_token.dead_corroborated.users_3" {
		t.Fatalf("after dead: %s %s", iss.Confidence, iss.Evidence.ConfidenceRule)
	}
}

func TestValidationLoop(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	var evs []Event
	for i, mono := range []int64{0, 1000, 2000} {
		ev := shell()
		ev.EventID = fmt.Sprintf("v%d", i)
		ev.Type = TypeFormOutcome
		ev.MonoMS = mono
		ev.Timestamp = baseMS + mono
		ev.FormKey = "signup"
		ev.Result = "validation_error"
		if i%2 == 0 {
			ev.FieldKeys = []string{"password", "email"}
		} else {
			ev.FieldKeys = []string{"email", "password"}
		}
		ev.ErrorCode = "mismatch"
		evs = append(evs, ev)
	}
	st.Apply(at(baseMS+2000), evs)
	iss, ok := findIssue(listAt(st, at(baseMS+2000), true).Issues, IssueValidationLoop, "", false)
	if !ok {
		t.Fatal("third validation_error should open one loop")
	}
	if iss.Occurrences != 1 || iss.Evidence.FormKey != "signup" {
		t.Fatalf("%+v", iss)
	}
	if iss.ElementSignature != "form:signup:fields:email,password" {
		t.Fatalf("signature %s", iss.ElementSignature)
	}
	want := "On /signup, form signup rejected email, password 3 times in one sitting, across 1 sittings and 1 users. The latest error code was mismatch."
	if iss.Observed != want {
		t.Fatalf("observed = %q", iss.Observed)
	}
	if iss.Explanation != "The same fields failed validation again, so the message or the rule is not enough to correct the form." {
		t.Fatalf("explanation %q", iss.Explanation)
	}
	if iss.Suggestion != "Put the error on those fields, name the expected format, and keep the entries already typed." {
		t.Fatalf("suggestion %q", iss.Suggestion)
	}
	if _, ok := findIssue(listAt(st, at(baseMS+2000), true).Issues, IssueFailedAction, "", false); ok {
		t.Fatal("validation_error must not create failed_action")
	}

	okEv := shell()
	okEv.EventID = "ok"
	okEv.Type = TypeFormOutcome
	okEv.FormKey = "signup"
	okEv.Result = "success"
	okEv.MonoMS = 3000
	okEv.Timestamp = baseMS + 3000
	st.Apply(at(baseMS+3000), []Event{okEv})

	again := shell()
	again.EventID = "v3"
	again.Type = TypeFormOutcome
	again.FormKey = "signup"
	again.Result = "validation_error"
	again.FieldKeys = []string{"email", "password"}
	again.MonoMS = 4000
	again.Timestamp = baseMS + 4000
	again.ErrorCode = "mismatch"
	st.Apply(at(baseMS+4000), []Event{again})
	iss, _ = findIssue(listAt(st, at(baseMS+4000), true).Issues, IssueValidationLoop, "", false)
	if iss.Occurrences != 1 || !strings.Contains(iss.Observed, " 3 times in one sitting") {
		t.Fatalf("success should end the run, got occ=%d %q", iss.Occurrences, iss.Observed)
	}

	st = NewStore()
	short := evs[:2]
	st.Apply(now, short)
	sub := shell()
	sub.EventID = "sub"
	sub.Type = TypeFormOutcome
	sub.FormKey = "signup"
	sub.Result = "submit_error"
	sub.ErrorCode = "timeout"
	sub.MonoMS = 5000
	sub.Timestamp = baseMS + 5000
	st.Apply(at(baseMS+5000), []Event{sub})
	if _, ok := findIssue(listAt(st, at(baseMS+5000), true).Issues, IssueValidationLoop, "", false); ok {
		t.Fatal("two validation errors plus submit_error are not a loop")
	}
	fail, ok := findIssue(listAt(st, at(baseMS+5000), true).Issues, IssueFailedAction, "", false)
	if !ok || fail.Evidence.Kind != KindSubmit {
		t.Fatal("submit_error should be failed_action")
	}
	if fail.Observed != "On /signup, form signup reported submit_error 1 times, across 1 users. The latest error code was timeout." {
		t.Fatalf("submit observed %q", fail.Observed)
	}

	wide := NewStore()
	var split []Event
	for i, mono := range []int64{0, ValidationLoopGapMS, ValidationLoopGapMS * 2} {
		ev := shell()
		ev.EventID = fmt.Sprintf("g%d", i)
		ev.Type = TypeFormOutcome
		ev.FormKey = "signup"
		ev.Result = "validation_error"
		ev.FieldKeys = []string{"email"}
		ev.MonoMS = mono
		ev.Timestamp = baseMS + mono
		split = append(split, ev)
	}
	wide.Apply(at(baseMS+ValidationLoopGapMS*2), split)
	if _, ok := findIssue(listAt(wide, at(baseMS+ValidationLoopGapMS*2), true).Issues, IssueValidationLoop, "", false); ok {
		t.Fatal("gap of 120000 must end the validation run")
	}
}

func TestFailedActionTemplates(t *testing.T) {
	now := at(baseMS)
	httpEv := shell()
	httpEv.EventID = "h1"
	httpEv.Type = TypeError
	httpEv.Kind = KindHTTP
	httpEv.Method = "get"
	httpEv.HTTPStatus = 500
	httpEv.Path = "/api/pay"
	httpEv.Route = "/checkout"
	httpEv.Truncated = true
	st := NewStore()
	st.Apply(now, []Event{httpEv})
	iss, ok := findIssue(listAt(st, now, true).Issues, IssueFailedAction, "", false)
	if !ok {
		t.Fatal("http 500")
	}
	if iss.ElementSignature != "http:GET:500:/api/pay" {
		t.Fatalf("signature %s", iss.ElementSignature)
	}
	if iss.Observed != "On /checkout, GET /api/pay returned status 500, 1 times, across 1 users. The count is a lower bound because the page stopped sending duplicates." {
		t.Fatalf("http observed %q", iss.Observed)
	}
	if iss.Explanation != "The server failed the request." {
		t.Fatalf("explanation %q", iss.Explanation)
	}
	if iss.Suggestion != "Show a retry for this action and fix the failing request." {
		t.Fatalf("suggestion %q", iss.Suggestion)
	}

	httpEv.EventID = "h2"
	httpEv.HTTPStatus = 404
	httpEv.Truncated = false
	st = NewStore()
	st.Apply(now, []Event{httpEv})
	iss, _ = findIssue(listAt(st, now, true).Issues, IssueFailedAction, "", false)
	if iss.Explanation != "The server rejected the request." || iss.Suggestion != "Show the error code on the form and correct the request this control sends." {
		t.Fatalf("404 text %q / %q", iss.Explanation, iss.Suggestion)
	}

	httpEv.EventID = "h0"
	httpEv.HTTPStatus = 0
	st = NewStore()
	st.Apply(now, []Event{httpEv})
	iss, _ = findIssue(listAt(st, now, true).Issues, IssueFailedAction, "", false)
	if iss.Explanation != "The browser could not complete the request." {
		t.Fatalf("status 0 explanation %q", iss.Explanation)
	}

	httpEv.EventID = "ok"
	httpEv.HTTPStatus = 200
	st = NewStore()
	if n := st.Apply(now, []Event{httpEv}); n != 1 {
		t.Fatalf("status 200 is still accepted, processed %d", n)
	}
	if _, ok := findIssue(listAt(st, now, true).Issues, IssueFailedAction, "", false); ok {
		t.Fatal("status 200 is not failed_action")
	}

	js := shell()
	js.EventID = "j1"
	js.Type = TypeError
	js.Kind = KindJS
	js.Name = "TypeError"
	js.MessageSlug = "cannot-read"
	js.File = "app.js"
	js.Function = "submit"
	js.Route = "/checkout"
	st = NewStore()
	st.Apply(now, []Event{js})
	iss, _ = findIssue(listAt(st, now, true).Issues, IssueFailedAction, "", false)
	if iss.ElementSignature != "js:TypeError:cannot-read:app.js:submit" {
		t.Fatalf("js signature %s", iss.ElementSignature)
	}
	if iss.Observed != "On /checkout, TypeError in app.js submit was uncaught, 1 times, across 1 users." {
		t.Fatalf("js observed %q", iss.Observed)
	}
	if iss.Explanation != "An uncaught exception interrupted the page." {
		t.Fatalf("js explanation %q", iss.Explanation)
	}
	if iss.Suggestion != "Handle this exception in submit so the control can be used again." {
		t.Fatalf("js suggestion %q", iss.Suggestion)
	}
}

func TestFlowDeadlineAndComplete(t *testing.T) {
	const window int64 = 60000
	start := baseMS

	openStore := NewStore()
	openStore.Apply(at(start+window), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account"}, window, false),
	})
	rep := openStore.Flows(at(start+window), DefaultWindow)
	if len(rep.Flows) != 1 || rep.Flows[0].Open != 1 || rep.Flows[0].Completed != 0 || rep.Flows[0].Abandoned != 0 {
		t.Fatalf("now == deadline should stay open: %+v", rep.Flows)
	}
	if _, ok := findIssue(listAt(openStore, at(start+window), true).Issues, IssueFlowAbandoned, "", false); ok {
		t.Fatal("open attempt must not write flow_abandoned")
	}

	closed := NewStore()
	closed.Apply(at(start+window+1), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account"}, window, false),
	})
	iss, ok := findIssue(listAt(closed, at(start+window+1), true).Issues, IssueFlowAbandoned, "signup", true)
	if !ok {
		t.Fatal("now == deadline+1 should write flow_abandoned")
	}
	if iss.Evidence.EndReason != EndWindow || iss.Evidence.Step != "account" {
		t.Fatalf("end %+v", iss.Evidence)
	}
	if iss.ElementSignature != "step:account:window" || iss.Route != "/signup" {
		t.Fatalf("sig %s route %s", iss.ElementSignature, iss.Route)
	}
	if iss.Confidence != ConfidenceLow {
		t.Fatalf("single window attempt should be low, got %s", iss.Confidence)
	}
	if _, ok := findIssue(listAt(closed, at(start+window+1), false).Issues, IssueFlowAbandoned, "", false); ok {
		t.Fatal("low confidence must be omitted unless IncludeLow")
	}
	wantObs := "Flow signup stopped at step account (window) for 1 of 1 attempts that reached this step, across 1 users."
	if iss.Observed != wantObs {
		t.Fatalf("observed %q", iss.Observed)
	}
	if iss.Explanation != "Those attempts reached account and did not reach the next step within the flow window." {
		t.Fatalf("explanation %q", iss.Explanation)
	}
	if iss.Suggestion != "At account, say what is still required and keep the attempt resumable." {
		t.Fatalf("suggestion %q", iss.Suggestion)
	}
	flows := closed.Flows(at(start+window+1), DefaultWindow)
	if flows.Flows[0].Open != 0 || flows.Flows[0].Abandoned != 1 || flows.Flows[0].Completed != 0 {
		t.Fatalf("counts %+v", flows.Flows[0])
	}

	closed.Apply(at(start+window+1), []Event{
		flowStep("e2", "a1", "signup", "account", "completed", "/signup", start+window+5000, nil, 0, false),
	})
	iss, ok = findIssue(listAt(closed, at(start+window+1), true).Issues, IssueFlowAbandoned, "signup", true)
	if !ok || iss.Evidence.EndReason != EndWindow {
		t.Fatal("a later completed must not remove flow_abandoned")
	}
	flows = closed.Flows(at(start+window+1), DefaultWindow)
	if flows.Flows[0].Completed != 0 || flows.Flows[0].Abandoned != 1 {
		t.Fatalf("completed cleared the abandonment: %+v", flows.Flows[0])
	}

	done := NewStore()
	done.Apply(at(start+1000), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account"}, window, false),
		flowStep("e2", "a1", "signup", "account", "completed", "/signup", start+1000, nil, 0, false),
	})
	if _, ok := findIssue(listAt(done, at(start+1000), true).Issues, IssueFlowAbandoned, "", false); ok {
		t.Fatal("complete on the final step of a one-step flow writes no issue")
	}
	rep = done.Flows(at(start+1000), DefaultWindow)
	if len(rep.Flows) != 1 || rep.Flows[0].Completed != 1 || rep.Flows[0].Attempts != 1 || rep.Flows[0].Open != 0 {
		t.Fatalf("completed count %+v", rep.Flows)
	}
	if len(rep.Flows[0].Steps) != 1 || rep.Flows[0].Steps[0].Step != "account" || rep.Flows[0].Steps[0].Reached != 1 || rep.Flows[0].Steps[0].Abandoned != 0 {
		t.Fatalf("steps %+v", rep.Flows[0].Steps)
	}
}

func TestFlowRestartResetAndSkip(t *testing.T) {
	const window int64 = 60000
	start := baseMS

	reset := NewStore()
	reset.Apply(at(start+window+1), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account", "profile"}, window, false),
		flowStep("e2", "a1", "signup", "account", "entered", "/signup", start+50000, nil, 0, false),
	})
	if _, ok := findIssue(listAt(reset, at(start+window+1), true).Issues, IssueFlowAbandoned, "", false); ok {
		t.Fatal("re-entering step 1 before a later step writes no issue")
	}
	rep := reset.Flows(at(start+window+1), DefaultWindow)
	if rep.Flows[0].Open != 1 || rep.Flows[0].Abandoned != 0 {
		t.Fatalf("clock should have reset, still open: %+v", rep.Flows[0])
	}

	restart := NewStore()
	restart.Apply(at(start+2000), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account", "profile"}, window, false),
		flowStep("e2", "a1", "signup", "profile", "entered", "/profile", start+1000, nil, 0, false),
		flowStep("e3", "a1", "signup", "account", "entered", "/signup", start+2000, nil, 0, false),
	})
	iss, ok := findIssue(listAt(restart, at(start+2000), true).Issues, IssueFlowAbandoned, "signup", true)
	if !ok {
		t.Fatal("re-entering step 1 after step 2 should write restart")
	}
	if iss.Evidence.EndReason != EndRestart || iss.Evidence.Step != "profile" || iss.Route != "/profile" {
		t.Fatalf("restart issue %+v sig %s", iss.Evidence, iss.ElementSignature)
	}
	if iss.Explanation != "The person started this flow again after reaching profile." {
		t.Fatalf("explanation %q", iss.Explanation)
	}

	skip := NewStore()
	skip.Apply(at(start+1000), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account", "profile", "review"}, window, false),
		flowStep("e2", "a1", "signup", "review", "completed", "/review", start+1000, nil, 0, false),
	})
	rep = skip.Flows(at(start+1000), DefaultWindow)
	if rep.Flows[0].Completed != 0 || rep.Flows[0].Open != 1 {
		t.Fatalf("skip-ahead completed must not complete: %+v", rep.Flows[0])
	}
	if rep.Flows[0].Steps[0].Reached != 1 || rep.Flows[0].Steps[2].Reached != 0 {
		t.Fatalf("steps %+v", rep.Flows[0].Steps)
	}
	if _, ok := findIssue(listAt(skip, at(start+1000), true).Issues, IssueFlowAbandoned, "", false); ok {
		t.Fatal("skip-ahead is not an abandonment while the window is open")
	}
}

func TestHostFailedAndAbandoned(t *testing.T) {
	now := at(baseMS + 1000)
	st := NewStore()
	st.Apply(now, []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", baseMS, []string{"account"}, 60000, true),
		flowStep("e2", "a1", "signup", "account", "failed", "/signup", baseMS+1000, nil, 0, false),
	})
	iss, ok := findIssue(listAt(st, now, false).Issues, IssueFlowAbandoned, "signup", true)
	if !ok {
		t.Fatal("host_failed should be visible without IncludeLow")
	}
	if iss.Evidence.EndReason != EndHostFailed || iss.Confidence != ConfidenceHigh {
		t.Fatalf("%s %s", iss.Evidence.EndReason, iss.Confidence)
	}
	if iss.Evidence.ConfidenceRule != "flow_abandoned.explicit_or_majority" {
		t.Fatalf("rule %s", iss.Evidence.ConfidenceRule)
	}
	if !iss.Important || iss.ImpactScore != 1001 {
		t.Fatalf("score %d important %v", iss.ImpactScore, iss.Important)
	}
	if iss.Explanation != "The host marked the attempt failed at account." {
		t.Fatalf("explanation %q", iss.Explanation)
	}
	if iss.Suggestion != "At account, say what is still required and keep the attempt resumable. Fix the failure at account before adding later steps." {
		t.Fatalf("suggestion %q", iss.Suggestion)
	}

	st = NewStore()
	st.Apply(now, []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", baseMS, []string{"account"}, 60000, false),
		flowStep("e2", "a1", "signup", "account", "abandoned", "/signup", baseMS+1000, nil, 0, false),
	})
	iss, ok = findIssue(listAt(st, now, false).Issues, IssueFlowAbandoned, "signup", true)
	if !ok || iss.Evidence.EndReason != EndHostAbandoned {
		t.Fatal("host abandoned")
	}
	if iss.Explanation != "The host marked the attempt abandoned at account." {
		t.Fatalf("explanation %q", iss.Explanation)
	}
}

func TestImportantStaysAndDoesNotFlip(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	st.Apply(now, []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", baseMS, []string{"account"}, 86400000, true),
		flowStep("e2", "a1", "signup", "account", "completed", "/signup", baseMS+1, nil, 0, false),
		flowStep("e3", "a2", "signup", "account", "entered", "/signup", baseMS+2, []string{"account"}, 86400000, false),
		deadOutcome("d1", "c1", "user-9", "sess-9", "he:save", "/signup", "signup"),
	})
	iss, ok := findIssue(listAt(st, now, true).Issues, IssueDeadClick, "signup", true)
	if !ok {
		t.Fatal("dead click")
	}
	if !iss.Important || iss.ImpactScore != 1001 {
		t.Fatalf("sticky important score=%d flag=%v", iss.ImpactScore, iss.Important)
	}

	st = NewStore()
	st.Apply(now, []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", baseMS, []string{"account"}, 86400000, false),
		flowStep("e2", "a1", "signup", "account", "entered", "/signup", baseMS+1, nil, 0, true),
		deadOutcome("d1", "c1", "user-9", "sess-9", "he:save", "/signup", "signup"),
	})
	iss, _ = findIssue(listAt(st, now, true).Issues, IssueDeadClick, "signup", true)
	if iss.Important || iss.ImpactScore != 101 {
		t.Fatalf("later event must not set important, score=%d flag=%v", iss.ImpactScore, iss.Important)
	}
}

func TestFlowWindowClampAndMissingSteps(t *testing.T) {
	start := baseMS
	st := NewStore()
	n := st.Apply(at(start+1001), []Event{
		flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, []string{"account"}, 1000, false),
	})
	if n != 1 {
		t.Fatalf("processed %d", n)
	}
	rep := st.Flows(at(start+1001), DefaultWindow)
	if len(rep.Flows) != 1 || rep.Flows[0].Open != 1 {
		t.Fatalf("window 1000 must clamp to 1800000 and stay open: %+v", rep.Flows)
	}
	if _, ok := findIssue(listAt(st, at(start+1001), true).Issues, IssueFlowAbandoned, "", false); ok {
		t.Fatal("clamped window should still be open")
	}

	st = NewStore()
	ev := flowStep("e1", "a1", "signup", "account", "entered", "/signup", start, nil, 60000, true)
	if st.Apply(at(start), []Event{ev}) != 1 {
		t.Fatal("processed")
	}
	if flows := st.Flows(at(start), DefaultWindow); len(flows.Flows) != 0 {
		t.Fatalf("missing steps must not open an attempt: %+v", flows.Flows)
	}
	dead := deadOutcome("d1", "c1", "user-1", "sess-1", "he:save", "/signup", "signup")
	st.Apply(at(start), []Event{dead})
	iss, _ := findIssue(listAt(st, at(start), true).Issues, IssueDeadClick, "signup", true)
	if iss.Important {
		t.Fatal("dropped flow event must not set important")
	}
}

func TestConfidenceFilter(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	st.Apply(now, []Event{
		deadOutcome("low", "c-low", "user-1", "sess-1", "ctrl:button:button#save", "/edit", ""),
		deadOutcome("med", "c-med", "user-1", "sess-1", "he:signup-submit", "/signup", ""),
	})
	def := listAt(st, now, false)
	if _, ok := findIssue(def.Issues, IssueDeadClick, "", false); !ok {
		t.Fatal("he: with one user is medium and belongs on the default list")
	}
	for _, iss := range def.Issues {
		if iss.ElementSignature == "ctrl:button:button#save" {
			t.Fatal("low confidence id signature was included")
		}
		if iss.ElementSignature == "he:signup-submit" && (iss.Confidence != ConfidenceMedium || iss.Evidence.ConfidenceRule != "dead_click.named_control") {
			t.Fatalf("he: one user = %s %s", iss.Confidence, iss.Evidence.ConfidenceRule)
		}
	}
	with := listAt(st, now, true)
	low, ok := findSig(with.Issues, "ctrl:button:button#save")
	if !ok {
		t.Fatal("IncludeLow should return the id signature")
	}
	if low.Confidence != ConfidenceLow || low.Evidence.ConfidenceRule != "dead_click.coarse_control" {
		t.Fatalf("low = %s %s", low.Confidence, low.Evidence.ConfidenceRule)
	}
	got, found := st.Get(now, low.ID, DefaultWindow)
	if !found || got.ID != low.ID {
		t.Fatal("Get returns low confidence")
	}
	if _, found := st.Get(now, "nope", DefaultWindow); found {
		t.Fatal("missing id")
	}

	slug := deadOutcome("slug", "c-slug", "user-2", "sess-2", "ctrl:button:button:save-draft", "/edit", "")
	st.Apply(now, []Event{slug})
	iss, ok := findSig(listAt(st, now, false).Issues, "ctrl:button:button:save-draft")
	if !ok || iss.Confidence != ConfidenceMedium {
		t.Fatalf("label slug should be medium, ok=%v %+v", ok, iss.Confidence)
	}
}

func TestRankTieBreak(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	st.Apply(now, []Event{
		deadOutcome("a", "ca", "user-a", "sess-a", "he:a", "/a", ""),
		deadOutcome("b", "cb", "user-b", "sess-b", "he:b", "/b", ""),
	})
	res := listAt(st, now, false)
	if len(res.Issues) != 2 {
		t.Fatalf("len %d", len(res.Issues))
	}
	if res.Issues[0].ImpactScore != res.Issues[1].ImpactScore || res.Issues[0].LastSeen != res.Issues[1].LastSeen {
		t.Fatal("fixture should tie on score and lastSeen")
	}
	if res.Issues[0].ID > res.Issues[1].ID {
		t.Fatalf("id order %s then %s", res.Issues[0].ID, res.Issues[1].ID)
	}
}

func TestApplyFiltersAndSweep(t *testing.T) {
	now := at(baseMS)
	st := NewStore()
	good := clickAt("c1", "he:save", "/home", "sess", "user", 1, baseMS, 0)
	emptySig := good
	emptySig.EventID = "c2"
	emptySig.ElementSignature = ""
	unknown := good
	unknown.EventID = "c3"
	unknown.Type = "scroll"
	dup := good
	n := st.Apply(now, []Event{good, emptySig, unknown, dup})
	if n != 1 {
		t.Fatalf("processed = %d, want 1", n)
	}
	if st.Apply(now, []Event{good}) != 0 {
		t.Fatal("duplicate event_id should not be accepted again")
	}

	later := at(baseMS).Add(8 * 24 * time.Hour)
	dead := deadOutcome("d1", "cx", "user", "sess", "he:save", "/home", "")
	st.Apply(now, []Event{dead})
	wide := st.List(Query{Now: later, Window: 30 * 24 * time.Hour, IncludeLow: true})
	if len(wide.Issues) != 1 {
		t.Fatalf("before sweep len=%d", len(wide.Issues))
	}
	st.Sweep(later)
	wide = st.List(Query{Now: later, Window: 30 * 24 * time.Hour, IncludeLow: true})
	if len(wide.Issues) != 0 {
		t.Fatalf("sweep should drop events older than 7 days, len=%d", len(wide.Issues))
	}
}

func TestDecodeBatch(t *testing.T) {
	if _, _, err := DecodeBatch([]byte("{")); err == nil {
		t.Fatal("invalid JSON should error")
	}
	events, legacy, err := DecodeBatch([]byte(`{"events":[{"eventType":"click","sessionId":"s"},{"eventType":"scroll","sessionId":"s"}]}`))
	if err != nil || !legacy || events != nil {
		t.Fatalf("legacy err=%v legacy=%v events=%v", err, legacy, events)
	}
	events, legacy, err = DecodeBatch([]byte(`{"sdk_version":"1.0.0","events":[{"event_id":"e1","type":"click","timestamp":1700000000000,"mono_ms":5,"page_seq":2,"anonymous_id":"a","session_id":"s","route":"/signup","element_signature":"he:signup-submit","flow_name":"signup"}]}`))
	if err != nil || legacy || len(events) != 1 {
		t.Fatalf("new err=%v legacy=%v len=%d", err, legacy, len(events))
	}
	ev := events[0]
	if ev.Type != "click" || ev.ElementSignature != "he:signup-submit" || ev.Timestamp != 1700000000000 || ev.MonoMS != 5 || ev.PageSeq != 2 || ev.FlowName != "signup" {
		t.Fatalf("decoded %+v", ev)
	}
	events, legacy, err = DecodeBatch([]byte(`{"events":[{"eventType":"click","type":"click","event_id":"e","anonymous_id":"a","session_id":"s"}]}`))
	if err != nil || legacy || len(events) != 1 || events[0].Type != "click" {
		t.Fatalf("mixed type should not be legacy: err=%v legacy=%v %+v", err, legacy, events)
	}
	events, legacy, err = DecodeBatch([]byte(`{"events":[{"event_id":"e","type":"scroll","anonymous_id":"a","session_id":"s","timestamp":1}]}`))
	if err != nil || legacy || len(events) != 1 || events[0].Type != "scroll" {
		t.Fatalf("unknown type should be returned: %+v err=%v", events, err)
	}
	st := NewStore()
	if n := st.Apply(at(1), events); n != 0 {
		t.Fatalf("Apply drops unknown types, processed %d", n)
	}
}

func TestReleaseSelection(t *testing.T) {
	now := at(baseMS + 10)
	first := deadOutcome("a", "ca", "user", "sess", "he:save", "/signup", "")
	first.Timestamp = baseMS
	first.AppVersion = "1.0.0"
	first.ReleaseID = ""
	second := deadOutcome("b", "cb", "user", "sess", "he:save", "/signup", "")
	second.Timestamp = baseMS + 10
	second.AppVersion = "1.2.0"
	second.ReleaseID = "rel-9"
	second.AppBuild = "88"
	second.AppNamespace = "web"
	st := NewStore()
	st.Apply(now, []Event{second, first})
	iss, ok := findIssue(listAt(st, now, true).Issues, IssueDeadClick, "", false)
	if !ok {
		t.Fatal("issue")
	}
	if iss.Release != "rel-9" || iss.Evidence.ReleaseID != "rel-9" {
		t.Fatalf("release %s %s", iss.Release, iss.Evidence.ReleaseID)
	}
	if iss.Evidence.FirstSeenAppVersion != "1.0.0" || iss.Evidence.LastSeenAppVersion != "1.2.0" {
		t.Fatalf("versions %s %s", iss.Evidence.FirstSeenAppVersion, iss.Evidence.LastSeenAppVersion)
	}
	if iss.FirstSeen != baseMS || iss.LastSeen != baseMS+10 {
		t.Fatalf("seen %d %d", iss.FirstSeen, iss.LastSeen)
	}
}

func applyDead(t *testing.T, users, occ int, flow string, important bool, route, sig string) Issue {
	t.Helper()
	st := NewStore()
	now := at(baseMS)
	var evs []Event
	if flow != "" {
		evs = append(evs, flowStep("flow-open", "att-score", flow, "account", "entered", route, baseMS, []string{"account"}, 86400000, important))
	}
	evs = append(evs, deadClicks(users, occ, flow, route, sig)...)
	st.Apply(now, evs)
	iss, ok := findIssue(listAt(st, now, true).Issues, IssueDeadClick, flow, true)
	if !ok {
		t.Fatalf("missing dead_click flow %q", flow)
	}
	return iss
}

func deadClicks(users, occ int, flow, route, sig string) []Event {
	out := make([]Event, 0, occ)
	for i := 0; i < occ; i++ {
		user := fmt.Sprintf("u%d", i%users)
		out = append(out, deadOutcome(fmt.Sprintf("e%d", i), fmt.Sprintf("c%d", i), user, "s"+user, sig, route, flow))
	}
	return out
}

func deadOutcome(eventID, clickID, user, session, sig, route, flow string) Event {
	ev := shell()
	ev.EventID = eventID
	ev.Type = TypeClickOutcome
	ev.AnonymousID = user
	ev.SessionID = session
	ev.Route = route
	ev.FlowName = flow
	ev.ElementSignature = sig
	ev.ClickID = clickID
	ev.Result = "dead"
	ev.Reason = "timeout"
	ev.ElapsedMS = 2500
	return ev
}

func clickAt(id, sig, route, session, user string, page int, ts, mono int64) Event {
	ev := shell()
	ev.EventID = id
	ev.Type = TypeClick
	ev.ElementSignature = sig
	ev.Route = route
	ev.SessionID = session
	ev.AnonymousID = user
	ev.PageSeq = page
	ev.Timestamp = ts
	ev.MonoMS = mono
	return ev
}

func flowOpen(name string, important bool) Event {
	return flowStep("flow-"+name, "att-"+name, name, "account", "entered", "/signup", baseMS, []string{"account"}, 86400000, important)
}

func flowStep(id, attempt, flow, step, status, route string, ts int64, steps []string, window int64, important bool) Event {
	ev := shell()
	ev.EventID = id
	ev.Type = TypeFlowStep
	ev.Timestamp = ts
	ev.AttemptID = attempt
	ev.FlowName = flow
	ev.Step = step
	ev.Status = status
	ev.Route = route
	ev.Steps = steps
	ev.WindowMS = window
	ev.Important = important
	ev.AnonymousID = "flow-user"
	ev.SessionID = "flow-sess"
	return ev
}

func findSig(issues []Issue, sig string) (Issue, bool) {
	for _, iss := range issues {
		if iss.ElementSignature == sig {
			return iss, true
		}
	}
	return Issue{}, false
}

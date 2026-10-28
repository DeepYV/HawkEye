package issues

import (
	"fmt"
	"strings"
)

const lowerBoundSentence = "The count is a lower bound because the page stopped sending duplicates."

// suggest fills the three sentences from section 8. The words are fixed.
func suggest(iss Issue) (observed, explanation, suggestion string) {
	switch iss.Type {
	case IssueRepeatedClick:
		observed = fmt.Sprintf(
			"On %s, %s was clicked repeatedly in %d bursts by %d users. The largest burst was %d clicks.",
			iss.Route, iss.ElementSignature, iss.Occurrences, iss.Users, iss.maxRun,
		)
		explanation = "People activated this control again before it acknowledged the first activation."
		suggestion = "After the first activation, disable the control or set aria-busy until the action finishes, and show that the action was accepted."
	case IssueDeadClick:
		observed = fmt.Sprintf(
			"On %s, %s was clicked and no response arrived within %d ms. This happened %d times, across %d users.",
			iss.Route, iss.ElementSignature, DeadClickDeadlineMS, iss.Occurrences, iss.Users,
		)
		explanation = "The control presents as clickable, and the page produced no navigation, host outcome, or disabled, aria-busy, or aria-expanded change on that element."
		suggestion = "Attach the missing action, or stop presenting the element as clickable. When the action starts, set disabled or aria-busy on this element."
	case IssueValidationLoop:
		observed = fmt.Sprintf(
			"On %s, form %s rejected %s %d times in one sitting, across %d sittings and %d users. The latest error code was %s.",
			iss.Route, iss.Evidence.FormKey, renderFields(iss.Evidence.FieldKeys), iss.maxRun, iss.Occurrences, iss.Users, renderCode(iss.Evidence.ErrorCode),
		)
		explanation = "The same fields failed validation again, so the message or the rule is not enough to correct the form."
		suggestion = "Put the error on those fields, name the expected format, and keep the entries already typed."
	case IssueFailedAction:
		observed, explanation, suggestion = suggestFailed(iss)
	case IssueFlowAbandoned:
		step := iss.Evidence.Step
		end := iss.Evidence.EndReason
		observed = fmt.Sprintf(
			"Flow %s stopped at step %s (%s) for %d of %d attempts that reached this step, across %d users.",
			iss.FlowName, step, end, iss.Occurrences, iss.Evidence.AttemptsReached, iss.Users,
		)
		switch end {
		case EndHostFailed:
			explanation = fmt.Sprintf("The host marked the attempt failed at %s.", step)
		case EndHostAbandoned:
			explanation = fmt.Sprintf("The host marked the attempt abandoned at %s.", step)
		case EndRestart:
			explanation = fmt.Sprintf("The person started this flow again after reaching %s.", step)
		default:
			explanation = fmt.Sprintf("Those attempts reached %s and did not reach the next step within the flow window.", step)
		}
		suggestion = fmt.Sprintf("At %s, say what is still required and keep the attempt resumable.", step)
		if end == EndHostFailed {
			suggestion += " " + fmt.Sprintf("Fix the failure at %s before adding later steps.", step)
		}
	}
	if iss.Evidence.Truncated {
		observed += " " + lowerBoundSentence
	}
	return observed, explanation, suggestion
}

func suggestFailed(iss Issue) (observed, explanation, suggestion string) {
	ev := iss.Evidence
	switch ev.Kind {
	case KindHTTP:
		status := 0
		if ev.Status != nil {
			status = *ev.Status
		}
		observed = fmt.Sprintf(
			"On %s, %s %s returned status %d, %d times, across %d users.",
			iss.Route, ev.Method, ev.Path, status, iss.Occurrences, iss.Users,
		)
		switch {
		case status >= 500:
			explanation = "The server failed the request."
		case status >= 400 && status <= 499:
			explanation = "The server rejected the request."
		default:
			explanation = "The browser could not complete the request."
		}
		if status >= 500 || status == 0 {
			suggestion = "Show a retry for this action and fix the failing request."
		} else {
			suggestion = "Show the error code on the form and correct the request this control sends."
		}
	case KindJS:
		observed = fmt.Sprintf(
			"On %s, %s in %s %s was uncaught, %d times, across %d users.",
			iss.Route, ev.Name, ev.File, ev.Function, iss.Occurrences, iss.Users,
		)
		explanation = "An uncaught exception interrupted the page."
		suggestion = fmt.Sprintf("Handle this exception in %s so the control can be used again.", ev.Function)
	default:
		observed = fmt.Sprintf(
			"On %s, form %s reported submit_error %d times, across %d users. The latest error code was %s.",
			iss.Route, ev.FormKey, iss.Occurrences, iss.Users, renderCode(ev.ErrorCode),
		)
		explanation = "The form was submitted and the host marked the submission failed."
		suggestion = "Show the error code on the form and return the person to the step that can correct it."
	}
	return observed, explanation, suggestion
}

func renderFields(keys []string) string {
	if len(keys) == 0 {
		return "-"
	}
	return strings.Join(keys, ", ")
}

func renderCode(code string) string {
	if code == "" {
		return "-"
	}
	return code
}

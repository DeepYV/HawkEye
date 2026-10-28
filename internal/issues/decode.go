package issues

import (
	"encoding/json"
	"errors"
	"strings"
)

// DecodeBatch parses an ingest body.
//
// Invalid JSON returns an error. If every event object has eventType and none
// has a non-empty type, legacy is true and events is nil so the caller can use
// the old ingest path. If any event has a non-empty type, legacy is false and
// the normalized events are returned. Unknown types are kept for Apply to drop.
func DecodeBatch(body []byte) ([]Event, bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		if err == nil {
			err = errors.New("invalid request body")
		}
		return nil, false, err
	}

	evRaw, ok := raw["events"]
	if !ok || strings.TrimSpace(string(evRaw)) == "null" {
		return nil, true, nil
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(evRaw, &arr); err != nil {
		return nil, false, err
	}
	if len(arr) == 0 {
		return nil, true, nil
	}

	allEventType := true
	anyType := false
	objects := make([]map[string]json.RawMessage, len(arr))
	for i, item := range arr {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(item, &obj); err != nil || obj == nil {
			return nil, false, errors.New("invalid request body")
		}
		objects[i] = obj
		if _, ok := obj["eventType"]; !ok {
			allEventType = false
		}
		if rawType, ok := obj["type"]; ok && nonEmptyString(rawType) {
			anyType = true
		}
	}
	if !anyType && allEventType {
		return nil, true, nil
	}

	events := make([]Event, 0, len(objects))
	for _, obj := range objects {
		ev, err := decodeEvent(obj)
		if err != nil {
			return nil, false, err
		}
		events = append(events, ev)
	}
	return events, false, nil
}

func nonEmptyString(raw json.RawMessage) bool {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return false
	}
	return s != ""
}

func decodeEvent(obj map[string]json.RawMessage) (Event, error) {
	buf, err := json.Marshal(obj)
	if err != nil {
		return Event{}, err
	}
	var m map[string]any
	if err := json.Unmarshal(buf, &m); err != nil || m == nil {
		return Event{}, errors.New("invalid request body")
	}
	ev := Event{
		EventID:          str(m, "event_id"),
		Type:             str(m, "type"),
		Timestamp:        num(m, "timestamp"),
		MonoMS:           num(m, "mono_ms"),
		PageSeq:          int(num(m, "page_seq")),
		AnonymousID:      str(m, "anonymous_id"),
		SessionID:        str(m, "session_id"),
		Route:            str(m, "route"),
		FlowID:           str(m, "flow_id"),
		FlowName:         str(m, "flow_name"),
		AppVersion:       str(m, "app_version"),
		AppBuild:         str(m, "app_build"),
		AppNamespace:     str(m, "app_namespace"),
		ReleaseID:        str(m, "release_id"),
		Environment:      str(m, "environment"),
		ElementSignature: str(m, "element_signature"),
		FormKey:          str(m, "form_key"),
		ClickID:          str(m, "click_id"),
		Result:           str(m, "result"),
		Reason:           str(m, "reason"),
		ElapsedMS:        num(m, "elapsed_ms"),
		FromRoute:        str(m, "from_route"),
		ToRoute:          str(m, "to_route"),
		Kind:             str(m, "kind"),
		FieldKeys:        strSlice(m, "field_keys"),
		ErrorCode:        str(m, "error_code"),
		Name:             str(m, "name"),
		MessageSlug:      str(m, "message_slug"),
		File:             str(m, "file"),
		Function:         str(m, "function"),
		Line:             int(num(m, "line")),
		Column:           int(num(m, "column")),
		Truncated:        boolean(m, "truncated"),
		Method:           str(m, "method"),
		Path:             str(m, "path"),
		Step:             str(m, "step"),
		AttemptID:        str(m, "attempt_id"),
		Steps:            strSlice(m, "steps"),
		WindowMS:         num(m, "window_ms"),
		Important:        boolean(m, "important"),
	}
	// status is a string on flow_step and a number on HTTP errors.
	if ev.Type == TypeFlowStep {
		ev.Status = str(m, "status")
	} else {
		ev.HTTPStatus = int(num(m, "status"))
	}
	return ev, nil
}

func str(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func num(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case float64:
		return int64(v)
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			f, _ := v.Float64()
			return int64(f)
		}
		return i
	default:
		return 0
	}
}

func boolean(m map[string]any, key string) bool {
	b, _ := m[key].(bool)
	return b
}

func strSlice(m map[string]any, key string) []string {
	arr, ok := m[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

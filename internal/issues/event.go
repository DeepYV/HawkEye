package issues

import "time"

// Event is one normalized wire event. JSON on the wire is snake_case; this
// struct is what Apply consumes after DecodeBatch.
type Event struct {
	EventID      string
	Type         string
	Timestamp    int64
	MonoMS       int64
	PageSeq      int
	AnonymousID  string
	SessionID    string
	Route        string
	FlowID       string
	FlowName     string
	AppVersion   string
	AppBuild     string
	AppNamespace string
	ReleaseID    string
	Environment  string

	ElementSignature string
	FormKey          string

	ClickID   string
	Result    string
	Reason    string
	ElapsedMS int64

	FromRoute string
	ToRoute   string
	Kind      string

	FieldKeys []string
	ErrorCode string

	Name        string
	MessageSlug string
	File        string
	Function    string
	Line        int
	Column      int
	Truncated   bool
	Method      string
	HTTPStatus  int
	Path        string

	Step      string
	Status    string
	AttemptID string
	Steps     []string
	WindowMS  int64
	Important bool
}

// Evidence is the camelCase evidence object on an issue.
type Evidence struct {
	ConfidenceRule      string   `json:"confidenceRule"`
	MaxBurst            int      `json:"maxBurst,omitempty"`
	Truncated           bool     `json:"truncated,omitempty"`
	FormKey             string   `json:"formKey,omitempty"`
	FieldKeys           []string `json:"fieldKeys,omitempty"`
	ErrorCode           string   `json:"errorCode,omitempty"`
	Kind                string   `json:"kind,omitempty"`
	Method              string   `json:"method,omitempty"`
	Status              *int     `json:"status,omitempty"`
	Path                string   `json:"path,omitempty"`
	Name                string   `json:"name,omitempty"`
	File                string   `json:"file,omitempty"`
	Function            string   `json:"function,omitempty"`
	Step                string   `json:"step,omitempty"`
	EndReason           string   `json:"endReason,omitempty"`
	AttemptsReached     int      `json:"attemptsReached,omitempty"`
	SessionCount        int      `json:"sessionCount,omitempty"`
	FirstSeenAppVersion string   `json:"firstSeenAppVersion,omitempty"`
	LastSeenAppVersion  string   `json:"lastSeenAppVersion,omitempty"`
	AppBuild            string   `json:"appBuild,omitempty"`
	AppNamespace        string   `json:"appNamespace,omitempty"`
	ReleaseID           string   `json:"releaseId,omitempty"`
}

// Issue is one grouped problem. firstSeen and lastSeen are Unix milliseconds.
type Issue struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	Route            string   `json:"route"`
	ElementSignature string   `json:"elementSignature"`
	FlowName         string   `json:"flowName"`
	Users            int      `json:"users"`
	Occurrences      int      `json:"occurrences"`
	ImpactScore      int      `json:"impactScore"`
	Confidence       string   `json:"confidence"`
	Observed         string   `json:"observed"`
	Explanation      string   `json:"explanation"`
	Suggestion       string   `json:"suggestion"`
	FirstSeen        int64    `json:"firstSeen"`
	LastSeen         int64    `json:"lastSeen"`
	Release          string   `json:"release"`
	Important        bool     `json:"important"`
	Evidence         Evidence `json:"evidence"`

	// maxRun is the largest burst or validation run. It is not part of the API.
	maxRun int
}

// Query selects a ranked issue list. Window <= 0 means 7 days. Limit <= 0 means no cap.
type Query struct {
	Now        time.Time
	Window     time.Duration
	IncludeLow bool
	Limit      int
}

// ListResult is the GET /v1/issues body.
type ListResult struct {
	WindowMs int64   `json:"windowMs"`
	Issues   []Issue `json:"issues"`
}

// FlowStepSummary is one step of a flow in the requested window.
type FlowStepSummary struct {
	Step      string `json:"step"`
	Reached   int    `json:"reached"`
	Abandoned int    `json:"abandoned"`
}

// FlowSummary is one named flow in the requested window.
type FlowSummary struct {
	Name      string            `json:"name"`
	Important bool              `json:"important"`
	Attempts  int               `json:"attempts"`
	Completed int               `json:"completed"`
	Abandoned int               `json:"abandoned"`
	Open      int               `json:"open"`
	Steps     []FlowStepSummary `json:"steps"`
}

// FlowReport is the GET /v1/flows body.
type FlowReport struct {
	WindowMs int64         `json:"windowMs"`
	Flows    []FlowSummary `json:"flows"`
}

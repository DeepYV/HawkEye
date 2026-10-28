package issues

import "time"

// Detection numbers from the logic contract. Thresholds live here so call sites
// do not invent their own.
const (
	RepeatedClickMin    = 3
	RepeatedClickGapMS  = 1000
	DeadClickDeadlineMS = 2500
	LinkBlurMS          = 1000
	ValidationLoopMin   = 3
	ValidationLoopGapMS = 120000
	FlowWindowDefaultMS = 1800000
	FlowWindowMinMS     = 60000
	FlowWindowMaxMS     = 86400000
	ErrorBudgetCount    = 10
	ErrorBudgetInterval = 10000
	MessageSlugCap      = 120
	AriaLabelCap        = 40
	RankWeightImportant = 1000
	RankWeightDefault   = 100
	OccurrenceScoreCap  = 999
	MaxFieldKeys        = 8
	DefaultWindow       = 7 * 24 * time.Hour
	RetentionWindow     = 7 * 24 * time.Hour
	DefaultListLimit    = 50
	MaxListLimit        = 200
)

const (
	TypeClick        = "click"
	TypeClickOutcome = "click_outcome"
	TypeNavigation   = "navigation"
	TypeFormOutcome  = "form_outcome"
	TypeError        = "error"
	TypeBusiness     = "business"
	TypeFlowStep     = "flow_step"

	IssueRepeatedClick  = "repeated_click"
	IssueDeadClick      = "dead_click"
	IssueValidationLoop = "validation_loop"
	IssueFailedAction   = "failed_action"
	IssueFlowAbandoned  = "flow_abandoned"

	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"

	KindSubmit = "submit"
	KindHTTP   = "http"
	KindJS     = "js"

	EndWindow        = "window"
	EndHostAbandoned = "host_abandoned"
	EndHostFailed    = "host_failed"
	EndRestart       = "restart"
)

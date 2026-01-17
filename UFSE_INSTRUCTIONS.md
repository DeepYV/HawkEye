# User Frustration Signal Engine (UFSE) - Implementation Instructions

## 1. Your Role

You are building **ONE isolated backend module** named:

**User Frustration Signal Engine (UFSE)**

This module is **pure logic**.

### What This Module Does NOT Know About:
- ❌ UI
- ❌ AI
- ❌ Jira
- ❌ Linear
- ❌ SDK internals

### What This Module Knows About:
- ✅ Sessions
- ✅ Events

---

## 2. Single Responsibility (Do Not Expand)

**Your responsibility is:**

> Convert a completed user session into zero or more high-confidence frustration incidents, while minimizing false positives.

**Critical Rule:** If confidence is insufficient → emit nothing.

### What You Must NOT Build:
- ❌ AI or LLM usage
- ❌ Ticket creation
- ❌ Dashboards
- ❌ Replay logic
- ❌ Error monitoring replacement
- ❌ User feedback handling

**If unsure → do not include**

---

## 3. What This Module Receives (Strict)

### Input:
- **Completed session object**
  - Ordered list of normalized events
  - Session metadata (route, device, app version)

### Events May Include:
- User interactions
- System feedback (errors, latency)
- Navigation

### Critical Constraints:
- ❌ You must **never assume business meaning**
- ❌ You must **never inspect raw DOM or PII**

---

## 4. What This Module Outputs

### Output:
- **Frustration Incident** (structured object)
- **OR nothing**

### Each Emitted Incident MUST Include:
1. **Incident ID** (unique identifier)
2. **Session ID** (source session)
3. **Frustration score** (0–100)
4. **Confidence level** (Low / Medium / High)
5. **Triggering signals** (which signals fired)
6. **Primary failure point** (what went wrong)
7. **Timestamp** (when it occurred)
8. **Human-readable explanation** (why this is frustration)

### Must NOT Include:
- ❌ Recommendations
- ❌ AI language
- ❌ Opinions

---

## 5. Session Isolation Rules (Critical)

**Events from different sessions must NEVER interact**

### Rules:
1. Signals must occur within the same session
2. Signals must occur within a bounded time window
3. Session boundaries are trusted and final

**Violation of this rule = invalid implementation**

### Time Window Specification:
- **Default time window:** 30 seconds
- Signals must occur within this window to be considered correlated
- Time window starts from first signal detection
- All correlated signals must complete within the window
- If time window expires before correlation completes → no incident emitted

---

## 6. Allowed Frustration Signals (Only These)

### A. Rage Interaction Signal

**Triggered ONLY if:**
- Same interaction target
- Repeated rapidly
- No success feedback in between

**Exclusions:**
- Single repetition ≠ rage

**Example:**
- User clicks button 5 times in 2 seconds with no success → ✅
- User clicks button twice with success after first click → ❌

---

### B. Blocked Progress Signal

**Triggered ONLY if:**
- User attempts a meaningful action
- System rejects (error or validation)
- User retries same action

**Exclusions:**
- Ignore first failure

**Example:**
- User submits form → validation error → retries → ✅
- User submits form → validation error → gives up → ❌ (needs abandonment signal)

---

### C. Abandonment Signal

**Triggered ONLY if:**
- User begins a flow
- Encounters friction (error, delay)
- Leaves before completion

**Exclusions:**
- Leaving alone is NOT frustration

**Example:**
- User starts checkout → encounters error → navigates away → ✅
- User views page → navigates away (no friction) → ❌

---

### D. Confusion Signal

**Triggered ONLY if:**
- Back-and-forth navigation
- Excessive scrolling or hovering
- No forward progress

**Exclusions:**
- Confusion alone is LOW severity

**Example:**
- User navigates: A → B → A → B → A (no progress) → ✅
- User scrolls up and down 20 times on same page → ✅
- User hovers over multiple elements without clicking → ✅

---

## 7. Correlation Rules (False Alarm Prevention)

**You MUST NOT emit an incident unless ALL are true:**

1. ✅ At least two independent signals
2. ✅ At least one system feedback signal
3. ✅ Signals occurred within the same session
4. ✅ Signals occurred within the time window

### Valid Examples:
- ✅ Rage clicks + API error → **VALID**
- ✅ Abandonment + validation error → **VALID**
- ✅ Confusion + slow page load + rage clicks → **VALID**

### Invalid Examples:
- ❌ Single error → **INVALID** (only one signal)
- ❌ Slow page load only → **INVALID** (no system feedback)
- ❌ Rage clicks alone → **INVALID** (no system feedback)
- ❌ Confusion alone → **INVALID** (needs correlation)

---

## 8. Scoring & Confidence Rules

### Each Session Must Compute:
- **Frustration score** (0–100)
- **Confidence level** (Low / Medium / High)

### Emission Rules:
- ✅ **Only High confidence incidents are emitted**
- ❌ Medium / Low confidence are discarded or stored internally

### Constraint:
- **No ML. Deterministic logic only.**

### Scoring Guidelines:
- **High Confidence (80-100):** Multiple correlated signals + system feedback
- **Medium Confidence (50-79):** Two signals + system feedback, but some ambiguity
- **Low Confidence (0-49):** Single signal or weak correlation

---

## 9. Explainability Requirement (Mandatory)

**Every emitted incident MUST clearly explain:**

> "Which signals occurred and why this crossed the threshold."

### Example Good Explanation:
> "User experienced rage clicks (5 clicks in 2 seconds) on submit button, followed by API error (500 status), then abandoned the checkout flow. Multiple correlated frustration signals indicate high-confidence incident."

### Example Bad Explanation:
> "User seems frustrated." ❌

**If explanation is unclear → discard incident.**

---

## 10. Configuration Rules

### Default Behavior:
- **Must work without configuration**

### Optional Overrides:
- Ignore routes
- Adjust thresholds
- Suppress admin/internal paths

**Configuration is OPTIONAL, never required.**

---

## 11. What You Must NOT Build

- ❌ AI or LLM usage
- ❌ Ticket creation
- ❌ Dashboards
- ❌ Replay logic
- ❌ Error monitoring replacement
- ❌ User feedback handling

**If unsure → do not include**

---

## 12. Acceptance Criteria (Must Pass)

This module is considered complete only if:

1. ✅ **No single event causes escalation**
   - Single errors don't trigger incidents
   - Requires correlation

2. ✅ **False positives are <5% in test data**
   - High precision over recall
   - Better to miss than to false alarm

3. ✅ **Every incident is explainable**
   - Clear signal correlation
   - Human-readable explanation

4. ✅ **Works out-of-the-box**
   - No configuration required
   - Sensible defaults

5. ✅ **Emits nothing when uncertain**
   - Silence is success
   - Trust > coverage

---

## 13. Design Philosophy (Follow Exactly)

1. **Conservative > aggressive**
   - Better to miss a signal than create false alarm
   - High threshold for emission

2. **Silence > noise**
   - If uncertain, emit nothing
   - Fewer, higher-quality incidents

3. **Explainable > clever**
   - Simple logic beats complex algorithms
   - Every decision must be traceable

4. **Trust > coverage**
   - User trust is paramount
   - False positives destroy trust

> **This module determines whether the product is trusted or ignored.**

---

## 14. Implementation Checklist

### Signal Detection
- [ ] Rage Interaction Signal implemented
- [ ] Blocked Progress Signal implemented
- [ ] Abandonment Signal implemented
- [ ] Confusion Signal implemented

### Correlation
- [ ] Requires at least 2 independent signals
- [ ] Requires at least 1 system feedback signal
- [ ] Signals must be within same session
- [ ] Signals must be within time window

### Scoring
- [ ] Frustration score calculation (0-100)
- [ ] Confidence level calculation (Low/Medium/High)
- [ ] Only High confidence incidents emitted

### Explainability
- [ ] Every incident has clear explanation
- [ ] Explanation lists triggering signals
- [ ] Explanation explains why threshold crossed

### Session Isolation
- [ ] No cross-session correlation
- [ ] Session boundaries respected
- [ ] Time windows enforced

### Testing
- [ ] No single event causes escalation
- [ ] False positives <5% in test data
- [ ] All incidents explainable
- [ ] Works without configuration

---

## FINAL NOTE

**If a feature:**
- Is not explicitly described
- Increases cleverness
- Reduces explainability
- Adds assumptions

**DO NOT IMPLEMENT IT.**

---

## Quick Reference Summary

### Core Rules:
1. ✅ **Input:** Completed session with normalized events
2. ✅ **Output:** High-confidence incidents OR nothing
3. ✅ **Signals:** 4 types (Rage, Blocked Progress, Abandonment, Confusion)
4. ✅ **Correlation:** Requires 2+ signals + system feedback
5. ✅ **Time Window:** 30 seconds (default)
6. ✅ **Confidence:** Only High confidence incidents emitted
7. ✅ **Explainability:** Every incident must have clear explanation

### Critical Constraints:
- ❌ No cross-session correlation
- ❌ No single-event escalations
- ❌ No ML/AI in this module
- ❌ No assumptions about business meaning
- ❌ No inspection of raw DOM or PII

### Design Philosophy:
- Conservative > aggressive
- Silence > noise
- Explainable > clever
- Trust > coverage

---

## Example Incident Output

```json
{
  "incidentId": "inc_abc123",
  "sessionId": "sess_xyz789",
  "frustrationScore": 85,
  "confidenceLevel": "High",
  "triggeringSignals": [
    "Rage Interaction Signal",
    "Blocked Progress Signal",
    "System Feedback Signal (API Error)"
  ],
  "primaryFailurePoint": "Checkout submission endpoint",
  "timestamp": "2024-01-15T10:30:45Z",
  "explanation": "User experienced rage clicks (5 clicks in 2 seconds) on checkout submit button, followed by API error (500 status), then retried the same action. Multiple correlated frustration signals within 10-second window indicate high-confidence frustration incident."
}
```
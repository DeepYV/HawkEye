# Frontend Observer SDK - Implementation Instructions

## 1. Your Role

You are building **ONE frontend module:**

**Frontend Observer SDK**

This SDK runs inside customer **React / Next.js applications**.

### Its Only Job:
> Observe and emit normalized, privacy-safe user interaction events

### It Does NOT:
- ❌ Analyze
- ❌ Decide
- ❌ Score
- ❌ Judge

---

## 2. Core Objective (Do Not Deviate)

The SDK must:

1. ✅ **Be extremely easy to integrate**
2. ✅ **Have near-zero performance impact**
3. ✅ **Be safe by default**
4. ✅ **Emit clean, normalized events**
5. ✅ **Never produce false signals itself**

**Critical Rule:** If unsure → do not capture.

---

## 3. Strict Scope Boundaries

### You MUST Build:
- ✅ SDK initialization
- ✅ Event observation
- ✅ Event normalization
- ✅ Event batching & sending
- ✅ Privacy & masking safeguards

### You MUST NOT Build:
- ❌ Session replay / video
- ❌ Frustration detection
- ❌ AI logic
- ❌ Ticket creation
- ❌ Dashboards or UI
- ❌ Business logic interpretation

**If you do any of the above → it is wrong.**

---

## 4. Integration Experience (Critical)

Integration must be:

- **One import**
- **One init call**
- **No configuration required**

### Example Experience (conceptual):
> "Install, initialize, forget."

**Optional configuration is allowed but never required.**

### Example Integration:
```javascript
// Simple, one-line integration
import { initObserver } from '@your-org/observer-sdk';

initObserver({ apiKey: 'your-api-key' });
```

That's it. No configuration needed.

---

## 5. SDK Initialization Responsibilities

On initialization, SDK must:

1. ✅ Validate API key exists
2. ✅ Generate a session identifier
3. ✅ Capture environment metadata
4. ✅ Register observers safely
5. ✅ Fail silently if misconfigured

**Critical Rule:** SDK must never crash the host app.

### Initialization Flow:
1. Validate API key format
2. Generate unique session ID (UUID v4)
3. Capture: browser, device, app version
4. Register event listeners (non-blocking)
5. If any step fails → fail silently, do not initialize

---

## 6. What the SDK Observes

### A. User Interaction Events (Allowed)

**Events:**
- Clicks
- Scroll start / end
- Input focus / blur
- Form submit attempts
- Navigation / route change

**Rules:**
- ✅ Capture intent, not content
- ❌ Never capture input values
- ❌ Never infer meaning

**Example:**
- ✅ Capture: "User clicked submit button on checkout form"
- ❌ Do NOT capture: "User entered credit card number 1234..."

---

### B. System Feedback Events (Allowed)

**Events:**
- JS runtime errors
- Network request failure / success (status only)
- Long task / slow response indicators
- Loading start / end markers

**Rules:**
- ✅ Metadata only
- ❌ No payloads
- ❌ No headers
- ❌ No response bodies

**Example:**
- ✅ Capture: "API request to /api/checkout failed with status 500"
- ❌ Do NOT capture: Request body, response body, headers

---

### C. Context Events (Allowed)

**Events:**
- Route / URL path
- Component identifier (if explicitly provided)
- App version
- Device & browser info

**Example:**
- ✅ Capture: "User navigated to /checkout"
- ✅ Capture: "App version: 1.2.3"
- ❌ Do NOT capture: Query parameters with PII

---

## 7. Privacy Rules (Non-Negotiable)

### MUST:
- ✅ Mask all input values
- ✅ Ignore password, OTP, credit card fields
- ✅ Ignore explicitly marked sensitive elements
- ✅ Avoid capturing inner text by default

### MUST NOT:
- ❌ Record keystrokes
- ❌ Capture PII
- ❌ Capture DOM snapshots
- ❌ Capture screenshots

**Rule:** If in doubt → drop the event.

### Automatic Field Masking:
The SDK must automatically detect and ignore:
- `input[type="password"]`
- `input[type="email"]` (configurable)
- Elements with `data-sensitive` attribute
- Elements matching sensitive selectors (configurable)

---

## 8. Event Normalization Rules

Each event emitted must be:

- ✅ **Flat** - No nested objects
- ✅ **Structured** - Consistent schema
- ✅ **Predictable** - Same input → same output
- ✅ **Deterministic** - No randomness

### Every Event Must Include:
1. **Event type** (string, enum)
2. **Timestamp** (ISO 8601)
3. **Session ID** (UUID)
4. **Route** (sanitized URL path)
5. **Target identifier** (hashed or abstracted)
6. **Metadata** (minimal, flat object)

### Must NOT Include:
- ❌ Raw DOM nodes
- ❌ Complex objects
- ❌ Circular references

### Example Normalized Event:
```json
{
  "eventType": "click",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "sessionId": "sess_abc123",
  "route": "/checkout",
  "target": {
    "type": "button",
    "id": "hashed_element_id",
    "selector": "button[data-testid='submit']"
  },
  "metadata": {
    "coordinates": { "x": 100, "y": 200 }
  }
}
```

---

## 9. Performance Constraints

The SDK must:

- ✅ Be non-blocking
- ✅ Use async observers
- ✅ Batch events
- ✅ Throttle high-frequency events
- ✅ Yield to browser idle time when possible

**Hard Requirement:** SDK must never noticeably impact user experience.

### Performance Targets:
- **Bundle size:** <10KB gzipped
- **Initialization:** <50ms
- **Event capture overhead:** <1ms per event
- **Memory footprint:** <5MB

### Optimization Strategies:
- Use `requestIdleCallback` for non-critical operations
- Debounce/throttle scroll and mousemove events
- Batch network requests (default: 10 events or 5 seconds)
- Use Web Workers for heavy processing (if needed)

---

## 10. Event Transport Rules

**Characteristics:**
- Events are buffered
- Sent in batches
- Sent over HTTP
- Retries are allowed
- Failures are silent

### If Network Fails:
- ✅ Drop events gracefully
- ✅ Do not block UI
- ❌ Do not log noisy errors

### Batching Strategy:
- **Default batch size:** 10 events
- **Default batch timeout:** 5 seconds
- **Max batch size:** 50 events
- **Retry strategy:** Exponential backoff (max 3 retries)

---

## 11. Session Handling (Frontend Side)

The SDK:

- ✅ Generates a session ID
- ✅ Resets session on page reload

**Must NOT:**
- ❌ Define session end logic

**Note:** Session lifecycle decisions belong to backend.

### Session ID Generation:
- Generate on SDK initialization
- Regenerate on page reload/navigation
- Format: UUID v4
- Store in memory (not localStorage/cookies for privacy)

---

## 12. Configuration (Optional Only)

### Allowed Optional Config:
- Disable specific event types
- Ignore specific routes
- Mark custom sensitive selectors
- Enable debug mode (console logs)

### Configuration Must:
- ✅ Be optional
- ✅ Have safe defaults
- ❌ Never be required

### Example Configuration:
```javascript
initObserver({
  apiKey: 'your-api-key',
  // All optional:
  disabledEvents: ['scroll'], // Optional
  ignoredRoutes: ['/admin'], // Optional
  sensitiveSelectors: ['.sensitive'], // Optional
  debug: false // Optional, default: false
});
```

---

## 13. Debug & Explainability (Developer Trust)

When debug mode is enabled:

- ✅ Log what events are captured
- ✅ Log why events are ignored
- ❌ Never expose sensitive data

**Purpose:** This is for developer trust, not end users.

### Debug Output Example:
```
[Observer SDK] Event captured: click on button#submit
[Observer SDK] Event ignored: input value in password field
[Observer SDK] Batch sent: 5 events
```

---

## 14. What You Must Explicitly Avoid

- ❌ Heavy dependencies
- ❌ Build-time plugins
- ❌ Framework-specific coupling
- ❌ Magic heuristics
- ❌ Silent breaking behavior

**Principle:** The SDK must feel boring, predictable, and safe.

### Dependency Policy:
- **Zero dependencies** preferred
- If dependency needed: <5KB, well-maintained, no breaking changes
- No polyfills unless absolutely necessary

---

## 15. Acceptance Criteria (Must Pass)

This module is complete only if:

1. ✅ **Can be integrated in <5 minutes**
   - One import, one init call
   - No configuration required

2. ✅ **Does not break host apps**
   - No crashes, no errors
   - Graceful degradation

3. ✅ **Captures only allowed events**
   - No PII, no sensitive data
   - Only normalized events

4. ✅ **Emits normalized data**
   - Consistent schema
   - Predictable format

5. ✅ **Respects privacy strictly**
   - Automatic field masking
   - No data leakage

6. ✅ **Does nothing when uncertain**
   - Fail silently
   - Drop events when in doubt

---

## 16. Design Philosophy (Follow Exactly)

1. **Observe, don't interpret**
   - Capture what happened, not why
   - No assumptions about user intent

2. **Safe > complete**
   - Better to miss an event than capture sensitive data
   - Privacy first, coverage second

3. **Invisible > powerful**
   - Users should never notice the SDK
   - Zero performance impact

4. **Predictable > clever**
   - Simple, straightforward logic
   - No magic, no heuristics

> **This SDK is the foundation of trust.**

---

## Implementation Checklist

### Core Functionality
- [ ] SDK initialization with API key validation
- [ ] Session ID generation
- [ ] Event observation (clicks, scroll, input, navigation)
- [ ] System feedback capture (errors, network, performance)
- [ ] Event normalization (flat, structured, predictable)
- [ ] Event batching and transport

### Privacy & Security
- [ ] Automatic field masking (password, OTP, credit card)
- [ ] PII detection and exclusion
- [ ] URL sanitization
- [ ] Error message sanitization
- [ ] No DOM snapshots or screenshots

### Performance
- [ ] Non-blocking event capture
- [ ] Event batching (10 events or 5 seconds)
- [ ] Throttling for high-frequency events
- [ ] Bundle size <10KB gzipped
- [ ] Initialization <50ms

### Integration
- [ ] One import, one init call
- [ ] No configuration required
- [ ] Optional configuration support
- [ ] Framework-agnostic core
- [ ] React and Next.js support

### Error Handling
- [ ] Fail silently on misconfiguration
- [ ] Never crash host app
- [ ] Graceful network failure handling
- [ ] No noisy error logs

### Testing
- [ ] Integration in <5 minutes
- [ ] No host app breakage
- [ ] Privacy compliance verified
- [ ] Performance benchmarks met
- [ ] Works without configuration

---

## FINAL INSTRUCTION

**If a feature:**
- Adds intelligence
- Makes assumptions
- Interprets behavior
- Risks privacy
- Increases bundle size unnecessarily

**DO NOT IMPLEMENT IT. Build only what is described here.**

---

## Example Event Schema

### Click Event:
```json
{
  "eventType": "click",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "sessionId": "sess_abc123",
  "route": "/checkout",
  "target": {
    "type": "button",
    "id": "hashed_id",
    "selector": "button[data-testid='submit']"
  },
  "metadata": {
    "coordinates": { "x": 100, "y": 200 }
  }
}
```

### Error Event:
```json
{
  "eventType": "error",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "sessionId": "sess_abc123",
  "route": "/checkout",
  "target": {
    "type": "javascript_error"
  },
  "metadata": {
    "errorType": "TypeError",
    "message": "Sanitized error message",
    "stack": "Sanitized stack trace"
  }
}
```

### Network Event:
```json
{
  "eventType": "network",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "sessionId": "sess_abc123",
  "route": "/checkout",
  "target": {
    "type": "api_request",
    "url": "/api/checkout"
  },
  "metadata": {
    "status": 500,
    "duration": 1234,
    "method": "POST"
  }
}
```
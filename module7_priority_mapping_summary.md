# Confidence → Priority Mapping Implementation

**Author:** John Smith (Solution Architect)  
**Date:** 2024-01-15  
**Status:** ✅ Implemented - Strict Policy Enforcement

---

## Implementation Summary

The confidence → priority mapping system has been implemented following the authoritative policy. This ensures predictable, boring (good), and trustworthy ticket prioritization.

---

## Core Principle (Non-Negotiable)

✅ **Confidence determines whether we create a ticket**  
✅ **Severity determines how urgent it is**  
✅ **Priority is never based on confidence alone**

---

## Confidence Bands (System-Level)

| Confidence Score | Meaning | System Action |
|-----------------|---------|---------------|
| < 0.50 | Weak signal | Ignore / shadow only |
| 0.50 – 0.69 | Possible issue | Observe, no ticket |
| 0.70 – 0.79 | Likely issue | ✅ Eligible for export |
| 0.80 – 0.89 | Very likely | ✅ Export normally |
| ≥ 0.90 | Certain | ✅ Export immediately |

**Rule:** ✅ No ticket is created below 0.70, ever.

---

## Priority Mapping Table (Implemented)

| Confidence | Severity | Ticket Priority |
|------------|----------|-----------------|
| 0.70–0.79 | Low | P4 (Backlog) |
| 0.70–0.79 | Medium | P3 |
| 0.70–0.79 | High | P2 |
| 0.70–0.79 | Critical | P2 |
| 0.80–0.89 | Low | P3 |
| 0.80–0.89 | Medium | P2 |
| 0.80–0.89 | High | P1 |
| 0.80–0.89 | Critical | P1 |
| ≥ 0.90 | Low | P3 |
| ≥ 0.90 | Medium | P2 |
| ≥ 0.90 | High | P1 |
| ≥ 0.90 | Critical | P0 |

---

## Jira / Linear Mapping (Implemented)

### Linear
- P0 → 4 (Urgent)
- P1 → 3 (High)
- P2 → 2 (Medium)
- P3 → 1 (Low)
- P4 → 0 (Backlog)

### Jira
- P0 → Blocker
- P1 → Highest
- P2 → High
- P3 → Medium
- P4 → Low

---

## Files Modified

### 1. `internal/exporter/eligibility.go`
- ✅ Enforces minimum confidence threshold (0.70)
- ✅ Configurable export threshold (default 0.70)
- ✅ No tickets below 0.70, ever

### 2. `internal/exporter/priority.go` (NEW)
- ✅ PriorityMapper implementation
- ✅ Maps confidence + severity to priority
- ✅ Jira and Linear priority mapping
- ✅ Deterministic mapping logic

### 3. `internal/exporter/engine.go`
- ✅ Integrates PriorityMapper
- ✅ Maps priority before formatting ticket
- ✅ Passes priority to formatter

### 4. `internal/exporter/formatter.go`
- ✅ Accepts priority parameter
- ✅ Includes priority in metadata
- ✅ Confidence transparency (always shown)

### 5. `internal/adapters/jira.go`
- ✅ Uses priority from ticket metadata
- ✅ Maps P0-P4 to Jira priorities
- ✅ Removed old severity-based mapping

### 6. `internal/adapters/linear.go`
- ✅ Uses priority from ticket metadata
- ✅ Maps P0-P4 to Linear priorities (0-4 scale)
- ✅ Removed old severity-based mapping

---

## Example Walkthroughs

### Example 1: High Confidence + Critical
**Incident:**
- Confidence: 0.91
- Severity: Critical
- Route: /checkout

**Result:**
- ✅ Exported
- Priority: P0 (Urgent)
- Jira: Blocker
- Linear: 4 (Urgent)

### Example 2: Medium Confidence + Medium Severity
**Incident:**
- Confidence: 0.74
- Severity: Medium
- Route: /settings

**Result:**
- ✅ Exported
- Priority: P3
- Jira: Medium
- Linear: 1 (Low)

### Example 3: Low Confidence (Not Exported)
**Incident:**
- Confidence: 0.65
- Severity: High
- Route: /dashboard

**Result:**
- ❌ Not exported (confidence < 0.70)
- Reason: "confidence_below_minimum"

---

## Export Safeguards (Anti-Noise)

✅ Export is blocked if:
- Incident is suppressed
- Export rate limit exceeded
- Confidence < 0.70 (system rule)
- Already exported

**Confidence ≠ permission to spam.**

---

## Transparency Rule

✅ Every exported ticket includes:
```
## Confidence
0.87 — consistent behavior observed across many sessions with identical interaction patterns
```

This builds trust.

---

## Why This Works

✅ Engineers are not overwhelmed  
✅ PMs trust the prioritization  
✅ High confidence ≠ always urgent  
✅ Low confidence never leaks out  
✅ Predictable and boring (that's good)

---

## Key Design Decisions

1. **Minimum Confidence Threshold:** 0.70 (non-negotiable)
2. **Priority Based on Both:** Confidence + Severity
3. **Conservative Defaults:** P4 for low confidence + low severity
4. **Transparency:** Confidence always shown in tickets
5. **Configurable:** Export threshold can be adjusted per customer

---

## Testing Recommendations

1. ✅ Test confidence < 0.70 → not exported
2. ✅ Test confidence 0.70-0.79 → P3/P4 priority
3. ✅ Test confidence 0.80-0.89 → P1/P2 priority
4. ✅ Test confidence ≥ 0.90 → P0/P1 priority
5. ✅ Test severity impact on priority
6. ✅ Test Jira priority mapping
7. ✅ Test Linear priority mapping
8. ✅ Test confidence transparency in tickets

---

**Status:** ✅ **IMPLEMENTED AND READY FOR TESTING**

**Final Rule:** The system earns trust by being conservative first, smart second.
# Module Architecture

## Architecture Principles (Non-Negotiable)

1. **Separation of concerns** - Each module has a single, well-defined responsibility
2. **Conservative defaults** - Default behavior prioritizes safety over coverage
3. **No single module is "smart" alone** - Intelligence emerges from module interaction
4. **Explainability over intelligence** - Clear explanations beat clever algorithms
5. **Silence is success when uncertain** - Better to emit nothing than false positives

---

## Core Modules

### Module 1: Frontend Observer SDK

**Purpose:** Capture normalized, privacy-safe user behavior

**Location:** Lives in customer app

**Characteristics:**
- Passive, lightweight
- No decisions made here

**Output:** Emits normalized events only

**Constraints:**
- No business logic
- No analysis
- Privacy-safe (field masking)

**See:** [Frontend SDK Detailed Instructions](./FRONTEND_SDK_INSTRUCTIONS.md)

---

### Module 2: Event Ingestion API

**Purpose:** Receive, validate, and store events

**Responsibilities:**
- Authentication (API key)
- Validation
- Rate limiting
- No analysis

**Output:** Forwards events to session manager

**Constraints:**
- No business logic
- No signal detection
- Pure validation and forwarding

---

### Module 3: Session Manager

**Purpose:** Define session boundaries correctly

**Responsibilities:**
- Groups events into sessions
- Enforces time & context isolation
- Prevents cross-session correlation

**Output:** Emits complete sessions downstream

**Constraints:**
- No signal detection
- No frustration analysis
- Pure session boundary management

**Critical Rules:**
- Session boundaries are trusted and final
- No cross-session correlation allowed

---

### Module 4: User Frustration Signal Engine (UFSE)

**Purpose:** Detect frustration & confusion

**Type:** CORE INTELLECTUAL PROPERTY

**Responsibilities:**
- Detects frustration & confusion
- Correlates signals
- Applies strict thresholds
- Emits only high-confidence incidents

**Output:** Frustration Incidents (high-confidence only)

**Input:**
- Completed session object
- Ordered list of normalized events
- Session metadata (route, device, app version)

**Constraints:**
- No UI knowledge
- No AI knowledge
- No Jira/Linear knowledge
- No SDK internals knowledge
- Only knows about sessions and events

**Emission Rules:**
- Only High confidence incidents are emitted
- Medium / Low confidence are discarded
- If confidence is insufficient → emit nothing

**See:** [UFSE Detailed Instructions](./UFSE_INSTRUCTIONS.md)

---

### Module 5: Incident Store

**Purpose:** Source of truth for validated incidents

**Responsibilities:**
- Stores incidents
- Enables deduplication
- Downstream consumers read from here

**Output:** Stored incidents (read-only for consumers)

**Constraints:**
- No signal detection
- No analysis
- Pure storage and retrieval

---

### Module 6: AI Interpretation Module

**Purpose:** Convert incidents into human-readable summaries

**Stage:** Later

**Characteristics:**
- Async
- Never in hot path
- Optional

**Constraints:**
- Post-incident only
- Never blocks core path
- Optional enhancement

---

### Module 7: Ticket Generator

**Purpose:** Create Jira / Linear tickets

**Responsibilities:**
- Consumes only high-confidence incidents
- Deduplicates
- Applies workflow rules

**Input:** High-confidence incidents from Incident Store

**Output:** Tickets in external systems

**Constraints:**
- No signal detection
- No analysis
- Pure integration logic

---

## Module Interaction Flow

```
Frontend Observer SDK
        │
        ▼
Event Ingestion API
        │
        ▼
Session Manager
        │
        ▼
User Frustration Signal Engine (UFSE)
        │
        ▼
Incident Store
        │
        ├──► AI Interpretation Module (async, later)
        │
        └──► Ticket Generator
```

## Data Formats

This module architecture is defined in:
- **`modules.json`** - JSON representation
- **`ufse_instructions.json`** - Detailed UFSE instructions

---

## Key Design Principles

### 1. Single Responsibility
Each module does ONE thing well. No module tries to be smart about everything.

### 2. Clear Boundaries
Modules communicate through well-defined interfaces. No shared state or hidden dependencies.

### 3. Conservative by Default
When in doubt, modules err on the side of caution. Better to miss a signal than create false alarms.

### 4. Explainability First
Every decision must be explainable. If you can't explain why something happened, it shouldn't happen.

### 5. Trust Through Silence
If a module is uncertain, it should emit nothing. Silence builds trust; false positives destroy it.
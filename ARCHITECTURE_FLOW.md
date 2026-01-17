# High-Level Architecture Flow

End-to-end system architecture showing the complete flow from customer application to ticket generation.

## Architecture Diagram

```
Customer App (React / Next.js)
        │
        ▼
Frontend Observer SDK
        │
        ▼
Event Ingestion API
        │
        ▼
Session Manager
        │
        ▼
Frustration Signal Engine (UFSE)
        │
        ▼
Incident Store
        │
        ├──► AI Interpretation Module (later)
        │
        └──► Ticket Generator (Jira / Linear)
```

## Component Details

### 1. Customer App
- **Type:** Application
- **Technologies:** React, Next.js
- **Role:** Customer application where users interact
- **Output:** User interactions and events

### 2. Frontend Observer SDK
- **Type:** SDK
- **Technologies:** TypeScript
- **Role:** Lightweight SDK that observes user interactions
- **Characteristics:**
  - <10KB gzipped
  - Browser-only
  - No dependencies
- **Output:** Batched events (JSON)

### 3. Event Ingestion API
- **Type:** API
- **Technologies:** Go, HTTP, JSON
- **Role:** Receives and validates events from SDK
- **Protocol:** HTTP (batched)
- **Format:** JSON
- **Output:** Validated events

### 4. Session Manager
- **Type:** Service
- **Technologies:** Go
- **Role:** Manages user sessions and aggregates events
- **Responsibilities:**
  - Session tracking
  - Event aggregation
  - Time-window management
- **Output:** Session context, aggregated events

### 5. Frustration Signal Engine (UFSE)
- **Type:** Engine
- **Technologies:** Go
- **Role:** Core engine that detects frustration signals
- **Pattern:** Stream → Window → Evaluate → Emit
- **Characteristics:**
  - Deterministic rule engine
  - Time-window evaluation
  - In-process Go workers
- **Output:** Detected incidents, signals

### 6. Incident Store
- **Type:** Storage
- **Technologies:** PostgreSQL, ClickHouse
- **Role:** Stores detected incidents and related data
- **Storage:**
  - Metadata: PostgreSQL
  - Events: ClickHouse
- **Downstream:**
  - AI Interpretation Module
  - Ticket Generator

### 7. AI Interpretation Module
- **Type:** Module
- **Technologies:** OpenAI, Anthropic
- **Role:** Interprets incidents using AI (async, post-incident)
- **Stage:** Later
- **Characteristics:**
  - Async processing
  - Post-incident only
  - Never in hot path
- **Parent:** Incident Store

### 8. Ticket Generator
- **Type:** Integration
- **Technologies:** REST API, Webhook
- **Role:** Generates tickets in external systems
- **Targets:** Jira, Linear
- **Priority:** Linear first
- **Parent:** Incident Store

## Data Flow

### Stage 1: Customer App → Frontend Observer SDK
- **Data:** User interactions, events
- **Direction:** Unidirectional

### Stage 2: Frontend Observer SDK → Event Ingestion API
- **Data:** Batched events (JSON)
- **Protocol:** HTTP (batched)
- **Compression:** Gzip

### Stage 3: Event Ingestion API → Session Manager
- **Data:** Validated events
- **Processing:** Validation, normalization

### Stage 4: Session Manager → Frustration Signal Engine
- **Data:** Session context, aggregated events
- **Processing:** Session aggregation, time-window management

### Stage 5: Frustration Signal Engine → Incident Store
- **Data:** Detected incidents, signals
- **Processing:** Rule evaluation, signal detection

### Stage 6: Incident Store → AI Interpretation Module
- **Data:** Incident data (async)
- **Processing:** Async AI interpretation
- **Note:** Out of hot path, post-incident only

### Stage 7: Incident Store → Ticket Generator
- **Data:** Incident data, metadata
- **Processing:** Ticket creation in external systems

## Key Principles

1. **Unidirectional Data Flow**
   - Data flows in one direction through the system
   - No circular dependencies
   - Clear separation of concerns

2. **Deterministic Core Path**
   - Core processing path (stages 1-5) is deterministic
   - No ML/AI in hot path
   - Predictable behavior

3. **AI Processing is Async and Out of Hot Path**
   - AI Interpretation Module processes asynchronously
   - Never blocks core incident detection
   - Post-incident processing only

4. **Separation of Concerns**
   - Each component has a single, well-defined responsibility
   - Clear interfaces between components
   - Easy to test and maintain

## Architecture Characteristics

### Hot Path (Deterministic)
- Customer App → Frontend Observer SDK → Event Ingestion API → Session Manager → Frustration Signal Engine → Incident Store

### Cold Path (Async)
- Incident Store → AI Interpretation Module (async, post-incident)

### Integration Path
- Incident Store → Ticket Generator (synchronous or async)

## Scalability Considerations

- **Frontend Observer SDK:** Stateless, scales with customer apps
- **Event Ingestion API:** Stateless, horizontal scaling
- **Session Manager:** Stateful, may require session affinity
- **Frustration Signal Engine:** Stateless processing, horizontal scaling
- **Incident Store:** Database-backed, vertical/horizontal scaling
- **AI Interpretation Module:** Async queue-based, independent scaling
- **Ticket Generator:** Stateless integration service, horizontal scaling

## Data Formats

This architecture flow is defined in:
- **`architecture_flow.json`** - JSON representation
- **`architecture_flow.yaml`** - YAML representation

These can be used for:
- Architecture visualization tools
- Documentation generation
- System design validation
- Team alignment
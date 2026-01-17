# Solution Architecture - Tech Stack

> **See also:** [High-Level Architecture Flow](./ARCHITECTURE_FLOW.md) for end-to-end system flow

## High-Level Principles

This tool must be:

1. **Invisible in customer apps** (low overhead)
2. **Deterministic** (not ML-heavy early)
3. **Explainable**
4. **Cheap to operate**
5. **Easy to integrate**
6. **Enterprise-ready later**

**Optimization Strategy:** Predictability > novelty

---

## Component Architecture

### A. Frontend SDK (Customer App)

**Language:** TypeScript

**Framework Support:**
- React
- Next.js (App Router + Pages)
- Framework-agnostic core (DOM-based)

**Key Characteristics:**
- <10KB gzipped
- No dependencies where possible
- Runs in browser only
- No build-time plugins required

**Rationale:**
- First-class TS dev experience
- Easy adoption
- Tree-shaking
- Good DX for OSS

---

### B. Event Transport Layer

**Protocol:**
- Primary: HTTP (batched)
- Future: WebSocket (optional)

**Format:** JSON

**Compression:** Gzip

**Rationale:**
- Reliable
- Easy debugging
- Firewall-friendly
- No infra surprises

**Avoid:**
- gRPC in browser
- Exotic protocols

---

### C. Backend Core (Frustration Engine)

**Language:** Go

**Why Go:**
- Deterministic behavior
- Low latency
- High throughput
- Easy concurrency
- Simple mental model
- Strong typing without complexity

---

### D. API Layer

**Framework:** net/http or chi

**Approach:** Minimal middleware

**Rationale:**
- Predictable
- Fast
- Easy to secure
- No magic

**Avoid:**
- Heavy frameworks
- Over-abstracted REST layers

---

### E. Data Storage

#### 1️⃣ Event Storage (Raw, Short-lived)

**Database:** ClickHouse

**Purpose:** Event streams

**Retention:** 7-30 days

**Rationale:**
- Excellent for event streams
- Cheap at scale
- Fast aggregations
- Perfect for session analysis

#### 2️⃣ Core Metadata & Config

**Database:** PostgreSQL

**Stores:**
- Projects
- API keys
- Threshold configs
- Incident records
- Ticket mappings

**Rationale:**
- Reliable
- ACID
- Familiar
- Easy migrations

---

### F. Session & Signal Processing

**Pattern:** Stream → Window → Evaluate → Emit

**Implementation:** In-process Go workers

**Evaluation:** Time-window evaluation

**Engine:** Deterministic rule engine

**Avoid Early:**
- Kafka
- Flink
- Spark

*You do not need them yet.*

---

### G. AI / LLM Layer (Later Stage)

**Model Access:**
- Providers: OpenAI / Anthropic
- Method: API-based

**Usage Pattern:**
- Async
- Post-incident only
- Never in the hot path

**Rationale:**
- Keeps core system deterministic
- Avoids latency & hallucination risk

---

### H. Integrations (Tickets)

**Approach:** Separate integration service

**Protocols:** Webhook + REST APIs

**Targets:**
- First: Linear (much cleaner than Jira)
- Future: Jira

---

### I. Infrastructure

**Cloud:** AWS or GCP

**Core Components:**
- Compute: Containers (ECS / GKE)
- Load Balancer
- Object storage (evidence blobs)
- Secrets manager

**Avoid:** Serverless for core processing (cold starts hurt trust)

---

### J. Observability (Dogfooding Matters)

**Metrics:**
- System: Prometheus
- Visualization: Grafana

**Logs:**
- Format: Structured logs
- Correlation: Correlation IDs per session

**Tracing:** OpenTelemetry

**Principle:** *If you can't trust your system, customers won't.*

---

## Security & Privacy

### Mandatory

- Field masking at SDK level
- Encryption in transit
- Encryption at rest
- Configurable data retention

### Future

- SOC2 readiness
- GDPR delete hooks
- Region-based data residency

---

## Data Formats

This architecture is defined in language-agnostic formats:

- **`tech_stack.json`** - JSON representation
- **`tech_stack.yaml`** - YAML representation
- **`tech_stack_schema.json`** - JSON Schema for validation

These can be used for:
- Documentation generation
- Configuration management
- Architecture validation
- Team alignment
- Tooling integration

---

## Architecture Decisions

### Why Not ML-Heavy Early?

Deterministic systems are:
- Predictable
- Debuggable
- Explainable
- Trustworthy

ML can be added later for enhancement, not core functionality.

### Why Go for Backend?

Go provides the perfect balance:
- Performance of compiled languages
- Simplicity of interpreted languages
- Excellent concurrency primitives
- Strong typing without complexity
- Fast compilation and deployment

### Why ClickHouse + PostgreSQL?

- **ClickHouse:** Optimized for time-series/event data, cheap at scale
- **PostgreSQL:** Reliable ACID transactions for critical metadata

### Why Avoid Serverless for Core?

Cold starts introduce unpredictable latency, which hurts user trust in a monitoring/frustration detection system.

---

## Implementation Phases

### Phase 1: Core (Deterministic)
- Frontend SDK (TypeScript)
- Backend (Go)
- Event transport (HTTP/JSON)
- Basic storage (ClickHouse + PostgreSQL)
- Rule-based processing

### Phase 2: Integration
- Ticket system integrations (Linear first)
- Enhanced observability
- Security hardening

### Phase 3: Enhancement
- AI/LLM layer (async, post-incident)
- Advanced analytics
- Enterprise features

---

## References

See `tech_stack.json` or `tech_stack.yaml` for the complete structured definition.
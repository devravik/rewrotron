# DATA-FLOW.md — End-to-End Pipeline & Event Streaming

> **Version:** 1.0  
> **Status:** Architecture Reference Guide  

---

## 1. End-to-End Information Lifecycle

Data flows sequentially through discrete stages of crystallization—from raw, noisy legacy code down to verified, idiomatic target code:

```text
[Legacy PHP Source Code] + [Live HTTP Requests / Seeds]
                     │
                     ▼
             1. RECONNAISSANCE
      • Static AST Parsing → Source Symbols, Calls, Queries
      • Runtime Observation → Correlated Execution Traces
                     │
                     ▼
             2. CRYSTALLIZATION
      • Reachability Graph Builder → Classify (ACTIVE, DORMANT, DEAD)
      • Behavior Synthesizer → BehaviorCases & BehaviorContracts
                     │
                     ▼
             3. PLANNING & SLICING
      • Migration Planner → Dependency-Ordered MigrationUnits
      • Architecture Mapper → Route & Model Clean Architecture Specs
                     │
                     ▼
             4. RECONSTRUCTION
      • Target Adapter (Laravel 13) → Idiomatic Code & Configurations
      • Clean Target Code committed to Target Project Workspace
                     │
                     ▼
             5. DIFFERENTIAL VERIFICATION
      • Replay Engine executes identical inputs against Source & Target
      • State Normalizer strips dynamic tokens (CSRF, Nonce, Time)
      • Comparator checks Status, Headers, Body, DB Mutations, Session
                     │
         ┌───────────┴───────────┐
         │                       │
      [PASS]                 [MISMATCH]
         │                       │
         ▼                       ▼
6. REPETITION & REGRESSION   7. DIAGNOSIS & REPAIR
   • Freeze Golden Master       • LLM Mismatch Analysis
   • Run next MigrationUnit     • Patch Target Implementation
                                • Re-run Verification (Loop)
```

---

## 2. Event Model & Audit Stream

Every significant state transition across the pipeline emits an append-only event persisted into `project.db` (and logged to `.rewrotron/logs/events.jsonl`):

```json
{
  "event_id": "evt_01HZX89AB1234567890",
  "project_id": "customer-portal",
  "timestamp": "2026-10-03T04:30:00.000Z",
  "event_type": "VERIFICATION_MISMATCH_DETECTED",
  "payload": {
    "case_id": "order-create-001",
    "unit_id": "orders-slice",
    "mismatch_type": "DATABASE_MUTATION_MISMATCH",
    "details": {
      "table": "orders",
      "expected_column": "created_at",
      "difference": "Target stored UTC ISO8601; Source stored Unix timestamp"
    }
  }
}
```

This guarantees that:
- Every action taken by autonomous agents is auditable.
- Any failed experiment can be rolled back to its exact prior state.
- Migration progress and quality metrics are computed deterministically from real events.

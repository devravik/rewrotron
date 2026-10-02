# PROJECT-LIFECYCLE.md — Project State Transitions & Mandatory Confirmation Gates

> **Version:** 1.0  
> **Status:** Domain State Machine Reference  

---

A Rewrotron project progresses through a deterministic finite state machine (FSM). Each state transition is validated and recorded in `.rewrotron/project.db`.

Crucially, **Gates 1, 2, and 3 are enforced by the engine**. Any attempt by an autonomous agent or command to bypass an unconfirmed gate results in a strict `ERROR: USER_CONFIRMATION_REQUIRED`.

## 1. Project Finite State Machine & Confirmation Gates

```text
 ┌─────────────────┐
 │   INITIALIZED   │  (rewrotron init / rewrotron detect executed)
 └────────┬────────┘
          │ (Environment, framework, runtime & route detection)
          ▼
 ┌─────────────────┐
 │    DETECTED     │  (Machine detection complete with confidence scores)
 └────────┬────────┘
          │
          │ ═════════════════════════════════════════════════════════
          │ GATE 1: SOURCE DETECTION CONFIRMATION
          │ Command: rewrotron confirm source (or MCP project.confirm_source)
          │ Persists: baseline facts, user overrides, configuration_hash
          │ ═════════════════════════════════════════════════════════
          ▼
 ┌─────────────────┐
 │ SOURCE_CONFIRMED│  (Baseline locked; deep analysis permitted)
 └────────┬────────┘
          │ (rewrotron analyze -> Source IR)
          ▼
 ┌─────────────────┐
 │    ANALYZED     │  (AST parsed, reachability graph built, code classified)
 └────────┬────────┘
          │ (rewrotron observe & discover -> Application IR & Behavior IR)
          ▼
 ┌─────────────────┐
 │    MODELED      │  (BehaviorContracts and BehaviorCases materialized)
 └────────┬────────┘
          │ (rewrotron plan -> proposes Target Stack, Architecture & Slices)
          ▼
 ┌─────────────────┐
 │ PLAN_SYNTHESIZED│  (Units topologically sorted; target architecture drafted)
 └────────┬────────┘
          │
          │ ═════════════════════════════════════════════════════════
          │ GATE 2: TARGET & MIGRATION PLAN CONFIRMATION
          │ Command: rewrotron confirm plan (or MCP project.confirm_plan)
          │ Persists: confirmed target language, framework, arch style
          │ ═════════════════════════════════════════════════════════
          ▼
 ┌─────────────────┐
 │  PLAN_CONFIRMED │  (Plan locked; code generation permitted)
 └────────┬────────┘
          │ (rewrotron migrate -> Target Adapter generates modern code)
          ▼
 ┌─────────────────┐
 │   MIGRATING     │  (Vertical slice code generation & verification loop)
 └────────┬────────┘
          │ (All units generated & verified with zero open defects)
          ▼
 ┌─────────────────┐
 │    VERIFIED     │  (Differential tests passing across 100% active units)
 └────────┬────────┘
          │ (rewrotron benchmark)
          ▼
 ┌─────────────────┐
 │   BENCHMARKED   │  (Performance non-regression confirmed)
 └────────┬────────┘
          │
          │ ═════════════════════════════════════════════════════════
          │ GATE 3: PRODUCTION CUTOVER CONFIRMATION
          │ Requires human sign-off; cannot be executed autonomously
          │ ═════════════════════════════════════════════════════════
          ▼
 ┌─────────────────┐
 │  READY_CUTOVER  │  (Target repository packaged for production deployment)
 └─────────────────┘
```

---

## 2. Invalidation & Drift Rules

- **`configuration_hash` Drift Invalidation:** If source files are modified on disk after `SOURCE_CONFIRMED`, the `configuration_hash` becomes stale. The project immediately transitions back to `DETECTED`, blocking further migration until the user re-confirms.
- **User Overrides Overrule Machine Inference:** If detection infers PHP 7.4 with 80% confidence, but the user specifies PHP 8.2, the confirmed value (`8.2`) is permanently recorded and passed to target adapters.

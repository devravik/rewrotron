# SYSTEM-DESIGN.md — Canonical System Architecture & Dependency Model

> **Status:** Authoritative Architecture Specification  
> **Subsystem:** Core Engine & Component Topology  
> **Language Baseline:** Go 1.22+  

---

## 1. Architectural Philosophy & Invariants

Rewrotron is architected around the following structural invariants:

1. **Complete Framework Neutrality in Core:** The Core Engine understands **applications**, not frameworks. It knows nothing about CodeIgniter, Laravel, Django, or Rails. All framework-specific syntax and conventions are encapsulated strictly in adapters.
2. **The Three-Tier Intermediate Representation (IR):**
   - **Source IR:** Syntax-level symbols, AST nodes, files, methods, raw queries, declared routes.
   - **Application IR:** Structural topology (controllers, services, data models, workflows, dependencies).
   - **Behavior IR:** Externally observable contracts (entrypoints, inputs, state transitions, business rules, side effects, responses).
3. **Pluggable Adapter Interfaces:** Source and target adapters implement standardized Go interfaces. The core interacts with adapters purely through these contracts—no `if framework == "codeigniter"` conditional branching in domain packages.
4. **Mandatory Gate Enforcement Engine:** Lifecycle progression is blocked by the engine until explicit user confirmation is recorded for Gate 1 (Source Detection), Gate 2 (Target & Migration Plan), and Gate 3 (Cutover).
5. **Acyclic Unidirectional Dependency Flow:** Dependencies only flow inward from external driving adapters toward domain services, and outward to abstracted infrastructure ports.

---

## 2. Global Component Topology

```text
┌────────────────────────┐  ┌────────────────────────┐  ┌────────────────────────┐
│     rewrotron CLI      │  │   rewrotron-mcp        │  │     Agent Skills       │
│    (cmd/rewrotron)     │  │  (cmd/rewrotron-mcp)   │  │   (skills/rewrotron)   │
└───────────┬────────────┘  └───────────┬────────────┘  └───────────┬────────────┘
            │                           │                           │
            └───────────────────────────┼───────────────────────────┘
                                        │ (Invokes CLI / MCP / Services)
                                        ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                           APPLICATION CORE (internal/)                         │
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                     Gate Enforcement Engine                              │  │
│  │  [ Gate 1: Detection ]   ➔   [ Gate 2: Plan ]   ➔   [ Gate 3: Cutover ]  │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌───────────────────────┐  ┌───────────────────────┐  ┌────────────────────┐  │
│  │   Detection Engine    │  │  Static Analyzer      │  │  Runtime Observer  │  │
│  │ (Detects Lang/Frame)  │  │ (Emits Source IR)     │  │ (Execution Traces) │  │
│  └───────────┬───────────┘  └───────────┬───────────┘  └─────────┬──────────┘  │
│              │                          │                        │             │
│              └──────────────────────────┼────────────────────────┘             │
│                                         ▼                                      │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                        Behavior Engine (internal/behavior)               │  │
│  │  • Source IR  ➔  Application IR  ➔  Behavior IR                          │  │
│  │  • Canonical Behavioral Contracts & Execution Cases                      │  │
│  └──────────────────────────────────────┬───────────────────────────────────┘  │
│                                         ▼                                      │
│  ┌───────────────────────┐  ┌───────────────────────┐  ┌────────────────────┐  │
│  │   Migration Engine    │  │  Verification Engine  │  │  Knowledge Engine  │  │
│  │ (Slices & Architect)  │  │ (Differential Replay) │  │ (Lessons/Patterns) │  │
│  └───────────┬───────────┘  └───────────┬───────────┘  └─────────┬──────────┘  │
└──────────────┼──────────────────────────┼────────────────────────┼─────────────┘
               │                          │                        │
               ▼                          ▼                        ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                        ADAPTERS & INFRASTRUCTURE LAYER                         │
│                                                                                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │ Source Adapters (implement SourceAdapter interface)                      │  │
│  │   • adapters/source/php/codeigniter3 (Planned reference)          │  │
│  │   • adapters/source/php/generic_php  (Fallback for non-standard PHP)     │  │
│  │   • adapters/source/python/django, etc.                                  │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │ Target Adapters (implement TargetAdapter interface)                      │  │
│  │   • adapters/target/php/laravel13    (Planned reference)          │  │
│  │   • adapters/target/go/std_http, adapters/target/rust/axum, etc.         │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│  ┌─────────────────────────────┐        ┌───────────────────────────────────┐  │
│  │ SQLite Relational Store     │        │ Filesystem Artifact Store         │  │
│  │ (.rewrotron/project.db)     │        │ (.rewrotron/artifacts/)           │  │
│  └─────────────────────────────┘        └───────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Pluggable Adapter Interfaces in Go

The Application Core communicates with source and target adapters strictly via canonical Go interfaces defined in `internal/domain`:

```go
package domain

import "context"

// SourceAdapter defines framework-specific inspection capabilities.
type SourceAdapter interface {
    // Name returns the unique identifier of the adapter (e.g. "codeigniter3", "generic-php").
    Name() string

    // Detect inspects a repository to determine framework match and confidence score.
    Detect(ctx context.Context, rootPath string) (DetectionResult, error)

    // Analyze parses the codebase and returns the Source IR.
    Analyze(ctx context.Context, project Project) (SourceIR, error)

    // Extract resolves framework conventions into neutral application structure.
    Extract(ctx context.Context, source SourceIR) (ApplicationIR, error)

    // Observe instruments only a sandbox copy and captures runtime traces.
    Observe(ctx context.Context, project Project, options ObserveOptions) (RuntimeTraceSet, error)
}

// TargetAdapter defines target framework code generation capabilities.
type TargetAdapter interface {
    // Name returns the target adapter identifier (e.g. "laravel13", "go-std-http").
    Name() string

    // Plan constructs a target-specific migration plan from Behavior IR.
    Plan(ctx context.Context, behavior BehaviorIR, config TargetConfig) (MigrationPlan, error)

    // Generate emits idiomatic target application code conforming to the requested architecture.
    Generate(ctx context.Context, plan MigrationPlan, outputDir string) (GenerationResult, error)

    // SetupVerification configures the target test harness for differential testing.
    SetupVerification(ctx context.Context, project Project) (VerificationHarness, error)
}
```

---

## 4. The Three Confirmation Gates

1. **Gate 1: Source Detection (`DETECTED` ➔ `SOURCE_CONFIRMED`)**
   - Automatically triggered upon `rewrotron detect` or `rewrotron init`.
   - Inspects language, framework, version, runtime, entrypoints, and dependencies.
   - Requires explicit user confirmation via `rewrotron confirm source` or MCP `project.confirm_source`.
   - Engine stores an immutable configuration hash. Material source or environment drift invalidates the current baseline; immutable prior approvals remain in history.

2. **Gate 2: Target & Migration Plan (`PLAN_SYNTHESIZED` ➔ `PLAN_CONFIRMED`)**
   - Proposes target language, framework, architectural style (e.g. `modular_monolith`), behavioral slices to migrate, and dead code candidates to drop.
   - Requires explicit user confirmation via `rewrotron confirm plan` or MCP `project.confirm_plan`.
   - The engine returns `ERROR: USER_CONFIRMATION_REQUIRED` if code generation is attempted prior to confirmation.

3. **Gate 3: Production Cutover (`VERIFIED` ➔ `READY_CUTOVER`)**
   - Requires passing evidence for 100% of the confirmed in-scope corpus, no required UNKNOWN dimensions, and current benchmark/session/rollback evidence.
   - Final deployment or live database cutover requires explicit human sign-off.


## 5. Construction and service boundary

The interface examples describe planned ports; their records are defined in [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md). `SourceAdapter.Analyze` returns neutral syntax records; `Extract` resolves framework routing/loading semantics. The core cannot infer CI3 conventions from generic syntax. `TargetAdapter.Generate` writes a staging directory and returns a patch manifest; it never overwrites customer edits directly. `Observe` receives a project whose execution paths resolve to an isolated sandbox copy.

`internal/runtime` registers concrete adapters and injects them into the gate-checked service layer. Concrete implementation paths and the Go import rules are owned by [COMPONENTS](COMPONENTS.md). The target-neutral planner checks declared capabilities before calling the target adapter. If a behavior cannot be generated from evidenced rules, return `ERR_UNSUPPORTED_CAPABILITY` with unit and rule references.

## 6. Durable operation contract

A service request carries project ID, request ID, idempotency key for mutations, expected subject hashes, and cancellation context. A result carries operation ID, artifact references, blockers, and captured hashes. Short state changes use one SQLite transaction; long operations use a captured input snapshot and publish through a current-subject/fencing check. See [RECOVERY](../08-storage/RECOVERY.md).

Before invoking adapters, the service checks gates, source drift, capability coverage, safe execution prerequisites, and unit dependencies. Adapters cannot write gate records, declare verification passes, or update project status. `status`, `next`, and MCP snapshots derive facts from the same records.

## 7. Current capability status

No adapter implementation or reference corpus has run in this repository. CI3 and Laravel are initial reference targets; maturity begins at `EXPERIMENTAL` after implementation and is promoted only on recorded corpus evidence. Generic PHP must return unsupported findings for unresolved framework conventions. Additional adapters reuse the same IR and verifier, but each pair still requires a tested capability intersection.

## 8. Operational records

[SERVICE-CONTRACTS](SERVICE-CONTRACTS.md) defines request context, infrastructure handles, capture/runner/lease fields, injection and compare-and-publish semantics shared by implementation tasks.

## 9. Target and dependency decision services

[ADR 0012](ADR/0012-target-specification-and-dependency-resolution.md) makes source detection, recommendations, requested target specifications and dependency resolution separate engine responsibilities. The flow is detect/review/Gate 1 → source/usage/behavior analysis → recommend/configure → immutable target revision → feasibility/plan → Gate 2 → generation/verification → Gate 3. Preliminary metadata recommendations may precede Gate 1; deep analysis cannot.

`internal/detection` coordinates neutral proposals; `internal/recommendations` emits evidence-backed decisions; `internal/target` validates and freezes user intent; `internal/dependency` builds usage/capability/constraint graphs. `internal/migration` remains the planner, `internal/core` enforces approval and publication, and `internal/knowledge` owns lessons. `TargetConfig` is the canonical rich [target specification](../04-migration/TARGET-SPECIFICATION.md), not merely a framework name. Ecosystem parsing/API extraction and generation remain adapter capabilities. Plans capture exact resolution and decision artifact hashes. A target change propagates along recorded dependency links while retaining unrelated source work; incomplete feasibility cannot become executable by approval alone.

## 10. Transparent execution boundary

[ADR 0013](ADR/0013-transparent-execution.md) and [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) make instrumentation a baseline service requirement. `internal/execution` validates typed events and projects active concurrent runs, repeatable stages, factual explanations, progress, findings/approval cards and next actions from the existing ledger. Engine and agent work must publish real start/boundary/result evidence; an agent summary is never completion authority.

All CLI/MCP/future desktop clients read the same committed projection at a sequence watermark. Execution labels do not replace lifecycle states. Progress never grants approval or parity, and telemetry does not change a business revision merely because it advances. Persistent human pause/continue constrains mission readiness independently of gates; resume reconciles without clearing pause. Concrete services receive an injected event sink, clock/redactor and repository; no framework/provider/desktop dependency enters the core.

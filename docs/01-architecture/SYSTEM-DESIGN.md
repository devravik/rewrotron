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
│  │   • adapters/source/php/codeigniter3 (Initial tested reference)          │  │
│  │   • adapters/source/php/generic_php  (Fallback for non-standard PHP)     │  │
│  │   • adapters/source/python/django, etc.                                  │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │ Target Adapters (implement TargetAdapter interface)                      │  │
│  │   • adapters/target/php/laravel13    (Initial tested reference)          │  │
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

    // Observe injects instrumentation hooks and captures runtime traces.
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

1. **Gate 1: Source Detection (`DETECTION_COMPLETE` ➔ `SOURCE_CONFIRMED`)**
   - Automatically triggered upon `rewrotron detect` or `rewrotron init`.
   - Inspects language, framework, version, runtime, entrypoints, and dependencies.
   - Requires explicit user confirmation via `rewrotron confirm source` or MCP `project.confirm_source`.
   - Engine stores an immutable configuration hash. Any subsequent source file mutations invalidate the hash.

2. **Gate 2: Target & Migration Plan (`PLAN_SYNTHESIZED` ➔ `PLAN_CONFIRMED`)**
   - Proposes target language, framework, architectural style (e.g. `modular_monolith`), behavioral slices to migrate, and dead code candidates to drop.
   - Requires explicit user confirmation via `rewrotron confirm plan` or MCP `project.confirm_plan`.
   - The engine returns `ERROR: USER_CONFIRMATION_REQUIRED` if code generation is attempted prior to confirmation.

3. **Gate 3: Production Cutover (`VERIFIED` ➔ `READY_CUTOVER`)**
   - Proves differential equivalence across 100% of active units.
   - Final deployment or live database cutover requires explicit human sign-off.

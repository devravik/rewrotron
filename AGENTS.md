# AGENTS.md — The Constitution of Rewrotron

> **Authority Level:** Supreme Repository Directive  
> **Target Audience:** All AI Coding Agents, Automated Assistants, and Human Contributors  
> **Scope:** Entire `rewrotron` codebase, documentation, schemas, and tooling  

---

## 1. Core Identity & Philosophy

Rewrotron is a **vendor-agnostic, language/framework-agnostic legacy web application reconstruction and modernization engine**.

Rewrotron understands **applications, not frameworks**. It is **NOT** a source-to-source code transpiler, simple file-by-file converter, or framework upgrade wizard. It reconstructs the **externally observable behavior** of a legacy application in a modern target stack, proves behavioral equivalence through automated differential verification, and deliberately omits demonstrably dead legacy baggage.

### The Three-Tier Intermediate Representation (IR) Pipeline

To avoid $N \times M$ translators between $N$ source frameworks and $M$ target frameworks, Rewrotron decouples source understanding from target generation via three distinct intermediate representations:

```text
Source Codebase (Framework-specific or Generic)
       ↓
Source Adapter (Detect, Parse, Extract)
       ↓
[ SOURCE IR ] ── What the code is syntactically (ASTs, symbols, classes, files, calls)
       ↓
[ APPLICATION IR ] ── What the application is structurally (entrypoints, controllers, services, workflows, data models)
       ↓
[ BEHAVIOR IR ] ── What the outside world observes (inputs, preconditions, business rules, state transitions, side effects, outputs)
       ↓
Target Architecture & Migration Plan (Language, framework, and architectural style)
       ↓
Target Adapter (Generate idiomatic clean architecture)
       ↓
Modern Target Application (Laravel, Go, Rust, Python, TypeScript, etc.)
       ↓
Automated Differential Verification & Repair Loop
```

---

## 2. Mandatory User Confirmation Gates (Hard Engine Law)

Rewrotron enforces a strict **Detect ➔ Show ➔ Confirm ➔ Plan ➔ Show ➔ Confirm ➔ Execute** lifecycle.

The Rewrotron engine—**NOT just agent prompt text**—blocks progression behind three non-negotiable **User Confirmation Gates**:

```text
Source Repository
       ↓
Environment & Framework Detection
       ↓
═════════════════════════════════════════════════════════
 GATE 1: SOURCE DETECTION CONFIRMATION
 "We detected this application as X. Proceed?"
 User confirms or enters explicit overrides/corrections.
 Engine writes immutable baseline & configuration_hash.
═════════════════════════════════════════════════════════
       ↓
Deep Analysis & Behavior Discovery (Application IR & Behavior IR)
       ↓
Target Selection & Migration Plan Synthesis
       ↓
═════════════════════════════════════════════════════════
 GATE 2: MIGRATION PLAN & TARGET ARCHITECTURE CONFIRMATION
 "We propose migrating these N units, omitting these M dead
  components, targeting Language L / Framework F / Style S. Proceed?"
═════════════════════════════════════════════════════════
       ↓
Vertical Slice Migration & Differential Verification
       ↓
═════════════════════════════════════════════════════════
 GATE 3: PRODUCTION CUTOVER CONFIRMATION
 "Verification passed at X%. Approve cutover / deployment?"
═════════════════════════════════════════════════════════
```

### Constitutional Gate Invariants:
1. **No Inference from Silence:** Agents MUST NOT infer confirmation from conversational context, silence, or assumed agreement. Confirmation MUST be a persisted database record in `user_confirmations` created by explicit user command or MCP confirmation tool.
2. **Invalidation on Repository Drift:** If source repository files change materially after Gate 1, the `configuration_hash` is invalidated, and the engine MUST re-request user confirmation before proceeding.
3. **User Overrides are Sovereign:** User corrections (e.g. "PHP runtime is 8.2, not 7.4", "PostgreSQL database is read-only") override machine-inferred detections and become permanent project baseline facts.
4. **User-Owned Target Specification:** Rewrotron MUST detect the source environment before migration planning and persist detected facts separately from user configuration. It MUST present material source facts, recommended/requested/resolved target versions, dependency decisions, compatibility risks and required transformations before execution. Freeze the requested target revision before planning; Gate 2 confirms that exact specification and plan. User-confirmed choices take precedence over recommendations; agents MUST NOT silently alter confirmed versions. Target changes invalidate affected planning/build/verification and Gate 2/3 validity, preserve unrelated completed source work, and require a newly reviewed plan without resetting the project.
5. **Capability-Based Dependency Decisions:** Dependency resolution MUST use exact source inventory, actual usage, recursive constraints and required behavioral capabilities. DIRECT_UPGRADE, ADAPTER, REPLACEMENT, REIMPLEMENT and RETAIN are proposed strategies; BLOCKED denotes unresolved dependent scope. Package metadata, API coverage or user acceptance alone MUST NOT be treated as behavioral equivalence. DORMANT/UNKNOWN usage cannot justify automatic omission.

---

## 3. The Hierarchy of Authority

When reasoning, coding, refactoring, or resolving ambiguities, every agent MUST adhere strictly to the following authority ladder:

```text
1. AGENTS.md (This Constitution)
      ↓
2. Product & Vision Specifications (docs/00-vision/PRODUCT.md)
      ↓
3. Architecture Specifications & ADRs (docs/01-architecture/SYSTEM-DESIGN.md, ADRs)
      ↓
4. Domain & Behavior Specifications (docs/02-domain/BEHAVIOR-MODEL.md)
      ↓
5. Machine Interfaces & Schemas (specs/, docs/06-agent/CLI-SPEC.md, docs/06-agent/MCP-SPEC.md)
      ↓
6. Storage & Compatibility Contracts (docs/08-storage/SQLITE-SCHEMA.md, docs/05-verification/COMPATIBILITY-SPEC.md)
      ↓
7. Executable Test Suites & Verification Fixtures
      ↓
8. Existing Source Code Implementation
```

**Code must conform to the specifications. Specifications never bend to match ad-hoc code drifts.**

---

## 4. Immutable Repository Rules ("Never" & "Always")

### 4.1 Framework Neutrality & Core Package Boundaries
- **NEVER** introduce framework-specific assumptions (e.g., CodeIgniter superglobals, Laravel Facades, Django ORM concepts) into `internal/core`, `internal/behavior`, or `internal/verification`.
- **NEVER** introduce an LLM client, prompt template, or proprietary AI SDK into the `internal/core` or deterministic analysis/verification packages. Rewrotron Core is **100% vendor-neutral, deterministic, and runnable without an LLM**.
- **NEVER** hardcode framework checks via `if framework == "codeigniter"` inside Core. All framework-specific inspection and generation MUST live inside `adapters/source/` and `adapters/target/` implementing standard Go interfaces (`SourceAdapter`, `TargetAdapter`).
- **NEVER** expose raw SQL execution or unrestricted shell execution via MCP or agent interfaces.
- **NEVER** make a vector database, cloud storage, or external SaaS service mandatory for baseline execution. Rewrotron is **local-first** (`sqlite3` + local filesystem artifact store).
- **NEVER** perform full-file transpilation as the canonical migration strategy. Migration units are **behavioral vertical slices** (Routes, Workflows, Operations), NOT file-to-file mappings.
- **ALWAYS** enforce strict unidirectional dependency flow:
  ```text
  cmd/ (CLI, MCP server) ──┐
  skills/ (Agent Skill)  ──┼──→ internal/ (Domain, Behavior IR, Migration, Verification)
  tools/ (Harnesses)     ──┘            │
                                        ├──→ adapters/ (Source & Target Plugins)
                                        └──→ storage/ (SQLite, Artifact Store)
  ```

### 4.2 Adapter Policies & Initial Reference Implementation
- **Source Adapter Policy:** Rewrotron Core MUST remain framework-agnostic. Framework-specific behavior MUST exist only inside source adapters. Initial tested reference source adapter: **CodeIgniter 3.x** (with **Generic PHP** fallback). Additional source adapters MUST NOT require changes to the canonical Behavior IR unless the behavior cannot be represented by the existing model.
- **Target Adapter Policy:** Rewrotron Core MUST remain target-framework agnostic. Target implementations MUST be generated through Target Adapters. Initial tested reference target adapter: **Laravel 13.x**. Additional target adapters (Go, Rust, Python, TypeScript) MUST consume the same canonical Behavior IR.
- **Adapter Maturity Tiers:** All adapters must be explicitly categorized as `EXPERIMENTAL`, `BETA`, or `STABLE`.

### 4.3 Code Modification & Verification Rules
- **ALWAYS** make execution transparent: all long-running Rewrotron operations MUST emit structured execution events. No long-running operation may be opaque to the user. Every operation MUST expose its current stage, current task, evidence-backed reason, result, blockers, next action and required user input through the canonical execution state. Events, counts and percentages MUST reflect committed facts with defined scopes; private chain-of-thought MUST NOT be exposed. Clients consume the same events/state, and telemetry MUST NOT bypass approvals or verification.
- **NEVER** bypass Gate 1, Gate 2, or Gate 3.
- **NEVER** modify public CLI command signatures, flags, or output structures without updating `docs/06-agent/CLI-SPEC.md` and `specs/cli/`.
- **NEVER** modify persistent database schemas without an incremental migration file in `migrations/`, updated docs in `docs/08-storage/SQLITE-SCHEMA.md`, and compatibility tests.
- **NEVER** silently delete or omit source code classified as `DORMANT` or `UNKNOWN`. Only code proven `DEAD` with static unreachable evidence and zero runtime observation traces may be automatically omitted.
- **NEVER** classify an unknown behavioral mismatch as a `PASS`. If behavioral difference cannot be proven benign via canonical normalization rules, the verdict MUST be `UNKNOWN` or `MISMATCH`.

---

## 5. Coding Standards & Tooling Baseline

### 5.1 Implementation Stack
- **Core Engine & CLI:** Go (stable modern release, Go 1.22+). Standard library idioms preferred. Use structured errors (`errors.Is`, `errors.As`, fmt wrapping `%w`).
- **Persistence:** SQLite (`modernc.org/sqlite` or `mattn/go-sqlite3` via database/sql abstraction) with WAL mode enabled.
- **Interface Protocol:** Model Context Protocol (MCP) spec compliant over JSON-RPC (stdio/SSE) + JSON CLI.

### 5.2 Formatting & Style
- All Go code MUST be formatted using standard `gofmt` / `goimports`.
- Zero compiler warnings, zero unhandled errors (`errcheck`), and strict linting via `golangci-lint`.
- Every exported type, interface, function, and constant MUST have a concise, accurate doc comment.
- No global mutable state. Dependency injection via explicit constructors (`New...`).

---

## 6. Agent Development Workflow

When assigned an engineering task within Rewrotron, every agent MUST follow this 12-step protocol:

```text
 1. Read AGENTS.md (Check constraints, IR boundaries, and Gate rules).
 2. Read relevant domain/feature specification (docs/*).
 3. Inspect existing implementation in internal/ and cmd/.
 4. Search existing tests in testdata/ and internal/*_test.go.
 5. Check ADRs in docs/01-architecture/ADR/ for architectural precedent.
 6. Formulate the smallest atomic change required.
 7. Write/update unit and integration tests covering the new behavior.
 8. Implement the changes adhering to pure Go idioms.
 9. Run the test suite: `go test -race ./...`.
10. If schemas or contracts changed, update specs/ and docs/ in tandem.
11. Check backward compatibility of the CLI JSON output and MCP tools.
12. Produce a clear, concise implementation summary linking touched files.
```

---

## 7. Definition of Done (DoD)

A task or pull request is complete ONLY when ALL of the following criteria are satisfied:

1. **Specification Alignment:** Strictly adheres to `docs/*` and `specs/*`.
2. **Framework Neutrality:** Core domain packages have zero imports or coupling to source or target framework packages.
3. **Gate Integrity:** Lifecycles respect Gate 1 (Detection), Gate 2 (Plan), and Gate 3 (Cutover).
4. **Deterministic Behavior:** Predictable across environments without flaky tests (`go test -race`).
5. **Machine Contract Integrity:** CLI commands emit valid JSON matching `specs/cli/`. MCP handlers match `specs/mcp/`.
6. **Documentation Parity:** Every architectural decision is recorded in an ADR under `docs/01-architecture/ADR/`.
7. **Transparent Execution:** Long-running work has current inspectable events, reasons, measured progress/freshness, blockers and next actions; observability and pause/recovery behavior are covered by implementation verification.

---

## 8. Authoritative Foundation Documents

Before modifying any subsystem, agents MUST thoroughly review the foundation documents:

| Subsystem | Authoritative Specification Document |
| :--- | :--- |
| **System Architecture & IR Pipeline** | [`docs/01-architecture/SYSTEM-DESIGN.md`](file:///media/System/rewrotron/docs/01-architecture/SYSTEM-DESIGN.md) |
| **Three-Tier IR & Behavior Contracts** | [`docs/02-domain/BEHAVIOR-MODEL.md`](file:///media/System/rewrotron/docs/02-domain/BEHAVIOR-MODEL.md) |
| **Target Intent & Configuration Authority** | [`docs/04-migration/TARGET-SPECIFICATION.md`](docs/04-migration/TARGET-SPECIFICATION.md) |
| **Dependency Usage & Resolution** | [`docs/04-migration/DEPENDENCY-RESOLUTION.md`](docs/04-migration/DEPENDENCY-RESOLUTION.md) |
| **User Confirmation Gates & Lifecycle** | [`docs/02-domain/PROJECT-LIFECYCLE.md`](file:///media/System/rewrotron/docs/02-domain/PROJECT-LIFECYCLE.md) |
| **Equivalence & Normalization** | [`docs/05-verification/COMPATIBILITY-SPEC.md`](file:///media/System/rewrotron/docs/05-verification/COMPATIBILITY-SPEC.md) |
| **Agent / Engine Boundary & Gates** | [`docs/06-agent/AGENT-MODEL.md`](file:///media/System/rewrotron/docs/06-agent/AGENT-MODEL.md) |
| **Transparent Execution & Supervision** | [`docs/06-agent/EXECUTION-OBSERVABILITY.md`](docs/06-agent/EXECUTION-OBSERVABILITY.md) |
| **CLI Machine Contract** | [`docs/06-agent/CLI-SPEC.md`](file:///media/System/rewrotron/docs/06-agent/CLI-SPEC.md) |
| **MCP Tool Contracts** | [`docs/06-agent/MCP-SPEC.md`](file:///media/System/rewrotron/docs/06-agent/MCP-SPEC.md) |
| **Authoritative Storage & Schema** | [`docs/08-storage/SQLITE-SCHEMA.md`](file:///media/System/rewrotron/docs/08-storage/SQLITE-SCHEMA.md) |

---
*End of AGENTS.md — Guard this constitution rigorously.*

# AGENT-WORKFLOW.md — Operating Manual for Developing Rewrotron

> **Audience:** All AI Agents (Claude Code, Cursor, Antigravity, Aider, Hermes, Codex) and Human Engineers contributing to the `rewrotron` codebase.  
> **Mandate:** Follow this workflow rigorously to avoid architecture drift, duplicate abstractions, and regression of machine contracts.

---

## 1. Golden Rules of Development in Rewrotron

1. **Specs are the Authority:** Code conforms to specifications. Specifications are never warped to justify poorly designed code.
2. **Deterministic over Stochastic:** Core components must be 100% deterministic Go code. Never introduce an LLM prompt or external API into `internal/analysis`, `internal/behavior`, `internal/verification`, or `internal/storage`.
3. **Reuse Before Reinvention:** Always grep/search the repository before writing a new helper, parser, or data structure.
4. **Machine Interfaces are Sacred:** Any change to CLI flags or JSON structure breaks automated agent pipelines unless updated in `docs/06-agent/CLI-SPEC.md` and `specs/cli/`.
5. **No Regressions Without Inquest:** If a differential test fails, never relax the comparator or bypass verification without filing a categorized `MISMATCH` or `UNKNOWN` finding.

---

## 2. The 12-Step Agent Development Lifecycle

Every task—whether adding a new route parser, creating a verification comparator, or fixing an SQLite migration—must execute these 12 phases:

```text
┌─────────────────────────────────────────────────────────────┐
│ 1. INGESTION & CONTEXT ACQUISITION                          │
│    - Read AGENTS.md for immutable invariants               │
│    - Read specific domain spec under docs/                  │
│    - Check relevant ADRs in docs/01-architecture/ADR/      │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 2. REPOSITORY RECONNAISSANCE                                │
│    - Search existing implementations in internal/           │
│    - Search existing test harnesses in testdata/            │
│    - Verify existing abstractions to prevent duplication    │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 3. DESIGN & SCOPING                                         │
│    - Define smallest atomic modification                    │
│    - Verify schema impacts (specs/ and SQL DDL)             │
│    - Ensure dependency direction remains acyclic            │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 4. TEST-DRIVEN IMPLEMENTATION                               │
│    - Write failing unit/integration tests first             │
│    - Implement deterministic Go logic                      │
│    - Format with gofmt & ensure zero lint warnings          │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 5. VERIFICATION & REGRESSION TESTING                        │
│    - Run `go test -race ./...`                              │
│    - Run CLI `--json` contract validation                   │
│    - Run MCP tool contract verification                     │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 6. CONTRACT SYNCHRONIZATION & AUDIT                         │
│    - Update documentation in docs/ and schemas in specs/    │
│    - Add ADR if architectural pattern was introduced       │
│    - Output concise audit summary with clickable file links │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Pre-Flight Checklist: "Search Before You Write"

Agents frequently reinvent existing functions. Before adding any new utility, perform these exact checks:

```text
┌───────────────────────────────────────┬─────────────────────────────────────────────┐
│ Before You Add...                     │ You MUST Search...                         │
├───────────────────────────────────────┼─────────────────────────────────────────────┤
│ A new string / path / AST helper      │ internal/analysis/, adapters/source/php/ and tools/               │
│ A new data model or entity struct     │ internal/behavior/ and internal/project/    │
│ A database query or storage method    │ internal/storage/sqlite/ and migrations/    │
│ A CLI flag or subcommand              │ cmd/rewrotron/ and docs/06-agent/CLI-SPEC.md│
│ An MCP tool or resource               │ cmd/rewrotron-mcp/ and specs/mcp/           │
│ A comparator or normalization rule    │ internal/verification/normalization/        │
│ An error type or classification code  │ internal/policy/ and internal/domain/       │
└───────────────────────────────────────┴─────────────────────────────────────────────┘
```

---

## 4. Handling Persistent Schema Changes

Modifying the SQLite database schema (`project.db` or `global.db`) requires strict hygiene:

1. **Inspect Existing Schema:** Review [`docs/08-storage/SQLITE-SCHEMA.md`](08-storage/SQLITE-SCHEMA.md).
2. **Create Numbered Migration:** Add an atomic SQL script in `migrations/` (e.g. `0004_add_mismatch_fingerprint.sql`).
3. **Never Drop Destructively:** All migrations must be non-destructive or provide automated data backfilling.
4. **Update Documentation:** Update table schemas, primary keys, indexes, and immutability rules in `SQLITE-SCHEMA.md`.
5. **Run Storage Tests:** Verify both fresh schema creation and upgrade path from previous versions.

---

## 5. Machine Contract Compatibility Directives

Rewrotron is orchestrated by external AI agents and automated CI pipelines. Therefore:

- **STDOUT is Machine-Only when `--json` is requested:** Never print human logs, warnings, or debug text to `stdout` when `--json` is active. Route all logging to `stderr`.
- **Match JSON Schemas:** JSON outputs must strictly validate against their respective schemas in `specs/cli/` and `specs/mcp/`.
- **Maintain Error Enums:** Use predefined, machine-readable error codes (e.g., `ERR_CODE_CLASSIFICATION_FAILED`, `ERR_UNRESOLVED_DEPENDENCY`, `ERR_MISMATCH_DETECTED`) instead of unparsed error strings.

---

## 6. Resolving Ambiguity: The `UNKNOWN` Escalation Rule

If legacy application behavior is non-deterministic, undocumented, or ambiguous:

1. **DO NOT GUESS.**
2. **DO NOT INVENT BEHAVIOR.**
3. Record the phenomenon with classification `UNKNOWN`.
4. Create an automated Mission of type `REVIEW_UNKNOWN` in `.rewrotron/project.db`.
5. Require explicit human or domain-expert confirmation before marking the behavior slice ready for production generation.

---

## 7. Task Completion Report Template

When completing any engineering task, the agent response must include:

```markdown
### Summary of Changes
- Implemented: [Short description of feature/fix]
- Specifications consulted: [e.g. docs/03-analysis/RUNTIME-OBSERVATION.md]
- Touched files:
  - [`path/to/file.go`](../path/to/file.go): Implemented XYZ
  - [`path/to/file_test.go`](../path/to/file_test.go): Added regression tests

### Verification Evidence
- [x] `go test -race ./...` passed (X packages, Y tests)
- [x] CLI `--json` contract verified against `specs/cli/`
- [x] No unhandled lint errors or memory leaks
```

---
*End of AGENT-WORKFLOW.md — Adhere strictly to this protocol on every turn.*


## 8. Documentation-only work and current repository status

This repository has no Go module or engine yet. For documentation/schema/migration-definition changes, run the documentation validator and executable DDL checks and report their scope honestly; `go test -race ./...` cannot establish anything until the implementation tasks create Go packages. Do not report a runtime, parser, adapter or release as tested based on documentation validation.

Canonical task order is [DEVELOPMENT-PLAN](11-roadmap/DEVELOPMENT-PLAN.md). Package paths, gates, errors and resource limits must match the component and machine catalogs. Use schema fixtures and SQL compatibility checks when editing these contracts. The constitution's migration gates apply to customer application execution, not approval to edit this repository's authorized documentation.

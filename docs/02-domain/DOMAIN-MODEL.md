# Domain entities and cardinality

The [glossary](../00-vision/GLOSSARY.md) owns vocabulary; [IR-CONTRACTS](IR-CONTRACTS.md) owns document fields; [SQLite schema](../08-storage/SQLITE-SCHEMA.md) owns persistence. Entity names are stable across Go, CLI, and MCP.

| Entity | Identity and ownership | Relationships and invariants |
| --- | --- | --- |
| Project | `project_id` | Owns one operational workspace and historical baselines |
| Detection | `detection_id` + manifest | Immutable evidence-backed proposal; not confirmation |
| Baseline | `configuration_hash` | Confirmed facts plus overrides and immutable manifest |
| Confirmation | `confirmation_id` | One human approval of one gate subject; append-only |
| SourceFile / SourceSymbol | Baseline + stable ID | File owns symbols; spans and parser completeness retained |
| Entrypoint | Baseline + `entrypoint_id` | HTTP/CLI/cron/event dispatch; many-to-many operations |
| Operation / BehaviorContract | `operation_id`, semantic version, revision | Governs observable rules and effects; several concrete cases |
| BehaviorCase | `case_id` + content hash | Binds contract version, fixture, inputs, expected dimensions |
| Workflow | `workflow_id` | Ordered/bounded steps referencing operations and bindings |
| ObservationTrace | `trace_id` | One actual run with completeness statuses; never inferred execution |
| TargetSpecification | `target_spec_hash` + immutable artifact | User-owned requested constraints and migration intent; Gate 2 binds the frozen revision |
| Recommendation | ID/revision + target subject | Evidence-backed proposal with type and explicit decision; never configuration authority |
| Dependency / Capability / Resolution | Baseline/target subject + stable IDs | Source usage and recursive constraints; selected strategy and cases are scoped to required behavior |
| FeasibilityAssessment | Target/source/graph subjects + artifact | Required compatibility and unknowns; scoped proceed decision is not parity |
| MigrationPlan | `plan_id` + `plan_hash` | Immutable target choices, units, omissions, scope and blockers |
| MigrationUnit | `unit_id` within plan | Many contracts/cases; explicit prerequisite units |
| VerificationRun | `run_id` + captured hashes | Owns case results and deterministic dimension verdicts |
| Mismatch | `mismatch_id` | A difference within a run; classification is separate from verdict |
| Experiment / RepairAttempt | Fingerprint + attempt ID | Hypothesis, before/after hashes, outcome and evidence |
| Mission | `mission_id` | Mutable scheduling projection, fenced leases, immutable attempt events |
| Finding / Blocker | `finding_id` | Required decision, missing evidence, or unsupported capability |
| Artifact | SHA-256 digest | Immutable redacted bytes; reachable references prevent collection |
| Benchmark | `benchmark_id` | Captured workload, environment and measurement report |
| Lesson / SkillVersion | ID + scope/revision | Project-only by default; explicit privacy-reviewed global promotion |

A new project may have zero source files, routes, contracts, cases, or units until their stages run. A case belongs to one contract revision and may contribute to multiple units/workflows. A verification result references a run, two observations, a case, and policy hash. A mismatch is derived from that result rather than directly owned by a source file. Dependencies can reference shared components without declaring every shared table a migration prerequisite.

Do not use `BehavioralCase`, `PARITY_ACHIEVED`, `DETECTION_COMPLETE`, or `APPROVED` as competing canonical names. Lifecycle and unit enums are defined in [PROJECT-LIFECYCLE](PROJECT-LIFECYCLE.md), mission enums in [MISSION-PROTOCOL](../06-agent/MISSION-PROTOCOL.md), and verdicts in [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md).

## Execution supervision entities

ExecutionRun identifies an operation attempt (initially the same ID); StageInstance and Task identify repeated/concurrent work with current subjects and public reasons. ExecutionEvent is a typed projection of the existing append-only ledger; ProgressMetric cites its scope/basis/accounting revision. ExecutionState and ExecutionPage are reconstructable read models, not new approval or verdict entities. Human ControlRequest is a durable event whose safe acknowledgment is distinct from lifecycle state. [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) owns these definitions.

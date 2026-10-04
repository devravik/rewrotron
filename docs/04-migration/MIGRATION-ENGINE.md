# Migration engine

**Owner:** Execution sequence and publication orchestration. [SYSTEM-DESIGN](../01-architecture/SYSTEM-DESIGN.md) owns pipeline; [MIGRATION-UNITS](MIGRATION-UNITS.md) owns scoped acceptance; [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) owns exact approvals.

## Guarded sequence

1. Acquire/seal source and profiles; bounded detection; explicit Gate 1.
2. Extract Source/Application IR; inventory dynamic uncertainty and operational roots; observe isolated scenarios with mechanism attestations.
3. Persist an assessment projection from available shared analysis/behavior/feasibility records; partial or blocked feasibility still permits completion. Stop here for assessment-only use. Await an explicit human continuation selection.
4. On a separate plan request, validate selected assessed work and current finding validity, record changes since assessment, complete applicable executable contracts/corpus and exact runtime/dependency/capability resolution and immutable performance policy, then create Target Implementation Model and ordered plan subject. In-scope blockers prevent execution, not assessment completion. Confirm Gate 2.
5. Claim ready unit and project mutation/resource ownership; stage adapter implementation using pinned inputs and revisions. Build in resolved environment.
6. Publish through before-hash/fence/journal checks; verify independent source/target fixtures and UNIT plus implemented regressions. Complete only scoped evidence; continue dependent units.
7. Diagnose MISMATCH/UNKNOWN separately. Repairs are bounded proposals, never comparator relaxation to hide faults. Every changed shared component reverifies currently implemented consumers in the active scope; future consumers inherit pending obligations.
8. Execute PROJECT corpus, required performance claim and release/rollback rehearsals; freeze release subject, confirm Gate 3 and package.

No execution chooses a lower user runtime or resolves latest after approval. A changed target/code/corpus/policy follows typed invalidation. Historic artifacts remain retained; reuse requires engine-validated certificates where allowed. Missing pinned bytes/capability blocks execution. Foreground/supervised lifetimes and crash reconciliation follow [RECOVERY](../08-storage/RECOVERY.md).

## Safety and repair budgets

Initial mutation is serialized per project, including shared files. Sandbox replay has independent databases/session/cache/files/sinks and validated reset boundaries; production/irreversible effects are prohibited. Retry after uncertain external acknowledgement is not automatic: reconcile receipts/idempotency first. Repair attempts are limited to three per incident subject; changing a trivial patch fingerprint does not reset this budget. Further attempts require explicit reviewed incident/cause change. Failure leaves actionable evidence and current UNKNOWN/MISMATCH.

Optimization is optional only after relevant scope parity; it produces new implementation bytes and scoped regression/performance evidence, never alters behavior obligations silently. A plausible target is not accepted until the verifier can reject deliberate wrong variants under the same approved policy.

The [assessment checkpoint](../02-domain/PROJECT-LIFECYCLE.md) is shared with this sequence. Assessment never enqueues implementation, creates Gate 2 receipts or publishes target code. Current version 2.1 plan envelopes bind assessment/selection/planning-basis artifacts through the existing approval subject decision refs; service equality/currentness checks precede Gate 2 and unit admission.

## Profile composition

Select SOURCE and TARGET packages independently under [SKILL-SPEC](../06-agent/SKILL-SPEC.md) and [SERVICE-CONTRACTS](../01-architecture/SERVICE-CONTRACTS.md). Source extraction establishes neutral obligations; target capabilities lower them through TIM. A source package never directly calls a specific target package. Framework/version matches or intersecting capability names are candidate filters only; exact qualification, executable obligation coverage and independent verification are required before affected execution/acceptance. CI3 → Laravel is the planned Reference Profile #1, not an execution default overriding target intent. New packages use the same lifecycle, unit scopes and gates.

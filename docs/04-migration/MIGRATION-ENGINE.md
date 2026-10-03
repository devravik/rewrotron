# Migration engine

**Owner:** Execution sequence and publication orchestration. [SYSTEM-DESIGN](../01-architecture/SYSTEM-DESIGN.md) owns pipeline; [MIGRATION-UNITS](MIGRATION-UNITS.md) owns scoped acceptance; [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) owns exact approvals.

## Guarded sequence

1. Acquire/seal source and profiles; bounded detection; explicit Gate 1.
2. Extract Source/Application IR; inventory dynamic uncertainty and operational roots; observe isolated scenarios with mechanism attestations.
3. Build evidenced Behavior IR and typed executable contracts/cases; review corpus/normalization and independent falsification controls.
4. Read requested target, resolve exact runtime/dependencies/capabilities, create Target Implementation Model and deterministic ordered plan subject; material blockers remain explicit. Confirm Gate 2.
5. Claim ready unit and project mutation/resource ownership; stage adapter implementation using pinned inputs and revisions. Build in resolved environment.
6. Publish through before-hash/fence/journal checks; verify independent source/target fixtures and UNIT plus implemented regressions. Complete only scoped evidence; continue dependent units.
7. Diagnose MISMATCH/UNKNOWN separately. Repairs are bounded proposals, never comparator relaxation to hide faults. Every changed shared component reverifies consumers.
8. Execute PROJECT corpus, required performance claim and release/rollback rehearsals; freeze release subject, confirm Gate 3 and package.

No execution chooses a lower user runtime or resolves latest after approval. A changed target/code/corpus/policy follows typed invalidation. Historic artifacts remain retained; reuse requires engine-validated certificates where allowed. Missing pinned bytes/capability blocks execution. Foreground/supervised lifetimes and crash reconciliation follow [RECOVERY](../08-storage/RECOVERY.md).

## Safety and repair budgets

Initial mutation is serialized per project, including shared files. Sandbox replay has independent databases/session/cache/files/sinks and validated reset boundaries; production/irreversible effects are prohibited. Retry after uncertain external acknowledgement is not automatic: reconcile receipts/idempotency first. Repair attempts are limited to three per incident subject; changing a trivial patch fingerprint does not reset this budget. Further attempts require explicit reviewed incident/cause change. Failure leaves actionable evidence and current UNKNOWN/MISMATCH.

Optimization is optional only after relevant scope parity; it produces new implementation bytes and scoped regression/performance evidence, never alters behavior obligations silently. A plausible target is not accepted until the verifier can reject deliberate wrong variants under the same approved policy.

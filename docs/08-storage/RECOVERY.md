# Crash recovery, target publication, and resume

## Operations

Persist an operation ID, request fingerprint, captured hashes and `RUNNING` status before long work. Checkpoints reference immutable artifacts; they record completed phases, not arbitrary resumable process memory. On restart, `resume` loads state from SQLite, validates schema/checksums, reconciles staged files and expired leases, checks drift, and returns actionable missions. Conversation history is not required.

## Recovery matrix

| Crash point | Recovery |
| --- | --- |
| Before artifact rename | Remove stale owned temporary file after active operation check |
| After rename, before metadata commit | Retain unreferenced object; reconcile/collect after grace period |
| During entity/event transaction | SQLite rollback; no gate or state transition committed |
| After commit, before JSON response | Same idempotency key returns committed result |
| During JSONL export | Resume from committed sequence; deduplicate event IDs |
| During target publication | Inspect patch journal and before/after hashes; finish or revert owned changes |
| During replay | Terminate owned process/container resources; restore fixture for a fresh run |
| Worker loses lease | Old fence cannot publish or complete; new owner revalidates checkpoints |
| User edits target during recovery | `ERR_TARGET_CONFLICT`; preserve user edits and patch evidence |

## Target patch journal

Generation writes staging files and a manifest containing safe relative paths, action, before/after hashes and unit/fence. Record PREPARED, check current Gate 1/2 subjects and fencing token, then record publication intent. Apply with atomic per-file replacement and a recoverable journal; multiple files are not falsely claimed to be one filesystem transaction. Persist completion/target tree hash only after every file is verified.

Serial publication prevents two leases from editing overlapping files. A process crash does not permit a unit to run from a partially published tree; startup reconciles the journal before build/replay. Rollback restores only files whose current digest matches the patch's after hash. Failed verification never overwrites subsequent human edits.

No time bound such as “rollback in 50 ms” is promised. Every rollback outcome is recorded and revalidated. Source files remain immutable throughout.

## Backups and upgrades

Use SQLite's consistent online backup API or close/checkpoint the DB under exclusive workspace coordination before copying it; never copy only a live main DB while ignoring WAL. Back up referenced artifact bytes and config, hash the manifest, and verify restore by reopening and checking references/schema/gates. A restored workspace rechecks actual source/target drift before work.

Migrations have monotonically numbered checksummed scripts. Failure leaves the prior version usable. A newer database is rejected, not downgraded. Stale idempotency results cannot count as current execution success; return their historical captured hashes and require revalidation for a changed subject.

## Configuration revision recovery

On resume, validate primary/legacy/split intent provenance and compare its semantic target hash with the current approved plan. Changed intent blocks new target execution/publication and invalidates affected Gates 2/3; retain unchanged Gate 1, source/behavior artifacts and historical unit results. Reconcile immutable target/decision artifact refs before exporting YAML views. A partly written `approvals.yaml` or `plan.yaml` is regenerated from SQLite and cannot grant authorization. Reuse target work only after recorded prerequisite and dependency checks under the successor plan. See [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md).

## Execution timeline and persistent control recovery

Rebuild live state/timeline from committed operation/mission/finding/confirmation records and execution events. Recheck liveness rather than trusting a pre-crash RUNNING display. Recovered attempts preserve old event/causation/checkpoint links and receive fresh operation IDs when replay is necessary. Human pause survives restart and `resume`; only explicit continue clears it after current readiness checks. Acknowledging pause waits for patch-journal consistency or fixture cancellation/reset evidence. Export/client disconnect does not stop the engine; an event-store failure stops admission at safe boundaries and never authorizes duplicate side effects. See [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md).

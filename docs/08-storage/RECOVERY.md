# Process ownership, crash recovery and coordinated backup

**Owner:** Operation lifetime, subprocess/resource ownership, publication reconciliation and backups. [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) owns state/event-window semantics.

## Foreground and supervised work

Foreground CLI owns its engine process and child groups; termination/disconnect of that process stops admission, cancels owned children, restores fixtures or leaves recoverable journal/checkpoint evidence. A subsequent resume reconciles and creates a new attempt, never claims a dead process is alive. Closing a read-only status watch affects only the watcher.

Explicit `--background` submits to an engine-owned local supervisor under an owner-only local IPC endpoint. All long MCP actions require supervised admission; missing supervisor/capability returns ERR_UNSUPPORTED_CAPABILITY. No implicit detached shell process. Supervisor startup acquires OS workspace ownership plus SQLite operation/resource claims, authenticates same-account clients locally, records instance ID/host boot ID/PID/start identity, launches owned child groups/containers, heartbeats every 15 seconds and renews leases. No public network daemon is part of this contract. PID alone cannot identify a process after reuse.

Shutdown stops admission and drains/cancels at safe boundaries; persistent human pause survives restart. Client reconnect reads operation IDs/checkpoints/current subjects and can inspect results without resubmission. Client connection state never determines supervised process lifetime. Owner crash marks attempts stale after lease expiry; recovery terminates/reconciles old owned resources before granting a new fence. Orphan detection uses project/run labels plus host/start identity, never kills an unrelated PID. Unknown resource ownership blocks cleanup/admission pending operator review. Target/resource claims and acknowledgements are durable; process memory is not.

## Publication and replay

Persist captured subject/fingerprint/operation before work. Artifacts publish bytes before metadata transaction; unreferenced bytes are reconciled after grace. Target patches have owned safe paths, before/after hashes and PREPARED/publication journal. Initial project mutation and publication are serialized. Hold supervisor/OS resource ownership across fence check and filesystem write; a SQL fence alone cannot prevent a stale process writing after a check. Lease handoff waits for old owner termination/reconciliation.

Atomic replacement is per file, not a multi-file filesystem transaction. Resume inspects journals and verifies each path before finishing/reverting owned changes. Rollback touches only current bytes matching the patch after hash; human edits produce conflict. Interrupted replay kills owned processes and restores all reset dimensions before a fresh attempt. Uncertain external receipt is reconciled before retry; never replay to fix a missing event. Paused/stale/error events are honest about unfinished cleanup.

## Coordinated backup and retention barrier

Under a brief publication/GC barrier, take a consistent SQLite online backup and capture its committed sequence. Derive the entire referenced artifact closure from that backed-up DB, pin objects, then copy/verify bytes and config. Archive manifest version 2 names project, schema/migration checksums, DB checksum, through_sequence, sorted object digests/sizes/retention classes and private-key policy. Release pins only after checksum/reference verification. A DB copy plus an unrelated later artifact copy is not a complete backup.

Restore into an empty owner-only workspace, verify manifest/DB/object digests, integrity_check/FKs/schema/canonical subject reconstruction and private correlation material, then recheck source/target drift before current execution. Missing object/archive/key yields BACKUP_INCOMPLETE, not valid gates. Private keys are only included by explicit secure policy and encrypted export. Archives preserve approval/release closure; moving bytes requires verified local retrieval by digest before resume.

Migrations are incremental, transactionally applied after preflight with monotonically numbered checksummed scripts. Do not rewrite 0001 or downgrade newer DBs. Preflight 0002 rejects cross-project refs/duplicate active work/results; never deletes history to satisfy a constraint. A failed upgrade leaves the prior schema usable. Old v1 subjects are readable but cannot become v2 approvals automatically.

## Required recovery tests

Interrupt before/after artifact rename/metadata commit, every patch journal phase, fixture replay, lease renewal, sink acceptance before ack, event persistence and DB/artifact backup. Kill foreground owner, kill MCP client with surviving supervisor, restart paused supervisor and simulate PID reuse. Verify no duplicated effects, invented PASS, lost approval or overwritten customer edits.

## Persistence failure and effect receipts

On normal event/DB/artifact persistence failure, stop admission and cancel owned runners at a safe boundary. Use the preallocated private journal to record operation/attempt, last committed event sequence, durable receipt identity, failure and cleanup obligations; flush its checksum/sequence chain. This emergency journal is recovery input, not a second event authority or approval ledger. Status reports PERSISTENCE_UNAVAILABLE with last durable N and clearly labels emergency/uncommitted facts; it must not invent a committed blocked event or PASS.

Controlled sinks durably record acceptance/business application before ACK, using reserved receipt capacity. If receipt persistence fails, do not ACK or blindly reissue the business effect. Reconcile actual sink application ID/receipt and attempt before any fresh replay; unknown acceptance remains UNKNOWN and blocks retry. If the emergency medium itself fails, emit bounded redacted stderr/IPC diagnostic, terminate owned processes where possible and mark the attempt untrusted on reopen from the durable admitted-operation record. Explain that the final failure detail could not be persisted; never claim guaranteed recovery from device loss.

Restart acquires ownership, verifies journal checksums/sequence and retained receipt bytes, imports reconciled facts transactionally into the canonical store, records cleanup completion, validates reset and only then releases slots. Torn records/incomplete receipts leave UNKNOWN and require evidence/operator review, never automatic replay. Recovery retains original human pause/approval state; journal entries cannot authorize implementation. Unsupported filesystem reservation/fsync semantics block this execution profile.

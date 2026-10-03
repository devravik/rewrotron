# Artifact classes, quotas and retention

**Owner:** Byte identity, availability, retention/admission and archive manifest. [RECOVERY](RECOVERY.md) owns coordinated DB/artifact backups. SQLite stores metadata, not huge payloads.

Agent-readable sanitized bytes are content-addressed as sha256 refs. Private source/fixtures live in an owner-only private vault and are not automatically exported or exposed by artifact read. Hash/source/AST/diagnostic/DOM/screenshot/build output can itself be sensitive. Redaction and source-derived export policy follow [SECURITY](../09-security/SECURITY.md); do not redact source and then claim its behavior is unchanged.

Write sanitized artifacts to owned temporary store paths, hash/size/fsync and atomically rename within one local filesystem. Verify an existing object before reuse. Commit registration/links/events after bytes; crash leaves reconcilable unreferenced objects. Readers accept registered refs only, validate ownership/digest/size, reject traversal/symlink/race escapes. Corrupt/unavailable protected bytes block reconstruction/readiness.

| Class | Retention / admission |
| --- | --- |
| Approval subjects/parent evidence, release evidence | Preserve complete immutable reconstructable closure indefinitely or in verified archive; never silent GC |
| Current contracts/cases/policies/required verification | Protected while current and while referenced by approval/release |
| Source snapshots/private fixtures | Protected while referenced; private archive policy and keys required |
| Failed experiments/lessons | Retain compact summaries and required evidentiary refs; unprotected raw diagnostics expire under reviewed policy |
| Raw traces/screenshots | Bounded capture/quota; protected selections retained, unprotected raw bytes default 30 days |
| Heartbeats/inspection | Coalesced bounded summaries; committed ledger events are not rewritten/deleted in initial implementation |
| Temporary/unreferenced bytes | Reconcile active attempts; eligible after 24-hour grace |
| Global knowledge | Sanitized versioned evidence and consumer/retraction graph; separate opt-in global store |

Initial declared workspace quota defaults: 50 GiB total artifacts/private snapshots, 2 GiB SQLite metadata, 2 GiB reserved host free space. Capture limits remain per profile. These are configurable admission limits, not measured scalability claims. Project setup shows limits; quota changes affecting capture/profile require updated evidence/plan. Protected bytes can exhaust quota: stop admission before capture/publication, checkpoint and report required space/archive action, never silently sample or delete required evidence. Unknown free space/unsupported filesystem blocks safe admission.

The initial scale target for future soak validation is 100,000 symbols, 10,000 cases and 100,000 execution events per active project. Actual supported scale is UNKNOWN until measured. Initial append-only event metadata is not compacted; long-lived work can require raising limits or archiving a closed workspace. Partitioned event epochs/online metadata compaction are explicitly deferred P2 with migration/restore requirements before claiming years of unbounded active history.

Archive v2 manifest includes schema version, project, DB checksum, committed sequence, migration checksums and sorted object entries `{ref,size_bytes,retention_class,private}`. Validate closure, checksums and restore controls before moving original bytes. Resolver must verify locally restored archived objects before execution; archive pointers never imply available evidence. Refuse exports containing unauthorized private/sensitive closure. Quotas are enforced across DB/WAL/temp/snapshot/artifact growth, not only final objects.

Assessment artifacts follow the shared retention/privacy rules. Preserve published report projections and their provenance/analysis closure for later review/continuation; never silently delete them merely because migration is blocked or assessment-only use stopped. Archive through integrity-checked manifests when admitted quotas require it. Any assessment/report closure consumed by an approved plan inherits protected approval evidence retention. Current applicability and historical retention are separate; missing archived bytes must be restored before reuse, never recreated as evidence.

## Reserved capacity and bounded admission

The 2 GiB free-space guard is an admission threshold, not a guarantee against concurrent writers. Before a mutating/observing operation reserve its declared worst-case artifact/temp/WAL/receipt budget in the local workspace allocator and a fully allocated 64 MiB private recovery journal on the same supported filesystem. Sparse allocation does not qualify. Initial policy permits one active mutation per project; cross-project reservations on the same volume are checked under an owner-level allocator lock. Unknown/unbounded capture size is refused or bounded by an approved profile that reports incomplete capture on limit, never successful truncation.

Recovery journal uses fixed preallocated 64 KiB slots, checksummed records and a monotonic sequence; reserve enough slots for the bounded attempt/receipt/cleanup count before starting. Each slot stores recovery-journal 1.0 plus bounded private receipt bytes or a reserved durable receipt ref. Fsync before acknowledging durability. Never reuse unreconciled slots; fewer available slots than required blocks admission. Reserve failures stop work before effects. Actual device failure may still prevent writes; no filesystem protocol promises durability on a failed device. RECOVERY defines explicit uncertainty in that case.

Quota accounting includes DB/WAL, private artifacts, temporary bytes, journal and reservations. Cleanup cannot delete protected evidence. Release reservation only after durable publication or completed reconciliation; replenish/test emergency capacity before new work. X74 exhausts artifact, WAL, temp and journal capacity independently, including after normal admission.

# Analysis and coherent source acquisition

**Owner:** Source snapshots and bounded analysis. [source-snapshot.schema.json](../../specs/project/source-snapshot.schema.json) records immutable identity; [CODE-CLASSIFICATION](CODE-CLASSIFICATION.md) owns omission obligations.

## Acquisition before Gate 1

Acquire via immutable VCS object export (including dirty/untracked content only under quiescence), filesystem snapshot, or quiesced copy. A live repeated scan is not an immutable snapshot: ABA changes and symlink swaps can evade it. If no consistent acquisition mechanism/quiescence is available, block snapshot publication. Operator quiescence is explicit recorded evidence, not assumed inactivity.

Define boundary including application/configuration/templates/assets/dependency/framework content, external generated sources and operational inputs; exclude only identified engine workspace/VCS/volatile output paths with reviewed rationale. Symlinks must be contained and acquisition must use race-resistant descriptor-relative access; ambiguous/escaping/swap paths fail. Record sorted POSIX paths, raw file SHA256, executable bit, symlink target and content-root hash. Content identity alone does not establish filesystem behavioral identity. Relevant metadata is captured separately under the filesystem semantic profile below. Snapshot storage is private/owner-only, sealed read-only and verified before use; original source is never instrumented.

Publish complete manifest/profile/closure evidence atomically with registered artifact identity and source snapshot row; never expose a partial snapshot as sealed. Detection/analysis/probes use the sealed bytes. Live-source drift is a separate monitored identity and triggers review/update, not mutation of existing baseline. Reopening verifies bytes, ownership and snapshot boundary. Rebase/update produces a successor snapshot with explicit changed manifest; typed validity determines reuse and new Gate 1 approval.

## Evidence and limits

Measure runtime/framework/dependencies/config profiles where safely possible; keep human assertions separate. Source version labels cannot establish vendor modifications or installed bytes. Parse status is PARSED/PARTIAL/UNSUPPORTED/FAILED per region. Dynamic dispatch/include/autoload/reflection/magic/eval/generated code, service locators/config/database-driven routing and production-only paths remain unresolved unless a supported safe mechanism records them. Runtime hits prove presence, not absence.

Operational roots include HTTP/admin/webhooks, CLI, cron/timezones, queues/workers, plugins, feature flags and environment/config/data-controlled loading. The customer supplies/reviews a versioned operational inventory; omitted root categories prevent boundary closure. No static evaluator executes PHP/dependency scripts. Unknown/unsupported regions prevent affected DEAD proofs and behavior acceptance.

Acceptance must falsify coherent acquisition with source mutation/ABA/symlink swaps; reject unsafe acquisition. Fixtures include DB-driven route, cron-only helper, modified vendor, mixed versions, dynamic method/class/include, reflection/autoloader/eval, unavailable extension, flag profiles and generic source. Successful parsing is not coverage completeness.

Assessment uses these same Source/Application IR, uncertainty and finding records. Partial parsing, missing operational roots or runtime inability can support a useful classified assessment report, while still blocking affected DEAD proofs or executable contracts. There is no separate assessment parser or analysis table. [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md) owns the checkpoint; later migration reuses only currently valid records with typed witnesses.

## Filesystem semantic profile and execution admission

Source snapshot 2.1 retains content identity and adds filesystem_profile_ref and execution_admission_ref. A filesystem profile lists exactly once each supported property: permissions, executable bits, mtime, ownership, hard links, symlink targets, case sensitivity, path resolution, working directory and directory ordering. Each property is REQUIRED (capture/preserve), VIRTUALIZED (pin mechanism and controlled semantics), IRRELEVANT (evidence under approved scope) or UNSUPPORTED (affected behavior unavailable). Capture metadata in the same coherent acquisition boundary as bytes; re-statting a live source later is not sufficient. A VCS export without original mtimes cannot assert their preservation. Private captured-state artifacts record path/value evidence; limits/platform gaps remain visible.

REQUIRED/VIRTUALIZED needs capture/mechanism and control evidence before execution; IRRELEVANT is derived from bounded dependency coverage, never an optimistic default. Every property has rationale/evidence and residual limitations. Same bytes with changed relevant mtime/permissions/profile yields a successor baseline/approval and invalidates affected observations. Unavailable relevant properties still permit static assessment but cannot support behavioral identity. Content hashes do not include all host metadata indiscriminately.

SECURITY owns execution-admission 1.0 and separately identified execution-variant 1.0. Their original_manifest_ref identifies the original bytes without referencing the future snapshot wrapper, so SOURCE approval closure stays acyclic. Variants never replace the private original. Gate 1 binds snapshot→filesystem/admission/variant closure. A changed execution variant/admission/profile creates a successor baseline and review; typed positive coverage may retain unrelated analysis.

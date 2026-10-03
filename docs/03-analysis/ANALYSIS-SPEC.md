# Analysis and coherent source acquisition

**Owner:** Source snapshots and bounded analysis. [source-snapshot.schema.json](../../specs/project/source-snapshot.schema.json) records immutable identity; [CODE-CLASSIFICATION](CODE-CLASSIFICATION.md) owns omission obligations.

## Acquisition before Gate 1

Acquire via immutable VCS object export (including dirty/untracked content only under quiescence), filesystem snapshot, or quiesced copy. A live repeated scan is not an immutable snapshot: ABA changes and symlink swaps can evade it. If no consistent acquisition mechanism/quiescence is available, block snapshot publication. Operator quiescence is explicit recorded evidence, not assumed inactivity.

Define boundary including application/configuration/templates/assets/dependency/framework content, external generated sources and operational inputs; exclude only identified engine workspace/VCS/volatile output paths with reviewed rationale. Symlinks must be contained and acquisition must use race-resistant descriptor-relative access; ambiguous/escaping/swap paths fail. Record sorted POSIX paths, raw file SHA256, executable bit, symlink target and content-root hash. Metadata is not identity. Snapshot storage is private/owner-only, sealed read-only and verified before use; original source is never instrumented.

Publish complete manifest/profile/closure evidence atomically with registered artifact identity and source snapshot row; never expose a partial snapshot as sealed. Detection/analysis/probes use the sealed bytes. Live-source drift is a separate monitored identity and triggers review/update, not mutation of existing baseline. Reopening verifies bytes, ownership and snapshot boundary. Rebase/update produces a successor snapshot with explicit changed manifest; typed validity determines reuse and new Gate 1 approval.

## Evidence and limits

Measure runtime/framework/dependencies/config profiles where safely possible; keep human assertions separate. Source version labels cannot establish vendor modifications or installed bytes. Parse status is PARSED/PARTIAL/UNSUPPORTED/FAILED per region. Dynamic dispatch/include/autoload/reflection/magic/eval/generated code, service locators/config/database-driven routing and production-only paths remain unresolved unless a supported safe mechanism records them. Runtime hits prove presence, not absence.

Operational roots include HTTP/admin/webhooks, CLI, cron/timezones, queues/workers, plugins, feature flags and environment/config/data-controlled loading. The customer supplies/reviews a versioned operational inventory; omitted root categories prevent boundary closure. No static evaluator executes PHP/dependency scripts. Unknown/unsupported regions prevent affected DEAD proofs and behavior acceptance.

Acceptance must falsify coherent acquisition with source mutation/ABA/symlink swaps; reject unsafe acquisition. Fixtures include DB-driven route, cron-only helper, modified vendor, mixed versions, dynamic method/class/include, reflection/autoloader/eval, unavailable extension, flag profiles and generic source. Successful parsing is not coverage completeness.

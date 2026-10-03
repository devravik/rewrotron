# Content-addressed artifact store

Durable bytes live at `.rewrotron/artifacts/sha256/<first-two-hex>/<remaining-hex>`. Public references are `sha256:<64 lowercase hex>`. SQLite records digest, relative path, media type, size, creation time, redaction manifest and entity links. Human-friendly trace/report names are exported indexes, not canonical storage paths.

## Write protocol

Redact before any staging write. Stream sanitized bytes to an owner-only temporary file within the store, compute digest and size, fsync, and atomically rename within the same filesystem. If that digest exists, verify its size/hash and reuse it. Sync the directory where supported. Then insert metadata/links/events in a SQLite transaction. If the DB commit fails, an unreferenced artifact is harmless and can be reconciled later.

Readers resolve only registered references, reject path traversal/symlink escapes, and validate hashes before using evidence. Corrupt or missing bytes produce `ERR_ARTIFACT_CORRUPT` and prevent a current pass/readiness claim. No raw artifact path supplied by an agent is opened directly.

## Retention and privacy

Default retention preserves every artifact reachable from current or historical confirmations, baselines, locked contracts, plans, runs, repair evidence and reports. Age or a PASS verdict alone never authorizes collection. Temporary/unreferenced writes can be collected only after a documented reconciliation grace period (24 hours by default) and active operation checks.

Customer archive/deletion policy is a separate reviewed workspace lifecycle; no public `clean` command is part of the initial catalog. Export builds a reviewed manifest, sanitizes credentials/PII, excludes correlation keys and private fixtures, and verifies destination ownership. A reusable diagnostic report is smaller than a raw customer archive.

Backups include SQLite, referenced artifacts, configuration and the private correlation key only when explicitly requested through a secure local backup policy. WAL-consistent backup and restore procedures are in [RECOVERY](RECOVERY.md).

## Decision-system views

Detection, source dependency inventory, recommendations, frozen target revisions, recursive resolution and feasibility are registered immutable artifacts. Human-readable `.rewrotron/{detection,recommendations,source-dependencies,dependencies,plan,approvals}.yaml` files are regenerable views. The optional `target.yaml` is user-owned input, never a generated recommendation view. `approvals.yaml` mirrors SQLite history and current subject validity; importing or editing it cannot create a confirmation.

Store bytes before transactional metadata/events. A crash during YAML export is recovered by regenerating from committed references. Target-only changes preserve unaffected source artifacts and immutable history; old target evidence stays retained and visibly stale. [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) defines file precedence and ownership.

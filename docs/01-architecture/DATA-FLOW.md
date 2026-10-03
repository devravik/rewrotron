# Data flow and publication

[SYSTEM-DESIGN](SYSTEM-DESIGN.md) owns the pipeline; [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) owns subject parents/hashes; [RECOVERY](../08-storage/RECOVERY.md) owns publication.

Each phase consumes immutable refs, captures exact source/plan/profile/resolution/tool identities and produces immutable artifacts plus current projections/events. Artifact bytes precede transactional metadata; a crash leaves reconcilable unreferenced bytes. Live source is not analysis input; sealed snapshot is. Recommendations cannot write intent; progress cannot alter approved decisions. Target model exists before approval, activation afterward. Unit correctness differs from project readiness. Event-window cursors describe committed events, not historical mutable state. JSONL is a regenerable export only.

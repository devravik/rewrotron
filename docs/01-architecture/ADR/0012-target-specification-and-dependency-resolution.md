# ADR 0012: User-owned targets and capability-based dependency resolution

- **Status:** Accepted specification decision; implementation pending
- **Date:** 2026-10-03
- **Basis:** Repository owner requested first-class target configuration and dependency resolution
- **Related:** ADRs 0009, 0010 and 0011; no existing confirmation gate is removed

## Context

Selecting a framework alone leaves runtime, infrastructure, compatibility and dependency decisions to individual agents. Package-name mapping misses actual usage and cross-language native implementations. Mutable recommendations can drift from customer intent and stale target work can accidentally appear current after a version change.

## Decision

Separate observed source facts, user target intent, immutable migration plans and executed results. Use root `rewrotron.yaml` as primary editable project intent, with optional exclusive `.rewrotron/target.yaml`. Preserve the prior project YAML input when no root file exists. Generated detection, recommendation, dependency, plan and approval views never overwrite intent or grant authority.

Capture a validated immutable target revision before planning. Gate 1 still protects deep source analysis; Gate 2 confirms the exact target revision, resolutions and plan together; Gate 3 confirms current release evidence. Approval stays in SQLite `user_confirmations`, not YAML. Explicit edits create successor intent requiring a new plan; recommendations cannot silently lower approved versions.

Use neutral `internal/detection`, `internal/recommendations`, `internal/target` and `internal/dependency` services. Existing `internal/migration`, `internal/core` and `internal/knowledge` own planning, gate enforcement and learning respectively. Ecosystem parsing/API metadata remains behind adapter ports. Reuse the artifact store, verifier and approval ledger.

Resolve dependencies by exact inventory, recursive constraints, actual usage and required capabilities. Strategies are DIRECT_UPGRADE, ADAPTER, REPLACEMENT, REIMPLEMENT and RETAIN; BLOCKED is an unresolved status. Metadata/API compatibility, capability coverage and behavioral verification are separate. Unknowns and dormant usage cannot be treated as dead baggage.

Record input subjects and reverse dependency links so target changes invalidate affected planning/build/verification while preserving unrelated source work and immutable history. Existing artifact/link/event/finding tables store the first design; no SQL schema modification is made here. Shared target and decision schemas extend the undeployed contracts additively; old documents remain readable, but incomplete historical plans must be rebuilt before execution under these requirements.

## Consequences

Users can review reproducible target and dependency decisions across sessions. Framework-neutral planners can support future languages without N×M package translators. Additional feasibility and characterization work is required, and unknown metadata may block dependent scope. Offline/local evidence and controlled acquisition keep SaaS and LLMs optional. Global resolution knowledge remains opt-in and evidence-scoped.

The [target specification](../../04-migration/TARGET-SPECIFICATION.md) and [dependency resolution](../../04-migration/DEPENDENCY-RESOLUTION.md) own detailed rules. A future interactive `configure` command requires its own synchronized registry/schema work; this decision does not add an undocumented public command.

# ADR 0013: Transparent execution as a core product contract

- **Status:** Accepted specification decision; implementation pending
- **Date:** 2026-10-03
- **Basis:** Repository owner requires continuous user supervision of long-running work
- **Related:** ADRs 0010–0012 and existing event, mission and recovery contracts

## Context

Behavioral reconstruction can take hours. A final result or opaque progress message does not explain discoveries, target changes, unresolved behavior or needed human decisions. Independent client logs can diverge from actual engine state, and guessed percentages can obscure failed verification or approval requirements.

## Decision

Transparent execution and behavior preservation are central product promises. Every long-running operation emits typed sanitized events and exposes current stage/task, factual reason, result, blockers, next action and user needs. `internal/execution` owns shared projections/explanations/progress, with one SQLite event ledger and existing operation/mission/finding/artifact/confirmation records.

Visible stages DETECT through FINALIZE are labels on repeatable/concurrent instances, not replacements for the durable lifecycle. OBSERVED/DECIDED/ACTION/RESULT distinguish facts, proposals, work and outcomes. Public explanations cite evidence and omit hidden reasoning. Percentages require persisted numerator/denominator/scope/basis; overall progress requires a versioned weighted work manifest and never substitutes for release gates.

CLI/MCP/desktop clients read the same committed state and timeline with sequence cursors. Add optional execution state to project snapshots and typed execution read/control contracts. Human `status --watch` preserves existing non-watch JSON envelopes; machine clients poll cursor-based reads. Durable pause/continue is separate from reconciliation and approval. Existing three gates remain authoritative; an event or UI button cannot forge approval. Initial MCP control is read-only; local human commands own pause/continue.

Existing events JSON payloads and artifact links store these records; no persistent table/column or initial DDL change is introduced. Long operations publish liveness at bounded intervals, acknowledge pause only at safe boundaries, and preserve history on drift/restart. Instrumentation is required in the first engine milestone; richer desktop rendering can follow without changing the event model.

## Consequences

Users can inspect current work, why it matters and what changed after returning, without relying on an agent conversation. Every producer must instrument real boundaries and handle event-store failure safely. Heartbeat, percentage, outcome and verification remain distinct. Replayable views improve diagnosis but do not restore external effects or prove equivalence. Live client/runtime validation remains an implementation deliverable.

[EXECUTION-OBSERVABILITY](../../06-agent/EXECUTION-OBSERVABILITY.md) owns the detailed contract and required negative fixtures.

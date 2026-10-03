# ADR 0014: Independent learning and reversible evidence consumption

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [LEARNING-SYSTEM](../../07-learning/LEARNING-SYSTEM.md). This ADR records the decision, not a duplicate field specification.

## Context

Repeated correlated oracles and degraded capture can promote false techniques.

## Problem

Correlated fixtures can promote a wrong oracle; open global prose cannot establish a versioned/retractable machine lifecycle. Historical RWT sequencing obscures when those missing contracts must exist.

## Decision

Retain project-local evidence first, global advisory opt-in learning only after M7 independent holdouts/mutants/quality evidence. M7 must specify typed revision/applicability/evaluation/consumer/retraction/outbox records and migrations before exposing promotion. Nothing in learning establishes compatibility, changes a target or bypasses approval.

## Alternatives considered

Promotion by context count/success rate; live registry consumption; independent pinned evidence/holdouts.

## Why rejected

Correlated oracles and degraded tools can inflate success; live revisions defeat reproducibility and transitive retraction.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Existing project lessons and typed validity links remain baseline. Global machine contracts are explicitly DEFERRED before M7, not claimed to be represented by a generic artifact alone. Every consumer must pin revisions and be discoverable by reverse dependency lookup; disclosure/promotion requires privacy controls.

## Migration/compatibility implications

No global tables/API invented during this documentation pass; no baseline migration depends on them. Retracted revisions remain immutable history; current consumers become stale, not erased. Successor knowledge requires explicit applicability and new evidence.

## Acceptance experiment / falsification test

X37–X39: correlated wrong oracle, tool that drops hard traces, leaked secret and failed holdout must reject promotion; crash/outbox/retraction finds every consumer while preserving old pinned approvals.

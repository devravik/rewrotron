# ADR 0010: Durable exact-subject gates and execution boundary

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../02-domain/CONFIRMATION-PROTOCOL.md) contains normative field/service details.

## Context

Approvals reconstruct retained bytes and must name relevant execution choices.

## Problem

Byte-only snapshots and secret substitutions can change behavior without changing a reported original tree.

## Decision

Keep SOURCE/PLAN/RELEASE canonical preimage 2.0. SOURCE binds snapshot 2.1 filesystem/admission/variant closure. PLAN binds exact execution/schema/target/decision/performance closure. Witnesses qualify actual later executions. Shared assessment is never approval.

## Alternatives considered

Hash mutable registries; hash bytes alone; retained exact subjects with relevant profile and variant references.

## Why rejected

Registries drift; content-only identity misses relevant filesystem/credential variants. Expected future execution witnesses cannot exist at approval time.

## Consequences

Relevant profile/admission changes require successor review; original private bytes and old approvals remain reconstructable.

## Machine-contract impact

Snapshot 2.1 plus filesystem-profile/execution-admission/execution-variant 1.0 via existing snapshot_ref. Approval-subject 2.0 hash fields and all existing vectors remain unchanged.

## Migration/compatibility implications

No rewriting migrations 0001/0002 or historical approval bytes. Changed transitive refs produce new subjects; historical execution closure is not silently reinterpreted.

## Acceptance experiment / falsification test

X01/X02/X55/X58/X68/X69: reconstruct old hashes after schema change, reject metadata/variant drift and assessment-as-token.

# ADR 0010: Durable exact-subject human gates

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [CONFIRMATION-PROTOCOL](../../02-domain/CONFIRMATION-PROTOCOL.md). This ADR records the decision, not a duplicate field specification.

## Context

Implicit hash construction/registry lookups and incoherent scans break approval identity.

## Problem

Pinning adapters without core interpreter/schema bytes lets the same approved rule execute with different semantics after upgrade. A mutable performance choice or assessment report can also be mistaken for implementation authority.

## Decision

Keep canonical SOURCE/PLAN/RELEASE subjects and exact local human gates. SOURCE and PLAN bind the core execution manifest via the existing RUNNER role. PLAN binds immutable assessment selection/basis and performance policy through decision_refs. Reconstruct retained bytes, independently check current execution prerequisites, and require successor approval for changed choices.

## Alternatives considered

Registry-derived hash; mutable YAML approval; retained versioned canonical subject. Version-label-only engine pins versus exact executable/build/schema closure.

## Why rejected

Registry changes make old preimages irreconstructable and edited YAML cannot establish human intent. Version labels alone cannot distinguish changed interpreter/schema bytes, so they cannot preserve the semantics of an old approval.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Execution-manifest 1.0 pins engine/build/dependency/schema/canonicalization/qualification bytes. Approval-subject 2.0 preimage remains unchanged; semantic checks require exact core binding and typed policy/assessment refs. Plan envelope 2.1 retains checkpoint references. New vector exercises existing binding fields; previous nine vectors remain unchanged.

## Migration/compatibility implications

Historical subjects remain reconstructable under retained schema/canonicalizer bytes; missing old executable blocks execution, not history. Never fetch/run an imported historical binary automatically or manufacture new receipts. Migrations 0001/0002 are untouched.

## Acceptance experiment / falsification test

X01/X02/X16/X55/X58/X59: reconstruct after restart/registry upgrade; replace core/schema bytes with the same version names; change performance policy or substitute assessment hash for plan token. Changed execution must block or require the applicable successor gate; original approved hash stays identical.

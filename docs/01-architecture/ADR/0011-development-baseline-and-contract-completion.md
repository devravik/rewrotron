# ADR 0011: Falsification before broad implementation

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [DEVELOPMENT-PLAN](../../11-roadmap/DEVELOPMENT-PLAN.md). This ADR records the decision, not a duplicate field specification.

## Context

Schema validation and read-only fixtures do not validate reconstruction.

## Problem

Generic milestone names and schema-only happy paths hide the hardest premise: plausible wrong reconstructions must be detected. Assessment cannot require successful executable planning, nor become a second analysis engine.

## Decision

M0 durable foundations immediately followed by M1 stateful wrong-target falsification; every milestone has concrete outcomes, prerequisites and evidence. Keep assessment as a successful shared-workflow checkpoint with explicit continuation and separate Gate 2. Add X56–X60 for execution/capture/identity/policy/interface gaps without postponing M1 behind broad support.

## Alternatives considered

Broad adapters first; read-only happy-path milestone; early stateful mutation falsification.

## Why rejected

Happy paths cannot detect a plausible wrong reconstruction; broad adapters compound an untested oracle.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Current schemas distinguish historical readers from execution admission; typed execution/qualification/performance artifacts use existing immutable storage. TESTING owns X01–X60; assessment X41–X55 remains its fifteen-test checklist. CLI/MCP share the canonical ERR_UNSUPPORTED_CAPABILITY code. Global contracts/profile DSL are explicit M7/M8 prerequisites.

## Migration/compatibility implications

Historical RWT identifiers are provenance aliases, not a competing task queue. No implementation or runtime evidence is asserted by specification edits. A failed M1 semantic mutant stops expansion; assessment findings/history remain retrievable.

## Acceptance experiment / falsification test

All seventeen M1 premise checks plus applicable X41–X58. Good synthetic target passes scoped acceptance, wrong mutations fail for their intended reasons, incomplete capture UNKNOWN, checkpoint termination/restart/reuse have no implementation authority. X59/X60 are assigned to their later claim gates.

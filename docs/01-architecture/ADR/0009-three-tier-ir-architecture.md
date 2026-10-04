# ADR 0009: Neutral IRs and qualified target lowering

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../04-migration/TARGET-ARCHITECTURE.md) contains normative field/service details.

## Context

Neutral syntax, application structure, behavior and target organization have distinct owners.

## Problem

Reference HTTP/numeric versions and prose provider defaults can become unintended universal assumptions.

## Decision

Keep three IRs and Target Implementation Model. Source/target framework skills package independently selected adapters through existing ports, categories and qualification records under SKILL-SPEC; no required source→target pair belongs in the core. CI3 → Laravel is Reference Profile #1 intended for first qualification, currently SPECIFIED. Use exact schema/profile qualification, opaque provider-qualified releases and digest identity; explicit response/session obligations constrain lowerings. Wider language/invocation facets wait for M8.

## Alternatives considered

N×M translators; universal unrestricted IR/language; narrow neutral records plus capability-qualified lowerings.

## Why rejected

Pairwise translators couple frameworks; universal semantics would delay falsification and falsely imply unsupported capabilities.

## Consequences

Current HTTP subset and unsupported computations remain explicit. Generated code is never verification evidence merely because a provider emitted it.

## Machine-contract impact

Resolution 2.1 includes provider interpretation and toolchain components; context/response 1.0 and target-model capability references carry obligations. Manifest 2.1 retains schema/qualification pins.

## Migration/compatibility implications

No automatic range widening or loss of unsupported source initialization. Historical schemas retain exact bytes; additional facets require versioned contracts.

## Acceptance experiment / falsification test

X40 adds compatible/incompatible directional package composition, reference regression and generator self-certification mutants at M8. X40/X58/X67/X75: preserve unsupported source facets, reject unqualified profile, catch dropped cookies and accept digest-pinned nonnumeric release.

Target-model 2.1 lowering_bindings make provider manifest, qualification, shape and obligation identities explicit; plain lowering_choice is never executable admission. X67 also rejects an unqualified/mismatched provider binding before generation.

Packaging uses existing manifest 2.1/qualification 1.0 and TIM lowering_bindings; no new service, registry or runtime target is introduced. Adding packages does not authorize an unsupported neutral operator. JEV is deferred and provides no architectural prerequisite.

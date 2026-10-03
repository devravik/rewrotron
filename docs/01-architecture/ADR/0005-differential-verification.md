# ADR 0005: Independent capture and downstream correspondence

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../05-verification/DIFFERENTIAL-TESTING.md) contains normative field/service details.

## Context

Discovery exists before any target/case and independent replay can diverge.

## Problem

Mandatory paired identity blocks source capture; stale builds and positional step matching can falsely pass.

## Decision

Publish side-local identities/context/effects first. Derive pairwise bijection and occurrence alignment afterward. STOP_AT_FIRST_DIVERGENCE preserves known MISMATCH. Actual process/build/endpoint witnesses and mechanism controls qualify every side.

## Alternatives considered

Placeholder paired target values; compare arrays by position; trust claimed tree; explicit source-local capture plus witnessed pairing.

## Why rejected

Placeholders invent future evidence, positions misalign branches, and a tree hash does not identify loaded code.

## Consequences

Read-only comparison needs independent reset, server-process capture, typed actual effects and trusted principal evidence; incomplete capture remains UNKNOWN.

## Machine-contract impact

Trace/effects/report/identity-map 2.1; execution-witness/local-identities/step-occurrence/occurrence-pairing/effect-payload/effect-occurrence 1.0. Attestation 2.1 and qualified data-model 2.0 remain required.

## Migration/compatibility implications

Old captures stay immutable history; no target values are inserted later. Old reports cannot become current by schema conversion.

## Acceptance experiment / falsification test

X62/X64/X65/X66/X70/X72/X80: publish before target, diverge branches, repeat requests, serve stale code, miss trigger effects and compare two SOURCE sides; expected infrastructure failures never count as semantic kills.

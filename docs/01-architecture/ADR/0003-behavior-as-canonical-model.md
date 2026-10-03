# ADR 0003: Executable behavior with isolated model execution

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../02-domain/BEHAVIOR-MODEL.md) contains normative field/service details.

## Context

Effectful models and real applications can both insert rows; shared state would manufacture parity.

## Problem

Namespaces, expected state and fixture labels lacked phase/provenance; headers and session continuity lacked executable obligations.

## Decision

Use the BEHAVIOR-MODEL role matrix and EVALUATION-CONTEXT ABI. Shadow actions never touch measured stores; checker contexts are actual-only. Typed response/session lowering is a bounded qualified capability.

## Alternatives considered

One interpreter with writable fixture handles; generation-only prose; isolated shadow simulation plus read-only checking.

## Why rejected

Shared handles can supply omitted target effects; prose hides response/session obligations. A general language is unnecessary for the narrow fixture.

## Consequences

Extra role/context controls are required before M1. Unsupported operators/providers remain UNKNOWN/blocked. Actual principal cannot derive from actor labels.

## Machine-contract impact

evaluation-context/principal-fixture/initial-state/response 1.0; additive pinned semantics response/session actions; existing artifact storage, no SQL rewrite.

## Migration/compatibility implications

Schema bytes and rwt-model-roles-1 qualification are pinned. Historical closures are retained; new capability consumption needs successor qualification/approval.

## Acceptance experiment / falsification test

X61/X63/X67/X77 plus X03/X56: absent target INSERT stays absent; poisoned expected owner cannot pass; redirects/cookies/session lifecycle are measured.

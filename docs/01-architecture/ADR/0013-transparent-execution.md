# ADR 0013: Owned execution and emergency persistence

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03 against review-2. [Canonical owner](../../08-storage/RECOVERY.md) contains normative field/service details.

## Context

Long work needs an engine owner and truthful current-state/event-window reporting.

## Problem

Disk exhaustion can prevent both an effect receipt and the event explaining its loss; heartbeat can conceal retries without progress.

## Decision

Keep foreground versus explicit supervisor ownership and state N/window M. Reserve bounded artifact/receipt budgets and a preallocated private recovery journal before effects. Stop on persistence failure and reconcile before retry.

## Alternatives considered

Best-effort stderr only; independent second event authority; reserved journal as reconciled safety input.

## Why rejected

Stderr alone loses receipts; a second authority conflicts with SQLite and cannot approve work. Device failure remains an explicit limitation even with reservations.

## Consequences

Emergency facts are marked uncommitted. Missing receipt stays UNKNOWN, no blind retry/PASS. Startup tests reservation durability and ownership.

## Machine-contract impact

recovery-journal 1.0; existing event/state/catalog schemas carry canonical status/errors. RECOVERY fixes slot/sequence/reconciliation; no new SQL authority or rewritten migration.

## Migration/compatibility implications

Existing committed sequences remain unchanged. Journal import creates ordinary canonical events only after reconciliation; old N is never fabricated current state.

## Acceptance experiment / falsification test

X26/X27/X74: disconnect client, exhaust WAL/artifact/receipt slots, crash before ack; preserve pause/approval, distinguish emergency facts, and reconcile without duplicate effects.

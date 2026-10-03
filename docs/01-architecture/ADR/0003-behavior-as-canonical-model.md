# ADR 0003: Executable behavior as canonical obligation

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [BEHAVIOR-MODEL](../../02-domain/BEHAVIOR-MODEL.md). This ADR records the decision, not a duplicate field specification.

## Context

Prose/open JSON cannot establish executable business rules or unseen behavior.

## Problem

A valid-looking contract cannot express which candidate row an UPDATE selects or bind the generated ID of an INSERT. Undefined short circuit, missing-value and terminal-transaction rules let two evaluators disagree while both claim parity.

## Decision

Keep the bounded closed vocabulary and explicit branches. Specify typed row context, optional write-result binding, single-assignment bindings, pointer/null/decimal rules, terminal transaction handling and positive async boundaries in BEHAVIOR-MODEL. Evaluator and target must qualify the exact consumed schema/operator bytes; unsupported semantics yield a named blocker/UNKNOWN, not an else branch or an implicit commit.

## Alternatives considered

Prose/open JSON evaluator; unrestricted DSL; small closed typed vocabulary.

## Why rejected

Prose has no deterministic interpreter and unrestricted evaluation adds unsafe unbounded semantics.

## Consequences

Implementation has an explicit narrow admission boundary and additional falsification evidence obligations. Missing qualification or required capture blocks the affected claim; historical bytes stay retained. Documentation validation does not establish runtime maturity.

## Machine-contract impact

Add row reference namespace and optional db_write.bind to semantics.schema; require positive SINK_ACK_QUIESCENCE duration in semantics/workflow schemas. Qualified data models identify every DB action. Service type/context/bound/transaction checks remain mandatory beyond JSON shape; no new SQL table is needed.

## Migration/compatibility implications

Retain old artifact bytes and their pinned interpreter/schema closure; additive fields do not retroactively change old approvals. Changed evaluator/schema execution choices require successor qualification and applicable approval. Descriptive v1 rules remain nonexecutable.

## Acceptance experiment / falsification test

X03–X08/X56: insert a row, bind its actual ID into response and dependent effects, then mutate predicate/principal/write/identity. Test wrong type, overflow, missing value, short circuit, invalid pointer, duplicate binding and terminal open transaction. Reject for the expected semantic reason, not unrelated boot failure.

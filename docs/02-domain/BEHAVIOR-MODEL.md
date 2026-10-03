# Three-tier behavior model

**Status:** Canonical domain specification. [IR-CONTRACTS](IR-CONTRACTS.md) defines records and provenance; [specs](../../specs/README.md) defines executable validation.

## Three distinct representations

| Tier | Describes | Producing responsibility | Cannot establish alone |
| --- | --- | --- | --- |
| Source IR | Files, syntax declarations, symbols, spans and typed call/include/reference edges | Source parser adapter | Framework dispatch or actual execution |
| Application IR | Entrypoints, handlers, services, workflows, data models and dependencies | Source adapter interpretation plus neutral graph services | Complete business behavior |
| Behavior IR | Operation rules, preconditions, concrete cases, state/effects, outputs and variability policies | Evidence-backed behavior services | All possible unseen input outcomes |

A source adapter interprets its framework before emitting neutral structure. The core does not infer CI3 loader semantics or Laravel middleware. Target adapters consume Behavior IR and discovered data models through a confirmed architecture plan.

## Graph relationships

Entrypoints can reach several operations; operations can be shared across entrypoints/workflows. A workflow references a bounded ordered sequence of operation cases and explicit value bindings. Contracts govern operations, and concrete cases bind exact versions, fixtures, input values and assertion policies. This is a graph, not a forced one-route/one-contract tree.

Entrypoints are HTTP, CLI, cron and event. HTTP dispatch requires method/path; other types require command/schedule/channel fields. Unsupported execution remains inventoried and unverified. Only explicit reviewed scope can exclude it from a migration claim.

## Contract content

Every contract names operation/version, dispatch, inputs, evidenced rules, required capture dimensions, DB/session/external effects, response/outcome assertions, evidence IDs and open questions. Authentication, authorization, error behavior and supported input boundaries are part of the public contract. Rules are `DECLARATIVE`, `OBSERVED_ONLY` or `UNRESOLVED`; unsupported representations block target generation.

Concrete executable cases supply fixture artifacts, principal/secret references, exact payloads and expected results. An input-only observation scenario supplies no claimed output. Source repetitions establish candidate assertions; three repeats do not prove all unobserved behavior.

## Canonical example

This synthetic contract is mirrored in [the machine example](../../specs/examples/behavior-contract.json) and validated by the documentation check. Evidence IDs and principal IDs reference synthetic fixture records. It illustrates a target-supported declarative rule, not a general expression evaluator.

```json
{
  "schema_version": "1.0",
  "operation_id": "orders-read",
  "version": "1.0.0",
  "description": "Read an order owned by the authenticated principal",
  "entrypoint": {
    "type": "http",
    "method": "GET",
    "path": "/orders/42",
    "content_types": []
  },
  "authentication": {
    "required": true,
    "principal_fixture_id": "customer-a"
  },
  "authorization": {
    "rule_id": "owner-only"
  },
  "inputs": {
    "parameters": []
  },
  "business_rules": [
    {
      "id": "owner-only",
      "description": "Return the order only to its owner",
      "representation": "DECLARATIVE",
      "evidence_ids": [
        "ev_owner"
      ],
      "data": {
        "operator": "principal_equals_row_owner"
      }
    }
  ],
  "database_effects": {
    "transactional": false,
    "mutations": []
  },
  "session_effects": {
    "mutations": []
  },
  "external_effects": {
    "events": []
  },
  "response": {
    "status": 200,
    "headers": {
      "content-type": "application/json"
    },
    "body": {
      "type": "json",
      "assertions": [
        {
          "json_pointer": "/order_id",
          "equals": 42
        }
      ]
    }
  },
  "evidence_ids": [
    "ev_owner"
  ],
  "required_dimensions": [
    "http",
    "body",
    "authentication",
    "authorization",
    "database",
    "session",
    "external"
  ],
  "open_questions": []
}
```

A failed ownership case is a separate concrete case governing the source's actual forbidden response. The engine must not invent Laravel's default 403/message if the source behaves differently.

## Business-rule completeness and portability

Free-text descriptions, SQL-like predicates and `{computed.total}` placeholders are explanatory; they do not automatically become executable programs. Each required computation must have structured, evidenced semantics that the adapter can implement or an explicitly accepted target proposal verified by cases. The core never evaluates contract strings as unrestricted SQL, PHP or shell.

Preserve observable coercion, money/decimal arithmetic, nulls, ordering, validation/error text, principal/permission effects, transactions, timezone, file and integration behavior. Target clean architecture is an implementation choice constrained by these contracts. Adding a new adapter requires verifying its supported capability intersection against the same corpus.

## Immutability and approval

Changed contracts/cases, fixtures, normalization, intentional changes or scope create new immutable revisions and corpus/policy hashes. Gate 2 approves that exact plan subject; old results remain historical. Unknown required behavior/capture is never converted into PASS. See [compatibility](../05-verification/COMPATIBILITY-SPEC.md), [coverage](../05-verification/COVERAGE.md) and [confirmation protocol](CONFIRMATION-PROTOCOL.md).

## Dependency boundaries

Application IR dependency edges establish structure and usage; required capability contracts establish behavior. A package replacement or native implementation must preserve the evidenced input/output/state/error subset actually used by the application. Package metadata and matching APIs do not establish parity. Reuse existing cases and traces for dependency characterization and affected end-to-end workflows; unknown usage/capabilities remain explicit. [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) owns strategy and approval rules; these decisions do not add framework-specific fields to canonical Behavior IR.

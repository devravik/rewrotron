# Coverage, denominators, and acceptance evidence

Percentages must always carry numerator, denominator, scope hash, and evidence age. An empty denominator is `null`/not evaluated, never 100%.

| Metric | Numerator | Denominator |
| --- | --- | --- |
| Entrypoint characterization | In-scope entrypoints with evidenced locked contracts and cases | All confirmed in-scope non-DEAD entrypoints |
| Required case compatibility | Current PASS case results | Required cases in the frozen corpus |
| Unit completion | Current COMPLETED units | Units in the confirmed plan |
| Observation completeness | Completely captured required dimensions | All required case/dimension pairs |
| Omission accountability | Omitted symbols with current valid DEAD proof | All automatically omitted symbols |

Inventory includes ACTIVE, DORMANT, UNKNOWN and DEAD, plus entrypoint types and unsupported capabilities. Runtime absence does not remove an entrypoint from the denominator. A proposed scope exclusion has an ID, reason, risk, user decision and plan hash; it remains visible in reports and does not count as migrated behavior.

A full application claim requires coverage of all agreed active behavior and disclosure of dormant/unknown components. A scoped migration can complete without the whole application being equivalent; the report names that scope. AUTH/permission failures, validation errors, business failures, repeated requests, and session continuity cases are required alongside happy paths.

Gate 3 requires 100% PASS for the confirmed required corpus, zero required UNKNOWN dimensions, no open prohibited mismatch, benchmark acceptance, current source/target/release hashes, an omission audit, and executable rollback/session evidence. “99% verified” alone never grants cutover readiness.

The report includes an inventory reconciliation: every discovered entrypoint maps to a migrated unit, valid DEAD omission, or explicit reviewed exclusion. Any entrypoint without a disposition is a coverage blocker.

## Live progress versus compatibility

Execution work metrics follow [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md): frozen basis/scope/accounting revision, numerator/denominator, evidence sequence and freshness. Optional overall work uses persisted weights and is null before that accounting basis exists. Work performed, discovered entities, completed tasks and current PASS cases are distinct metrics. Retries do not double-count, changed scope can reduce progress, stale PASS leaves the current numerator, and 100% work completion never grants Gate 3.

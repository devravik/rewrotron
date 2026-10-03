# Behavior model and bounded executable semantics

**Owner:** Behavior meanings, executable vocabulary and scenario semantics. [IR-CONTRACTS](IR-CONTRACTS.md) owns identities/provenance; schemas under specs/behavior define shape. Current executable contract/case/scenario records are v2; v1 DECLARATIVE open JSON is historical/descriptive and not an executable rule language.

## Claim terminology

| Term | Establishment |
| --- | --- |
| Observed equivalence | Named observations match under the approved profile/policy; no claim about unseen inputs |
| Contract equivalence | All required assertions/cases in an identified acceptance scope pass with sufficient capture |
| Semantic equivalence | Equivalent under explicitly named semantics and an applicable demonstrated comparator/proof method; not inferred universally from finite cases |
| Unverified behavior | Specified/reachable behavior without current applicable verification |
| Unknown behavior | Insufficient/conflicting evidence or missing semantics/capture |
| Unsupported behavior | Named mechanism/operator/protocol lacks declared implementation capability |
| Intentional change | Human-approved explicit policy/contract difference; separately displayed, never an accidental mismatch masked by a waiver |

PASS/MISMATCH/UNKNOWN remain the three verdicts. UNSUPPORTED and BLOCKED are capability/readiness reasons, not extra successful verdicts. Unsupported in-scope behavior yields UNKNOWN and blocks scoped acceptance. Known violations yield MISMATCH even when other dimensions are UNKNOWN; either blocks. Successful cases do not establish corpus completeness.

## Interpreter vocabulary

[semantics.schema.json](../../specs/behavior/semantics.schema.json) is a closed tagged language, never PHP, SQL, shell or arbitrary expressions. Limits: depth 32, at most 1,000 expression nodes per contract, 100 actions/branches/array elements and 100 executed scenario steps. Limit violations return ERR_CAPABILITY_UNSUPPORTED with node/operator/limit. No dynamic evaluator or unrestricted extension code.

Values: UTF-8 string, boolean, exact integer/decimal represented as canonical decimal strings, null, RFC3339 timestamp instant, bytes by registered ref, typed object and ordered array. Absent reference is MISSING, distinct from null, empty string, zero and false. Reference namespaces are input, actor, state, result, clock and binding, with JSON Pointer and declared type. Wrong type or absent required binding produces UNKNOWN/ERR_UNBOUND_VALUE; it cannot quietly take an else branch. No implicit PHP/Python/JS truthiness/coercion.

Operators: literal/ref; eq with equal declared types (decimals compare exact numeric value); lt/lte for same numeric or string type (strings byte order); before/at_or_before for timestamp instants; in tests typed array membership; and/or/not require boolean operands; exists/is_missing/is_null inspect references; add/sub exact numeric same-type arithmetic; array/object construct values; coerce permits only INTEGER_TO_DECIMAL and CANONICAL_STRING_TO_INTEGER. Integer range is signed 64-bit for executable arithmetic; larger transport IDs remain exact strings and are not coerced. Decimal arithmetic preserves exact rational base-10 values with max 38 digits and 18 fractional digits; exceeding limits returns unsupported, never rounds. Unsupported operators fail schema validation; capability mismatch after valid parsing blocks execution.

An ordered branch chooses the first true boolean guard; last guard must be literal true unless exhaustive disjointness is established by a supported deterministic check. Unknown earlier guard yields UNKNOWN, not fall-through. Actions: assert, set_state, bounded db_read, db_write, begin/commit/rollback, emit, respond, error. DB actions are connection/catalog/table-qualified and translated only through discovered safe identifiers and typed ports, never raw user SQL. The initial evaluator supports one active transaction per connection; nested/distributed transactions are unsupported. begin/commit/rollback IDs must pair; error/exit outcomes define rollback behavior explicitly.

Preconditions run before actions, postconditions at the declared completion boundary, invariants before/after each state transition. Authorization uses explicit actor/state predicates, including negative wrong-principal cases; descriptive authorization text cannot substitute for a predicate. Business decisions involving time use controlled clock/reference predicates, never broad timestamp masking. Emit records effect class, destination, payload and idempotency relationship; the runner enforces safety independently.

## Scenarios and async boundaries

[workflow.schema.json](../../specs/behavior/workflow.schema.json) v2 identifies actors/principal fixtures, initial state, environment/profile, inputs, guarded steps, bindings, transition edges, terminal success/failure/unsupported states, expected effects and bounded execution. Each step names its actor, operation and immutable request. Bindings evaluate before request execution against the preceding committed observations; each names value expression, exact selector and PERCENT_ENCODED/HEADER_VALUE/JSON_VALUE encoding. A selector into raw_target names an explicit `{binding-name}` template marker, a header selector indexes its ordered record, and body_json selector is a JSON Pointer into a declared JSON body. Only declared sites may change; header CR/LF is rejected, percent encoding uses UTF-8 RFC3986 component escaping, and JSON_VALUE preserves typed exact scalar values. Conflicting selectors, undeclared markers, non-JSON body transformation or unavailable observations block. Persist effective request bytes separately; unbound base requests retain original bytes. Each step and transition list requires a final explicit fallback guard or deterministic exhaustiveness evidence. Exactly one next_step_id or terminal_id is non-null. IDs resolve, all planned paths are bounded, and a path exceeding max_executed_steps is unsupported, never successful truncation. The first subset permits an acyclic finite graph or explicitly unrolled paths; unbounded loops/concurrency/streams are rejected.

Observation scenarios carry inputs only and reference a workflow/profile; they cannot assert invented legacy outcomes. Verification cases add reviewed assertions and source evidence. Separate role/account/inventory/payment/feature/time outcomes are explicit paths or separate cases, never flattened into one happy path. Each side maintains its own cookie jars and workflow state; actors cannot share a jar unintentionally.

Completion modes: PROCESS_EXIT, HTTP_RESPONSE for demonstrably synchronous scope, or SINK_ACK_QUIESCENCE with expected receipt count, predicate, timeout and positive quiescence duration. Queue dispatch is not delivery; sink acceptance and worker acknowledgement are separate evidence. Timeout, late uncorrelated effect or missing receipt yields UNKNOWN/MISMATCH as appropriate. General eventual-consistency proofs and arbitrary distributed schedules are deferred.

## Evidence obligations

Every contract/branch/assertion links source evidence and open uncertainty. [Runtime observation](../03-analysis/RUNTIME-OBSERVATION.md) supplies mechanism attestations; [compatibility](../05-verification/COMPATIBILITY-SPEC.md) supplies relation-preserving comparisons. Fixtures, corpus partitions and comparator mutation tests require independent review, so one agent cannot create both a flawed oracle and its own certificate. Unsupported semantics stay explicit in Behavior IR and feasibility; they are not discarded during target lowering.

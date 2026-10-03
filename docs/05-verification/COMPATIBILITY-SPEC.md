# Behavioral compatibility and deterministic verdicts

**Status:** Normative. Applies to every verification run and migration report.

## Claim and scope

A passing finite corpus demonstrates compatibility for those cases, their input/state conditions, required observation dimensions, and approved comparison policies. It is not mathematical proof of every possible execution. A report always names its configuration, plan, corpus, target tree, environment, and policy hashes and discloses uncovered entrypoints and approved intentional changes.

Required dimensions are HTTP status, redirects, structured payload/body, authentication, authorization, session/flash/cookies, committed DB state, external effects, HTML/form behavior, files/assets, and error behavior. Cache semantics and non-HTTP exit codes/events are included when applicable. Missing instrumentation is `UNKNOWN`; an empty effect list means “observed none” only when the completeness status is `COMPLETE`.

## Verdict set and aggregation

The only verdicts are `PASS`, `MISMATCH`, `UNKNOWN`.

- `PASS`: Every required dimension was completely observed and satisfies the versioned policy.
- `MISMATCH`: At least one complete dimension proves a prohibited difference.
- `UNKNOWN`: No proved mismatch, but a required dimension cannot be evaluated safely or completely.

Aggregation precedence is `MISMATCH` > `UNKNOWN` > `PASS`. Reports preserve all per-dimension results even when a mismatch takes precedence. A timeout, malformed trace, unsafe execution, missing source baseline, ambiguous row matching, or unavailable sink cannot pass.

Mismatch classification (`DEFECT`, `EXPECTED_CHANGE`, `UNKNOWN`) is a separate diagnostic label. An intentional API, status, authorization, or business-value change is never made compatible by a generic normalizer. It requires a human-approved contract exception, new policy/corpus/plan hash, and rerun. The report marks `has_expected_changes: true`; strict parity claims disclose the changed behavior. An old result remains immutable.

## Comparison modes

| Dimension | Default | Permitted scoped alternative |
| --- | --- | --- |
| Non-HTTP exit codes/events | Exact declared outcome, stdout/stderr and effects | Explicit supported event assertions; unsupported harness is UNKNOWN |
| Status, error code, deterministic business values | `EXACT` | Explicit intentional-change exception |
| Redirects | Exact path, query multimap and redirect chain | Approved base-origin substitution; path escaping/order remain meaningful unless specified |
| JSON | `SEMANTIC` | Compare objects independent of key order; array order, types, missing/null and values remain meaningful |
| HTML | `EXACT` | Explicit DOM assertions preserving text, form actions/names, accessibility, scripts, assets and whitespace-sensitive elements |
| Headers | Lowercase names, multivalue fields | Per-field transport noise policy; never collapse `Set-Cookie` |
| Cookies | Name/domain/path/flags/expiry semantics and lifecycle | Opaque-value relation rules preserving principal and continuity |
| DB | Typed committed row state | Approved bijective generated-ID mapping applied to all references |
| Files, binary downloads | Exact bytes/hash | Explicit format-specific comparator; unsupported format returns UNKNOWN |
| Side effects | Ordered destination/payload/count | Declared commutative groups with relation checks |

JSON numbers are parsed without float64 rounding. Numerically equivalent decimal spelling may compare equal, but a string `"1"` differs from number `1`; duplicate keys are rejected. No unconditional float tolerance is allowed for money. Arrays retain order. HTML whitespace can affect visible content and must not be stripped universally. SQL text/query count is diagnostic evidence; different SQL can produce the same committed business state.

## Dynamic values

Repeated source runs identify candidate variability, not automatically approved rules. A rule records ID, version, exact paths/selectors, source repeat trace IDs, format validator, comparison operator, relation group, and human approval reference. It is included in the Gate 2 policy subject.

Supported initial operators are format-and-relation checks for opaque identifiers, approved origin mapping, and explicitly bounded time windows. Time bounds derive from the approved environment/workload, not a universal 60-second tolerance. A rule cannot ignore a whole response, principal, amount, permission, count, or DB table.

Opaque values must be nonempty, satisfy format constraints, and preserve equality/inequality relationships within and across workflow steps. CSRF/session values must also preserve security and lifecycle behavior: rejection of invalid/replayed tokens where applicable, cookie flags and rotation, authentication continuity, logout invalidation, and cross-user separation. Replacing every token with one mask would hide these differences.

## Publication and acceptance

A run has execution status `COMPLETED`, `FAILED`, or `CANCELLED` and a verdict. Infrastructure failure produces diagnostic evidence and `UNKNOWN`; the CLI distinguishes operational failure from a completed mismatch. Source/target/corpus/policy drift prevents publishing a current passing run. All agreed cases and required dimensions must pass before a unit completes. Excluded cases and unreachable features remain visible in [coverage](COVERAGE.md).

Required counterexamples include changed amount, wrong user, duplicate side effect, missing DB write, wrong redirect, removed form control, collapsed cookies, swapped generated IDs, unobserved session mutation, timestamp outside window, and loss of transaction rollback. Each must yield MISMATCH or UNKNOWN as appropriate.

## Dependency characterization

A dependency decision has separate package/platform compatibility, used-API coverage and behavioral evidence. A compatible release or complete capability mapping cannot produce PASS without complete observations for the applicable cases. Characterize the application's required subset, including errors/edge cases/state/effects, against the exact source and selected target implementations, then rerun dependent workflows. Use the same verdicts, comparators, scope denominators and immutable run records. Target or transitive dependency changes invalidate affected results; retained source characterization is reusable only with unchanged subjects. See [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md).

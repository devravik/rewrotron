# Runtime observation and instrumentation completeness

**Status:** Normative. Requires valid Gate 1 and the [container profile](../09-security/SANDBOX.md). The original source tree is never instrumented.

## Capture architecture

A reverse proxy records actual HTTP bytes/status/headers while the source adapter's probe records resolved dispatch, DB operations, session/flash/cache mutations and filesystem effects. External effects are intercepted by isolated sinks. A trace ID correlates all components; probe routing metadata is neutralized into Application IR references.

CI3 hook/driver instrumentation belongs in the source adapter. Proxy capture covers early exits and actual outgoing responses; hook coverage can be incomplete for cached responses, constructor failures, `exit`, direct output and shutdown errors. An HTTP response alone cannot establish session/DB completeness.

| Dimension | Evidence | Incomplete case |
| --- | --- | --- |
| Request/response | Proxy byte capture, multivalue headers, redirect steps | Truncation, unsupported protocol, timeout |
| Dispatch | Probe resolved entrypoint/symbol | Custom remap/dynamic code not intercepted |
| DB | Query/transaction diagnostics plus committed snapshots | Unsupported connection/driver/trigger coverage |
| Session/flash/cookies | Driver-level snapshots and lifecycle | Early shutdown, external store missing |
| External effects | HTTP/SMTP/queue sink receipts and network denial | Unknown protocol, blocked unmatched destination |
| Non-HTTP execution | Exit codes/stdout/stderr/event results and declared bindings | Unsupported command/consumer harness |
| Files/cache | Declared-volume and store snapshots | Untracked write path or inaccessible data |

Each required dimension is `COMPLETE`, `PARTIAL`, `UNAVAILABLE` or `NOT_APPLICABLE`. Only a justified contract can use NOT_APPLICABLE. Empty arrays with unavailable instrumentation never mean no effects.

## Trace contract

Traces validate against `specs/schemas/trace-schema.json` and contain source/target environment identity, baseline hash, case/operation IDs, timing, completeness, request/response or non-HTTP execution result, dispatch, state/effect observations, redaction manifest and artifact refs. All bodies use registered content-addressed references after redaction. Header values may be strings or arrays to preserve `Set-Cookie` and duplicate fields.

SQL comments are not inserted universally for correlation because they can alter caching/plans or malformed queries. Prefer probe-side trace context and connection/event metadata. Any instrumentation that changes execution must be detected through an instrumented/uninstrumented control fixture and recorded.

## Baseline observation

Use at least three full source repetitions from reset DB/session/cache/file/sink state. Variation creates candidate rules; approved comparison semantics are defined by [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md). Never auto-mask all tokens or convert a changed value into an accepted modernization.

Cancellation and failures preserve partial sanitized traces and blocker records. Trace publication rechecks source drift and records adapter/probe versions and capture limits. A runtime hit can establish ACTIVE classification; absence cannot alone establish DEAD.

# Existing source async qualification attempt

Owners: [RUNTIME-OBSERVATION causality](../../docs/03-analysis/RUNTIME-OBSERVATION.md#trace-and-causality),
[BEHAVIOR-MODEL completion](../../docs/02-domain/BEHAVIOR-MODEL.md#bounded-completion-counter-and-clock-b4),
and [COMPATIBILITY-SPEC](../../docs/05-verification/COMPATIBILITY-SPEC.md).
X12 owns the HTTP-success-without-delivery falsifier.

```bash
python3 tools/fixtures/source_async.py /tmp/source-async-attempt-new
PYTHONPATH=tools/fixtures python3 -m unittest discover -s tools/fixtures -p 'test_*.py'
```

This measures the existing closed CI3 source behavior without rewriting it:
`/create` commits an item and a pending jobs row; an explicitly invoked application
worker selects pending rows, posts to the local sink and stores the returned receipt.
The sink persists payload/business_application_id and returns receipt/applied flags.
There is no timer/due column or delayed scheduler in that fixture. Native scheduler,
event/routine/privilege census and actual job state come from fixed server reads.
A ready pending row is not proof of a qualified deadline/due horizon.

Two independently built disposable runs execute the real Alice login, CSRF issuance
and `/create`. Withhold worker invocation and observe HTTP 202, a pending job, null
receipt and no sink acceptance. X12 completion must remain UNKNOWN. Then invoke
the actual worker and capture its exit, the committed acked job/receipt and matching
sink payload/business application ID. These are genuine observed relations but
remain untimed/unpropagated stages, not a complete causal delivery ABI.

The current fixture lacks the causal envelope in its queue, worker parent/cause/span
propagation, independently captured wire receipt/ACK and delivery activity offsets
relative to initiating dispatch. Each existing raw operation has its own anchor;
subtracting different anchors or matching item IDs cannot repair this gap. Worker
exit, stored status `acked` and finite silence do not establish the B4 logical ACK
counter, predicate, fixed deadline, queue eligibility and resettable quiescence.
No positive horizon or due status is invented; both remain unavailable.

The tool uses the unchanged role-isolated SOURCE read capability and controlled
raw intervals. Schema/implementation/profile/control refs, actual source/build/
runtime/window/run/launch identities and raw queue/state/sink captures are retained
in a new namespace. Qualification 1.0 is deliberately INCONCLUSIVE with exact
missing-capability reasons. An imported status or structurally valid defaults
cannot establish support. Altered capture scope/hash, fabricated default horizon,
foreign model/plan fields and changed queue bytes are rejected.

Outcome B stops at this instrumentation checkpoint. Completing the missing ABI
requires a separately qualified trusted job-envelope propagation and independent
receipt/ACK timing mechanism within the disposable boundary; these facts cannot
be retrofitted into existing records. No scheduler architecture, fixture behavior,
whole-source witness/occurrence/trace, Behavior IR, auth contract or target is added.
DB intermediate visibility and async causal/completion ABI remain PARTIAL. Old
windows-b1, health-horizon-b1, role qualification and all prior source bytes remain
unchanged. Actual application-only negative-control writes remain contained in
fresh disposable stores; the observer/checker performs no writes or sink delivery.

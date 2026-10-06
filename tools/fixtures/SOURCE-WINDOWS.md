# Controlled SOURCE intervals and fresh captures

Owners: [BEHAVIOR-MODEL clock](../../docs/02-domain/BEHAVIOR-MODEL.md#bounded-completion-counter-and-clock-b4),
[RUNTIME-OBSERVATION](../../docs/03-analysis/RUNTIME-OBSERVATION.md#source-local-publication-and-trusted-execution-witnesses),
and [Trace 2.1](../../specs/schemas/trace-schema.json).

```bash
PYTHONPATH=tools/fixtures python3 -m unittest tools/fixtures/test_source_windows.py tools/fixtures/test_source_roles.py tools/fixtures/test_ci3_reference.py
python3 tools/fixtures/source_windows.py /tmp/source-windows-new
```

Use the existing pinned CI3 runtime. The command seals the source, validates the
prior exact role qualification and implementation closure, independently builds
and executes the unchanged reference scenario twice. Each scenario receives
constructor-bound SOURCE driver/observer capabilities. A separate function globals
allocation supplies those capabilities without monkey-patching the legacy module,
editing its bytes, reimplementing its scenarios or changing old qualifications.
The observer uses the existing qualified SourceCapture boundary for every read.
No target, comparator, model, plan, approval or caller timing input is accepted.

The supervisor allocates a unique raw window identity and logical operation visit
before dispatch. After a separate PRE capture, it samples monotonic nanoseconds and
UTC once at dispatch, recording the sampling bracket. All subsequent activity/end
timestamps derive from that monotonic anchor. Activity millisecond offsets round
up; timer reads round down. No application/worker wall time or business clock is
used. The interval ends after the actual return and POST capture; it does not
pretend capture continued to a full timeout. PRE snapshots precede dispatch and
are not declared to be events inside the operation interval.

Immutable raw records bind run, operation/visit, source manifest, runtime lock,
content build, actual result/state, clock samples and prior control reference.
Repeated payloads are not merged. Original scenario responses associate with
windows by in-memory response object identity, never array index or equal payload;
publication verifies operation, result and state correspondence. Independent runs
have distinct physical identities. Concurrent dispatches can overlap; allocation
and registration are locked. Final witnessed occurrence identity remains a later
publication concern; these raw IDs are not StepOccurrence artifacts.

Failed dispatch/capture retains INCOMPLETE evidence and any actually measured end.
Missing or invalid boundaries cannot qualify. Qualification executes deliberate
altered-scope/hash and missing-boundary controls against real intervals, plus an
actual invalid-delivery control under the live supervisor clock. Unit tests also
exercise reversed/zero intervals, forged UTC, rounding, duplicate/reused identity,
foreign inputs, capture failures and overlapping repeated execution. Pinned builds
and the existing bounded response reproduction projection must agree between runs;
that projection authorizes no normalization of security-sensitive values.

`qualification.json` uses existing qualification 1.0 for the exact raw-interval
profile. `control-result.json` references fresh raw corpora, windows, actual launches,
failed-delivery controls and immutable implementation/profile/build refs. Each run
retains raw snapshots/responses, actors' request cookies and the existing scenario
control outcomes. Hashes cover canonical JSON without storage newline; implementation,
runtime and schema pins cover raw bytes. Compressed build logs preserve original bytes.

These are raw capture records, not completed Trace 2.1. The familiar observation
window start/end/control fields are retained; schedule_horizon_ms, queue_state_ref
and due_beyond_horizon remain null because qualified completion/scheduling evidence
is unavailable. They cannot be substituted with zero/false to satisfy trace schema.
A controlled interval establishes attribution, not synchronous closure or async
receipt/causal completeness. DB intermediate visibility and async causal/completion
ABI remain PARTIAL. No ExecutionWitness, StepOccurrence, ledger or trace is published.
Old corpora, role qualification, projection and attestation remain unchanged.

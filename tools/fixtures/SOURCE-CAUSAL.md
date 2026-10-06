# Disposable source causal carrier qualification

Owner: [RUNTIME-OBSERVATION](../../docs/03-analysis/RUNTIME-OBSERVATION.md#trace-and-causality)
and existing [causality envelope](../../specs/observation/causality.schema.json).
This closes one smallest prerequisite: producer/queue/worker/receiver carrier
propagation. It does not qualify the complete scheduled/async scope.

```bash
python3 tools/fixtures/source_causal.py /tmp/source-causal-new
PYTHONPATH=tools/fixtures python3 -m unittest discover -s tools/fixtures -p 'test_*.py'
```

The original CI3 fixture, dependencies, locks and historical evidence are unchanged.
The tool first validates and seals that source, then applies four reviewed overlay
files in a separate temporary disposable revision. Its own successor source manifest
and runtime lock bind the changed controller/seed/worker/sink bytes. No claim of
transparent original-source equivalence is made. This is an explicit fixture-only
metadata extension, not a customer instrumentation variant or production queue.

The trusted fixture runner allocates an initiating envelope using its actual run,
request and source launch context. The real `/create` application persists it in
its job row, assigns actual job/effect identifiers, and commits the transaction.
The real worker reads that stored carrier, inherits a runner-provided span through
its launch environment, preserves the parent/cause relationship and sends the
carrier as a transport header. The local receiver captures that actual header in
its durable acceptance file. Fixed observer SQL and file reads retrieve both
stages; the checker does not create metadata rows or downstream effects.

Two independent builds/runs execute the source role handle controls in the new
profile. HTTP 202 with worker withheld remains pending/UNKNOWN. After worker
execution, validate project/run, initiating request/workflow/trace, job/effect and
worker parent/cause/span against the measured queue and receiver records. Same
business IDs across isolated runs do not establish cross-run identity. The fixture
metadata field is neither authentication nor a target/model/plan/approval input.
Wrong/missing cause, run, job or foreign envelope fields fail qualification.

Qualification 1.0 is SUPPORTED only for the new revision's carrier capability,
with exact overlay/implementation/source/runtime/build/profile/control refs.
Receiver/queue values are actual application metadata; the worker's emitted PID
is not an independently witnessed process identity. No causal completion is inferred
from that field. Root role qualification records for B1 are not retroactively widened;
the same executable role controls are run against the successor source revision.

No independently captured wire ACK, delivery offsets from initiating dispatch,
qualified horizon/due semantics, receipt counter or quiescence is established.
Those require the next independent timing/control prerequisite. Completion stays
UNKNOWN; DB intermediate visibility and aggregate async causal/completion ABI stay
PARTIAL. No original B1 support upgrade, traces/witnesses/occurrences, Behavior IR,
auth contract, M3/TIM or target generation. The new revision must be explicitly
selected as source evidence in later admission; original-source observations
cannot be silently relabeled as observations of this revision.

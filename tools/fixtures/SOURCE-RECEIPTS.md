# Independent receiver transport timing

Owners: [RUNTIME-OBSERVATION](../../docs/03-analysis/RUNTIME-OBSERVATION.md#trace-and-causality),
[BEHAVIOR-MODEL B4](../../docs/02-domain/BEHAVIOR-MODEL.md#bounded-completion-counter-and-clock-b4)
and [COMPATIBILITY-SPEC](../../docs/05-verification/COMPATIBILITY-SPEC.md).
This is one bounded prerequisite, not complete async qualification.

```bash
python3 tools/fixtures/source_receipts.py /tmp/source-receipts-new
PYTHONPATH=tools/fixtures python3 -m unittest discover -s tools/fixtures -p 'test_*.py'
```

A separate successor disposable source revision reuses the causal overlay but
routes its worker to the supervisor-owned local receiver. Original B1, causal
revision artifacts and earlier source files are unchanged. The successor worker
bytes and runtime lock are retained. No transparent-equivalence claim is made.

The independent Python receiver runs on the owned internal Docker bridge, using
the pinned supervisor Python runtime. The producer/worker has no host mounts or
receiver evidence-file access. The receiver exposes HTTP delivery, not an event-log
write API. Its address is derived from the internal labeled network and verified
inside that subnet; it never selects a production endpoint. Same-account hostile
host/Docker access remains outside the mediated isolation boundary.

Before /create dispatch, sample monotonic nanoseconds and UTC once. The actual
source creates the job; observer SQL binds the receiver to its actual stored
carrier/job/item before worker execution. Incoming payload and envelope must match
that binding. The receiver samples actual arrival, assigns receipt/ACK IDs, writes
the payload/business identity to an exclusive durable file and fsyncs it. It then
samples durable acceptance, sends the actual ACK HTTP body, flushes it, and samples
response-send time. All offsets round activity upward from that initiating anchor;
worker/application wall timestamps never supply timing.

Receipt means independent durable acceptance. ACK here means the receiver's actual
transport response-send boundary. Send/flush does not prove worker consumption.
The worker's actual stored receipt is captured as additional observed state, but
its independently witnessed ACK-consumption span is unavailable. Therefore no
ACKNOWLEDGED effect occurrence, logical ACK counter, horizon, due status, fixed
quiescence or completion is asserted. A timer/sleep or empty event list is not used
to fill those gaps. Completion remains UNKNOWN and aggregate async ABI PARTIAL.

Two independently built disposable runs capture live receipt/ACK-send evidence,
exact source/runtime/build/launch/run/window identity, original source reference,
causal carrier, independent receiver identity, durable file hashes and raw response
bytes. Qualification 1.0 is SUPPORTED only for that exact transport-timing profile.
No worker PID identity is inferred from its application-emitted carrier. Original
role qualification is not widened; its executable controls run in the new profile.

Falsifiers exercise HTTP 202 with worker withheld (no receipt), malformed/foreign
cause/run/request/job/span, forged producer ACK input, missing ACK, replayed delivery,
wrong execution/window/runtime/source, altered timestamps/hash and wrong receipt
reference. Producer attempts to write receiver evidence fail under its read-only,
unmounted filesystem. Replays are explicitly rejected in this narrow profile;
no retry/deduplication qualification is advertised. Actual independent acceptance
can only result from delivered payload, not producer-side ACK=true assertions.

The evidence contains no completed Trace, StepOccurrence or ExecutionWitness.
DB intermediate visibility remains PARTIAL. Horizon/due, receipt-counter eligibility,
quiescence and independent worker ACK consumption need subsequent qualifications.

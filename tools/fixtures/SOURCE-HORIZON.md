# Trace 2.1 horizon qualification boundary

Decision: **Case C**. [RUNTIME-OBSERVATION](../../docs/03-analysis/RUNTIME-OBSERVATION.md#trace-21-horizon-applicability-and-admission)
and [BEHAVIOR-MODEL](../../docs/02-domain/BEHAVIOR-MODEL.md#bounded-completion-counter-and-clock-b4)
agree with the current schema. Every qualifying window has a meaningful horizon,
including zero for proven synchronous scope. There is no async-only applicability
or generic unavailable escape hatch. Required missing evidence blocks publication.
No Trace 2.1 schema bytes, historical vectors or old corpora change.

```bash
python3 tools/fixtures/source_horizon.py /tmp/source-health-horizon-new
PYTHONPATH=tools/fixtures python3 -m unittest tools/fixtures/test_source_horizon.py tools/fixtures/test_source_windows.py tools/fixtures/test_source_roles.py tools/fixtures/test_ci3_reference.py
```

The smallest runtime qualification is the existing pinned `/health` route. It
reads SELECT 1 and returns synchronously; it declares no scheduled emission.
The canonical HTTP_RESPONSE completion record uses the existing driver's five
second bound, zero expected receipts and zero quiescence. No model actions or
Behavior IR are interpreted. Each independent build/run retains real dispatch
anchors, source/build/runtime/launch identity and before/after source state.

The trusted supervisor captures native event scheduler state, event/routine
catalog counts, application database/global event privileges and the actual
job queue using fixed read-only SQL. The native scheduler may be ON or OFF;
qualification requires no installed source events/routines, no application event
privileges and a positively observed empty job queue in both phases. This closed
pinned read-only route cannot create new scheduled work; inspection and declared
scope, not silence or a guessed default, establish horizon zero and no work due
outside it. The original source-role read capability remains unchanged.

`queue-proof.json` pins both actual censuses, raw bytes, originating window/hash,
source/runtime/build scope and completion profile. `observation-window.json`
projects measured boundaries into the existing Trace 2.1 window shape, with
schedule_horizon_ms=0 and due_beyond_horizon=false only after those checks pass.
The projection does not edit its raw window. A null/unobserved census, foreign
profile, incomplete return, exceeded deadline, state mutation or scheduled/pending
work blocks this capability, regardless of whether defaults would validate.

A real application-only negative control logs Alice in, obtains her observed
CSRF token and creates a job through `/create`. Its subsequent `/health` interval
still records HTTP success but **cannot qualify** as the empty-queue synchronous
profile. No checker inserts the job, sends the sink effect or repairs application
state. Native catalog/privilege/unsupported-mode negatives run in focused tests.
Two independent executions reproduce content-build identity and qualified scope;
physical run/window identities and timings remain actual distinct values.

Qualification 1.0, profile/completion records, immutable implementation and schema
refs, raw positive and pending-job control captures and launch records are retained
under a new evidence namespace. Qualification is limited to this profile; status
SUPPORTED cannot be used for arbitrary routes or imported as engine admission.
These are qualified window subrecords, not published complete traces/witnesses.

The 62 windows-b1 records remain historically accurate with unavailable fields;
this experiment does not backfill or qualify them. Scheduled/async source scopes
still need admitted bounded completion, inspected scheduling and causal evidence.
A valid horizon never implies an ACK, receipt, quiescence or full async completion.
DB intermediate visibility and async causal/completion ABI remain PARTIAL.

# SOURCE capture role qualification

Canonical owners: [BEHAVIOR-MODEL execution roles](../../docs/02-domain/BEHAVIOR-MODEL.md#execution-roles),
[RUNTIME-OBSERVATION](../../docs/03-analysis/RUNTIME-OBSERVATION.md#source-local-publication-and-trusted-execution-witnesses),
and [SERVICE-CONTRACTS qualification](../../docs/01-architecture/SERVICE-CONTRACTS.md#exact-implementation-qualification).
The accepted [ADR 0003](../../docs/01-architecture/ADR/0003-behavior-as-canonical-model.md) supplies the X61 falsification boundary.

`source_roles.py` introduces a successor, mediated SOURCE capture capability.
Its observer receives only a fixed read closure, returns detached immutable
actual bytes, and accepts an exact SOURCE/snapshot request. Its checker handle
reads those bytes without measured store, driver, preparation, sink or expectation
handles. The trusted supervisor retains preparation and application delivery.
A copied model allocation is a negative control, not an executable Behavior IR.
This is same-account mediated enforcement, not protection against hostile Python
introspection or an operator with Docker access.

```bash
python3 tools/fixtures/source_roles.py /tmp/source-roles-new-a
python3 tools/fixtures/source_roles.py /tmp/source-roles-new-b
PYTHONPATH=tools/fixtures python3 -m unittest tools/fixtures/test_source_roles.py tools/fixtures/test_ci3_reference.py
```

Use the existing fixture's exact Python, image, archive and Docker pins. Output
must be new. The harness copies and validates the sealed fixture, builds offline,
starts fresh disposable SOURCE resources, and delivers a real health request.
No target executable exists in this path. Controls run against the actual
read-only observer of that runtime. Each foreign target/model/comparator/mapping/
plan/approval input is rejected before observation. Deliberately poisoned shadow
writes do not appear in actual DB/files/sinks. Detached checker reads cannot alter
captures. Direct INSERT using the actual observer DB principal must produce the
specific privilege refusal; unrelated command failure cannot qualify it. Captured
state must remain identical afterward. The trusted supervisor independently retains
the DB server query audit: observer queries may only read, apart from the explicitly
denied INSERT control. Snapshot equality cannot conceal a write followed by repair. Cleanup is required on failure and success.

The required positive/negative control outcomes are declared in `control-suite.json`.
Incomplete, duplicate or failed outcomes cannot produce qualification. Unit
falsifiers substitute plausible permissive/writable capture implementations and
must fail the same executable suite. Those unit tests do not supply runtime evidence.

`qualification.json` uses existing qualification 1.0 and names capability
`rwt-model-roles-1` only for the exact SOURCE input profile and implementation
closure. `control-result.json` retains source manifest, runtime lock, content build,
loaded image, real execution/container launch, actual capture and control references.
`artifact-index.json` exposes qualification/result digests. JSON references hash
canonical JSON without the storage newline; runtime/implementation/schema/diagnostic
refs hash exact raw bytes. Build logs are retained losslessly compressed.

Equivalent inputs reproduce implementation/profile/source/content-build identities,
actual read-only capture and control outcomes. Independent executions have distinct
launch IDs and execution IDs; their result and qualification hashes consequently
differ. HTTP Date and image metadata remain actual evidence, not normalized values.

This qualifies neither the old capture implementation nor old retained corpora.
Consumers must resolve the exact profile/implementation/schema/control closure;
an imported SUPPORTED field alone cannot establish authority. No engine store,
customer approval, witness, ledger, occurrence or trace is published by this tool.
Controlled observation windows remain absent. DB intermediate visibility and async
causal/completion ABI remain PARTIAL. Target lowering, model interpretation and
session/auth qualification are outside this profile.

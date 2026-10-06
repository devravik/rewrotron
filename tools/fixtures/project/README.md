# Source capture projection prerequisite

Owner: [EVALUATION-CONTEXT](../../../docs/02-domain/EVALUATION-CONTEXT.md), with
capture provenance and publication dependencies owned by
[RUNTIME-OBSERVATION](../../../docs/03-analysis/RUNTIME-OBSERVATION.md) (M2).
This is the first, smallest pre-Unit-5 work item: projecting existing SOURCE
capture fragments to context/response 1.0. It introduces no new M4 unit or evaluator.

```bash
go build -trimpath -buildvcs=false -o /tmp/rwt-source-project ./tools/fixtures/project
/tmp/rwt-source-project fixtures/ci3-reference/evidence /tmp/new-source-contexts
go test -count=1 ./tools/fixtures/project
```

Output must be new. The command verifies the current source manifest and each
original canonical corpus digest, uses named attempt/step identity rather than
array index, and writes content-addressed artifacts plus per-run artifact indexes.
It does not execute PHP, query a store, alter input captures, or write measured
application state. It reuses the existing lossless JSON parser, tagged values,
context validator and response validator. Projector identity names exact native
bytes. Checked-in records remain reproducible under the pinned Go toolchain.

Captured status, body, ordered headers and repeated cookie attributes are retained.
Raw byte artifacts retain binary bodies and unsupported cookie representations.
Only observed Set-Cookie attributes are supplied; absent values remain null and
Max-Age=0 remains distinct. Unknown or duplicate attributes leave the response
unavailable rather than silently flattening it. JSON numbers retain exact lexical
tags; duplicate-key/unrepresentable JSON remains MISSING.

File content/mode and safe JSON session snapshots are projected directly from
before/after capture fragments. DB TSV, server diagnostics, sink snapshots and
request metadata remain preserved in immutable original corpus/fragment artifacts.
There is no qualified native DB type mapping/read/intermediate-visibility capture;
therefore DB values stay MISSING, rather than guessing tags from schema or expected
rows. Queue/delivery completion stays MISSING. No snapshot difference is promoted
to a proven causal effect, transaction, delivery, rotation or invalidation event.
Effect reference arrays contain no registered typed effects; they do not certify
absence. The retained raw fragments and explicit limitations remain required inputs.

Login request names may select an intended subject; other unrecorded actor
selections remain MISSING. Neither actor labels, HTTP 401/200 nor stored principal
fields establish actual authentication. The actual-principal namespace remains
MISSING pending an independently qualified capture channel. Raw session principal
fields remain tagged observed data, not authentication verdicts. Effective request
headers/body were not captured, so the closed input namespace is unavailable;
redacted passwords are never substituted with the marker or fixture expectations.
No controlled business clock was captured; HTTP Date never replaces it.

There are no model bindings or row assertions in this pre-contract capture
projection. PRE contexts have no result. Worker captures retain before/after
state, with unavailable HTTP result. No future case, target, model or approval
is required or generated. Existing capture artifacts and PARTIAL attestation
remain unchanged; DB intermediate visibility and async causal/completion ABI
remain PARTIAL permanently within this qualification boundary.

All indexes explicitly report admission UNKNOWN. These are typed data projections,
not engine-admitted captures or qualified execution witnesses. Missing runner/role,
principal, input-byte and clock provenance still prevents dependent acceptance.
The next observation work item is current execution-witness/identity/occurrence/
trace publication and capture-provenance closure where actually observable;
unavailable evidence must remain UNKNOWN. Executable Behavior IR, the fixture's
Unit-4 session contract, M3 plan/TIM/approvals and Unit 5 remain separate later work.

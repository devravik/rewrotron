# Verification requirements for engine development

## Profiles

1. **Deterministic Go tests:** Pure hashing/state/graph/comparison tests and local SQLite/artifact integration with no containers, PHP, network or LLM.
2. **Contracts:** Every JSON Schema, local reference, CLI/MCP registry, success/error fixture, lifecycle enum and command output, including unknown version/field and empty queue.
3. **Storage/recovery:** Fresh migration, version upgrades/checksums, foreign keys, append-only rules, transactional events, leases/fences, crash publication and backup/restore.
4. **Adapter/runtime fixtures:** Synthetic PHP parsing/routing/probe/target generation, pinned isolated environments, instrumented control runs and safe effects.
5. **End-to-end:** Gate-aware complete scoped CI3→Laravel migrations plus intentional mismatch/unknown/repair and process restart.
6. **Performance:** Independent reproducible workloads after corpus correctness; report inconclusive results explicitly.

[FIXTURE-CORPUS](FIXTURE-CORPUS.md) maps scenarios to paths. Tests target observable behavior and failure cases, not trivial copies of implementation internals. Inject clocks/IDs/runners for repeatability.

## Mandatory checks after implementation

Run `go test -race ./...`, `go vet ./...`, pinned lint, schema/CLI/MCP contract checks, migration compatibility and the affected fixture suite. Complete runtime/reference checks before claiming adapter support. Measure core coverage with a target above 85%; prioritize gate, drift, publication, reference ownership and uncertainty branches rather than gaming a percentage. A coverage number alone is not proof of correctness.

Test errors separately from verdicts: completed mismatch/unknown is exit 2; unsafe or incomplete execution error is exit 1 with UNKNOWN diagnostic evidence. No case can pass solely because status is 200. Counterexamples are required for every approved normalizer.

For documentation-only work before Go packages exist, run `python3 tools/docs/validate.py` and report exactly its scope. Do not present initial DDL validation as executed Go migration compatibility or runtime testing. CI profiles and the initial checks are built in RWT-001/002.

## Evidence artifact

Record command, tool/dependency versions, exit status, executed profile/corpus, hashes and limitations. Mark incomplete release-required environments as blockers; skips are never equivalent to a release pass. No test/check is pre-marked complete in a checklist.

## Transparent execution acceptance

Test the canonical event/state contracts and real long-running producer instrumentation. Required cases include heartbeats while a subprocess waits, stale worker display, concurrent stage/task runs, privacy in every verbosity, cursor gaps/duplicates/fixed page watermarks, client-specific deltas, zero/unknown denominators, changed weighted scope, failed execution versus MISMATCH/UNKNOWN, pause at journal/replay boundaries, restart/resume retaining pause and continue with stale gates. Test telemetry failure without repeated side effects and approval events without actual confirmation. [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) owns acceptance details. JSON fixtures prove shape only; runtime/race/recovery/client evidence remains required.

# Data flow and publication boundaries

| Stage | Input | Durable result | Guard |
| --- | --- | --- | --- |
| Initialize/detect | Readable source and bounded metadata scan | Detection proposal, manifest | No source execution or deep parsing |
| Confirm baseline | Reviewed detection and overrides | Immutable baseline and Gate 1 record | Explicit human approval |
| Analyze | Confirmed source snapshot | Source IR, Application IR, graph and findings | Valid Gate 1 |
| Observe | Sandbox copy, fixture, scenario | Redacted traces and completeness map | Valid Gate 1, isolated environment |
| Discover | Structural graph and traces | Contracts, cases, workflow relations, unknowns | Evidence/provenance required |
| Configure/freeze | User intent and recommendations | Immutable target revision, provenance; no gate approval | Schema validation; explicit choices never replaced by recommendations |
| Resolve/assess | Source usage, capability/constraint graph, requested target | Recursive target graph, API differences, feasibility and decision artifacts | Valid Gate 1 for deep analysis; unknown required compatibility blocks dependent scope |
| Plan | Behavior IR, frozen target, resolved dependencies, scope | Immutable plan, ordered units, omission and decision review | Valid Gate 1; target feasibility and generation blockers visible |
| Confirm plan | Exact plan subject | Gate 2 record | Current baseline and human approval |
| Generate/repair | Confirmed plan, ready unit, before hashes | Staged patch, target tree hash, provenance | Valid Gates 1/2 and mission fence |
| Verify | Identical fixtures and inputs in independent environments | Run report, dimensions, verdicts | Fresh source/target/corpus/policy hashes |
| Benchmark | Passing corpus, pinned workload | Measurement and acceptance report | Equivalent behavior under measured load |
| Confirm/package | Current release evidence and runbooks | Gate 3 record, immutable package | Current subjects and human approval |

Each result captures upstream hashes and adapter/tool versions. Large results use `sha256:<digest>` artifact references. A result is published only after artifact persistence succeeds and the captured subjects still match current state. Events and entity updates share a SQLite transaction.

JSONL event output is an export of committed records. It can lag or be rebuilt; no filesystem write is claimed to be atomically committed with SQLite. See [EVENT-MODEL](../08-storage/EVENT-MODEL.md).

Source traces are evidence, not executable instructions. Proposals from an agent or workflow heuristic include confidence and source references. The engine validates them and records unresolved questions before they can enter a locked corpus. A change to evidence or scope creates a new plan rather than silently rewriting a confirmed plan.

## Shared execution flow

Every long-running stage records structured starts, factual rationale, observations/decisions/actions/results, progress/liveness and blockers through `internal/execution` and the existing SQLite event ledger. Domain effects and their result events commit together; live clients see only committed state at a sequence watermark. CLI, MCP and future desktop render the same projection and registered drill-down refs. Approval cards delegate to existing confirmations; control pause is a persisted scheduler guard, not a new lifecycle gate. See [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md).

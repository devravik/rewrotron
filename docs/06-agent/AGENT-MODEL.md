# Agent and deterministic engine boundary

The engine owns gates, state, hashing, scheduling, sandbox execution, persistence and verifier verdicts. An optional external agent proposes interpretations, architectures, code patches, diagnosis and knowledge. The core runs without an LLM or provider credentials.

Agents read snapshots and evidence through the CLI/MCP, claim fenced missions, propose target patches through staged generation, and submit completion evidence. Mediated agent operations do not permit direct SQLite mutation; same-account shell/filesystem access is outside this enforcement boundary. Through engine services they cannot, invent observations, approve their own proposals, mark uncertain code DEAD, overwrite source files, or alter deterministic results.

The three approvals are exact subjects described in [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md). Chat agreement alone is not a persisted approval. CLI human actions or MCP calls consuming a separately issued human token create `user_confirmations`. YAML flags and agent-controlled actor names cannot establish authority. The local account remains the trust boundary.

## Operating loop

1. Detect and present the source proposal and uncertainties.
2. Wait for explicit persisted Gate 1 approval; report required missing inputs.
3. Analyze and observe safely, then materialize evidenced contracts/cases.
4. Propose a capability-checked plan and present omissions/exclusions/blockers.
5. Wait for Gate 2, then peek and claim ready missions.
6. Generate/repair target slices; invoke the engine verifier and inspect every required dimension.
7. Complete only with current fenced evidence; UNKNOWN creates observation/review work.
8. Benchmark the passing corpus and prepare release/rollback/session evidence.
9. Request persisted Gate 3 approval of the current release subject; package the runbook.

Direct deterministic fixtures implement a complete vertical slice without an agent. General business-rule interpretation can require customer/agent assistance, but uncertainty remains explicit rather than fabricated target logic. Client compatibility is established by protocol fixtures and at least two actual client sessions before a stable integration claim.

## Target authority and dependency decisions

Agents inspect separate detection, requested target, recommendation, dependency and plan artifacts. They may propose changes with evidence, but cannot edit user-owned target versions as a fallback. A candidate explicit user edit requires recomputed feasibility and a newly reviewed Gate 2 plan. Resolve required capabilities rather than copying all declared packages; preserve dormant/unknown usage. BLOCKED dependencies create affected-unit findings and human decisions, never invented equivalence. Resume uses SQLite confirmations and current hashes, never YAML approval flags. [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) and [DEPENDENCY-RESOLUTION](../04-migration/DEPENDENCY-RESOLUTION.md) define these engine-enforced obligations.

## Visible agent execution

An agent works through mission → reason/action → evidence/result → engine-validated completion → next mission. Canonical events/state update with those committed steps; the agent cannot replace them with an opaque activity message or make its narration an observed fact. Explanations cite goals/evidence and label hypotheses without exposing hidden reasoning. Inspect [EXECUTION-OBSERVABILITY](EXECUTION-OBSERVABILITY.md) for stage/task/result, liveness, blockers and next-action obligations. Persistent human pause stops new mission readiness; read-only inspection remains available.

## Reproducibility and trust

Current generation requires v2 executable contracts and the approved Target Implementation Model; arbitrary prose is not interpreted truth. Pins include source snapshot, exact resolution, adapters/comparators, tools/skills/knowledge and acceptance scope. Another client resumes through persisted subjects/missions/evidence without conversation history. Same-account users can bypass mediated access externally; SECURITY owns that limitation. An unavailable LLM does not prevent deterministic fixture replay/acceptance, but unsupported interpretation/generation remains BLOCKED rather than invented.

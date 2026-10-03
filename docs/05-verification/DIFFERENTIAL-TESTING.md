# Differential replay algorithm

## Preconditions

A run captures current Gate 1/2 validity, source and target hashes, case/contract/corpus/policy hashes, source/target adapter versions, sandbox image digests, DB schema and fixture hashes, locale, timezone, seed, and tool versions. Non-HTTP inventory remains visible even when a target adapter cannot execute it.

## Execution sequence

1. Validate the case and every artifact/reference before starting a process. Reject unsafe effects or unsupported required dimensions.
2. Create independent source and target environments from the same logical fixture. They never share a writable DB, session store, filesystem, or outbound-effect sink.
3. Restore all relevant state: DB rows and sequences, sessions, cache, files, queues, clocks/random providers where controllable. Verify restoration hashes. Restore failure yields UNKNOWN.
4. Execute the same scenario. Preserve each side's cookie jar through its own workflow. Extract values such as CSRF tokens and generated IDs using explicit bindings; never replay a source-side token into the target.
5. Capture request/response, redirect chain, status/exit code, committed DB state, session/flash, files/cache, and sink events. Automatic redirects are disabled unless a case declares the chain to follow. Each step has a 30-second default timeout, a maximum of 10 declared redirects, and 16 MiB capture limit; configured increases become policy inputs.
6. Persist redacted artifacts and completeness statuses. Timeout/cancellation/truncation prevents complete assertions; preserve partial evidence without declaring PASS.
7. Compare using [COMPATIBILITY-SPEC](COMPATIBILITY-SPEC.md). Store per-dimension results, aggregate verdict, captured hashes and diff artifact.
8. Recheck drift and target fences before accepting the result as current. Append events and results atomically. Clean up processes, mounts and temporary state even on cancellation.

Workflows reset state before each complete scenario, not between steps. Cases can run concurrently only with independent environment/fixture instances. A DB rollback around a request is insufficient when the application commits, sends effects, or uses multiple connections.

## Repair coordination

MISMATCH queues a repair mission with the contract, both traces, exact dimension differences, relevant symbols, and previous attempts. UNKNOWN queues an observation or review mission. No provider is invoked by the deterministic engine; external agents consume the mission through CLI/MCP.

A repair stage records its hypothesis and target before hashes, validates paths and edits, runs build/static checks, then reruns the affected case and regression corpus. Maximum automatic attempts are three per mismatch/current context. A changed context requires a new fingerprint and cannot silently reset the recorded attempt history. Exhaustion creates a blocker with prior evidence and a concrete requested decision.

Regression impact can restrict early feedback, but a unit/project completion claim requires the full confirmed corpus. Comparator or case changes are proposals requiring a new policy/corpus and plan approval. Failed code patches are reverted only if their current hashes still match the applied patch; customer edits are never overwritten by rollback.

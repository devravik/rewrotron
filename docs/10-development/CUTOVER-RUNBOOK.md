# Cutover and rollback runbook contract

The first engine release prepares packages/runbooks after Gate 3. It does not change DNS, load balancers, production schema, credentials or customer traffic. Operators execute their infrastructure-specific reviewed procedure.

## Required release inputs

Record source/target release IDs and hashes, dependency locks, source/target environment settings, preserved DB/schema identity, migration scope and exclusions, current passing verification/benchmark reports, external integration mappings, session continuity evidence, rollback owner and exact executable operator commands. Customer commands are supplied and validated in staging; the engine must not fabricate them from a provider name.

A completed runbook includes preflight health/data-backup checks, safe deployment of the target, route ownership/canary plan, consistency and write ownership, smoke workflows, session checks, external-effect checks, monitoring window, acceptance criteria and rollback steps. Source/target dual writes require a separately approved consistency design and are not the initial default.

## Rollback readiness

Before approval, perform a staging rehearsal of the supplied rollback commands and record results. Define concrete triggers for mismatch/error/latency/session/effect failure using the agreed workload and monitoring. Show who decides, how traffic is restored, what data may need reconciliation, and how source versions/config/session state are retained. Preserving the schema reduces risk but does not make all new target writes reversible.

`report` assembles the release manifest and computes `release_hash` only when all required inputs/evidence exist. Missing commands, unproved session continuity, incomplete scope reconciliation or inconclusive required benchmarks prevent Gate 3. `confirm cutover` approves that subject; `cutover prepare` verifies it again before writing the package. A changed package/runbook requires a new approval.

The migration report records approval time/actor, limitations, intentional changes, evidence references and package checksum. Its existence is not evidence that production deployment has occurred.

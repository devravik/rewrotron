# Customer inputs and blocker handling

This checklist supplies customer facts that engineering cannot determine from repository documentation. A missing input produces a finding with code, affected stage/units, evidence and requested action; unrelated work can continue.

| Required by | Input | Source and refusal behavior |
| --- | --- | --- |
| Detection/Gate 1 | Readable source root, include/dependency roots, runtime/framework corrections | Bounded metadata evidence; unresolved facts shown for review |
| Gate 1/analysis | Operational HTTP/CLI/cron/event roots, generated code/config paths, exclusions | Source plus customer operational manifest; unknown roots prevent DEAD proof |
| Observation | Sandbox runtime image/cache, application startup, safe fixture data and DB engine/schema/settings | Doctor validates; no host/production fallback |
| Observation/auth | Synthetic accounts/roles, secret references, session driver/format/cookie policy | Missing safe accounts or session capture blocks required cases |
| Modeling | Critical workflows, business invariants and failure paths | Existing tests/traces or explicit customer declarations; undocumented rules remain unresolved |
| Planning/Gate 2 | Target path/runtime/framework/style, supported scope, intentional changes and omission review | Target config is explicit; unsupported capabilities fail before generation |
| Verification | Resettable DB/cache/session/files/queues and outbound mock behavior | Fixture restore/completeness evidence required |
| Benchmark | Workload mix/concurrency, quotas, memory instrumentation and acceptance policy | Missing environment evidence is inconclusive |
| Gate 3 | Session-continuity strategy, approved release, traffic/runback ownership and rollback criteria | Readiness blocked until evidenced runbooks and approval |
| Optional exports/global learning | Reviewed privacy/export manifest and destination | Local-only default |

## Supplied configuration

`rewrotron.yaml` at the workspace root is schema-validated user intent; `.rewrotron/project.yaml` remains a legacy input only when the root file is absent. An optional exclusive `.rewrotron/target.yaml` holds the same target object. Generated detection/recommendation/approval views cannot overwrite intent or grant authority. See [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md). The target section may be absent at initialization; before planning it must specify language/framework/version/runtime, architecture, preserve-existing DB policy and a writable target path outside source roots. The reference recommendation is php/laravel, modular_monolith, PHP 8.3+, but the user specifies the requested runtime constraint. No recommendation silently fills or replaces explicit choices. Supply target infrastructure/compatibility settings, and explicit dependency strategy decisions for the actual required capabilities; recursive feasibility and characterization requirements are reviewed before execution.

Input scenarios are placed in `.rewrotron/import/scenarios/<scenario_id>.json`, validated against `specs/behavior/observation-scenario.schema.json`. They contain concrete synthetic inputs, fixture refs, principal IDs, required dimensions and bindings without claiming expected outputs. `observe --case <scenario_id>` first resolves a locked case by ID, then an imported scenario; duplicate IDs are rejected. `discover` derives candidate assertions from the observed evidence and creates canonical cases.

Operational entrypoints are supplied as `.rewrotron/import/entrypoints.json`, a JSON array of canonical entrypoint dispatch records. Include schedules/timezones and argument/message schemas; the analyzer merges them with source discoveries and provenance. Customer declarations can establish roots, never fabricate runtime hits.

Secret refs use `env:NAME`. Mock endpoint/provider fixtures must define request matching and response/effect behavior; unmatched calls fail closed and are reported. No live default gateway, payment, email or object-store bucket is invented.

## Blocking decision format

A finding includes `finding_id`, code, severity, affected IDs, evidence IDs and message; the persistence record adds required action and status. A review response names the exact finding and subject revision. If it changes baseline facts, re-confirm Gate 1; target/scope/corpus/policy changes require Gate 2; release/runbook changes require Gate 3. Scope exclusions remain visible in every report.

The development plan can eliminate missing engineering contracts; it cannot supply undocumented customer behavior or demonstrate a runtime that has not been executed.


## Agent patch submission

An agent prepares `.rewrotron/import/patches/<patch_id>/manifest.json` matching `specs/migration/patch.schema.json`, with new bytes in that directory under the same relative target file paths. `patch stage` / `migration.stage_patch` resolves only this project-owned import directory, validates paths/before-after hashes, reads bounded bytes, rejects customer secrets/symlinks and registers content artifacts. Existing content refs must belong to the same project. The engine returns a registered patch artifact ref; `patch apply` invokes the same fenced publication path as adapter generation.

The manifest names an active ready unit and live fencing token. It cannot write outside the configured target, grant approvals or change contract/verifier policy. Missing token/lease, unsupported paths or mismatched before hashes fail before mutation. Reverification and mission completion evidence remain mandatory after application.

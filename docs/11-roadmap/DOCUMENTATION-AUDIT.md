# Documentation audit and gap resolution

**Date:** 2026-10-03. **Reviewed baseline:** commit `4d26680`, with the repository constitution supplied by the owner. **Scope:** README, the original project exploration, every documentation subsystem/ADR, and all seven original draft JSON schemas. Repository inspection found documentation/schema artifacts and LICENSE, with no Go module, runtime implementation, migrations, executable tests or fixture corpus.

## Review outcome

The architecture already established behavior reconstruction, three IR tiers, framework neutrality, vertical slices and mandatory gates. Development would still require undocumented decisions about persistence, approvals, hashes, execution, interfaces and acceptance. This review completes the specification boundaries and puts external validation/customer facts into explicit owned tasks; it does not certify an engine that has not been built.

## Finding register

“Specified” below means corrected in documents/contracts with an implementation task. Runtime closure requires the referenced task's executed evidence.

| ID | Gap/contradiction found | Resolution and owning document | Task evidence |
| --- | --- | --- | --- |
| A01 | Planned CI3/Laravel/future adapters advertised as stable/tested | Truthful implementation status and maturity criteria in PRODUCT/README | 007–018,025 |
| A02 | Mathematical/unforgeable proof claims exceeded finite evidence | Corpus-bounded compatibility and integrity-only hashes | 008,012,022 |
| A03 | Init blocked before a DB existed to record approval | Minimal init/detection store before Gate 1, no deep work | 004–006 |
| A04 | `--auto-confirm` could bypass showing reviewed subjects | Removed; exact explicit persisted approvals | 005,006 |
| A05 | Four competing lifecycle/state vocabularies | One project state machine; mission/unit statuses separate | 001,005,019 |
| A06 | Undefined material repository drift/configuration hash | Byte manifests, exclusions, environment identity, three subject hashes, invalidation matrix | 004,005 |
| A07 | Plan approval stored config hash without exact plan/corpus/target identity | Immutable plan_hash and required parents; release_hash separately | 005,014,022 |
| A08 | Agent-provided identity confused with human approval | Trusted actor/channel and separate subject-bound human approval token; local trust limit explicit | 005,020 |
| A09 | Gate 3 had no command/tool and a 99% shortcut | Exact source/plan/cutover confirmation catalog; complete required corpus and runbooks | 006,020,022 |
| A10 | Framework parsing/hooks in neutral internal packages; conflicting paths | Concrete adapter ownership, neutral ports, composition package and import checks | 001,007,008,010 |
| A11 | Source → Application IR transformation undefined | SourceAdapter.Extract and typed evidence/reference records | 007,008 |
| A12 | Behavior graph/cardinality confused cases/contracts/routes | Many-to-many entrypoint/operation/workflow/unit relationships | 011,014 |
| A13 | Non-HTTP types required HTTP fields in schemas and were missing storage | Type-specific dispatch/outcomes, generic entrypoints, explicit unsupported inventory | 008,011,013 |
| A14 | Natural-language rules and placeholders treated as executable behavior | Supported declarative rule representations and generation capability blockers | 011,014,016 |
| A15 | No concrete input-only observation/bootstrap scenario | Imported scenario schema, synthetic principal/fixtures, discovery path | 009–011 |
| A16 | CI3 routing incorrectly said no native HTTP verb routing | Corrected verb arrays, precedence, `_remap`/convention edge cases | 008 |
| A17 | CI3 hooks claimed to avoid source changes while editing hooks.php | Disposable instrumented copy; proxy/control traces and completeness | 009,010 |
| A18 | Proxy env vars claimed to isolate arbitrary untrusted apps | Container/network/mount/egress enforcement; local mode limited to trusted fixtures | 009,025 |
| A19 | Redaction erased replay tokens or claimed complete secret guarantees | Synthetic secret refs, in-memory assertions, keyed correlation, withheld UNKNOWN dimensions | 003,010,013 |
| A20 | Empty effects silently meant no effects | Required per-dimension completeness and unavailable→UNKNOWN | 010,012,013 |
| A21 | Blanket JSON/HTML/token/ID normalization could hide defects | Scoped exact/semantic rules, numeric types, multivalue cookies and relation checks | 012,017,018 |
| A22 | EXPECTED_CHANGE was both verdict and mismatch classification | Three deterministic verdicts, separate approved exception and immutable rerun | 001,012,013 |
| A23 | DB diff ignored observable IDs and relied on query logs | Typed committed snapshots, triggers/rollback, exact IDs or approved bijection | 012,018 |
| A24 | Password/session continuity guessed or missing | Detected profiles, bridge/security cases, explicit unproven blocker | 017 |
| A25 | Performance product promise conflicted with automatic 20%/2x allowance | Paired protocol, confidence/uncertainty, reviewed deviation separate | 021 |
| A26 | Coverage percentages had no scope/denominators | Explicit numerator/denominator/null and entrypoint reconciliation | 022 |
| A27 | CLI envelope had no command-specific catalogs or full errors/dry-run | Strict schemas, complete command registry, exact exits/errors/options | 001,006,013,019,022 |
| A28 | MCP catalog lacked input/output schemas, lifecycle, trust/resources | Versioned stdio protocol and complete tool schemas/catalog | 020 |
| A29 | Persistent DDL omitted plans/baselines/runs/edges/leases/checkpoints | Initial executable 35-table migration and normalized versioned records | 002 |
| A30 | Append-only declarations conflicted with cascading deletion/no triggers | SQL triggers, RESTRICT ownership, immutable evidence and mutable projections | 002 |
| A31 | Dual SQLite/JSONL writes falsely claimed atomic audit durability | Transactional SQLite events; recoverable cursor-based JSONL export | 002,015 |
| A32 | Artifact hashing/paths/GC could discard needed evidence | Content addressing, artifact-first write protocol, reachability retention | 003,015 |
| A33 | Crash resume and <50ms rollback had no real semantics | Publication journal, before hashes, idempotency and recovery matrix | 015,019 |
| A34 | Missions had no atomic leases/fences/empty-queue/evidence completion | Explicit protocol, readiness/order/limits and current-subject completion | 019,020 |
| A35 | External agents had no governed repair patch submission seam | Project-import patch manifest, target-only stage/apply and publication checks | 015,019,020 |
| A36 | Global learning could cross privacy boundaries automatically | Local lessons first, reviewed opt-in promotion with independent evidence | 023,024 |
| A37 | Go/tool/parser/framework versions assumed or mutually inconsistent | Dependency compatibility task, pinned protocol, supported source syntax/runtime matrix | 001,007,009,020 |
| A38 | Roadmap postponed verifier/CLI until after generation and duplicated IDs | Dependency-ordered 27 tasks and canonical milestone/backlog mappings | Entire development plan |
| A39 | Customer/fixture/release inputs and blocker ownership missing | PROJECT-INPUTS, fixture corpus, decision/validation register | 007,009,017,021,022,025 |
| A40 | No actionable cutover/rollback or independent target deliverable | Supplied staging-tested operator runbook/session strategy; Gate 3 package | 016,017,022,025 |
| A41 | Absolute machine-specific links and fragmented doc navigation | Portable links, documentation/schema index and owner map | Document validation |
| A42 | No executable checks for draft schemas/DDL/document coherence | Offline schema/reference/example/registry/link/DDL validator | Current artifacts; Go checks in 001/002 |

## Added development detail

The development plan includes 27 dependency-ordered tasks, owner roles, exact file paths, shared interfaces/records, negative/counterexample fixtures, commands, evidence and requirement traceability. New detailed specifications cover confirmations/hashes, IR provenance, mission leases, errors, auth/session continuity, coverage, recovery, fixture matrix, customer inputs and cutover/rollback. ADR 0011 records the concrete baseline choices.

Machine artifacts now cover source/application/behavior IR, operation contracts/cases/scenarios/workflows, non-HTTP dispatch, target capabilities/plans/patches/releases, confirmations/snapshots, traces/completeness, missions, benchmark/verifier reports and command/tool inputs. Canonical descriptive metadata remains structured but open where semantic code validation is required; schemas alone do not prove graph ownership, safe generation or gate correctness.

The original project exploration retains product scenarios and rationale; duplicate contradictory implementation sections now point to their owner. Existing ADR decisions remain, with clarification of unsupported claims and the current initial protocol/evidence baseline. AGENTS.md and its three customer migration gates are preserved.

## Validation scope and remaining evidence

Run `python3 tools/docs/validate.py` to check all schema metaschemas/local references, positive and negative examples, CLI/MCP registry coverage, portable local links and initial SQL/document parity. It also executes fresh SQL creation and focused ownership/append-only/lifecycle/transaction checks. Run `git diff --check` for whitespace issues.

No Go/race, container probe, real framework generation, client session, performance or production test is claimed. Those checks require implementation and the explicit validation inputs V01–V10 in [DECISION-REGISTER](DECISION-REGISTER.md). Scope-specific customer decisions remain visible; there is no promise that undocumented application facts cannot block a migration.

## Primary-source corrections

- [Laravel release policy](https://laravel.com/framework/docs/releases) confirms PHP 8.3 as Laravel 13's minimum.
- [MCP transports](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports) and [lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle) define the pinned stdio/protocol behavior and Streamable HTTP distinction.
- [PHP parser repository](https://github.com/VKCOM/php-parser) documents the selected Go backend and PHP syntax range.
- [CI3 routing](https://codeigniter.com/userguide3/general/routing.html?highlight=uri) documents verb-specific route arrays; [session documentation](https://codeigniter.com/userguide3/libraries/sessions.html?highlight=sessions) makes driver-dependent session storage explicit. Direct fetches were blocked; indexed official documentation supported these corrections. Fixture/runtime availability remains a validation task.

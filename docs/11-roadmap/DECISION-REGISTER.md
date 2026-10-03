# Engineering decisions and remaining validation

**Status:** Decisions below are specification choices. External validation rows are deliberately uncompleted and have explicit task owners. No customer baseline or production gate is approved by this file.

## Fixed baseline choices

| ID | Decision | Owner/specification |
| --- | --- | --- |
| D01 | Go engine, local SQLite WAL/artifacts, no mandatory provider/cloud | Constitution, ADR 0001/0002/0011 |
| D02 | modernc SQLite via database/sql; standard-library CLI; Go PHP parser behind adapter | ADR 0011, RWT-001/002/007 |
| D03 | Canonical states and three hash subjects; minimal init before Gate 1; no auto-confirm | Lifecycle/confirmation protocol |
| D04 | Domain ports, concrete composition package and build-time adapters | SYSTEM-DESIGN/COMPONENTS |
| D05 | CI3 3.1.13 fixture, Laravel 13/PHP 8.3+, preserved MySQL sandbox schema | PRODUCT, adapter specs |
| D06 | MCP 2025-11-25 stdio; remote HTTP/SSE deferred | MCP-SPEC |
| D07 | PASS/MISMATCH/UNKNOWN; approved change separate; scoped relational normalization | COMPATIBILITY-SPEC |
| D08 | Independent fixture state; container isolation; no proxy-only security promise | SANDBOX/DIFFERENTIAL-TESTING |
| D09 | SQLite events authoritative, artifact-first metadata commit, recoverable JSONL export | Storage/recovery specs |
| D10 | Fenced at-least-once missions, three attempts/context, evidence-only completion | MISSION-PROTOCOL |
| D11 | Current corpus/scope denominators; no empty 100%; no 99% cutover shortcut | COVERAGE |
| D12 | Goal non-regression; paired measurements and explicit inconclusive result | PERFORMANCE |
| D13 | Default active-session continuity; unsupported format blocks readiness | AUTH-AND-SESSION |
| D14 | Gate 3 package/runbook only; live infrastructure runner deferred | CUTOVER-RUNBOOK |
| D15 | Project lessons first; reviewed opt-in global promotion/skills later | LEARNING-SYSTEM |
| D16 | Apache-2.0 repository license; customer dependency notices preserved | LICENSE/RELEASE |

## Required validation, not hidden assumptions

| ID | Deliverable/input | Responsible task/role | Can start now? | Refusal/fallback |
| --- | --- | --- | --- | --- |
| V01 | Exact supported Go/modernc/parser/schema/lint versions and module ownership | RWT-001 core maintainer | Yes: inspect available tooling and pin compatible versions | Fail bootstrap until compatible; never invent current release |
| V02 | Fixture archive/license/checksum, PHP and MySQL image digests, cache/locks | RWT-007/009 adapter/runtime maintainers | Yes: synthetic reference fixture setup | Static characterization can continue; runtime blocked |
| V03 | Actual probe completeness and observer control parity | RWT-010 observer maintainer | After V02 | Required dimensions UNKNOWN |
| V04 | Concrete evidenced business-rule representations/generator capabilities | RWT-011/014/016 behavior/target maintainers | After source corpus | Unsupported rule blocks unit generation |
| V05 | Actual source-session/password interoperability | RWT-017 auth maintainer | After first target fixture | ACTIVE_SESSION_CONTINUITY_UNPROVEN; no cutover readiness |
| V06 | Go MCP library implementation and two actual client sessions | RWT-020/024 interface maintainer | Protocol fixtures first | CLI remains baseline; no untested client support claim |
| V07 | Target performance under intended workload with adequate counters | RWT-021 performance maintainer | After correctness | Inconclusive result blocks required benchmark readiness |
| V08 | Customer cutover commands/owner/monitoring and staging rollback rehearsal | RWT-022 release maintainer/customer | Customer supplied per project | Package not ready; no invented provider commands |
| V09 | Native platform runner coverage, license/SBOM/signing custody/security contact | RWT-025 release maintainer | After runtime/release profile | Publish only verified platform scope; no release until required contact/custody assigned |
| V10 | Customer source facts, jobs/integrations, safe fixture/accounts and workflows | PROJECT-INPUTS, project owner | Supplied per migration | Concrete findings and dependency blockers |

No unresolved design choice is disguised as a runtime guarantee. Defaults describe reference fixtures and proposed plans; customer facts require source evidence or explicit overrides. Any new ambiguity is added with ID, owner, affected task, evidence and decision needed before dependent work proceeds.

## Target and dependency decision extension

D17: Root user intent, separate detection/recommendations, frozen rich target and explicit decision types are fixed by [ADR 0012](../01-architecture/ADR/0012-target-specification-and-dependency-resolution.md). D18: Resolve exact dependency graphs by required capability/usage and verify selected implementations; target-only changes retain unrelated source progress. These are specification decisions, not implemented features.

Validation owner RWT-014 must prove recursive feasibility, used-API and behavioral separation, decision provenance, target constraint enforcement and selective invalidation. Ecosystem metadata/API acquisition must record local evidence/freshness and handle offline UNKNOWN; no live package availability or version mapping is claimed by these docs. A future configure UI/API requires explicit contract work before exposure.

## Transparent execution decision

D19: Shared structured execution events/state, factual explanations, defined progress, timeline/deltas and persistent human control are baseline engine requirements under [ADR 0013](../01-architecture/ADR/0013-transparent-execution.md). Visible stages do not replace lifecycle states; existing SQLite events remain the ledger. Human watch does not change JSON framing; MCP reads bounded pages.

RWT-006 owns first live-state/heartbeat/watch/control validation; RWT-015/019 own safe pause/recovery; RWT-020 owns MCP/client cursor validation. Actual long-running liveness, event-store failure recovery, privacy and desktop/client behavior are untested until those implementation tasks run. Desktop rendering may follow; engine transparency may not.

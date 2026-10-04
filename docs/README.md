# Documentation guide

**Review baseline:** 2026-10-03. This repository contains specifications, schemas, and incremental SQL migration definitions. M0 has implemented/tested bounded Go foundation primitives, but no runnable engine binaries, runtime adapters or qualified migration corpus yet. Commands and package paths describe the implementation to build.

Start with [PRODUCT](00-vision/PRODUCT.md), then [SYSTEM-DESIGN](01-architecture/SYSTEM-DESIGN.md), [PROJECT-LIFECYCLE](02-domain/PROJECT-LIFECYCLE.md), and [DEVELOPMENT-PLAN](11-roadmap/DEVELOPMENT-PLAN.md). The [audit](11-roadmap/DOCUMENTATION-AUDIT.md) records gaps and the [decision register](11-roadmap/DECISION-REGISTER.md) distinguishes specification decisions from validation work.

## Authority and maintenance

[AGENTS.md](../AGENTS.md) establishes the authority order. A document's word “accepted” describes a specification decision, never implementation status or proof of testing. Lower-level summaries link to the owning specification and do not introduce competing definitions. Schemas and SQL encode the approved contracts; discrepancies are defects to fix together.

| Topic | Owner | Supporting detail |
| --- | --- | --- |
| Product scope and success | [PRODUCT](00-vision/PRODUCT.md) | [Goals](00-vision/GOALS.md), [non-goals](00-vision/NON-GOALS.md), [glossary](00-vision/GLOSSARY.md) |
| Package boundaries and ports | [SYSTEM-DESIGN](01-architecture/SYSTEM-DESIGN.md) | [Components](01-architecture/COMPONENTS.md), [service records](01-architecture/SERVICE-CONTRACTS.md), [ADRs](01-architecture/ADR/0011-development-baseline-and-contract-completion.md) |
| Three IR tiers and evidence | [BEHAVIOR-MODEL](02-domain/BEHAVIOR-MODEL.md) | [IR contracts](02-domain/IR-CONTRACTS.md), [domain entities](02-domain/DOMAIN-MODEL.md) |
| Shared assessment checkpoint | [PROJECT-LIFECYCLE](02-domain/PROJECT-LIFECYCLE.md) | [Report schema](../specs/reports/assessment-report.schema.json), [selection schema](../specs/project/assessment-selection.schema.json) |
| States and readiness | [PROJECT-LIFECYCLE](02-domain/PROJECT-LIFECYCLE.md) | [Confirmation protocol](02-domain/CONFIRMATION-PROTOCOL.md) |
| Approval subjects and execution closure | [CONFIRMATION-PROTOCOL](02-domain/CONFIRMATION-PROTOCOL.md) | [Execution manifest](../specs/project/execution-manifest.schema.json), [service qualification](01-architecture/SERVICE-CONTRACTS.md) |
| Analysis and omission | [ANALYSIS-SPEC](03-analysis/ANALYSIS-SPEC.md) | [Classification](03-analysis/CODE-CLASSIFICATION.md), [adapter](04-migration/CI3-ADAPTER.md) |
| Runtime capture and replay | [RUNTIME-OBSERVATION](03-analysis/RUNTIME-OBSERVATION.md) | [Sandbox](09-security/SANDBOX.md), [differential testing](05-verification/DIFFERENTIAL-TESTING.md) |
| User target and feasibility | [TARGET-SPECIFICATION](04-migration/TARGET-SPECIFICATION.md) | [Decision ADR](01-architecture/ADR/0012-target-specification-and-dependency-resolution.md), [target schema](../specs/project/target.schema.json) |
| Dependency usage and decisions | [DEPENDENCY-RESOLUTION](04-migration/DEPENDENCY-RESOLUTION.md) | [Resolution schema](../specs/migration/dependency-resolution.schema.json), [learning](07-learning/KNOWLEDGE-MODEL.md) |
| Framework skill packaging and composition | [SKILL-SPEC](06-agent/SKILL-SPEC.md) | [Adapter ports/qualification](01-architecture/SERVICE-CONTRACTS.md), [TIM](04-migration/TARGET-ARCHITECTURE.md) |
| Planning and generation | [MIGRATION-ENGINE](04-migration/MIGRATION-ENGINE.md) | [Laravel adapter](04-migration/LARAVEL-ADAPTER.md), [auth](04-migration/AUTH-AND-SESSION.md) |
| Equivalence and release evidence | [COMPATIBILITY-SPEC](05-verification/COMPATIBILITY-SPEC.md) | [Performance](05-verification/PERFORMANCE.md), [coverage](05-verification/COVERAGE.md) |
| Transparent execution | [EXECUTION-OBSERVABILITY](06-agent/EXECUTION-OBSERVABILITY.md) | [Event ledger](08-storage/EVENT-MODEL.md), [ADR 0013](01-architecture/ADR/0013-transparent-execution.md), [execution schemas](../specs/execution/state.schema.json) |
| CLI and MCP | [CLI-SPEC](06-agent/CLI-SPEC.md), [MCP-SPEC](06-agent/MCP-SPEC.md) | [Errors](06-agent/ERROR-CATALOG.md), [missions](06-agent/MISSION-PROTOCOL.md), [schemas](../specs/README.md) |
| SQLite, events, artifacts | [SQLITE-SCHEMA](08-storage/SQLITE-SCHEMA.md) | [Events](08-storage/EVENT-MODEL.md), [recovery](08-storage/RECOVERY.md) |
| Permissions and privacy | [SECURITY](09-security/SECURITY.md) | [Secrets](09-security/SECRET-HANDLING.md), [input checklist](10-development/PROJECT-INPUTS.md) |
| Contributor and release checks | [DEVELOPMENT](10-development/DEVELOPMENT.md) | [Testing](10-development/TESTING.md), [fixtures](10-development/FIXTURE-CORPUS.md), [cutover](10-development/CUTOVER-RUNBOOK.md), [release/install owner](10-development/RELEASE.md), [user installation guide](10-development/INSTALLATION.md), [implementation Skill usage](10-development/CONTRIBUTING.md#4-using-the-implementation-skill) |
| Development order | [DEVELOPMENT-PLAN](11-roadmap/DEVELOPMENT-PLAN.md) | [Milestones](11-roadmap/MILESTONES.md), [backlog](11-roadmap/BACKLOG.md) |

The [project plan](../REWROTRON_PROJECT_PLAN.md) summarizes the current milestone sequence and references its owners; historical phases are superseded. Each task in the development plan has an owner role, prerequisites, files, failure cases, verification, and completion evidence.

## Reading examples correctly

An example customer database, runtime, role, URL, or workflow is illustrative. It cannot become a detected baseline fact without evidence or an explicit user override. Framework support means a declared capability checked against a corpus, not that every possible application can be migrated automatically. `UNKNOWN` remains visible and prevents an equivalence claim for the affected scope.

## Additional canonical owners after red-team revision

CONFIRMATION-PROTOCOL owns subject serialization/human gates; PROJECT-LIFECYCLE owns state/readiness. IR-CONTRACTS owns neutral identity/versioning; BEHAVIOR-MODEL owns executable semantics and claim terminology. TARGET-ARCHITECTURE owns the Target Implementation Model; MIGRATION-UNITS owns acceptance scopes. DOMAIN-MODEL owns typed invalidation/reuse. ARTIFACT-STORE owns retention/quota/archive; RECOVERY owns process lifetime/publication/backup. LEARNING-SYSTEM owns promotion/independence/retraction. PERFORMANCE owns explicit statistical claims. These details are referenced, not redefined by overview documents/ADRs.

[Red-team resolution matrix](11-roadmap/RED-TEAM-RESOLUTION.md) is an audit/disposition record, not a competing specification. Attached review is retained unchanged as external architectural input. All current runtime/adapter/client claims remain unvalidated until corresponding evidence is produced.

[EVALUATION-CONTEXT](02-domain/EVALUATION-CONTEXT.md) owns the versioned observation-to-expression ABI and fixture/principal projection. BEHAVIOR-MODEL owns execution roles/response lowering; RUNTIME-OBSERVATION owns side-local identities/execution witnesses; DIFFERENTIAL-TESTING owns downstream pairwise correspondence and occurrence pairing. Review-2 dispositions extend the existing RED-TEAM-RESOLUTION audit record; the review contains two overlapping passes, which are mapped separately without making either a competing specification.

CI3 → Laravel is Reference Profile #1 intended for first qualification, currently SPECIFIED. References in adapter/static-routing/compatibility documents are profile details, not universal core assumptions. M8 qualifies pluggable directional packages; additional technologies and optional AI selection remain unimplemented/deferred.

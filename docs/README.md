# Documentation guide

**Review baseline:** 2026-10-03. This repository contains specifications, schemas, and initial SQL migration definitions. It has no Go engine, binaries, runtime adapters, or executed migration corpus yet. Commands and package paths describe the implementation to build.

Start with [PRODUCT](00-vision/PRODUCT.md), then [SYSTEM-DESIGN](01-architecture/SYSTEM-DESIGN.md), [PROJECT-LIFECYCLE](02-domain/PROJECT-LIFECYCLE.md), and [DEVELOPMENT-PLAN](11-roadmap/DEVELOPMENT-PLAN.md). The [audit](11-roadmap/DOCUMENTATION-AUDIT.md) records gaps and the [decision register](11-roadmap/DECISION-REGISTER.md) distinguishes specification decisions from validation work.

## Authority and maintenance

[AGENTS.md](../AGENTS.md) establishes the authority order. A document's word “accepted” describes a specification decision, never implementation status or proof of testing. Lower-level summaries link to the owning specification and do not introduce competing definitions. Schemas and SQL encode the approved contracts; discrepancies are defects to fix together.

| Topic | Owner | Supporting detail |
| --- | --- | --- |
| Product scope and success | [PRODUCT](00-vision/PRODUCT.md) | [Goals](00-vision/GOALS.md), [non-goals](00-vision/NON-GOALS.md), [glossary](00-vision/GLOSSARY.md) |
| Package boundaries and ports | [SYSTEM-DESIGN](01-architecture/SYSTEM-DESIGN.md) | [Components](01-architecture/COMPONENTS.md), [service records](01-architecture/SERVICE-CONTRACTS.md), [ADRs](01-architecture/ADR/0011-development-baseline-and-contract-completion.md) |
| Three IR tiers and evidence | [BEHAVIOR-MODEL](02-domain/BEHAVIOR-MODEL.md) | [IR contracts](02-domain/IR-CONTRACTS.md), [domain entities](02-domain/DOMAIN-MODEL.md) |
| States, gates, hashes | [PROJECT-LIFECYCLE](02-domain/PROJECT-LIFECYCLE.md) | [Confirmation protocol](02-domain/CONFIRMATION-PROTOCOL.md) |
| Analysis and omission | [ANALYSIS-SPEC](03-analysis/ANALYSIS-SPEC.md) | [Classification](03-analysis/CODE-CLASSIFICATION.md), [adapter](04-migration/CI3-ADAPTER.md) |
| Runtime capture and replay | [RUNTIME-OBSERVATION](03-analysis/RUNTIME-OBSERVATION.md) | [Sandbox](09-security/SANDBOX.md), [differential testing](05-verification/DIFFERENTIAL-TESTING.md) |
| User target and feasibility | [TARGET-SPECIFICATION](04-migration/TARGET-SPECIFICATION.md) | [Decision ADR](01-architecture/ADR/0012-target-specification-and-dependency-resolution.md), [target schema](../specs/project/target.schema.json) |
| Dependency usage and decisions | [DEPENDENCY-RESOLUTION](04-migration/DEPENDENCY-RESOLUTION.md) | [Resolution schema](../specs/migration/dependency-resolution.schema.json), [learning](07-learning/KNOWLEDGE-MODEL.md) |
| Planning and generation | [MIGRATION-ENGINE](04-migration/MIGRATION-ENGINE.md) | [Laravel adapter](04-migration/LARAVEL-ADAPTER.md), [auth](04-migration/AUTH-AND-SESSION.md) |
| Equivalence and release evidence | [COMPATIBILITY-SPEC](05-verification/COMPATIBILITY-SPEC.md) | [Performance](05-verification/PERFORMANCE.md), [coverage](05-verification/COVERAGE.md) |
| Transparent execution | [EXECUTION-OBSERVABILITY](06-agent/EXECUTION-OBSERVABILITY.md) | [Event ledger](08-storage/EVENT-MODEL.md), [ADR 0013](01-architecture/ADR/0013-transparent-execution.md), [execution schemas](../specs/execution/state.schema.json) |
| CLI and MCP | [CLI-SPEC](06-agent/CLI-SPEC.md), [MCP-SPEC](06-agent/MCP-SPEC.md) | [Errors](06-agent/ERROR-CATALOG.md), [missions](06-agent/MISSION-PROTOCOL.md), [schemas](../specs/README.md) |
| SQLite, events, artifacts | [SQLITE-SCHEMA](08-storage/SQLITE-SCHEMA.md) | [Events](08-storage/EVENT-MODEL.md), [recovery](08-storage/RECOVERY.md) |
| Permissions and privacy | [SECURITY](09-security/SECURITY.md) | [Secrets](09-security/SECRET-HANDLING.md), [input checklist](10-development/PROJECT-INPUTS.md) |
| Contributor and release checks | [DEVELOPMENT](10-development/DEVELOPMENT.md) | [Testing](10-development/TESTING.md), [fixtures](10-development/FIXTURE-CORPUS.md), [cutover](10-development/CUTOVER-RUNBOOK.md) |
| Development order | [DEVELOPMENT-PLAN](11-roadmap/DEVELOPMENT-PLAN.md) | [Milestones](11-roadmap/MILESTONES.md), [backlog](11-roadmap/BACKLOG.md) |

The original [project exploration](../REWROTRON_PROJECT_PLAN.md) supplies product context and examples. Its phase names, historical backlog, and illustrative structures are not a second implementation contract. Each task in the development plan has an owner role, prerequisites, files, failure cases, verification, and completion evidence.

## Reading examples correctly

An example customer database, runtime, role, URL, or workflow is illustrative. It cannot become a detected baseline fact without evidence or an explicit user override. Framework support means a declared capability checked against a corpus, not that every possible application can be migrated automatically. `UNKNOWN` remains visible and prevents an equivalence claim for the affected scope.

# GLOSSARY.md — Canonical Terminology & Definitions

> **Version:** 1.0  
> **Status:** Binding Vocabulary Definition  

---

This glossary establishes the unambiguous terms used across Rewrotron documentation, specifications, schemas, and source code. All agents must use these definitions consistently.

### Terms

- **Source Application (`SourceApp`)**  
  The existing legacy web application (e.g., CodeIgniter 3.x project) serving as the baseline oracle for behavior.

- **Target Application (`TargetApp`)**  
  The newly generated application written in the modern framework (e.g., Laravel 13.x on PHP 8.3+) that reconstructs the legacy behavior.

- **Behavior Model (`BehaviorModel`)**  
  The canonical, language-agnostic representation of an application's execution logic, state transitions, inputs, constraints, and side effects.

- **Behavior Case (`BehaviorCase`)**  
  A single, concrete, executable scenario testing a specific path through an entry point with explicit inputs, preconditions, expected outputs, and database mutations.

- **Behavioral Contract (`BehaviorContract`)**  
  The formal, framework-independent specification governing an operation or workflow, defining required status codes, headers, response schemas, and side effects.

- **Migration Unit (`MigrationUnit`)**  
  The smallest independently migratable and verifiable externally observable slice of functionality (e.g., an authenticated route, an API endpoint, or a coherent workflow).

- **Static Reachability Graph**  
  A directed graph mapping entrypoints (controllers, CLI commands, cron tasks) to downstream symbols, classes, methods, database tables, and external calls via static analysis.

- **Runtime Observation Trace (`ObservationTrace`)**  
  A structured, correlated log of an actual execution through the legacy application, capturing incoming HTTP payload, session state, database queries executed, cache access, and outbound response.

- **Golden Master (`GoldenMaster`)**  
  A stable, repeatable baseline recording of the legacy application's responses and side-effects under controlled preconditions, used for differential comparison.

- **Differential Verification (`DiffVerification`)**  
  The automated process of replaying an identical `BehaviorCase` against both the Source and Target applications and comparing the observed outputs and side effects.

- **Mismatch (`Mismatch`)**  
  A discrepancy recorded within a verification run. Its diagnostic classification is `EXPECTED_CHANGE`, `DEFECT`, or `UNKNOWN`; the run verdict remains `PASS`, `MISMATCH`, or `UNKNOWN`.

- **Finding (`Finding`)**  
  An analytical fact discovered during static analysis or runtime observation (e.g., dead code candidate, raw SQL concatenation, deprecated superglobal usage).

- **Mission (`Mission`)**  
  An atomic unit of work assigned to an agent or internal subsystem (e.g., `ANALYZE_SOURCE`, `MIGRATE_UNIT`, `REPAIR_MISMATCH`).

- **Skill (`Skill`)**  
  A modular, versioned package of instructions, rules, and scripts that instructs an agent how to execute a specific category of missions.

- **Experiment Fingerprint (`Fingerprint`)**  
  A deterministic cryptographic hash (SHA-256) of all inputs, source ASTs, migration templates, and database state defining an exact reproduction context.


- **Baseline:** Immutable confirmed source/environment facts identified by `configuration_hash`.
- **Plan subject:** Immutable units, scope, corpus, policy and target choices identified by `plan_hash`.
- **Release subject:** Target tree, dependency locks, verification, benchmark, session and runbook evidence identified by `release_hash`.
- **Completeness:** Whether a required trace dimension is COMPLETE, PARTIAL, UNAVAILABLE or NOT_APPLICABLE.
- **Fencing token:** Monotonically increasing mission lease revision; stale owners cannot publish results.
- **Scope exclusion:** A disclosed, explicitly reviewed omission from the migration's agreed behavior; it does not count as migrated or verified.
- **Expected change:** A separately approved contract exception and new policy subject, never a fourth verification verdict.

ID formats and enum owners are defined in [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md) and [TERMINOLOGY](../02-domain/TERMINOLOGY.md).

## Migration decision vocabulary

- **Detected environment:** Evidence-backed observed source facts, separate from user intent; corrections are recorded in the confirmed baseline.
- **Target specification:** User-owned requested runtime/framework/infrastructure/architecture/compatibility constraints, frozen as a planning revision and approved with Gate 2.
- **Recommendation:** An evidence-backed proposed action typed REQUIRED, RECOMMENDED, OPTIONAL, INFORMATIONAL or BLOCKING; never authority to mutate a target.
- **Feasibility:** Scoped assessment of whether required target capabilities and recursive constraints can be satisfied; not behavioral equivalence.
- **Dependency capability:** The input/output/state/error behavior actually required from a package/service, independent of package name.
- **Resolution strategy:** DIRECT_UPGRADE, ADAPTER, REPLACEMENT, REIMPLEMENT or RETAIN; BLOCKED is an unresolved status.
- **Selective invalidation:** Mark changed and transitively dependent artifacts stale while preserving unrelated current evidence and all history.

## Execution vocabulary

- **Execution stage:** A visible label for a repeatable/concurrent work instance, distinct from durable project lifecycle state.
- **Execution event:** A committed structured observation, decision, action or result with rationale and evidence; the existing SQLite ledger supplies order.
- **Execution state:** The canonical live projection answering current work, why, findings/changes, blockers, next action and user needs.
- **Progress basis:** A persisted scope/accounting revision and work definition supporting a metric's numerator/denominator; unavailable basis gives no guessed percentage.
- **Public explanation:** A factual evidence-linked justification; it does not expose hidden reasoning.
- **Human pause:** Persistent control acknowledged at safe checkpoints; resume/reconnect does not clear it.
- **Last-seen watermark:** A client-owned sequence used for return-session changes; reading does not approve or acknowledge work for other clients.

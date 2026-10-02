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
  A detected discrepancy between the Source and Target execution during differential verification, classified as either `EXPECTED_CHANGE`, `DEFECT`, or `UNKNOWN`.

- **Finding (`Finding`)**  
  An analytical fact discovered during static analysis or runtime observation (e.g., dead code candidate, raw SQL concatenation, deprecated superglobal usage).

- **Mission (`Mission`)**  
  An atomic unit of work assigned to an agent or internal subsystem (e.g., `ANALYZE_SOURCE`, `MIGRATE_UNIT`, `REPAIR_MISMATCH`).

- **Skill (`Skill`)**  
  A modular, versioned package of instructions, rules, and scripts that instructs an agent how to execute a specific category of missions.

- **Experiment Fingerprint (`Fingerprint`)**  
  A deterministic cryptographic hash (SHA-256) of all inputs, source ASTs, migration templates, and database state defining an exact reproduction context.

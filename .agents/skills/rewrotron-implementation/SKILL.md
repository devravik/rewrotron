---
name: rewrotron-implementation
description: Implement one authorized Rewrotron development unit using the repository's existing contracts, workflow, validation and durable checkpoints.
metadata:
  version: 1.0.1
  author: Rewrotron Core Team
---

# Rewrotron implementation

This is an execution procedure. Repository instructions and canonical owners remain authoritative; this skill adds no lifecycle, missions, approvals or progress model.

## Reconstruct and select

Discover applicable AGENTS.md files from the repository root through the affected paths. Read the [workflow index](../../../docs/AGENT-WORKFLOW.md) and [development plan](../../../docs/11-roadmap/DEVELOPMENT-PLAN.md); use the [milestone](../../../docs/11-roadmap/MILESTONES.md) and [backlog](../../../docs/11-roadmap/BACKLOG.md) indexes when relevant.

Reconstruct current state from Git, authoritative documentation, schemas, fixtures, tests and existing execution/mission/progress records. Conversation history is not authority. Inspect pending changes before choosing the smallest currently authorized unit whose prerequisites are satisfied.

Use the existing [agent model](../../../docs/06-agent/AGENT-MODEL.md), [rules](../../../docs/06-agent/AGENT-RULES.md) and [workflow](../../../docs/06-agent/AGENT-WORKFLOW.md) for mission ownership and gates. Distinguish repository-engine development from customer migration execution. During authorized bootstrap, before engine state exists, use existing Git checkpoints and the development plan; do not fabricate engine missions, confirmations or completion records.

## Implement and validate

Identify the unit's canonical owner through the [documentation index](../../../docs/README.md). Read only its relevant specification, accepted decisions, dependencies, schemas, vectors, fixtures, tests and implementation. Discover dependencies before editing.

Implement the smallest compliant unit. An unspecified semantic decision is a blocker to dependent work: identify its owner and record the missing rule. Do not silently redesign frozen architecture or extend the unit into deferred capabilities.

Follow [development](../../../docs/10-development/DEVELOPMENT.md) and [testing](../../../docs/10-development/TESTING.md): focused tests, authoritative contract validation, then broader required checks. Report actual evidence and unexecuted checks separately. The [skill specification](../../../docs/06-agent/SKILL-SPEC.md) owns revision recording and portability; record the consumed skill revision/content identity with the checkpoint.

## Optional context-cost optimization

Use pxpipe when available and compatible. It is optional context transport, never authority or a workflow prerequisite. Before relying on it, inspect the installed/current pxpipe version and verify a validated rendering profile for the **exact model ID**. Unknown model identity, unsupported/unvalidated profile or uncertain economics means normal text; another model's support is insufficient.

First remove unnecessary context, select relevant authoritative sources, reuse durable repository state and compact stale history when appropriate. Classify remaining blocks before compression:

- **Keep native:** code being edited, schemas, JSON/YAML contracts, SQL, hashes/digests, identifiers, paths, command arguments, credentials/secrets, machine-readable outputs, approval subjects, golden vectors, exact fixtures/errors and all byte-sensitive material where one character changes semantics. Do not image/transform these without independently verified losslessness; this policy grants no permission to expose secrets.
- **Candidates:** historical conversation, older closed tool history, verbose logs, exploratory output, bulk prose, redundant context and broad repository material used for understanding rather than exact reproduction. Mixed blocks retain precision-critical parts natively.

Use pxpipe only for eligible material with measured/known worthwhile token or cost savings, never merely because it is large. Prefer native text under uncertainty. Keep the immediate task, current instructions, active tool state, exact contracts and values native. Before precision-sensitive edits, re-read exact authoritative files and re-fetch any code, fields, identifiers, hashes or paths recalled from imaged context; cheaper recollection cannot replace repository evidence.

If newly enabled, perform a small repository-specific validation before broad use: prose, long tool output and historical context, plus code, JSON, schemas and hashes/identifiers as native-retention controls. Record limitations through existing lessons/decisions or the existing bootstrap checkpoint mechanism, not a new tracker. pxpipe output is not durable project state; successors must resume without it from the sources already named above.

## Record and continue

Record outcomes, evidence, blockers and next action in the existing state mechanism. Runtime mission completion uses guarded engine services. Bootstrap continuation uses Git's existing history: checkpoint task-owned changes with the unit, owning specification, exact checks/results, remaining scope and next authorized unit in the commit record. Do not create a separate status or handoff ledger.

Inspect the checkpoint for unintended changes and preserve unrelated work. Leave a successor enough repository evidence to reconstruct prerequisites and continue independently. Select the next authorized unit explicitly; do not treat completing this unit as authorization to build everything.

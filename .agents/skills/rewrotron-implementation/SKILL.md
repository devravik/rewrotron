---
name: rewrotron-implementation
description: Implement one authorized Rewrotron development unit using existing repository contracts, relevant available skills, focused validation and durable checkpoints.
metadata:
  version: 2.3.0
  author: Rewrotron Core Team
---

# Rewrotron implementation

Execution procedure for building Rewrotron itself. Applicable `AGENTS.md` files and canonical owners remain authoritative. This skill adds no workflow, milestones, missions, approvals or progress ledger. [Usage and examples](../../../docs/10-development/CONTRIBUTING.md#4-using-the-implementation-skill) are derivative instructions.

## 1. Reconstruct and select one unit

- Discover applicable `AGENTS.md` files; inspect Git status, recent checkpoints, existing tests/code and available progress/execution records. Conversation history is not authority. Preserve unrelated edits.
- Consult the active milestone section of [DEVELOPMENT-PLAN](../../../docs/11-roadmap/DEVELOPMENT-PLAN.md). Choose the smallest currently authorized unit with satisfied prerequisites; do not rebuild completed work unless relevant inputs changed or evidence conflicts.
- Follow [repository workflow](../../../docs/AGENT-WORKFLOW.md), [AGENT-MODEL](../../../docs/06-agent/AGENT-MODEL.md) and [AGENT-RULES](../../../docs/06-agent/AGENT-RULES.md). Before engine state exists, use Git/plan checkpoints; never fabricate missions or customer confirmations.
- Establish the unit's owner, dependency scope, outputs, acceptance checks and stopping condition before editing. If required semantics are absent or contradictory, name the exact gap and owner. Implement only an independent specified prerequisite when authorized; do not silently substitute another task or design the missing rule.

## 2. Find only relevant context and aids

- Use [docs/README](../../../docs/README.md) to locate topic owners. Search headings/symbols first; read bounded relevant sections and exact schemas, vectors, fixtures and existing code. Load another document only to answer a specific unresolved question.
- For development testing, diagnosis or validation, inspect the current agent's available skill catalog for a relevant existing helper. Load only the helper needed for this unit. Helper instructions are advisory beneath repository authority; unavailable helpers or their optional dependencies fall back to repository checks. Never create/install a skill simply to proceed.
- Domain migration/analysis/verification routing follows [SKILL-SPEC](../../../docs/06-agent/SKILL-SPEC.md) and `skills/manifest.json`. Use only eligible entries whose actual package exists and whose version/content identity can be verified. A placeholder hash or `STABLE` label alone proves nothing. Absent, unmatched or invalid entries fall back to canonical specs; the operation category is not a transformation capability.
- Inspect `tools/agent/capabilities.json` or run `tools/agent/bootstrap status` for optional aids. Availability/configuration flags do not prove operation, model compatibility, authentication or quality. Use only compatible approved aids; do not install or change interception/client configuration without explicit operator authorization. Absence never blocks work.
- Prefer exact search for location, bounded reads for code, a usable graph/index for impact narrowing, and native output for decisions/diagnostics. Indexes and output filters are not authority; re-read exact source and preserve exit codes/unmasked failure details.

Framework migration knowledge belongs in source/target capability packages under [SKILL-SPEC](../../../docs/06-agent/SKILL-SPEC.md), with executable behavior supplied by qualified adapters through neutral IR/TIM. This implementation procedure stays general: it implements the authorized repository unit, never chooses a mandatory framework pair or certifies generated migration output.

## 3. Preserve precision; use pxpipe only when qualified

Context economy order: omit unnecessary material → load relevant owners → reuse durable state → compact stale history where appropriate → consider eligible bulk transport → re-fetch exact bytes before acting.

**Always native:** current task/instructions/tool state; source being edited; schemas; JSON/YAML contracts; SQL; machine outputs; approval subjects; golden vectors/exact fixtures; errors; identifiers/hashes/digests/paths/arguments; secrets and all byte-sensitive data. Do not expose secrets. If one character can change semantics, keep exact text.

**Potential candidates:** bulk explanatory prose, historical conversation, older closed tool history, verbose noncritical logs and broad exploratory context. Large size alone is not eligibility. Prefer native text when exactness may matter; exact diagnostics within a bulk log remain native.

Use pxpipe when available and compatible, only for noncritical readable context with worthwhile measured/known economics. Before relying on it, check the installed/current version and a validated rendering profile for the exact active model identifier. Unknown, unsupported or unvalidated model means normal text; another model's profile or a bootstrap CONFIGURED flag is insufficient. Operator/runtime controls interception; do not issue self-wrapping session commands.

Before newly enabling bulk use, validate representative repository prose/log/history transport and native retention of code, JSON, schemas and identifiers/hashes. Record limitations through existing lesson/decision/checkpoint mechanisms. Re-fetch exact identifiers, paths, contracts and code from disk before precision-sensitive edits; expected savings never justify ambiguous image recollection. Compression is ephemeral: a successor must resume without pxpipe.

## 4. Implement and verify the bounded unit

- Discover direct dependencies and callers, inspect existing implementation, then make the smallest contract-compliant change. Use bounded slices where they reduce verification or recovery risk; avoid giant cross-milestone tasks.
- Add meaningful tests for required behavior and plausible wrong implementations. Use existing schemas/fixtures/vectors and test registry; do not invent expected semantics or alter valid tests/contracts to obtain a pass.
- Follow [DEVELOPMENT](../../../docs/10-development/DEVELOPMENT.md) and [TESTING](../../../docs/10-development/TESTING.md): focused tests → applicable contract validation → required broad checks. Before completion/commit, inspect actual exit codes and relevant outputs. Structural validity is not runtime evidence.
- On failure, distinguish contract gap, implementation defect, incorrect test, environment failure or transient failure before editing. Resolve the confirmed cause; no silent downgrade, blind effect retry or weakened acceptance. Report unmet prerequisites and unexecuted checks.

## 5. Record and checkpoint for another agent

- Update progress using the existing mechanism. During bootstrap, record the unit, owner, changed files, actual checks/results, blockers and next authorized unit in Git checkpoint records; after engine services exist, use their applicable mission/state protocol. Create no STATUS/HANDOFF file or second ledger.
- Record consumed skill/tool/knowledge versions and content identities under [SKILL-SPEC](../../../docs/06-agent/SKILL-SPEC.md). A helper's name or live mutable location is not a reproducibility pin.
- Inspect the scoped diff and required validation, then create a clean Git checkpoint consistent with [CONTRIBUTING](../../../docs/10-development/CONTRIBUTING.md). Do not include unrelated work, push automatically or claim unfinished work committed. Preserve safe unfinished state with exact pointers if checkpointing is blocked.
- Report authorized unit/owner, changed files, actual validation outcomes, remaining blockers, checkpoint reference/state and next authorized unit. Completed unit is not completed milestone. Stop at the requested boundary; the next agent reconstructs from repository state, not conversation summaries.

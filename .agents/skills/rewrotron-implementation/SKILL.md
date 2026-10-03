---
name: rewrotron-implementation
description: Implement one authorized Rewrotron development unit using the repository's existing contracts, workflow, validation and durable checkpoints.
metadata:
  version: 1.3.0
  author: Rewrotron Core Team
---

# Rewrotron implementation

This is an execution procedure. Repository instructions and canonical owners remain authoritative; this skill adds no lifecycle, missions, approvals or progress model.

## 1. Reconstruct and select

Discover applicable `AGENTS.md` files from the repository root through the affected paths. Do not load broad documentation upfront:

1. **Reconstruct durable state:** Inspect Git history (`git status`, `git log -n 1`), existing schemas, fixtures, tests and commit records. Conversation history is not authoritative state; durable repository records are.
2. **Locate active unit:** Consult [DEVELOPMENT-PLAN.md](../../../docs/11-roadmap/DEVELOPMENT-PLAN.md) only for the current milestone queue (M0–M8). Consult [MILESTONES.md](../../../docs/11-roadmap/MILESTONES.md) or [BACKLOG.md](../../../docs/11-roadmap/BACKLOG.md) only when explicitly relevant.
3. **Select minimal unit:** Choose the smallest currently authorized unit whose prerequisites are satisfied.
4. **Distinguish contexts:** Separate repository-engine bootstrap from customer migration execution. During bootstrap (before engine state exists), use Git checkpoints and the development plan; do not fabricate engine missions or confirmations. Follow [AGENT-MODEL.md](../../../docs/06-agent/AGENT-MODEL.md), [AGENT-RULES.md](../../../docs/06-agent/AGENT-RULES.md), and [AGENT-WORKFLOW.md](../../../docs/06-agent/AGENT-WORKFLOW.md).
5. **No redundant investigation:** Do not repeat an analysis already established by an authoritative repository artifact, passing validation, or a verified commit unless the current task introduces conflicting evidence or requires revalidation.

## 2. Context Router and progressive acquisition

Route and load context progressively using [docs/README.md](../../../docs/README.md) topic ownership:

* **Context budget:** Before loading a non-required document, state internally what unresolved question it answers. If no specific question exists, do not load it. Prefer one relevant section over a whole document. Stop acquiring context once the unit's implementation and validation requirements are sufficiently determined.
* **Search before read:** For large files, locate the required heading, symbol, test, table, or contract term before reading surrounding content. Use bounded reads whenever practical.
* **REQUIRED NOW:** The active unit definition in [DEVELOPMENT-PLAN.md](../../../docs/11-roadmap/DEVELOPMENT-PLAN.md), the relevant section of its topic owner, and the exact files/tests under edit.
* **OPTIONAL IF NEEDED:** Direct dependency specifications, accepted ADRs, and referenced schema definitions.
* **RE-READ BEFORE EDIT:** The exact target file, contract, or test fixture from disk. Never rely on conversational recall, prior summaries, or earlier tool outputs.
* **DO NOT LOAD:** Out-of-scope milestones, historical backlogs, broad architecture narratives without a specific unresolved question, and unchanged material whose relevant contents are already available in the current context.

## 3. Context economy and Precision Guard

Classify information into three tiers, governed by the **Precision Guard**: *If one character can change semantics, preserve and re-read exact bytes from disk.*

* **Tier 0 — Always native / exact (Precision Guard — never compress, approximate, or recall from memory):**
  * Source code being authored or edited.
  * SQL migrations, DDL statements, queries, and connection settings.
  * Schemas, JSON/YAML machine contracts, and API signatures.
  * Cryptographic hashes, checksums, digests, and golden vectors.
  * Identifiers, paths, table/column names, and command arguments.
  * Secrets, tokens, and credentials (never expose secrets).
  * Exact compiler errors, linter diagnostics, test failure traces, and machine-readable output.
  * All byte-sensitive material where a single character alters semantics.
* **Tier 1 — Potentially compressible (Optional transport):**
  * Explanatory architectural narratives, historical conversation turns, closed tool history, and verbose test logs.
* **Tier 2 — Omit (Never load or carry forward):**
  * Stale conversation history, redundant summaries, resolved investigation logs, and out-of-scope documentation.

*Optimization principle:* Prefer omitting context over compressing it; prefer targeted slices over full files; prefer fresh disk reads over recalled summaries.

## 4. Optional context compression (pxpipe policy)

`pxpipe` is an optional transport cost optimization, never a project dependency or workflow prerequisite. Its absence must never block or delay work:

* **Policy, not tool invocation:** Do not self-launch or run agent-wrapping commands (`pxpipe warp`, etc.). Interception is configured externally by the operator or runtime environment.
* **Support requirement:** Enable compression only when available and when the active runtime confirms a validated rendering profile for the exact model ID; otherwise operate entirely in native text.
* **Lossy transport boundary:** Imaged context can suffer visual confabulation on dense strings. Never compress Tier 0 material. Re-read any recalled identifier, hash, path, or code from disk before editing.
* **Non-durable:** Compressed context is ephemeral. Repository state, Git commits, and code artifacts must remain fully reconstructable without `pxpipe`.

## 5. Implementation, decomposition, and failure triage

Implement the authorized unit with calibrated effort and systematic triage:

* **Task decomposition:** Subordinate to the milestone plan, internally break multi-step units into bounded slices (e.g. `connection → discovery → validation → execution → targeted verification`). Each slice must define explicit scope, exact inputs, expected output, targeted validation, and a stopping condition. Do not decompose mechanically. Create a separate slice only when it reduces reasoning scope, validation scope, or recovery risk. Do not author monolithic multi-file changes across slices.
* **Reasoning budget calibration:**
  * *Low reasoning:* Routine boilerplate, mechanical edits, formatting, simple tests, standard file operations.
  * *Moderate reasoning:* Standard unit implementation, debugging test failures, mapping domain models.
  * *High reasoning:* Complex architectural conflicts, subtle concurrency/race conditions, cryptographic/canonical hashing contracts, or cross-spec invariant resolution.
* **Failure triage before edits:** When a command or test fails, classify the failure before modifying code:
  * `CONTRACT`: Failure violates a canonical specification → Re-read the authoritative contract; do not alter product code or tests to weaken contracts.
  * `IMPLEMENTATION`: Logic error in newly authored code → Inspect changed code and focused test diff.
  * `TEST`: Test expectation or fixture is defective → Re-verify test against owning specification before editing test.
  * `ENVIRONMENT`: Toolchain, OS, filesystem, or driver issue → Resolve environment/configuration; do not modify product code to mask environment failure.
  * `TRANSIENT`: Resource contention or timeout → Verify status and clean resources before retrying; do not rewrite working logic.
* **Verification sequence:** Follow [DEVELOPMENT.md](../../../docs/10-development/DEVELOPMENT.md) and [TESTING.md](../../../docs/10-development/TESTING.md). Run focused unit tests first, authoritative contract validation (`python3 tools/docs/validate.py` or schema checks) second, and required broad checks third. Record consumed skill revision/content identity with the checkpoint per [SKILL-SPEC.md](../../../docs/06-agent/SKILL-SPEC.md).
* **Portability:** Strictly agent- and platform-agnostic (Codex, agy, OpenCode, Hermes, etc.). Relies exclusively on standard files, terminal commands, and git operations.

## 6. Concise output and durable handoff

Minimize conversational token bloat and enable independent continuation:

* **No output bloat:** Do not echo full files, massive unified diffs, raw verbose logs, or repository documentation.
* **Completion report:** State only:
  * Authorized unit and owning specification.
  * Files modified or created.
  * Tests/validation executed and exact pass/fail results.
  * Unresolved issues or blockers, if any.
  * Git commit reference or checkpoint state.
  * Next authorized unit.
  Do not explain implementation details unless needed to understand a failure, tradeoff, contract deviation, or unresolved issue.
* **Durable handoff:** Create a Git checkpoint/commit according to the repository's existing checkpoint conventions when the authorized unit or slice is ready for checkpointing, recording the unit, owning spec, checks, and next unit. Subsequent turns or successor agents must reconstruct state from repository files and commits without relying on conversation history. Do not create a separate handoff database or status ledger.

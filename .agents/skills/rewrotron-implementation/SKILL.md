---
name: rewrotron-implementation
description: Implement one authorized Rewrotron development unit using the repository's existing contracts, workflow, validation and durable checkpoints.
metadata:
  version: 2.1.0
  author: Rewrotron Core Team
---

# Rewrotron implementation

This is an execution procedure. Repository instructions and canonical owners remain authoritative; this skill adds no lifecycle, missions, approvals or progress model.

## 1. Reconstruct and select

Discover applicable `AGENTS.md` files from the repository root through the affected paths. Do not load broad documentation upfront:

1. **Reconstruct durable state:** Inspect Git history (`git status`, `git log -n 1`), existing schemas, fixtures, tests and commit records. Conversation history is disposable; durable repository records are authoritative.
2. **Locate active unit:** Consult [DEVELOPMENT-PLAN.md](../../../docs/11-roadmap/DEVELOPMENT-PLAN.md) only for the current milestone queue (M0–M8). Consult [MILESTONES.md](../../../docs/11-roadmap/MILESTONES.md) or [BACKLOG.md](../../../docs/11-roadmap/BACKLOG.md) only when explicitly relevant.
3. **Select minimal unit:** Choose the smallest currently authorized unit whose prerequisites are satisfied.
4. **Discover efficiency capabilities:** Run `tools/agent/bootstrap status` or inspect `tools/agent/capabilities.json` to detect available capabilities (`context_compression`, `terminal_output_optimizer`, `code_graph`, `session_handoff`). Use approved capabilities when present; provision missing capabilities via `tools/agent/bootstrap install <cap>` only with explicit operator approval; otherwise fall back gracefully to native execution. Absence of optional tools never blocks work.
5. **Distinguish contexts:** Separate repository-engine bootstrap from customer migration execution. During bootstrap (before engine state exists), use Git checkpoints and the development plan; do not fabricate engine missions or confirmations. Follow [AGENT-MODEL.md](../../../docs/06-agent/AGENT-MODEL.md), [AGENT-RULES.md](../../../docs/06-agent/AGENT-RULES.md), and [AGENT-WORKFLOW.md](../../../docs/06-agent/AGENT-WORKFLOW.md).
6. **No redundant investigation:** Do not repeat an analysis already established by an authoritative repository artifact, passing validation, or a verified commit unless current evidence conflicts, a new detail is required, or relevant files changed.

## 2. Context Router and impact-based discovery

Route and load context progressively using [docs/README.md](../../../docs/README.md) topic ownership:

* **Context budget:** Before loading a non-required document, state internally what unresolved question it answers; if none, do not load. Stop acquiring context once requirements are determined.
* **Graph & impact narrowing:** Where a symbol index, ctags, or dependency graph (`code_graph`) is available, use it to narrow affected symbols, callers/callees, and impact surface before reading files. The graph is an index, never authority; always re-read exact source before editing.
* **Search before read:** Locate headings, symbols, or contract terms (e.g. via `rg` or symbol lookup) before reading surrounding text. Use bounded line ranges (`sed -n`, slice reads) whenever practical.
* **Capability skill routing:** For specialized migration, transformation, or verification units, consult `skills/manifest.json` via small bounded manifest lookup (or CLI `skills search`) for domain capabilities (`migration`, `analysis`, `verification`). The `operation` category designates execution entrypoints and is never selected as a capability. If a matching `STABLE` capability skill exists, load only its `SKILL.md` (and specific `references/` on demand). If absent, malformed, or unmatched, fall back directly to canonical specifications in [docs/README.md](../../../docs/README.md); skill absence never blocks execution.
* **Context Router:**
  * `REQUIRED NOW`: The active unit definition in [DEVELOPMENT-PLAN.md](../../../docs/11-roadmap/DEVELOPMENT-PLAN.md), the relevant section of its topic owner, and the exact files/tests under edit.
  * `OPTIONAL IF NEEDED`: Direct dependency specifications, accepted ADRs, and referenced schema definitions.
  * `RE-READ BEFORE EDIT`: The exact target file, contract, or test fixture from disk. Never rely on conversational recall or prior summaries.
  * `DO NOT LOAD`: Out-of-scope milestones, historical backlogs, broad architecture narratives without a specific unresolved question, and unchanged material whose relevant contents are already in context.

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

## 4. Tool output efficiency and selection principle

Use the cheapest reliable mechanism that answers the current question:

* **Tool selection hierarchy:**
  * Symbol or definition location → `rg` or symbol index (`code_graph`).
  * Surrounding implementation → bounded line read (`sed -n`, line slice).
  * Impact surface → dependency graph or symbol index if available.
  * Exact SQL, schema, or contract → direct file read from disk.
  * High-level bulk transport → pxpipe (`context_compression`) if available and eligible.
  * Decision or diagnostic verification → native command output.
* **Locate → narrow → inspect:** Minimize terminal output. Prefer `git diff --stat`, `git diff -- <file>`, `git log -n 3 -- <path>`, or targeted test commands over full repository or log dumps.
* **Terminal output filtering (RTK):** When RTK or similar output filters are detected via `tools/agent/bootstrap`, use them to suppress redundant command boilerplate. RTK is optional; its absence must never block work. Exact diagnostic output must remain unmasked when triaging failures or verifying contracts.

## 5. Optional context compression (pxpipe policy)

`pxpipe` is an optional transport cost optimization, never a project dependency or workflow prerequisite. Its absence must never block or delay work:

* **Policy, not tool invocation:** The skill defines *eligibility*; the operator/runtime environment controls *interception*. Do not run self-wrapping commands (`pxpipe warp`, etc.).
* **Support requirement:** Enable compression only when available and when the active runtime confirms a validated rendering profile for the exact model ID; otherwise operate entirely in native text.
* **Lossy transport boundary:** Imaged context can suffer visual confabulation on dense strings. Never compress Tier 0 material. Re-read any recalled identifier, hash, path, or code from disk before editing.
* **Non-durable:** Compressed context is ephemeral. Repository state, Git commits, and code artifacts must remain fully reconstructable without `pxpipe`.

## 6. Implementation, decomposition, and failure triage

Implement the authorized unit with calibrated effort and systematic triage:

* **Task decomposition:** Subordinate to the milestone plan, internally break multi-step units into bounded slices (e.g. `connection → discovery → validation → execution → targeted verification`). Each slice must define explicit scope, exact inputs, expected output, targeted validation, and a stopping condition. Do not decompose mechanically: create a separate slice only when it reduces reasoning scope, validation scope, or recovery risk. Do not author monolithic multi-file changes across slices.
* **Reasoning budget calibration:**
  * *Low reasoning:* Routine boilerplate, mechanical edits, formatting, simple tests, standard file operations.
  * *Moderate reasoning:* Standard unit implementation, debugging test failures, mapping domain models.
  * *High reasoning:* Complex architectural conflicts, subtle concurrency/race conditions, cryptographic/canonical hashing contracts, or cross-spec invariant resolution.
* **Failure triage before edits:** When a command or test fails, classify the failure before modifying code:
  * `CONTRACT`: Failure violates a canonical specification → Re-read authoritative contract; do not alter product code or tests to weaken contracts.
  * `IMPLEMENTATION`: Logic error in newly authored code → Inspect changed code and focused test diff.
  * `TEST`: Test expectation or fixture is defective → Re-verify test against owning specification before editing test.
  * `ENVIRONMENT`: Toolchain, OS, filesystem, or driver issue → Resolve environment/configuration; do not modify product code to mask environment failure.
  * `TRANSIENT`: Resource contention or timeout → Verify status and clean resources before retrying; do not rewrite working logic.
* **Verification sequence:** Follow [DEVELOPMENT.md](../../../docs/10-development/DEVELOPMENT.md) and [TESTING.md](../../../docs/10-development/TESTING.md). Run focused unit tests first, authoritative contract validation (`python3 tools/docs/validate.py` or schema checks) second, and required broad checks third. Record consumed skill revision/content identity with the checkpoint per [SKILL-SPEC.md](../../../docs/06-agent/SKILL-SPEC.md).
* **Portability:** Strictly agent- and platform-agnostic (Codex, agy, OpenCode, Hermes, etc.). Relies exclusively on standard files, terminal commands, and git operations.

## 7. Concise output, session boundaries, and durable handoff

Minimize conversational token bloat and enable independent continuation:

* **Caveman output discipline:** Prefer dense, structured status updates (`changed`, `validated`, `blocked`, `next`). Never repeat the user's prompt, narrate obvious tool actions, explain unchanged architecture, or dump full files/diffs. Never omit failure details, assumptions, or blockers necessary for decisions or auditability.
* **Pointers over prose:** Cite exact file paths, commit hashes, line ranges, and issue IDs instead of writing narrative explanations.
* **Completion report:** State only:
  * Authorized unit and owning specification.
  * Files modified or created.
  * Tests/validation executed and exact pass/fail results.
  * Unresolved issues or blockers, if any.
  * Git commit reference or checkpoint state.
  * Next authorized unit.
  Do not explain implementation details unless needed to understand a failure, tradeoff, contract deviation, or unresolved issue.
* **Session boundary transition:** When a session becomes context-heavy:
  1. Finish or checkpoint the current bounded unit/slice where safe.
  2. Persist durable state to Git and repository artifacts using existing mechanisms.
  3. Carry forward only minimal unresolved pointers.
  4. Resume in a fresh session from repository state rather than replaying conversation history. Never treat generated summaries as authoritative.
* **Durable handoff:** Create a Git checkpoint/commit according to the repository's existing checkpoint conventions when the authorized unit or slice is ready for checkpointing, recording the unit, owning spec, checks, and next unit. Subsequent turns or successor agents must reconstruct state from repository files and commits without relying on conversation history. Do not create a separate handoff database or status ledger.

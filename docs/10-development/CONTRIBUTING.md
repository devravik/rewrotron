# CONTRIBUTING.md — Code Review & Contribution Guidelines

> **Version:** 1.0  
> **Status:** Contribution Guidelines  

---

## 1. Commit Message Convention

All commits must follow Conventional Commits:
- `feat(analyzer)`: Add support for CI3 modular extensions
- `fix(comparator)`: Correctly normalize floating-point JSON numbers
- `docs(specs)`: Update CLI JSON schema for verify command
- `refactor(storage)`: Streamline SQLite transaction retries

---

## 2. Pull Request Checklist

1. Review [`AGENTS.md`](../../AGENTS.md) to ensure boundary rules are respected.
2. Follow [`docs/AGENT-WORKFLOW.md`](../AGENT-WORKFLOW.md).
3. Ensure no LLM or cloud API dependencies are introduced into `internal/`.
4. Ensure all unit and integration tests pass cleanly with `go test -race ./...`.


## 3. Review evidence

Include requirement/task IDs, consulted specification, exact checks and results, touched schemas/migrations/docs, and any unexecuted runtime profiles. Documentation-only changes use the document validator while the Go implementation is absent. Use [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md) as the implementation queue; do not claim a fixture or adapter is complete before its acceptance evidence exists.


## 4. Using the implementation skill

Use [rewrotron-implementation](../../.agents/skills/rewrotron-implementation/SKILL.md) to ask a coding agent to implement one authorized development unit **in Rewrotron itself**. It is an execution procedure over the existing repository workflow, not a customer migration operator, model connection or autonomous background process. The [installation guide](INSTALLATION.md) describes the separate planned end-user tool/operator Skill experience.

### Start a session

Open your coding agent with the Rewrotron repository as its working directory. If the client discovers repository Skills, select `rewrotron-implementation` through its own interface, or use this prompt:

> Use the rewrotron-implementation skill. Reconstruct current progress from Git and existing repository records. Identify the next currently authorized implementation unit in DEVELOPMENT-PLAN.md whose prerequisites are satisfied. Implement the smallest compliant unit, run required validation, and leave a durable checkpoint with the evidence and next authorized unit. Do not invent unspecified semantics or redesign frozen architecture.

If the client does not discover the Skill, use the file-reading fallback:

> Read .agents/skills/rewrotron-implementation/SKILL.md from this repository and follow its procedure for the next currently authorized implementation unit. Obey applicable AGENTS.md instructions and the owning specifications. Reconstruct state from the repository rather than conversation history.

Skill discovery varies by client; these prompts do not imply tested compatibility with every agent. A client must be able to read repository files and perform the permitted development actions. No provider-specific session commands or extra Rewrotron model credentials are required by the Skill.

### Request a bounded unit

When you already know the unit, replace the bracketed description before sending:

> Use rewrotron-implementation for [specific development unit]. Check its current authorization and prerequisites, read its owning contracts and relevant code/tests, then implement and validate only that unit. If required semantics are missing or contradictory, report the exact owner and gap instead of guessing.

The prompt names work; it does not waive repository instructions, prerequisite checks or customer execution gates. The [development plan](../11-roadmap/DEVELOPMENT-PLAN.md) remains the sole implementation queue. Do not ask the Skill to build the entire platform in one unit.

### Resume or switch agents

Open the same repository/working tree in the successor agent and send:

> Use rewrotron-implementation. Resume from Git, existing repository progress records and any available execution/mission state. Inspect unfinished changes before selecting work; preserve unrelated edits. Continue only the next currently authorized unit whose prerequisites are satisfied. Do not rely on the previous agent's conversation history.

Keep Git checkpoints and required repository artifacts available to the successor. Before engine state exists, the Skill uses Git and the development plan; it must not fabricate missions or approvals. Once applicable engine state exists, use its existing records as defined by the authoritative workflow. No separate STATUS/HANDOFF file or tracking system is introduced by these instructions.

### Practical examples to copy

These are prompts to send to your coding agent while it is working in the Rewrotron repository. They are not shell commands, a second task queue, or instructions to implement all the examples together. Named units remain subject to the current plan and prerequisite checks; if already complete, inspect their recorded evidence instead of rebuilding them.

**Example 1 — Start the next M0 unit without remembering previous work**

> Use rewrotron-implementation. Inspect Git status, recent checkpoints and the M0 section of docs/11-roadmap/DEVELOPMENT-PLAN.md. Determine what is already implemented and select the smallest remaining authorized M0 unit with satisfied prerequisites. Implement that one unit, run the required checks, and checkpoint the result. Report what is complete and which unit can follow; do not claim all of M0 is complete from this one change.

**Example 2 — Work specifically on the SQLite migration foundation**

> Use rewrotron-implementation for the next ready portion of the M0 SQLite initialization and migration foundation. Read docs/08-storage/SQLITE-SCHEMA.md, the relevant recovery rules, existing SQL migrations and storage tests/code. Confirm the pinned driver/toolchain requirements and prerequisites first. Select one bounded change, test its success and failure behavior against the contracts, and leave a checkpoint. Do not rewrite applied migrations or start runtime adapters.

The storage owner is [SQLITE-SCHEMA](../08-storage/SQLITE-SCHEMA.md); [Recovery](../08-storage/RECOVERY.md) owns upgrade/recovery behavior. This example does not assert that storage is always the next ready unit.

**Example 3 — Implement a bounded source-snapshot change**

> Use rewrotron-implementation for a ready M0 source-snapshot acquisition unit. Check the current implementation and prerequisites, then follow the owning sections of docs/03-analysis/ANALYSIS-SPEC.md and the referenced snapshot contracts. Select the smallest missing supported behavior, include relevant X02 controls, validate it, and checkpoint. Do not execute the legacy application or broaden the snapshot boundary without an authoritative rule.

Use this only when the [development plan](../11-roadmap/DEVELOPMENT-PLAN.md) and existing evidence establish that its prerequisites are ready. X02 is defined by [Testing](TESTING.md).

**Example 4 — Repair a real failing test**

First paste the exact command and its actual failure output below the prompt, keeping precision-sensitive diagnostics in native text:

> Use rewrotron-implementation to investigate the attached failing test. Reproduce the failure and classify it using the Skill's failure triage. Check the owning specification before changing code or expectations. Fix the smallest confirmed cause, rerun the focused test and required validation, and checkpoint the evidence. If the failure exposes an unspecified rule, report the exact gap rather than inventing a fix.

Do not manufacture a failure for this example or weaken a valid test to obtain a pass.

**Example 5 — Continue in another agent or a fresh session**

> Use rewrotron-implementation. This is a fresh session with no trusted prior conversation context. Inspect the same repository's Git checkpoints, unfinished changes, existing progress records and any applicable execution/mission state. Verify the last unit's recorded result where needed, then continue the next currently authorized unit with satisfied prerequisites. Preserve unrelated work and leave enough durable evidence for the next agent.

This works by reading shared repository state, not by requiring a particular agent brand or forwarding an entire chat transcript.

**Example 6 — Finish the current bounded unit before stopping**

> Use rewrotron-implementation to bring only the current authorized unit to a safe stopping point. Run its applicable validation, record the actual results and unresolved work through the existing checkpoint mechanism, and create a Git checkpoint when consistent with the repository workflow. Identify the next authorized unit, but do not start it in this session. If unfinished work cannot safely be committed, report its exact working-tree state instead of claiming completion.

### What to expect at the checkpoint

The Skill's completion report points to the authorized unit/owner, changed files, actual validation results, remaining blockers, Git checkpoint and next authorized unit. Its own procedure owns how to produce that evidence; this guide does not redefine it. A successful unit does not imply a complete milestone, working migration engine or tested client integration.

Optional efficiency tools such as pxpipe are governed by the Skill and are not required to start or resume. Repository files, exact contracts, tests and durable records remain authoritative. See [Development](DEVELOPMENT.md), [Testing](TESTING.md), and [repository workflow](../AGENT-WORKFLOW.md) for their owning requirements.

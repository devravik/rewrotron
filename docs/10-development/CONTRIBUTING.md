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

### What to expect at the checkpoint

The Skill's completion report points to the authorized unit/owner, changed files, actual validation results, remaining blockers, Git checkpoint and next authorized unit. Its own procedure owns how to produce that evidence; this guide does not redefine it. A successful unit does not imply a complete milestone, working migration engine or tested client integration.

Optional efficiency tools such as pxpipe are governed by the Skill and are not required to start or resume. Repository files, exact contracts, tests and durable records remain authoritative. See [Development](DEVELOPMENT.md), [Testing](TESTING.md), and [repository workflow](../AGENT-WORKFLOW.md) for their owning requirements.

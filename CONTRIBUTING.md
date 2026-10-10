# Contributing to Rewrotron

**Owner:** Contributor guidelines & code review standards  
**Detailed Guide:** [docs/10-development/CONTRIBUTING.md](docs/10-development/CONTRIBUTING.md)  
**Agent Skill:** [`.agents/skills/rewrotron-implementation/SKILL.md`](.agents/skills/rewrotron-implementation/SKILL.md)  
**Constitution:** [`AGENTS.md`](AGENTS.md)  

---

## 1. Unit-Based Development Discipline

Rewrotron strictly follows a **one-authorized-unit-per-run** development discipline:
1. Every engineering task is decomposed into an authoritative unit graph in [`docs/11-roadmap/DEVELOPMENT-PLAN.md`](docs/11-roadmap/DEVELOPMENT-PLAN.md).
2. Work proceeds strictly through: **Inspect → Implement → Validate → Commit → Stop**.
3. Do not combine multiple roadmap units in a single pull request. Keep changes focused and atomic.

---

## 2. Invariants You Must Never Break

- **Preservation of Uncertainty:** When source behavior is dynamic or unknown, report `UNKNOWN`. **`UNKNOWN` never becomes `PASS` by assertion, waiver, or missing capture.**
- **Neutral Core:** Core IRs (`Source IR`, `Application IR`, `Behavior IR`, `Target Implementation Model`) remain framework- and language-neutral. Never import framework-specific packages into core domain logic.
- **Human Gates:** Gates 1, 2, and 3 are persistent human authorization points. Code cannot auto-approve human gates.
- **Falsification First:** New verifiers, comparators, or sandboxes must include falsification tests proving they catch deliberate semantic regressions and unauthorized access.

---

## 3. Pull Request Checklist

Before submitting a pull request, ensure:
1. `go test -race ./...` passes without race conditions.
2. `go vet ./...` passes with zero issues.
3. `python3 tools/docs/validate.py` passes (all schemas, fixtures, and local documentation links valid).
4. Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat(scope): ...`
   - `fix(scope): ...`
   - `docs(scope): ...`
   - `test(scope): ...`

---

## 4. Using the Implementation Skill

If working with AI coding agents, instruct the agent to use the repository's native implementation skill:
```text
Use .agents/skills/rewrotron-implementation/SKILL.md.
Follow the authoritative DEVELOPMENT-PLAN.md queue.
Execute one authorized unit, validate with race tests, and commit atomically.
```

For detailed prompts and contributor examples, see [docs/10-development/CONTRIBUTING.md](docs/10-development/CONTRIBUTING.md).

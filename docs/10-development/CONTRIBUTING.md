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

1. Review [`AGENTS.md`](file:///media/System/rewrotron/AGENTS.md) to ensure boundary rules are respected.
2. Follow [`docs/AGENT-WORKFLOW.md`](file:///media/System/rewrotron/docs/AGENT-WORKFLOW.md).
3. Ensure no LLM or cloud API dependencies are introduced into `internal/`.
4. Ensure all unit and integration tests pass cleanly with `go test -race ./...`.

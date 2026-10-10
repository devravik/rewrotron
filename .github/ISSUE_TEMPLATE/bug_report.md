---
name: Bug Report
about: Create a report to help us improve Rewrotron
title: "fix(scope): [short description of bug]"
labels: ["bug"]
assignees: ""
---

### Summary of Defect

A clear and concise description of the bug.

### Affected Scope & Component

- **Component:** (e.g. `internal/domain`, Source Adapter, Target Adapter, Comparator, CLI, MCP)
- **Profile / Technology:** (e.g. CodeIgniter 3, Laravel 13, Go stdlib)
- **Roadmap Unit:** (if applicable)

### Steps to Reproduce

1. Provide minimal reproduction commands or fixture.
2. Run command: `go test -race -v ...` or CLI invocation.
3. Observe unexpected behavior or error.

### Expected Behavior

What you expected to happen according to the authoritative specifications (`docs/` and `specs/`).

### Actual Output & Diagnostics

```text
Paste exact failure output, test traces, or diff here.
```

### Environment Information

- **OS & Architecture:** (e.g. Linux amd64, Ubuntu 24.04)
- **Go Version:** (`go version`)
- **Git Commit Hash:** (`git rev-parse HEAD`)

### Invariant Check

- [ ] Does this bug involve `UNKNOWN` falsely turning into `PASS`?
- [ ] Does this involve sandbox containment or secret redaction?
- [ ] Does this violate human authorization gates (Gate 1, 2, or 3)?

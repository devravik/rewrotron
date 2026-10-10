---
name: Feature / Adapter Request
about: Propose a new capability, technology adapter, or roadmap unit
title: "feat(scope): [short description of proposal]"
labels: ["enhancement"]
assignees: ""
---

### Proposed Capability

Describe the capability or adapter you would like Rewrotron to support (e.g. Django source adapter, Spring Boot target adapter, new verification dimension).

### Rationale & Problem

Why should this be added? What legacy applications or modern stacks does this address?

### Neutral Core Alignment

- How does this compose through `Source IR` → `Application IR` → `Behavior IR` → `Target Implementation Model`?
- Confirm that no framework-specific dependencies are added to the deterministic neutral core:
  - [ ] Yes, this capability cleanly separates through pluggable adapter ports (`SourceAdapter` / `TargetAdapter`).

### Falsification & Verification Plan

How will this capability be tested?
- [ ] Bounded reference fixture identified or created.
- [ ] Differential execution witnesses and compatibility test cases defined.
- [ ] Negative falsification control included (deliberate defect caught as `MISMATCH`).

### Applicable Specifications

Which specifications in `docs/` or `specs/` govern this work?

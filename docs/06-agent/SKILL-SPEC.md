# SKILL-SPEC.md — Agent Skill Architecture & Packaging Specification

> **Version:** 1.0  
> **Status:** Agent Skill Reference Specification  

---

## 1. Structure of a Rewrotron Skill

The Rewrotron Skill teaches an LLM agent how to operate Rewrotron correctly. It is packaged in `skills/rewrotron/`:

```text
skills/rewrotron/
├── SKILL.md                  # Main entrypoint with YAML frontmatter & core instructions
├── references/               # Detailed cheat-sheets loaded on demand
│   ├── lifecycle.md          # State machine transitions
│   ├── ci3.md                # CodeIgniter 3 architecture quirks
│   ├── laravel13.md          # Laravel 13 best practices
│   ├── compatibility.md      # Equivalence & comparison rules
│   ├── troubleshooting.md    # Common mismatch remedies
│   └── safety.md             # Sandbox constraints & redaction
└── scripts/                  # Deterministic helper scripts
    ├── bootstrap.sh          # Initialize project environment
    ├── inspect.sh            # Inspect traces and mismatches
    └── report.sh             # Generate human-readable summaries
```

---

## 2. Frontmatter Contract

`SKILL.md` must declare standard frontmatter:

```yaml
---
name: rewrotron
description: Reconstructs legacy CodeIgniter 3 applications into modern Laravel 13 using behavior modeling and automated differential verification.
version: 1.0.0
author: Rewrotron Core Team
---
```

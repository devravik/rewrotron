# SKILL-SPEC.md — Agent Skill Architecture & Packaging Specification

> **Version:** 1.0  
> **Status:** Agent Skill Reference Specification  

---

## 1. Structure of a Rewrotron Skill

The Rewrotron Skill teaches an LLM agent how to operate Rewrotron correctly. Its planned packaging is `skills/rewrotron/`:

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
description: Operates guarded Rewrotron services for evidence-bounded application reconstruction.
version: 1.0.0
author: Rewrotron Core Team
---
```


## 3. Portability and operating contract

The Skill is a planned external orchestration package, not engine implementation. Its default flow uses the canonical CLI/MCP catalogs, waits for persisted gate approvals, claims fenced missions, distinguishes unsupported/UNKNOWN behavior, and completes only with verified current evidence. Reference documents load on demand; scripts invoke fixed domain commands and never provide a shell/SQL escape hatch.

Record Skill version/content hash when used; a client without skill support can perform the same flow manually through CLI. Test with at least two client integrations, interrupted/resumed sessions and a blocked gate. No client brand is required by core services.

No Skill/client integration is implemented or validated here. Pin exact consumed revisions, never use evolving live instructions as approved execution truth. CLI fallback supports all decision actions/gates.

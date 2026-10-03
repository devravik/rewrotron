# SKILL-SPEC.md — Agent Skill Architecture & Packaging Specification

> **Version:** 1.1
> **Status:** Agent Skill Reference Specification

---

## 1. Skill Packaging and Directory Layout

The repository distinguishes execution entrypoints from domain capability packages:

* `.agents/skills/`: Agent CLI execution entrypoints discovered directly by coding agents (such as `.agents/skills/rewrotron-implementation/`).
* `skills/`: Repository capability packages and engine operator skills, indexed by `skills/manifest.json`:
  * `skills/rewrotron/`: Core engine operator skill.
  * `skills/capabilities/<category>/<skill-name>/`: Scoped domain capabilities (e.g., `skills/capabilities/migration/ci3-controller-to-laravel/`).

The planned packaging for the core engine operator is `skills/rewrotron/`:

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

## 2. Frontmatter and Capability Package Contract

`SKILL.md` must declare standard frontmatter:

```yaml
---
name: rewrotron
description: Operates guarded Rewrotron services for evidence-bounded application reconstruction.
version: 1.0.0
author: Rewrotron Core Team
---
```

### Capability skill packages

Specialized capability skills (such as migration, analysis, and verification packages) follow a modular structure:

```text
skills/capabilities/<category>/<skill-name>/
├── SKILL.md                  # Concise entrypoint (< 100 lines) with frontmatter
└── references/               # Deep cheat-sheets loaded only on demand
    ├── mapping.md            # Syntax / AST transformation table
    ├── edge-cases.md         # Framework idiosyncrasies and exceptions
    └── verification.md       # Focused unit test and verification commands
```

Capability skills must declare extended frontmatter:

```yaml
---
name: ci3-controller-to-laravel
description: Transforms CI3 controllers, methods, and view loaders to Laravel 13 single-action or resource controllers.
version: 1.0.0
status: STABLE
category: migration
source_framework: codeigniter@3.*
target_framework: laravel@13.*
tags:
  - controller
  - http
  - routing
  - view-loader
provenance:
  lessons:
    - les-01h8abc
author: Rewrotron Core Team
---
```

**Advisory invariants:** Capability skills are strictly advisory aids for target generation and repair. They cannot grant Gate 1, Gate 2, or Gate 3 approval, cannot override or alter deterministic verifier verdicts (PASS, MISMATCH, UNKNOWN), cannot turn UNKNOWN into PASS, cannot declare code DEAD, and cannot alter canonical machine contracts or schemas.

### Canonical content-hash preimage (`rwt-skill-package-1`)

To ensure deterministic content-addressing across clients and environments without divergence:

1. Enumerate all regular files within the skill package directory (`SKILL.md`, `references/*`, `scripts/*`), excluding VCS and ephemeral files.
2. Represent each file by its relative POSIX path, raw byte length (`size_bytes`), and SHA-256 digest of raw file bytes (`sha256`, 64 lowercase hex characters).
3. Construct a canonical JSON manifest with key `"identity": "rwt-skill-package-1"` and `"entries"` sorted strictly by unnormalized UTF-8 path bytes.
4. Serialize using [rwt-cjson-1](../02-domain/CONFIRMATION-PROTOCOL.md).
5. Compute the SHA-256 digest of the serialized UTF-8 bytes: `content_hash = "sha256:" + lowercase_hex(SHA256(preimage))`.

---

## 3. Skill Routing Manifest

The repository maintains a repository skill discovery index at `skills/manifest.json` (a non-authoritative discovery cache; canonical specifications remain the sole authority):

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "version": "1.0",
  "skills": [
    {
      "id": "operation.rewrotron",
      "name": "rewrotron",
      "version": "1.0.0",
      "status": "STABLE",
      "category": "operation",
      "tags": ["core", "engine", "orchestration", "assessment", "migration"],
      "description": "Operates guarded Rewrotron services for evidence-bounded application reconstruction.",
      "path": "skills/rewrotron",
      "content_hash": "sha256:0000000000000000000000000000000000000000000000000000000000000000"
    }
  ]
}
```

### Deterministic routing semantics

1. When working on a specialized unit, an agent consults `skills/manifest.json` via small bounded manifest lookup (or CLI `skills search --query <domain>`).
2. **Category filtering:** Capability skill routing evaluates ONLY domain capability categories (`category IN ('migration', 'analysis', 'verification')`). The `operation` category designates execution entrypoints or engine operator skills (such as `operation.rewrotron`), which govern process flow and are never selected as domain transformation capabilities.
3. **Status filtering:** Only capability skills with `status IN ('STABLE')` are eligible for automatic routing (or `EXPERIMENTAL` when explicitly requested).
4. **Metadata filtering & optional bounded selection:** The router applies deterministic low-cost metadata filtering against trigger tags and framework predicates. When multiple candidate skills match, an optional `bounded_decision` capability (such as TypeSafe Jev via `TYPESAFE_API_KEY`) may provide an advisory choice ranking to select 1–2 skills; if absent or disabled, deterministic tag rank is used directly.
5. The agent loads only the selected `SKILL.md` (and specific `references/*.md` on demand).
6. **Absence/malformation fallback:** If `skills/manifest.json` is absent, malformed, or unreadable, the agent falls back directly to canonical specifications in `docs/`. Absence or corruption of the manifest never blocks execution.

See [SKILL-EVOLUTION](../07-learning/SKILL-EVOLUTION.md) for lifecycle states and promotion criteria, and [LEARNING-SYSTEM](../07-learning/LEARNING-SYSTEM.md) for multi-context validation rules.

---

## 4. Portability and operating contract

The Skill is a planned external orchestration package, not engine implementation. Its default flow uses the canonical CLI/MCP catalogs, waits for explicit human assessment continuation decisions and persisted gate approvals, claims fenced missions, distinguishes unsupported/UNKNOWN behavior, and distinguishes assessment completion from implementation acceptance. An assessment-only run can complete with blocked/unknown feasibility; migration unit completion still requires scoped current verification. Reference documents load on demand; scripts invoke fixed domain commands and never provide a shell/SQL escape hatch.

Record Skill version/content hash when used; a client without skill support can perform the same flow manually through CLI. Test with at least two client integrations, interrupted/resumed sessions and a blocked gate. No client brand is required by core services.

No Skill/client integration is implemented or validated here. Pin exact consumed revisions, never use evolving live instructions as approved execution truth. CLI fallback supports all decision actions/gates.

The Skill presents the shared assessment report and may stop successfully there; it must not automatically plan or generate. An explicit human local assessment decision selects continuation, followed by a separate plan request and Gate 2 approval. It cannot treat report completion, recommendation acceptance or conversation as authorization. See [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md).

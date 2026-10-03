# Rewrotron

Rewrotron is a vendor-neutral legacy web application reconstruction engine being designed around observable behavior, canonical intermediate representations and differential verification.

**Current status:** Specification baseline. This repository contains documentation, machine schemas and the incremental SQL migration definitions. The Go engine, CLI/MCP binaries, runtime fixtures and adapters have not been implemented or tested yet.

## Architecture

```text
Source adapters → Source IR → Application IR → Behavior IR
    → assessment report → explicit human selected-work/continuation decision
    → exact target resolution + executable plan → explicit approval
    → target adapters → differential verification/repair
```

Core services remain framework/LLM-independent. Migration units are routes, operations or workflows. Initial reference work targets CodeIgniter 3.1.x → Laravel 13/PHP 8.3+, with preserved schemas and independent sandbox databases. Generic PHP begins with inventory/extraction; additional languages and frameworks are roadmap candidates.

Every customer migration enforces persisted source confirmation (Gate 1), exact target/plan confirmation (Gate 2), and release/cutover confirmation (Gate 3). Unknown observations cannot pass. Dead-code omission requires current evidence; source files remain unchanged.

Transparent execution is a core product requirement alongside behavioral preservation: every long-running operation must show its work, factual rationale, discoveries, changes, blockers and next action through shared committed events/state. [Execution observability](docs/06-agent/EXECUTION-OBSERVABILITY.md) defines live inspection, honest progress, timelines, human decisions and pause/recovery. These are specified requirements; implementation remains pending.

## Start here

- [Documentation index](docs/README.md) — authoritative specifications and reading order.
- [Development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md) — dependency-ordered tasks, files, interfaces, failure cases and acceptance evidence.
- [Documentation audit](docs/11-roadmap/DOCUMENTATION-AUDIT.md) — identified gaps and resolutions.
- [Decision register](docs/11-roadmap/DECISION-REGISTER.md) — fixed engineering choices and remaining validation/operational inputs.
- [Product](docs/00-vision/PRODUCT.md), [system design](docs/01-architecture/SYSTEM-DESIGN.md), [repository constitution](AGENTS.md).
- [CLI](docs/06-agent/CLI-SPEC.md), [MCP](docs/06-agent/MCP-SPEC.md), [schema registry](specs/README.md).
- [Development checks](docs/10-development/DEVELOPMENT.md) and [customer inputs](docs/10-development/PROJECT-INPUTS.md).

The [project plan](REWROTRON_PROJECT_PLAN.md) summarizes current sequencing; the development plan/specifications own implementation decisions. Planned command examples in the CLI specification become executable as its tasks are completed.

## Validate the specification artifacts

```bash
python3 -m pip install 'jsonschema>=4.10,<5'
python3 tools/docs/validate.py
```

This checks documentation links, schemas/references, contract examples, registries and incremental SQLite DDL and approval vectors. It does not establish runtime migration correctness.

## License

[Apache-2.0](LICENSE).

The [red-team resolution report](docs/11-roadmap/RED-TEAM-RESOLUTION.md) maps findings to canonical owners, dispositions and remaining experiments. No finite corpus implies universal equivalence.

Assessment is a successful stopping checkpoint in the [shared workflow](docs/02-domain/PROJECT-LIFECYCLE.md). Its report may expose blocked/unknown migration feasibility; it is neither an executable plan nor implementation approval. Optional continuation revalidates persisted evidence and requires separate work selection, planning and Gate 2 approval. These capabilities are SPECIFIED, not implemented.

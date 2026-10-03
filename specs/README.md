# Machine specification registry

**Status:** Version 1.0 initial contracts; no deployed CLI/MCP client compatibility is claimed. Schemas use JSON Schema draft 2020-12 and stable logical IDs under `https://rewrotron.dev/specs/`. IDs are identifiers, not a requirement for a hosted website or runtime network resolution.

## Groups

| Directory | Contract |
| --- | --- |
| `common.schema.json` | IDs/hashes/artifacts, multivalue headers, evidence, errors, entrypoints, completeness |
| `project/` | Intent config, prepared source subject, persisted confirmation, current snapshot |
| `ir/` | Source/Application/Behavior IR and discovered data models |
| `behavior/` | Contracts, concrete cases, input-only scenarios, bounded workflow bindings |
| `schemas/` | Detection and observation trace documents |
| `migration/` | Plan, target patch manifest and approved release subject |
| `reports/` | Inventory reconciliation, evidence, intentional changes and release limitations |
| `verification/` | Deterministic run and performance reports |
| `missions/` | Queue/lease/completion record |
| `adapters/` | Versioned capabilities, maturity, profiles and limitations |
| `cli/` | Envelope, complete command registry and command input schemas |
| `mcp/` | Complete tool registry, per-tool input schemas and empty-queue result |
| `examples/` | Positive/negative contract validation fixtures |

`cli/commands.json` and `mcp/tools.json` are catalogs rather than instance schemas. Each schema reference must resolve locally. CLI success data uses its registered result schema; source prepare has a declared response variant. Dry-run/help/version rendering follows CLI-SPEC and never invents execution evidence.

Strict canonical objects reject unknown fields to catch spelling/version drift. Descriptive metadata, rule representation data, assertions, provider-specific fixture configuration and namespaced extensions allow structured content but cannot bypass service semantic validation. Cross-project identity, hash recomputation, graph references/cycles, approval trust, normalization relations and ownership require deterministic code checks in the development plan.

A versioned field addition changes compatible instance fixtures; a required field/removal changes the contract version after deployment. Update every affected CLI/MCP/doc example and storage migration when needed. The initial SQL migration replaces only an undeployed draft, not an existing customer DB.

## Check

```bash
python3 tools/docs/validate.py
```

The check validates metaschemas, offline references, both catalogs, positive/negative instances, local Markdown links, initial SQL/document parity, creation/foreign-key/immutability/transaction checks. Go runtime/driver/race/adapter checks remain implementation tasks.

## Target and dependency decision schemas

`project/target.schema.json` is the shared rich target object referenced by project intent and plans. Its additions preserve flat v1 language/framework fields; language names remain open and adapter-validated. Root/split YAML is parsed strictly before validating the JSON data model. `project/recommendation.schema.json` records typed evidence-backed proposals; `migration/dependency-resolution.schema.json` records capability mappings, strategy and decision status.

The plan's additive `decision_system` section binds frozen intent, exact resolved target and immutable recommendation/dependency/feasibility artifacts. Existing documents without the section remain structurally readable; new approval/execution services must require it and rebuild incomplete historical plans. Shape validation is not approval or feasibility: service fixtures must check graph constraints, evidence freshness, capability completeness, explicit human decisions and selective invalidation. Existing CLI/MCP flag-based planning calls remain accepted; target/style inputs are now optional assertions so planning can use configuration alone. SQL DDL is unchanged.

## Execution contracts

`execution/` contains measured progress, structured committed events, live state and cursor-page schemas. Snapshot `execution` is additive so existing stored snapshots remain valid; new services must provide it. Visible stages/control labels are separate from lifecycle and verifier enums. State/event/approval references require service ownership/freshness checks; JSON Schema cannot prove arithmetic, sequence ordering, gate authority, safe checkpoints or event redaction.

CLI status gains optional watch/verbosity; `execution show`, human `pause`/`continue` and MCP `project.get_execution` have synchronized catalog/input/output schemas. JSON watch remains rejected so existing single-envelope framing is preserved. No SQL DDL or MCP transport change is introduced. [EXECUTION-OBSERVABILITY](../docs/06-agent/EXECUTION-OBSERVABILITY.md) owns semantics and runtime acceptance.

# IR records, provenance, and versioning

**Status:** Normative record contract. Companion schemas are under `specs/ir/`.

## Shared identity and evidence

Each IR document has `schema_version: "1.0"`, `project_id`, `configuration_hash`, and a stable document ID. A reference must resolve within the same project and baseline. IDs use lowercase ASCII letters/digits, `_`, and `-`; hashes are lowercase 64-character SHA-256 hex. References to artifacts use `sha256:<digest>`, resolved only through the artifact store.

An evidence record includes an `evidence_id`, `kind` (`STATIC`, `RUNTIME`, `USER_OVERRIDE`, `AGENT_PROPOSAL`), `artifact_ref`, optional symbol/trace references, and a description. Confidence expresses an estimate, not proof. `AGENT_PROPOSAL` cannot establish an observation, unlock a gate, prove DEAD, or set a verifier verdict. All derived records list evidence IDs. Immutable IR documents are stored as artifacts; searchable records are indexed in SQLite.

## Source IR

| Record | Required fields | Meaning |
| --- | --- | --- |
| File | `file_id`, `path`, `sha256`, `syntax_status` | Content identity; `PARSED`, `PARTIAL`, `UNSUPPORTED`, `FAILED` |
| Symbol | `symbol_id`, `file_id`, `name`, `kind`, `start_line`, `end_line` | Neutral class/function/method/constant/property/closure declaration |
| Edge | `edge_id`, `from_id`, `to_id` or null, `kind`, `resolved`, `evidence_ids` | Call/include/inheritance/reference; unresolved edge retains expression and affected scope |
| Finding | `finding_id`, `code`, `severity`, `affected_ids`, `evidence_ids`, `message` | Parse failure, dynamic dispatch, environment dependency, or coverage gap |

Raw backend ASTs may be auxiliary artifacts. Parser-specific node types never cross the adapter boundary. Positions are 1-based lines; spans are inclusive. A partial parse retains inventory and findings but cannot justify a closed reachability proof for that region.

## Application IR

Entrypoints have `entrypoint_id`, `type`, `dispatch`, `handler_symbol_ids`, `evidence_ids`, and `coverage_state`. Types are `http`, `cli`, `cron`, `event`. HTTP dispatch includes method/path/priority; CLI includes executable identity and argument schema; cron includes schedule/timezone and command; event includes channel and message schema. Method and path are required only for HTTP.

Components use neutral roles such as handler, service, data model, renderer, or integration. Dependency edges distinguish reads, writes, dispatch, and state sharing. Models list discovered tables/columns, keys, nullability, types, collation and relationships; schema evidence is required before DB comparisons.

Operations can be reached by multiple entrypoints. Workflows reference ordered operation steps and explicit bindings between outputs and later inputs. The graph is not a containment tree: a single operation or authentication component can participate in several workflows.

Source adapters resolve their conventions into these records. Unknown handlers, dynamic schedules, callbacks, route precedence, and plugin loading retain an unresolved finding. The generic core does not invent handlers.

## Behavior IR

`BehaviorIR` lists immutable `contract_refs`, `case_refs`, workflow records, evidence IDs, and open questions. Contract IDs name operations; contract `version` is a semantic version string; storage `revision` is a positive integer. Cases bind a precise contract version and fixture identity. A corpus is an ordered case list plus fixture and policy hashes.

Business rules have IDs, descriptions, evidence IDs, a representation (`DECLARATIVE`, `OBSERVED_ONLY`, `UNRESOLVED`), and optional structured data. An observed example cannot specify all unseen input combinations. A target adapter can accept only representations it declares it can implement; unresolved computations produce a blocker. Placeholder expressions in documentation are explanatory and are not evaluated as PHP, SQL, shell, or an unrestricted expression language.

Cases specify concrete inputs, fixture IDs, principal IDs from synthetic accounts, timeout, redirects, assertion modes, and required observation dimensions. Workflow steps explicitly extract and bind dynamic values. Fresh independent cookie jars and DB copies are used per side; a workflow preserves each side's state across its own steps.

## Revision rules and rejection behavior

Unknown schema versions fail `ERR_SCHEMA_VERSION`; unknown canonical fields fail schema validation. Optional `extensions` use namespaced keys and never change canonical semantics. Adding a required field requires a new schema version and compatibility fixtures. Changed IR content creates a new immutable artifact and invalidates dependent plans through hashes.

Unresolved IDs, duplicate IDs, cross-project references, reversed spans, cycles in inheritance, nonexistent handler symbols, contradictory route priorities, and evidence pointing outside the approved baseline are rejected with `ERR_INVALID_CONTRACT`. Workflow cycles may exist in the application but must have an explicit bounded executable scenario; the migration dependency DAG condenses strongly connected units.


## Persistent identity versus document-local IDs

Document-local symbol/file/unit/case IDs are scoped by baseline or plan. SQLite row IDs are unique versioned record identities, built from the baseline/plan hash prefix and local ID or allocated deterministically by the ID port. Repository indexes retain the local ID in their canonical document; queries always include project and baseline/plan. A stable local symbol name cannot overwrite a prior baseline's evidence row.

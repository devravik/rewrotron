# Components and package ownership

**Status:** Planned package map. See [SYSTEM-DESIGN](SYSTEM-DESIGN.md) and ADR 0011. These paths do not currently contain an implementation.

| Package | Responsibility | Permitted inputs and outputs |
| --- | --- | --- |
| `cmd/rewrotron` | Entry point and signal handling | CLI request → application services → envelope/exit code |
| `cmd/rewrotron-mcp` | Entry point and stdio lifecycle | JSON-RPC → MCP handler → same application services |
| `internal/runtime` | Composition and adapter registration | Concrete constructors → injected neutral ports |
| `internal/domain` | Records, typed IDs, enums, errors, ports | Standard library and canonical types only |
| `internal/core` | Gate-checked application services | Domain requests → persisted state and operation results |
| `internal/project` | Configuration, manifests, workspace opening | Root `rewrotron.yaml`/legacy input → validated intent; generated views have no approval authority |
| `internal/detection` | Neutral source/environment proposal coordination | Bounded adapter facts → separate immutable detection artifacts |
| `internal/recommendations` | Evidence-backed proposed actions and decision views | Facts + requested target + compatibility evidence → typed recommendations; never mutates intent |
| `internal/target` | Rich target validation, immutable revisions and impact analysis | User configuration → frozen target subject and selective invalidation |
| `internal/dependency` | Recursive constraints, actual usage, capabilities and strategies | Adapter inventories + evidenced usage + target → feasibility/resolution artifacts and blockers |
| `internal/analysis` | Framework-neutral graph and reachability algorithms | Source/Application IR → evidence-backed classifications |
| `internal/observer` | Capture coordination, correlation, redaction | Sandbox adapter observations → canonical trace artifacts |
| `internal/behavior` | Workflow synthesis and contract validation | Application IR + traces → Behavior IR and case corpus |
| `internal/migration` | Unit DAG, plan synthesis and patch coordination | Behavior IR + target capabilities → confirmed plan and patches |
| `internal/verification` | Replay orchestration and neutral comparisons | Case + source/target traces + policy → deterministic report |
| `internal/execution` | Structured instrumentation, repeatable stage/task projections, measured progress, explanations and human control views | Committed events/operations/missions/findings/confirmations → one live state/timeline; no approval or verdict authority |
| `internal/missions` | Queue, leases, retries, evidence-based completion | Valid state + dependencies → fenced mission lease |
| `internal/benchmark` | Workload coordination and comparison | Frozen corpus + measurements → benchmark report |
| `internal/knowledge` | Project lesson records and opt-in promotion | Verified experiments → sanitized versioned knowledge |
| `internal/cli` | Argument parsing, envelopes, human rendering | Public CLI catalog, no business or framework rules |
| `internal/mcp` | Protocol and tool/resource dispatch | Schema-validated inputs; trusted human approval handling |
| `internal/storage/sqlite` | Domain repository implementations and migrations | Bound SQL parameters; transactional records/events |
| `internal/storage/artifacts` | Content addressing, staging, reads, retention | Redacted bytes → immutable artifact references |
| `internal/sandbox` | Container lifecycle and controlled tool runner | Fixed executable specifications; no unrestricted shell |
| `adapters/source/php/parser` | PHP parser integration | PHP bytes → neutral syntax inventory |
| `adapters/source/php/codeigniter3` | CI3 routing, loaders, hooks, session extraction | Source IR → framework-resolved Application IR and traces |
| `adapters/source/php/generic_php` | Generic PHP inventory and limited entrypoint extraction | Explicit unsupported findings; no assumed CI3 conventions |
| `adapters/target/php/laravel13` | Laravel blueprint, generation, harness | Canonical behavior rules → idiomatic target patch |
| `tools/php`, `testdata` | Probe sources and synthetic corpus | Versioned, hashable fixture artifacts |

## Dependency enforcement

`internal/domain` imports no other repository package. Neutral services can import domain and other neutral services in an acyclic graph. Framework adapters import domain records/ports; the composition package imports concrete adapters and storage. `internal/core`, `internal/behavior`, and `internal/verification` never import concrete framework packages, SDKs, prompts, or parser-specific AST types.

PHP parsing, CI3 hook semantics, Laravel middleware, and framework template generation are adapter responsibilities. Shared graph algorithms and JSON/session/DB comparison are neutral responsibilities. Table identifiers belong to discovered data models, not hardcoded framework conventions.

No second artifact store or normalization implementation is introduced. `internal/storage/artifacts` owns durable bytes; `internal/verification` owns comparison. `internal/observer` invokes a shared redactor before persistence. Enforce the import graph with an architecture check in RWT-001 and fixtures in RWT-007/008.

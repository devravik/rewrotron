# Target architecture and Target Implementation Model

**Owner:** Target-specific implementation choices between Behavior IR and generated code. Core stores neutral references/ownership; target adapters interpret their own lowerings. [target-model.schema.json](../../specs/migration/target-model.schema.json) defines current 2.1 immutable model shape.

## Purpose and lifecycle

Behavior IR specifies observed obligations, not component organization. The Target Implementation Model names exact resolved environment, component/module ownership, shared consumers, safe target paths, behavior/contract mappings, lowering choices, runtime/build capabilities, implementation constraints and unresolved capability gaps. It is not another source IR and cannot invent behavior or rewrite source evidence.

A proposal is built before Gate 2, with model_id, positive revision, project identity, Behavior IR/resolution refs and explicit component records. Component owner_unit_id is unique; consumer units and prerequisites must agree with the plan DAG. Paths cannot overlap owners ambiguously. Lowering choices (including validation, transactions, sessions, framework defaults) are target-adapter choices backed by constraints/capabilities. Missing in-scope capabilities block plan feasibility.

Gate 2 hashes the model ref; activation occurs only after approval. No plan_hash is embedded in that model, avoiding a circular hash. Execution uses [progress.schema.json](../../specs/migration/progress.schema.json) keyed by model ref/plan/entity: PLANNED → GENERATING → BUILDING → VERIFYING → VERIFIED, or BLOCKED. Progress and target tree hashes never mutate the approved model.

Changing ownership/lowering/resolution creates a successor model and Gate 2. Code patches implementing existing choices preserve Gate 2 and invalidate applicable verification/release. Shared component changes reverify all currently implemented consumers in the active acceptance scope. Future consumers receive a pending obligation checked when they are implemented; they cannot block the current owner unit. Target publication validates safe paths, before hashes, resource claims, current subjects and a recoverable multi-file journal. Human edits conflict rather than being overwritten.

## Reference versus universal model

The planned Laravel modular-monolith adapter may lower to controllers, services/actions, data access and views. These are adapter conventions, not core mandatory classes. Go standard-library targets use framework=null in exact resolution; Rust/Python targets may need other component organizations. Required runtime/build capabilities and contract mappings persist across all choices. No target generates solely from source file names.

Generated code manifests map every owned path to component/unit/contract and pinned generation revisions. Scaffold/lock/build/configuration/compatibility harness/report artifacts are part of the release closure. A clean build is runtime evidence; a schema-valid model is only structural evidence. The first falsification milestone builds the smallest model needed for a stateful synthetic scenario before broader generation.

## Qualified provider lowerings

Current model 2.1 components contain lowering_bindings: capability_id, exact provider_manifest_ref, qualification_ref, executable obligation_refs and input/output shape refs. The provider manifest is an adapter manifest 2.1 for explicit built-in registration or 2.2 for package-callable binding; qualification is the existing typed qualification artifact for the exact profile/schema/implementation. Services match the capability, schemas and control suite to each required obligation. lowering_choice is explanatory text, not executable authority. Primitive core capabilities need no provider binding; each nonprimitive requirement (password verification, session codec, template rendering) must resolve exactly one qualified binding or remain a blocking gap. Providers used for target generation never become observation/comparison authority. M1 uses only independently controlled fixture providers, not arbitrary PHP/authentication schemes.

## Target skill boundary

A target framework skill packages the adapter/provider lowerings, architecture conventions, templates, dependency constraints and harness assets described by [SKILL-SPEC](../06-agent/SKILL-SPEC.md). TIM is its explicit contract with neutral behavior: every choice retains obligation and provider qualification refs. No target class, mobile widget, platform API or framework lifecycle leaks backward into neutral IR as a universal requirement. Represent genuine source-independent behavior only through supported versioned facets; unsupported facets remain gaps. The generator consumes approved TIM and cannot establish acceptance. A co-distributed harness retains separate role/qualification and cannot self-certify generation.

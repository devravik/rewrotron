# Migration framework skills and agent skill packaging

**Owner:** Skill packaging, discovery and the boundary between guidance and executable adapter capabilities. **Status:** SPECIFIED. [SERVICE-CONTRACTS](../01-architecture/SERVICE-CONTRACTS.md) owns deterministic ports and qualification; [TARGET-ARCHITECTURE](../04-migration/TARGET-ARCHITECTURE.md) owns the Target Implementation Model (TIM). No framework skill or migration profile is runtime-qualified yet.

## Existing layouts and distinct responsibilities

| Package | Role | Authority |
| --- | --- | --- |
| `.agents/skills/rewrotron-implementation/` | Portable procedure for implementing Rewrotron itself, one authorized development unit at a time | References repository owners; no framework semantics or second workflow |
| `skills/rewrotron/` (planned) | Operator guidance for the shared assessment/migration workflow | Calls existing guarded CLI/MCP; grants no gates |
| `skills/capabilities/analysis/<source-name>/` (planned) | Source framework skill | Packages source knowledge and references a SOURCE adapter |
| `skills/capabilities/migration/<target-name>/` (planned) | Target framework skill | Packages target lowering knowledge and references a TARGET adapter |
| `skills/capabilities/verification/<name>/` (planned) | Probe, harness or comparator guidance | References separately qualified observation/comparison providers |

Reuse these category directories and `skills/manifest.json`; do not create a parallel source/target registry. Source and target names identify individual technologies, not mandatory pairs. CI3 and Laravel packages are Reference Profile #1 candidates, currently SPECIFIED. Their adapter specifications remain [CI3-ADAPTER](../04-migration/CI3-ADAPTER.md) and [LARAVEL-ADAPTER](../04-migration/LARAVEL-ADAPTER.md).

A migration framework skill is a capability package, not merely prompts. It contains concise `SKILL.md` guidance plus only the necessary declarative metadata, mappings, templates, fixtures and deterministic adapter assets or references. The engine consumes installed adapter manifests and qualified ports, not natural-language instructions as executable semantics. Adding a framework supplies a package and adapter implementation through existing ports: built-in registration may require a rebuild, and separately installed invocation becomes available only after the specified external binding is implemented and qualified. A genuinely new neutral semantic capability requires an explicit versioned contract decision; a skill cannot extend the behavior interpreter implicitly.

## Source and target contracts

| Concern | Source skill | Target skill |
| --- | --- | --- |
| Identity/profile | Framework/language detection rules, supported version/profile declarations | Target framework/language, build/runtime/platform profile declarations |
| Deterministic implementation | Parser/analyzer, conventions, source semantics, extraction/probe assets | TIM lowerings, component/structure/dependency constraints, templates/generators and harness assets |
| Neutral boundary | Source → Source IR → Application IR → Behavior IR with provenance/uncertainty | Behavior obligations → proposed TIM → approved TIM → target implementation |
| Capabilities | Extracted/observed constructs, unresolved dynamic behavior, limitations and capture coverage | Supported operators/effects, required target facilities, unresolved lowering gaps |
| Evidence | Detection/extraction/capture controls and source fixtures | Build/lowering controls, target fixtures and independently executed verification |
| Agent guidance | Optional interpretation and characterization hints | Optional implementation/repair hints; no acceptance authority |

Neither package depends on the other's framework identity to interpret neutral contracts. A pair-specific recipe may remain advisory, but cannot replace neutral IR, TIM or exact-profile admission. No source skill emits Laravel classes into neutral IR; no target skill requires CI3 controller names as input semantics.

## Machine-readable contract: reuse existing records

| Information | Existing owning record |
| --- | --- |
| Discovery/category/tags/framework selectors | `skills/manifest.json` discovery entry and SKILL frontmatter; advisory filtering only |
| Package identity | Skill version and `rwt-skill-package-1` content hash |
| Executable role/identity/bytes | [Adapter manifest](../../specs/adapters/manifest.schema.json) 2.2 for package-callable bindings: direction, adapter_id/version, implementation_ref, binding and operations; 2.1 for existing explicit built-in registration |
| Advertised framework/version/capabilities/limitations | Manifest capabilities, supported_profiles and limitations; profiles are descriptive, never executable constraints |
| Exact supported input and control evidence | [Qualification](../../specs/adapters/qualification.schema.json): schema_refs, capability_ids, input_profile_ref, control suite/results, limitations/status |
| Source-required semantics and uncertainty | IR/contract/scenario records plus actual capture/coverage attestations |
| Target-required facilities and lowering providers | Exact resolution and TIM 2.1 component capabilities, gaps and lowering_bindings (provider manifest, qualification, obligation/shape refs) |
| Independent comparison capability | Pinned comparator/probe/runner bindings, qualifications and approved comparison policies |

Source/target packages must contain `package.json` conforming to [package.schema.json](../../specs/skills/package.schema.json); its structured adapter reference, not guidance prose, identifies the exact adapter manifest. Consumed skill revisions and installed adapter bytes/schema/qualification closure are pinned through [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md); current registries cannot reconstruct historical approvals. Package-callable adapters use branch 2.2 of the existing adapter manifest; this is not a second adapter manifest format. Branch 2.1 remains usable for explicit built-in registration only, with the existing exact qualification checks. A generic executable profile constraint language is an explicit M8 prerequisite, not inferred from open supported_profiles JSON.

## Composition and admission

1. Discover source and target packages independently; detection remains an evidence-backed proposal.
2. Validate installed manifests, exact bytes/schema/profile qualifications and applicable source extraction/capture capabilities through existing services.
3. Derive required neutral obligations from characterized application behavior and selected scope, not from package names or a union of advertised tags.
4. Resolve each required target lowering through TIM and each required observation/comparison dimension through independently qualified providers. Match exact obligation, input/output shape, profile, implementation and schema refs.
5. Record unsupported/unresolved obligations as typed capability gaps. Missing lowering blocks affected executable planning; missing required capture/comparison remains UNKNOWN. Assessment can still finish with gaps.
6. Freeze the selected package/provider revisions in the exact approved plan. Execute only after Gate 2; independent verifier derives PASS / MISMATCH / UNKNOWN.

Capability tag intersection is insufficient: `routing` and `navigation` are not equivalent without an executable neutral obligation and qualified lowering. Partial compatibility requires an explicit resolution or scope decision and a successor feasible plan; approval cannot waive missing correctness evidence. Neither qualifications of each individual adapter nor a compatible schema alone qualify the composed migration profile: run its end-to-end positive and deliberate-negative controls under the named scope.

The initial architecture admits exact qualified profiles. M8 demonstrates pluggable composition with a substantially different technology and specifies any additional neutral facets/constraints before executing them. React Native/Expo → Kotlin/Jetpack Compose/Room/Android APIs is one future candidate, not a commitment or currently supported profile. Reactive/native lifecycle, navigation, offline persistence or platform effects that lack qualified semantics remain explicit gaps; never flatten them into HTTP.

## Package structure and guidance

`SKILL.md` has name/description and version metadata; optional references load on demand. A package may include deterministic mappings/templates/fixtures or reference separately distributed pinned adapter bytes. Do not execute arbitrary package scripts through the agent interface: helpers use fixed guarded commands/qualified Runner ToolSpecs under the existing security/gate policy.

Discovery entries retain existing id/name/version/status/category/tags/description/path/content_hash fields. Framework selectors used for guidance are descriptive; no `source_framework` plus `target_framework` pair is mandatory. `operation` entries are not transformation capabilities. `STABLE` is a distribution/routing label, never runtime maturity or evidence. Missing package, placeholder digest or absent qualification cannot establish a usable migration capability. The current index is empty because no domain capability/operator package is present; `.agents` implementation guidance is independently discoverable by coding agents.

Routing uses deterministic category/tag/profile filtering and stable tie-breaking by entry ID; load only relevant guidance. No match falls back to canonical documentation for investigation, never to unsupported execution. JEV and other bounded-decision integrations are deferred; there is no required AI selector. Agents may propose code through their existing coding capability, but the engine and verifier retain authority. Framework knowledge belongs in these packages, not in the general implementation procedure.

## Canonical package identity (`rwt-skill-package-1`)

1. Enumerate regular package files, excluding VCS/ephemeral files.
2. Record `path` (relative POSIX path), raw `size_bytes` and `sha256` (64 lowercase hex digits over raw bytes), using [package.schema.json inventory](../../specs/skills/package.schema.json#/$defs/inventory). Refuse symlinks, special files, traversal/absolute paths and duplicate paths; inventory all published package files, including package.json and SKILL.md. Excluded VCS/ephemeral files never become executable inputs.
3. Form `{identity: "rwt-skill-package-1", entries: [...]}` with entries strictly sorted by unnormalized UTF-8 path bytes.
4. Serialize with rwt-cjson-1 and hash those bytes; prefix the lowercase digest with `sha256:`.

Pin all executable assets and consumed revisions separately through the approved closure. The guidance hash does not prove adapter bytes are installed, safe or qualified. Updating a package cannot retroactively change historical approval/evidence.

## Verification and portability

Source/target skills may supply control fixtures, platform harnesses and test guidance. Expected obligations are independently reviewed and approved; generation assets do not establish actual capture truth or comparator authority. A target package cannot mark its own output PASS, broaden normalization, fill missing effects or turn UNKNOWN/UNSUPPORTED_CAPABILITY into PASS. Any co-distributed harness/comparator must be independently qualified and receive only its role-specific handles.

Use the shared assessment checkpoint, mission/state system and three human gates. Skills introduce no new workflow. CLI remains the complete fallback when a client lacks skill support; client support remains UNVERIFIED until integration evidence exists. [TESTING](../10-development/TESTING.md) X40/X58/X61/X67 owns neutrality, profile admission, checker separation and lowering falsification; M8 adds composition controls to X40.

## Deterministic package resolution and installation

Discovery ≠ executable binding ≠ composed-profile qualification. The following are independent admission facts, not a second lifecycle/state database:

| Fact | Required evidence |
| --- | --- |
| Discoverable | Entry in the existing discovery index; no executable authority |
| Resolvable | Exact package ID/version, verified rwt-skill-package-1 content hash and schema-valid package.json with exact adapter manifest digest |
| Executable | The pinned adapter has an available, validated BUILTIN or EXTERNAL binding under SERVICE-CONTRACTS; existence/labels alone are insufficient |
| Qualified | Engine accepts exact applicable implementation/schema/profile/control evidence; package guidance cannot confer this |
| Usable | Selected composition satisfies current obligations, TIM lowerings, independent verification capabilities and normal gates |

Installing means acquiring and verifying bytes in a private staging location, publishing immutable local package contents, then making discovery metadata available. It does not execute install hooks, start an adapter, qualify a profile, or authorize customer execution. Acquisition follows SECURITY/SANDBOX. Current external resolution/loading is SPECIFIED, not IMPLEMENTED or QUALIFIED; built-in composition may require registering the implementation and rebuilding the distribution. No package-install CLI is claimed here.

Resolution is deterministic: select exact index ID/version (never latest/range); verify its content_hash over the package; read root package.json; require package_id/package_version equality with discovery and SKILL version metadata; resolve adapter.manifest_ref through the existing verified artifact store; require adapter_id/version/direction equality with the referenced manifest. Missing refs, duplicate IDs/versions, wrong digest or mismatched identity refuse resolution. Manifest location is not inferred from a name, tag or README. One package selects one primary SOURCE or TARGET adapter; additional helpers are pinned input assets or separately qualified provider packages, not implicit adapters. An incompatible package schema is refused, never interpreted as prose fallback for execution.

The descriptor is included in the package content hash, but does not contain that hash (avoids self-reference). Adapter manifests/executable bytes may be co-distributed or separately acquired: copy and verify them through existing artifact registration before binding. A local file name is not an artifact identity. Discovery remains advisory and mutable; freeze exact descriptor, package tree, adapter manifest and consumed assets in the approval closure. Historical subjects never resolve through the current index. No new registry, service, SQL table or workflow is introduced.

[SERVICE-CONTRACTS](../01-architecture/SERVICE-CONTRACTS.md#executable-adapter-binding) owns executable binding and composed qualification. Changing package, binding, executable, operation schemas or profile requires successor qualification and applicable SOURCE/PLAN approval; no upgrade rewrites previous evidence. A skill can express/lower supported IR/TIM semantics, but cannot add an operator or an invocation category by installing assets.


Persist the canonical inventory itself as a verified artifact: its raw canonical-byte digest equals the rwt-skill-package-1 content hash. Resolve package.json by its inventory entry and raw digest. PLAN decision_refs retains inventory and descriptor refs plus consumed file closure; SOURCE retains them via package_refs in its engine execution-manifest 1.1, not a nonexistent SOURCE decision_refs field; composition source_package_content_ref/target_package_content_ref pin the inventories, while source_package_ref/target_package_ref pin the descriptors. The engine verifies the inventory-to-descriptor relationship before qualification/admission; matching only a package name/version is insufficient.

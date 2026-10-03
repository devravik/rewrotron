# User intent and exact target resolution

**Owner:** Configuration precedence, environment choices and feasibility. [target.schema.json](../../specs/project/target.schema.json) v2 is requested intent; [resolved-environment.schema.json](../../specs/project/resolved-environment.schema.json) is an immutable exact resolution. They are not interchangeable.

## Separate records

| Record | Authority / mutability |
| --- | --- |
| Detection | Immutable measured/inferred facts with provenance; assertions cannot overwrite measurements |
| Human assertions/corrections | Immutable separate claims; disagreement with measured runtime remains a finding |
| Target intent | User-owned input; successor revisions on edits |
| Recommendations | Engine/agent proposals with evidence; never mutate intent |
| Resolution | Exact runtime/services/build/dependency bytes, profile and feasibility refs |
| Decision | Immutable proposal selected for the reviewed plan; Gate 2 approval applies transitively |
| Progress | Mutable/reconstructable execution projection keyed by approved refs |
| Verification | Immutable scoped result under exact captured subjects |

## Human files and precedence

`rewrotron.yaml` at workspace root is the primary user input. Source is separate and auto_detect is bounded discovery, not approval. Optional `.rewrotron/target.yaml` is an explicitly configured split target input; it cannot implicitly compete with root target. If both supply different target values, reject ERR_CONFIG_CONFLICT until the user selects one owner. Legacy project.yaml imports are explicit, reviewed and preserve provenance; generated `.rewrotron/{detection,recommendations,dependencies,plan,approvals}.yaml` are read-only regenerable views. Editing approvals.yaml never authorizes work.

Precedence: hard engine constraints (reject incompatible choice, never replace it) → current explicit user intent/overrides → project defaults → detected suggestions → recommendations → defaults. A new explicit edit replaces requested intent and invalidates Gate 2; an old confirmation cannot force a new edit back to its old value. Do not silently apply RECOMMENDED/OPTIONAL decisions.

Intent v2 names language, framework/version, runtime version/OS/CPU, architecture, target path, preserved DB policy and deployment profile. Optional web/database/cache/queue/storage/package manager and compatibility policies become explicit in resolution when required. Standard-library targets use framework=`none`, version=null; resolution framework=null. Wildcards/ranges/latest may be intent constraints only, never exact resolution.

## Resolution before execution

Resolve exact runtime/framework/service/package-manager versions, content-addressed distributions, build and deployment profile, locks and feasibility evidence. Same version with modified bytes is a different resolution. Missing runtime/version/platform/native extension/lock/license or availability evidence remains BLOCKED/UNKNOWN. Feasibility recursively checks dependencies and used capabilities, not package names. A user choosing PHP 8.4 cannot be downgraded to 8.3; show blocker/alternatives and wait for a successor intent if needed.

Every material recommendation reports required/recommended/optional/informational/blocking class, affected scope, source evidence, consequences and alternatives. Human choice flows through `decision propose`/`decision show` and Gate 2; all action types have CLI/MCP parity. [DEPENDENCY-RESOLUTION](DEPENDENCY-RESOLUTION.md) owns frozen strategies versus progress.

Target edits re-resolve affected inputs and create a successor plan; unchanged source Gate 1/analysis/characterization remain reusable when typed validity predicates hold. Target-only edits never restart unrelated discovery. Verification under an old target remains history; it cannot be relabeled current without a fresh run. [DOMAIN-MODEL](../02-domain/DOMAIN-MODEL.md) owns certificates and conservative invalidation.

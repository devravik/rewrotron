# Confirmation protocol, baseline identity, and drift

**Status:** Normative specification. Applies to every CLI, MCP, worker, and direct service entrypoint. See [lifecycle](PROJECT-LIFECYCLE.md).

## Subjects approved by a human

| Gate | Subject | Required parent subjects | Permitted work |
| --- | --- | --- | --- |
| `GATE_1_SOURCE_DETECTION` | `configuration_hash` | Detection ID and source manifest hash | Deep static analysis, isolated observation, modeling, planning |
| `GATE_2_PLAN_CONFIRMED` | `plan_hash` | Valid Gate 1, Behavior IR, corpus and policy hashes | Target bootstrap, generation, repair, sandbox verification |
| `GATE_3_CUTOVER` | `release_hash` | Valid Gates 1/2, current complete verification and benchmark reports | Package the approved release and cutover runbook |

`user_confirmations` stores gate type, subject hash, configuration hash, actor, channel, action ID, reviewed payload artifact, overrides, and creation time. Gate records are append-only. Validity is computed by matching the current subject and parents; a status flag in YAML never grants authority. Gate 3 never grants access to production by itself.

## Exact hash construction

All digests are lowercase SHA-256 hex. Hash input is UTF-8 canonical JSON with recursively sorted object keys, no insignificant whitespace, no duplicate keys, and arrays in semantic order. Hash documents use integers, strings, booleans, and null; confidence floats and timestamps are excluded. Preserve file bytes, line endings, and Unicode; do not normalize code contents. The canonicalizer has cross-platform golden vectors.

`source_manifest_hash` hashes a sorted array of `{path, sha256, executable, symlink_target}` records. Paths are relative POSIX paths sorted by UTF-8 bytes. Inventory includes application code, configuration, lockfiles, templates, public assets, and included framework/dependency files. Exclude only `.git/`, `.rewrotron/`, the explicitly identified workspace-owned `rewrotron.yaml` when source and workspace are co-located, and explicitly listed volatile output/cache/log directories. Show the engine configuration exclusion at Gate 1; never exclude legacy application configuration on this basis. The exclusions themselves are hashed, shown at Gate 1, and must not conceal code or configuration. A changed exclusion list invalidates Gate 1.

Use `lstat`; record in-root symlink targets and the resolved file content; reject escape symlinks, devices, unreadable files, or source paths overlapping the target directory. Report duplicate aliases consistently. Read each file with before/after metadata and recheck the manifest after the scan. If files changed while reading, return `ERR_SOURCE_DRIFT` and retain no partial baseline. A Git commit ID alone is insufficient: untracked and dirty files count.

`configuration_hash` hashes `{schema_version, source_manifest_hash, source_adapter_id, source_adapter_version, confirmed_facts, overrides, exclusions, environment_identity}`. Environment identity contains runtime version, DB engine/version, schema hash, collation/timezone, sanitized endpoint identifiers, and sandbox profile digest. Raw credentials and changing business rows are excluded. A DB schema change is material; ordinary fixture row changes create new fixture/case hashes instead.

`target_spec_hash` hashes `{target_config, migration_settings, override_provenance}` after strict intent validation. The target configuration contains requested constraints, never agent-selected alternatives. Planning captures an immutable revision before synthesis; human authority comes from Gate 2 on the resulting plan.

`plan_hash` hashes `{configuration_hash, behavior_ir_hash, corpus_hash, policy_hash, target_config, decision_system, adapter_versions, ordered_units, dependency_edges, omissions, scope_exclusions, unresolved_items}`. `decision_system` captures the target specification hash and frozen specification artifact reference, exact resolved versions, dependency resolution, feasibility and reviewed recommendation decisions as immutable artifact references. These digests bind their full content transitively. New executable plans require this section; older documents without it remain historical/readable and must be rebuilt before approval. Preserve route precedence and workflow step order. Changing a target version, contract, fixture, normalization rule, scope, omission proof, or unit dependency invalidates Gate 2 and Gate 3.

`release_hash` hashes `{configuration_hash, plan_hash, target_tree_hash, dependency_lock_hashes, verification_report_hash, benchmark_report_hash, cutover_runbook_hash, rollback_runbook_hash, session_strategy_hash}`. Any target patch or release evidence change invalidates Gate 3.

## Human action and MCP trust

The actor identifier is derived from the local authenticated operating-system session or a trusted client identity, never from an agent-controlled input field. A local human can run a confirmation command with the exact reviewed subject hash. For MCP, confirmation requires a single-use approval token issued through a local human approval command. Token minting is not an MCP tool.

`rewrotron approval issue --gate source|plan|cutover --subject-hash <hash>` displays the current subject, requires an explicit terminal `approve <hash>`, and creates a 10-minute, project/gate/subject-bound random token. Noninteractive automation must supply the approval through the human confirmation CLI; it cannot mint tokens by passing `--yes`. Store only a token digest, expiry, actor and consumed time. The issuing command is intentionally a human-only operation; a client cannot establish that an agent is human merely by claiming an actor name.

MCP confirmation tools consume the token and write the confirmation in one SQLite transaction after rechecking drift and parents. This enforces a separate human channel, not cryptographic proof of human intent on a host where an agent already has unrestricted terminal access. The trust boundary is the local user account and client integration; stronger remote authorization requires a separate design.

## Request and transaction rules

1. Detection creates an immutable proposal and displays `detection_id` and `source_manifest_hash`.
2. Source confirmation references both. Overrides use a whitelist of factual keys: runtime version, framework/version, application root, DB engine/version/access, timezone/collation, and configured non-HTTP entrypoints. Unknown keys fail `ERR_INVALID_ARGUMENT`. Overrides preserve old values and provenance.
3. A Gate 1 override changes the configuration subject. The human receives the merged facts and resulting hash before approval. MCP must first call `project.prepare_source_confirmation` to obtain that merged subject; it cannot introduce new overrides while consuming an approval token.
4. Plan confirmation accepts only `plan_hash`, and cutover confirmation only `release_hash`. Changing preferences requires synthesizing and reviewing a new plan/release.
5. Within `BEGIN IMMEDIATE`, validate parents, subject freshness and token, insert confirmation/event, and change state. Commit all or none.
6. An idempotency key plus identical payload returns the existing result. A reused key with a different payload returns `ERR_IDEMPOTENCY_CONFLICT`. A consumed token is reusable only as part of the same successful idempotent request.

## Drift and invalidation

Recheck source and environment subjects before every expensive or state-changing operation, before accepting a worker result, and immediately before Gate 2/3 confirmation. Source changes return the project to `DETECTED`; old confirmations, IRs, plans, generated code, and reports remain historical and are marked stale through new events. No history is deleted.

A new normalization policy or case corpus can keep Gate 1 valid while invalidating Gates 2/3. A target-only patch keeps Gate 2 valid when it implements the confirmed plan; it invalidates release verification and Gate 3. Unsupported/unreadable source changes fail closed. A mid-run drift aborts publication and queues re-detection; source generation and observation never mutate the baseline.

## Required negative cases

Missing approval, YAML `confirmed: true`, agent-supplied actor, stale manifest, mismatched plan hash, expired/wrong-project token, concurrent confirmation, repeated token use, dry-run, override mismatch, and crash before commit must never unlock a gate. A successful repeat request must return the original confirmation ID without adding duplicate records.

## Target-only changes and resuming approved work

Changing requested runtime/framework/infrastructure or a dependency strategy creates a new target/plan subject, invalidating Gates 2/3 while leaving unchanged Gate 1 and source characterization valid. Recompute the transitive affected artifact set; retain unrelated IR, route discovery and history. Worker results captured against an old subject cannot publish as current. Reused target work needs checked prerequisites and current evidence under the successor confirmed plan. Editing generated `approvals.yaml` or `plan.yaml` cannot perform this transition. A root intent-file edit is separate from legacy source drift as defined by the manifest exclusion above. See [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md).

## Execution approval presentation

An APPROVAL_REQUIRED event/card exposes exact gate/subject, evidence and affected scope. APPROVAL_RECEIVED is recorded only in the transaction persisting the matching `user_confirmations` record; a display event cannot create approval. Stale cards fail current-subject checks. Alert receipt/dismissal, `requires_user`, progress and human pause/continue are not gate confirmations. The shared execution layer delegates approval authority to this protocol.

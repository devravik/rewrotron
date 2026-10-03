# Domain records, identity and validity

**Owner:** Record responsibilities and typed invalidation/reuse. Detailed behavior, approvals, target, execution and storage semantics belong to their topic owners. IDs are document-local within project/baseline/plan; persistent row IDs name immutable versions, never overwrite a prior baseline. All references resolve with project ownership and exact artifact hash checks.

## State separation

Detected facts, human assertions, target intent, recommendations, exact resolution, decision proposals, approval receipts, progress and verification have separate identities. Gate 2 hashes immutable decision refs, not their execution status. Target model components own implementation paths; their progress is a separate projection. Historical approval/result rows are append-only; current validity is derived against current subjects.

## Typed dependency graph

[validity-graph.schema.json](../../specs/project/validity-graph.schema.json) defines immutable directed edges from prerequisite ref to dependent ref: SOURCE_IR, SOURCE_CONTRACT, ENVIRONMENT_OBSERVATION, ADAPTER_EXTRACTION, DEPENDENCY_IMPLEMENTATION, COMPONENT_UNIT, NORMALIZATION_RESULT, KNOWLEDGE_ARTIFACT, SKILL_RESULT and TOOL_EVIDENCE. The edge predicate is SAME_CONTENT, SAME_PROFILE, SAME_CAPABILITY, SAME_RULES or SAME_FENCE, backed by witness refs. Generic artifact_links index these typed graph artifacts; they do not by themselves prove reuse. The graph is frozen with a plan; new observations/consumption produce successor graph revisions rather than editing approved graph bytes.

SAME_CONTENT compares bytes and parse prerequisites; SAME_PROFILE compares all declared environment/reset/operational dimensions; SAME_CAPABILITY requires pinned manifests/content and current control tests; SAME_RULES compares semantics/corpus/policy/relevant normalization; SAME_FENCE verifies current resource owner during publication, not future reuse. Missing edges, unknown impact or unverifiable witness conservatively invalidate affected artifact categories. Compute transitive reverse closure. Node IDs/hashes and edge references must resolve in one authorized project; global sanitized knowledge refs have explicit imported project-owned copies and consumption provenance.

## Historical versus current

Historical work remains retained. Retained bytes need not be currently valid. Characterization may be reusable under unchanged source/profile/capture inputs. Implementation may be reusable under unchanged obligations/resolution/lowering/prerequisites. Current approval requires matching subject/parents. Current verified work requires an applicable current scoped report. These five states must not be collapsed into a single complete flag.

[reuse-certificate.schema.json](../../specs/project/reuse-certificate.schema.json) binds old/new plan, artifact, graph, checked edge refs and evidence. Only the engine issues it after validation. Initial selective reuse permits characterization/implementation bytes; verification always reruns against the successor plan/target. Reuse certificates cannot remap old PASS to a new environment. A certificate is retained but validity can cease when parents change.

Pinned knowledge/tool/skill consumers are typed edges. Retraction marks current descendants stale and queues revalidation; it never deletes approved history. Status queries explain affected work and retained artifacts. Source snapshot drift and target edits have distinct propagation roots.

## Assessment findings and progression records

[PROJECT-LIFECYCLE](PROJECT-LIFECYCLE.md) owns the shared checkpoint. Assessment reports capture existing analysis records, not a competing truth model. Report completion, migration feasibility, finding validity, user selection, plan readiness and verification are separate. INPUT_ASSESSMENT_FINDING edges connect actual source/target/profile/adapter/policy/evidence prerequisites to a finding; FINDING_ASSESSMENT_REPORT connects findings to captured report projections; ASSESSMENT_SELECTION binds report to human selected work; SELECTION_PLAN connects selection/planning basis to the executable plan. Use the existing SAME_CONTENT/PROFILE/CAPABILITY/RULES predicates with exact witnesses. A successor report can reuse valid original analysis refs without a prior plan. The planning-basis artifact records per-finding predicate checks; plan-to-plan reuse certificates remain for implementation/characterization reuse across plans, never a prerequisite to assessment-only reporting.

An immutable historical report remains valid evidence of what was assessed, while its current applicability can be STALE/UNKNOWN. Current assessment views name current project sequence and per-finding validity; report as_of_sequence is captured provenance, not historical projection of arbitrary mutable rows. Re-evaluate only affected findings/downstream plan inputs, preserving target-independent source characterization. Missing typed coverage conservatively invalidates; global report staleness must not imply every constituent finding is stale.

Shared pre-plan work scopes are immutable registered artifacts under work-scope.schema; their paths resolve only within the named snapshot. They are not verifier acceptance scopes and require no approved corpus. Report/selection/basis refs must resolve that exact scope; chosen executable units/cases lower it only during planning.

## Execution and performance dependency edges

IMPLEMENTATION_QUALIFICATION links implementation/schema/profile/control-suite/result prerequisites to an exact qualification. EXECUTION_RESULT links an execution/adapter manifest or qualification to its consumed analysis/implementation/evidence/result. PERFORMANCE_RESULT links approved policy/workload/environment/method/raw evidence to benchmark/release artifacts. SAME_CONTENT/PROFILE/CAPABILITY/RULES witnesses are checked as applicable; manifests cannot be considered unchanged merely because their version names match. A successor interpreter/schema/control/policy invalidates the reverse affected closure, retaining history and target-independent source facts where proved reusable. Emit these edges when consuming inputs, not retrospectively guessed at resume. Qualification/control closures and their consumption graphs must avoid self/future-result hash cycles.

## Positive prerequisite/consumption coverage

[consumption-coverage.schema.json](../../specs/project/consumption-coverage.schema.json) records immutable coverage for a prerequisite scope: discovery method/revision, input refs and roots actually examined, required prerequisite categories, category status/evidence, unresolved candidates, excluded mechanisms and closure controls. Record consumption atomically with artifact publication; coverage describes prior inputs, not the dependent artifact's future digest, avoiding a cycle. Reuse certificates 2.1 bind artifact, graph and coverage refs; service joins coverage to the consumed artifact's publication record.

Sufficient coverage means every prerequisite category relevant to the artifact has a CLOSED engine-derived category with reviewed roots/inputs and applicable controls, or an evidenced IRRELEVANT category. A producer assertion or empty graph cannot establish this. Dynamic/unexamined/excluded relevant mechanisms mean PARTIAL/UNKNOWN and invalidate the affected category conservatively. Compare examined input set with actual runner/generator/projector consumption; an omitted edge or unregistered input fails closure. Covers contracts, implementations, shared components, dependency decisions, assessment findings, result validity and learned consumers. Verification results never become current in a successor plan without rerun. Historical observations remain retained even when coverage is insufficient for reuse.

Global detached-consumer coverage/retraction is deferred to M7. Until its typed index/reconciliation exists, global knowledge is advisory and pinned; no automatic promotion or activation.

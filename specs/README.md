# Machine-contract registry

Current execution records are explicitly versioned; the owning service selects the current required branch, never a historical fallback. Legacy v1 descriptive/artifact records remain readable historical branches; they do not satisfy current generation/approval/acceptance. IR envelope and CLI output remain v1 where unchanged. Catalog contract_version=2.1 and MCP protocol_version=2025-11-25 are separate identities. No deployed engine compatibility is claimed.

Schemas use JSON Schema 2020-12 and local logical IDs under https://rewrotron.dev/specs/. Resolve offline; this namespace does not require a hosted service. cli/commands.json and mcp/tools.json are authoritative catalogs; their output_variants explicitly cover admission versus completed reports. Shape validation is not domain truth: services check ownership, ref availability, canonical hash/order, DAG/scenario links, typed execution/capture, scope completeness and human authority.

| Group | Current responsibilities |
| --- | --- |
| project | requested intent, exact resolution, coherent snapshot, immutable decision/subject, typed validity/reuse, confirmation/current state |
| behavior | bounded typed executable semantics, guarded workflows, cases/input-only observation, lossless request |
| observation | canonical causality, mechanism attestation, operational profile, qualified effects, full reset report |
| analysis | conservative omission proof bundle |
| ir | neutral three-tier envelopes and provenance |
| migration | plan subject, immutable target model, separate progress, patch/release |
| verification | scoped report, relational normalization and declared performance claim |
| execution | v2 correlated events/current-state plus bounded event-window page |
| operations | shared result/admission, artifact content and coordinated archive manifest |
| cli/mcp | all interfaces to the same guarded engine including typed decision proposals |
| examples | historical/current positive-negative structural fixtures and approval canonical vectors |

Apply migrations 0001 then 0002; do not rewrite old DDL/history/approvals. Public sha256 refs map to project-owned artifact row IDs; protected closures stay available or in verified restorable archives. Current subject preimages never consult mutable registries. [Confirmation owner](../docs/02-domain/CONFIRMATION-PROTOCOL.md) and [testing owner](../docs/10-development/TESTING.md) define semantic/falsification obligations.

Run `python3 tools/docs/validate.py` for schema/ref/catalog/fixture/link/hash-vector/incremental-DDL checks only. Runtime/Go/driver/race/client/security/performance tests remain pending milestones.

Assessment is owned by PROJECT-LIFECYCLE. reports/assessment-report is an immutable projection of the existing assessed state; reports/assessment-view exposes current validity separately. project/assessment-selection is an explicit human progression decision, never an approval token; migration/assessment-planning-basis records changes/reuse evidence. Current executable plan envelopes are 2.1 and bind these refs through unchanged v2 PLAN subject decision_refs. Old plan envelopes remain historical-readable. Catalog contract_version is 2.1; transport and CLI envelopes retain their separate versions. No new analysis engine, SQL lifecycle enum or migration table is implied. JSON schema checks cannot establish human identity, non-mutation, continuation authority or input currentness; those require the specified service/falsification tests.

project/work-scope is the shared pre-plan boundary (snapshot, critical workflows, requirements/constraints and selected work); it does not require the executable corpus demanded by verification/acceptance-scope. All new example cases and hash vectors are synthetic structural fixtures with illustrative refs, not recorded application observations or runtime validation.

Current branch map: approval-subject 2.0; plan 2.1; adapter manifest 2.1; attestation 2.1; data-model 2.0; benchmark 2.1; execution-manifest/qualification/performance-policy 1.0. Their distinct versions do not imply that all records share one version. CONFIRMATION-PROTOCOL owns engine/schema pins and performance policy membership in the approval closure; SERVICE-CONTRACTS owns exact-profile qualification. BEHAVIOR-MODEL owns additive row/write-binding semantics and interpreter limits. No new SQL tables are required: these immutable artifacts and their typed reference edges use existing storage. MCP advertises offline local schema bundles under MCP-SPEC, not network fetch URLs.

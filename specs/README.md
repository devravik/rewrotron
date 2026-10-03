# Machine-contract registry

Current execution records are explicitly versioned v2. Legacy v1 descriptive/artifact records remain readable historical branches; they do not satisfy current generation/approval/acceptance. IR envelope and CLI output remain v1 where unchanged. Catalog contract_version=2.0 and MCP protocol_version=2025-11-25 are separate identities. No deployed engine compatibility is claimed.

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

# Milestones and evidence gates

| Milestone | Task boundary | Required evidence |
| --- | --- | --- |
| M0 Specification baseline | Documentation/schema review | Local schemas/links/registries/DDL checked; external validation register visible |
| M1 Safe foundation | RWT-001–006 | Durable workspace, bounded detection, explicit Gate 1, drift and negative refusals, valid JSON outputs, canonical execution events/state, live inspection and safe human control |
| M2 Characterization | RWT-007–013 | Complete reference IR/trace/corpus, omission counterexamples, safe reset/egress, comparator MISMATCH/UNKNOWN detection |
| M3 First slice | RWT-014–016 | Exact Gate 2, staged independent Laravel fixture, regression and crash/conflict recovery |
| M4 Reference application | RWT-017–022 | Auth/session/CRUD/effects/files, full required corpus and coverage reconciliation, benchmark evidence, Gate 3 package/runbooks |
| M5 Reference release | RWT-023–025 | Optional privacy learning/Skill, two clients, supported OS/capabilities, security/recovery/licenses/signatures |
| M6 Extensions | RWT-026–027 | New adapter intersection verified with unchanged canonical IR |

M0 does not freeze every future implementation detail or assert approval by unnamed stakeholders. Runtime milestones require executed results, not documentation status. Product claims remain bounded by tested scope and capability maturity.

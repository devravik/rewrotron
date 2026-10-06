# Behavioral units and acceptance scopes

**Owner:** Schedulable unit boundaries and completion rules. A unit is a behavioral vertical slice of operations/workflows; shared code belongs to one explicit component owner in the Target Implementation Model. Strongly connected behavior prerequisites condense into a schedulable unit. A cyclic execution DAG is a plan blocker.

## Three acceptance scopes

| Scope | Required corpus | Meaning |
| --- | --- | --- |
| UNIT | Exact approved cases mapped to this unit plus required prerequisite checks | Unit can complete; future dependent cases need not exist in target yet |
| IMPLEMENTED_REGRESSION | Applicable approved cases of all currently implemented units, including shared-component consumers | Current implemented baseline remains valid |
| PROJECT | Entire approved in-scope corpus and mandatory dimensions | Whole-project correctness prerequisite for release |

[acceptance-scope.schema.json](../../specs/verification/acceptance-scope.schema.json) identifies kind, unit IDs, sorted case IDs, corpus ref and scope hash. Hash scope as rwt-cjson-1 of that record without scope_hash. Report exact scope, plan/target/profile/policy identities and every case outcome. Absent/UNKNOWN/MISMATCH cases never count as completion. Scope selection is derived by guarded engine services, not a supplied agent convenience filter.

## Generation admission uses the current unit scope

Before staging a unit, derive its UNIT acceptance scope from the exact current Gate 2 plan, never from a caller-supplied case filter. Admission requires current source snapshot/qualification, M3 checkpoint/plan/TIM, exact target resolution and valid human approval bindings. These closure checks remain mandatory even for a small slice. Source evidence must remain source-only; target choices, comparator expectations and approvals cannot supply missing observations.

Within that closure, resolve the selected unit's contract refs, approved cases and their applicable evidence dimensions, required completed prerequisites, and exact qualified target capabilities. A contract must be executable where required; a reference or structurally valid placeholder does not suffice. Required session/auth/CSRF semantics remain owned by AUTH-AND-SESSION and M4 Unit 4. Missing, stale, contaminated, unsupported, UNKNOWN or insufficient evidence for any required obligation blocks this unit. A plan's blocking unresolved item affecting this unit, its cases/contracts/capabilities or prerequisites remains blocking; unresolved source-symbol/component dependencies must be resolved by the guarded closure check, not silently excluded.

Unrelated future/other-unit evidence does not block generation. For example, a synchronous routing/auth/session slice with sufficient applicable evidence can be ADMITTED while a separate async-completion slice remains BLOCKED/UNKNOWN. An async obligation that is an actual dependency of the first slice still blocks it. This does not establish that the reference fixture already supplies the synchronous slice's executable contracts, current session evidence or approved M3 closure.

The internal `EvaluateGenerationAdmission` predicate consumes the existing plan envelope and acceptance-scope models. Its injected guarded resolver validates artifact/currentness/approval closure and resolves sufficient evidence for the derived requirements; the predicate does not mint approvals or attestations. Results report UNIT scope (included cases, corpus and scope hash), excluded PROJECT cases, contract/capability/prerequisite requirements, retained unresolved items/limitations and ADMITTED/BLOCKED with reasons. ADMITTED means generation readiness, **not verifier PASS, unit completion or project acceptance**. Results are transient; currentness, ownership and fences must be checked again before staging/publication. This is not a new public CLI/MCP or stored artifact contract.

Complete UNIT and applicable IMPLEMENTED_REGRESSION verification still follow the rules below. PROJECT acceptance still requires the entire approved applicable corpus; excluded unit-admission cases remain project obligations and cannot be counted as PASS. M4 Unit 8 must enforce all applicable M4 composition/corpus obligations, including async where required; whole-project performance/release acceptance remains owned by M6. No scoped admission result upgrades PARTIAL DB intermediate visibility or PARTIAL async causal/completion qualification. Trace 2.1 still requires qualified horizon/due evidence for **every window it publishes**: this rule scopes which evidence is required, not the validity of any required trace/window.

## Completion and scheduling

A unit completes when its UNIT cases pass, all applicable implemented-regression cases pass, required prerequisites are current valid completed units, and no in-scope blocking UNKNOWN/MISMATCH remains. Prerequisite acceptance never depends on future dependent-unit cases. Login can complete before checkout; checkout subsequently checks login regressions. Invalidation can remove current completion without deleting historical success. Once all units are implemented, PROJECT acceptance still executes the complete approved corpus; adding or dropping cases requires successor plan approval.

A scoped exception/exclusion changes the approved scope and is visible; it never makes the excluded case PASS. Shared component ownership/resource claims and the serial baseline prevent overlapping work. Completion evidence includes target tree/model refs, scoped reports, current fences and prerequisites. Benchmark acceptance is distinct from slice correctness and release requires all applicable performance obligations.

## Shared-component maintenance

An owner change carries a pending regression obligation for every mapped consumer. Check all currently implemented consumers in the active scope now; future consumers discharge their obligation at implementation. No future case is required for owner completion. Consumption coverage must justify the mapped set; missing/incomplete consumer coverage invalidates conservatively rather than treating an empty list as sufficient. Acceptance scope identifies corpus/current target and implemented-unit revision.

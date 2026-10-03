# Behavioral units and acceptance scopes

**Owner:** Schedulable unit boundaries and completion rules. A unit is a behavioral vertical slice of operations/workflows; shared code belongs to one explicit component owner in the Target Implementation Model. Strongly connected behavior prerequisites condense into a schedulable unit. A cyclic execution DAG is a plan blocker.

## Three acceptance scopes

| Scope | Required corpus | Meaning |
| --- | --- | --- |
| UNIT | Exact approved cases mapped to this unit plus required prerequisite checks | Unit can complete; future dependent cases need not exist in target yet |
| IMPLEMENTED_REGRESSION | Applicable approved cases of all currently implemented units, including shared-component consumers | Current implemented baseline remains valid |
| PROJECT | Entire approved in-scope corpus and mandatory dimensions | Whole-project correctness prerequisite for release |

[acceptance-scope.schema.json](../../specs/verification/acceptance-scope.schema.json) identifies kind, unit IDs, sorted case IDs, corpus ref and scope hash. Hash scope as rwt-cjson-1 of that record without scope_hash. Report exact scope, plan/target/profile/policy identities and every case outcome. Absent/UNKNOWN/MISMATCH cases never count as completion. Scope selection is derived by guarded engine services, not a supplied agent convenience filter.

## Completion and scheduling

A unit completes when its UNIT cases pass, all applicable implemented-regression cases pass, required prerequisites are current valid completed units, and no in-scope blocking UNKNOWN/MISMATCH remains. Prerequisite acceptance never depends on future dependent-unit cases. Login can complete before checkout; checkout subsequently checks login regressions. Invalidation can remove current completion without deleting historical success. Once all units are implemented, PROJECT acceptance still executes the complete approved corpus; adding or dropping cases requires successor plan approval.

A scoped exception/exclusion changes the approved scope and is visible; it never makes the excluded case PASS. Shared component ownership/resource claims and the serial baseline prevent overlapping work. Completion evidence includes target tree/model refs, scoped reports, current fences and prerequisites. Benchmark acceptance is distinct from slice correctness and release requires all applicable performance obligations.

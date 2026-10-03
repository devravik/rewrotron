# Migration engine contract

**Status:** Normative. Target work requires valid Gates 1 and 2. Generation reconstructs behavior slices from canonical IR through a target adapter.

## Unit progression

`DISCOVERED → PLANNED → GENERATING → GENERATED → VERIFYING → VERIFIED → COMPLETED`. A failure creates `BLOCKED` with concrete findings and preserves earlier evidence. A repair returns the unit to GENERATED/VERIFYING and invalidates current completion evidence.

1. Characterize the entrypoints, behavior rules, required dimensions, fixtures and cases.
2. Check target adapter capability coverage and dependency readiness.
3. Include the unit and its risks in the immutable plan approved at Gate 2.
4. Claim a fenced mission and stage target files/patches with before hashes.
5. Run target build/static checks and differential cases; persist immutable reports.
6. Repair proved defects using recorded hypotheses, maximum three attempts per current context; missing evidence creates an observation/review mission.
7. Reverify the current full regression corpus and record provenance before completion.

A unit completes only with current PASS results in all required cases/dimensions, no unexplained mismatch, committed state/effect equivalence, target quality checks, and hashed artifacts. An approved intentional change is recorded as a policy exception and rerun; it is never a fourth verdict. Performance optimization requires prior correctness and fresh benchmark/regression evidence.

## Plan stability and generation safety

The plan names source/configuration, IR/corpus/policy hashes, target/runtime/dependency versions, architecture, ordered units and dependencies, omissions, exclusions and blockers. Source drift invalidates Gate 1; changes to plan subjects invalidate Gate 2/3. A target implementation patch of the same confirmed behavior keeps Gate 2 but invalidates release results.

Target publication and rollback are specified in [RECOVERY](../08-storage/RECOVERY.md); authentication continuity in [AUTH-AND-SESSION](AUTH-AND-SESSION.md). Adapters cannot mark themselves verified. Independent units can run after an unrelated blocker, while prerequisite and conflicting-file work remains blocked.

Generation cannot derive arbitrary business logic from free-text descriptions alone. Unsupported computations return `ERR_UNSUPPORTED_CAPABILITY` with exact rule/evidence references. External agents may propose implementations; the same build, gates, staged publication and differential verifier govern them.

## Target and dependency planning prerequisites

Freeze the validated user [target specification](TARGET-SPECIFICATION.md) before synthesis. Compare detected/recommended/requested/exact resolved versions and run [dependency resolution](DEPENDENCY-RESOLUTION.md) recursively using actual usage and required capabilities. Include evidence-backed recommendation decisions, API changes, adapters/native implementations, characterization cases and unresolved scope in the plan review. REQUIRED actions need acceptance or target/scope revision; BLOCKING issues prevent dependent execution; optional changes need opt-in. Recommendations never mutate the target.

New plans include `decision_system` with `target_spec_hash`, `target_spec_ref`, resolved target versions and immutable recommendation/dependency/feasibility refs. Gate 2 binds these inputs through the plan hash. Historic plans missing this section are readable but must be rebuilt before execution. Target edits invalidate the current approval and affected target evidence; preserve unrelated source work and use recorded dependency links to revalidate retained artifacts under the new plan. A feasibility success is not behavioral verification.

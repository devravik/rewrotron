# Vocabulary index

This file owns names; linked specifications own exact semantics/fields. Do not redefine them in summaries.

| Term | Canonical meaning owner |
| --- | --- |
| Observed / contract / semantic equivalence; unverified / unknown / unsupported; intentional change | [BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md) |
| Source / Application / Behavior IR; evidence / provenance | [IR-CONTRACTS](../02-domain/IR-CONTRACTS.md) |
| Target intent / resolution / recommendation / feasibility | [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md) |
| Target Implementation Model / component ownership | [TARGET-ARCHITECTURE](../04-migration/TARGET-ARCHITECTURE.md) |
| BehaviorContract / BehaviorCase / bounded scenario / actor / completion boundary | [BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md) |
| UNIT / IMPLEMENTED_REGRESSION / PROJECT acceptance | [MIGRATION-UNITS](../04-migration/MIGRATION-UNITS.md) |
| PASS / MISMATCH / UNKNOWN; normalization / identity relations | [COMPATIBILITY-SPEC](../05-verification/COMPATIBILITY-SPEC.md) |
| DEAD under boundary B / omission proof | [CODE-CLASSIFICATION](../03-analysis/CODE-CLASSIFICATION.md) |
| Immutable approval subject / configuration_hash / plan_hash / release_hash | [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md) |
| Mechanism-derived completeness / attestation / causality | [RUNTIME-OBSERVATION](../03-analysis/RUNTIME-OBSERVATION.md) |
| Decision / progress / typed validity / reuse certificate | [DOMAIN-MODEL](../02-domain/DOMAIN-MODEL.md) |
| Mission / work fingerprint / lease / fence / resource claim | [MISSION-PROTOCOL](../06-agent/MISSION-PROTOCOL.md) |
| Current execution state N / event window M / heartbeat / progress / pause | [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) |
| Foreground owner / supervised operation / journal / backup closure | [RECOVERY](../08-storage/RECOVERY.md) |
| Pattern / skill / tool revision / promotion / consumer / retraction | [LEARNING-SYSTEM](../07-learning/LEARNING-SYSTEM.md) |

Validation maturity is defined by [AGENTS](../../AGENTS.md): DESIGNED, SPECIFIED, STRUCTURALLY_VALIDATED, IMPLEMENTED, UNIT_TESTED, RUNTIME_VALIDATED, INTEGRATION_VALIDATED, RELEASE_VALIDATED. Adapter distribution tier is separate from maturity. All named client integrations remain UNVERIFIED until actual tests.

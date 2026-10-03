# Terminology

The single vocabulary owner is [GLOSSARY](../00-vision/GLOSSARY.md). Use `BehaviorCase`, `BehaviorContract`, `Entrypoint`, `MigrationUnit`, `ObservationTrace`, and `VerificationRun` consistently. A route is an HTTP dispatch pattern; an endpoint is a concrete address; an operation is behavior reached by an entrypoint; a workflow is a sequence of operations.

`PASS`, `MISMATCH`, and `UNKNOWN` are verifier verdicts. `EXPECTED_CHANGE`, `DEFECT`, and `UNKNOWN` are mismatch classifications. A classification does not mutate an old verdict. `configuration_hash`, `plan_hash`, and `release_hash` identify different gate subjects and must never be used interchangeably.

A specification being accepted does not mean its runtime is implemented. An adapter's implementation status and its maturity are separate facts. Planned support cannot be described as tested or stable.

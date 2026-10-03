# Golden master and baseline revisions

A golden master is a reproducible source observation corpus for a specific source configuration, contract revision, fixture and comparison policy. It is evidence for those scenarios.

Run each source case at least three times from a fully reset fixture. Compare all required dimensions across repeats. Stable values become exact assertions; variable fields become candidates requiring scoped validators and relationships. Changing a value does not prove that it is semantically irrelevant. Unexplained variance creates UNKNOWN and an observation/review mission.

Lock a master only when fixture restoration, source health, instrumentation completeness and required security cases are demonstrated. Save the original redacted trace references, their content hashes, canonical assertions, approved normalization rules, creation evidence and environment identity in an immutable artifact. SQLite indexes the corpus hash and active revision; exported `.rewrotron/contracts/` files are views of these records.

A changed source, fixture, rule or scope produces a new master revision and invalidates dependent verification/plan subjects. Never overwrite the old master to make a target patch pass. Re-baselining requires an explicit reason, evidence and renewed affected gate approval.

# Knowledge records and privacy

A project lesson records `lesson_id`, project/unit/experiment IDs, observed finding, proposed remediation, source/target/corpus/policy hashes, immutable evidence references, applicability, limitations and outcome. A single success is evidence for that case, not a universal confidence score.

A reusable pattern additionally records pattern ID/version, supported capability predicates, sanitized examples, positive/negative fixture refs, evaluation outcomes, originating evidence count, approval actor/action and deprecation history. Store no customer identifiers, private URLs, credential hashes or raw traces in global records.

Framework examples must be evidence-based. CI3 session formats depend on the configured driver; do not assume encrypted session cookies or a particular legacy cipher. Session bridges must satisfy [AUTH-AND-SESSION](../04-migration/AUTH-AND-SESSION.md).

Project lessons are local-first and can exist without global storage. Global pattern schema/migration is developed in RWT-023 with privacy and regression fixtures before promotion is exposed. Pending global functionality cannot block a baseline migration.

## Dependency resolution experience

Resolution lessons include exact source/target versions, used capability predicates, selected strategy/adapter/native implementation, transitive constraints, characterization and application corpus/policy/environment subjects, and successful/failed alternatives. Report evidence counts and failures rather than inventing a numerical confidence from small samples. Known mappings remain scoped recommendations; they cannot override a user target, satisfy Gate 2 or bypass current differential verification. Global promotion follows the existing opt-in privacy and multi-context validation policy.

# Experiment identity and repair history

An experiment context is canonical JSON containing source configuration, plan/corpus/policy hashes, case/workflow and fixture hashes, target before-tree hash, adapter/tool/dependency versions, hypothesis/patch hash, sandbox identity, and skill/prompt versions when an external agent was used. Its SHA-256 digest is the fingerprint; ambiguous string concatenation is prohibited.

A failed identical context cannot be retried as if it were a new technique. Record every attempt, failure and rollback; changed inputs create a new fingerprint. Outcome and evidence are immutable. Cache hits return historical subjects and must be revalidated before counting as current success.

Patch rollback follows [RECOVERY](../08-storage/RECOVERY.md): restore only owned files with matching current after hashes. There is no fixed rollback latency guarantee. Preserve conflict evidence and human edits. External effects are confined to resettable sinks; Git commits cannot reverse a payment or production DB write.

Changed fingerprints preserve history but do not reset the three-attempt incident budget without a reviewed cause/context change. Pin exact content refs for consumed tool/skill/knowledge revisions; consumer edges support transitive retraction. LEARNING-SYSTEM owns independent evidence/quality requirements.

# Differential replay and scoped acceptance

**Owner:** Replay protocol and acceptance aggregation. [COMPATIBILITY-SPEC](COMPATIBILITY-SPEC.md) owns comparison semantics; [MIGRATION-UNITS](../04-migration/MIGRATION-UNITS.md) owns the three scopes.

## Independent environments

Source and target never share writable verification databases, Redis/session namespaces, files or effect endpoints. “Same MySQL initially” means preserved production schema/behavior, not one test database. Each side starts from independently restored equivalent fixture state and a recorded reset report. [DATABASE-COMPARISON](DATABASE-COMPARISON.md) defines multi-connection reset. Each workflow preserves its side's state across steps; each fresh scenario restores all dimensions.

Pin source snapshot, profiles, exact runtime/build, contract/corpus/policy, probes/comparators/runners and target tree. Raw HTTP targets/ordered headers/body refs preserve empty/repeated/query distinctions. Core parses large business numbers exactly; machine clients send immutable raw bytes refs or typed strings, not rounded JSON values. Match logical identities relationally across DB/response/links/effects and actor partitions.

## Protocol

Verify readiness/fences/gates, reset both sides, check isolation controls and capture controls, execute bounded steps/guards/bindings, await declared async completion, collect attestations and state/effects, compare under approved policies, publish immutable per-case and scoped reports transactionally. Every trace uses the canonical causality envelope. Limit/truncation/missing mechanisms remain incomplete. SQL rollback alone is not a reset certificate.

Known disagreement produces MISMATCH; insufficient/unavailable required evidence produces UNKNOWN; PASS requires all required assertions and dimensions. Intentional changes are approved comparison-policy obligations with explicit difference reporting, not a fourth verdict or unknown waiver. Run reports contain acceptance_scope/corpus/plan/target/profile identities. UNIT excludes unimplemented future dependents; implemented regressions check current affected components; PROJECT requires the full approved in-scope corpus. Aggregation cannot count duplicate case rows, old target results or excluded cases as current PASS.

## Oracle independence

Before acceptance, test negative targets missing a DB write, switching owner identity, duplicating effect, returning a wrong redirect, mishandling token lifecycle and suppressing a capture dimension. Approved normalizers must retain these failures. Expected behavior is reviewed against source observations, supplied critical cases and independent holdouts; it is not established solely by the generator's own tests. When observation cannot safely reach production-only/opaque/irreversible behavior, report a blocker or explicit unsupported/excluded scope.

Current v2 reports populate every dimension with applicable/verdict/reason/evidence refs. Not-applicable has verdict=null and contract justification; applicable dimensions have only the three verdicts. PASS cannot contain an applicable UNKNOWN/MISMATCH or zero cases. Service checks additionally require the exact corpus scope and cannot accept an agent-selected not-applicable flag. The legacy v1 report dimension shape had an impossible `$ref` requirement; it is historical non-executable data, never a current report contract.

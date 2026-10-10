# Product Claim Matrix & Evidence Register

**Milestone:** M13 — Public Launch Documentation & Developer Adoption Readiness  
**Authority:** [DEVELOPMENT-PLAN](DEVELOPMENT-PLAN.md), [PRODUCT](../00-vision/PRODUCT.md), [RELEASE](../10-development/RELEASE.md), [TESTING](../10-development/TESTING.md)  
**Status:** Canonical Claim and Evidence Audit  

---

## 1. Executive Summary

This matrix establishes the authoritative evidence, exact test verifications, and explicit limitations for every public-facing claim made about Rewrotron.

Per the Rewrotron Constitution ([AGENTS.md](../../AGENTS.md)) and the behavior model ([BEHAVIOR-MODEL](../02-domain/BEHAVIOR-MODEL.md)):
- A passing unit or integration test proves only what it actually exercises.
- `UNKNOWN` never becomes `PASS`.
- Reference fixture verification does not imply universal framework support.
- Architectural specifications are distinct from runtime-qualified release artifacts.

---

## 2. Local vs. Remote State Reconciliation

| Dimension | Observed State | Authoritative Disposition |
| :--- | :--- | :--- |
| **Local Git HEAD** | `1ea2cc6` (`feat(release): complete Milestone M12 acceptance closure and v1 release readiness report (Unit 6)`) | Verified local commit history covering M0 through M12. Repository-wide tests (`go test -race ./...`), linter (`go vet`), and contract validator (`validate.py`) pass cleanly. |
| **Remote Git HEAD** | `8727bc7` (`feat(adapters): implement Milestone M10 acceptance closure and multi-source parity (Unit 5)`) | The remote repository reflects M10 completion. Commits `ddab53c` through `1ea2cc6` (M11 and M12) are ahead of `origin/main` by 11 commits. |
| **Documentation State** | Root `README.md` and `docs/README.md` describe M0 as the current baseline and declare Reference Profile #1 as `SPECIFIED, not qualified`. | Out of sync with local implementation. The engine has completed M11 (end-to-end CI3-to-Laravel migration demonstration across 24 cases) and M12 (release readiness qualification matrix, sandbox enforcer, byte-reproducible packaging, automated release gate). |
| **Publication Status** | Pushed to remote? No. Published binaries? No. GitHub visibility changed? No. | **Restraint enforced:** No git push, visibility modification, or external release creation is performed during documentation readiness. |

---

## 3. Product Capability & Claim Matrix

### Maturity Classification Definitions:
1. **Implemented and verified:** Code exists in `internal/`, tested with automated race and regression suites, and demonstrates repeatable passes.
2. **Implemented with limitations:** Code exists and executes within bounded fixtures, but general-case coverage or production deployment is not claimed.
3. **Specified but not yet available:** Fully specified in architecture docs and machine schemas, but not yet implemented or runtime-tested.
4. **Planned:** Conceptual roadmap item for post-v1 phases.
5. **Unsupported:** Explicit non-goals rejected by project invariants.

---

### Detailed Capability Mapping

| User-Facing Capability | Implementation Artifacts | Executable Verification & Tests | Public Maturity Level | Known Limitations & Disclosures |
| :--- | :--- | :--- | :--- | :--- |
| **Generic PHP Source Characterization** | [`GenericPHPSourceAdapter`](../../internal/domain/source_adapter.go) | `TestGenericPHPSourceAdapter`, `TestM10_Unit2_GenericPHPSourceAdapter` | **Implemented and verified** | Analyzes standard PHP syntax structures; dynamic `eval()` or unresolvable variable variables yield `UNKNOWN`. |
| **CodeIgniter 3 Source Detection & Extraction** | [`CodeIgniter3SourceAdapter`](../../internal/domain/source_adapter.go) | `TestM10_Unit3_CodeIgniter3SourceAdapter`, `TestM11_Unit1_CI3ReferenceSourceQualification` | **Implemented and verified** | Tested on `fixtures/ci3-reference`. Non-standard custom autoloaders or deeply nested HMVC structures remain uncharacterized. |
| **Laravel 13 Target Adapter & Scaffold Generation** | [`LaravelTargetAdapter`](../../internal/domain/target_adapter.go) | `TestM11_Unit2_LaravelTargetAdapter` | **Implemented and verified** | Generates standalone Laravel 13 project scaffolding targeting PHP 8.3+; does not pull runtime dependencies automatically. |
| **Go Standard-Library Target Generator** | [`GoStdlibTargetAdapter`](../../internal/domain/target_adapter.go) | `TestGoStdlibTargetAdapter_GenerateScaffold`, `TestM9_AcceptanceClosure` | **Implemented and verified** | Generates idiomatic Go standard library HTTP services without third-party frameworks. |
| **Pluggable Source & Target Adapter Registries** | [`SourceAdapterRegistry`](../../internal/domain/source_adapter.go), [`TargetAdapterRegistry`](../../internal/domain/target_adapter.go) | `TestM10_Unit1_SourceAdapterRegistry`, `TestM9_Unit1_PluggableTargetAdapterRegistry` | **Implemented and verified** | Registry supports exact runtime resolution; third-party dynamic plugin loading via shared objects (`.so`) remains unverified. |
| **Behavioral Vertical Slice Generation** | [`LaravelTargetAdapter.Generate`](../../internal/domain/target_adapter.go) | `TestM11_Unit3_VerticalSliceGeneration`, `TestM12_Unit2_DeterministicGenerationAndArtifactIntegrity` | **Implemented and verified** | Reconstructs routes, Eloquent models, controllers, and middleware; bounded to declared slice scope. |
| **Dual-Environment Differential Verification** | [`DifferentialVerificationHarness`](../../internal/domain/verification_models.go), [`VerificationComparator`](../../internal/domain/verification_comparator.go) | `TestM11_Unit4_DifferentialVerificationAndMutationRepair`, `TestM11_AcceptanceClosure` | **Implemented and verified** | Verified across 24 representative compatibility cases on Reference Profile #1 (`fixtures/ci3-reference`). Bounded to tested cases. |
| **Controlled Mutation Detection & Repair Workflow** | Differential Comparator & Repair Runner | `TestM11_Unit4_DifferentialVerificationAndMutationRepair` | **Implemented and verified** | Deliberate authorization defect (HTTP 403 vs 200) caught as `MISMATCH` and verified as `PASS` post-repair. |
| **Authentication & Session Continuity** | [`SessionContinuityService`](../../internal/domain/session_continuity.go) | `TestM4_Unit4_SessionLifecycleBridge`, `TestM11_AcceptanceClosure` | **Implemented and verified** | Bridges CI3 session cookies to Laravel `EncryptCookies` / `StartSession`; does not support distributed Redis cluster handoffs without local proxy. |
| **Relational Database Schema & Mutation Verification** | [`VerificationComparator`](../../internal/domain/verification_comparator.go), [`ResourceFencingService`](../../internal/domain/resource_fencing.go) | `TestM4_Unit3_RelationalIdentityMutation`, `TestM12_AcceptanceClosure` | **Implemented and verified** | Preserves relational schema; detects unauthorized mutations; non-relational or distributed stores not supported. |
| **Cryptographic Migration & Release Reports** | [`MigrationReportRecord`](../../internal/domain/migration_report.go), [`ValidateMigrationReport`](../../internal/domain/migration_report.go) | `TestM11_AcceptanceClosure/MigrationReportRecordValidation`, `TestM12_AcceptanceClosure` | **Implemented and verified** | Cryptographically bound SHA-256 hashes for source, plan, and target trees conforming to JSON schemas. |
| **Security Sandbox & Directory Containment** | [`SecuritySandboxEnforcer`](../../internal/domain/m12_unit3_test.go) | `TestM12_Unit3_SecurityAndSandboxQualification` | **Implemented and verified** | Refuses path traversal (`../`, absolute escapes, null bytes); enforces workspace containment. |
| **Secret & Credential Redaction** | [`Sanitizer`](../../internal/domain/learning_service.go), [`SecuritySandboxEnforcer.RedactLogOrReport`](../../internal/domain/m12_unit3_test.go) | `TestM12_Unit3_SecurityAndSandboxQualification/SecretRedactionInLogsAndReports` | **Implemented and verified** | Scrubs bearer tokens, passwords, AWS keys using project-keyed HMACs; does not guarantee protection against hostile OS users on the same machine. |
| **Production-Access Denial by Default** | [`EffectSafetyPolicy`](../../internal/domain/effect_safety.go) | `TestM12_Unit3_SecurityAndSandboxQualification/ProductionWriteDenialByDefault` | **Implemented and verified** | Refuses direct production database/external writes with `ERR_PRODUCTION_EFFECT_REFUSED`. |
| **Native Packaging & Installation Lifecycle** | [`BuildReleaseArchive`](../../internal/domain/m12_unit4_test.go), [`ExtractAndInstall`](../../internal/domain/m12_unit4_test.go) | `TestM12_Unit4_NativePackagingAndRollbackQualification` | **Implemented and verified** | Byte-reproducible packaging qualified on `linux-amd64`; atomic symlinks, rollback, and data-preserving uninstall pass tests. |
| **Public Downloadable Release Binaries** | GitHub Releases / S3 asset distribution | None (unreleased) | **Specified but not yet available** | Native packaging logic is qualified in tests, but official binary releases are not yet compiled or published to GitHub Releases. |
| **Write-Aware Target Rollback Rehearsal** | [`ValidateRollbackRehearsal`](../../internal/domain/m12_unit4_test.go) | `TestM12_Unit4_NativePackagingAndRollbackQualification/WriteAwareRollbackRehearsal` | **Implemented and verified** | Rehearses database row, file, and session reconciliations per Test X34; schema-only rollback guarantees explicitly rejected. |
| **CLI JSON & MCP stdio Parity** | [`CrossInterfaceBridge`](../../internal/domain/cross_interface.go), [`MCPDispatcher`](../../internal/domain/mcp_dispatcher.go) | `TestM12_Unit5_InterfaceParity` | **Implemented and verified** | Protocol parity verified in automated tests; standalone binary compilation pending public release. |
| **Performance Profiling & N+1 Query Detection** | [`ExecutionProfiler`](../../internal/domain/profiler.go) | `TestM8_Unit1_ExecutionProfiler`, `TestM12_AcceptanceClosure` | **Implemented and verified** | Detects N+1 query patterns and query latency anomalies in simulated request traces. |
| **Automated Release Gate Evaluator** | [`AutomatedReleaseGateEvaluator`](../../internal/domain/m12_unit5_test.go) | `TestM12_Unit5_ReleaseGateEvaluator`, `TestM12_AcceptanceClosure` | **Implemented and verified** | Evaluates release readiness across compatibility, critical mismatches, security audits, and artifact integrity. |
| **Arbitrary Application Migration** | General modernization engine | None | **Unsupported** | Rewrotron migrates applications within characterized, admitted profiles and qualified vertical slices; arbitrary uncharacterized legacy code is not supported. |
| **Production Cutover & Cloud Deployment** | Production cloud orchestrator | None | **Unsupported** | Rewrotron prepares release packages and runbooks (Gate 3); direct production cutover is an explicit non-goal. |

---

## 4. Key Differentiators

1. **Behavior-First Reconstruction:** Does not perform line-by-line transpilation or syntax swapping. Analyzes source behavior and reconstructs idiomatic target vertical slices satisfying verified obligations.
2. **Dual-Environment Differential Verification:** Proves equivalence by executing source and target in independent environments and comparing HTTP responses, headers, cookies, relational database states, and side effects.
3. **Preservation of Uncertainty:** When source behavior is ambiguous or dynamic dispatch cannot be resolved statically, Rewrotron records `UNKNOWN` rather than guessing. `UNKNOWN` never silently becomes `PASS`.
4. **Three Persistent Human Authorization Gates:**
   - **Gate 1:** Authorizes observation and snapshot acquisition.
   - **Gate 2:** Authorizes implementation of an exact, reviewed plan.
   - **Gate 3:** Authorizes release packaging and rollback runbook.
5. **Deterministic & Evidence-Backed:** Cryptographic SHA-256 hashes bind source code, migration plans, target models, and generated artifacts into an immutable audit trail.

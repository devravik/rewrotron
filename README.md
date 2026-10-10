# Rewrotron

**Rewrotron is an open-source modernization engine that reconstructs legacy application behavior in modern stacks using verifiable observation evidence and differential contracts.**

Rather than performing blind source-to-source syntax transpilation or unpredictable LLM file rewriting, Rewrotron models application obligations across bounded vertical slices, generates idiomatic target code, and proves behavioral compatibility through dual-environment execution witnesses.

---

## What Problem Does Rewrotron Solve?

Legacy applications hide critical business behavior across controllers, global state, dynamic database queries, session lifecycles, background jobs, and framework-specific quirks. Replacing these applications presents a classic dilemma:
- **Manual rewrites** take months or years, lose subtle edge-case behaviors, and introduce regressions.
- **Automated file transpilers** carry legacy anti-patterns forward into modern languages without leveraging target idioms.
- **Direct LLM code generation** suffers from hallucinations, incomplete context windows, and inability to prove runtime equivalence.

Rewrotron solves this by organizing reconstruction around **behavioral vertical slices**—complete functional workflows (such as user authentication, stateful checkout, or authorization gating). It verifies each slice in an isolated target environment against the legacy source before declaring success.

---

## Current Project Maturity & Verified Evidence

| Area | Current Status | Evidence & Qualification |
| :--- | :--- | :--- |
| **Engine Foundation & Core IR** | **Implemented & Verified** | 3-tier IR (Source → Application → Behavior) and Target Implementation Model 2.1 qualified across M0–M4. |
| **Source Adapters** | **Implemented & Verified** | Generic PHP syntax characterization and CodeIgniter 3.x adapter qualified on [`fixtures/ci3-reference`](fixtures/ci3-reference) in M10. |
| **Target Adapters** | **Implemented & Verified** | Idiomatic Laravel 13 target adapter (PHP 8.3+, Eloquent, middleware) and Go standard-library HTTP generator qualified in M9 & M11. |
| **End-to-End Demonstration** | **Demonstrated (24 Cases)** | Dual-environment differential execution verified across 24 compatibility cases in Reference Profile #1 (CodeIgniter 3 → Laravel 13). |
| **Defect Detection & Repair** | **Demonstrated & Verified** | Deliberate authorization defect (HTTP 403 vs 200) detected as `MISMATCH`, diagnosed, repaired, and re-verified as `PASS` (M11 Unit 4). |
| **Security & Sandbox Isolation** | **Implemented & Verified** | Strict workspace containment, traversal refusal (`../`, null bytes), HMAC secret scrubbing, and production write denial (`ERR_PRODUCTION_EFFECT_REFUSED`). |
| **Packaging & Rollback Rehearsal** | **Qualified in Test Suites** | Byte-reproducible packaging (`linux-amd64`), atomic symlinks, and write-aware rollback rehearsal (X34) pass automated suites. |
| **Downloadable Release Binaries** | **Pending Publication** | Local packaging is qualified, but precompiled binaries on GitHub Releases are not yet published. Evaluation currently runs from source. |

See the authoritative [Product Claim Matrix](docs/11-roadmap/PRODUCT-CLAIM-MATRIX.md) for full mapping of capabilities, test verifications, and known limitations.

---

## How Rewrotron Works

```text
  Legacy Application (e.g. CI3 Reference)
                   │
                   ▼
  1. Capture & Extract Source IR
     ├─ Route definitions, controllers, database models, session hooks
     └─ Preserve uncertainty (unresolved dynamic paths remain UNKNOWN)
                   │
                   ▼
  2. Gate 1: Human Confirmation of Coherent Snapshot
                   │
                   ▼
  3. Plan Bounded Vertical Slices & Synthesize Target Model (TIM 2.1)
                   │
                   ▼
  4. Gate 2: Explicit Plan Approval
                   │
                   ▼
  5. Generate Idiomatic Target Code (e.g. Laravel 13)
     ├─ Routes (`routes/web.php`), Eloquent models, controllers, middleware
     └─ Deterministic, byte-reproducible generation with cryptographic digests
                   │
                   ▼
  6. Differential Verification Harness
     ├─ Source Witness Execution ──┐
     │                             ├─► Comparator (HTTP Status, Headers, Body, Cookies, DB State)
     ├─ Target Witness Execution ──┘
     │
     ├──► PASS: Verified Equivalent
     └──► MISMATCH: Defect Diagnosed ──► Repair Workflow ──► Re-Verify
                   │
                   ▼
  7. Gate 3: Release Package & Write-Aware Rollback Runbook
```

---

## Why Rewrotron Is Different

1. **Behavioral Vertical Slices, Not Files:** Migrates complete functional slices across routing, models, controllers, and state rather than translating file-by-file.
2. **Dual-Environment Differential Verification:** Proves equivalence by executing source and target in independent environments and comparing HTTP responses, headers, session cookies, database mutations, and side effects.
3. **Preservation of Uncertainty:** When legacy code contains ambiguous or unresolvable dynamic dispatch, Rewrotron records `UNKNOWN` rather than hallucinating an implementation. **`UNKNOWN` never silently becomes `PASS`.**
4. **Three Persistent Human Authorization Gates:**
   - **Gate 1:** Authorizes observation and snapshot acquisition.
   - **Gate 2:** Authorizes execution of an exact, reviewed migration plan.
   - **Gate 3:** Authorizes release packaging and rollback runbook.
5. **Cryptographic Provenance:** SHA-256 hashes bind source code, migration plans, target models, and generated artifacts into an immutable audit trail conforming to machine schemas.

---

## Quickstart: Validate & Explore Today

You can inspect, build, and run the verified demonstration directly from the repository.

### Prerequisites
- **Go:** 1.25.0+ (pinned in [`tools/toolchain.json`](tools/toolchain.json))
- **Python:** 3.10+ (with `jsonschema>=4.10,<5` for contract validation)

### 1. Clone & Validate Repository Contracts
```bash
git clone https://github.com/devravik/rewrotron.git
cd rewrotron

# Validate all schemas, fixtures, and documentation links
python3 -m pip install 'jsonschema>=4.10,<5'
python3 tools/docs/validate.py
```

### 2. Run the End-to-End CI3-to-Laravel Differential Verification Demonstration
Execute the 24-case differential verification suite and controlled mutation detection/repair workflow:
```bash
# Run the differential verification and mutation repair test
go test -race -v ./internal/domain -run "TestM11_Unit4_DifferentialVerificationAndMutationRepair"
```
*Expected result:* The test executes 24 compatibility cases, injects a deliberate behavioral defect (authorization bypass), verifies that the harness detects it as `MISMATCH`, applies a diagnosed repair, and re-verifies that the target returns `PASS`.

### 3. Run the Release Qualification & Automated Gate Evaluator
```bash
# Run the Milestone M12 acceptance closure test
go test -race -v ./internal/domain -run "TestM12_AcceptanceClosure"
```
*Expected result:* Evaluates all 9 release qualification requirements, verifies deterministic generation, validates security sandbox path containment and secret redaction, and evaluates the automated release gate to a clean `PASS`.

### 4. Run Full Test Suite with Race Detection
```bash
go test -race -count=1 ./...
go vet ./...
```

---

## Documentation & Learning Paths

| If you want to... | Read |
| :--- | :--- |
| Understand verified capabilities and known limitations | [Product Claim Matrix](docs/11-roadmap/PRODUCT-CLAIM-MATRIX.md) |
| Review the project principles and authority structure | [AGENTS.md](AGENTS.md) |
| Explore the system architecture and component boundaries | [System Design](docs/01-architecture/SYSTEM-DESIGN.md) |
| Follow the human confirmation protocol and lifecycle gates | [Confirmation Protocol](docs/02-domain/CONFIRMATION-PROTOCOL.md) & [Project Lifecycle](docs/02-domain/PROJECT-LIFECYCLE.md) |
| Read the CodeIgniter 3 source adapter specification | [CI3 Adapter](docs/04-migration/CI3-ADAPTER.md) |
| Understand differential testing and comparator policies | [Differential Testing](docs/05-verification/DIFFERENTIAL-TESTING.md) |
| Inspect security boundaries, sandboxing, and secret scrubbing | [Security Policy](docs/09-security/SECURITY.md) & [Sandbox](docs/09-security/SANDBOX.md) |
| Review the development plan and milestone progress | [Development Plan](docs/11-roadmap/DEVELOPMENT-PLAN.md) |
| Learn how to contribute to Rewrotron | [Contributing Guidelines](docs/10-development/CONTRIBUTING.md) |

---

## Community & Contributing

Contributions are welcome! Please review our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before submitting pull requests.

Rewrotron uses a strict unit-based development discipline managed via [`.agents/skills/rewrotron-implementation/SKILL.md`](.agents/skills/rewrotron-implementation/SKILL.md). All pull requests must include unit test evidence, pass race detection (`go test -race ./...`), and satisfy `python3 tools/docs/validate.py`.

---

## Security

To report a security vulnerability or sandbox escape, please follow the instructions in [Security Policy](SECURITY.md). Do not file public GitHub issues for security vulnerabilities.

---

## License

[Apache-2.0](LICENSE) © 2026 Dev Ravi & Rewrotron Contributors.

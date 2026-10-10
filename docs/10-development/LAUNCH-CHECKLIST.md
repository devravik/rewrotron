# Public Launch Checklist & Repository Publication Readiness

**Milestone:** M13 — Public Launch Documentation & Developer Adoption Readiness  
**Authority:** [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md), [RELEASE](RELEASE.md), [INSTALLATION](INSTALLATION.md), [PRODUCT-CLAIM-MATRIX](../11-roadmap/PRODUCT-CLAIM-MATRIX.md)  
**Status:** Canonical Publication Action Register  

---

## 1. Executive Summary

This checklist establishes the factual, actionable requirements for publicly launching the `devravik/rewrotron` repository on GitHub.

Per the mandatory execution protocol:
- **No external publication actions** (such as `git push`, changing repository visibility, creating tags, or publishing releases) are performed automatically by coding agents.
- All technical and documentation prerequisites must be satisfied and verified locally before operator authorization.

---

## 2. Technical & Repository Readiness Audit

| Verification Item | Requirement | Verification Method | Status |
| :--- | :--- | :--- | :--- |
| **Commit History Reconciliation** | Local `main` contains complete M0–M13 milestones; ahead of `origin/main` by documented commits. | `git log --oneline` audit against remote `8727bc7` | :white_check_mark: **Verified** |
| **Clean Working Tree** | No uncommitted changes, untracked artifacts, or dangling files. | `git status -sb` | :white_check_mark: **Verified** |
| **Contract & Schema Validation** | All 154 machine schemas, 387 fixtures, SQL migrations, and local links pass validator. | `python3 tools/docs/validate.py` | :white_check_mark: **Verified (PASS)** |
| **Repository-Wide Race Tests** | Full Go test suite passes cleanly with race detector enabled across all packages. | `go test -race -count=1 ./...` | :white_check_mark: **Verified (PASS)** |
| **Static Analysis / Lint** | Zero compile or vet warnings across all packages. | `go vet ./...` | :white_check_mark: **Verified (PASS)** |
| **Differential Verification Proof** | 24 compatibility cases verified; mutation detection and repair proven on Reference Profile #1. | `TestM11_Unit4_DifferentialVerificationAndMutationRepair` | :white_check_mark: **Verified (PASS)** |
| **Automated Release Gate** | Unified gate evaluator evaluates release qualification matrix, security, and packaging to `PASS`. | `TestM12_AcceptanceClosure` | :white_check_mark: **Verified (PASS)** |
| **Credential & Secret Protection** | Zero raw credentials, bearer tokens, AWS keys, or private customer data in repository code, logs, or fixtures. | `TestM12_Unit3_SecurityAndSandboxQualification` & repository audit | :white_check_mark: **Verified (PASS)** |

---

## 3. Documentation & Adoption Parity Audit

| Deliverable | Location | Owning Standard | Status |
| :--- | :--- | :--- | :--- |
| **Root Landing Page** | [`README.md`](../../README.md) | Modernized product narrative, visual architecture flow, verified maturity matrix, source quickstart | :white_check_mark: **Complete** |
| **Product Claim Matrix** | [`docs/11-roadmap/PRODUCT-CLAIM-MATRIX.md`](../11-roadmap/PRODUCT-CLAIM-MATRIX.md) | Authoritative mapping of user-facing claims to tests, implementation files, and limitations | :white_check_mark: **Complete** |
| **5-Minute Quickstart Guide** | [`docs/10-development/QUICKSTART.md`](QUICKSTART.md) | Step-by-step developer evaluation instructions tested directly from clean checkout | :white_check_mark: **Complete** |
| **CI3-to-Laravel Walkthrough** | [`docs/04-migration/CI3-LARAVEL-WALKTHROUGH.md`](../04-migration/CI3-LARAVEL-WALKTHROUGH.md) | Technical deep-dive into Reference Profile #1, 24 cases, and mutation repair | :white_check_mark: **Complete** |
| **Code of Conduct** | [`CODE_OF_CONDUCT.md`](../../CODE_OF_CONDUCT.md) | Contributor Covenant v2.1 with reporting email `security@rewrotron.dev` | :white_check_mark: **Complete** |
| **Security Policy** | [`SECURITY.md`](../../SECURITY.md) | Supported versions, mediated trust boundaries, and private vulnerability disclosure | :white_check_mark: **Complete** |
| **Contributor Entrypoint** | [`CONTRIBUTING.md`](../../CONTRIBUTING.md) | Root guidelines linking to unit-based development skill and invariants | :white_check_mark: **Complete** |
| **GitHub Issue Templates** | [`.github/ISSUE_TEMPLATE/`](../../.github/ISSUE_TEMPLATE/) | Bug report and feature request templates enforcing invariant checklists | :white_check_mark: **Complete** |
| **GitHub Pull Request Template** | [`.github/PULL_REQUEST_TEMPLATE.md`](../../.github/PULL_REQUEST_TEMPLATE.md) | PR template enforcing unit-based discipline, race checks, and schema validation | :white_check_mark: **Complete** |
| **License** | [`LICENSE`](../../LICENSE) | Authoritative Apache-2.0 license file present | :white_check_mark: **Complete** |

---

## 4. GitHub Repository Configuration Checklist (Operator Actions)

The following settings are recommended for the human repository administrator when publishing the repository publicly:

### A. Repository Details
- **Description:** `Behavior-first legacy application modernization engine with differential verification contracts.`
- **Website / Homepage:** `https://github.com/devravik/rewrotron`
- **Topics / Tags:**
  - `legacy-modernization`
  - `codeigniter`
  - `laravel`
  - `php-modernization`
  - `differential-testing`
  - `golang`
  - `software-architecture`
  - `behavior-driven-development`

### B. Security & Community Settings
- **Private Vulnerability Reporting:** Enable on `Settings > Code security and analysis > Private vulnerability reporting`.
- **Branch Protection:**
  - Protect `main` branch.
  - Require status checks to pass before merging (`go test -race`, `validate.py`).
  - Require conventional commit PR titles.
- **Issues & Discussions:**
  - Enable GitHub Issues using the committed `.github/ISSUE_TEMPLATE/`.
  - Disable public wiki (documentation is authoritative in `docs/`).

---

## 5. Post-Launch Next Steps

1. **Compiling Precompiled Binary Releases:**
   - Execute native packaging (`BuildReleaseArchive`) as part of a formal GitHub Actions release workflow to compile `rewrotron` and `rewrotron-mcp` binaries for `linux-amd64` and attach signed checksums.
2. **Community Feedback & Adapter Proposals:**
   - Review incoming community issues for additional source/target candidates (e.g. Django, Spring Boot, Go targets).
3. **Qualification of Additional Source/Target Profiles:**
   - Proceed with roadmap expansion for additional language adapters through neutral IR/TIM composition.

# Quickstart: 5-Minute Developer Evaluation Guide

**Owner:** Developer adoption & evaluation  
**Status:** Runnable from source checkout  
**References:** [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md), [PRODUCT-CLAIM-MATRIX](../11-roadmap/PRODUCT-CLAIM-MATRIX.md), [INSTALLATION](INSTALLATION.md)  

---

## Overview

This guide walks you through evaluating Rewrotron on your local development machine in under 5 minutes.

Rewrotron is an open-source modernization engine that reconstructs legacy application behavior in modern technology stacks using differential verification contracts.

> [!NOTE]
> Precompiled standalone binaries on GitHub Releases are pending publication. Today, Rewrotron is evaluated directly from a source checkout on Linux (`amd64`).

---

## 1. System Prerequisites

Before running the evaluation checks, ensure you have:
- **Go:** `1.25.0+` (specified in [`tools/toolchain.json`](../../tools/toolchain.json))
- **Python:** `3.10+`
- **Git**

Verify your local environment:
```bash
go version
python3 --version
```

---

## 2. Clone the Repository

```bash
git clone https://github.com/devravik/rewrotron.git
cd rewrotron
```

---

## 3. Step 1: Validate Contracts & Machine Schemas (30 Seconds)

Rewrotron uses strict JSON Schema contracts and hash vectors for all IR representations, migration plans, and reports.

Install the JSON schema validator and run the repository contract suite:
```bash
python3 -m pip install 'jsonschema>=4.10,<5'
python3 tools/docs/validate.py
```

**Expected output:**
```text
PASS: 154 schemas; 387 positive/negative fixtures; 71 command/tool entries; ...
```

---

## 4. Step 2: Run the End-to-End CI3-to-Laravel Demonstration (60 Seconds)

Execute the end-to-end differential verification test on Reference Profile #1 (CodeIgniter 3 → Laravel 13).

This demonstration:
1. Loads the legacy CodeIgniter 3 reference fixture ([`fixtures/ci3-reference`](../../fixtures/ci3-reference)).
2. Extracts the Application IR and Target Implementation Model.
3. Generates the idiomatic Laravel 13 vertical slice (routes, models, controllers, middleware).
4. Executes 24 compatibility cases comparing HTTP status codes, headers, response bodies, session cookies, and database state.
5. Injects a deliberate behavioral defect (authorization bypass) and proves the comparator catches it as `MISMATCH`.
6. Applies a diagnosed repair and confirms the target re-evaluates to `PASS`.

```bash
go test -race -v ./internal/domain -run "TestM11_Unit4_DifferentialVerificationAndMutationRepair"
```

**Expected output:**
```text
=== RUN   TestM11_Unit4_DifferentialVerificationAndMutationRepair
--- PASS: TestM11_Unit4_DifferentialVerificationAndMutationRepair (0.12s)
PASS
ok  	github.com/devravik/rewrotron/internal/domain	...
```

For a detailed breakdown of each compatibility case and the mutation repair workflow, see the [CI3-to-Laravel Walkthrough](../04-migration/CI3-LARAVEL-WALKTHROUGH.md).

---

## 5. Step 3: Run the Release Gate Evaluator (30 Seconds)

Evaluate the automated release gate:
```bash
go test -race -v ./internal/domain -run "TestM12_AcceptanceClosure"
```

**Expected output:**
```text
=== RUN   TestM12_AcceptanceClosure
=== RUN   TestM12_AcceptanceClosure/ReleaseQualificationMatrixEvaluation
=== RUN   TestM12_AcceptanceClosure/DeterministicGenerationAndArtifactIntegrity
=== RUN   TestM12_AcceptanceClosure/SecuritySandboxAndSecretRedaction
=== RUN   TestM12_AcceptanceClosure/NativePackagingAndRollbackReadiness
=== RUN   TestM12_AcceptanceClosure/AutomatedReleaseGateEvaluation
=== RUN   TestM12_AcceptanceClosure/CryptographicReleaseReadinessReportValidation
=== RUN   TestM12_AcceptanceClosure/CrossMilestoneRegressionProtection
--- PASS: TestM12_AcceptanceClosure (0.11s)
PASS
```

---

## 6. Step 4: Run the Full Test Suite with Race Detection

To verify complete repository health across all packages (`internal/domain`, `internal/storage`, tools):
```bash
go test -race -count=1 ./...
go vet ./...
```

---

## Next Steps

- Explore the [CI3-to-Laravel Walkthrough](../04-migration/CI3-LARAVEL-WALKTHROUGH.md) for deep technical details.
- Review the [Product Claim Matrix](../11-roadmap/PRODUCT-CLAIM-MATRIX.md) to see what is implemented vs. specified.
- Read [CONTRIBUTING.md](CONTRIBUTING.md) to understand how to contribute using the unit-based discipline.

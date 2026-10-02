# DIFFERENTIAL-TESTING.md — Dual Replay & Differential Verification Algorithm

> **Authority Level:** Mandatory Verification Contract  
> **Status:** Authoritative Core Specification  
> **Subsystems:** `internal/verification`, Replay Engine, Repair Loop  

---

## 1. The Differential Verification Algorithm

Differential testing is the mathematical foundation of Rewrotron's verification guarantee.

```text
               ┌────────────────────────────────────────────────────────┐
               │              Behavior Case (Input & Seed)              │
               └───────────────────────────┬────────────────────────────┘
                                           │
                     ┌─────────────────────┴─────────────────────┐
                     ▼                                           ▼
       ┌───────────────────────────┐               ┌───────────────────────────┐
       │   Legacy Execution (CI3)  │               │ Target Execution (Laravel)│
       └─────────────┬─────────────┘               └─────────────┬─────────────┘
                     ▼                                           ▼
       ┌───────────────────────────┐               ┌───────────────────────────┐
       │   Capture Source Trace    │               │   Capture Target Trace    │
       │ (HTTP, DB, Session, Time) │               │ (HTTP, DB, Session, Time) │
       └─────────────┬─────────────┘               └─────────────┬─────────────┘
                     │                                           │
                     └─────────────────────┬─────────────────────┘
                                           ▼
                             ┌───────────────────────────┐
                             │ Dynamic Value Normalizer  │
                             │ (Mask CSRF, UUID, Time)   │
                             └─────────────┬─────────────┘
                                           ▼
                             ┌───────────────────────────┐
                             │   Multi-Tier Comparator   │
                             │ (Status, Headers, Bodies, │
                             │  DB Mutations, Sessions)  │
                             └─────────────┬─────────────┘
                                           ▼
                             ┌───────────────────────────┐
                             │       CLASSIFICATION      │
                             │  • PASS                   │
                             │  • EXPECTED_CHANGE        │
                             │  • MISMATCH               │
                             │  • UNKNOWN                │
                             └───────────────────────────┘
```

---

## 2. Verdict Classifications

Every differential run produces one of four mutually exclusive classifications:

1. **`PASS`**: Outputs and side effects match perfectly under `EXACT` or `SEMANTIC` comparison rules.
2. **`EXPECTED_CHANGE`**: Outputs differ in an approved, modernized manner documented in the contract (e.g. Modern UUIDv4 format instead of md5 hash, or Laravel's JSON error envelope instead of plain text, provided it was intentional).
3. **`MISMATCH`**: Target execution violates the contract (e.g. status code 500 instead of 200, missing database INSERT, wrong column value).
4. **`UNKNOWN`**: Difference detected whose legitimacy cannot be determined automatically.

---

## 3. What Happens After a Mismatch?

When a `MISMATCH` or `UNKNOWN` verdict is reached, the following workflow is automatically triggered:

```text
Differential Run produces MISMATCH
    ↓
Record Mismatch in project.db (mismatches table)
    ↓
Generate Diagnostic Diff Artifact (.rewrotron/artifacts/reports/mismatch_*.json)
    ↓
Calculate Mismatch Severity (CRITICAL, HIGH, MEDIUM, LOW)
    ↓
Spawn or Update Mission: REPAIR_MISMATCH
    ↓
Invoke LLM Diagnostic Prompt / Agent Repair Loop:
  - Input: Legacy Trace, Target Trace, Unified Diff, Contract
  - Output: Target Code Patch OR Contract Normalization Proposal
    ↓
Apply Patch to Target Application
    ↓
Re-run Differential Verification (Max 3 automated attempts)
    ↓
If still failing after 3 attempts ──► Escalate to Human Review Gate
```

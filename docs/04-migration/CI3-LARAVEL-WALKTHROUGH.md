# CI3-to-Laravel End-to-End Migration Walkthrough

**Owner:** Migration engineering & target qualification  
**Status:** Runnable demonstration on Reference Profile #1  
**References:** [CI3-ADAPTER](CI3-ADAPTER.md), [LARAVEL-ADAPTER](LARAVEL-ADAPTER.md), [TARGET-ARCHITECTURE](TARGET-ARCHITECTURE.md), [DIFFERENTIAL-TESTING](../05-verification/DIFFERENTIAL-TESTING.md), [PRODUCT-CLAIM-MATRIX](../11-roadmap/PRODUCT-CLAIM-MATRIX.md)  

---

## 1. Executive Summary

Reference Profile #1 modernizes legacy **CodeIgniter 3.x** applications targeting **Laravel 13.x (PHP 8.3+)**, initially preserving the underlying relational database schema.

This walkthrough documents the end-to-end vertical slice demonstration qualified in Milestone M11:
- Source discovery and Application IR extraction from [`fixtures/ci3-reference`](../../fixtures/ci3-reference).
- Target Implementation Model 2.1 mapping to idiomatic Laravel constructs.
- Deterministic code generation: `routes/web.php`, Eloquent models, controllers, and session middleware.
- Dual-environment differential execution across **24 compatibility cases**.
- Controlled mutation injection, defect detection (`MISMATCH`), and repair verification (`PASS`).

---

## 2. Legacy Source: The CI3 Reference Fixture

The reference source fixture resides in [`fixtures/ci3-reference`](../../fixtures/ci3-reference) and represents a realistic legacy application:
- **Routing:** URI segments mapped via `application/config/routes.php` and dynamic controller dispatch.
- **Stateful Authentication:** Custom session management using `CI_Session` and cookie-backed credentials.
- **Relational Data:** MySQL tables with active query-builder models (`items`, `users`, `orders`).
- **Authorization Gating:** Resource-ownership checks (e.g. users cannot access other users' private items).

---

## 3. Vertical Slice Lowering: CodeIgniter 3 → Laravel 13

Rewrotron maps legacy behaviors into idiomatic Laravel 13 components:

| Legacy CI3 Component | Rewrotron Intermediate Representation | Modern Laravel 13 Lowering |
| :--- | :--- | :--- |
| Dynamic URI routes in `routes.php` | Application IR Entrypoints & Route Map | Explicit named routes in `routes/web.php` |
| `CI_Controller` actions | Behavioral Vertical Slice Directives | Thin controllers (`App\Http\Controllers\ReconstructedController`) |
| `CI_Model` query builder queries | Relational Identity & Mutation Schemas | Idiomatic Eloquent models (`App\Models\Item`, `App\Models\User`) |
| `CI_Session` flashdata & user data | Session Continuity Contract | Laravel `EncryptCookies` & `StartSession` middleware |
| Custom database transaction calls | Atomic Transaction Bounds | `DB::transaction()` closures preserving rollback safety |

---

## 4. The 24 Compatibility Cases

Differential verification executes source and target in independent simulated execution witnesses across 24 test cases, categorized into four behavioral dimensions:

1. **Read-Only Catalog & Health (6 Cases):**
   - Health check endpoints (`/health`).
   - Catalog browsing and pagination.
   - Public product lookups returning identical JSON payloads and HTTP 200 statuses.

2. **Authentication & Session Lifecycle (6 Cases):**
   - User login with valid credentials (session cookie issuance).
   - Login rejection with invalid credentials (HTTP 401).
   - Session continuity across subsequent requests.
   - Logout and session invalidation.

3. **Stateful Mutations & Database Transactions (6 Cases):**
   - Item creation writing exact rows to the relational database.
   - Balance updates within an atomic transaction.
   - Transaction rollback on simulated mid-flight failure, leaving database unchanged.

4. **Authorization Gating & Resource Ownership (6 Cases):**
   - Alice accesses her own private item (HTTP 200).
   - Bob attempts to access Alice's private item without permission (HTTP 403 Forbidden).
   - Unauthenticated access to protected resources (HTTP 401 Unauthorized).

---

## 5. Controlled Mutation Detection & Repair Demonstration

To prove that the verification comparator is sensitive to real semantic regressions and cannot be fooled by synthetic passes, Rewrotron tests a deliberate behavioral defect:

### A. Defect Injection
A mutation is injected into the target controller: Bob's unauthorized request to read Alice's item is altered to return HTTP 200 with Alice's data instead of HTTP 403 Forbidden:
```go
// Deliberate defect: Bob is authorized to read Alice's item (HTTP 200 instead of 403)
mutantInput := buildCase("item-read-bob-denied", 403, `{"error":"owner"}`, dbItemAlice, cookieAuth)
for _, targetCtx := range mutantInput.TargetContexts {
    targetCtx.Namespaces.Result.Status = 200
    targetCtx.Namespaces.Result.Body = TaggedValue{
        Type:  ValueTypeString,
        Value: `{"id":1,"owner":1,"name":"Widget A"}`,
    }
}
```

### B. Mismatch Detection
The differential comparator evaluates source against target. It immediately detects the discrepancy and halts with `MISMATCH`:
```text
Verdict: MISMATCH
Failure: HTTP Status code divergence (Expected: 403, Actual: 200)
Namespace: Result.Status
Evidence Ref: sha256:...
```

### C. Repair Resolution
The defect is repaired in the target implementation to correctly enforce ownership. The comparator re-evaluates the repaired target and returns `PASS` with full cryptographic evidence:
```text
Verdict: PASS
All 24 compatibility cases verified equivalent.
```

---

## 6. How to Run the Walkthrough

You can execute this exact walkthrough on your local machine:
```bash
go test -race -v ./internal/domain -run "TestM11_Unit4_DifferentialVerificationAndMutationRepair"
```

To run the complete end-to-end acceptance closure (including report schema validation):
```bash
go test -race -v ./internal/domain -run "TestM11_AcceptanceClosure"
```

---

## 7. Known Scope Limitations

1. **Preserved Relational Schema:** Reference Profile #1 retains the existing MySQL schema. Dynamic schema alteration or automated NoSQL migrations are not performed.
2. **Asynchronous Background Queues:** Complex asynchronous worker pools or distributed messaging buses (e.g. RabbitMQ/Kafka) are out of scope for the reference profile.
3. **No Arbitrary PHP:** Verification guarantees apply strictly to the characterized CodeIgniter 3 patterns and vertical slices under the approved Gate 2 plan.

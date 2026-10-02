# COMPATIBILITY-SPEC.md — Equivalence Definition & Normalization Specification

> **Authority Level:** Mandatory Verification Contract (Critical Foundation Document #2)  
> **Status:** Authoritative Core Specification  
> **Subsystems:** `internal/verification`, Differential Comparator, Repair Loop  

---

## 1. What Exactly Does "Same Application" Mean?

To eliminate hand-waving and subjective claims of compatibility, Rewrotron defines behavioral equivalence across 11 explicit dimensions. Two implementations are **behaviorally equivalent** if and only if they produce identical observable outputs and side effects under identical preconditions.

### The 11 Dimensions of Equivalence

1. **HTTP Status Code Equivalence:** Status codes must match exactly (e.g. 200, 201, 302, 400, 401, 403, 404, 500). A 200 returned where a 302 was expected is an unconditional failure.
2. **Redirect Behavior Equivalence:** Redirect targets (via `Location` header) must resolve to the identical relative path and query parameters.
3. **API Payload Equivalence:** JSON/XML responses must carry equivalent data structures, data types, keys, and values.
4. **Authentication Equivalence:** An unauthenticated request must be challenged identically; an authenticated request must authenticate the identical principal and permissions.
5. **Authorization Equivalence:** Requests lacking required permissions must be forbidden with identical status codes and error bodies.
6. **Session State Equivalence:** Flash messages, session variable sets/unsets, and persistence across consecutive requests must match.
7. **Database State Equivalence:** The set of database tables mutated (INSERT, UPDATE, DELETE), the columns affected, and the final relational records must be identical.
8. **External Side-Effect Equivalence:** External API calls, email dispatches, message queue publishes, and webhook triggers must match in destination and payload.
9. **HTML Equivalence:** When rendering server-side HTML views, semantic structure, DOM elements, forms, and input fields must be equivalent.
10. **File & Asset Behavior:** File uploads, downloads, and generated PDFs must produce identical content or valid representations.
11. **Error Behavior Equivalence:** Validation errors and intentional business failure responses must emit equivalent error keys, status codes, and user messages.

---

## 2. The Three Comparison Modes

Equivalence is evaluated using a 3-tier policy:

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        THE 3 COMPARISON MODES                          │
├────────────────────────────────────────────────────────────────────────┤
│ 1. EXACT (Byte-for-Byte)                                               │
│    • Applied when values are strictly deterministic.                   │
│    • Required for: HTTP status codes, redirect paths, static JSON      │
│      keys, database table names, column names, business rule values.   │
├────────────────────────────────────────────────────────────────────────┤
│ 2. SEMANTIC (Canonical Semantic Equivalence)                           │
│    • Applied to structured data formats where formatting varies.        │
│    • JSON: Ignores key ordering, formatting whitespace, float notation│
│    • HTML: Normalizes whitespace, tag indentation, attribute quotes.   │
│    • SQL: Normalizes SQL whitespace and keyword casing.                │
├────────────────────────────────────────────────────────────────────────┤
│ 3. NORMALIZED_DYNAMIC (Dynamic-Value Normalization)                    │
│    • Applied ONLY to fields demonstrated to vary across repeat runs.    │
│    • Nonces, CSRF tokens, session IDs, UUIDs, request IDs, timestamps. │
│    • Values are extracted, validated against format schemas (e.g.      │
│      UUIDv4 or ISO8601), and replaced with canonical masks (<UUID>).   │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. The Invariant Rule on Uncertainty

> ### **CRITICAL INVARIANT:**  
> **UNKNOWN DIFFERENCES ARE NEVER PASSES.**

If a differential verification run detects a discrepancy between the Source Application and Target Application:
1. If the difference is explicitly governed by an approved `NORMALIZED_DYNAMIC` or `EXPECTED_CHANGE` rule, it is classified as a pass.
2. If the difference is NOT governed by an approved rule, the verifier **MUST classify the result as `MISMATCH` or `UNKNOWN`**.
3. It is **STRICTLY FORBIDDEN** for an AI agent, repair algorithm, or human developer to automatically convert an `UNKNOWN` difference into a `PASS` without recording an approved rationale in the experiment ledger.

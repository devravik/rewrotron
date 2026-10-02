# TERMINOLOGY.md — Precise Domain Terminology & Abbreviations

> **Version:** 1.0  
> **Status:** Standard Reference Dictionary  

---

This document resolves subtle distinctions between similar terms across the Rewrotron system.

## 1. Distinction Matrix

### 1.1 "Behavior Case" vs. "Behavior Contract"
- **Behavior Case (`BehaviorCase`):** A specific, concrete test input/output instance (e.g. `POST /checkout` with cart ID 42 and customer Alice expecting status 302). It is executable.
- **Behavior Contract (`BehaviorContract`):** The general, parameterized rule governing the endpoint under all circumstances (e.g. `POST /checkout` requires authentication, deducts inventory, sends confirmation email, redirects to `/orders/{id}`).

### 1.2 "Migration Unit" vs. "Source File"
- **Migration Unit (`MigrationUnit`):** A functional vertical slice (e.g., User Authentication or Order Creation). It spans routes, business rules, database mutations, and views.
- **Source File:** A physical file in the legacy repository (e.g. `application/controllers/Order.php`). **A source file is never a migration unit.**

### 1.3 "Observation Trace" vs. "Golden Master"
- **Observation Trace (`ObservationTrace`):** The raw, recorded runtime trace from an execution of the legacy application.
- **Golden Master (`GoldenMaster`):** A sanitized, normalized, and locked baseline recording verified across multiple runs to serve as the ground-truth benchmark during differential verification.

### 1.4 "Mismatch" vs. "Finding"
- **Finding:** A static or runtime observation about the legacy codebase (e.g. dead code candidate, raw SQL string concatenation, session variable usage).
- **Mismatch:** A differential verification failure between legacy and target executions (e.g., status code difference, header mismatch, divergent SQL mutations).

### 1.5 "Lesson" vs. "Pattern"
- **Lesson:** An insight learned from diagnosing a specific mismatch in the current project (e.g., CI3 active record inserts timestamps in local timezone; Laravel Eloquent defaults to UTC).
- **Pattern:** A generalized, multi-project lesson promoted into global knowledge after passing cross-project validation criteria.

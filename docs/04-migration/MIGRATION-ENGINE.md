# MIGRATION-ENGINE.md — Vertical Slice Migration Engine Specification

> **Authority Level:** Mandatory Migration Contract  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** `internal/migration`, Target Adapters, Repair Loop  

---

## 1. Canonical Migration Strategy

Rewrotron explicitly rejects whole-file or line-by-line transpilation. Converting `LegacyController.php` directly into `ModernController.php` merely reproduces legacy structural defects, God-class antipatterns, and tight coupling in the modern target framework.

### The Behavioral Slicing Paradigm
```text
Route / Entrypoint
        ↓
Behavioral Slice (Inputs, DB Mutations, Session, Views)
        ↓
Migration Unit
        ↓
Idiomatic Clean Architecture Target Implementation
```

Target code is generated from the **Behavioral Contract**, using modern idioms (Form Requests, Action Classes, Eloquent Models, Resource Collections, Blade Templates), completely independent of legacy file structures.

---

## 2. The 7-Stage Migration Lifecycle for a Unit

Every `MigrationUnit` proceeds through these 7 mandatory phases:

```text
┌─────────────────────────────────────────────────────────────┐
│ 1. DISCOVER                                                 │
│    Extract routes, static dependencies, and entrypoints.     │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 2. CHARACTERIZE                                             │
│    Synthesize BehaviorContract and executable BehaviorCases.│
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 3. PLAN                                                     │
│    Resolve dependencies, shared models, and migration order.│
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 4. GENERATE                                                 │
│    Target adapter emits clean, idiomatic target code.       │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 5. VERIFY                                                   │
│    Execute differential testing against legacy oracle.       │
└──────────────────────────────┬──────────────────────────────┘
                               │
                ┌──────────────┴──────────────┐
                │                             │
             [PASS]                       [MISMATCH]
                │                             │
                │              ┌──────────────▼──────────────┐
                │              │ 6. REPAIR (Loop)            │
                │              │    Diagnose mismatch &      │
                │              │    patch target code.       │
                │              └──────────────┬──────────────┘
                │                             │
                ▼                             ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. OPTIMIZE & FREEZE                                        │
│    Clean up redundant queries, benchmark latency, and       │
│    freeze Golden Master verification suite.                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Progression Criteria: When May the Engine Proceed?

The migration engine is **strictly prohibited** from advancing to the next `MigrationUnit` unless all of the following conditions are met:

1. **Zero Open Defects:** All differential test cases for the current unit pass without divergence or unclassified failures.
2. **All Mismatches Formally Justified:** Any differences in output are documented as `EXPECTED_CHANGE` (e.g. CSRF token structure or ISO8601 formatting) with an approved normalization rule.
3. **Target Cleanliness Verified:** Target code passes PHP syntax validation (`php -l`), static analysis (PHPStan level 8+), and coding standards (`pint` or `php-cs-fixer`).
4. **Database State Equivalence Confirmed:** Post-execution database diff confirms identical rows, columns, and relations modified.
5. **Regression Verification:** Previously passed `MigrationUnits` are re-verified to ensure the new unit did not introduce breaking cross-cutting side effects.

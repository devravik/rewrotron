# PRODUCT.md — Rewrotron Product Specification

> **Status:** Authoritative Foundation Specification  
> **Target Release:** Rewrotron 1.0  
> **Domain:** Universal Legacy Web Application Reconstruction & Modernization  

---

## 1. Product Statement & Core Promise

Rewrotron is an autonomous, behavior-driven modernization engine designed to reconstruct complex legacy web applications into modern, clean-architecture target stacks.

Rewrotron understands **applications, not frameworks**. A legacy web application is characterized by its externally observable behavior, entrypoints, state transitions, business rules, and side effects—not the syntax or idiosyncrasies of the legacy framework in which it was authored.

### The Product Promise
> **"Rewrotron reconstructs an existing web application's observable behavior in a modern target implementation through a universal behavior model, allowing unnecessary legacy implementation baggage to be omitted, target architectures to be substantially modernized, and equivalence to be mathematically proven via automated differential verification."**

---

## 2. The $N + M$ Architectural Advantage: The Three-Tier IR

Traditional migration tools build point-to-point transpilers (e.g. CI3 to Laravel, Django to FastAPI, Rails to Go). Supporting $N$ source frameworks across $M$ target frameworks would require $N \times M$ fragile custom translators:

```text
POINT-TO-POINT MATRIX (Fragile, N × M translators):
CI3 ─────────► Laravel
CI3 ─────────► Go
CI3 ─────────► Rust
Django ──────► Go
Rails ───────► Go
Symfony ─────► Laravel
```

Rewrotron solves this through the **Universal Intermediate Representation Pipeline**, reducing complexity to **$N$ Source Adapters + $M$ Target Adapters**:

```text
Source Frameworks (PHP, Python, Ruby, Java, Node, .NET)
                     │
                     ▼
             [ Source Adapters ]
                     │
                     ▼
             1. SOURCE IR
   (Syntax, ASTs, symbols, classes, methods, calls, declared routes)
                     │
                     ▼
           2. APPLICATION IR
   (Controllers, services, data models, workflows, dependencies)
                     │
                     ▼
             3. BEHAVIOR IR
   (Entrypoints, inputs, state transitions, business rules, side effects, outputs)
                     │
                     ▼
             [ Target Adapters ]
                     │
                     ▼
Target Stacks (Laravel, Go std/fiber, Rust axum, Python fastapi, TypeScript nest)
```

With this architecture:
- Adding a **Django** source adapter immediately unlocks migration to **Laravel, Go, Rust, and Python**.
- Adding a **Rust (Axum)** target adapter immediately unlocks migration from **CodeIgniter, Laravel, Symfony, Django, and Rails**.

---

## 3. Separation of Framework from Architecture

Rewrotron does not merely allow the user to select a target framework; it decouples **Language**, **Framework**, and **Architectural Style**:

```yaml
target:
  language: go
  framework: std-http
  architecture:
    style: modular_monolith
    service_layer: domain_driven
  database:
    preserve_existing: true
  api:
    compatibility: strict
```

Target code generation conforms to the requested clean architecture style rather than slavishly mirroring the legacy code's structural flaws.

---

## 4. The 3 Hard User Confirmation Gates

To guarantee that autonomous agents never execute destructive actions, misclassify a project, or operate on false assumptions, Rewrotron enforces **Three Hard Confirmation Gates**:

```text
┌─────────────────────────────────────────────────────────────┐
│ 1. SOURCE DETECTION CONFIRMATION (Gate 1)                   │
│    - Engine automatically detects language, framework,      │
│      version, runtime, routes, database, integrations.      │
│    - Presents findings with confidence scores.              │
│    - REQUIRES explicit user confirmation / overrides        │
│      before deep analysis or project initialization.        │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 2. TARGET & MIGRATION PLAN CONFIRMATION (Gate 2)            │
│    - Proposes target stack, architecture style,             │
│      behavioral units to migrate, and dead code to drop.    │
│    - REQUIRES explicit user approval before code generation.│
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ 3. PRODUCTION CUTOVER CONFIRMATION (Gate 3)                 │
│    - Verifies differential equivalence across all units.     │
│    - REQUIRES human sign-off before production cutover,     │
│      destructive migration, or live deployment.             │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. Adapter Availability & Maturity Tiers

Rewrotron distinguishes **architectural support** (the core handles any web application) from **adapter implementation maturity**:

### Source Adapters
| Framework / Platform | Maturity Baseline | Role |
| :--- | :--- | :--- |
| **CodeIgniter 3.x** | `STABLE` | Initial reference source adapter |
| **Generic PHP** | `BETA` | Fallback for non-framework or custom legacy PHP |
| **Laravel (Legacy)** | `BETA` | Modernization of legacy Laravel 4/5/6 |
| **Symfony 2/3** | `EXPERIMENTAL` | Enterprise legacy PHP |
| **Django / Flask** | `EXPERIMENTAL` | Python monoliths |
| **Ruby on Rails** | `EXPERIMENTAL` | Legacy Rails applications |

### Target Adapters
| Target Stack | Maturity Baseline | Role |
| :--- | :--- | :--- |
| **Laravel 13.x (PHP 8.3+)** | `STABLE` | Initial reference target adapter |
| **Go (std-http / Fiber)** | `BETA` | Cloud-native microservices / modular monolith |
| **Rust (Axum)** | `EXPERIMENTAL` | Ultra-high performance modernization |
| **Python (FastAPI)** | `EXPERIMENTAL` | Modern Python async services |

---

## 6. Definition of a "Successful Migration"

A migration unit or complete application is deemed **successfully migrated** if and only if:
1. **Full Behavior IR Coverage:** 100% of discovered active endpoints and workflows have an executable behavioral contract.
2. **Differential Verification Pass:** Replaying all recorded and synthesized test cases produces an identical or normalized match between source and target:
   - HTTP status codes match exactly.
   - Redirect headers match exactly.
   - Deterministic response bodies match exactly (byte-for-byte or normalized canonical semantic comparison).
   - Relational database mutations (INSERTs, UPDATEs, DELETEs) match in table, columns, and values.
   - Session flash data and state transitions match.
3. **Zero Unexplained Mismatches:** Every discrepancy is classified as `EXPECTED_CHANGE` (e.g. upgraded CSRF tokens) with recorded justification. Zero `UNKNOWN` differences.
4. **Performance Non-Regression:** The target stack response latency and memory consumption are equal to or better than the legacy baseline under identical load.
5. **Independence:** The target application runs completely independently without requiring Rewrotron at runtime.

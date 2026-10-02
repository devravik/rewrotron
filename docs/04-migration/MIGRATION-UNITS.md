# MIGRATION-UNITS.md — Migration Unit Granularity & Slicing Heuristics

> **Version:** 1.0  
> **Status:** Migration Reference Guide  

---

## 1. Granularity Guidelines

A `MigrationUnit` must be sized appropriately to balance isolation with cohesion:

- **Too Small (Anti-Pattern):** An individual helper function or single private method. Cannot be tested with end-to-end HTTP requests.
- **Too Large (Anti-Pattern):** An entire billing or admin module with 50 routes. Feedback loops become unmanageably long and debugging mismatches becomes overwhelming.
- **Optimal Granularity:**
  - A single atomic REST endpoint (e.g. `POST /api/v1/users`).
  - A cohesive CRUD resource set (e.g. `GET/POST/PUT/DELETE /products`).
  - A tightly-coupled workflow (e.g. `POST /checkout/address` -> `POST /checkout/pay`).

---

## 2. Dependency Topological Sorting

Migration units frequently depend on shared entities:
- Common Database Tables (e.g. `users`, `settings`).
- Authentication & Session Subsystems.
- Shared Layout Views and UI Partials.

The Migration Planner constructs a Directed Acyclic Graph (DAG) of migration units and topologically sorts them so foundational slices (e.g. Authentication, User Profile) are migrated and verified before downstream transactional slices (e.g. Orders, Checkout).

# ADR 0004: Vertical Slice Migration Units Over File-by-File Translation

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
A major cause of failed migration initiatives is the "horizontal" or "file-by-file" approach (migrating all models first, then all controllers, then all views). This leaves the application broken and unexecutable for months until the entire migration is complete.

## Decision
Decompose application migration into autonomous **Vertical Slices** called **Migration Units**. A Migration Unit encapsulates an independently migratable and verifiable externally observable slice of functionality (e.g. `POST /auth/login`, `GET /products`, `POST /orders/checkout`).

## Alternatives Considered
- **File-by-file migration:** Migrating `User_model.php` to `User.php`, then `Auth.php` to `AuthController.php`. Rejected because individual files cannot be differentially tested in isolation.
- **Big-Bang complete rewrite:** Migrating the entire codebase before running any tests. Rejected due to extreme regression risks and long feedback loops.

## Consequences
- **Positive:** Each Migration Unit can be planned, generated, differentially verified, repaired, and marked "Done" independently; rapid feedback loops; incremental verification.
- **Negative:** Shared models and database tables must be reconciled across multiple vertical slices as they are introduced.

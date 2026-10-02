# ADR 0003: Behavior Model as Authority Over Source Syntax

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
Legacy applications contain decades of architectural rot: monolithic procedural scripts, global state mutations, spaghetti controllers, and unreachable branches. If a modernization engine treats the legacy source code's structure as the blueprint, it will reproduce the exact same architectural defects in the new language or framework.

## Decision
Treat legacy source code exclusively as **evidence of intended behavior**, never as structural authority. The canonical **Behavior Model** (`BehaviorContract`, `BehaviorCase`, state transitions, input/output validation, and database side-effects) is the sole authority from which target code is generated.

## Alternatives Considered
- **Direct AST Translation:** Rewriting CodeIgniter AST nodes directly into Laravel AST nodes. Rejected because it preserves bad legacy architecture, God objects, and dead code.
- **Specification-Only from Human Documentation:** Relying on human specifications. Rejected because legacy apps rarely possess accurate or up-to-date documentation.

## Consequences
- **Positive:** Enables true modernization; target code adheres to clean architectural patterns; completely dead code is omitted naturally; target stack can be swapped (e.g. Laravel today, Go or Rust tomorrow) without changing the behavior model.
- **Negative:** Requires deeper observation, reachability analysis, and contract synthesis before any code generation can start.

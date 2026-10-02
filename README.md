# Rewrotron

> **The Universal Behavior-Driven Legacy Application Reconstruction Engine**

Rewrotron is an open-source, vendor-neutral modernization engine that reconstructs the externally observable behavior of legacy web applications in modern software stacks.

Rewrotron understands **applications, not frameworks**. A legacy web application is characterized by its externally observable behavior, entrypoints, state transitions, business rules, and side effects—not the syntax or idiosyncrasies of the legacy framework in which it was authored.

---

## The $N + M$ Architectural Advantage: Three-Tier IR

Instead of building fragile $N \times M$ point-to-point transpilers, Rewrotron decouples source understanding from target generation via three progressive intermediate representations:

```text
Source Codebase (PHP, Python, Ruby, Java, Node, .NET)
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

- **Pluggable Source Adapters:** Extract legacy code into the universal model (CodeIgniter 3.x is the initial tested reference, alongside Generic PHP fallback, with Laravel, Symfony, Django, Rails planned).
- **Pluggable Target Adapters:** Render the universal model into modern stacks (Laravel 13.x is the initial tested reference, alongside Go std-http/Fiber, Rust Axum, and Python FastAPI).
- **Architecture Decoupled from Framework:** Users configure target language, framework, and architectural style (e.g. `modular_monolith` or `clean_architecture`).

---

## Mandatory Engine-Enforced Confirmation Gates

Rewrotron enforces a strict **Detect ➔ Show ➔ Confirm ➔ Plan ➔ Show ➔ Confirm ➔ Execute** lifecycle. Agents cannot infer confirmation from silence or conversational context:

1. **Gate 1: Source Detection Confirmation:** Engine inspects source repository and presents findings with confidence scores. Deep analysis is blocked until explicit user confirmation is recorded.
2. **Gate 2: Target & Migration Plan Confirmation:** Engine presents proposed target stack, architecture style, and behavioral migration units. Code generation is blocked until explicit user confirmation is recorded.
3. **Gate 3: Production Cutover Confirmation:** Requires human sign-off before any live deployment or production data modification.

---

## Standard CLI Flow

```bash
# 1. Inspect repository and detect framework, runtime, routes, database
rewrotron detect /path/to/legacy-app

# 2. Confirm detected baseline or provide explicit overrides (Unlocks Gate 1)
rewrotron confirm source --php 8.2

# 3. Perform static reachability analysis (Emits Source IR)
rewrotron analyze

# 4. Observe live or synthetic traffic (Captures runtime execution traces)
rewrotron observe

# 5. Synthesize Application IR and canonical Behavior IR contracts
rewrotron discover

# 6. Formulate migration units and target architecture plan
rewrotron plan --target go --style modular_monolith

# 7. Review and confirm target migration plan (Unlocks Gate 2)
rewrotron confirm plan

# 8. Migrate a vertical behavioral slice (e.g., auth or orders)
rewrotron migrate --unit orders-create

# 9. Run automated differential verification comparing source vs target
rewrotron verify --unit orders-create

# 10. Check migration progress and query the next recommended mission
rewrotron status
rewrotron next
```

---

## Documentation & Specification Index

- [AGENTS.md](file:///media/System/rewrotron/AGENTS.md) — The Repository Constitution & Invariants
- [docs/AGENT-WORKFLOW.md](file:///media/System/rewrotron/docs/AGENT-WORKFLOW.md) — Step-by-Step Developer & Agent Workflow
- [docs/00-vision/PRODUCT.md](file:///media/System/rewrotron/docs/00-vision/PRODUCT.md) — Product Promise, Scope, and $N+M$ IR Advantage
- [docs/01-architecture/SYSTEM-DESIGN.md](file:///media/System/rewrotron/docs/01-architecture/SYSTEM-DESIGN.md) — Canonical System Architecture & Dependency Graph
- [docs/02-domain/BEHAVIOR-MODEL.md](file:///media/System/rewrotron/docs/02-domain/BEHAVIOR-MODEL.md) — Three-Tier IR & Behavioral Contract Specification
- [docs/02-domain/PROJECT-LIFECYCLE.md](file:///media/System/rewrotron/docs/02-domain/PROJECT-LIFECYCLE.md) — Project State Machine & Mandatory Gates
- [docs/05-verification/COMPATIBILITY-SPEC.md](file:///media/System/rewrotron/docs/05-verification/COMPATIBILITY-SPEC.md) — Equivalence Definition & Normalization Policies
- [docs/06-agent/AGENT-MODEL.md](file:///media/System/rewrotron/docs/06-agent/AGENT-MODEL.md) — Agent vs Deterministic Engine Boundary & Gate Enforcement
- [docs/06-agent/CLI-SPEC.md](file:///media/System/rewrotron/docs/06-agent/CLI-SPEC.md) — Machine & Human Command Line Specifications
- [docs/06-agent/MCP-SPEC.md](file:///media/System/rewrotron/docs/06-agent/MCP-SPEC.md) — Model Context Protocol Tools & Resources Contract
- [docs/08-storage/SQLITE-SCHEMA.md](file:///media/System/rewrotron/docs/08-storage/SQLITE-SCHEMA.md) — Relational Data Definition (Tables, DDL, Confirmation Gates)

---

## License

Rewrotron is licensed under the Apache License, Version 2.0. See [LICENSE](file:///media/System/rewrotron/LICENSE) for details.

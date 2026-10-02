# ARCHITECTURE.md — High-Level Component Boundaries & Modularity

> **Version:** 1.0  
> **Status:** Architecture Reference Guide  

---

## 1. Clean Architecture & Port/Adapter Pattern

Rewrotron adheres to hexagonal / ports-and-adapters architecture to ensure complete decoupling of core modernization rules from runtime environments and target frameworks:

```text
               ┌────────────────────────────────────────────────────────┐
               │                    Driving Adapters                    │
               │   CLI (`cmd/rewrotron`)   |   MCP (`cmd/rewrotron-mcp`)│
               └───────────────────────────┬────────────────────────────┘
                                           │
                                           ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                                 Domain Boundary                                  │
│                                                                                  │
│   ┌──────────────────────────────────────────────────────────────────────────┐   │
│   │                              Domain Entities                             │   │
│   │   Project, Route, BehaviorCase, BehaviorContract, MigrationUnit, etc.     │   │
│   └──────────────────────────────────┬───────────────────────────────────────┘   │
│                                      │                                           │
│   ┌──────────────────────────────────▼───────────────────────────────────────┐   │
│   │                              Domain Services                             │   │
│   │   ReachabilityAnalyzer, TraceNormalizer, DiffComparator, PlanScheduler  │   │
│   └──────────────────────────────────┬───────────────────────────────────────┘   │
│                                      │                                           │
│   ┌──────────────────────────────────▼───────────────────────────────────────┐   │
│   │                             Port Interfaces                              │   │
│   │   StoragePort, SourceAdapterPort, TargetAdapterPort, SandboxPort         │   │
│   └──────────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────┬───────────────────────────────────────────┘
                                       │
                                       ▼
               ┌────────────────────────────────────────────────────────┐
               │                    Driven Adapters                     │
               │   SQLite Database | Local Filesystem | Docker Sandbox  │
               │   CI3 Source Parser | Laravel 13 Code Generator        │
               └────────────────────────────────────────────────────────┘
```

---

## 2. Core Modules in `internal/`

1. **`internal/project`**: Project workspace initialization, `.rewrotron` directory lifecycle, configuration parsing (`project.yaml`), and state tracking.
2. **`internal/policy`**: Equivalence policies (Exact, Semantic, Normalized Dynamic), safety constraints, and sandbox access controls.
3. **`internal/analyzer`**: PHP AST parsers, reachability call graph builders, database query analyzers, and symbol resolution engines.
4. **`internal/observer`**: HTTP proxy/middleware interceptors, execution trace recorders, and dynamic correlation engines.
5. **`internal/behavior`**: Canonical behavior graph synthesis, contract validation, and workflow state machine builders.
6. **`internal/migration`**: Dependency topological sorting, vertical slice planner, and code generation coordinator.
7. **`internal/verification`**: Dual replay runner, dynamic variance normalizers, diff matchers, and failure diagnostic classifier.
8. **`internal/storage`**: SQLite repository implementations, database migration runners, and transaction managers.
9. **`internal/artifacts`**: Artifact lifecycle, hashing, deduplication, content-addressable storage, and sensitive data redaction.
10. **`internal/missions`**: Task scheduler, priority heuristic calculation, and agent mission ledger.

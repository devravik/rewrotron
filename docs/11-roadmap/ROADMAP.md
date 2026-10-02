# ROADMAP.md — Engineering Phases & Strategic Evolution

> **Version:** 1.0  
> **Status:** Strategic Roadmap  

---

## The 11-Phase Implementation Roadmap

Rewrotron's engineering journey is structured across 11 discrete phases:

```text
Phase 0: Foundation & Contracts
  ├── AGENTS.md, Specs, Schemas, SQLite DDL, CI Baseline.
  └── Delivery: Fully specified contracts and test harnesses.

Phase 1: CodeIgniter 3 Static Analyzer
  ├── Go AST parser, symbol indexer, routes.php parser, reachability graph.
  └── Delivery: `rewrotron analyze` populates symbols and routes.

Phase 2: CI3 Runtime Observer
  ├── Injected hooks, trace recorder, SQLite trace writer, redaction.
  └── Delivery: `rewrotron observe` records ground-truth traces.

Phase 3: Behavior & Workflow Discovery
  ├── Behavior graph builder, contract synthesizer, workflow miner.
  └── Delivery: `rewrotron discover` emits BehaviorContracts.

Phase 4: Laravel 13 Target Generator
  ├── Clean Architecture templates, route mapper, Eloquent model generator.
  └── Delivery: `rewrotron migrate` emits idiomatic Laravel 13 code.

Phase 5: Differential Verifier & Repair Loop
  ├── Replay engine, multi-tier comparator, dynamic normalizer, patch loop.
  └── Delivery: `rewrotron verify` proves equivalence or drives repair.

Phase 6: Agent Interfaces (CLI & MCP)
  ├── JSON CLI envelopes, MCP server (`rewrotron-mcp`), Agent Skill.
  └── Delivery: Claude Code, Cursor, Antigravity autonomous operation.

Phase 7: Knowledge & Continuous Learning
  ├── Lesson recorder, pattern promoter, experiment ledger, skill evolution.
  └── Delivery: Automatic cross-project knowledge accumulation.

Phase 8: Performance Modernization
  ├── Profiler, N+1 query optimizer, caching suggestions, load benchmark.
  └── Delivery: Proven target performance non-regression.

Phase 9: Additional Target Adapters
  ├── Go (Gin / Fiber), Python (FastAPI), Rust (Axum).
  └── Delivery: Multi-language reconstruction from identical behavior models.

Phase 10: Additional Source Adapters
  ├── Zend Framework, CakePHP, Symfony 2/3, raw procedural PHP.
  └── Delivery: Broad legacy enterprise modernization support.
```

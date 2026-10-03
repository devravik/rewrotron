# GOALS.md — Explicit System Goals for Rewrotron

> **Version:** 1.0  
> **Status:** Approved Target Objectives  

---

## 1. Primary Objectives (v1 Baseline)

### G1: Behavior-First Reconstruction
- Reconstruct web applications based on verifiable external behavior rather than source syntax.
- Ensure all business rules, database mutations, session state, and external side-effects are captured in canonical contracts before code generation.

### G2: Vertical Slice Migration Pipeline
- Support granular migration of discrete endpoints, workflows, and routes independently.
- Enable incremental verification, repair, and signoff rather than all-or-nothing big-bang cutovers.

### G3: CodeIgniter 3.x to Laravel 13.x Reference Implementation
- Provide a robust reference adapter migrating PHP CodeIgniter 3.1.x applications to idiomatic Laravel 13.x (PHP 8.3+).
- Support standard CodeIgniter paradigms: controllers, models, helpers, libraries, hooks, and views.

### G4: Automated Differential Verification
- Implement a dual-replay engine comparing source and target outputs under identical preconditions.
- Support multi-tier equivalence matching:
  - Exact Byte-for-Byte comparison.
  - Canonical Semantic comparison (JSON formatting, HTML whitespace).
  - Normalized Dynamic comparison (UUIDs, timestamps, session IDs, CSRF tokens).

### G5: Objective Dead Code Elimination
- Formally identify, classify, and safely eliminate dead legacy code.
- Provide reproducible, integrity-hashed reachability and observation evidence justifying why omitted code is not present in the target application.

### G6: Vendor-Neutral Agent Architecture
- Core engine, CLI, analyzers, observers, and verifiers implemented in pure Go without vendor lock-in.
- Full Model Context Protocol (MCP) server compliance and JSON CLI outputs allowing any AI model or agent framework to operate Rewrotron.

### G7: Local-First and Enterprise Safe
- Default operation runs locally on customer premises with SQLite storage and local filesystem artifacts.
- Enforced sandbox policies and negative security fixtures against unauthorized production mutations, third-party network egress, or accidental credential leakage.

---

### G8: Transparent Execution and User Supervision

- Every long-running task exposes stage, task, factual reason, discoveries, changes, results, blockers, next action and needed decisions.
- One committed execution event/state model powers CLI/MCP and future desktop clients; progress has a persisted accounting basis and freshness.
- Timeline/drill-down, return-session changes, verbosity, exact-subject approvals and safe persistent pause/continue are first-release engine requirements.
- Public explanations cite evidence without exposing hidden reasoning; no fake percentages, success claims or alert-based approvals.

See [EXECUTION-OBSERVABILITY](../06-agent/EXECUTION-OBSERVABILITY.md) for normative semantics.

## 2. Secondary Objectives (Phase 2+)

- **Multi-Target Ecosystem:** Extensible target architecture supporting Go (Gin/Fiber), Python (FastAPI), and Rust (Axum/Actix-web).
- **Multi-Source Adapters:** Support for legacy Zend Framework, CakePHP 2/3, Symfony 2/3, and raw procedural PHP monoliths.
- **Continuous Learning & Skill Evolution:** Autonomous accumulation of pattern lessons and reviewed proposals for migration skills across repeated runs.
- **Automated Performance Optimization:** Static and dynamic profiling identifying redundant queries, N+1 patterns, and caching opportunities during migration.

## Evidence and release scope

These are target objectives, not implemented capability claims. The initial supported profile, maturity requirements, finite-corpus limits and project acceptance criteria are owned by [PRODUCT](PRODUCT.md). Dynamic values require scoped relation validation; generic HTML whitespace stripping is not an approved comparator.

# AGENT-MODEL.md — Agent & Engine Boundary Specification

> **Authority Level:** Mandatory Agent Interface Contract (Critical Foundation Document #3)  
> **Status:** Authoritative Core Specification  
> **Target Audience:** All AI Agent Integrations, Subsystem Developers, MCP Servers  

---

## 1. The Separation of Responsibilities

Rewrotron enforces a strict **bifurcation of responsibility**:
- The **Deterministic Engine** is the sole authority on state, scope, data persistence, hashing, test execution, gate enforcement, and comparison.
- The **AI Agent** is an advisory intelligence responsible for semantic interpretation, code synthesis, architecture design, and diagnostic reasoning.

```text
┌───────────────────────────────────────┬───────────────────────────────────────┐
│     DETERMINISTIC ENGINE POWERS       │         AI AGENT CAPABILITIES         │
├───────────────────────────────────────┼───────────────────────────────────────┤
│ • Mandatory Confirmation Gate Checks  │ • Semantic Business Rule Inference    │
│ • Relational State Persistence        │ • Target Architecture Design          │
│ • Experiment Hashing & Fingerprinting │ • Clean Code Synthesis (Laravel, Go)  │
│ • Reachability Graph Traversal        │ • Mismatch Diagnosis & Root Cause     │
│ • Runtime Trace Capture & Redaction   │ • Ambiguity Resolution Proposals      │
│ • Exact / Semantic Diff Comparisons   │ • Test Generation & Edge Case Defense │
│ • Mission Scheduling & Prioritization │ • Refactoring & Modernization Styling │
│ • Sandboxing & Security Enforcement   │                                       │
└───────────────────────────────────────┴───────────────────────────────────────┘
```

---

## 2. Invariant Boundaries ("What Agents May NEVER Decide")

1. **Agents CANNOT Bypass Confirmation Gates:** The engine enforces Gate 1 (Source Detection), Gate 2 (Target & Plan), and Gate 3 (Cutover). If an agent attempts to call `migration.start` or generate target code before Gate 1 and Gate 2 are recorded in `user_confirmations`, the engine returns `ERROR: USER_CONFIRMATION_REQUIRED`.
2. **Agents CANNOT Infer Confirmation from Silence:** Confirmation requires an explicit, authenticated user action (a CLI confirmation command or MCP confirmation tool call) that writes a persistent cryptographic record into SQLite.
3. **Agents CANNOT Alter Comparison Results:** An agent cannot declare a differential failure to be a "pass". The verifier emits `PASS`, `MISMATCH`, or `UNKNOWN`.
4. **Agents CANNOT Delete Uncertain Code:** An agent cannot mark code as `DEAD` unless the deterministic reachability graph and runtime trace logs prove zero references.
5. **Agents CANNOT Directly Mutate `project.db`:** Agents interact with project state exclusively via standardized MCP tools or CLI commands.

---

## 3. The 3 Engine-Enforced Confirmation Gates

```text
┌─────────────────────────────────────────────────────────────┐
│ GATE 1: SOURCE DETECTION CONFIRMATION                       │
│ Trigger: rewrotron detect (or init)                         │
│ Output: Language, framework, version, runtime, confidence.   │
│ Blocking Rule: Deep analysis blocked until user confirms.   │
│ Invalidation: If source repository files change,            │
│               configuration_hash invalidates Gate 1.        │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ GATE 2: TARGET SELECTION & MIGRATION PLAN CONFIRMATION      │
│ Trigger: rewrotron plan                                     │
│ Output: Target language, framework, architectural style,    │
│         migration units, dead code omission proofs.         │
│ Blocking Rule: Target code generation blocked until user    │
│                confirms plan.                               │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│ GATE 3: PRODUCTION CUTOVER CONFIRMATION                     │
│ Trigger: rewrotron cutover                                  │
│ Output: Differential verification audit report (>= 99%).    │
│ Blocking Rule: Deployment and production database mutations  │
│                strictly blocked without human sign-off.     │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. Agent Autonomous Loop: Next-Mission Flow

When an agent operates in autonomous mode, it executes this standard loop:

```text
1. Agent runs `rewrotron detect --json` (or calls `project.detect`)
2. Engine emits structured detection findings with confidence metrics.
3. Agent presents findings to human user in chat.
4. Human user confirms or provides overrides:
   `rewrotron confirm source --php 8.2` (or calls `project.confirm_source`)
5. Engine records confirmation and unlocks Gate 1.
6. Agent runs deep analysis (`rewrotron analyze`, `rewrotron observe`, `rewrotron discover`).
7. Agent runs `rewrotron plan --target go --style modular_monolith`.
8. Agent presents proposed migration plan to human user.
9. Human user confirms plan:
   `rewrotron confirm plan` (or calls `project.confirm_plan`)
10. Engine records confirmation and unlocks Gate 2.
11. Autonomous migration loop begins:
    a. Agent requests next mission: `research.get_next_mission`
    b. Agent generates target vertical slice code conforming to target adapter.
    c. Agent triggers differential verification: `verification.run_case`
    d. If PASS: Agent marks mission complete and requests next mission.
    e. If MISMATCH: Agent enters diagnostic repair loop.
```

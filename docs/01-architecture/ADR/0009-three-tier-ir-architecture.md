# ADR 0009: Three-Tier Intermediate Representation (Source IR, Application IR, Behavior IR)

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
Directly translating from a source framework AST (e.g. CodeIgniter 3) to a target framework (e.g. Laravel 13) couples the modernization engine to specific framework pairs ($N \times M$ translators). Furthermore, legacy syntax often obscures the true architectural structure of the application.

## Decision
Introduce a **Three-Tier Intermediate Representation (IR)** pipeline that decouples source understanding from target generation:

```text
Source Code
    ↓
1. SOURCE IR (What the code is syntactically: ASTs, symbols, files, calls, declared routes)
    ↓
2. APPLICATION IR (What the application is structurally: controllers, services, models, workflows)
    ↓
3. BEHAVIOR IR (What the outside world observes: inputs, business rules, state transitions, side effects, outputs)
    ↓
Target Adapters (Laravel, Go, Rust, Python, TypeScript)
```

## Alternatives Considered
- **Direct AST Translation:** Mapping PHP AST nodes directly to target framework AST nodes. Rejected because it couples frameworks ($N \times M$), carries forward legacy God classes, and prevents cross-language migration.
- **Single Flat Behavior Model:** Attempting to infer observable behavior directly from raw AST without modeling application architecture. Rejected because structural concepts (like controllers, services, models, and workflows) provide critical context for synthesizing clean target code.

## Consequences
- **Positive:**
  - $N$ source adapters + $M$ target adapters enable $N \times M$ migration combinations.
  - Adding a new target adapter (e.g. Go or Rust) immediately works for all supported source frameworks.
  - Clean separation of concerns: source adapters understand legacy quirks; target adapters generate clean modern idioms.
- **Negative:**
  - Requires maintaining transformations between Source IR, Application IR, and Behavior IR.

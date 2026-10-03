# LEARNING-SYSTEM.md — Long-Term Learning & Skill Evolution Specification

> **Authority Level:** Mandatory Learning Contract  
> **Status:** Authoritative Core Specification  
> **Subsystems:** `internal/learning`, `internal/knowledge`, Skill Manager  

---

## 1. Dual-Tier Learning Topology

Rewrotron prevents "hallucinated knowledge" by strictly separating project-specific learning from global cross-project knowledge:

```text
┌────────────────────────────────────────────────────────┐
│                   PROJECT LEARNING                     │
│  Scope: Single customer migration (.rewrotron/project.db)│
│  Entities: Observations, Mismatches, Repair Attempts,  │
│            Project Lessons.                            │
└───────────────────────────┬────────────────────────────┘
                            │ (Formal Promotion Gate)
                            ▼
┌────────────────────────────────────────────────────────┐
│                    GLOBAL LEARNING                     │
│  Scope: Multi-project reusable knowledge (~/.rewrotron/)│
│  Entities: Validated Patterns, Framework Idioms,       │
│            Global Skills, Stable Migration Rules.       │
└────────────────────────────────────────────────────────┘
```

---

## 2. Knowledge Evolution Hierarchy

Learning progresses through a disciplined lifecycle:

```text
1. OBSERVATION
   └─ Raw trace of runtime execution or mismatch diff.
        ↓
2. LESSON
   └─ Specific diagnostic conclusion learned from fixing a mismatch.
        ↓
3. PATTERN HYPOTHESIS
   └─ Proposed generalization of the lesson (e.g. "CI3 Cart stores serialized arrays in DB").
        ↓
4. EXPERIMENT
   └─ Empirical testing of the hypothesis across multiple test cases.
        ↓
5. VALIDATED PATTERN
   └─ Confirmed heuristic that passed evidence thresholds.
        ↓
6. SKILL PROPOSAL
   └─ Encapsulation of validated pattern into an updated Skill reference or script.
        ↓
7. SKILL VERSIONING (Experimental -> Stable -> Deprecated)
```

---

## 3. Evidence Thresholds for Promotion

### 3.1 Project Lesson -> Global Knowledge
A project lesson may be promoted to Global Knowledge **ONLY IF**:
1. **Multi-Project Occurrence:** Observed and successfully resolved across at least three independent fixture/application contexts, including negative cases. Multiple units in one app alone are not evidence of cross-project generality.
2. **Deterministic Verification:** The remediation pattern produced a 100% differential verification pass rate in all instances.
3. **No Regressions:** Caused zero side-effect regressions across the global benchmark corpus.
4. **Human / Architect Approval:** Explicit human approval of the sanitized export/pattern plus the automated regression gate. Automated validation alone does not authorize crossing project privacy boundaries.

### 3.2 Experimental Skill -> Stable Skill
A skill version progresses from `EXPERIMENTAL` to `STABLE` **ONLY IF**:
1. It has been executed in at least 20 repair missions with $\ge 90\%$ first-attempt success rate.
2. It has zero rollbacks in the experiment ledger.
3. Execution time and token overhead fall within established bounds.


## 4. Initial versus later implementation

Initial learning records project lessons and failed repair fingerprints only. Pattern generalization, global search and skill synthesis are later RWT-023/024 work. A proposed skill cannot execute arbitrary code through MCP. Generated scripts require explicit capability metadata, safe runner contracts, synthetic evaluation and versioned approval.

Promotion requires an opt-in reviewed manifest, customer-data removal, at least three independent contexts, 100% passing supported cases and no new required UNKNOWNs/regressions. First-attempt skill metrics are reported with attempt denominators; success rates cannot override verifier evidence. Stable skill thresholds remain minimum evidence requirements, not a substitute for security/privacy checks.

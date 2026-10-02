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
1. **Multi-Project Occurrence:** Observed and successfully resolved across $\ge 3$ distinct migration units or projects.
2. **Deterministic Verification:** The remediation pattern produced a 100% differential verification pass rate in all instances.
3. **No Regressions:** Caused zero side-effect regressions across the global benchmark corpus.
4. **Human / Architect Approval:** Flagged for sign-off or validated by an automated regression gate.

### 3.2 Experimental Skill -> Stable Skill
A skill version progresses from `EXPERIMENTAL` to `STABLE` **ONLY IF**:
1. It has been executed in at least 20 repair missions with $\ge 90\%$ first-attempt success rate.
2. It has zero rollbacks in the experiment ledger.
3. Execution time and token overhead fall within established bounds.

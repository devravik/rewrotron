# NON-GOALS.md — Explicit Out-of-Scope Boundaries

> **Version:** 1.0  
> **Status:** Binding Negative Constraints  

---

To prevent scope creep, design dilution, and catastrophic false assumptions, Rewrotron explicitly declares the following items as **NON-GOALS**:

## 1. General Anti-Patterns & Out-of-Scope Behaviors

- **NG1: Not a Generic AST Transpiler**  
  Rewrotron does not perform 1-to-1 AST conversions from PHP to PHP, Go, or Python. It does not attempt to map every line of legacy code to a corresponding line in the new codebase.

- **NG2: Not a Blind Framework Upgrade Wizard**  
  Rewrotron is not Rector or PHPStan-Upgrade. While upgrade tools alter code in-place within the same architecture, Rewrotron reconstructs applications from the ground up into clean target architectures.

- **NG3: No Autonomous Production Cutover**  
  Rewrotron will never automatically switch DNS, deploy directly to live production infrastructure, or execute unreviewed schema migrations on live production databases. The initial engine packages a Gate 3-approved runbook only; an actual production runner is deferred.

- **NG4: No Deletion of Uncertain or Dormant Code Without Approval**  
  Rewrotron never assumes code is dead merely because a short observation window did not trigger it. DORMANT or UNKNOWN is never automatically omitted; a visible scope exclusion needs Gate 2 and is not a DEAD proof. Source files are never deleted by migration.

- **NG5: No Invention of Undocumented Infrastructure**  
  Rewrotron will never invent message queues, external microservices, cloud cache clusters, or database engines that were not present in the legacy source or explicitly configured in `rewrotron.yaml`.

- **NG6: No Guarantees on Unobserved / Unspecified Behaviors**  
  Rewrotron cannot and does not guarantee behavioral equivalence for dynamic behaviors that were never observed, never tested, and not statically reachable. Uncovered spaces are explicitly flagged as `UNKNOWN`.

- **NG7: No Mandatory Cloud or SaaS Dependency**  
  Supported deterministic replay of acquired pinned bytes requires no internet; initial dependency/runtime acquisition may require network or supplied cache. Rewrotron core will never require an internet connection, a cloud account, a remote database, or an external LLM subscription to run deterministic analysis, observation, and verification.

- **NG8: Target Architecture Need Not Resemble Legacy Architecture**  
  Rewrotron does not preserve anti-patterns (such as 3,000-line God controllers or direct SQL queries in template views) for the sake of familiarity. Clean architecture is prioritized.

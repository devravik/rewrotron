# SKILL-EVOLUTION.md — Automated Skill Synthesis & Deprecation Rules

> **Version:** 1.0  
> **Status:** Learning Reference Guide  

---

## 1. Skill Lifecycle States

Every Skill in `skills/` or `~/.rewrotron/skills/` is versioned and tagged with one of four lifecycle states:

1. **`PROPOSED`**: Synthesized by an agent after resolving a repeated class of mismatches; awaiting automated test validation.
2. **`EXPERIMENTAL`**: Usable on non-critical migration units with strict rollback monitoring.
3. **`STABLE`**: Standard production skill recommended for all applicable units.
4. **`DEPRECATED`**: Superseded by a superior pattern or obsolete due to target framework changes.

---

## 2. Automated Deprecation Rule

A skill is automatically downgraded or deprecated if:
- Its success rate falls below 75% over the last 15 attempts.
- An update to the target framework (e.g. Laravel 13.x minor release) renders the generated code invalid or insecure.


## 3. Reproducibility and authority

Pin the skill version/content hash in each experiment. Deprecation creates a new registry event; historical skill content remains retrievable. Promotion and generated-script activation require the reviewed evidence/privacy process in LEARNING-SYSTEM. Missing usage samples are “insufficient evidence,” never a computed success rate of 100%.

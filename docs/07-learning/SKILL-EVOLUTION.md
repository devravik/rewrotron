# Skill revision lifecycle

[LEARNING-SYSTEM](LEARNING-SYSTEM.md) owns independence/quality/promotion/retraction; [SKILL-SPEC](../06-agent/SKILL-SPEC.md) owns orchestration interfaces. PROPOSED/EXPERIMENTAL/STABLE/DEPRECATED are registry lifecycle labels, not permission or validation evidence. Experimental activation requires reviewed safe runner capabilities and independent benchmarks; “non-critical” is not a safety exemption.

Immutable version/digest/applicability is pinned per consumer. Evidence degradation/security/mutation failure blocks further activation and creates retraction/successor records with transitive consumer revalidation. Low success rates may trigger review but cannot be the sole semantic judgment. Missing denominators are insufficient evidence. Global synthesis is deferred to M7; no current generated skills/tools are created by this revision.

## State transitions and evidence rules

The four skill lifecycle states correspond to SQLite schema constraints in [SQLITE-SCHEMA](../08-storage/SQLITE-SCHEMA.md) (`skill_versions.status`):

1. **PROPOSED:**
   * *Genesis:* Synthesized from one or more append-only `lessons` or proposed by a developer/agent.
   * *Package:* Staged under `skills/capabilities/<category>/<name>/` with `SKILL.md` frontmatter referencing source `lessons` and `experiments`.
   * *Persistence:* Staged row in `skill_versions` with status `PROPOSED`.
   * *Usability:* Draft and inspection only. Never automatically routed or executed in production migrations.

2. **EXPERIMENTAL:**
   * *Prerequisites:* Clean AST/syntax validation and successful execution in at least one isolated sandbox fixture or unit test with zero regressions on existing suites.
   * *Persistence:* Status updated to `EXPERIMENTAL` in `skill_versions`.
   * *Usability:* Usable only under explicit operator opt-in or heightened sandbox verification. Never permitted to waive or relax verifier checks.

3. **STABLE:**
   * *Prerequisites:* Promoted only after satisfying the complete [LEARNING-SYSTEM](LEARNING-SYSTEM.md) threshold:
     * Minimum three independent fixture/application contexts (independent author/template derivation).
     * 100% detection of designated negative semantic mutants.
     * Zero holdout regressions.
     * Zero newly introduced UNKNOWNs.
     * Explicit human operator promotion review.
   * *Persistence:* Status updated to `STABLE` in `skill_versions` and registered in `skills/manifest.json`.
   * *Usability:* Automatically discoverable and eligible for deterministic routing during standard execution.

4. **DEPRECATED:**
   * *Genesis:* Triggered by evidence degradation, security vulnerability, mutation failure, or supersession by a newer content-addressed revision.
   * *Retraction:* Transitive invalidation marks downstream consumers in [validity-graph](../../specs/project/validity-graph.schema.json) as `STALE`, requiring revalidation without erasing historical approvals. Disallowed for new work.

## Content-addressed immutability and provenance

Every skill version is content-addressed:

* The `document_ref` in `skill_versions` references an immutable artifact (`artifacts.id`) containing the `rwt-skill-package-1` SHA-256 digest defined in [SKILL-SPEC](../06-agent/SKILL-SPEC.md).
* Frontmatter provenance links the skill version directly to supporting `lessons` and `experiments` records (with SHA-256 context fingerprints).
* Past revisions in `skill_versions` are immutable; revisions append new rows with incremented semantic versions rather than modifying existing entries.

## Self-improvement boundary

The boundary between evolvable capabilities and human-controlled authority is strictly separated:

* **Evolvable via validated learning pipeline:**
  * Syntax transformation templates and AST mapping tables in `references/*.md`.
  * Troubleshooting references and framework edge-case workarounds.
  * Diagnostic verification recipes (sandbox inspection commands).
  * Trigger tags and keyword metadata in `skills/manifest.json`.
  * Candidate `PROPOSED` skill packages derived from append-only `lessons`.

* **Strictly human-controlled (engine invariants — never self-modifying):**
  * Constitutional authority and rules in [AGENTS.md](../../AGENTS.md).
  * System architecture and topic ownership in `docs/00-vision/` and `docs/01-architecture/`.
  * Security invariants, sandbox boundaries, and credential handling in `docs/09-security/`.
  * Machine contracts and JSON schemas in `specs/`.
  * Confirmation protocol and Gate 1, Gate 2, Gate 3 approvals in [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md).
  * Verifier verdicts (PASS, MISMATCH, UNKNOWN) and differential testing semantics in [DIFFERENTIAL-TESTING](../05-verification/DIFFERENTIAL-TESTING.md).
  * SQLite database schemas, migrations, and append-only database triggers in [SQLITE-SCHEMA](../08-storage/SQLITE-SCHEMA.md).
  * Milestone roadmap sequencing in [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md).
  * Promotion of any skill to `STABLE` status.

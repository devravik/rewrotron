# ADR 0011: Falsification before broad implementation

**Status:** Accepted specification decision; implementation pending. Revised 2026-10-03. Detailed contract owner: [DEVELOPMENT-PLAN](../../11-roadmap/DEVELOPMENT-PLAN.md). This ADR records the decision, not a duplicate field specification.

## Context

Schema validation and read-only fixtures do not validate reconstruction.

## Problem

Unverified assumptions can produce false acceptance or nonreproducible work; the owning contract must distinguish evidence and current authority.

## Decision

M0 durable foundation then M1 stateful wrong-target rejection before broad adapters/autonomy; maturity evidence vocabulary; unsupported limits explicit. Assessment is a stopping checkpoint in shared analysis, with independent report completion and no parallel engine; add X41–X55 across existing milestones without postponing M1 falsification.

## Alternatives considered

Broad adapters first; read-only happy-path milestone; early stateful mutation falsification.

## Why rejected

Happy paths cannot detect a plausible wrong reconstruction; broad adapters compound an untested oracle.

## Consequences

More explicit immutable refs and required negative tests; unsupported inputs block rather than fabricate support. Old evidence remains readable and historical. No runtime maturity is asserted.

## Machine-contract impact

Relevant v2 schemas/catalogs and migration 0002 represent subjects, exact resolution, semantics/scenarios, scope, causality, decisions/process/resources or current-state windows as applicable. Existing v1 artifact shapes are read-only history; envelope/protocol versions remain separately pinned. Topic owner defines precise fields and service invariants. New assessment report/view/selection/planning-basis artifacts use existing immutable storage. Plan envelope 2.1 adds checkpoint refs, hashed through unchanged PLAN subject v2 decision_refs; old subjects/vectors remain reconstructable, not current executable authorization.

## Migration/compatibility implications

Do not rewrite migration 0001 or existing approvals. Apply checksummed 0002 after ownership/duplicate preflight; require successor approval for changed execution choices. Legacy parsing never silently creates current acceptance. No deployed engine predecessor is claimed. Old plan wrappers remain historical-readable; current execution requires a checkpoint-linked successor plan and its explicit Gate 2 approval. A separate assessment pipeline and implicit continuation were rejected because they duplicate analysis and conflate information with authorization.

## Acceptance experiment / falsification test

M1 17 premise checks plus applicable X41–X55 shared checkpoint/non-authorization/reuse tests. Record actual command/profile/versions and expected semantic failure before claiming validation; schema-valid positive records alone do not establish this decision works.

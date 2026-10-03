# Confirmation protocol and canonical approval subjects

**Owner:** All gate identities, preimages, human authorization and invalidation. Current subjects use schema version 2.0 and [approval-subject.schema.json](../../specs/project/approval-subject.schema.json). Old v1 artifacts remain readable historical evidence, never inferred current approvals.

| Gate | Exact subject | Authorizes |
| --- | --- | --- |
| GATE_1_SOURCE_DETECTION | SOURCE subject/configuration_hash | Analysis, isolated observation, modeling and planning against sealed source |
| GATE_2_PLAN_CONFIRMED | PLAN subject/plan_hash | Generation, controlled publication, repair and sandbox verification |
| GATE_3_CUTOVER | RELEASE subject/release_hash | Approved release package/runbook; no initial production execution |

## Hash preimage v2

Subjects are immutable registered artifacts. SOURCE binds sealed snapshot, detection, separate human assertions and exact source/probe bindings. PLAN binds configuration_hash, snapshot, Behavior IR, corpus, policy, user intent, exact resolution, target model, immutable decision refs, validity graph, all adapter/runner/comparator bindings, consumed skill/tool/knowledge revisions, ordered units, dependency edges, omission proofs, exclusions and blockers. RELEASE binds configuration/plan, target tree, whole-project report, benchmark, runbook and rollback rehearsal. Schema-required null is distinct from absent; optional absent fields have no implicit hash default. All relevant choices must be explicit before hashing. Proposal refs are not confirmation refs: normalization and dependency decisions are approved transitively by the plan, avoiding circular hashes.

`subject_hash = SHA256(rwt-cjson-1(subject))`, represented as 64 lowercase hex. Neither subject_hash, timestamps, actor, mutable progress, approval receipt nor registry lookups are part of this preimage. Nested refs hash exact artifact bytes; the same hash under another project does not confer ownership. Execution requires registered same-project/transitively authorized refs, size/hash/schema verification and available pinned implementations. A digest alone cannot prove bytes exist or are trusted.

## rwt-cjson-1 serialization

UTF-8 JSON, no BOM or trailing newline, no whitespace outside strings. Recursively order object keys by their unnormalized UTF-8 byte sequence. Preserve Unicode code points; reject invalid surrogate code points, duplicate keys and non-finite/floating numbers. Values are objects, arrays, strings, signed 64-bit integers, booleans or null. Integers use shortest decimal, no leading zeros or negative zero. Strings escape quote/backslash and U+0000–001F using JSON short escapes for backspace/tab/newline/formfeed/carriage return and lowercase `\u00xx` otherwise; do not escape slash or non-ASCII. Arrays retain schema-defined semantic order.

Set-valued refs/IDs sort lexicographically by UTF-8 bytes and reject duplicates before serialization. adapter_bindings sort `(role,adapter_id,version,manifest_ref,implementation_ref)`. Dependency edges sort `(prerequisite_id,dependent_id)`. ordered_units use deterministic topological order: among ready units highest priority then unit_id bytes; cycles block. Unit case_ids, contract_refs and capabilities are sets. Omission/exclusion/decision/revision refs are sets. A supplied noncanonical set/order is rejected, not silently reinterpreted during approval.

[approval-hash-vectors.json](../../specs/examples/approval-hash-vectors.json) contains preimages, canonical bytes and expected digests. Structural Python checks are not proof the eventual Go canonicalizer passes. Implementations must reproduce vectors, shuffled object key cases, non-ASCII/escape/null cases and rejected duplicate/float/invalid-order cases.

## Reconstruction and upgrades

Load the original subject bytes and its complete immutable reference closure from SQLite/artifacts; validate using its retained schema/canonicalizer version. Never rebuild from the current adapter registry, latest locks or generated YAML. Store approved payload ref, schema/canonicalization version, hash and receipts. Upgrading engine/adapters does not rewrite historical subjects. A missing old executable/schema prevents executing/rechecking that subject; it does not alter what was approved. A successor execution plan names new versions and requires Gate 2. No migration may manufacture approvals for upgraded subjects.

## Human channel and tokens

Confirmation is an append-only user_confirmations record, not inferred silence/conversation or a YAML flag. Local CLI `confirm source/plan/cutover` requires explicit exact subject hash and trusted human context. `approval issue` locally issues a short-lived, one-use token bound to project, gate, exact subject, human context and expiration; MCP confirmation tools consume it transactionally. Agents cannot issue tokens through MCP, supply their own human identity or turn DECIDED/APPROVAL_RECEIVED events into receipts. Same-account CLI use remains a trust boundary, not OS isolation. Dry-run never creates token/receipt.

Idempotent retries return the existing receipt; different payload under the same key conflicts. Confirmation, reviewed payload, receipt/event and current validity change commit atomically. Validate current parents, availability, feasibility and subjects immediately before commit. Explicit overrides are immutable human assertions alongside measured facts; contradictions with doctor measurements remain findings, not falsified detection.

## Invalidation

Live customer drift prompts a successor coherent snapshot and Gate 1; previously sealed snapshot/history remain intact. Target intent, dependency strategy, corpus, comparator, normalization, lowering/ownership choice or consumed execution revision changes require successor PLAN/Gate 2 and invalidate Gate 3. Implementation-only code changes within approved choices retain Gate 2 but invalidate results/release. [DOMAIN-MODEL](DOMAIN-MODEL.md) owns typed selective reuse; unknown impact invalidates conservatively. Progress/heartbeats never change an approval hash. Gate 3 requires current whole-project PASS, no in-scope unresolved blocking UNKNOWN/MISMATCH, required accepted performance and rollback/session evidence. Percent progress is never readiness.

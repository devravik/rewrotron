# Security and data boundary

**Owner:** Threat model, mediated permissions, private data and export/egress limits. Local-first does not mean no data can leave the machine.

## Three trust levels

1. Mediated engine enforcement: all CLI/MCP/Skill operations validate schemas, ownership, paths, gates, capabilities and safe runners; no raw SQL or unrestricted shell ports. This is enforceable within engine services.
2. Trusted same-account local operation: users/agents with filesystem/shell access can edit DBs, read private files or call human CLI outside those services. Rewrotron cannot claim protection from a hostile same-account agent. Tokens prevent mediated agent impersonation, not local root/owner compromise.
3. Hostile-agent isolation: separate OS identities/process sandboxes/egress policy would be required; deferred capability, not an implemented guarantee.

Runtime source/target are untrusted programs and require the declared sandbox profile, independent synthetic/sanitized DBs and denied production endpoints/mounts. Container support is a prerequisite, not proof of hostile-code isolation. Missing safe profile produces a blocker. Gate 3 initially approves a release package, not permission for a production runner.

## Private source versus evidence

Sealed source can contain hardcoded credentials and must remain byte-exact in a private owner-only vault. Agent-readable evidence is separately sanitized: source snippets, ASTs, diagnostics, logs, traces, DOM/screenshots, dependency archives and generated code all undergo secret/PII checks before public registration/display/export. If faithful evidence cannot be sanitized, retain private ref and report unavailable disclosure rather than leaking or claiming a transformed source is equivalent. Private identity manifests stay private; exported correlations use workspace-keyed digests where low-entropy values could leak. Keys never enter public/global artifacts.

Redaction precedes agent-readable staging and includes failure paths. Secret canaries test each output/artifact/client/export path. Context export is opt-in with an exact disclosed sanitized manifest; global promotion requires separate reviewed export. An external agent/client may send anything it can access to its provider; core has no provider SDK/telemetry by default and cannot control that client's egress. Report this trust boundary plainly.

## Acquisition and execution

Package acquisition uses reviewed registries/provenance/locks/digests/license inventory; hooks/scripts/native build require controlled runner capability and disabled egress except explicit acquisition policy. Version labels do not authenticate modified vendor bytes. Archives reject traversal, symlink escapes, device files and bombs. Descriptor-relative path operations plus ownership/hash checks must address symlink-swap races; initial unsupported platform protections block write/acquisition rather than advertise OS isolation.

Use fixed runner executable/typed args, cancellation/resource/time/output quotas; never concatenate application data into shell. Source text is untrusted data, not instructions. Production credentials/payment/SMS/email/cloud endpoints are prohibited in replay. Dangerous real effects cannot be approved merely by a normal gate; safe sink or unsupported scope is required. [Compatibility](../05-verification/COMPATIBILITY-SPEC.md) and [SANDBOX](SANDBOX.md) define effect/isolation obligations.

All client integrations are UNVERIFIED pending tests. Cross-project artifact/operation references are rejected both by service validation and storage ownership constraints where represented. Backups preserve protected evidence and private-key policy; immutable history does not license exporting secrets.

## Credential-bearing source admission

Private preservation does not authorize execution. Before Gate 1 runtime admission, statically inspect approved source/config/dependency scope and operator credential inventory without executing it. Unknown credential disposition is unsafe, not evidence of absence. execution-admission 1.0 yields ADMITTED_ORIGINAL only with synthetic/absent or evidenced revoked credentials; ADMITTED_VARIANT with a separately reviewed non-production execution variant; otherwise REFUSED or STATIC_ONLY. Denied network alone never admits live secrets into a container. Assessment can complete with runtime blocked.

execution-variant 1.0 records original manifest/content identity, variant manifest/content identity, exact transform and per-path before/after refs, credential policy, review evidence and behavioral limitations. Only approved synthetic/revoked replacements may execute. It is a separate immutable private artifact, not a sanitized overwrite of original source. Characterize the variant under its own witness/profile; assessment explicitly says original runtime behavior is unobserved where the replacement matters. Verification supports the approved variant's bounded obligations only; independent evidence is required for any original-source claim. Gate 1 reviews the variant/admission closure; Gate 2 must carry corresponding scope limitations. No silent substitution or automatic claim of equivalence.

## Capture, parser and vulnerability admission

Capture authority derives from a qualified channel controlled by the harness/service boundary, correlated to actual process/store state. Source-emitted log/assertion text is application output, not trusted principal/write/completeness evidence. Read-only capture credentials and immutable result registration prevent model/checker writes; service-process egress applies to DB, workers, queues, native helpers and sinks too. Same-account hostile access remains outside this guarantee.

Before parser/decompressor allocation, apply the input profile's budgets. Initial reference defaults: 16 MiB per source file, 256 MiB aggregate expanded archive, 100,000 entries, 1,000,000 AST nodes per file, depth 256, 30 seconds/file and 512 MiB parser-worker memory. The worker is a cancellable resource-limited process; source parsing never executes source. Crossing any bound leaves affected parsing incomplete/UNKNOWN and prevents DEAD; changing limits is a new profile, not silent truncation. Limits are specified defaults, not measured capacity. Reject archive traversal/device/symlink escapes before extraction.

An observed legacy vulnerability produces an attributed security finding and immutable plan decision: PRESERVE_WITH_REVIEWED_RISK, INTENTIONAL_CHANGE, EXCLUDE_SCOPE or BLOCK_RELEASE. Preserve/alter is never inferred by an agent. An intentional change has new executable obligations and visible difference report; strict parity does not certify security. Engine safety rules (including irreversible replay/credential restrictions) cannot be waived as compatibility. Secret canaries cover ASTs, diagnostics, private hashes, contexts, emergency journals and exports; derived sensitivity propagates. Unknown redaction/capture blocks disclosure rather than downgrading evidence requirements.

Admission records preserve original credential_status and separate variant_credential_status. ADMITTED_VARIANT requires the latter SYNTHETIC_OR_ABSENT or REVOKED with review/control evidence; unknown/live replacement credentials cannot be admitted. Revocation evidence must apply to the actual embedded credential without disclosing it.

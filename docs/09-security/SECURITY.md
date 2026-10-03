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

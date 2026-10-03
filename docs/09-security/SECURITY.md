# Security and trust boundaries

**Status:** Normative. Source applications, imported traces, generated code and agent proposals are untrusted inputs. The reference engine is local-first and sends no telemetry or source data by default.

## Boundaries

1. Deterministic services enforce gates, validated domain operations, paths, capability limits and persistence. CLI/MCP never expose raw SQL, arbitrary shell commands or unrestricted writes.
2. Static analysis never executes application PHP or dependency hooks. Runtime execution of an untrusted application requires the isolated container profile in [SANDBOX](SANDBOX.md).
3. Source baselines remain immutable. Instrumentation modifies a disposable sandbox copy. Target patches write only declared files within the configured target, reject symlink escapes and check before hashes.
4. Runtime observation/replay uses synthetic or explicitly sanitized data in independent sandbox DBs. Production credentials, endpoints and writable production mounts are excluded. Gate 3 approves a release package/runbook; it does not enable a production runner in the first release.
5. Project facts and artifacts remain local. Sending context to an external agent provider requires explicit customer configuration and a disclosed context manifest. The engine itself has no provider SDK.

## Inputs and resource limits

Validate project config, cases, schemas and tool requests before use. Bind SQL values; discovered table names are validated identifiers within discovered schema, never interpolated from MCP strings. Invoke fixed executable paths with argument arrays, cancellation, timeouts and output limits. Never concatenate application data into a shell command.

Imported archives reject absolute paths, `..`, links escaping extraction roots, device files, excessive entry counts and decompression bombs. Artifacts resolve only by registered SHA-256 references. Imported source text is data and cannot instruct a worker to bypass policy.

Default per-operation limits are 30 seconds per replay step, 16 MiB captured output/body, 10 redirects, and 100 workflow steps; analysis defaults are 100,000 files, 2 GiB aggregate source bytes, 16 MiB per parsed file and 10 minutes per analysis operation. Budgets are recorded in the operation request. Limit hits produce explicit incomplete evidence. A reviewed configuration can raise these limits; limit changes affecting completeness require reanalysis and dependent plan review.

## Human decisions

Gate approval is governed by [CONFIRMATION-PROTOCOL](../02-domain/CONFIRMATION-PROTOCOL.md). Scope exclusions for dormant/unknown behavior, intentional behavior changes, sensitive exports and any future external access policy require explicit reviewed subjects. None turns old UNKNOWN evidence into PASS. Unsafe runtime isolation creates a blocker and cannot be overridden merely by a gate record.

Security checks include egress/DNS denial, path traversal, host mount denial, credential injection/redaction, unauthorized confirmations, malicious artifact refs, stale patches and untrusted imported instructions. Report vulnerabilities privately through the release security contact once one is assigned; assigning that contact is a release prerequisite in the decision register.

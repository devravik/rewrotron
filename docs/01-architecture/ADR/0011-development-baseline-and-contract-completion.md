# ADR 0011: Development baseline and complete engine contracts

- **Status:** Accepted specification decision; implementation unstarted
- **Date:** 2026-10-03
- **Basis:** Documentation review requested by the repository owner
- **Related:** ADRs 0001–0010 remain binding; this ADR clarifies their implementation boundaries.

## Context

The repository has documentation and permissive draft schemas but no runtime. Conflicting state names, advertised adapter maturity, hash definitions, CLI commands, incomplete persistence, and blanket normalization rules make independent implementation unsafe. A runtime must distinguish configuration intent, immutable human approval, evidence, and mutable progress.

## Decisions

1. Use the canonical lifecycle and hash subjects in PROJECT-LIFECYCLE and CONFIRMATION-PROTOCOL. `init` may create the minimal workspace needed to persist detection and confirmations; it never performs deep analysis or auto-confirms. Remove `--auto-confirm`.
2. Framework code lives in `adapters/source/php/{codeigniter3,generic_php}` and `adapters/target/php/laravel13`. Neutral interfaces and records live in `internal/domain`. `internal/runtime` is the composition package importing concrete adapters/storage and injecting them into neutral services. The constitution's diagram describes orchestration dependency direction; domain packages do not import concrete framework packages. External adapters initially require building a new binary; a stable out-of-process plugin ABI is deferred.
3. Use `database/sql` with `modernc.org/sqlite`, a standard-library CLI parser, `github.com/VKCOM/php-parser` behind the PHP adapter, and draft 2020-12 JSON Schema. Pin compatible dependency versions and a supported Go toolchain in task RWT-001; never infer toolchain versions from document dates. The parser's documented syntax range is PHP 5, 7, and 8.0–8.2; unsupported syntax yields an explicit incomplete analysis result.
4. Use MCP protocol revision `2025-11-25` over stdio for the first release. Streamable HTTP is a later separately reviewed transport. Legacy standalone SSE is not the implementation baseline. No sampling or provider SDK is necessary.
5. Initial reference profiles are CI3 3.1.13, a fixture-specific pinned source PHP runtime, Laravel 13 with PHP 8.3+, and MySQL sandbox databases with preserved schemas. Exact patch versions, image digests, and Composer locks are captured by the fixture task before execution. Generic PHP supports inventory and unresolved findings initially; generation capability is explicitly limited. All adapters begin `EXPERIMENTAL` when implemented; none is currently tested.
6. The deterministic verdict set is `PASS`, `MISMATCH`, `UNKNOWN`. `EXPECTED_CHANGE` is a human-approved contract exception, recorded separately; historical results remain immutable and verification reruns against a new policy hash. Strict compatibility reports show these exceptions.
7. Compare committed logical DB state, ordered effects, and session semantics. Keep identifiers exact unless an approved bijective identity rule preserves every public reference. Normalization has explicit scopes, evidence, validators, and cross-field relationships; no universal HTML whitespace removal or token masking.
8. Project SQLite is authoritative. Immutable artifact writes precede transactional metadata/events; exported JSONL is recoverable output, not a second atomic ledger. Append-only events and confirmations reject SQL updates/deletes. Baseline records are never overwritten on drift.
9. Mission leases use transactional compare-and-swap with fencing tokens. Target generation uses staged patches and before hashes; source files are immutable and probes only modify sandbox copies. Local proxy variables are not a security boundary: container execution is required for untrusted applications in the reference release.
10. Initial cutover support produces a verified package and runbook after Gate 3. It does not implement live infrastructure deployment. Global knowledge export is opt-in and requires explicit human approval plus privacy and corpus checks.

## Consequences

The engine can be implemented in dependency order with contract fixtures before framework generation. Unsupported behavior becomes a recorded blocker; finite differential cases provide evidence for a declared scope, not proof of all possible application executions. Customer choices remain explicit inputs and cannot be manufactured by the scheduler.

## Sources checked on 2026-10-03

- [Laravel release policy](https://laravel.com/framework/docs/releases): Laravel 13 requires PHP 8.3.
- [MCP 2025-11-25 transports](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports): stdio and Streamable HTTP; the latter replaces legacy HTTP+SSE.
- [PHP parser project](https://github.com/VKCOM/php-parser): Go parser and documented syntax support.

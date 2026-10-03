# Release and distribution

Use semantic versioning for the engine and explicit versions for machine contracts, schema migrations, adapters, probes and skills. Breaking public shapes/required fields or incompatible persisted data require a documented upgrade path and appropriate major version after the first public release. No existing deployed API is claimed in this specification-only repository.

## Release evidence

Require supported-profile tests, `go test -race ./...`, lint, schema/CLI/MCP contracts, previous-version DB upgrades, fresh fixture migrations, security/redaction/egress checks, restart/recovery, independent target builds, license/SBOM review, signed checksums and documentation parity. Pin tool/runtime/dependency versions and image digests; publish the tested adapter capability/maturity matrix and known limitations.

Target platforms are Linux amd64/arm64, macOS amd64/arm64 and Windows amd64. The first planned runtime profile is Linux amd64; it is SPECIFIED, not runtime validated. Cross-compilation is insufficient to claim each OS's path/permissions/process behavior works; native smoke tests and driver packaging are release tasks. Container runtime availability is separate from CLI binary portability.

Deliver CLI and MCP binaries, checksum/signature manifest, Skill bundle, licenses/notices and upgrade/recovery instructions. Confirm static linking and race/driver behavior with the selected dependencies rather than promising it in advance. Assign release signing custody and a security contact before publication; these operational facts are explicitly tracked in the decision register.

Production rollback must rehearse target writes/session/external receipt reconciliation; preserved schema is insufficient. Version/driver/WAL, native platforms and client matrix require actual evidence before RELEASE_VALIDATED.

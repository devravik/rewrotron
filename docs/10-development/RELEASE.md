# Release and distribution

Use semantic versioning for the engine and explicit versions for machine contracts, schema migrations, adapters, probes and skills. Breaking public shapes/required fields or incompatible persisted data require a documented upgrade path and appropriate major version after the first public release. No existing deployed API is claimed in this specification-only repository.

## Release evidence

Require supported-profile tests, `go test -race ./...`, lint, schema/CLI/MCP contracts, previous-version DB upgrades, fresh fixture migrations, security/redaction/egress checks, restart/recovery, independent target builds, license/SBOM review, signed checksums and documentation parity. Pin tool/runtime/dependency versions and image digests; publish the tested adapter capability/maturity matrix and known limitations.

Target platforms are Linux amd64/arm64, macOS amd64/arm64 and Windows amd64. The first planned runtime profile is Linux amd64; it is SPECIFIED, not runtime validated. Cross-compilation is insufficient to claim each OS's path/permissions/process behavior works; native smoke tests and driver packaging are release tasks. Container runtime availability is separate from CLI binary portability.

Deliver CLI and MCP binaries, checksum/signature manifest, Skill bundle, licenses/notices and upgrade/recovery instructions. Confirm static linking and race/driver behavior with the selected dependencies rather than promising it in advance. Assign release signing custody and a security contact before publication; these operational facts are explicitly tracked in the decision register.

Production rollback must rehearse target writes/session/external receipt reconciliation; preserved schema is insufficient. Version/driver/WAL, native platforms and client matrix require actual evidence before RELEASE_VALIDATED.


## Installation contract (SPECIFIED, not an available release)

This document owns installation/distribution behavior. The [installation and first-assessment guide](INSTALLATION.md) is a user-facing procedure that references existing CLI, Skill, security and approval owners. Product installation does not initialize a customer project, acquire its runtime, issue approvals or execute a migration. Product distribution is distinct from the customer Gate 3 release package/runbook.

### Release contents and acquisition

The first qualified installation profile is Linux amd64. Publish a versioned `rewrotron-<version>-linux-amd64.tar.gz` with one top-level directory of the same name, containing:

- `bin/rewrotron` and `bin/rewrotron-mcp`.
- `share/rewrotron/` with required schemas, catalogs and runtime resources; binaries resolve resources relative to the versioned installation, not the user's working directory or a source checkout.
- `skills/rewrotron/`, the engine operator package defined in [SKILL-SPEC](../06-agent/SKILL-SPEC.md), with its complete referenced resources and recorded revision/content hash. The repository implementation skill is not a substitute for this operator skill.
- `docs/` containing the installation guide, supported-profile/client matrix, upgrade/recovery instructions and release notes; `LICENSE`, notices and SBOM.

The release record lists exact engine/build/contract identities and all downloadable filenames, sizes and SHA-256 checksums. Publish its signature plus verification instructions and a trusted signing identity obtained independently of the downloaded archive. A checksum from the same untrusted source alone does not authenticate a release. Public release hosting endpoint, signature format/verifier and signing custody are operational inputs still required before M6 publication; neither an installer URL nor a trusted signing key is implied by this specification. Publish only tested platform packages; additional OS/architecture archive formats and layouts are defined and qualified when admitted.

Provide a downloadable installer alongside the manual route. For Linux it is a reviewable shell script using explicitly documented standard tools; users download and verify it before execution. Do not advertise an unverified download-and-execute command. Authenticate the installer and its selected package against the published release verification policy. Runtime use of installed binaries must not require Go, Python, a source checkout or a separate SQLite server. Disclose any installer/signature-verifier dependencies and native shared-library prerequisites; bundled SQLite does not imply unverified static linking.

### Manual and installer routes

Both routes implement the same sequence: select exact engine version and supported OS/architecture → obtain/authenticate release metadata → verify archive bytes → validate archive entries → extract into a private staging directory → check complete resources and executable identities → publish installation → report how to invoke it. Refuse unsupported profiles, missing signature/digest, changed bytes, absolute/traversing archive paths, escaping links and unexpected existing destinations before activation. Never silently select another platform/version, overwrite unrelated files or follow an installation path into customer source/target/workspace roots. Failure before activation preserves the existing installation and removes only owned staging files; uncertain cleanup is reported.

Default Linux locations are `$HOME/.local/share/rewrotron/releases/<version>/` and entrypoints `$HOME/.local/bin/rewrotron` and `$HOME/.local/bin/rewrotron-mcp`. Both entrypoints resolve through one managed `current` pointer so activation selects a single complete release. Reject unmanaged existing entrypoints/pointers instead of overwriting them. Explicit operator-selected installation/entrypoint directories are supported under the same ownership/containment checks. Install without administrator privileges; do not edit shell startup files, agent configuration or PATH automatically. Print the actual executable paths and shell-appropriate PATH instructions, then verify version/help from an unrelated working directory with no source checkout. Manual installation uses the same versioned layout and pointer rules.

The installer offers explicit install/version selection and uninstall actions documented with the shipped installer; these are not new engine CLI commands. Omitted version selects the exact version bound by the authenticated installer/release record, never an unverified floating `latest`. Unsupported acquisition or network failure offers the manual/local-archive route rather than disabling verification. Offline installation requires previously acquired metadata, signatures, verifier trust material and archive bytes. Homebrew/other package managers, automatic updates and broad platforms are later conveniences, not initial release requirements.

### Upgrade, rollback and uninstall

Upgrades stage and verify a successor release without touching customer workspaces. Refuse activation while owned foreground work or a supervisor is active; the operator first stops/drains it according to [RECOVERY](../08-storage/RECOVERY.md). Atomically replace the managed current pointer only after validation, retaining previous releases required by pinned approvals/execution evidence. Installing a binary never migrates project SQLite state automatically. Opening a project follows the existing checksummed migration/preflight/backup rules; an incompatible newer project schema is refused by an older binary. Switching the binary back does not reverse DB migrations or confer current approvals.

Uninstall removes only the selected installation's managed entrypoints/current pointer and explicitly selected owned release files, after the same process checks. Customer source/target repositories, migration workspaces, SQLite/artifacts, credentials, caches and backups are preserved by default. Never garbage-collect approval/release closure bytes through uninstall. Retained installation bytes needed by a project must remain retrievable under [ARTIFACT-STORE](../08-storage/ARTIFACT-STORE.md); refuse removal when this cannot be established and report the affected retention obligation. Removing client configuration or a copied Skill requires a separate explicit action and matching ownership/revision checks; uninstall must not erase other client settings.

### Packaging and publication acceptance

M6 builds from a named Git revision with pinned tooling; produces platform archives, installer, resource/Skill bundle, release metadata/signatures/SBOM; runs X36 against those exact archives; and publishes verified assets plus the supported capability/client matrix. Record artifact digests, native environment, build commands, signer and test evidence. Rebuilding the named revision under the same pinned build profile must reproduce unsigned payload bytes; signatures/timestamps are recorded separately. Missing reproducibility or authenticity evidence blocks publication. Choice of build automation is an implementation detail, not a new workflow.

Installation success is separate from migration readiness. [CLI-SPEC](../06-agent/CLI-SPEC.md) owns `doctor` before/after workspace initialization; unavailable optional MCP/Skill integration never prevents CLI-only use. Legacy container images, databases, extensions and safe fixture inputs are checked for the selected project, not all installed on the user's machine as product-install prerequisites. Actual client setup and the first-assessment walkthrough are qualified in M5; final package installation is qualified by X36 in M6. No release capability is established by this documentation.

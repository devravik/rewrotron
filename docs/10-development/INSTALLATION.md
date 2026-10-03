# Installation and your first assessment

**Status: SPECIFIED user procedure, not an available installation.** No migration CLI/MCP release exists yet. This guide describes the release experience to qualify in M5/M6; do not run the planned commands until using a released, tested package. [RELEASE](RELEASE.md) owns package/install rules; [CLI-SPEC](../06-agent/CLI-SPEC.md) owns command syntax. This guide creates no new lifecycle, approval or agent workflow.

## 1. Install the tool

The first supported release will target Linux amd64. Use the published release page to choose an exact version and check its tested profile/client matrix. A portable CLI does not by itself establish that your source runtime is supported.

Choose either route:

- **Manual:** Download the archive plus release metadata/signature. Follow the release's verification instructions with its independently trusted signing identity. Verify before extracting; reject unsafe archive paths. Install the complete archive into the versioned user-local layout and create only the managed entrypoints described in RELEASE.
- **Installer:** Download and authenticate the release's installer before executing it. Select the exact version or use its authenticated bound version. It performs the same verification, safe extraction and activation; it does not initialize a project or approve migration.

The default Linux layout is:

```text
~/.local/share/rewrotron/releases/<version>/
    bin/rewrotron
    bin/rewrotron-mcp
    share/rewrotron/     # Required schemas/catalogs/resources
    skills/rewrotron/   # Engine operator Skill and its references
    docs/
~/.local/share/rewrotron/current  # Managed selected-release pointer
~/.local/bin/rewrotron           # Managed entrypoints through current
~/.local/bin/rewrotron-mcp
```

Add the reported entrypoint directory to PATH yourself, or use the full executable path. The installer must not edit your shell startup files. Check `rewrotron --version`, `rewrotron --help`, and `rewrotron doctor` from a directory outside the package/source checkout. Doctor before a workspace exists may return `state: null`; missing project/runtime information is not evidence that binary installation failed.

The release download URL, signature-verifier procedure and installer invocation are intentionally unavailable until published and tested. The shipped copy of this guide must replace those operational gaps with exact verified instructions before release. Go, Python and a SQLite server are not user runtime prerequisites. Any native libraries or installer verification utilities must be disclosed by that release.

## 2. Keep tool, source, target and state separate

For example:

```text
/work/customer-old/        # Existing source repository
/work/customer-new/        # New repository where target code will live
/work/customer-migration/  # Configuration, SQLite metadata and evidence
```

The tool is installed separately. Source/target roots must not overlap. Keep migration evidence private and durable; do not blindly commit credentials, raw traces or snapshots to Git. [PROJECT-INPUTS](PROJECT-INPUTS.md) owns the input checklist and import layout; [SECURITY](../09-security/SECURITY.md) owns data/export boundaries.

## 3. Initialize and review the source

Planned released commands:

```bash
rewrotron init /work/customer-old --project /work/customer-migration --json
rewrotron doctor --project /work/customer-migration --json
```

Inspect detected facts, uncertainties and the coherent snapshot. Prepare/review the source subject and explicitly confirm Gate 1 through the local human CLI using the returned detection ID, manifest hash and subject hash. Follow [CLI-SPEC's lifecycle example](../06-agent/CLI-SPEC.md#example-lifecycle-planned-commands); do not copy example hashes or allow an agent to fabricate approval.

Describe user-owned intent in `/work/customer-migration/rewrotron.yaml`, including `/work/customer-new` as target path. Use the versioned configuration contract in [TARGET-SPECIFICATION](../04-migration/TARGET-SPECIFICATION.md); a recommendation cannot replace your runtime choice. Target intent may be incomplete during assessment, but must satisfy selected-scope planning requirements before an executable plan.

Provide critical workflows, operational roots (including cron/jobs), constraints, business requirements and available dependency information. For runtime observation supply safe synthetic identities/data, disposable services and supported sink profiles. Missing infrastructure can yield a static/available-evidence assessment with explicit gaps; it does not permit production credentials or host fallback. Doctor checks project-specific prerequisites rather than installing every possible runtime.

## 4. Choose CLI-only or connect an agent

CLI-only use requires no agent subscription. For a compatible coding agent, either allow invocation of the installed CLI or configure its stdio MCP server using the installed `rewrotron-mcp` path. Use only release-tested client instructions listing client version, exact executable/arguments, resource paths, workspace access and supported protocol. Paths containing spaces must be represented correctly by that client's configuration format. Do not advertise a client as supported without those tests.

The optional operator Skill comes from the installed `skills/rewrotron/` package, defined by [SKILL-SPEC](../06-agent/SKILL-SPEC.md). Copy/link its complete resources into a tested client's documented Skill discovery location, recording the consumed revision/hash; do not overwrite another skill or track a mutable live version silently. Without Skill support use CLI/MCP directly. The repository `rewrotron-implementation` skill teaches contributors to build the tool, not operate customer migrations.

The agent supplies its own model authentication. CLI is the complete fallback; no second model account is required by core services. External agents have their own data/egress policies. [AGENT-MODEL](../06-agent/AGENT-MODEL.md) owns authority and [MCP-SPEC](../06-agent/MCP-SPEC.md) owns connection semantics.

A suitable assessment request is:

> Assess the source at /work/customer-old for the target at /work/customer-new. Use /work/customer-migration for durable state. Present missing evidence and target/dependency alternatives. Stop at assessment; request required human confirmation and do not generate target code.

This prompt is not approval or registered scope. The operator must obtain a registered work-scope artifact under [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md) and [work-scope schema](../../specs/project/work-scope.schema.json); paths and a chat description alone are not `scope_ref`. The release walkthrough must demonstrate its actual guarded creation/registration route and returned ref. This guide does not invent a `scope create` command absent from the catalog. If registration cannot be supplied through supported interfaces, the walkthrough is blocked and must not claim assessment onboarding works.

## 5. Run and read the assessment, then stop

With Gate 1 and an actual registered scope, run the existing `assess --scope-ref` command. The following shell values are placeholders; replace them with refs returned by guarded services before use:

```bash
ASSESSMENT_SCOPE_REF='replace-with-registered-scope-ref'
rewrotron assess --scope-ref "$ASSESSMENT_SCOPE_REF" --observation AVAILABLE_ONLY --project /work/customer-migration --json

ASSESSMENT_REPORT_REF='replace-with-returned-assessment-ref'
rewrotron assessment show --assessment-ref "$ASSESSMENT_REPORT_REF" --project /work/customer-migration --json
rewrotron report --kind assessment --assessment-ref "$ASSESSMENT_REPORT_REF" --format markdown --project /work/customer-migration --json
```

`AVAILABLE_ONLY` uses available evidence; it does not demand launching a runtime. Select `SAFE_RUNTIME` only when the configured observation capabilities and admission requirements are met. A persisted report may complete with BLOCKED/UNKNOWN feasibility and static-only limitations. Missing evidence remains visible; missing persistence is an operation failure.

Stop here for assessment-only use. No plan approval, target writes or automatic continuation follows. Later work requires an explicit human CONTINUE_PLANNING selection, a separate executable plan request and exact Gate 2 confirmation. See [PROJECT-LIFECYCLE](../02-domain/PROJECT-LIFECYCLE.md). Persisted findings are reused only when current validity checks permit it; conversation history is not authority.

## Updates and removal

Follow the release installer/manual instructions; stop owned work and the supervisor before activating another version. Installing a binary alone does not upgrade project databases. Keep pinned release bytes and follow the existing backup/migration/recovery rules when reopening projects.

Uninstall preserves source/target repositories, migration workspaces and evidence by default. Removing a copied Skill/client connection is a separate explicit operation. [RELEASE](RELEASE.md) owns safe removal and retained-byte obligations; [RECOVERY](../08-storage/RECOVERY.md) owns process and project recovery.

# Rewrotron

**Rewrotron is an open-source tool being built to reconstruct legacy applications in a user-selected technology stack, using evidence about how the application behaves.**

You bring an old application and describe the target you want. Rewrotron is designed to inspect the source, observe supported workflows, explain the required changes, and produce a reviewable assessment. If you choose to continue, it plans the selected work, requests your approval, and reconstructs and verifies the application one behavioral slice at a time.

Its two central promises are to preserve the behavior covered by approved contracts and to show you what it is doing, why, and what remains uncertain.

## What can I use today?

**Early development: M0 foundation work has begun; no milestone is complete.** The repository contains specifications, machine schemas, SQL migration definitions, structural validation, and a small Go implementation of canonical JSON serialization and hashing with tests.

There is **no runnable migration engine, Rewrotron CLI/MCP binary, or runtime-validated source/target adapter yet**. The user workflow below describes the specified product. The quick start runs the repository checks available today.

The first planned reference migration is **CodeIgniter 3 → Laravel**, using independent environments and initially preserving the database schema. The core is designed to remain language, framework, and agent vendor neutral. Other source/target combinations require adapters and validation; they are not supported merely because the architecture can describe them.

## What problem does it solve?

A legacy application often hides important behavior in controllers, helpers, database queries, session state, scheduled jobs, and third-party libraries. Rewriting those files does not establish that the replacement still works for its users.

Rewrotron organizes the migration around **behavioral vertical slices**: a complete operation or workflow, such as signing in, creating an order, or rejecting an unauthorized request. A slice can span several files and services.

For example, a successful order submission may need to:

- Authenticate the user and check ownership.
- Validate input and write the correct database records.
- Queue a notification once.
- Update session state and redirect to the correct page.

The target must satisfy the approved obligations for that workflow. Internal architecture can improve, and intentional behavior changes must be explicit.

## What will Rewrotron do?

| Capability | What you get |
| --- | --- |
| Source and environment detection | Findings about the application, runtime, framework, dependencies, and operational entrypoints, with evidence and gaps. |
| Migration assessment | Required changes, dependency options, constraints, risks, blockers, alternatives, and missing evidence. |
| Target configuration and feasibility | Your requested environment kept separate from recommendations, followed by exact resolution for executable planning. |
| Behavior characterization | Supported observations and executable contracts for selected workflows, including state and side effects. |
| Incremental reconstruction | An approved plan and implementation in bounded slices, with explicit component ownership. |
| Differential verification | Comparison of actual source and target evidence in independent environments. |
| Transparent, resumable execution | Persisted work, decisions, evidence, current task, blockers, and next action. |
| Release preparation | An approved release package and runbook after required verification and performance/rollback evidence. Initial release scope does not deploy production. |

These are product requirements, not a list of features already implemented. See the [development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md) for delivery order.

## How does it work?

```text
Coherent source snapshot + environment and operational scope
    ↓
Analyze source + observe supported behavior
    ↓
Assessment report
    ↓
Your decision: stop, gather more evidence, change scope/target, or continue
    ↓
Resolve exact target + produce executable plan
    ↓
Your explicit approval
    ↓
Implement selected behavioral slices
    ↓
Verify actual behavior → repair mismatches → verify again
    ↓
Performance evidence + approved release package
```

Internally, source adapters extract a **Source IR** (source structure). The core organizes it into an **Application IR** (application relationships) and a **Behavior IR** (behavior obligations with evidence and uncertainty). A **Target Implementation Model** records how those obligations map to target components before generation. These are inspectable representations rather than reliance on an agent's conversation memory.

The deterministic engine owns state, approvals, safety checks, and verification results. Coding agents can propose implementations and explanations through the specified CLI/MCP/Skill interfaces. An agent's assertion does not establish correctness.

Project metadata is designed to live locally in SQLite, with large artifacts on disk. Local-first describes storage and execution defaults; an external coding agent or an explicitly enabled integration can still transmit data under its own configuration.

### What does verification mean?

A result is bounded by the approved source snapshot, environment, workflow cases, contracts, comparison policies, and exact target resolution:

- **PASS:** The required comparisons and obligations passed for the named scope.
- **MISMATCH:** Actual behavior violates a required comparison or obligation.
- **UNKNOWN:** Required evidence or supported capture is missing; no passing claim can be made for that affected scope.

Finite tests cannot prove every possible application behavior or that the test corpus is complete. Expected model state cannot fill a missing observation, and the checker cannot manufacture application effects to make a target pass. Safe sinks and disposable environments are required for supported side-effect tests; real payments and other irreversible actions are never blindly replayed.

## How will a user use it?

### 1. Bring the application and describe the scope

Supply readable source, available runtime/dependency information, critical workflows, business constraints, and safe test data/accounts. Include entrypoints outside HTTP, such as cron or workers. Missing information remains a visible gap.

Rewrotron captures a coherent snapshot and presents detected facts. **Gate 1** is your persisted approval to analyze and observe that source under the reviewed boundary.

### 2. Choose the target and get an assessment

Describe your target in the planned root configuration file, `rewrotron.yaml`: language/runtime, framework, database, cache, architecture, and related constraints. Recommendations never silently replace your choices. If you request PHP 8.4 and a recommendation suggests 8.3, your request remains 8.4; infeasibility is reported.

The assessment explains what was examined, what needs to change, why, and what is unknown. Static-only findings are labeled with their limitations. An assessment can complete successfully even when migration is blocked.

**You can stop here.** Assessment-only use is a supported checkpoint in the same workflow; it does not authorize implementation.

### 3. Select work, review the plan, and approve it

To continue, explicitly select the work and request planning. Rewrotron revalidates relevant evidence, resolves exact target components, and prepares an executable plan. **Gate 2** approves that exact plan separately from assessment or continuation.

A changed target or other relevant input invalidates affected downstream work and approval. Historical evidence remains available; reuse requires current validity checks.

### 4. Follow implementation and verification

Inspect current work, its factual rationale, discoveries, blockers, and next steps. Complete slices are checked against their approved cases and applicable already-implemented regressions. Whole-project acceptance is a separate, stricter scope.

Required performance and rollback evidence precede **Gate 3**, which approves release packaging and its runbook.

The exact planned commands and arguments are in the [CLI specification](docs/06-agent/CLI-SPEC.md). CLI JSON is the complete vendor-neutral fallback; MCP and Skills use the same guarded engine services. Named coding-agent integrations remain unverified until exercised.

## Quick start: explore and validate the repository today

This starts contributor exploration and checks, not a customer migration. Python 3 and Go 1.25.0 are needed for the commands below; the pinned bootstrap tooling is recorded in [tools/toolchain.json](tools/toolchain.json). First-time dependency/toolchain acquisition may need network access.

```bash
git clone https://github.com/devravik/rewrotron.git
cd rewrotron

# Validate documentation links, schemas, fixtures, SQL definitions, and vectors.
python3 -m pip install 'jsonschema>=4.10,<5'
python3 tools/docs/validate.py

# Test the currently implemented Go foundation.
GOTOOLCHAIN=go1.25.0 go test -race ./...
GOTOOLCHAIN=go1.25.0 go vet ./...
```

Structural validation checks specification artifacts. Go tests exercise the existing bootstrap code. Neither demonstrates a working migration, adapter, approval service, or verifier.

For implementation work, read [AGENTS.md](AGENTS.md), then select the currently authorized unit from the [development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md). The repository's [implementation skill](.agents/skills/rewrotron-implementation/SKILL.md) references the existing workflow and durable checkpoints for compatible coding agents.

## Where should I look next?

| If you want to… | Read |
| --- | --- |
| Remember the product's purpose and limits | [Product contract](docs/00-vision/PRODUCT.md) |
| Understand the architecture | [System design](docs/01-architecture/SYSTEM-DESIGN.md) |
| Understand assessment, continuation, and approvals | [Project lifecycle](docs/02-domain/PROJECT-LIFECYCLE.md) and [confirmation protocol](docs/02-domain/CONFIRMATION-PROTOCOL.md) |
| Understand planned installation and first assessment | [Installation guide](docs/10-development/INSTALLATION.md) and [release contract](docs/10-development/RELEASE.md) |
| Prepare a legacy application | [Customer inputs](docs/10-development/PROJECT-INPUTS.md) |
| Understand target choices and dependency handling | [Target specification](docs/04-migration/TARGET-SPECIFICATION.md) and [dependency resolution](docs/04-migration/DEPENDENCY-RESOLUTION.md) |
| Understand verification and its boundaries | [Differential testing](docs/05-verification/DIFFERENTIAL-TESTING.md) |
| Find planned user commands | [CLI specification](docs/06-agent/CLI-SPEC.md) and [MCP specification](docs/06-agent/MCP-SPEC.md) |
| See what will be built next | [Development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md), [roadmap](docs/11-roadmap/ROADMAP.md), and [project plan](REWROTRON_PROJECT_PLAN.md) |
| Find a specification or machine contract | [Documentation index](docs/README.md) and [schema registry](specs/README.md) |
| Review engineering checks and falsification tests | [Development](docs/10-development/DEVELOPMENT.md) and [testing](docs/10-development/TESTING.md) |
| Inspect prior architectural review dispositions | [Red-team resolution record](docs/11-roadmap/RED-TEAM-RESOLUTION.md) |

The README is an introduction. [AGENTS.md](AGENTS.md) and the canonical owners listed in the documentation index govern implementation and resolve authority; dated review-baseline descriptions are not current implementation inventories.

## License

[Apache-2.0](LICENSE).

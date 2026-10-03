# Rewrotron — Implementation-Grade Project Plan

**Document status:** Product exploration and historical design rationale; implementation contracts are maintained under docs/ and specs/
**Version:** 1.1 documentation review
**Date:** 2026-10-03
**Repository model:** One open-source monorepo
**Primary implementation language:** Go
**Local persistence:** SQLite
**Primary agent integration:** MCP + Agent Skill + CLI
**Initial source framework:** CodeIgniter 3.x
**Initial target framework:** Laravel 13.x
**Future targets:** Go, Python, Rust, additional PHP frameworks


> **Reading this document:** This exploration preserves product scenarios and rationale. The repository currently contains specifications, schemas and an initial SQL migration definition, with no Go engine or executed adapter corpus. The [documentation index](docs/README.md), [canonical development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md), [decision register](docs/11-roadmap/DECISION-REGISTER.md) and owning specifications supersede its historical phase/backlog numbering and illustrative layouts. Acceptance applies to an explicit finite corpus and scope, not every possible application execution.

---

## 0. Executive decision summary

This document defines Rewrotron as a vendor-agnostic legacy-application reconstruction and modernization engine.

Rewrotron is **not** a source-to-source transpiler. It is a system that reconstructs the externally observable behavior of an existing application in a new implementation while deliberately avoiding unnecessary legacy implementation details and improving internal architecture and performance.

The canonical workflow is:

```text
Legacy application
    ↓
Static analysis
    ↓
Runtime observation
    ↓
Request / workflow discovery
    ↓
Behavior model
    ↓
Executable compatibility cases
    ↓
Target architecture
    ↓
Target implementation
    ↓
Differential verification
    ↓
Repair mismatches
    ↓
Performance optimization
    ↓
Regression verification
    ↓
Release candidate
```

The first production scope is deliberately constrained:

```text
CodeIgniter 3.x application
        →
Laravel 13.x application
        →
existing relational database retained initially
```

The first release will **not** attempt arbitrary PHP-to-any-language migration. It will establish the core behavior-reconstruction architecture and CI3→Laravel adapter. Go, Rust and Python targets will reuse the same behavior model and verification system after the CI3→Laravel path becomes reliable.

---

# 1. Product definition

## 1.1 Product statement

> Rewrotron takes an existing web application as a source of truth, discovers its actual behavior, reconstructs that behavior in a modern target stack, removes demonstrably unnecessary legacy implementation, and proves behavioral equivalence through automated differential testing.

## 1.2 Primary customer problem

Legacy applications often contain:

- obsolete framework conventions;
- dead or duplicated code;
- tightly coupled controllers/models/helpers;
- outdated runtime versions;
- poor test coverage;
- undocumented business rules;
- slow database access;
- implicit framework behavior;
- custom authentication/session mechanisms;
- integrations nobody wants to rewrite manually.

A conventional rewrite requires engineers to understand the old system and manually preserve behavior. Rewrotron's primary value is reducing this reverse-engineering and validation effort.

## 1.3 Customer promise

The customer must be able to say:

> "This is the same application from a user and integration perspective, but the implementation is modernized, unnecessary legacy code is not carried forward, and the new implementation has evidence proving compatibility."

## 1.4 Explicit non-goals

Rewrotron v1 is not:

- a generic PHP transpiler;
- a framework upgrade wizard only;
- an arbitrary code formatter;
- a static-only dead-code remover;
- an automatic production cutover system;
- a tool that guarantees behavior it has never observed or specified;
- an autonomous system allowed to delete uncertain code;
- a requirement that the target architecture resemble the source architecture.

---

# 2. Non-negotiable design principles

## 2.1 Behavior over source structure

The source code is evidence about behavior. The behavior model is the migration authority.

A controller in CI3 does not need to become a controller containing the same logic in Laravel. The target may use controllers, services, repositories, jobs, policies, queries and adapters in a completely different arrangement.

## 2.2 Observe before rewriting

No target code generation begins for a behavior slice until that slice has enough source understanding to define a compatibility contract.

## 2.3 Deterministic tooling before LLM reasoning

Use deterministic tooling for:

- parsing;
- route extraction;
- symbol indexing;
- dependency analysis;
- trace collection;
- request replay;
- response comparison;
- database comparison;
- benchmarks;
- hashing;
- coverage calculation.

Use LLMs for:

- semantic interpretation;
- architecture inference;
- business-rule inference;
- migration planning;
- code generation;
- mismatch diagnosis;
- ambiguous-case analysis.

## 2.4 Never silently delete uncertain behavior

Code is classified as `ACTIVE`, `DORMANT`, `DEAD`, or `UNKNOWN`. Only `DEAD` code may be removed automatically.

## 2.5 The old application is the oracle, not absolute truth

The legacy application's observed behavior is the compatibility target. Rewrotron must distinguish:

- intended behavior;
- observed behavior;
- accidental behavior;
- unstable behavior;
- environment-specific behavior.

If behavior is inconsistent across repeated observations, the contract must model the variability rather than invent a single expected value.

## 2.6 No false claim of "exact"

Rewrotron can only guarantee equivalence for behavior that is represented by test/observation coverage or a stronger proof. The system must report coverage and unknown space explicitly.

## 2.7 Local-first

No source code, database content, credentials, traces or project data are uploaded to a Rewrotron service by default.

## 2.8 Vendor-agnostic

The Rewrotron engine must remain independent of Claude Code, Cursor, Codex, Hermes, Aider or any specific LLM API.

The canonical interfaces are:

1. CLI;
2. machine-readable CLI JSON;
3. MCP server;
4. Agent Skill package.

---

# 3. Initial technology baseline

## 3.1 Core

Go is the primary implementation language for the Rewrotron core, CLI, MCP server, storage layer and deterministic services.

The constitutional floor is Go 1.22+. RWT-001 selects and pins a supported stable toolchain compatible with all dependency floors; no current patch release is inferred from this document date.

## 3.2 Local storage

SQLite is the default project datastore.

Use:

- ordinary relational tables for canonical state;
- JSON columns for flexible payloads;
- FTS5 for full-text retrieval where appropriate;
- a filesystem artifact store for large traces and generated outputs.

Do not make a vector database mandatory in v1.

Semantic retrieval may be added later behind an interface.

## 3.3 Initial source framework

CodeIgniter 3.x is the first source adapter. The first reference implementation targets the CodeIgniter 3.1.x family, with runtime version detected per project rather than assumed.

CodeIgniter's documented execution model includes front-controller initialization, routing, security processing, controller/resource loading and view rendering. Its hooks expose lifecycle points including `pre_controller`, `post_controller_constructor`, `post_controller`, `display_override`, `cache_override`, and `post_system`. These are used by Rewrotron's runtime observer where applicable.

## 3.4 Initial target framework

Laravel 13.x is the first target adapter.

Laravel 13 requires PHP 8.3 or newer. The target PHP version must be explicitly pinned in each generated project.

The initial migration should retain the existing relational database rather than simultaneously changing database engine/schema semantics.

## 3.5 MCP

The first implementation targets MCP protocol revision 2025-11-25 over stdio, using only capabilities required by Rewrotron. The core implementation should remain useful through ordinary CLI execution even when an LLM client has no MCP support.

---

# 4. User workflow

The complete planned CLI flow is maintained in [CLI-SPEC](docs/06-agent/CLI-SPEC.md). Initialize/detect a minimal workspace, explicitly confirm the current source subject, analyze/observe/discover, synthesize the target plan, explicitly confirm the plan hash, generate and verify behavioral units, benchmark and prepare runbooks, then confirm the release hash before packaging.

An external agent operates through the same persisted gates/services and cannot infer confirmation from chat or create its own authoritative ledger. Source facts and safe runtime inputs must satisfy [PROJECT-INPUTS](docs/10-development/PROJECT-INPUTS.md).

---

# 5. Repository architecture

A single monorepo owns engine, adapters, tooling, fixtures, schemas, migrations and documentation. The exact implementation package map and dependency direction are maintained in [COMPONENTS](docs/01-architecture/COMPONENTS.md).

Framework parsing/hooks live in `adapters/source/php/`; target generation lives in `adapters/target/php/`. Neutral domain records/ports and services live in `internal/`. `internal/runtime` wires concrete implementations; core/behavior/verification do not import framework packages. CLI/MCP drive the same application services. Machine schemas live in `specs/`, SQL in `migrations/`, synthetic fixtures in `testdata/`. Generated customer targets are outside source roots.

---

# 6. Project storage model

Each migrated application receives a self-contained `.rewrotron` directory.

```text
customer-app/
├── application/
├── system/
├── public/
├── rewrotron.yaml                # Primary user intent
└── .rewrotron/
    ├── project.yaml              # Identity/settings view or legacy input
    ├── target.yaml               # Optional exclusive target input
    ├── detection.yaml
    ├── recommendations.yaml
    ├── source-dependencies.yaml
    ├── dependencies.yaml
    ├── plan.yaml
    ├── approvals.yaml            # SQLite export; no approval authority
    ├── project.db
    ├── lock
    ├── artifacts/
    │   ├── source/
    │   ├── traces/
    │   ├── snapshots/
    │   ├── screenshots/
    │   ├── reports/
    │   └── generated/
    ├── contracts/
    ├── cases/
    ├── missions/
    ├── mappings/
    └── logs/
```

The project database is the authoritative record of project state.

Global reusable knowledge is stored separately:

```text
~/.rewrotron/
├── global.db
├── cache/
├── tools/
└── knowledge/
```

Project knowledge must never be promoted into global knowledge automatically without passing the learning/evaluation policy.

---

# 7. Project configuration

Root `rewrotron.yaml` is the primary user-owned configuration; optional exclusive `.rewrotron/target.yaml` separates target intent. Detection and recommendations are generated artifacts and never overwrite either. [TARGET-SPECIFICATION](docs/04-migration/TARGET-SPECIFICATION.md) owns precedence, rich target fields, frozen revisions, feasibility and selective invalidation; [DEPENDENCY-RESOLUTION](docs/04-migration/DEPENDENCY-RESOLUTION.md) owns actual usage, recursive constraints, capability mappings and explicit strategy decisions. Gate 2 binds the exact specification and plan; no fourth gate is introduced.

Canonical configuration is validated by [project.schema.json](specs/project/project.schema.json). It expresses intent and secret references, never gate authority. A target may be absent at initialization; before planning it includes language, framework/version/runtime, architecture, preserved-database policy and an explicit safe output path. Minimal initialization precedes Gate 1 solely to persist detection and approval.

Use docker execution for customer source code; local mode is restricted to trusted development fixtures. Production and external writes remain denied. Confirmed baseline facts/overrides and operational state are authoritative SQLite records; YAML `confirmed` flags are rejected.

---

# 8. Lifecycle state machine

Canonical states and transitions are defined in [PROJECT-LIFECYCLE](docs/02-domain/PROJECT-LIFECYCLE.md):

```text
INITIALIZED → DETECTED → [Gate 1] SOURCE_CONFIRMED → ANALYZED → MODELED
→ PLAN_SYNTHESIZED → [Gate 2] PLAN_CONFIRMED → MIGRATING → VERIFIED
→ BENCHMARKED → [Gate 3] READY_CUTOVER
```

Operation/unit/mission blockers are distinct from project state. [CONFIRMATION-PROTOCOL](docs/02-domain/CONFIRMATION-PROTOCOL.md) defines source/configuration, plan and release hashes, exact human subjects, overrides, idempotency and drift invalidation. No old illustrative state names should be implemented as competing enums.

---

# 9. Static analysis engine

## 9.1 Purpose

Build an inventory and candidate behavior graph before executing the application.

## 9.2 Required outputs

The analyzer must identify, where detectable:

- PHP files;
- namespaces;
- classes;
- interfaces;
- traits;
- functions;
- methods;
- constants;
- properties;
- inheritance;
- calls;
- dynamic calls;
- include/require relationships;
- Composer dependencies;
- CI3 controllers;
- CI3 models;
- libraries;
- helpers;
- hooks;
- config files;
- route definitions;
- view files;
- database access;
- external HTTP access;
- session access;
- filesystem access;
- CLI entry points;
- cron-related files where identifiable.

## 9.3 Parser strategy

The analyzer must have a parser interface:

```go
type PHPParser interface {
    ParseFile(ctx context.Context, path string) (*ASTFile, error)
}
```

The initial implementation may use a PHP AST backend compatible with the detected source syntax.

If syntax cannot be parsed, the affected operation/unit receives a `BLOCKED` finding rather than falling back to regex and pretending the analysis is complete.

## 9.4 Static reachability

Construct a directed graph:

```text
symbol → referenced symbol
class → method
method → call
file → include
route → controller method
hook → callable
cron candidate → callable
```

Reachability is a candidate, not proof of execution for dynamic PHP.

---

# 10. Code classification

Every source component receives a classification:

### ACTIVE

Observed in at least one correlated runtime trace in the approved baseline. Static reachability without observation is DORMANT.

### DORMANT

No runtime evidence yet, but a plausible static/configuration path exists.

### DEAD

No static path, no dynamic path, no configured entry point, and no external operational reference discovered under the enabled analysis scope.

### UNKNOWN

Insufficient evidence.

Only `DEAD` is eligible for automatic omission.

A generated migration report must list all omitted components and the evidence supporting omission.

---

# 11. Runtime observer

The runtime observer is the core of Rewrotron.

## 11.1 Execution environment

The source application runs in an isolated environment whenever practical:

```text
rewrotron sandbox
├── web server
├── PHP runtime
├── source application
├── database
├── cache
├── mail sink
└── outbound HTTP proxy
```

The exact PHP version is detected from the project and pinned in the sandbox.

## 11.2 CI3 instrumentation

Where possible, use CodeIgniter hooks instead of modifying the framework core.

Probe lifecycle:

```text
pre_system
    ↓
request start
    ↓
route/controller observation
    ↓
pre_controller
    ↓
post_controller_constructor
    ↓
controller execution
    ↓
post_controller
    ↓
response capture
    ↓
post_system
```

Use `display_override` only when necessary for final-response capture.

The probe must never alter business behavior except for collecting telemetry and, where explicitly configured, replacing unsafe external side effects with test sinks.

## 11.3 Recorded request information

For each request:

```text
method
scheme
host
port
path
query
headers (redacted according to policy)
cookies (redacted/hashed as required)
body hash
body schema
route
controller
method
status
response headers
response body hash
response body artifact
redirect chain
session changes
DB queries
DB transaction boundaries
external HTTP calls
filesystem effects
execution time
memory
```

Secrets must not be written to normal logs.

---

# 12. Traffic and workflow discovery

Rewrotron must discover more than URLs.

It must identify workflows.

Example:

```text
GET /login
    ↓
POST /login
    ↓
GET /dashboard
    ↓
POST /orders
    ↓
GET /orders/123
```

This becomes one stateful workflow.

## 12.1 Sources of discovery

Use, in order of confidence:

1. existing automated tests;
2. recorded staging traffic provided by the customer;
3. application/server logs where supplied;
4. declared API collections;
5. static route discovery;
6. browser-driven exploration;
7. AI-guided exploration.

## 12.2 Entry point types

Rewrotron must model:

- HTTP routes;
- API endpoints;
- webhooks;
- CLI commands;
- scheduled jobs;
- queue consumers;
- event handlers;
- administrative actions.

HTTP routes are the first priority, but other entry points cannot be silently ignored.

---

# 13. Behavior graph

The behavior graph is the central intermediate representation.

A behavior node represents an externally relevant operation.

```text
Route
  ↓
Authentication
  ↓
Authorization
  ↓
Validation
  ↓
Business Logic
  ↓
Database
  ↓
External Services
  ↓
Session / Cache / Files
  ↓
Response
```

Each node may reference source symbols, runtime evidence and generated target symbols.

## 13.1 Required graph entities

```text
Application
Route
Workflow
Request
AuthenticationState
Role
Resource
Component
BusinessOperation
DatabaseOperation
ExternalCall
SideEffect
Response
Contract
TestCase
MigrationUnit
```

---

# 14. Behavioral contract

The canonical example and fields are maintained in [BEHAVIOR-MODEL](docs/02-domain/BEHAVIOR-MODEL.md), [IR-CONTRACTS](docs/02-domain/IR-CONTRACTS.md) and [behavior schemas](specs/README.md). Contracts govern operations; concrete cases bind exact versions, fixtures and expected dimensions. HTTP, CLI, cron and event dispatch have type-specific fields.

Business descriptions are not executable code. A required computation must have evidenced declarative semantics supported by the target adapter or remain unresolved. Workflows have explicit ordered steps and per-side variable bindings; authentication, state/effects and failure paths are part of the contract.

---

# 15. Comparison policy

[COMPATIBILITY-SPEC](docs/05-verification/COMPATIBILITY-SPEC.md) defines EXACT, SEMANTIC and scoped dynamic relation operators. No blanket HTML whitespace or token/ID masking is allowed. JSON types, array order, null/missing, decimals, status, redirects, principal/permissions and committed effects remain meaningful.

The verdicts are PASS/MISMATCH/UNKNOWN; expected modernization is a separately approved contract exception requiring new policy/plan subjects and rerun. Required incomplete capture is UNKNOWN, and historical results are immutable.

---

# 16. Golden master generation

For every behavior case:

```text
Run legacy
   ↓
Capture result
   ↓
Run again
   ↓
Detect deterministic vs variable fields
   ↓
Build canonical contract
   ↓
Freeze contract
```

Golden-master cases must be reproducible.

Each case stores:

```text
input
preconditions
source environment identifier
expected response
expected side effects
normalization rules
artifact hashes
creation timestamp
```

---

# 17. Test data and data isolation

Rewrotron must create deterministic fixtures where possible.

Test data must be separated from production data.

Preferred flow:

```text
source DB snapshot
      ↓
restore into sandbox DB
      ↓
create test identities
      ↓
run cases
      ↓
destroy sandbox
```

Never modify a production database during automatic behavior discovery.

---

# 18. Database behavior capture

Capture tables read/written, query/transaction diagnostics, committed pre/post-state and trigger/procedure effects in independent identically seeded environments. The [DB comparison contract](docs/05-verification/DATABASE-COMPARISON.md) defines typed row/null/decimal/timezone comparison and ambiguity handling.

IDs are exact by default. An explicit approved bijective mapping can account for generated IDs only when all foreign-key, URL, session, later request and external-effect references preserve the same relationships. Correlation ambiguity is UNKNOWN, not an ignored difference.

---

# 19. External side-effect handling

Side effects must be classified:

```text
PURE
READ_ONLY
LOCAL_MUTATION
EXTERNAL_MUTATION
IRREVERSIBLE
```

Automatic differential execution is allowed only according to a project safety policy.

Examples:

- sending email → mail sink;
- payment → provider sandbox/mock;
- webhook → local capture endpoint;
- cloud object upload → test bucket;
- SMS → mock provider;
- third-party write → disabled unless explicit sandbox configuration exists.

If safe isolation cannot be established, the case becomes `NEEDS_HUMAN_REVIEW`.

---

# 20. Authentication and session compatibility

Authentication is part of the public behavior contract.

The target must preserve where applicable:

- login URL;
- HTTP method;
- validation behavior;
- failure behavior;
- success redirect;
- session establishment;
- permissions;
- logout;
- password reset workflow;
- remember-me behavior;
- cookie scope and semantics.

## 20.1 Existing password hashes

Prefer a compatibility adapter that can continue verifying the existing password hash format rather than forcing a mass password reset.

Migration of password hashes is optional optimization and must not break login behavior.

## 20.2 Existing sessions

Transparent active-session continuity is required by default.

If the source session mechanism cannot be safely decoded and reproduced, Rewrotron must report:

```text
ACTIVE_SESSION_CONTINUITY_UNPROVEN
```

and block the cutover-readiness gate rather than silently accepting re-login.

---

# 21. Migration unit

Do not migrate one source file at a time.

Migrate by **vertical behavior slice**.

Example:

```text
POST /orders
    ├── route
    ├── auth
    ├── validation
    ├── business logic
    ├── DB reads
    ├── DB writes
    ├── external effects
    └── response
```

This entire slice becomes a migration unit.

Source files that implement several behaviors may be split into multiple target modules.

---

# 22. Target architecture generation

Rewrotron should produce an architecture proposal before generating substantial target code.

For Laravel 13, default proposal:

```text
app/
├── Http/
│   ├── Controllers/
│   ├── Middleware/
│   └── Requests/
├── Models/
├── Services/
├── Repositories/     # only when justified
├── Policies/
├── Jobs/
└── Support/
```

The target architecture must be derived from observed behavior and complexity.

Do not create repositories/services solely to imitate the old source hierarchy.

---

# 23. Laravel adapter

The Laravel adapter must know how to map behavior contracts to:

- routes;
- controllers;
- middleware;
- validation;
- authentication;
- authorization;
- sessions;
- Blade views;
- JSON API responses;
- database access;
- jobs;
- scheduled commands;
- configuration;
- environment variables.

Laravel's HTTP testing capabilities can serve as a target-side execution mechanism for status, headers, content, JSON, session and authentication assertions.

Generated compatibility tests must remain framework-independent at the contract level even if Laravel-specific test code is generated underneath.

---

# 24. View migration

Views require special handling because exact user-visible output matters.

Migration sequence:

```text
source HTML observation
      ↓
DOM/HTML analysis
      ↓
asset inventory
      ↓
view dependency mapping
      ↓
Blade target
      ↓
render comparison
```

The target view may be structurally different internally but must satisfy the configured presentation contract.

Compatibility checks include:

- title;
- visible text;
- forms;
- links;
- input names;
- input values;
- action URLs;
- classes/IDs when contractually relevant;
- scripts/assets;
- important ARIA attributes;
- DOM structure where required.

---

# 25. API compatibility

For APIs, the compatibility contract must preserve:

- path;
- method;
- accepted content types;
- request validation;
- authentication;
- authorization;
- response status;
- response schema;
- error schema;
- headers that matter to clients;
- pagination semantics;
- sorting semantics;
- filtering semantics;
- idempotency behavior;
- rate/error semantics where part of the documented behavior.

An OpenAPI document may be generated as a derivative artifact, but the internal Behavior IR remains authoritative.

---

# 26. Differential runner

The differential runner is a first-class subsystem.

```text
Compatibility Case
        │
        ├── Legacy
        │      ↓
        │   Observation
        │
        └── Target
               ↓
            Observation
                 │
                 ▼
             Comparator
                 │
       ┌─────────┼─────────┐
       ▼         ▼         ▼
     PASS     MISMATCH    UNKNOWN
                 │
                 ▼
             Diagnosis
                 │
                 ▼
               Repair
                 │
                 ▼
              Retest
```

## 26.1 Diff dimensions

```text
HTTP
  status
  headers
  cookies
  redirect chain
  body

STATE
  database
  session
  cache
  filesystem

EXTERNAL
  outgoing requests
  mail
  queue
  webhook

PERFORMANCE
  latency
  CPU
  memory
  DB query count
```

---

# 27. Mismatch classification

Every mismatch gets one classification:

```text
TARGET_BUG
SOURCE_VARIABILITY
NORMALIZATION_MISSING
TEST_DATA_PROBLEM
ENVIRONMENT_DIFFERENCE
EXPECTED_MODERNIZATION
UNKNOWN
```

`UNKNOWN` must not be automatically resolved as success.

---

# 28. LLM mismatch repair loop

When deterministic verification detects a mismatch, Hermes/Claude/Codex/Cursor/Aider receives a compact evidence package.

Example:

```text
CASE: order-create-001

Source:
  status: 302
  location: /orders/918
  DB: order created, inventory decremented

Target:
  status: 200
  body: JSON error
  DB: no change

Likely mismatch:
  target validation rejects a value accepted by source.
```

The LLM must not receive the entire database or project unless necessary.

It receives:

1. contract;
2. source evidence;
3. target evidence;
4. relevant source symbols;
5. relevant target symbols;
6. previous attempted fixes.

---

# 29. Repair attempt ledger

Every repair attempt is recorded.

```text
repair_attempt
├── mismatch_id
├── hypothesis
├── files_changed
├── tests_run
├── result
├── before_hash
├── after_hash
└── created_at
```

This prevents the agent from repeatedly trying the same failed fix.

---

# 30. Research and migration ledger

The complete project table catalog and executable initial DDL are maintained in [SQLITE-SCHEMA](docs/08-storage/SQLITE-SCHEMA.md) and [0001_initial.sql](migrations/0001_initial.sql). Source/IR, detection/baseline/approval, case/workflow, plan/unit, run/result/mismatch, experiment/repair, mission/operation and artifact/learning records are explicitly separated.

There is no existing deployed schema; future schema changes require incremental numbered migrations and compatibility fixtures. Project/global state have separate privacy and migration namespaces.

---

# 31. Append-only event log

The [event model](docs/08-storage/EVENT-MODEL.md) defines event IDs, project/sequence ordering, actor/request attribution and sanitized payloads. State transitions and events commit in one SQLite transaction. Append-only triggers prevent history mutation/deletion.

JSONL is a recoverable export of committed events, not a second atomic ledger. Evidence artifacts and baseline/contract/plan records are needed for rebuilding supported projections; an event log alone cannot roll back files or external effects.

---

# 32. Experiment fingerprinting

Every operation that could be repeated receives a deterministic fingerprint.

Conceptually:

```text
project
+ entrypoint
+ input signature
+ authentication state
+ test identity
+ target state
+ technique
+ relevant environment hash
```

Before execution:

```text
check fingerprint
    ↓
found?
 ┌──┴──────┐
 yes       no
  │         │
check state execute
  │
 same → skip
 changed → revalidate
```

This is a strict requirement for resume support.

---

# 33. Resume semantics

When Rewrotron starts, it must not rely on the LLM's previous conversation.

It loads:

```text
project state
current phase
open migrations
open mismatches
unresolved unknowns
coverage gaps
previous repair attempts
recent lessons
next recommended task
```

The agent receives a compact project snapshot.

Example:

```text
Project: customer-portal
State: MIGRATING

Routes: 100%
Behavior cases: 78%
Migration units: 63%
Compatibility: 91%

Open mismatches: 17
Blocked items: 2
Unverified routes: 34

Next task:
  Fix POST /orders validation mismatch
```

---

# 34. Mission and task system

The [mission protocol](docs/06-agent/MISSION-PROTOCOL.md) defines readiness, states, context/fingerprint, dependencies, priority, atomic lease claiming, renewal, fencing and evidence-based completion. Engineering repository tasks and customer migration missions are distinct.

`next` peeks without mutation. Claim/renew/complete require the current trusted owner and live fence. Expired workers cannot publish. Completion requires current verifiable evidence rather than a free-text agent assertion.

---

# 35. Next-task selection

Ready tasks are sorted deterministically by priority descending, creation time ascending and mission ID ascending. Applicable gates, safe environment, prerequisite completion and current subject hashes determine readiness before priority. The mission specification defines concrete severity priorities and three-attempt current-context limits.

An empty/blocked queue returns a typed null mission and explicit reason; it does not invent a new task or spin indefinitely.

---

# 36. Skills

The Rewrotron Skill teaches an LLM client how to operate Rewrotron.

It must not contain the entire migration implementation.

Recommended structure:

```text
skills/rewrotron/
├── SKILL.md
├── references/
│   ├── lifecycle.md
│   ├── ci3.md
│   ├── laravel13.md
│   ├── compatibility.md
│   ├── troubleshooting.md
│   └── safety.md
└── scripts/
    ├── bootstrap.sh
    ├── inspect.sh
    └── report.sh
```

The Skill must instruct the agent to use the CLI/MCP as the source of truth.

---

# 37. Vendor-neutral agent interfaces

[CLI-SPEC](docs/06-agent/CLI-SPEC.md) and [MCP-SPEC](docs/06-agent/MCP-SPEC.md), with the complete [schema registries](specs/README.md), own every supported command/tool/resource and response. Stdio MCP uses JSON-RPC lifecycle/capabilities and typed structured results; CLI uses one JSON envelope with exit/error semantics.

No arbitrary SQL/shell/write or result-override interface exists. Confirmation tools consume approval for exact human-reviewed subjects; actors cannot be claimed by an agent-controlled field. Every essential migration can run through CLI without an LLM.

---

# 38. Capability registry

Rewrotron must maintain a registry of reusable tools and migration capabilities.

A capability contains:

```text
id
name
version
description
inputs
outputs
language
location
status
tags
applicability
usage_count
success_count
failure_count
false_positive_count
```

Before creating new code, the agent should be able to search for existing capabilities.

This prevents repeated implementation of the same analysis or conversion helper.

---

# 39. Generated scripts and tool lifecycle

New tools can be created in three stages:

```text
EPHEMERAL
    ↓
EXPERIMENTAL
    ↓
VALIDATED
    ↓
STABLE
```

A generated tool cannot become stable solely because the agent says it works.

It must have:

- automated tests;
- reproducible outputs;
- version metadata;
- usage history;
- acceptable failure rate;
- no unresolved critical correctness issue.

Stable tools may be promoted into the monorepo's reusable tool library.

---

# 40. Knowledge system

Two distinct knowledge scopes are mandatory.

## 40.1 Project knowledge

Examples:

```text
custom MY_Controller performs authentication
session stored in database
payment service uses provider sandbox
reports use special SQL function
```

Project knowledge must remain isolated by default.

## 40.2 Global knowledge

Examples:

```text
common CI3 query-builder migration pattern
common CodeIgniter hook pattern
common session compatibility technique
common Blade migration pattern
```

Global knowledge requires evidence before promotion.

---

# 41. Learning system

The initial learning system is procedural rather than model-weight training.

The system learns:

- reusable transformations;
- architecture patterns;
- framework mappings;
- failure patterns;
- mismatch causes;
- successful repair strategies;
- dead-code evidence patterns;
- target-specific conventions.

It should not automatically fine-tune an LLM in v1.

## 41.1 Learning candidate lifecycle

```text
Observation
   ↓
Lesson candidate
   ↓
Repeated evidence
   ↓
Pattern candidate
   ↓
Evaluation
   ↓
Global knowledge
```

## 41.2 Negative learning

Failed migrations must also create knowledge.

Example:

```text
Transformation X failed because source used a dynamic hook.
```

Record this so future plans can avoid the same incorrect transformation.

---

# 42. Skill improvement

A skill can be improved only through an explicit process:

```text
problem observed
    ↓
skill gap identified
    ↓
proposed change
    ↓
benchmark against historical cases
    ↓
regression suite
    ↓
review
    ↓
promote version
```

Every skill version remains available for historical reproducibility.

---

# 43. Global benchmark corpus

Create a small set of representative migration fixtures.

Example:

```text
fixtures/
├── ci3-simple-crud
├── ci3-auth-session
├── ci3-hooks
├── ci3-rest-api
├── ci3-legacy-mixed-code
├── ci3-dynamic-models
└── ci3-realistic-portal
```

Every migration engine change must run against the benchmark corpus.

The final benchmark should include at least one intentionally messy real-world-like application with:

- dynamic loading;
- custom base controllers;
- hooks;
- APIs;
- forms;
- sessions;
- file uploads;
- cron/CLI;
- external integration adapters;
- nontrivial DB queries.

---

# 44. Verification levels

Rewrotron reports multiple levels of confidence.

## Level 0 — Static confidence

Only static analysis exists.

## Level 1 — Runtime observed

Behavior observed in legacy runtime.

## Level 2 — Contract verified

Target passes the generated contract.

## Level 3 — Workflow verified

End-to-end workflow passes across multiple state transitions.

## Level 4 — Regression corpus verified

Historical cases all pass.

## Level 5 — Production-like shadow verified

A customer-approved staging or shadow environment passes representative traffic.

No migration is labeled production-ready below the project's configured minimum verification level.

---

# 45. Performance optimization

Performance optimization starts only after behavior parity for the relevant scope.

Metrics:

```text
p50
p95
p99
CPU
memory
DB query count
DB execution time
external-call time
response size
cold start
```

The optimizer should identify:

- N+1 queries;
- unnecessary DB reads;
- repeated computation;
- redundant serialization;
- blocking external calls;
- avoidable memory allocation;
- cache opportunities.

The old system remains the behavioral oracle while performance becomes an optimization objective.

---

# 46. Optimization acceptance rule

Every optimization must pass:

```text
behavioral regression suite
+
new performance benchmark
```

A change is accepted only when:

```text
compatibility: PASS
AND
performance: improved or explicitly justified
```

Do not trade correctness for benchmark numbers.

---

# 47. Security model

Rewrotron handles potentially sensitive source code, credentials and application data.

## 47.1 Default controls

- sandbox source and target environments;
- deny production by default;
- redact secrets from logs;
- hash or tokenize sensitive values where possible;
- store artifacts locally;
- no default telemetry upload;
- explicit outbound-network policy;
- explicit external-write policy;
- audit all agent-mediated destructive operations.

## 47.2 Credentials

Credentials may be loaded into runtime environments, but:

- never place them in source code;
- never store plaintext credentials in events;
- never include them in LLM context unless necessary;
- redact secrets from traces;
- support environment-variable references instead of values.

## 47.3 Agent permissions

The Rewrotron agent should have separate permissions for:

```text
READ_SOURCE
WRITE_TARGET
RUN_SOURCE_SANDBOX
RUN_TARGET_SANDBOX
READ_ARTIFACTS
WRITE_PROJECT_STATE
EXTERNAL_NETWORK
PRODUCTION_ACCESS
DESTRUCTIVE_ACTIONS
```

Production access and destructive actions are denied unless explicitly enabled by the user.

---

# 48. Artifact management

Large artifacts belong on the filesystem.

Examples:

```text
HTTP traces
HAR files
screenshots
DOM snapshots
DB dumps
query logs
benchmark outputs
compiler logs
```

SQLite stores:

```text
artifact ID
sha256
path
size
mime type
creation time
project
case/mission association
```

Artifact hashes are used to detect unchanged data and avoid duplicate storage.

---

# 49. Logging

Logging levels:

```text
ERROR
WARN
INFO
DEBUG
TRACE
```

Logs are structured JSON internally.

Every agent-triggered operation includes:

```text
project_id
mission_id
agent_session_id
tool_id
skill_version
```

This allows us to determine why a migration produced a result.

---

# 50. Concurrency

[MISSION-PROTOCOL](docs/06-agent/MISSION-PROTOCOL.md) and [RECOVERY](docs/08-storage/RECOVERY.md) define short SQLite write transactions, expected project revision, at-least-once leases with monotonically increasing fencing tokens, and target patch ownership.

Long subprocess/parser work does not hold DB transactions. Publication checks current hashes/fences and before-file digests; independent cases have separate state, and overlapping target files publish serially. A stale worker cannot complete or overwrite customer edits.

---

# 51. Multi-agent support

Rewrotron must support multiple LLM clients eventually working on one project.

```text
Project DB
    │
 ┌──┼────────┐
 ▼  ▼        ▼
Codex Claude Hermes
```

Each agent claims missions through Rewrotron.

The agents do not maintain independent task backlogs.

---

# 52. Human review gates

The three mandatory gates are source baseline, target/plan, and release/cutover, enforced through immutable persisted exact-subject approvals. Scope exclusions, intentional behavior changes and sensitive cross-project exports require explicit reviewed decisions included in their owning subjects/policies.

The initial engine does not implement production writes. A Gate 3 package contains verified evidence and customer-supplied staging-tested operator/rollback commands. Agents cannot convert UNKNOWN into PASS or silently omit dormant/unknown behavior.

---

# 53. CLI specification

The complete supported command and flag catalog is [CLI-SPEC](docs/06-agent/CLI-SPEC.md), with typed inputs/results in [commands.json](specs/cli/commands.json). Initialization/detection/confirmation, analysis/observation/discovery, planning/generation/verification, benchmark/status/missions, report/export/learning/skills and Gate 3 package operations are explicitly documented.

Unknown/legacy speculative commands are rejected; `--auto-confirm` is removed. Dry-run never writes state, claims leases, executes source or grants approval. Use the canonical registry rather than this exploration to implement commands.

---

# 54. `doctor` command

`rewrotron doctor` performs environment checks.

It must verify:

- Go tool availability where required;
- PHP runtime availability;
- Composer availability;
- source project dependencies;
- database connectivity;
- browser runtime if needed;
- Docker/container runtime if sandbox mode requires it;
- filesystem permissions;
- available disk space;
- target framework prerequisites.

It reports exact remediation commands where possible.

---

# 55. `status` command

Required output:

```text
Project
Source
Target
Phase
Overall compatibility
Behavior coverage
Route coverage
Migration coverage
Open mismatches
Blocked items
Latest mission
Next mission
```

Machine JSON must expose the same facts in structured form.

---

# 56. `report` command

The migration report must include:

```text
Executive summary
Source inventory
Target inventory
Behavior coverage
Routes
Workflows
Migrated components
Omitted components
Omission evidence
Compatibility results
Open mismatches
Performance comparison
Known limitations
Required human decisions
Cutover checklist
```

---

# 57. Generated target repository

The generated Laravel repository should be clean enough to be maintained independently after migration.

It must not require Rewrotron at runtime unless the customer explicitly wants Rewrotron verification agents.

It must include:

```text
README.md
.env.example
composer.json
application code
tests/
compatibility-tests/
configuration documentation
migration report
```

The application itself must not depend on `.rewrotron/project.db` in production.

---

# 58. Compatibility test packaging

Compatibility tests should be retained in the target repository under:

```text
tests/Compatibility/
```

The underlying canonical cases remain in Rewrotron's project data.

Generated tests are derivatives and can be regenerated.

---

# 59. Cutover strategy

Rewrotron should eventually support a controlled cutover.

Preferred sequence:

```text
Legacy 100%
     ↓
Target deployed separately
     ↓
Read-only verification
     ↓
Shadow traffic where safe
     ↓
Canary subset where supported
     ↓
Increase target traffic
     ↓
Full target
     ↓
Legacy retained for rollback
```

Initial releases should provide configuration and runbooks rather than automatically changing a production load balancer.

---

# 60. Rollback strategy

A migration is not complete unless rollback is defined.

Required metadata:

```text
source version
 target version
 DB schema state
 config version
 feature flags
 route ownership
 session strategy
 rollback trigger
 rollback command/runbook
```

Rollback triggers may include:

- compatibility regression;
- error rate increase;
- performance regression;
- missing external effects;
- authentication/session failure.

---

# 61. Testing strategy for Rewrotron itself

## Unit tests

Every pure component must have unit tests.

Priorities:

- parser transformations;
- fingerprinting;
- normalization;
- graph algorithms;
- coverage calculation;
- route extraction;
- contract comparison.

## Integration tests

Test:

- SQLite persistence;
- CLI;
- MCP;
- source analyzer;
- runtime probe;
- differential runner.

## Fixture tests

Every supported CI3 pattern gets a fixture.

## End-to-end tests

At least one full:

```text
CI3 repo
→ analyze
→ observe
→ contracts
→ Laravel
→ differential verification
```

## Regression tests

Every previously fixed mismatch becomes a regression case.

---

# 62. Quality metrics

Do not use lines of migrated code as the main success metric.

Primary metrics:

```text
behavior coverage
contract pass rate
workflow pass rate
mismatch rate
unknown rate
migration automation rate
manual intervention count
false-deletion count
performance delta
migration time saved
```

A particularly important metric is:

> percentage of representative customer behavior that reaches verified parity without manual code rewriting.

---

# 63. Initial success thresholds

Customer success is defined by [PRODUCT](docs/00-vision/PRODUCT.md), [COVERAGE](docs/05-verification/COVERAGE.md) and [PERFORMANCE](docs/05-verification/PERFORMANCE.md). All required in-scope cases/dimensions must pass; required unknowns and uncovered inventory block readiness. Percentages name explicit denominators and a zero denominator is not 100%.

Release support requires reproducible fixtures, independent target builds, gate/drift/recovery/security checks, omission evidence, actual session continuity, measured performance and complete reviewed runbooks. These are future executed deliverables, not claims established by this document.

---

# 64. Phase 0 — Repository and contracts foundation

## Deliverables

- monorepo;
- Go module;
- CLI skeleton;
- SQLite schema;
- project initialization;
- event log;
- JSON command output;
- lock handling;
- artifact store;
- configuration loading.

## Exit criteria

A new directory can be initialized and reopened with no loss of state.

```bash
rewrotron init ./sample
rewrotron status --json
rewrotron resume
```

---

# 65. Phase 1 — CI3 static analyzer

## Deliverables

- PHP parser abstraction;
- CI3 detector;
- Composer inspection;
- route extractor;
- controller/model/library/helper detector;
- hook detector;
- symbol index;
- static call graph;
- code classification.

## Exit criteria

Against the benchmark CI3 fixtures:

- all routes are detected where statically resolvable;
- all benchmark controllers/models/helpers are classified;
- hooks are identified;
- parser failures are explicit.

---

# 66. Phase 2 — CI3 runtime observer

## Deliverables

- sandbox launcher;
- CI3 hook probe;
- request recorder;
- response recorder;
- DB query recorder;
- session observer;
- external HTTP capture;
- artifact writer.

## Exit criteria

A benchmark workflow produces a complete trace without changing application behavior.

---

# 67. Phase 3 — Behavior discovery

## Deliverables

- workflow graph;
- state model;
- behavior case generator;
- dynamic-value detector;
- golden-master generator;
- coverage engine.

## Exit criteria

For the realistic CI3 fixture:

- login workflow discovered;
- CRUD workflow discovered;
- API workflow discovered;
- DB side effects captured;
- stable vs dynamic response fields identified.

---

# 68. Phase 4 — Laravel generator

## Deliverables

- Laravel 13 project bootstrap;
- route generator;
- controller generator;
- validation generator;
- authentication compatibility adapter;
- DB service layer;
- view migration;
- configuration mapper;
- basic compatibility test generator.

## Exit criteria

The generated target boots and passes the first 20–30 representative compatibility cases from the benchmark application.

The exact number is configurable; the principle is that migration must begin with small verified slices.

---

# 69. Phase 5 — Differential verifier

## Deliverables

- HTTP comparator;
- JSON comparator;
- HTML comparator;
- cookie/session comparator;
- DB logical comparator;
- external side-effect comparator;
- mismatch classifier;
- repair evidence packaging.

## Exit criteria

The system can identify a deliberately introduced target bug and produce a precise mismatch report.

---

# 70. Phase 6 — Agent integration

## Deliverables

- Rewrotron Agent Skill;
- MCP server;
- project snapshot resource;
- mission tools;
- verification tools;
- knowledge tools;
- skill registry;
- CLI JSON documentation.

## Supported clients

Initial validation matrix:

```text
OpenAI Codex CLI
Claude Code
Cursor
Aider
Hermes
custom CLI/headless loop
```

The product is considered vendor-agnostic when all essential functionality works without agent-specific internal code.

---

# 71. Phase 7 — Learning system

## Deliverables

- project lessons;
- global knowledge;
- capability registry;
- tool usage telemetry;
- skill versioning;
- benchmark-based skill evaluation;
- promotion workflow.

## Exit criteria

A migration lesson from Project A can be retrieved and reused in Project B without copying Project A's private data into Project B.

---

# 72. Phase 8 — Performance modernization

## Deliverables

- profiler integration;
- query-count comparison;
- benchmark runner;
- optimization mission generation;
- optimization reports.

## Exit criteria

The system can identify at least one controlled performance regression and one controlled optimization while maintaining compatibility.

---

# 73. Phase 9 — Additional target adapters

Order:

```text
CI3 → Go
CI3 → Python
CI3 → Rust
```

The common Behavior IR and compatibility suite must be reused.

No target-specific migration engine may duplicate source behavior extraction logic.

---

# 74. Phase 10 — Additional source adapters

Candidate sources:

```text
Laravel legacy
Symfony legacy
Custom PHP
CodeIgniter 2
```

Priority should be determined by actual customer demand.

---

# 75. Team structure

For a small team, roles can be combined.

## Technical lead / architecture

Owns:

- core architecture;
- contracts;
- repository structure;
- extension points;
- release quality.

## Runtime/analysis engineer

Owns:

- PHP analysis;
- CI3 instrumentation;
- behavior graph;
- DB/runtime observation.

## Migration engine engineer

Owns:

- Laravel adapter;
- generators;
- mapping system;
- differential fixes.

## Agent/tooling engineer

Owns:

- MCP;
- CLI;
- Agent Skill;
- client compatibility;
- capability registry.

## QA/benchmark engineer

Owns:

- fixture corpus;
- regression suites;
- compatibility benchmark;
- performance benchmark.

---

# 76. Code ownership rules

No subsystem may contain LLM prompts as the primary implementation of deterministic behavior.

All important transformations must have:

- explicit interfaces;
- tests;
- structured outputs;
- reproducible fixtures.

Prompts are versioned like code.

---

# 77. Error handling policy

Errors are categorized:

```text
USER_ERROR
ENVIRONMENT_ERROR
SOURCE_ERROR
TARGET_BUILD_ERROR
ANALYSIS_ERROR
OBSERVATION_ERROR
VERIFICATION_ERROR
AGENT_ERROR
INTERNAL_ERROR
```

Every error should include:

```text
category
message
retryable
recommended_action
artifact reference
```

LLM-generated explanations must never replace the underlying deterministic error.

---

# 78. Recovery policy

Long-running migration tasks must be restartable.

A process crash must not corrupt project state.

Required patterns:

- SQLite transactions;
- atomic artifact writes;
- mission leasing;
- immutable events;
- checkpoints;
- resumable subprocesses;
- generated code written to temporary paths before replacement.

---

# 79. Git integration

Git is used for generated target repositories and optionally source analysis.

Rewrotron should:

- detect current branch;
- record source commit hash;
- record target commit hash;
- create logical migration commits if configured;
- never rewrite customer history without explicit instruction.

Each migration unit can optionally become a commit:

```text
migrate route /orders
fix mismatch /orders
verify /orders
```

---

# 80. Deterministic build requirements

Every generated target must record:

```text
source commit
Rewrotron version
adapter version
prompt/skill versions when AI was used
target framework version
runtime version
dependency lockfile hash
contract corpus hash
```

This enables reproduction of a migration result.

---

# 81. License and dependency policy

Recommended Rewrotron license: Apache-2.0.

Reason: permissive open-source distribution with an explicit patent grant suitable for both community and commercial adoption.

All dependencies must pass a license-policy check before release.

Avoid introducing a dependency solely because an LLM selected it.

Generated customer code must retain required notices from third-party dependencies.

---

# 82. Distribution

Initial distribution:

```text
GitHub source
prebuilt CLI binaries
```

Platforms:

```text
Linux amd64
Linux arm64
macOS amd64
macOS arm64
Windows amd64
```

The MCP server should be distributed as part of the same release artifact.

---

# 83. CI pipeline

Every pull request runs:

```text
go test ./...
go vet ./...
static analysis
schema validation
SQLite migration tests
CLI integration tests
MCP protocol tests
benchmark smoke tests
fixture migrations
license check
secret scan
```

Nightly/extended CI runs the full migration corpus.

---

# 84. Documentation requirements

The repository must maintain:

```text
README
Architecture
Getting Started
CI3 source adapter
Laravel target adapter
CLI reference
MCP reference
Skill reference
Project format
Behavior contract format
Contributing
Security
Troubleshooting
```

Every public command and MCP tool must be documented.

---

# 85. Developer experience

A contributor should be able to:

```bash
git clone ...
cd rewrotron
make bootstrap
make test
make test-fixtures
make build
```

No cloud account should be required for core tests.

---

# 86. Risk register

## Risk: dynamic PHP behavior

Mitigation:

- runtime tracing;
- explicit UNKNOWN state;
- agent-assisted analysis;
- no false reachability claims.

## Risk: migration produces behavior drift

Mitigation:

- golden master;
- differential execution;
- vertical migration slices;
- regression suite.

## Risk: dead-code removal breaks hidden functionality

Mitigation:

- ACTIVE/DORMANT/DEAD classification;
- cron/CLI/webhook analysis;
- safe omission rules.

## Risk: external integration side effects

Mitigation:

- sandbox providers;
- mock sinks;
- external-write policy;
- human gate.

## Risk: LLM repeats failed fixes

Mitigation:

- repair ledger;
- experiment fingerprinting;
- mission state.

## Risk: project state becomes huge

Mitigation:

- SQLite metadata;
- filesystem artifacts;
- hashes;
- compact snapshots;
- archival policies.

## Risk: agent-specific lock-in

Mitigation:

- CLI first;
- JSON contracts;
- MCP;
- portable Skill.

## Risk: global knowledge contaminates customer projects

Mitigation:

- project/global scope separation;
- promotion gate;
- confidence/evidence metadata.

## Risk: generated code becomes unmaintainable

Mitigation:

- architecture proposal;
- target framework conventions;
- generated test suite;
- code quality gates;
- post-generation refactoring.

---

# 87. Key architecture decisions

These decisions are locked unless deliberately changed through an ADR:

| Decision | Choice |
|---|---|
| Repository | Single monorepo |
| Core language | Go |
| Local database | SQLite |
| Large artifacts | Filesystem/object-compatible abstraction |
| Global semantic search | Optional; interface first |
| Agent integration | CLI + JSON + MCP + Skill |
| Source model | Behavior-first |
| Migration granularity | Vertical behavior slice |
| Initial source | CodeIgniter 3 |
| Initial target | Laravel 13.x |
| Initial DB strategy | Retain existing DB |
| Verification | Differential |
| Dead-code deletion | Only proven DEAD |
| Learning | Procedural/knowledge-based first |
| Cloud dependency | None for core |
| Production access | Denied by default |
| License | Apache-2.0 |

---

# 88. Development backlog

The dependency-ordered [DEVELOPMENT-PLAN](docs/11-roadmap/DEVELOPMENT-PLAN.md) supersedes the historical first-sprint list. RWT-001–027 there are canonical task IDs with exact files, interfaces, negative fixtures, verification and evidence; see [BACKLOG](docs/11-roadmap/BACKLOG.md).

Start with safe persistence/artifacts/configuration/gates/CLI, then source characterization and the verifier, then confirmed target slices, auth/transactions/effects, agent integration, measured performance and release packages. Optional learning/extensions follow proven baseline operation.

---

# 89. First end-to-end demonstration

Use the scenario in the [development plan](docs/11-roadmap/DEVELOPMENT-PLAN.md): detect and persist Gate 1, analyze/observe login and CRUD in reset independent sandboxes, lock cases, synthesize/confirm Gate 2, generate target slices, demonstrate deliberate MISMATCH and incomplete-capture UNKNOWN, repair with fences/history, resume after crash, rerun corpus, benchmark, rehearse session/rollback evidence, confirm Gate 3 and package.

No production mutation occurs. A read-only first slice validates plumbing; full auth/state/integration evidence is required before broad adapter maturity or whole-app claims.

---

# 90. Definition of Done for a migration unit

A migration unit is complete only when:

```text
[ ] source behavior observed
[ ] contract generated
[ ] target implementation generated
[ ] target builds
[ ] compatibility tests generated
[ ] legacy baseline recorded
[ ] target comparison executed
[ ] all deterministic differences resolved
[ ] all unknown differences classified
[ ] side effects verified
[ ] authorization/authentication verified
[ ] relevant performance measured
[ ] mapping recorded
[ ] mission closed
[ ] artifacts hashed
```

---

# 91. Definition of Done for a project

A project is complete when:

```text
[ ] agreed entrypoints analyzed
[ ] agreed workflows discovered
[ ] compatibility corpus frozen
[ ] target build reproducible
[ ] compatibility corpus passes
[ ] no unresolved critical mismatch
[ ] source omissions documented
[ ] DORMANT/UNKNOWN components disclosed
[ ] performance report generated
[ ] external integrations validated
[ ] authentication/session continuity validated
[ ] rollback plan documented
[ ] production deployment plan documented
```

---

# 92. What Rewrotron should learn over time

The long-term system should accumulate:

```text
framework → framework mappings
source pattern → target pattern
legacy pattern → modernization strategy
business pattern → canonical behavior model
mismatch → root cause
root cause → repair strategy
architecture → recommended migration architecture
technology → known compatibility issue
skill → measured success rate
```

The system should eventually be able to encounter a new CI3 application and say, based on accumulated evidence:

```text
This application resembles 14 previously migrated applications.

Known patterns:
- custom MY_Controller authentication
- query-builder repositories embedded in models
- DB-backed sessions
- scheduled report generation

Recommended approach:
- preserve DB initially
- implement authentication compatibility adapter
- migrate report generation separately
- retain legacy session reader until cutover verification
```

That is the intended compounding effect.

---

# 93. Long-term evolution

After CI3→Laravel is mature:

```text
                    Rewrotron Core
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        Source Adapters  Behavior IR  Target Adapters
              │                         │
        ┌─────┼──────┐           ┌─────┼──────┐
        ▼     ▼      ▼           ▼     ▼      ▼
       CI3  Laravel  PHP        Laravel Go    Rust
                             
```

Then add:

- CI3→Go;
- CI3→Rust;
- CI3→Python;
- Laravel→Go;
- Symfony→Go;
- custom PHP→Go;
- database modernization;
- frontend modernization;
- service extraction.

The Behavior IR and differential verifier remain common.

---

# 94. Long-term product direction

Rewrotron can eventually expose three modes.

## Local developer mode

```text
repo → modernized repo
```

## Assisted migration mode

```text
repo + team review → staged migration
```

## Enterprise modernization mode

```text
repo + staging + traffic
→ continuous parity
→ migration
→ performance optimization
→ controlled cutover
```

The open-source core remains usable in all three.

---

# 95. Final architectural model

```text
                            REWROTRON
                                │
                    ┌───────────┴───────────┐
                    │   Agent Interfaces    │
                    ├───────────────────────┤
                    │ CLI                   │
                    │ JSON CLI              │
                    │ MCP                   │
                    │ Agent Skill           │
                    └───────────┬───────────┘
                                │
                         REWROTRON CORE
                                │
          ┌─────────────────────┼─────────────────────┐
          ▼                     ▼                     ▼
      Analyzer              Observer              Project
          │                     │                   State
          └──────────────┬──────┴──────────────┬──────┘
                         ▼                     ▼
                   BEHAVIOR IR           SQLite + Events
                         │
                         ▼
                  Migration Planner
                         │
                 ┌───────┴───────┐
                 ▼               ▼
             Target Adapter   LLM Agent
                 │               │
                 └───────┬───────┘
                         ▼
                   Generated App
                         │
                         ▼
                 Differential Runner
                         │
               ┌─────────┴─────────┐
               ▼                   ▼
             PASS              MISMATCH
               │                   │
               ▼                   ▼
          Benchmark             Diagnosis
               │                   │
               └─────────┬─────────┘
                         ▼
                       Repair
                         │
                         ▼
                     Regression
                         │
                         ▼
                      Learning
                    ┌────┴────┐
                    ▼         ▼
                 Project    Global
                 Knowledge Knowledge
```

---

# 96. Final implementation rule

The most important rule for the team is:

> **Do not build a smarter translator. Build a better measurement system.**

If Rewrotron can accurately answer:

```text
What does this application actually expose?
What inputs does it accept?
What state does it depend on?
What does it change?
What does the user see?
What external systems does it call?
What behavior is deterministic?
What behavior is variable?
What code is actually required?
```

then the target implementation becomes a tractable engineering problem.

If those questions cannot be answered, generating more target code will not solve the migration problem.

The first goal is therefore not "convert CI3 to Laravel."

The first goal is:

> **Build a machine-verifiable behavioral model of a real CI3 application.**

Once that exists, Laravel is the first rendering of that model. Go, Rust and Python become additional renderers rather than entirely new migration systems.

---

# 97. Decision and readiness register

The [decision register](docs/11-roadmap/DECISION-REGISTER.md) records fixed engineering choices and external validation inputs with explicit owners/tasks. The repository owner's documentation request authorizes this specification review. It does not confirm any customer application's Gate 1/2/3 or certify runtime behavior.

Proceed through the development plan; missing fixture/runtime/customer/session/release inputs become concrete tracked blockers. No blanket “approve the entire architecture again” step is required to begin already-authorized safe engineering work.

# Transparent execution as a central product requirement

Behavior preservation and continuous user supervision are the two governing product promises. [EXECUTION-OBSERVABILITY](docs/06-agent/EXECUTION-OBSERVABILITY.md) and [ADR 0013](docs/01-architecture/ADR/0013-transparent-execution.md) own the shared execution stages/events/state, factual explanations, measured progress, timeline, return-session changes, decision cards, verbosity and safe persistent control. `internal/execution` projects existing operational records; CLI/MCP/future desktop consume the same committed facts. UI stages do not replace lifecycle states, and progress/alerts cannot grant approval or parity. This requirement belongs to the first engine milestone; runtime/client implementation is pending.

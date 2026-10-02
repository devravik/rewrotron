# Rewrotron — Implementation-Grade Project Plan

**Document status:** Team review draft
**Version:** 1.0
**Date:** 2026-10-03
**Repository model:** One open-source monorepo
**Primary implementation language:** Go
**Local persistence:** SQLite
**Primary agent integration:** MCP + Agent Skill + CLI
**Initial source framework:** CodeIgniter 3.x
**Initial target framework:** Laravel 13.x
**Future targets:** Go, Python, Rust, additional PHP frameworks

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

The implementation baseline should be pinned in CI. At project inception, Go 1.27.1 is the current stable patch release available in the official release history.

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

The MCP server should target the current stable MCP specification available at project inception, 2026-07-28, using only capabilities required by Rewrotron. The core implementation should remain useful through ordinary CLI execution even when an LLM client has no MCP support.

---

# 4. User workflow

The primary user workflow is:

```bash
rewrotron init /path/to/legacy-app
```

Then:

```bash
rewrotron analyze
rewrotron doctor
rewrotron observe
rewrotron discover
rewrotron plan
rewrotron migrate
rewrotron verify
rewrotron benchmark
rewrotron status
```

For agent clients, the user can simply expose the Rewrotron skill and/or MCP server.

The agent can then be instructed:

> Migrate this CodeIgniter application to Laravel using Rewrotron. Preserve externally observable behavior. Do not change production. Work only inside the configured sandbox.

The agent must then use Rewrotron's state rather than creating its own competing project ledger.

---

# 5. Repository architecture

Everything belongs in a single repository.

```text
rewrotron/
├── cmd/
│   ├── rewrotron/
│   └── rewrotron-mcp/
│
├── internal/
│   ├── project/
│   ├── config/
│   ├── policy/
│   ├── analyzer/
│   │   ├── php/
│   │   ├── routes/
│   │   ├── symbols/
│   │   ├── dependencies/
│   │   └── database/
│   ├── observer/
│   │   ├── http/
│   │   ├── ci3/
│   │   ├── database/
│   │   ├── session/
│   │   ├── filesystem/
│   │   └── external/
│   ├── behavior/
│   │   ├── graph/
│   │   ├── contracts/
│   │   ├── normalization/
│   │   └── workflows/
│   ├── migration/
│   │   ├── planner/
│   │   ├── mappings/
│   │   ├── generator/
│   │   └── adapters/
│   ├── verification/
│   │   ├── http/
│   │   ├── browser/
│   │   ├── database/
│   │   ├── sideeffects/
│   │   └── normalization/
│   ├── benchmark/
│   ├── knowledge/
│   ├── learning/
│   ├── skills/
│   ├── tools/
│   ├── artifacts/
│   └── events/
│
├── adapters/
│   ├── source/
│   │   └── codeigniter3/
│   └── target/
│       ├── laravel13/
│       ├── go/
│       ├── python/
│       └── rust/
│
├── tools/
│   ├── php/
│   ├── browser/
│   ├── database/
│   ├── diff/
│   └── fixtures/
│
├── skills/
│   └── rewrotron/
│       ├── SKILL.md
│       ├── references/
│       └── scripts/
│
├── schemas/
│   ├── project/
│   ├── behavior/
│   ├── migration/
│   ├── verification/
│   └── mcp/
│
├── migrations/
├── testdata/
│   ├── ci3-fixtures/
│   ├── laravel-fixtures/
│   └── compatibility-cases/
│
├── docs/
│   ├── architecture.md
│   ├── contracts.md
│   ├── adapters.md
│   ├── agent-integration.md
│   ├── contributing.md
│   └── security.md
│
├── Makefile
├── go.mod
├── README.md
└── LICENSE
```

Generated customer projects do not belong inside the Rewrotron source tree.

---

# 6. Project storage model

Each migrated application receives a self-contained `.rewrotron` directory.

```text
customer-app/
├── application/
├── system/
├── public/
└── .rewrotron/
    ├── project.yaml
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

`project.yaml` is human-readable metadata; SQLite is the authoritative operational store.

Example:

```yaml
version: 1
project_id: customer-portal
source:
  type: codeigniter3
  path: .
  runtime:
    php: auto

target:
  type: laravel
  version: "13.x"
  php: "8.3"

environment:
  mode: sandbox
  database:
    engine: mysql
    schema_policy: retain

compatibility:
  requirement: behavior_equivalent
  active_session_continuity: required
  exact_html: required_when_deterministic
  redirects: exact
  status_codes: exact

safety:
  production_access: denied
  external_writes: denied
  payment_actions: sandbox_only

agent:
  mode: vendor_agnostic
```

Any value not represented here is obtained from detected environment information or explicitly requested configuration. Rewrotron must not silently invent infrastructure details.

---

# 8. Lifecycle state machine

Every project has an explicit state.

```text
CREATED
  ↓
DIAGNOSED
  ↓
ANALYZED
  ↓
OBSERVATION_READY
  ↓
DISCOVERING
  ↓
MODELED
  ↓
CONTRACTS_READY
  ↓
MIGRATION_PLANNED
  ↓
MIGRATING
  ↓
VERIFYING
  ↓
REPAIRING
  ↓
PARITY_ACHIEVED
  ↓
OPTIMIZING
  ↓
FINAL_VERIFY
  ↓
READY_FOR_CUTOVER
```

Additional terminal or blocking states:

```text
BLOCKED
FAILED
NEEDS_HUMAN_REVIEW
ABANDONED
```

A project may return to an earlier state when new evidence invalidates an assumption.

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

If syntax cannot be parsed, the project enters `BLOCKED` for that portion rather than falling back to regex and pretending the analysis is complete.

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

Observed or strongly established as reachable.

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

Every migrated behavior slice gets a canonical contract.

Example:

```yaml
case_id: order-create-001
entrypoint:
  type: http
  method: POST
  path: /orders

auth:
  state: authenticated
  identity: test-user-a

input:
  content_type: application/x-www-form-urlencoded
  fields:
    product_id: integer
    quantity: integer

preconditions:
  product_exists: true
  inventory_available: true

response:
  status: 302
  redirect:
    exact_path: /orders/{created_order_id}

side_effects:
  database:
    inserts:
      - table: orders
    updates:
      - table: inventory
  session:
    flash_keys:
      - order_created

external_effects:
  email: none
```

The contract is target-framework independent.

---

# 15. Exact response policy

"Exactly the same" must be implemented as a layered comparison system.

## 15.1 Byte-exact

Use when values are deterministic.

Examples:

- status code;
- deterministic redirect path;
- static JSON values;
- deterministic HTML.

## 15.2 Canonical semantic comparison

Use for structured outputs where formatting may vary.

Examples:

- JSON object ordering;
- whitespace differences;
- SQL formatting;
- HTML formatting.

## 15.3 Dynamic-value normalization

Detect values that legitimately vary.

Examples:

```text
UUID
timestamp
nonce
CSRF token
request ID
session ID
random filename
```

A field becomes dynamic only when repeated source observations demonstrate variability.

The normalization rule must be stored in the contract.

## 15.4 Unknown

If Rewrotron cannot determine whether a difference is legitimate, the result is `UNKNOWN`, not `PASS`.

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

For every behavior case that touches the DB, capture:

```text
tables read
queries
parameters (secret-redacted)
transaction boundaries
rows inserted
rows updated
rows deleted
stored procedure calls
```

## 18.1 DB comparison policy

Do not require identical physical identifiers.

If the old system creates:

```text
order_id = 918
```

and the new system creates:

```text
order_id = 247
```

Rewrotron maps the logical entities and compares their business fields.

The mapping engine must support:

```text
source identity → target identity
```

through deterministic correlation rules.

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

# 30. Research/migration ledger

The project database must record every meaningful operation.

Core tables:

```text
projects
project_state
source_files
source_symbols
source_routes
source_components
assets
workflows
behavior_cases
contracts
observations
experiments
migrations
migration_units
mappings
diffs
repair_attempts
benchmarks
findings
lessons
skills
skill_runs
knowledge_refs
events
locks
```

---

# 31. Append-only event log

Every important state change produces an event.

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    entity_type TEXT,
    entity_id TEXT,
    payload_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    agent_session_id TEXT
);
```

Examples:

```text
PROJECT_CREATED
SOURCE_ANALYSIS_COMPLETED
ROUTE_DISCOVERED
WORKFLOW_DISCOVERED
CASE_CREATED
CASE_VERIFIED
MIGRATION_STARTED
MIGRATION_COMPLETED
DIFF_FOUND
DIFF_RESOLVED
SKILL_USED
LESSON_RECORDED
GLOBAL_PATTERN_CANDIDATE
```

Events are never mutated to rewrite history.

Derived state can be rebuilt from events plus immutable artifacts when practical.

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
Phase: VERIFYING

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

# 34. Mission/task system

Rewrotron must generate explicit missions rather than relying on the LLM to improvise the backlog.

Mission fields:

```text
mission_id
project_id
type
target_entity
objective
preconditions
dependencies
priority
estimated_cost
status
attempt_count
created_at
completed_at
```

Mission types:

```text
ANALYZE_SOURCE
DISCOVER_ROUTE
OBSERVE_WORKFLOW
GENERATE_CONTRACT
MIGRATE_UNIT
VERIFY_CASE
REPAIR_MISMATCH
ANALYZE_DEAD_CODE
BENCHMARK
OPTIMIZE
REVIEW_UNKNOWN
```

---

# 35. Next-task selection

The project manager should rank tasks using:

```text
compatibility impact
+ dependency readiness
+ behavior coverage gap
+ mismatch severity
+ migration progress
+ estimated effort
+ historical success of the required skill
```

This is a scheduling heuristic, not an LLM opinion.

The agent may override the next task only by recording a reason.

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

## 37.1 CLI

Every essential operation must be possible without an LLM.

Examples:

```bash
rewrotron status
rewrotron status --json
rewrotron analyze --json
rewrotron discover --json
rewrotron next --json
rewrotron verify --case order-create-001 --json
rewrotron report --json
```

## 37.2 JSON contract

Human-oriented CLI output is optional. Machine output is mandatory.

A machine-readable command must emit valid JSON to stdout and diagnostics to stderr.

## 37.3 MCP

The MCP server exposes semantic operations rather than raw SQL.

Do not expose:

```text
execute_sql
write_file_anywhere
shell_any_command
```

as Rewrotron-specific management tools.

Preferred tools:

```text
project.get_snapshot
project.get_status
project.pause
project.resume

research.get_next_mission
research.start_mission
research.complete_mission
research.search_history

analysis.get_routes
analysis.get_components
analysis.get_call_graph

behavior.get_case
behavior.create_case
behavior.record_observation

migration.get_plan
migration.get_unit
migration.record_mapping

verification.run_case
verification.get_mismatches
verification.classify_mismatch

knowledge.search_project
knowledge.search_global
knowledge.record_lesson

skills.search
skills.record_usage
skills.propose
skills.evaluate

reports.generate
reports.get
```

## 37.4 MCP resources

Where helpful, expose read-only resources such as:

```text
rewrotron://project/{id}/snapshot
rewrotron://project/{id}/coverage
rewrotron://project/{id}/contract/{case_id}
rewrotron://project/{id}/mismatch/{id}
```

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

SQLite is sufficient for the initial single-machine architecture.

Use file/project locks and DB transactions so two agents cannot accidentally mutate the same migration unit simultaneously.

Mission claiming must be atomic.

Example logical operation:

```text
BEGIN
select pending mission
lock mission
mark running
COMMIT
```

Only one worker owns a mission at a time.

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

Human approval is required at these gates by default:

1. enabling any production target;
2. approving destructive/external side effects;
3. accepting an `UNKNOWN` compatibility difference as intentional;
4. approving deletion of `DORMANT` or `UNKNOWN` code;
5. final production cutover.

Everything else can be automated.

---

# 53. CLI specification

Minimum command set:

```text
rewrotron init
rewrotron doctor
rewrotron analyze
rewrotron observe
rewrotron discover
rewrotron model
rewrotron plan
rewrotron migrate
rewrotron verify
rewrotron benchmark
rewrotron status
rewrotron next
rewrotron resume
rewrotron learn
rewrotron skills
rewrotron report
rewrotron export
```

Every command supports:

```text
--json
--project <path>
--quiet
--verbose
```

Where meaningful, mutating operations also support:

```text
--dry-run
```

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

The exact thresholds should be configurable, but the default v1 release gate is:

```text
100% of in-scope compatibility cases pass
0 unresolved critical mismatches
0 source components incorrectly deleted in benchmark corpus
100% of generated compatibility cases reproducible
100% of migrations produce a complete report
0 plaintext secrets in Rewrotron logs
0 unapproved production writes in automated tests
```

For an actual customer project, 100% refers to the agreed compatibility corpus and defined scope, not every theoretical input the application could possibly receive.

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

# 88. First sprint backlog

The first development sprint should create these concrete tickets.

## RWT-001 — Monorepo bootstrap

Create Go module, package structure, Makefile, CI.

Acceptance:

```text
make test
make build
```

pass on Linux.

## RWT-002 — Project initializer

Implement `.rewrotron/project.yaml` and SQLite initialization.

## RWT-003 — Event log

Implement append-only events and project reconstruction tests.

## RWT-004 — Artifact store

Implement content-addressed artifact storage.

## RWT-005 — CLI framework

Implement command routing and `--json` output.

## RWT-006 — Project snapshot

Implement `rewrotron status` and `rewrotron next`.

## RWT-007 — PHP runtime detector

Detect PHP version and Composer metadata from a CI3 project.

## RWT-008 — CI3 detector

Identify CI3 installation, application directory and configuration.

## RWT-009 — Route extractor

Extract static CI3 routes into the project database.

## RWT-010 — Source symbol index

Parse CI3 fixture and persist symbols/calls.

## RWT-011 — CI3 probe

Create runtime instrumentation package using hooks.

## RWT-012 — Request trace

Record request and response metadata.

## RWT-013 — DB trace

Record queries and transaction metadata.

## RWT-014 — Behavior case format

Implement canonical case schema.

## RWT-015 — Golden master

Create repeated-observation normalization.

## RWT-016 — Differential runner

Compare legacy and a deliberately modified target fixture.

## RWT-017 — Laravel bootstrap

Generate an empty Laravel 13 target.

## RWT-018 — First vertical slice

Migrate one real CI3 route into Laravel.

## RWT-019 — Compatibility test generator

Generate Laravel tests from canonical cases.

## RWT-020 — MCP server

Expose project snapshot and next-mission tools.

## RWT-021 — Rewrotron Skill

Create agent instructions for the first end-to-end migration.

---

# 89. First end-to-end demonstration

The first meaningful demonstration should be a small but real CI3 application, not a toy hello-world project.

Required flow:

```text
1. Point Rewrotron at CI3 repo
2. Detect framework/runtime
3. Analyze source
4. Discover routes
5. Boot sandbox
6. Login using test account
7. Record dashboard workflow
8. Record one create/update/read workflow
9. Capture DB effects
10. Generate compatibility cases
11. Bootstrap Laravel 13
12. Migrate one vertical slice
13. Run legacy and Laravel against same cases
14. Show exact/semantic comparison
15. Intentionally introduce a target mismatch
16. Let agent diagnose it
17. Repair it
18. Re-run verification
19. Produce migration report
```

If this demo works reliably, the core architecture is validated.

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

# 97. Team approval checklist

Before implementation starts, the team should explicitly approve:

```text
[ ] Single monorepo
[ ] Go core
[ ] SQLite project DB
[ ] Filesystem artifact store
[ ] CLI is canonical baseline
[ ] MCP is integration layer
[ ] Agent Skill is orchestration layer
[ ] CI3 is first source adapter
[ ] Laravel 13 is first target adapter
[ ] Existing DB retained in initial migration
[ ] Behavior IR is canonical
[ ] Vertical-slice migration
[ ] Differential testing is mandatory
[ ] Dead code requires evidence
[ ] Production access denied by default
[ ] Project and global knowledge separated
[ ] Learning is evidence-based
[ ] Generated tools are versioned/promoted
[ ] No provider-specific core logic
[ ] Apache-2.0 license
```

Once these are approved, development can begin with Phase 0 and the first 21 backlog items without requiring architectural decisions during implementation.

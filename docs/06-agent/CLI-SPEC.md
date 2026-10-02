# CLI-SPEC.md — Command Line Interface Specification (Machine & Human)

> **Authority Level:** Mandatory Machine Interface Contract (Critical Foundation Document #4)  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** `cmd/rewrotron`, External Agents, Automation Scripts  

---

## 1. Global CLI Invariants

1. **Standard Flags Supported on Every Command:**
   - `--json`: Format all output as a valid, single JSON envelope on `stdout`.
   - `--project <path>`: Explicit path to project directory containing `.rewrotron/` (defaults to current working directory).
   - `--quiet`: Suppress non-essential output.
   - `--verbose`: Enable detailed debug tracing on `stderr`.
   - `--dry-run`: Simulate mutating operations without persisting changes.
2. **Stdout vs. Stderr Separation:** When `--json` is active:
   - **`stdout`** is strictly reserved for the machine JSON response envelope.
   - **`stderr`** is used for progress bars, human diagnostics, and warning logs.
3. **Exit Code Conventions:**
   - `0`: Success / Passed.
   - `1`: Operational error (invalid arguments, missing files, doctor check failed).
   - `2`: Verification failure (differential test mismatch detected).
   - `3`: Human gate blocked (action requires manual approval).
   - `4`: User confirmation required (Gate 1 or Gate 2 pending confirmation).

---

## 2. Standard JSON Response Envelope

When `--json` is specified, all commands return this standard envelope (matching `specs/cli/cli-output.schema.json`):

```json
{
  "$schema": "https://rewrotron.dev/schemas/cli/cli-output.json",
  "version": "1.0",
  "command": "detect",
  "success": true,
  "exit_code": 0,
  "timestamp": "2026-10-03T05:00:00.123Z",
  "duration_ms": 142.8,
  "data": {},
  "errors": []
}
```

---

## 3. Subcommand Specifications

### 3.1 `rewrotron detect [path]`
- **Purpose:** Automatically inspect repository and detect language, framework, version, runtime, entrypoints, database, external integrations, confidence scores, and potential migration concerns.
- **Arguments:** `[path]` (Optional repository path; defaults to `.`).
- **Options:** `--fast`, `--json`.
- **JSON Data Example:**
  ```json
  {
    "language": "PHP",
    "framework": "CodeIgniter",
    "detected_version": "3.1.13",
    "php_compatibility": "likely PHP 7.4+",
    "application_type": "Server-rendered web application",
    "routes_count": 184,
    "controllers_count": 73,
    "models_count": 112,
    "views_count": 241,
    "database_engine": "MySQL",
    "external_integrations_count": 8,
    "authentication_type": "custom session-based authentication",
    "cron_entrypoints_count": 7,
    "confidence": {
      "framework": 0.99,
      "version": 0.96,
      "database": 0.91
    },
    "potential_concerns": [
      "dynamic controller loading",
      "custom hooks in application/config/hooks.php",
      "14 unresolved dynamic dependencies"
    ],
    "gate_status": "AWAITING_SOURCE_CONFIRMATION"
  }
  ```

### 3.2 `rewrotron confirm <gate>`
- **Purpose:** Record explicit user confirmation / overrides for Gate 1 (`source`) or Gate 2 (`plan`).
- **Subcommands / Arguments:**
  - `rewrotron confirm source`: Unlocks Gate 1.
    - Options: `--php <ver>`, `--framework <name>`, `--database <engine>`, `--override <key=val>`.
  - `rewrotron confirm plan`: Unlocks Gate 2.
    - Options: `--target-lang <lang>`, `--target-framework <fw>`, `--target-style <style>`.
- **Side Effects:** Writes persistent confirmation record to `user_confirmations` in `project.db`, computes and stores `configuration_hash`.

### 3.3 `rewrotron init [path]`
- **Purpose:** Initialize project workspace and `.rewrotron` directory. Automatically triggers detection.
- **Options:** `--source-type <type>`, `--target-type <type>`, `--auto-confirm` (requires explicit flag).

### 3.4 `rewrotron doctor`
- **Purpose:** Verify required system dependencies, compilers, runtimes, database connectivity, and container sandboxes.

### 3.5 `rewrotron analyze`
- **Purpose:** Parse Source ASTs, build symbol tables, extract call graphs, and emit **Source IR**.
- **Prerequisite:** Gate 1 (`SOURCE_CONFIRMED`). Fails with exit code 4 if Gate 1 is unconfirmed.

### 3.6 `rewrotron observe`
- **Purpose:** Instrument legacy application runtime and record execution traces.

### 3.7 `rewrotron discover`
- **Purpose:** Mine multi-step workflows and synthesize canonical **Application IR** and **Behavior IR** contracts.

### 3.8 `rewrotron plan`
- **Purpose:** Formulate vertical slice migration plan and proposed target architecture.
- **Options:** `--target <lang/framework>`, `--style <style>`, `--recalculate`.
- **Side Effects:** Generates proposed plan and sets project state to `PLAN_SYNTHESIZED` (Gate 2 pending).

### 3.9 `rewrotron migrate`
- **Purpose:** Invoke target adapter to generate modern target application code.
- **Prerequisite:** Gate 2 (`PLAN_CONFIRMED`). Fails with exit code 4 if Gate 2 is unconfirmed.
- **Options:** `--unit <unit_id>`, `--all`, `--dry-run`.

### 3.10 `rewrotron verify`
- **Purpose:** Execute automated differential verification between legacy source and modern target.

### 3.11 `rewrotron benchmark`
- **Purpose:** Execute load benchmarking and latency/memory comparison.

### 3.12 `rewrotron status`
- **Purpose:** Display migration progress, gate confirmation states, and coverage metrics.

### 3.13 `rewrotron next`
- **Purpose:** Query the Mission Manager for the next highest-priority autonomous agent task.

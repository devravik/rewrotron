# MCP-SPEC.md — Model Context Protocol Tools & Resources Specification

> **Authority Level:** Mandatory Machine Interface Contract (Critical Foundation Document #5)  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** `cmd/rewrotron-mcp`, MCP Clients (Claude Code, Cursor, Antigravity)  

---

## 1. Principles of MCP Tool Design

1. **Domain Operations Only:** Exposes high-level domain operations; never raw SQL or unrestricted shell execution.
2. **Mandatory Confirmation Gate Enforcement:** Mutating operations (`migration.generate_code`) verify that Gate 1 (`SOURCE_CONFIRMED`) and Gate 2 (`PLAN_CONFIRMED`) are recorded. If unconfirmed, the tool returns an error code `ERR_CONFIRMATION_REQUIRED`.
3. **Auditability:** Every tool invocation emits an event into `project.db`.

---

## 2. MCP Tools Catalog

### 2.1 `project.detect`
- **Description:** Runs automated environment and framework inspection on the source repository, returning language, framework, version, entrypoints, database, confidence scores, and potential concerns.
- **Input Schema:**
  ```json
  {
    "type": "object",
    "properties": {
      "path": { "type": "string", "description": "Repository path (defaults to root)" }
    }
  }
  ```
- **Output Schema:** Structured detection findings with confidence metrics and `gate_status`.
- **Permissions:** Read-only. Idempotent: Yes. Side effects: Records detection in `source_detections`.

### 2.2 `project.confirm_source` (Gate 1)
- **Description:** Records explicit human confirmation / overrides for the detected source application baseline. Unlocks Gate 1.
- **Input Schema:**
  ```json
  {
    "type": "object",
    "required": ["confirmed_framework"],
    "properties": {
      "confirmed_framework": { "type": "string" },
      "runtime_version": { "type": "string" },
      "database_engine": { "type": "string" },
      "overrides": { "type": "object", "additionalProperties": { "type": "string" } }
    }
  }
  ```
- **Output Schema:**
  ```json
  {
    "confirmed": true,
    "gate": "GATE_1_SOURCE_DETECTION",
    "status": "SOURCE_CONFIRMED",
    "configuration_hash": "string"
  }
  ```
- **Permissions:** Read/Write. Idempotent: Yes. Side effects: Inserts into `user_confirmations`.

### 2.3 `project.confirm_plan` (Gate 2)
- **Description:** Records explicit human confirmation of target stack, architectural style, and migration slices. Unlocks Gate 2.
- **Input Schema:**
  ```json
  {
    "type": "object",
    "required": ["target_language", "target_framework"],
    "properties": {
      "target_language": { "type": "string" },
      "target_framework": { "type": "string" },
      "architecture_style": { "type": "string", "default": "modular_monolith" },
      "approved_unit_ids": { "type": "array", "items": { "type": "string" } }
    }
  }
  ```
- **Output Schema:**
  ```json
  {
    "confirmed": true,
    "gate": "GATE_2_PLAN_CONFIRMED",
    "status": "PLAN_CONFIRMED"
  }
  ```
- **Permissions:** Read/Write. Idempotent: Yes. Side effects: Inserts into `user_confirmations`.

### 2.4 `project.get_snapshot`
- **Description:** Returns the overall project status, gate states, configuration, phase, coverage metrics, and health checks.
- **Output Schema:**
  ```json
  {
    "project_id": "string",
    "phase": "string",
    "gate_states": {
      "gate_1_source": "CONFIRMED",
      "gate_2_plan": "CONFIRMED",
      "gate_3_cutover": "PENDING"
    },
    "source_framework": "string",
    "target_language": "string",
    "target_framework": "string",
    "overall_compatibility": "number"
  }
  ```

### 2.5 `research.get_next_mission`
- **Description:** Fetches the next highest-priority task assigned to the AI agent by the Mission Manager.
- **Output Schema:** Canonical `Mission` object.

### 2.6 `research.complete_mission`
- **Description:** Marks an active mission as completed with an optional outcome summary.

### 2.7 `behavior.get_case`
- **Description:** Retrieves the complete specification of a Behavior Case including preconditions, inputs, and expected outputs.

### 2.8 `verification.run_case`
- **Description:** Executes differential verification for a specific case against both source and target environments.

### 2.9 `knowledge.record_lesson`
- **Description:** Persists an empirical diagnostic lesson learned during a repair attempt.

### 2.10 `skills.search`
- **Description:** Searches available migration and repair skills by keyword or category.

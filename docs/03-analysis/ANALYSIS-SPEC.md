# ANALYSIS-SPEC.md — Unified Static Analysis & Code Classification Specification

> **Authority Level:** Mandatory Analysis Contract  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** `internal/analyzer`, `internal/observer`, Code Classifier  

---

## 1. Scope & Objective

The Static Analysis Engine reverse-engineers legacy application source code into a structured, relational inventory of reachable symbols, routes, entrypoints, and data access patterns.

Static analysis provides the **structural hypothesis** of how the legacy application is built. This hypothesis is subsequent confirmed, enriched, or corrected by runtime observation.

---

## 2. Extracted Entities & Inventory

During `rewrotron analyze`, the engine extracts and persists into SQLite:

1. **Source Files (`source_files`):** Path, file hash (SHA-256), lines of code, byte size, file type (`controller`, `model`, `view`, `config`, `helper`, `library`, `migration`, `asset`).
2. **Symbols (`source_symbols`):** Classes, interfaces, traits, methods, functions, constants, class properties.
3. **Calls & Call Graph:** Caller-callee edges between methods and functions, including instantiated classes and static invocations.
4. **Declared & Implicit Routes:**
   - Routes declared in routing configuration (`application/config/routes.php`).
   - Implicit convention-based routes (`class/method/param` URI matching in CodeIgniter).
5. **Hooks & Interceptors:** Framework lifecycle hooks (e.g. `pre_controller`, `post_controller_constructor`, `post_controller`, `post_system`).
6. **Dependencies:** Composer packages, third-party libraries bundled in `application/libraries/` or `system/`.
7. **Models & Queries:** Active Record calls (`$this->db->get()`, `$this->db->insert()`), raw SQL strings, query builder invocations.
8. **Views & Templates:** View render calls (`$this->load->view('name', $data)`), partial inclusions, inline PHP scripts.
9. **External Service Invocations:** cURL calls, Guzzle clients, file uploads, raw socket connections, email dispatches (`$this->email->send()`).
10. **Background & Batch Entrypoints:** CLI controller commands, cron scripts, worker loops.

---

## 3. The 4-Tier Code Classification Taxonomy

To ensure agents never delete code without empirical proof, every symbol, method, and file MUST be classified into exactly one of four objective states:

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        CODE CLASSIFICATION TAXONOMY                    │
├────────────────────────────────────────────────────────────────────────┤
│ 1. ACTIVE                                                              │
│    • Statically reachable from a known entrypoint (HTTP, CLI, Cron)    │
│    • AND/OR observed executing during runtime observation traces.       │
│    Action: MUST be reconstructed in Target Application.                │
├────────────────────────────────────────────────────────────────────────┤
│ 2. DORMANT                                                             │
│    • Statically reachable from an entrypoint                           │
│    • BUT NEVER observed executing across all recorded runtime traces.   │
│    Action: Reconstruct in Target OR flag for human review; NEVER drop.  │
├────────────────────────────────────────────────────────────────────────┤
│ 3. DEAD                                                                │
│    • Proven STATICALLY UNREACHABLE from any public route, hook, CLI,   │
│      or cron entrypoint.                                               │
│    • AND has ZERO executions across all runtime observation traces.     │
│    • AND no dynamic dispatch (e.g. $this->$variable()) targets it.     │
│    Action: SAFELY OMIT from Target Application. Log reachability proof.│
├────────────────────────────────────────────────────────────────────────┤
│ 4. UNKNOWN                                                             │
│    • Reachability cannot be mathematically proven (e.g. uses eval(),    │
│      dynamic method calls $$method(), reflection, or plugin hooks).     │
│    Action: MUST be preserved or escalated to human review gate.         │
│            NEVER automatically deleted.                                │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Objective Classification Rules

An AI agent or analyzer algorithm MUST apply these strict boolean rules when classifying a symbol `S`:

```text
Is S executed in any Runtime Observation Trace?
  ├── YES ──► ACTIVE
  └── NO
        ├── Is S reachable via static call graph from any Entrypoint?
        │     ├── YES ──► DORMANT
        │     └── NO
        │           ├── Does the codebase contain dynamic calls ($this->$name)
        │           │   or eval() that could resolve to S?
        │           │     ├── YES ──► UNKNOWN
        │           │     └── NO  ──► DEAD
```

### Omission Audit Rule
When code classified as `DEAD` is omitted from the target application, Rewrotron generates an **Omission Evidence Record** in `.rewrotron/artifacts/reports/omissions.json` containing:
- Symbol name and file location.
- SHA-256 hash of the legacy file.
- Static graph traverse proof showing lack of inward edges from all entrypoints.
- Runtime observation verification log demonstrating zero hits.

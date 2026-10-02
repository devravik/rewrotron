# RUNTIME-OBSERVATION.md — Runtime Instrumentation & Trace Specification

> **Authority Level:** Mandatory Analysis Contract  
> **Status:** Authoritative Core Specification  
> **Subsystem:** `internal/observer`, CodeIgniter 3 Instrumentation  

---

## 1. Overview & Execution Pipeline

Runtime observation dynamically instruments the legacy application to record ground-truth execution timelines under real or synthetic traffic.

```text
Incoming Request
    ↓
[Observer Reverse Proxy / Entry Interceptor] (Generates X-Rewrotron-Trace-ID)
    ↓
[CI3 pre_controller Hook] (Captures URI, HTTP Verb, Headers, Session State)
    ↓
[Controller Execution]
    ↓
[CI3 DB Driver Interceptor] (Captures SQL queries, bound params, execution time)
    ↓
[CI3 Cache / Session Drivers] (Captures read/write cache operations & flash data)
    ↓
[External HTTP Interceptor] (Mocks / captures outbound cURL or Guzzle calls)
    ↓
[CI3 post_controller / display_override] (Captures rendered HTML / JSON & Status)
    ↓
Outgoing Response + Persisted Observation Trace JSON
```

---

## 2. What Gets Captured & Where

| Phase / Component | Capture Location | Captured Metadata |
| :--- | :--- | :--- |
| **HTTP Request** | Reverse Proxy / Front Controller | Method, URL, headers, query params, form body, raw body, client IP |
| **Routing & Dispatch** | `pre_controller` hook | Resolved controller class, method, URI segments |
| **Session Baseline** | Controller constructor | Initial session variables, cookie IDs, authentication tokens |
| **Database Queries** | `CI_DB_driver` wrapper | Exact SQL query text, parameter bindings, duration (μs), affected rows |
| **Session Mutations** | `post_controller` hook | Set/unset session keys, flash data added |
| **External Requests** | Mock HTTP handler | Outbound URLs, request payloads, simulated or observed responses |
| **HTTP Response** | `display_override` hook | Status code, response headers, response body (or artifact reference) |

---

## 3. Canonical Trace Schema

Every observation produces a structured JSON record matching `specs/schemas/trace-schema.json`:

```json
{
  "trace_id": "trc_01HZX9876543210ABCDEFG",
  "project_id": "customer-portal",
  "environment": "source_legacy",
  "timestamp": "2026-10-03T04:45:00.123456Z",
  "duration_ms": 42.5,
  "request": {
    "method": "POST",
    "path": "/api/v1/cart/add",
    "query": {},
    "headers": {
      "Content-Type": "application/x-www-form-urlencoded",
      "User-Agent": "Rewrotron-Observer/1.0"
    },
    "body_redacted": false,
    "body_ref": "artifacts/traces/trc_01HZX9876543210ABCDEFG_req.bin"
  },
  "route": {
    "matched_pattern": "api/v1/cart/add",
    "controller": "Cart_api",
    "action": "add_item"
  },
  "session": {
    "initial": {
      "user_id": 104,
      "cart_token": "tok_abc123"
    },
    "mutations": [
      {
        "op": "SET",
        "key": "cart_count",
        "value": 3
      }
    ]
  },
  "database": {
    "queries": [
      {
        "query_index": 0,
        "sql": "SELECT * FROM products WHERE id = ? AND active = 1",
        "bindings": [42],
        "duration_us": 1240,
        "rows_returned": 1
      },
      {
        "query_index": 1,
        "sql": "INSERT INTO cart_items (cart_id, product_id, qty) VALUES (?, ?, ?)",
        "bindings": [12, 42, 1],
        "duration_us": 2310,
        "rows_affected": 1
      }
    ]
  },
  "external_calls": [],
  "response": {
    "status": 200,
    "headers": {
      "Content-Type": "application/json; charset=UTF-8"
    },
    "body_ref": "artifacts/traces/trc_01HZX9876543210ABCDEFG_resp.bin",
    "body_sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  }
}
```

---

## 4. Correlation IDs & Distributed Tracing

1. **`X-Rewrotron-Trace-ID`**: Injected into every incoming request header. Propagated through PHP environment (`$_SERVER['HTTP_X_REWROTRON_TRACE_ID']`).
2. **Database Query Tagging**: Injected as an SQL comment prepended to queries:
   ```sql
   /* rewrotron_trace_id: trc_01HZX9876543210ABCDEFG */ SELECT * FROM users ...
   ```
3. **Artifact Linking**: Request and response bodies exceeding 64KB are streamed directly into `.rewrotron/artifacts/traces/` and referenced by file path hash.

---

## 5. Security & Sensitive Data Redaction

Prior to persisting traces to SQLite or the artifact store, the redaction engine sanitizes:
- **Sensitive Headers:** `Authorization`, `Cookie`, `X-Api-Key`.
- **Sensitive Body Fields:** Fields matching `(?i)(password|secret|credit_card|cvv|token|api_key|ssn)`.
- **Masking Format:** Replaced with `[REDACTED_BY_REWROTRON]`.

---

## 6. Repeat Observation & Nondeterminism Detection

To detect dynamic variability vs static expectations:
1. **Replay Count:** Every observation case is replayed at least 2 times against the legacy application under identical initial database seeds.
2. **Variability Differential:** If the legacy response changes (e.g. CSRF token changes, timestamp shifts, UUID changes), the field is flagged in the `BehaviorContract` as **`DYNAMIC_VARIABLE`** and assigned a normalization rule.

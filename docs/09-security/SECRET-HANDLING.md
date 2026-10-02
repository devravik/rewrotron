# SECRET-HANDLING.md — Credential Masking & Storage Protection

> **Version:** 1.0  
> **Status:** Security Reference Guide  

---

## 1. Automated Secret Scrubbing Pipeline

Before any trace, observation, or error log is written to SQLite or the artifact store, it passes through the `internal/artifacts/redact` filter:

```text
Raw HTTP Request / SQL Query / Trace
              │
              ▼
    Regex & AST Secret Matcher
  - Password fields
  - Bearer / Basic Auth headers
  - Database connection strings
  - Private RSA / PGP keys
              │
              ▼
   Masked with `[REDACTED_BY_REWROTRON]`
              │
              ▼
  Persisted to .rewrotron/
```

This guarantees that developer machines, git commits, or shared experiment logs never leak customer credentials.

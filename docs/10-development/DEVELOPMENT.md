# DEVELOPMENT.md — Local Development Setup & Tooling

> **Version:** 1.0  
> **Status:** Development Reference Guide  

---

## 1. Prerequisites

- **Go:** 1.22+ installed and available on PATH.
- **PHP:** 8.2 or 8.3 with standard extensions (pdo, mbstring, sqlite3, curl).
- **Composer:** Modern Composer 2.x for PHP test dependencies.
- **Docker:** (Optional but recommended) for containerized sandboxing.

---

## 2. Standard Make Commands

```bash
# Compile binaries (rewrotron & rewrotron-mcp)
make build

# Run complete test suite with race detector
make test

# Run linter
make lint

# Validate all JSON schemas against fixtures
make test-schemas

# Run end-to-end reference fixture migration
make test-e2e
```

# TESTING.md — Testing Pyramid & Quality Gate Standards

> **Authority Level:** Mandatory Development Quality Specification  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** All Go packages, CI/CD, Contributor PRs  

---

## 1. The Rewrotron Testing Pyramid

Rewrotron's mission is proving behavioral equivalence; its own codebase must maintain immaculate testing standards.

```text
                  ▲
                 / \
                /   \     8. Global Benchmark Corpus (Latency, Memory)
               /     \
              /───────\   7. Full Project End-to-End Migrations
             /         \
            /───────────\ 6. Differential Verifier Integration Suites
           /             \
          /───────────────\ 5. Machine Contract Tests (specs/cli, specs/mcp)
         /                 \
        /───────────────────\ 4. Golden Master Fixture Replays (testdata/)
       /                     \
      /───────────────────────\ 3. Storage Layer & Migration Compatibility
     /                         \
    /───────────────────────────\ 2. Analyzer & AST Parsing Tests
   /                             \
  /───────────────────────────────\ 1. Pure Unit Tests (internal/*_test.go)
```

---

## 2. Test Category Definitions

1. **Unit Tests:** High-speed, isolated Go tests testing individual functions, normalizers, hashers, and state machine transitions. Zero network, zero external processes.
2. **Analyzer Tests:** AST fixture tests validating that PHP constructs (classes, traits, dynamic calls, complex queries) are accurately parsed into symbol graphs.
3. **Storage & Migration Tests:** Validates that SQLite schemas initialize cleanly and upgrade without data loss across versions in `migrations/`.
4. **Fixture Replays:** Verifies that synthetic CodeIgniter 3 test applications in `testdata/ci3-fixtures/` are correctly observed and modeled.
5. **Contract Tests:** Asserts that every CLI command `--json` output and MCP server response strictly validates against its JSON Schema in `specs/`.
6. **Differential Verifier Tests:** Runs the comparator against known identical and divergent traces to confirm accurate categorization (`PASS`, `EXPECTED_CHANGE`, `DEFECT`, `UNKNOWN`).
7. **End-to-End Migration Suites:** Runs complete end-to-end migrations on reference benchmark applications (e.g. CI3 sample e-commerce store -> Laravel 13).
8. **Benchmarks:** Profiling test execution to prevent memory leaks and throughput regressions.

---

## 3. Pre-Merge Verification Checklist

A pull request or feature branch is considered complete **ONLY IF**:
- [x] `go test -race ./...` passes with zero race conditions.
- [x] Code coverage across `internal/` exceeds 85%.
- [x] `golangci-lint run` reports zero warnings.
- [x] All schema contracts in `specs/` validate against test fixtures.
- [x] An ADR is added in `docs/01-architecture/ADR/` for any new architectural pattern.

# Development setup and available checks

## Current repository

The repository is a specification baseline. There is no `go.mod`, Makefile, Go implementation, container fixture corpus or engine binary yet. The initial SQL migration defines the store to build; it is not a deployed customer DB. Start with M0 in [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md).

For documentation/schema checks now:

```bash
python3 -m pip install 'jsonschema>=4.10,<5'
python3 tools/docs/validate.py
```

The validator uses local schema IDs, checks fixtures/registries/links and executes the incremental SQLite DDL and approval hash vectors. This does not establish Go driver, race, framework runtime or migration correctness.

## Planned implementation prerequisites

Go floor is 1.25.0 for the specified modernc.org/sqlite v1.46.2/libc v1.70.0 pin; M0 records the exact toolchain/module/checksums and verifies embedded SQLite 3.51.3. A dependency requiring a later Go version raises the build baseline explicitly; do not silently use an incompatible minimum. Use `modernc.org/sqlite`, standard-library CLI parsing, and the selected Go PHP parser inside the adapter. Pin exact versions and checksums in `go.mod/go.sum` and a tool manifest.

Source PHP is detected/pinned per fixture/application. Target PHP is 8.3+ for Laravel 13; source and target runtimes are independent. Composer 2 and a container runtime/image cache are required for runtime fixture tasks. No cloud account or LLM subscription is needed. Go-only tests cannot require PHP/containers/network.

## Make targets to create in M0

| Target | Command/behavior |
| --- | --- |
| `make build` | Build CLI/MCP binaries with injected version metadata |
| `make test` | `go test -race ./...` for deterministic tests |
| `make lint` | `go vet ./...` and pinned `golangci-lint run` |
| `make test-schemas` | Offline schema/registry/example/document checks |
| `make test-fixtures` | Pinned container source observation corpus |
| `make test-e2e` | Gate-aware CI3→Laravel fixture migrations |
| `make benchmark-fixtures` | Reproducible separate measurement workflow |

Runtime tests report an explicit skip only in a named development profile; release CI must provision and execute them. Every command's prerequisites and remediation are recorded by doctor. Shared fixture images/dependencies are cached for offline replay.

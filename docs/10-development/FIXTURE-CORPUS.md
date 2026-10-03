# Fixture corpus and acceptance matrix

All fixtures are synthetic, redistribution-safe, versioned and free of real customer credentials. RWT-007 establishes the source archive checksum/license, source/runtime images, DB settings/schema, startup commands and Composer locks. Pin image digests and exact versions before any runtime claims.

| Fixture path | Required behavior and negative mutation |
| --- | --- |
| `testdata/ci3-fixtures/routing` | Config/verb/convention routes, regex precedence, default/404, `_remap`, nested controllers; deliberately swap route order |
| `testdata/ci3-fixtures/dynamic` | Dynamic loaders/includes, partial parse, custom bases, unreachable cycle, cron-only roots; falsely classify a dynamic target DEAD |
| `testdata/ci3-fixtures/read-json` | Deterministic JSON and authorization failure; alter number/type/status |
| `testdata/ci3-fixtures/auth-session` | Login/roles, cookies, CSRF, flash, rotation/logout/active source session; switch principal or ignore invalid token |
| `testdata/ci3-fixtures/crud-workflow` | Multi-step write/read, transaction rollback, decimal/null/timezone/generated IDs; skip audit insert or break FK mapping |
| `testdata/ci3-fixtures/hooks-effects` | Existing hooks, early exits/cache, HTTP/mail sink; duplicate side effect or lose capture |
| `testdata/ci3-fixtures/views-files` | Forms/assets/whitespace, upload/download; remove input or change binary bytes |
| `testdata/ci3-fixtures/realistic-portal` | Combined custom controllers/dynamic loading/auth/DB/effects/CLI inventory; exercise complete report and limitations |
| `testdata/compatibility-cases` | Canonical contracts/cases and expected trace/normalization verdicts |
| `testdata/contracts` | CLI/MCP success/error/gate/empty-queue/cancellation/dry-run schemas |
| `testdata/storage` | Migration versions, invalid relations, immutable records, contention/recovery checkpoints |

Source instrumentation control compares uninstrumented/instrumented HTTP/state behavior. Every trace identifies required capture dimensions; missing probes produce UNKNOWN. Include deliberate regressions in every comparator dimension rather than tests that only echo implementation behavior.

Keep HTTP and non-HTTP inventory visible. The first HTTP replay engine marks unsupported command/consumer cases explicitly; full release support is limited to capability-tested operations. A combined fixture with unsupported cron cannot be reported as whole-app equivalent without a reviewed exclusion.

Runtime corpus pass evidence contains seed/schema/image/lock/probe hashes and cleanup outcome. A first read-json pass is a smoke slice; auth and transactional workflows are required before BETA. Stable release needs the full published supported matrix, two agent/client integrations, restart recovery, security checks and release runbooks.

# Falsification and evidence requirements

**Owner:** Acceptance experiment registry. All X01–X40 are specified experiments, NOT executed runtime results. [FIXTURE-CORPUS](FIXTURE-CORPUS.md) references this registry; milestone assignments are in [DEVELOPMENT-PLAN](../11-roadmap/DEVELOPMENT-PLAN.md).

The core question is whether the engine detects a plausible WRONG reconstruction. Positive targets alone cannot establish verifier quality. Each test records approved subject/corpus/policy/target/profile, deliberate fault, independent expected failure reason, actual verdict and evidence. A mutant rejected for unrelated setup failure is not a semantic kill. Unknown capture must produce UNKNOWN; known behavior defects must MISMATCH. No passing corpus proves its own completeness.

| ID | Subject | Falsification obligation | Milestone |
| --- | --- | --- | --- |
| X01 | Approval identity | Reconstruct subject/hash after restart and adapter registry upgrade; shuffle keys; reject float/duplicate/invalid order; changed adapter bytes must require successor approval | M0 |
| X02 | Coherent source | Mutate and ABA files during scan, swap symlink, add untracked generated source; reject live mixed baseline or show sealed consistent bytes | M0 |
| X03 | Executable semantics | Unsupported operator, wrong type, missing/null/empty distinction, overflow, unknown guard must reject/block, never run arbitrary JSON or fall through | M1 |
| X04 | Scenario branching | Two actors, inventory/payment/account/flag branches, explicit terminal state and async bound; wrong branch/state must mismatch; unbounded path unsupported | M1 |
| X05 | Missing write | Good response without required DB INSERT/UPDATE must mismatch despite status/body equality | M1 |
| X06 | Wrong principal | Write/respond for another account with plausible ID must mismatch across actor/DB/response/effect relations | M1 |
| X07 | Duplicate effect | Worker accepts effect then crashes before ack; duplicate email/payment simulator receipt must fail effect cardinality/idempotency | M1 |
| X08 | Incomplete observer | Remove required mechanism/control/connection/causality or truncate capture; required dimension must UNKNOWN, not empty PASS | M1 |
| X09 | Dynamic and cron roots | Dynamic include/class/method/reflection/autoload/database dispatch, dormant admin/plugin/serialized class and missed cron must prevent false DEAD | M2 |
| X10 | Modified environment | Vendor/framework patch, mixed versions, production flags/settings and assertions contradicting measurements retain separate facts/gaps | M2 |
| X11 | Probe bypass | PDO outside framework, native IO/raw socket and alternate connection escape must invalidate affected completeness | M2 |
| X12 | Async email | HTTP success before queued delivery must wait for causal ack/quiescence; missing/late/unrelated delivery fails | M2 |
| X13 | Runtime acquisition | Missing PHP extension/image/platform or incompatible lock blocks safe boot; no host/older-runtime fallback | M2 |
| X14 | Reset dimensions | DDL/implicit commits/temp tables/nontransactional rows/sequences/multi-DB identity/Redis TTL-lock-queue contamination must fail reset or use fresh instance | M2 |
| X15 | Exact resolution | 13.*, latest, range or missing runtime/distribution rejected as resolution; standard-library framework=null supported | M3 |
| X16 | User decisions | Reject recommendation, supply alternate adapter, request evidence; proposed exception cannot grant PASS/safety waiver; only exact Gate 2 activates | M3 |
| X17 | Decision immutability | PLANNED→VERIFYING→VERIFIED progress leaves plan hash unchanged; strategy/lowering changes invalidate approval | M3 |
| X18 | Recursive dependencies | Replacement has incompatible transitive/native/license dependency; apparent API compatibility changes retries; block/characterize rather than upgrade blindly | M3 |
| X19 | Session/CSRF lifecycle | Rotation/expiry/logout/reset/remember-me/throttle/cookie attrs/cross-user/replayed CSRF/concurrent sessions and unsafe serialization fail without supported safe bridge | M4 |
| X20 | Identity normalizers | Broken foreign key/URL/session principal or token lifetime, business refund clock boundary, meaningful random order must survive normalization as mismatch | M4 |
| X21 | Lossless inputs | Absent versus empty/raw/repeated query, blank header, 9007199254740993 numeric ID, decimal amount and JSON absence/null remain distinct end-to-end | M4 |
| X22 | Duplicate/shared work | Two agents submit duplicate work or edit shared auth file; unique active work/resource ownership prevents duplicate admission/lost edits | M4 |
| X23 | Fenced publication | Old owner pauses after fence check, expires and resumes; old process cannot write after resource transfer; shared/disjoint conflicts handled under serial baseline | M4 |
| X24 | Idempotency/resume | Crash after commit before response returns existing result; stale project/incident/lease cannot publish; trivial context edit cannot reset three-attempt budget | M5 |
| X25 | Ownership isolation | Cross-project artifacts/operations/missions/token/source refs fail service and SQL checks; independent workspace/private data remain isolated | M0 |
| X26 | Process lifetime | Kill foreground owner, kill MCP client with running supervisor, orphan/PID reuse/reconnect, restart human pause; correct owner survival/cleanup and no invented progress | M1/M5 |
| X27 | Execution truth | Read old through_sequence while rows advance: current N distinct from event M; event store failure stops admission; heartbeat not progress; unread counts exclude inspection | M1/M5 |
| X28 | Storage recovery | Apply migrations and corrupted/checksum/version cases, WAL checkpoint/reader contention, coordinated DB/artifact backup during publication and missing object restore; no valid incomplete backup | M0/M6 |
| X29 | Typed invalidation | PHP8.3→8.4 at 70%: preserve unchanged source discovery, invalidate relevant dependency/target/verification; unknown edge conservative; no old PASS relabel | M3 |
| X30 | Quota/privacy | Disk full in capture/publication/backup preserves protected subjects; secret canaries in AST/diagnostic/DOM/screenshot/log/generated source never escape approved boundary | M0 |
| X31 | Vendor-neutral transport | Same project resumed by two named clients and CLI-only agent; numeric/framing/JSON/error/gate parity; unsupported MCP negotiation honest | M5 |
| X32 | Trust boundary | Same-account hostile file edit is detected when mediated service revalidates, but cannot be advertised as prevented OS access; sandbox egress controls executed | M5 |
| X33 | Statistical claim | Identical noisy trials, insufficient p99 samples, zero denominator, known slowdown and warm/cold order effects yield correct explicit claim/inconclusive outcome | M6 |
| X34 | Write-aware rollback | Target changes sessions/orders/payment simulator state before rollback; source continuity/data reconciliation tested or release blocks | M6 |
| X35 | Scope deadlock | Login prerequisite completes on its UNIT+implemented regressions before checkout; project missing checkout cannot VERIFIED/READY_CUTOVER | M1/M6 |
| X36 | Distribution | Native supported OS path/process/fsync/driver/image controls and licenses; cross-compilation cannot confer runtime maturity | M6/M8 |
| X37 | Contaminated learning | Three correlated generated fixtures with same wrong oracle and holdout failure cannot promote | M7 |
| X38 | Low-quality capability | Experimental tool drops hard traces; apparent success improves but quality/unknown/capture/mutation denominators reject activation | M7 |
| X39 | Retraction | Pattern/skill/tool revision retracted; all current consumer artifacts/results discovered and stale; old pinned approvals/history retained; promotion crash reconciles | M7 |
| X40 | Neutral representations | Django→Go, Laravel→Rust, custom PHP→Python structural/effect/async fixtures preserve uncertainty and fail incompatible PHP-only lowering; unsupported protocol explicit | M8 |

## Execution profiles and claims

Documentation-only: `python3 tools/docs/validate.py` checks local schemas/refs/catalogs/examples/hash vectors and incremental SQLite DDL/invariants. It does not run the Go canonicalizer/driver/race/adapter/runtime/client/security/performance experiments.

Implementation: pure deterministic tests with injected clocks/IDs/ports, SQLite/artifact integration, actual `go test -race ./...`, vet/pinned lint, contract compatibility, sandbox controls, deliberate target/comparator mutants and independent end-to-end holdouts. Native/release runtime environments cannot be skipped and called validated. Capture quality/comparator negative tests precede normalizer/tool promotion.

Every result names versions/checksums, exact commands/exit codes, executed fixture/profile and limitations. Maturity transitions require evidence. Unsupported mechanisms remain blocked/UNKNOWN; fields/assertions may not be invented to make fixtures pass. No test is pre-marked completed by this specification.

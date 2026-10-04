# Reference adapter compatibility map

The [canonical compatibility specification](COMPATIBILITY-SPEC.md) owns verdicts and comparison rules. This table covers Reference Profile #1 (CI3 → Laravel), currently SPECIFIED. Framework APIs are implementation choices rather than evidence of compatibility; other profiles use the same independent obligations/verdicts with their own qualified providers.

| Behavior | CI3 extraction | Laravel implementation | Required evidence |
| --- | --- | --- | --- |
| Routing | Config, verb routes, conventions, overrides | Explicit ordered routes | Method/path/precedence and fallback cases |
| Input validation | Actual source validation/coercion | Requests/actions preserving observed errors | Accepted/rejected payloads and exact status/messages |
| Sessions | Detected driver, cookie settings, userdata/flash | Explicit compatible store/bridge | Principal continuity, rotation, flags and logout |
| DB effects | Transactions, query traces and row snapshots | Query builder/model with exact schema mapping | Committed state and rollback |
| HTML | Source rendered pages and assets | Blade/output preserving contract | Text/forms/assets and selected DOM or exact bytes |
| CSRF | Detected settings and invalid-token behavior | Explicit middleware compatible with case | Missing/invalid/replayed tokens and lifecycle |

Neither Eloquent defaults nor Laravel validation defaults automatically match a source. Session interoperability is specified in [AUTH-AND-SESSION](../04-migration/AUTH-AND-SESSION.md). Unknown framework extensions block the affected capability.

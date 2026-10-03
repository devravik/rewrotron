# Laravel 13 target adapter

**Validation maturity:** SPECIFIED. **Distribution tier:** proposed EXPERIMENTAL after implementation; no runtime evidence exists. Path: `adapters/target/php/laravel13`. Laravel 13 requires PHP 8.3+, according to the [release policy](https://laravel.com/framework/docs/releases). Pin the runtime patch, framework patch, dependencies and Composer lock in every generated target.

## Inputs and capability validation

Consume v2 executable Behavior IR contracts, data models, exact resolved environment, approved Target Implementation Model and plan. Generation does not translate CI3 files. Supported rule/entrypoint/effect capabilities are declared in the adapter manifest. Unknown computations, drivers, middleware semantics, non-HTTP consumers or session formats block their unit rather than becoming fabricated implementations.

Generate a modular monolith by default: explicit routes, lean controllers, actions/services, Requests/Policies only when they preserve actual behavior, schema-aware DB access, views/assets and compatibility tests. Form Requests can otherwise change status/error formats; default middleware can change CSRF, redirects, rate behavior and sessions. Choose and test each deliberately.

## Preserved database

No schema/engine migration is executed in the initial generation profile. Models declare exact tables, primary keys, timestamp behavior, casts and relationships from discovery. Eloquent is optional: query builder or explicit SQL inside the generated application may be necessary to preserve behavior. The engine/MCP never offers unrestricted SQL execution.

Use independent sandbox DB copies. Preserve transaction boundaries, decimal arithmetic, null/coercion, timezone, triggers and affected row semantics. Generated migration files are fixture/bootstrap definitions when needed, never commands to mutate a production schema.

## Safe target publication

Bootstrap and slice generation occur only after Gate 2. The adapter emits files to staging plus a patch manifest `{path, before_sha256, after_sha256, action, unit_id}`. New files require absence; modifications require matching before hashes; escaping paths/symlinks and unexpected customer edits fail `ERR_TARGET_CONFLICT`. Generate `.env.example` with references/placeholders and no customer credentials.

Run syntax checks, PHPStan level 8 and Pint against the pinned fixture tool versions, then differential cases and the regression corpus. Failure preserves diagnostics and reverts only owned unchanged patches. Provenance records source/plan/adapter/target/dependency/corpus hashes. Targets run independently without Rewrotron, SQLite metadata or source-framework runtime.

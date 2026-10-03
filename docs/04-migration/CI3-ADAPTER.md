# CodeIgniter 3 source adapter

**Implementation:** Planned. **Initial implemented maturity:** EXPERIMENTAL. **Reference family:** CI3 3.1.x; first fixture pins 3.1.13. Path: `adapters/source/php/codeigniter3`.

## Detection and extraction

Bounded detection reads framework constants/files, Composer metadata, application/system root layout and config text without PHP execution. Conflicting/custom roots yield findings and a reviewed override. Runtime compatibility is distinct from the runtime actually executing the source; never infer the latter solely from a syntax floor.

Analyze via the PHP parser, then `Extract` resolves routes, HTTP verb arrays, conventions, `_remap`, nested controllers, custom `MY_Controller`/`MY_Model`, inheritance, autoload, helpers/libraries/models/views, hooks, environment config and DB/session/cache access into neutral records. Dynamic loaders and HMVC/third-party extensions are unresolved capabilities until fixtures support them.

A generic PHP fallback preserves inventory and uncertainty. It cannot silently apply CI3 conventions when detection is ambiguous. Framework/version confidence always includes supporting file/span evidence.

## Probe behavior

Prepare a disposable sandbox copy and overlay versioned probe code and hook configuration. Never patch `application/config/hooks.php` in the original source. Preserve existing hooks and their order; avoid replacing an application's output hook. Proxy capture records final bytes, including early exits and cache responses; probe completeness states describe gaps.

Source hooks alone do not capture every query, session driver mutation, file write, outbound socket or early failure. Required dimensions use driver wrappers/snapshots and controlled sinks. An unsupported driver causes UNKNOWN for the affected case. Run instrumented/uninstrumented fixture controls to detect observer-induced behavior changes.

## Required adapter fixtures

Config/convention/verb routing and precedence; default/404; `_remap`; loader/autoload; custom base classes; hook ordering; session drivers/flashdata; cache/early exit; transactional DB writes; dynamic models; cron/CLI roots; views/assets. Compare expected neutral IR, source hits, and completeness flags. Promote maturity only on the evidence requirements in PRODUCT.

Primary routing/session references are the [CI3 routing](https://codeigniter.com/userguide3/general/routing.html?highlight=uri) and [session guide](https://codeigniter.com/userguide3/libraries/sessions.html?highlight=sessions). Runtime support for a specific customer remains a doctor/fixture result.

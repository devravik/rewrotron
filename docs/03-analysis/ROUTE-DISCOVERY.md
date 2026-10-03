# Route and entrypoint discovery

The CI3 adapter resolves framework routing into neutral Application IR. Sources are static route configuration, controller conventions, supplied tests/traffic/logs, and sandbox observations. Every discovered candidate has provenance, dispatch priority, HTTP method (or `ANY` when undetermined), handler, classification and uncertainty.

CI3 supports HTTP verb-specific route arrays, regular-expression routes and URI conventions. Preserve route declaration precedence, `(:num)`/`(:any)` semantics, capture substitution, default controller, 404 override, `translate_uri_dashes`, nested controller directories and case-sensitive filesystem behavior. Public controller methods are candidates; inherited/private/reserved methods and custom `_remap` need adapter semantics rather than blanket exposure.

Method-body checks refine behavior contracts; they do not replace explicit dispatch rules. A route accepting GET but returning an error for it differs from a route that cannot dispatch GET. Observe invalid methods, trailing slashes, encoded paths, missing segments and unknown routes in fixtures.

Do not evaluate dynamic route PHP during static extraction. Record unresolved route expressions and `_remap` behavior; sandbox evidence may resolve concrete cases without proving all dynamic routes. Merge candidates by dispatch identity and evidence, not just path: overlapping patterns and method priorities must coexist.

The complete entrypoint inventory also includes CLI, cron and consumers. Supplied operational commands become roots with explicit provenance. Unsupported execution is visible in coverage; it cannot be omitted because the first replay engine is HTTP-focused.

Source: [CI3 routing documentation](https://codeigniter.com/userguide3/general/routing.html?highlight=uri), supports verb-specific route arrays; checked on 2026-10-03.

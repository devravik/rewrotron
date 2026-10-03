# Workflow discovery and executable scenarios

A workflow is a bounded guarded business scenario graph across operations, with state continuity and explicit value bindings. Login → dashboard and cart → checkout → confirmation are candidate examples; the customer supplies the actual critical scenarios.

Evidence can come from existing tests, declared API collections, supplied staging traces, redirects, rendered form/link targets, and related session/state mutations. Discovery heuristics propose sequences with evidence and confidence. They do not prove that adjacent traffic belongs to one user or that a visible link is safe to crawl.

A workflow record includes ID, operation IDs, bounded branches, initial fixture/principal, each request, extracted variables, later bindings, expected final state, required observation dimensions and unresolved questions. Use independent source/target cookie jars. Preserve state across steps; reset before the whole workflow. Never share raw session/CSRF tokens between implementations.

Cases cover authenticated and unauthenticated paths, permitted/forbidden roles, validation and business failures, logout, token rotation/expiry where applicable, and retry behavior. Cyclic applications need bounded test scenarios; automatic exploration stops at configured request/redirect/step limits and records coverage gaps.

Discovered workflows contribute edges to the application graph and migration unit dependencies. A heuristic result enters a confirmed corpus only after its assertions and safe execution are evidenced. Missing session mutation capture produces UNKNOWN, even if every response is HTTP 200.

Current executable representation is workflow.schema v2 under BEHAVIOR-MODEL. Source-derived scenarios carry evidence; unbounded/unsupported transitions remain explicit, never flattened.

Use BEHAVIOR-MODEL bounded scenarios and DIFFERENTIAL-TESTING logical occurrences/first-divergence policy. Discovery captures side-local outcomes before cases/targets exist; it never manufactures paired identity or forces a happy branch. Identical requests in different occurrences retain distinct identities.

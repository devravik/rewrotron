# Static parsing and reachability

PHP syntax handling belongs in `adapters/source/php/parser`; CI3 interpretation belongs in `adapters/source/php/codeigniter3`. Neutral graph algorithms live in `internal/analysis`. The chosen backend is [VKCOM/php-parser](https://github.com/VKCOM/php-parser), which documents PHP 5, PHP 7, and PHP 8.0–8.2 syntax. Pin the backend revision in RWT-007. Newer syntax is inventoried as unsupported until corpus evidence establishes support.

## Parsing contract

Never execute `routes.php`, Composer hooks, includes or application bootstrap for static detection. Parse source bytes from the confirmed manifest, retain spans and errors, and return neutral Source IR records. A failed/partial parse creates a finding and prevents DEAD proof for its affected region. Regex inventory can assist detection but cannot replace an AST while reporting analysis complete.

Collect classes, interfaces, traits, inheritance, functions, methods, closures, declarations, imports/includes, calls, potential entrypoints, configuration expressions and dependency metadata. ASTs are auxiliary artifacts; canonical symbols and edges reference baseline/file hashes and evidence.

## Reachability

Roots include HTTP dispatch, CLI commands, hooks, scheduled jobs, event consumers, templates/assets and supplied operational entrypoints. Construct typed call/include/reference/dispatch edges. Traverse from every approved root, accounting for symbol aliases and inheritance. Static reachability means possible execution, not actual observation.

An unresolved dynamic function/method/class/include, reflection, `eval`, plugin registry, or parse failure records the expression and candidate scope. If candidates cannot be bounded, the affected application region remains UNKNOWN. Observing one dynamic target does not prove no other target can execute. Runtime evidence adds resolved paths; it does not erase unresolved alternatives.

Bound parsing by file count/bytes/time and return a completeness report on limit/cancellation. A limit cannot silently truncate inventory. A companion reflection tool, if eventually needed, runs only in an isolated sandbox under Gate 1; it is an execution capability requiring its own evidence and cannot bootstrap arbitrary source on the host.

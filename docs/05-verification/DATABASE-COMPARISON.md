# Database state comparison

Compare application data in independent identically seeded databases using the same engine/version, schema, collation, timezone, SQL mode, isolation level and initial sequences. Retaining the schema does not permit source and target to share a writable database.

## Capture

Record the pre-state, committed post-state, transaction boundaries, tables read/written, and trigger/procedure outcomes. Query logs alone cannot prove mutations. Snapshot all in-scope tables initially; narrow coverage only after dependency/trigger analysis proves the affected set. A required table without capture is UNKNOWN.

Compare rows by declared primary/unique keys, stable column ordering, typed values, nulls, binary data, exact decimal representations, and date/time semantics. Row order is irrelevant for a table, but response query ordering remains observable. A table without a stable key uses a complete multiset comparison only when rows can be uniquely correlated; ambiguity yields UNKNOWN.

## Generated keys

IDs remain exact by default because they may appear in URLs, API payloads, files and integrations. If the plan explicitly approves different generated IDs, create a one-to-one mapping using an unambiguous creation event or stable business key. Enforce uniqueness, foreign-key integrity, cardinality and consistency across response URLs, session values, subsequent workflow requests, emails and other effects. Never mask IDs independently in different dimensions. If mapping is ambiguous, return UNKNOWN.

## Failure paths and concurrency

Check failed validation, insufficient permissions, constraint failures, partial transactions, and duplicate/replayed requests. A failed operation must leave the same committed state and effect counts as the source. Triggers and stored procedures are application behavior and must be included. Concurrency/locking behavior is a separate explicit scenario; sequential cases do not establish concurrent equivalence.

The first execution profile is MySQL with preserved customer schemas. PostgreSQL/other engines can be detected and inventoried but require a declared sandbox/reset/comparison capability before replay. Fixtures for signed/unsigned types, decimal precision, null/missing values, timezone, generated keys, triggers, rollback, multi-table writes and ambiguous keys are required in RWT-012.

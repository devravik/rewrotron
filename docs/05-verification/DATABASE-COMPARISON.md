# Application database identity, comparison and reset

**Owner:** Behavior of retained databases and full fixture reset. Preserving schema does not establish equivalent connection semantics. Source/target writable verification DBs are independent.

Capture connection_id, catalog/schema/table, engine exact bytes/version, driver, SQL mode, collation, timezone, isolation, autocommit, locking and connection-local state. Qualified reads/mutations include transaction IDs/order, generated values/sequences, triggers/routines/views/generated columns, precision/null/implicit cast semantics, errors and committed state. Raw SQL text is diagnostic; equivalent queries need not be text-equal. Read sets/transaction/locking behavior remain required where contract-relevant; unsupported visibility yields UNKNOWN even if final rows match.

Multi-DB effects are never keyed only by table name. No implicit cross-database atomicity claim. SQL without ORDER BY keeps observed ordering unless approved semantics explicitly permit another relation. Corrupted/legacy data and implicit MySQL conversions require characterization, not automatic target cleanup.

## Reset profiles

Each profile declares all mutable DB connections, rows/schema/triggers/routines/sequences/temp tables/connection state, session/cache/queue namespaces/files/sinks, initial identities and control assertions. Reset reports prove exact initial state within that boundary and no leftover workers. Transaction rollback alone is insufficient for auto-increment burns, DDL/implicit commits/nontransactional tables/temp tables or side effects. Use fresh owned DB instances from verified fixtures when restoration cannot prove these dimensions. Unsupported reset blocks replay.

Each side opens independent logical connections mapped to equivalent fixtures. Preserve intra-workflow state; reset between fresh scenarios/trials. Redis session/cache/locks/queues/flags need separate owned namespaces and restored TTL/clock behavior, not a wildcard cache clear on shared infrastructure. Disallow production access or same writable fixture on both sides.

Tests must remove a write, write to a same-named table on another DB, alter transaction boundaries, consume sequences on rollback, leave temp tables, perform DDL, change collation/timezone and contaminate Redis. Each must produce MISMATCH, invalid reset or UNKNOWN—not PASS.

## Qualified data models

[Data-model schema](../../specs/ir/data-model.schema.json) version 2.0 is required for current generation/DB actions. Each model binds connection_id, catalog, nullable schema, table and server_behavior_ref alongside columns/keys/evidence. Catalog names the database/catalog within the logical connection; schema is null only for an engine with no separate schema namespace. server_behavior_ref retains the exact driver/server/settings acquisition manifest for this connection, including the behavior dimensions listed above. Absence or an opaque unqualified model is a capability gap, not a default connection.

Relationships identify from_columns and a fully qualified destination plus evidence. FOREIGN_KEY requires actual discovered constraint evidence on an engine that supports that relationship; cross-database application relationships are OBSERVED_LOGICAL or INFERRED_LOGICAL, never invented physical constraints. Validate equal column arity, resolved column/table IDs and referenced evidence before use. An inferred relation cannot certify identity equivalence without characterization. Legacy unversioned descriptive models remain inspectable but must be re-extracted/resolved before current lowering. Application IR envelope v1 can reference qualified v2 models; this does not upgrade legacy child records implicitly.

Columns retain native type spelling and explicit canonical value_type/type_mapping_ref. The mapping witness pins adapter/profile/control evidence for native conversion, precision, timezone and null semantics. UNKNOWN with null mapping preserves an unsupported type honestly; an in-scope action/assertion needing it blocks rather than treating it as a string. Nullability is separate from absence and native coercion. Known mappings require witness refs; free native type names cannot silently select host-language conversions. Characterization can retain unknown columns, while current typed lowering requires supported mappings for every consumed value.

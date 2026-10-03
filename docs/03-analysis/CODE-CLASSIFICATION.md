# Conservative code classification and omission proof

**Owner:** ACTIVE / DORMANT / DEAD / UNKNOWN. Classification is mutually exclusive, evaluated in this order: ACTIVE if current trusted runtime evidence establishes execution (remaining uncertainty stays attached to the record); UNKNOWN if not observed and relevant parsing/dispatch/environment/root uncertainty remains; DORMANT if not observed, resolution is sufficient and a known reachable path exists; DEAD only if neither observed nor reachable and the entire omission bundle below is satisfied; otherwise UNKNOWN. Static reachability alone is never ACTIVE. Counts use exactly one state per candidate and identify the graph/profile/corpus revision. DEAD means no reachable evidence under approved analysis boundary B, never “proven never used.”

Automatic omission requires one immutable proof bundle included in the approved Gate 2 plan:

1. Closed operational root inventory: customer-approved HTTP/admin/webhook/CLI/cron/worker/plugin/flag/profile/config/data roots; supported enumeration mechanisms, complete declared categories, no unknown loading roots and current closure controls.
2. Complete relevant parsing and supported resolution with current graph/analyzer/adapter/source/profile identities.
3. No static reachable path and no unresolved candidate path into the omitted component, including reflection, serialized class names, database dispatch and dynamic loading.
4. Usable runtime corpus under the same boundary, with applicable capture controls, no relevant incomplete dimensions and zero hits.
5. Explicit omitted IDs, affected edges, evidence refs, boundary, validity predicates and approval subject.

A user assertion “no cron” alone cannot close a boundary: require configuration/deployment/service inventory checks and known excluded limitations. If operational evidence cannot establish closure, classification remains UNKNOWN; scope exclusion is a separate visible human decision, not automatic DEAD. Feature flags and production-only roots absent from the profile remain unknown. Never auto-omit DORMANT, UNKNOWN or unresolved dynamic candidates. Dependency “unused” and development-only are not DEAD synonyms.

Source/profile/analyzer/graph changes invalidate the proof bundle transitively. Historical proof remains retained but cannot omit under a successor plan without validated predicates. Falsification: forgotten scheduled job, plugin, admin path, serialized class, runtime-generated source or unreadable file must prevent omission.

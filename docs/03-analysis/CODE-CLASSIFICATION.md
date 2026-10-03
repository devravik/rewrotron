# Classification and omission evidence

| Classification | Rule | Target treatment |
| --- | --- | --- |
| `ACTIVE` | At least one correlated runtime hit in the approved baseline | Characterize and reconstruct |
| `DORMANT` | Proven possible static/configured path, no observed hit | Retain in scope or explicit reviewed exclusion |
| `UNKNOWN` | No hit and reachability cannot be closed because of dynamic edges, incomplete roots/parsing, or missing operations | Preserve or create a reviewed scope decision |
| `DEAD` | Closed static proof of no path from any approved root, no unresolved candidate path, and zero observed hits | Automatic target omission eligible with proof |

A known reachable symbol remains DORMANT even when other uncertainty exists; the uncertainty is recorded separately and prevents omission. Absence of observation is never sufficient for DEAD. Runtime evidence dominates: an observed symbol is ACTIVE even if the static graph missed it.

## Proof bundle

A DEAD proof records baseline/manifest hash, symbol and file/span, complete root inventory, graph hash, unresolved edge analysis, parser completeness, operational-input declarations, runtime corpus/window and zero hit count, analyzer version, and a machine-checkable reachability result. Inbound calls can exist among dead symbols; zero inbound edges is neither necessary nor sufficient. The required property is no reachable path from any approved root.

Cryptographic hashes protect bundle integrity and reproducibility, not the truth or completeness of the analysis. There is no “unforgeable” or universal cryptographic proof of dead PHP. No runtime corpus available means insufficient evidence for automatic omission in the initial engine.

A file is DEAD only if all executable symbols and top-level statements are proven dead and it provides no reachable templates, assets, configuration, include-time side effects or operational resources. Individual symbols may be DEAD in a mixed ACTIVE file. Omission applies to target reconstruction; the source repository is never edited or deleted.

Omission candidates and proofs appear in the Gate 2 plan. A stale baseline, new route/job, parse failure, unresolved dynamic edge or runtime hit invalidates the proof and dependent plan. RWT-008 includes fixtures for mutually unreachable cycles, dynamic loaders, cron-only routines, templates, includes, missing roots, and incomplete parsing.

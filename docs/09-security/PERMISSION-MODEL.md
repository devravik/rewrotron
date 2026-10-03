# Domain operation permissions

| Operation | Agent allowed | Preconditions / changes |
| --- | --- | --- |
| Execution state/timeline/drill-down | Yes | Sanitized same-project read; no shared seen marker, work claim or control mutation |
| Human pause/continue | Local human channel only | Persistent scheduler control; no gate approval; MCP inspection is read-only |
| Snapshot, case/plan lookup, mission peek | Yes | Registered project; sanitized read; audit may append |
| Detect | Yes | Bounded metadata read; persists proposal, never confirms |
| Analyze, sandbox observe/discover | Yes | Valid Gate 1; isolated runtime for observation |
| Plan proposal | Yes | Valid Gate 1; disclosed capabilities and scope |
| Source/plan/cutover confirmation | Human-approved channel only | Exact subject and persisted approval; MCP consumes a scoped token |
| Mission claim/renew/complete | Yes | Atomic leases; correct fence and completion evidence |
| Generate/repair | Yes | Valid Gates 1/2; target-only staged patch; current before hashes |
| Verify/benchmark | Yes | Confirmed scope, safe environments and current inputs |
| Omit DEAD in generated target | Yes | Current proof included in approved plan; source never deleted |
| Exclude DORMANT/UNKNOWN scope | Proposal only | Explicit reviewed plan change; limitations disclosed |
| Export/global lesson promotion | Human approval when sensitive | Sanitization and exact reviewed manifest |
| Production writes/deploy/traffic changes | Not implemented | Gate 3 prepares a package/runbook only |

Read-only means business state is not changed; an invocation can still append an audit event. `next` peeks without acquiring a lease. `research.claim_mission` is a mutation. Arbitrary code deletion, `config.enable_production` and raw SQL/shell tools are not public MCP operations.

Typed decision propose/show (MCP decisions.propose/get) are Gate 1 operations only; actions include accepting/rejecting/selecting alternatives/requesting evidence/proposing exceptions. Gate 2 alone authorizes frozen selected decisions. These are mediated permission rules; hostile same-account filesystem/shell access is outside engine isolation, as defined by SECURITY.

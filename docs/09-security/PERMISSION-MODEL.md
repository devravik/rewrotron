# Domain operation permissions

| Operation | Agent allowed | Preconditions / changes |
| --- | --- | --- |
| Execution state/timeline/drill-down | Yes | Sanitized same-project read; no shared seen marker, work claim or control mutation |
| Human pause/continue | Local human channel only | Persistent scheduler control; no gate approval; MCP inspection is read-only |
| Snapshot, case/plan lookup, mission peek | Yes | Registered project; sanitized read; audit may append |
| Detect | Yes | Bounded metadata read; persists proposal, never confirms |
| Analyze, sandbox observe/discover | Yes | Valid Gate 1; isolated runtime for observation |
| Assessment report/read | Yes | Shared guarded evidence; no implementation or approval |
| Assessment work/continuation decision | Local human channel only | Exact report/selected scope; no Gate 2 authority |
| Plan proposal | Yes | Valid Gate 1 plus current explicit human CONTINUE_PLANNING selection; disclosed capabilities and selected scope |
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

## Shared assessment boundary

`assess`/`assessment.run` use Gate 1 and existing read-only snapshot/isolated observation permissions; source and production are never mutated. Report publication writes only local project metadata/artifacts/events. `assessment show`/`assessment.get` remain inspection with existing privacy/export controls. Human-only local `assessment decide` records work/continuation, not Gate 2 permission. No MCP assessment method issues a human selection/token. Execution still requires the exact Gate 2 PLAN subject including selection/basis closure, and implementation admission must reject report/selection digests as authorization. Same-account external filesystem/shell bypass remains outside mediated enforcement.

# Error catalog and recovery

Every error has `code`, `message`, `category`, `retryable`, `recommended_action`, optional structured `details` and `artifact_ref`. Use the same domain codes across CLI and MCP. Go errors wrap underlying causes with `%w`; callers classify through typed errors/`errors.Is`/`errors.As`.

| Code | CLI exit | Retryable | Required action |
| --- | --- | --- | --- |
| ERR_INVALID_ARGUMENT / ERR_INVALID_CONTRACT | 1 | No | Fix invalid field using schema and details |
| ERR_SCHEMA_VERSION | 1 | No | Use compatible binary/schema; never silently downgrade |
| ERR_PROJECT_NOT_FOUND | 1 | No | Initialize/register the selected workspace |
| ERR_CONFIRMATION_REQUIRED | 4 for Gate 1/2, 3 for Gate 3 | No | Review and approve the exact current subject |
| ERR_SOURCE_DRIFT | 4 | No | Re-detect, review changes and re-confirm baseline |
| ERR_ASSESSMENT_DECISION_REQUIRED | 4 | No | Review assessment and explicitly select CONTINUE_PLANNING through the local human channel |
| ERR_PLAN_STALE | 4 | No | Re-synthesize and confirm current plan |
| ERR_APPROVAL_INVALID | 3 | No | Issue a valid human approval for the project/gate/subject |
| ERR_IDEMPOTENCY_CONFLICT | 1 | No | Use new request key or original identical payload |
| ERR_UNSUPPORTED_CAPABILITY | 1 | No | Supply capability/evidence or explicitly revise scope |
| ERR_UNSAFE_ENVIRONMENT | 1 | No | Establish the required isolation/sinks |
| ERR_FIXTURE_RESTORE | 1 | Conditional | Correct fixture/reset/environment; preserve UNKNOWN report |
| ERR_TARGET_CONFLICT | 1 | No | Review changed files and produce a fresh patch |
| ERR_LEASE_LOST | 1 | No | Stop publication; reconcile and acquire fresh lease |
| ERR_STORAGE_BUSY | 1 | Yes, bounded | Retry transaction without repeating external execution |
| ERR_ARTIFACT_CORRUPT | 1 | No | Restore verified bytes; do not accept affected evidence |
| ERR_VERIFICATION_MISMATCH | 2 | No | Repair proved difference using report |
| ERR_VERIFICATION_UNKNOWN | 2 | No | Obtain missing evidence or reviewed decision |
| ERR_PERFORMANCE_REGRESSION / ERR_PERFORMANCE_UNKNOWN | 2 | No | Repair or gather sufficient measurements |
| ERR_CANCELLED / ERR_TIMEOUT | 1 | Conditional | Resume from safe checkpoint/replay a fresh fixture |
| ERR_INTERNAL | 1 | No automatic retry | Preserve sanitized diagnostics and fix engine fault |

Categories match `specs/common.schema.json`. Machine contracts never require parsing human messages. Approval errors must identify the relevant gate/hash without exposing tokens. Fixture/environment failures preserve partial UNKNOWN evidence but exit 1 because execution did not complete safely. Retryability means a recorded bounded strategy exists, not an agent can retry indefinitely.

## Execution error presentation

The canonical execution view combines the existing code with stage/task/scope, concrete difference, evidence refs, recorded attempts, labeled hypothesis and next action. A task failure is separate from MISMATCH/UNKNOWN; paused, disconnected and stale-heartbeat states are not verification verdicts. `ERR_INVALID_ARGUMENT` covers invalid/future event cursors, watch/JSON or watch/quiet/dry-run combinations, and unknown verbosity. Reading developer details retains the same redaction/ownership rules. UI Skip proposes a scope/plan revision, never a silent PASS or exclusion.

Version 2 reasons: ERR_UNSUPPORTED_CAPABILITY (mechanism/operator/supervisor missing), ERR_UNBOUND_VALUE (missing typed input), ERR_CONFIG_CONFLICT (competing intent), BACKUP_INCOMPLETE (unavailable closure), ERR_QUOTA_EXCEEDED (admission paused), ERR_SUBJECT_VERSION (retained subject cannot execute). Include stage/scope/ref/evidence/next action; unsupported in-scope behavior yields UNKNOWN, never PASS.

A persisted useful assessment with BLOCKED/UNKNOWN migration feasibility is successful report completion, not ERR_UNSUPPORTED_CAPABILITY solely because migration is infeasible. Acquisition errors are reported as classified findings where possible; inability to persist the report remains an operation error. A stale selection/changed chosen scope must be reviewed before planning, never silently substituted.

ERR_UNSUPPORTED_CAPABILITY is the single public capability-refusal code across CLI/MCP/domain owners. Unsupported operator, profile, supervisor or pinned implementation details are typed diagnostic reasons; they do not create alternate spellings or a successful verification verdict. Historical spelling in the retained review is quoted input, not an interface alias.

## Review-2 refusal reasons

| Code | Exit | Automatic retry | Meaning / required action |
| --- | --- | --- | --- |
| ERR_STALE_EXECUTION | 1 | No | Loaded build/process/endpoint witness mismatches; terminate/reconcile and qualify fresh execution |
| ERR_INVALID_COMPARISON | 1 | No | Same-side or incompatible scenario/occurrence evidence; supply actual distinct SOURCE/TARGET captures |
| ERR_TYPE | 1 | No | Schema-qualified evaluation value type mismatch; affected assertion UNKNOWN, never implicit coercion |
| ERR_PERSISTENCE_UNAVAILABLE | 1 | No | Canonical store unavailable; report last committed sequence and clearly labeled emergency diagnostics, reconcile before retry |

PERSISTENCE_UNAVAILABLE is an emergency diagnostic state, not a committed execution enum or verifier verdict. Known semantic violations still produce MISMATCH under valid capture; setup/admission refusal is not a mutation kill. Unsupported profile/provider/encoding/horizon continues to use ERR_UNSUPPORTED_CAPABILITY with a specific reason.

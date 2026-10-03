# Error catalog and recovery

Every error has `code`, `message`, `category`, `retryable`, `recommended_action`, optional structured `details` and `artifact_ref`. Use the same domain codes across CLI and MCP. Go errors wrap underlying causes with `%w`; callers classify through typed errors/`errors.Is`/`errors.As`.

| Code | CLI exit | Retryable | Required action |
| --- | --- | --- | --- |
| ERR_INVALID_ARGUMENT / ERR_INVALID_CONTRACT | 1 | No | Fix invalid field using schema and details |
| ERR_SCHEMA_VERSION | 1 | No | Use compatible binary/schema; never silently downgrade |
| ERR_PROJECT_NOT_FOUND | 1 | No | Initialize/register the selected workspace |
| ERR_CONFIRMATION_REQUIRED | 4 for Gate 1/2, 3 for Gate 3 | No | Review and approve the exact current subject |
| ERR_SOURCE_DRIFT | 4 | No | Re-detect, review changes and re-confirm baseline |
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

Version 2 reasons: ERR_CAPABILITY_UNSUPPORTED (mechanism/operator/supervisor missing), ERR_UNBOUND_VALUE (missing typed input), ERR_CONFIG_CONFLICT (competing intent), BACKUP_INCOMPLETE (unavailable closure), ERR_QUOTA_EXCEEDED (admission paused), ERR_SUBJECT_VERSION (retained subject cannot execute). Include stage/scope/ref/evidence/next action; unsupported in-scope behavior yields UNKNOWN, never PASS.

package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

var allowedExecutionControlModes = map[string]struct{}{
	"RUNNING":         {},
	"PAUSE_REQUESTED": {},
	"PAUSED":          {},
	"IDLE":            {},
	"COMPLETED":       {},
	"FAILED":          {},
}

var allowedExecutionStages = map[string]struct{}{
	"DETECT":   {},
	"CONFIRM":  {},
	"ANALYZE":  {},
	"OBSERVE":  {},
	"MODEL":    {},
	"PLAN":     {},
	"MIGRATE":  {},
	"VERIFY":   {},
	"REPAIR":   {},
	"OPTIMIZE": {},
	"FINALIZE": {},
}

var allowedExecutionRunStatuses = map[string]struct{}{
	"PENDING":        {},
	"RUNNING":        {},
	"WAITING_USER":   {},
	"BLOCKED":        {},
	"PAUSED":         {},
	"COMPLETED":      {},
	"FAILED":         {},
	"CANCELLED":      {},
	"STALE":          {},
	"NOT_APPLICABLE": {},
}

var allowedExecutionLiveness = map[string]struct{}{
	"FRESH":       {},
	"STALE":       {},
	"NOT_RUNNING": {},
}

var allowedExecutionGateTypes = map[string]struct{}{
	"GATE_1_SOURCE_DETECTION": {},
	"GATE_2_PLAN_CONFIRMED":   {},
	"GATE_3_CUTOVER":          {},
}

// ExecutionStateError describes an invalid execution state record without echoing private contents.
type ExecutionStateError struct {
	Reason string
	Cause  error
}

// Error identifies the failed execution state obligation.
func (e *ExecutionStateError) Error() string {
	return "invalid execution state: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ExecutionStateError) Unwrap() error { return e.Cause }

// ValidateExecutionState validates a canonical live execution projection conforming to
// specs/execution/state.schema.json.
// It verifies schema version 1.0, required top-level projection fields, lifecycle and control
// mode enums, monotonically ordered sequences and RFC 3339 timestamps, valid active runs with
// stage tasks and subjects, stage instances, progress metrics, structured blockers, pending approvals,
// and closure references.
func ValidateExecutionState(raw []byte) error {
	record, err := parseExecutionStateJSON(raw)
	if err != nil {
		return err
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"lifecycle_state",
		"control_mode",
		"control_reason",
		"latest_event_sequence",
		"observed_at",
		"focus_run_id",
		"active_runs",
		"stage_instances",
		"last_activity_at",
		"last_heartbeat_at",
		"liveness",
		"progress_metrics",
		"blockers",
		"pending_approvals",
		"next_action",
		"user_request",
		"subjects",
		"recent_change_refs",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ExecutionStateError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ExecutionStateError{Reason: "object fields"}
	}

	var version string
	if string(record["schema_version"]) == "null" || json.Unmarshal(record["schema_version"], &version) != nil || version != "1.0" {
		return &ExecutionStateError{Reason: "unsupported schema version"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &ExecutionStateError{Reason: "project identity"}
	}

	var lifecycle string
	if string(record["lifecycle_state"]) == "null" || json.Unmarshal(record["lifecycle_state"], &lifecycle) != nil {
		return &ExecutionStateError{Reason: "invalid lifecycle_state"}
	}
	if _, ok := allowedOperationLifecycleStates[lifecycle]; !ok {
		return &ExecutionStateError{Reason: "invalid lifecycle_state"}
	}

	var controlMode string
	if string(record["control_mode"]) == "null" || json.Unmarshal(record["control_mode"], &controlMode) != nil {
		return &ExecutionStateError{Reason: "invalid control_mode"}
	}
	if _, ok := allowedExecutionControlModes[controlMode]; !ok {
		return &ExecutionStateError{Reason: "invalid control_mode"}
	}

	var controlReason string
	if string(record["control_reason"]) == "null" || json.Unmarshal(record["control_reason"], &controlReason) != nil || strings.TrimSpace(controlReason) == "" {
		return &ExecutionStateError{Reason: "control_reason"}
	}

	var latestSeq int64
	if string(record["latest_event_sequence"]) == "null" || json.Unmarshal(record["latest_event_sequence"], &latestSeq) != nil || latestSeq < 0 {
		return &ExecutionStateError{Reason: "latest_event_sequence"}
	}

	var observedAt string
	if string(record["observed_at"]) == "null" || json.Unmarshal(record["observed_at"], &observedAt) != nil {
		return &ExecutionStateError{Reason: "observed_at"}
	}
	if _, err := time.Parse(time.RFC3339, observedAt); err != nil {
		return &ExecutionStateError{Reason: "observed_at timestamp", Cause: err}
	}

	if string(record["focus_run_id"]) != "null" {
		var focusID string
		if json.Unmarshal(record["focus_run_id"], &focusID) != nil || !validAssessmentReportID(focusID) {
			return &ExecutionStateError{Reason: "focus_run_id"}
		}
	}

	if err := validateExecutionActiveRuns(record["active_runs"]); err != nil {
		return err
	}

	if err := validateExecutionStageInstances(record["stage_instances"]); err != nil {
		return err
	}

	if err := validateExecutionOptionalTimestamp(record["last_activity_at"], "last_activity_at"); err != nil {
		return err
	}

	if err := validateExecutionOptionalTimestamp(record["last_heartbeat_at"], "last_heartbeat_at"); err != nil {
		return err
	}

	var liveness string
	if string(record["liveness"]) == "null" || json.Unmarshal(record["liveness"], &liveness) != nil {
		return &ExecutionStateError{Reason: "invalid liveness"}
	}
	if _, ok := allowedExecutionLiveness[liveness]; !ok {
		return &ExecutionStateError{Reason: "invalid liveness"}
	}

	if err := validateExecutionProgressMetrics(record["progress_metrics"]); err != nil {
		return err
	}

	if err := validateExecutionBlockers(record["blockers"]); err != nil {
		return err
	}

	if err := validateExecutionPendingApprovals(record["pending_approvals"]); err != nil {
		return err
	}

	var nextAction string
	if string(record["next_action"]) == "null" || json.Unmarshal(record["next_action"], &nextAction) != nil || strings.TrimSpace(nextAction) == "" {
		return &ExecutionStateError{Reason: "next_action"}
	}

	if string(record["user_request"]) != "null" {
		var userReq string
		if json.Unmarshal(record["user_request"], &userReq) != nil || strings.TrimSpace(userReq) == "" {
			return &ExecutionStateError{Reason: "user_request"}
		}
	}

	if err := validateExecutionSubjects(record["subjects"]); err != nil {
		return err
	}

	if err := validateExecutionRecentChangeRefs(record["recent_change_refs"]); err != nil {
		return err
	}

	return nil
}

func validateExecutionActiveRuns(raw json.RawMessage) error {
	var runs []json.RawMessage
	if err := json.Unmarshal(raw, &runs); err != nil {
		return &ExecutionStateError{Reason: "active_runs array", Cause: err}
	}

	expectedRunFields := []string{
		"run_id",
		"operation_id",
		"stage",
		"stage_instance_id",
		"status",
		"current_task",
		"last_activity_at",
		"last_heartbeat_at",
		"subjects",
	}

	for _, rawRun := range runs {
		var runObj map[string]json.RawMessage
		if err := json.Unmarshal(rawRun, &runObj); err != nil || runObj == nil {
			return &ExecutionStateError{Reason: "run object", Cause: err}
		}

		for _, req := range expectedRunFields {
			if _, present := runObj[req]; !present {
				return &ExecutionStateError{Reason: "missing required run field"}
			}
		}
		if len(runObj) != len(expectedRunFields) {
			return &ExecutionStateError{Reason: "run object fields"}
		}

		var runID string
		if string(runObj["run_id"]) == "null" || json.Unmarshal(runObj["run_id"], &runID) != nil || !validAssessmentReportID(runID) {
			return &ExecutionStateError{Reason: "run_id"}
		}

		var opID string
		if string(runObj["operation_id"]) == "null" || json.Unmarshal(runObj["operation_id"], &opID) != nil || !validAssessmentReportID(opID) {
			return &ExecutionStateError{Reason: "run operation_id"}
		}

		var stage string
		if string(runObj["stage"]) == "null" || json.Unmarshal(runObj["stage"], &stage) != nil {
			return &ExecutionStateError{Reason: "run stage"}
		}
		if _, ok := allowedExecutionStages[stage]; !ok {
			return &ExecutionStateError{Reason: "run stage"}
		}

		var stageInstID string
		if string(runObj["stage_instance_id"]) == "null" || json.Unmarshal(runObj["stage_instance_id"], &stageInstID) != nil || !validAssessmentReportID(stageInstID) {
			return &ExecutionStateError{Reason: "run stage_instance_id"}
		}

		var status string
		if string(runObj["status"]) == "null" || json.Unmarshal(runObj["status"], &status) != nil {
			return &ExecutionStateError{Reason: "run status"}
		}
		if _, ok := allowedExecutionRunStatuses[status]; !ok {
			return &ExecutionStateError{Reason: "run status"}
		}

		if err := validateExecutionCurrentTask(runObj["current_task"]); err != nil {
			return err
		}

		if err := validateExecutionOptionalTimestamp(runObj["last_activity_at"], "run last_activity_at"); err != nil {
			return err
		}

		if err := validateExecutionOptionalTimestamp(runObj["last_heartbeat_at"], "run last_heartbeat_at"); err != nil {
			return err
		}

		if err := validateExecutionSubjects(runObj["subjects"]); err != nil {
			return err
		}
	}
	return nil
}

func validateExecutionCurrentTask(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var taskObj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &taskObj); err != nil || taskObj == nil {
		return &ExecutionStateError{Reason: "current_task object", Cause: err}
	}

	expectedTaskFields := []string{
		"task_id",
		"mission_id",
		"title",
		"reason",
		"status",
		"result",
		"next_action",
		"evidence_refs",
	}

	for _, req := range expectedTaskFields {
		if _, present := taskObj[req]; !present {
			return &ExecutionStateError{Reason: "missing required task field"}
		}
	}
	if len(taskObj) != len(expectedTaskFields) {
		return &ExecutionStateError{Reason: "current_task object fields"}
	}

	var taskID string
	if string(taskObj["task_id"]) == "null" || json.Unmarshal(taskObj["task_id"], &taskID) != nil || !validAssessmentReportID(taskID) {
		return &ExecutionStateError{Reason: "task_id"}
	}

	if string(taskObj["mission_id"]) != "null" {
		var mID string
		if json.Unmarshal(taskObj["mission_id"], &mID) != nil || !validAssessmentReportID(mID) {
			return &ExecutionStateError{Reason: "task mission_id"}
		}
	}

	var title string
	if string(taskObj["title"]) == "null" || json.Unmarshal(taskObj["title"], &title) != nil || strings.TrimSpace(title) == "" {
		return &ExecutionStateError{Reason: "task title"}
	}

	var reason string
	if string(taskObj["reason"]) == "null" || json.Unmarshal(taskObj["reason"], &reason) != nil || strings.TrimSpace(reason) == "" {
		return &ExecutionStateError{Reason: "task reason"}
	}

	var status string
	if string(taskObj["status"]) == "null" || json.Unmarshal(taskObj["status"], &status) != nil {
		return &ExecutionStateError{Reason: "task status"}
	}
	if _, ok := allowedExecutionRunStatuses[status]; !ok {
		return &ExecutionStateError{Reason: "task status"}
	}

	if string(taskObj["result"]) != "null" {
		var result string
		if json.Unmarshal(taskObj["result"], &result) != nil || strings.TrimSpace(result) == "" {
			return &ExecutionStateError{Reason: "task result"}
		}
	}

	if string(taskObj["next_action"]) != "null" {
		var nextAction string
		if json.Unmarshal(taskObj["next_action"], &nextAction) != nil || strings.TrimSpace(nextAction) == "" {
			return &ExecutionStateError{Reason: "task next_action"}
		}
	}

	var refs []json.RawMessage
	if err := json.Unmarshal(taskObj["evidence_refs"], &refs); err != nil {
		return &ExecutionStateError{Reason: "task evidence_refs array", Cause: err}
	}
	for _, ref := range refs {
		if !validProfileRef(ref, false) {
			return &ExecutionStateError{Reason: "task evidence_refs reference"}
		}
	}

	return nil
}

func validateExecutionStageInstances(raw json.RawMessage) error {
	var stages []json.RawMessage
	if err := json.Unmarshal(raw, &stages); err != nil {
		return &ExecutionStateError{Reason: "stage_instances array", Cause: err}
	}

	expectedStageFields := []string{
		"stage_instance_id",
		"run_id",
		"stage",
		"status",
		"summary",
		"evidence_refs",
		"artifact_refs",
	}

	for _, rawStage := range stages {
		var stageObj map[string]json.RawMessage
		if err := json.Unmarshal(rawStage, &stageObj); err != nil || stageObj == nil {
			return &ExecutionStateError{Reason: "stage instance object", Cause: err}
		}

		for _, req := range expectedStageFields {
			if _, present := stageObj[req]; !present {
				return &ExecutionStateError{Reason: "missing required stage instance field"}
			}
		}
		if len(stageObj) != len(expectedStageFields) {
			return &ExecutionStateError{Reason: "stage instance object fields"}
		}

		var stageInstID string
		if string(stageObj["stage_instance_id"]) == "null" || json.Unmarshal(stageObj["stage_instance_id"], &stageInstID) != nil || !validAssessmentReportID(stageInstID) {
			return &ExecutionStateError{Reason: "stage instance stage_instance_id"}
		}

		var runID string
		if string(stageObj["run_id"]) == "null" || json.Unmarshal(stageObj["run_id"], &runID) != nil || !validAssessmentReportID(runID) {
			return &ExecutionStateError{Reason: "stage instance run_id"}
		}

		var stage string
		if string(stageObj["stage"]) == "null" || json.Unmarshal(stageObj["stage"], &stage) != nil {
			return &ExecutionStateError{Reason: "stage instance stage"}
		}
		if _, ok := allowedExecutionStages[stage]; !ok {
			return &ExecutionStateError{Reason: "stage instance stage"}
		}

		var status string
		if string(stageObj["status"]) == "null" || json.Unmarshal(stageObj["status"], &status) != nil {
			return &ExecutionStateError{Reason: "stage instance status"}
		}
		if _, ok := allowedExecutionRunStatuses[status]; !ok {
			return &ExecutionStateError{Reason: "stage instance status"}
		}

		var summary string
		if string(stageObj["summary"]) == "null" || json.Unmarshal(stageObj["summary"], &summary) != nil || strings.TrimSpace(summary) == "" {
			return &ExecutionStateError{Reason: "stage instance summary"}
		}

		var evRefs []json.RawMessage
		if err := json.Unmarshal(stageObj["evidence_refs"], &evRefs); err != nil {
			return &ExecutionStateError{Reason: "stage instance evidence_refs array", Cause: err}
		}
		for _, ref := range evRefs {
			if !validProfileRef(ref, false) {
				return &ExecutionStateError{Reason: "stage instance evidence_refs reference"}
			}
		}

		var artRefs []json.RawMessage
		if err := json.Unmarshal(stageObj["artifact_refs"], &artRefs); err != nil {
			return &ExecutionStateError{Reason: "stage instance artifact_refs array", Cause: err}
		}
		for _, ref := range artRefs {
			if !validProfileRef(ref, false) {
				return &ExecutionStateError{Reason: "stage instance artifact_refs reference"}
			}
		}
	}
	return nil
}

func validateExecutionOptionalTimestamp(raw json.RawMessage, fieldName string) error {
	if string(raw) == "null" {
		return nil
	}
	var ts string
	if json.Unmarshal(raw, &ts) != nil {
		return &ExecutionStateError{Reason: fieldName}
	}
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		return &ExecutionStateError{Reason: fieldName + " timestamp", Cause: err}
	}
	return nil
}

func validateExecutionProgressMetrics(raw json.RawMessage) error {
	var metrics []json.RawMessage
	if err := json.Unmarshal(raw, &metrics); err != nil {
		return &ExecutionStateError{Reason: "progress_metrics array", Cause: err}
	}
	for _, rawMetric := range metrics {
		if err := ValidateExecutionProgress(rawMetric); err != nil {
			return &ExecutionStateError{Reason: "invalid progress metric item", Cause: err}
		}
	}
	return nil
}

func validateExecutionBlockers(raw json.RawMessage) error {
	var blockers []json.RawMessage
	if err := json.Unmarshal(raw, &blockers); err != nil {
		return &ExecutionStateError{Reason: "blockers array", Cause: err}
	}

	expectedBlockerFields := []string{
		"finding_id",
		"message",
		"reason",
		"affected_ids",
		"next_action",
		"requires_user",
		"options",
		"evidence_refs",
	}

	for _, rawBlocker := range blockers {
		var bObj map[string]json.RawMessage
		if err := json.Unmarshal(rawBlocker, &bObj); err != nil || bObj == nil {
			return &ExecutionStateError{Reason: "blocker object", Cause: err}
		}

		for _, req := range expectedBlockerFields {
			if _, present := bObj[req]; !present {
				return &ExecutionStateError{Reason: "missing required blocker field"}
			}
		}
		if len(bObj) != len(expectedBlockerFields) {
			return &ExecutionStateError{Reason: "blocker object fields"}
		}

		var fID string
		if string(bObj["finding_id"]) == "null" || json.Unmarshal(bObj["finding_id"], &fID) != nil || !validAssessmentReportID(fID) {
			return &ExecutionStateError{Reason: "blocker finding_id"}
		}

		var msg string
		if string(bObj["message"]) == "null" || json.Unmarshal(bObj["message"], &msg) != nil || strings.TrimSpace(msg) == "" {
			return &ExecutionStateError{Reason: "blocker message"}
		}

		var reason string
		if string(bObj["reason"]) == "null" || json.Unmarshal(bObj["reason"], &reason) != nil || strings.TrimSpace(reason) == "" {
			return &ExecutionStateError{Reason: "blocker reason"}
		}

		var affIDs []json.RawMessage
		if err := json.Unmarshal(bObj["affected_ids"], &affIDs); err != nil {
			return &ExecutionStateError{Reason: "blocker affected_ids array", Cause: err}
		}
		for _, affID := range affIDs {
			var id string
			if string(affID) == "null" || json.Unmarshal(affID, &id) != nil || !validAssessmentReportID(id) {
				return &ExecutionStateError{Reason: "blocker affected_ids identity"}
			}
		}

		var nextAction string
		if string(bObj["next_action"]) == "null" || json.Unmarshal(bObj["next_action"], &nextAction) != nil || strings.TrimSpace(nextAction) == "" {
			return &ExecutionStateError{Reason: "blocker next_action"}
		}

		var reqUser bool
		if string(bObj["requires_user"]) == "null" || json.Unmarshal(bObj["requires_user"], &reqUser) != nil {
			return &ExecutionStateError{Reason: "blocker requires_user"}
		}

		var opts []json.RawMessage
		if err := json.Unmarshal(bObj["options"], &opts); err != nil {
			return &ExecutionStateError{Reason: "blocker options array", Cause: err}
		}
		for _, opt := range opts {
			var optStr string
			if string(opt) == "null" || json.Unmarshal(opt, &optStr) != nil || strings.TrimSpace(optStr) == "" {
				return &ExecutionStateError{Reason: "blocker options entry"}
			}
		}

		var evRefs []json.RawMessage
		if err := json.Unmarshal(bObj["evidence_refs"], &evRefs); err != nil {
			return &ExecutionStateError{Reason: "blocker evidence_refs array", Cause: err}
		}
		for _, ref := range evRefs {
			if !validProfileRef(ref, false) {
				return &ExecutionStateError{Reason: "blocker evidence_refs reference"}
			}
		}
	}
	return nil
}

func validateExecutionPendingApprovals(raw json.RawMessage) error {
	var approvals []json.RawMessage
	if err := json.Unmarshal(raw, &approvals); err != nil {
		return &ExecutionStateError{Reason: "pending_approvals array", Cause: err}
	}

	expectedFields := []string{
		"gate_type",
		"subject_hash",
		"reason",
		"next_action",
		"evidence_refs",
	}

	for _, rawApp := range approvals {
		var appObj map[string]json.RawMessage
		if err := json.Unmarshal(rawApp, &appObj); err != nil || appObj == nil {
			return &ExecutionStateError{Reason: "pending approval object", Cause: err}
		}

		for _, req := range expectedFields {
			if _, present := appObj[req]; !present {
				return &ExecutionStateError{Reason: "missing required approval field"}
			}
		}
		if len(appObj) != len(expectedFields) {
			return &ExecutionStateError{Reason: "pending approval object fields"}
		}

		var gate string
		if string(appObj["gate_type"]) == "null" || json.Unmarshal(appObj["gate_type"], &gate) != nil {
			return &ExecutionStateError{Reason: "approval gate_type"}
		}
		if _, ok := allowedExecutionGateTypes[gate]; !ok {
			return &ExecutionStateError{Reason: "approval gate_type"}
		}

		var hash string
		if string(appObj["subject_hash"]) == "null" || json.Unmarshal(appObj["subject_hash"], &hash) != nil || !isManifestDigest(hash) {
			return &ExecutionStateError{Reason: "approval subject_hash"}
		}

		var reason string
		if string(appObj["reason"]) == "null" || json.Unmarshal(appObj["reason"], &reason) != nil || strings.TrimSpace(reason) == "" {
			return &ExecutionStateError{Reason: "approval reason"}
		}

		var nextAction string
		if string(appObj["next_action"]) == "null" || json.Unmarshal(appObj["next_action"], &nextAction) != nil || strings.TrimSpace(nextAction) == "" {
			return &ExecutionStateError{Reason: "approval next_action"}
		}

		var evRefs []json.RawMessage
		if err := json.Unmarshal(appObj["evidence_refs"], &evRefs); err != nil {
			return &ExecutionStateError{Reason: "approval evidence_refs array", Cause: err}
		}
		for _, ref := range evRefs {
			if !validProfileRef(ref, false) {
				return &ExecutionStateError{Reason: "approval evidence_refs reference"}
			}
		}
	}
	return nil
}

func validateExecutionSubjects(raw json.RawMessage) error {
	var subj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &subj); err != nil || subj == nil {
		return &ExecutionStateError{Reason: "subjects object", Cause: err}
	}

	expectedSubjectFields := []string{
		"configuration_hash",
		"plan_hash",
		"target_spec_hash",
		"target_tree_hash",
		"corpus_hash",
		"policy_hash",
	}

	for _, req := range expectedSubjectFields {
		if _, present := subj[req]; !present {
			return &ExecutionStateError{Reason: "missing required subject field"}
		}
	}
	if len(subj) != len(expectedSubjectFields) {
		return &ExecutionStateError{Reason: "subjects object fields"}
	}

	for _, field := range expectedSubjectFields {
		rawHash := subj[field]
		if string(rawHash) == "null" {
			continue
		}
		var hash string
		if json.Unmarshal(rawHash, &hash) != nil || !isManifestDigest(hash) {
			return &ExecutionStateError{Reason: field}
		}
	}
	return nil
}

func validateExecutionRecentChangeRefs(raw json.RawMessage) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil {
		return &ExecutionStateError{Reason: "recent_change_refs array", Cause: err}
	}
	for _, ref := range refs {
		if !validProfileRef(ref, false) {
			return &ExecutionStateError{Reason: "recent_change_refs reference"}
		}
	}
	return nil
}

func parseExecutionStateJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ExecutionStateError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ExecutionStateError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ExecutionStateError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ExecutionStateError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ExecutionStateError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ExecutionStateError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ExecutionStateError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ExecutionStateError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ExecutionStateError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ExecutionStateError{Reason: "trailing input"}
	}
	return record, nil
}

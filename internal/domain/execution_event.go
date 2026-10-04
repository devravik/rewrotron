package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

var allowedExecutionEventTypes = map[string]struct{}{
	"STAGE_STARTED":        {},
	"STAGE_COMPLETED":      {},
	"TASK_STARTED":         {},
	"TASK_COMPLETED":       {},
	"TASK_BLOCKED":         {},
	"DISCOVERY_FOUND":      {},
	"DECISION_MADE":        {},
	"ACTION_STARTED":       {},
	"ACTION_COMPLETED":     {},
	"WARNING":              {},
	"ERROR":                {},
	"APPROVAL_REQUIRED":    {},
	"APPROVAL_RECEIVED":    {},
	"MILESTONE_REACHED":    {},
	"PROGRESS_UPDATED":     {},
	"HEARTBEAT":            {},
	"PAUSE_REQUESTED":      {},
	"EXECUTION_PAUSED":     {},
	"EXECUTION_CONTINUED":  {},
	"EXECUTION_CANCELLED":  {},
	"EXECUTION_RECOVERED":  {},
	"ARTIFACT_INVALIDATED": {},
}

var allowedExecutionEventKinds = map[string]struct{}{
	"OBSERVED": {},
	"DECIDED":  {},
	"ACTION":   {},
	"RESULT":   {},
}

var allowedExecutionEventSeverities = map[string]struct{}{
	"INFO":     {},
	"WARNING":  {},
	"ERROR":    {},
	"CRITICAL": {},
}

var allowedExecutionEventProducers = map[string]struct{}{
	"ENGINE":  {},
	"ADAPTER": {},
	"AGENT":   {},
}

// ExecutionEventError describes an invalid execution event record without echoing private contents.
type ExecutionEventError struct {
	Reason string
	Cause  error
}

// Error identifies the failed execution event obligation.
func (e *ExecutionEventError) Error() string {
	return "invalid execution event: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ExecutionEventError) Unwrap() error { return e.Cause }

// ExecutionCausalityError describes an invalid execution causality record without echoing private contents.
type ExecutionCausalityError struct {
	Reason string
	Cause  error
}

// Error identifies the failed execution causality obligation.
func (e *ExecutionCausalityError) Error() string {
	return "invalid execution causality: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ExecutionCausalityError) Unwrap() error { return e.Cause }

// ValidateExecutionEvent validates a committed structured execution event record conforming
// to specs/execution/event.schema.json (schema_version 1.0 or 2.0).
// It verifies required schema fields, valid sequence numbers (integer >= 1), allowed enums
// for stages, event types, event kinds, severity, and producers.
// It verifies cross-field constraints:
// - requires_user == true requires a non-empty user_request
// - APPROVAL_REQUIRED and APPROVAL_RECEIVED require an approval object
// - APPROVAL_RECEIVED requires a non-null confirmation_id and producer ENGINE
// - PROGRESS_UPDATED requires a non-null progress record conforming to ValidateExecutionProgress
// - ACTION_COMPLETED, TASK_COMPLETED, and STAGE_COMPLETED require a non-empty result
// - Schema 2.0 requires a valid causality record conforming to specs/observation/causality.schema.json.
func ValidateExecutionEvent(raw []byte) error {
	record, err := parseExecutionEventJSON(raw)
	if err != nil {
		return err
	}

	var version string
	if string(record["schema_version"]) == "null" || json.Unmarshal(record["schema_version"], &version) != nil {
		return &ExecutionEventError{Reason: "schema_version"}
	}
	if version != "1.0" && version != "2.0" {
		return &ExecutionEventError{Reason: "unsupported schema_version"}
	}

	expectedFields := []string{
		"schema_version",
		"event_id",
		"sequence",
		"project_id",
		"run_id",
		"operation_id",
		"mission_id",
		"task_id",
		"stage",
		"stage_instance_id",
		"type",
		"kind",
		"severity",
		"producer",
		"requires_user",
		"created_at",
		"causation_event_id",
		"title",
		"summary",
		"reason",
		"result",
		"next_action",
		"user_request",
		"entity_type",
		"entity_id",
		"subjects",
		"evidence_refs",
		"artifact_refs",
		"progress",
		"approval",
		"details",
	}
	if version == "2.0" {
		expectedFields = append(expectedFields, "causality")
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ExecutionEventError{Reason: "missing required field: " + field}
		}
	}
	if len(record) != len(expectedFields) {
		return &ExecutionEventError{Reason: "unexpected additional properties"}
	}

	var eventID string
	if string(record["event_id"]) == "null" || json.Unmarshal(record["event_id"], &eventID) != nil || !validAssessmentReportID(eventID) {
		return &ExecutionEventError{Reason: "invalid event_id"}
	}

	var sequence int64
	if string(record["sequence"]) == "null" || json.Unmarshal(record["sequence"], &sequence) != nil || sequence < 1 {
		return &ExecutionEventError{Reason: "sequence must be integer >= 1"}
	}
	if strings.ContainsAny(string(record["sequence"]), ".eE") {
		return &ExecutionEventError{Reason: "sequence must be integer"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &ExecutionEventError{Reason: "invalid project_id"}
	}

	var runID string
	if string(record["run_id"]) == "null" || json.Unmarshal(record["run_id"], &runID) != nil || !validAssessmentReportID(runID) {
		return &ExecutionEventError{Reason: "invalid run_id"}
	}

	var opID string
	if string(record["operation_id"]) == "null" || json.Unmarshal(record["operation_id"], &opID) != nil || !validAssessmentReportID(opID) {
		return &ExecutionEventError{Reason: "invalid operation_id"}
	}

	if string(record["mission_id"]) != "null" {
		var missionID string
		if json.Unmarshal(record["mission_id"], &missionID) != nil || !validAssessmentReportID(missionID) {
			return &ExecutionEventError{Reason: "invalid mission_id"}
		}
	}

	if string(record["task_id"]) != "null" {
		var taskID string
		if json.Unmarshal(record["task_id"], &taskID) != nil || !validAssessmentReportID(taskID) {
			return &ExecutionEventError{Reason: "invalid task_id"}
		}
	}

	var stage string
	if string(record["stage"]) == "null" || json.Unmarshal(record["stage"], &stage) != nil {
		return &ExecutionEventError{Reason: "invalid stage"}
	}
	if _, ok := allowedExecutionStages[stage]; !ok {
		return &ExecutionEventError{Reason: "invalid stage enum"}
	}

	var stageInstID string
	if string(record["stage_instance_id"]) == "null" || json.Unmarshal(record["stage_instance_id"], &stageInstID) != nil || !validAssessmentReportID(stageInstID) {
		return &ExecutionEventError{Reason: "invalid stage_instance_id"}
	}

	var evType string
	if string(record["type"]) == "null" || json.Unmarshal(record["type"], &evType) != nil {
		return &ExecutionEventError{Reason: "invalid type"}
	}
	if _, ok := allowedExecutionEventTypes[evType]; !ok {
		return &ExecutionEventError{Reason: "invalid type enum"}
	}

	var kind string
	if string(record["kind"]) == "null" || json.Unmarshal(record["kind"], &kind) != nil {
		return &ExecutionEventError{Reason: "invalid kind"}
	}
	if _, ok := allowedExecutionEventKinds[kind]; !ok {
		return &ExecutionEventError{Reason: "invalid kind enum"}
	}

	var severity string
	if string(record["severity"]) == "null" || json.Unmarshal(record["severity"], &severity) != nil {
		return &ExecutionEventError{Reason: "invalid severity"}
	}
	if _, ok := allowedExecutionEventSeverities[severity]; !ok {
		return &ExecutionEventError{Reason: "invalid severity enum"}
	}

	var producer string
	if string(record["producer"]) == "null" || json.Unmarshal(record["producer"], &producer) != nil {
		return &ExecutionEventError{Reason: "invalid producer"}
	}
	if _, ok := allowedExecutionEventProducers[producer]; !ok {
		return &ExecutionEventError{Reason: "invalid producer enum"}
	}

	var reqUser bool
	if string(record["requires_user"]) == "null" || json.Unmarshal(record["requires_user"], &reqUser) != nil {
		return &ExecutionEventError{Reason: "requires_user must be a boolean"}
	}

	var createdAt string
	if string(record["created_at"]) == "null" || json.Unmarshal(record["created_at"], &createdAt) != nil {
		return &ExecutionEventError{Reason: "invalid created_at"}
	}
	if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
		if _, errNano := time.Parse(time.RFC3339Nano, createdAt); errNano != nil {
			return &ExecutionEventError{Reason: "created_at must be RFC3339 date-time"}
		}
	}

	if string(record["causation_event_id"]) != "null" {
		var causeID string
		if json.Unmarshal(record["causation_event_id"], &causeID) != nil || !validAssessmentReportID(causeID) {
			return &ExecutionEventError{Reason: "invalid causation_event_id"}
		}
	}

	var title string
	if string(record["title"]) == "null" || json.Unmarshal(record["title"], &title) != nil || len(title) < 1 {
		return &ExecutionEventError{Reason: "invalid title"}
	}

	var summary string
	if string(record["summary"]) == "null" || json.Unmarshal(record["summary"], &summary) != nil || len(summary) < 1 {
		return &ExecutionEventError{Reason: "invalid summary"}
	}

	var reason string
	if string(record["reason"]) == "null" || json.Unmarshal(record["reason"], &reason) != nil || len(reason) < 1 {
		return &ExecutionEventError{Reason: "invalid reason"}
	}

	var hasResult bool
	if string(record["result"]) != "null" {
		var res string
		if json.Unmarshal(record["result"], &res) != nil || len(res) < 1 {
			return &ExecutionEventError{Reason: "invalid result"}
		}
		hasResult = true
	}

	if string(record["next_action"]) != "null" {
		var nextAct string
		if json.Unmarshal(record["next_action"], &nextAct) != nil || len(nextAct) < 1 {
			return &ExecutionEventError{Reason: "invalid next_action"}
		}
	}

	var hasUserRequest bool
	if string(record["user_request"]) != "null" {
		var userReq string
		if json.Unmarshal(record["user_request"], &userReq) != nil || len(userReq) < 1 {
			return &ExecutionEventError{Reason: "invalid user_request"}
		}
		hasUserRequest = true
	}

	var entityType string
	if string(record["entity_type"]) == "null" || json.Unmarshal(record["entity_type"], &entityType) != nil || len(entityType) < 1 {
		return &ExecutionEventError{Reason: "invalid entity_type"}
	}

	var entityID string
	if string(record["entity_id"]) == "null" || json.Unmarshal(record["entity_id"], &entityID) != nil || !validAssessmentReportID(entityID) {
		return &ExecutionEventError{Reason: "invalid entity_id"}
	}

	if err := validateExecutionEventSubjects(record["subjects"]); err != nil {
		return err
	}

	var evRefs []json.RawMessage
	if string(record["evidence_refs"]) == "null" || json.Unmarshal(record["evidence_refs"], &evRefs) != nil {
		return &ExecutionEventError{Reason: "evidence_refs must be an array"}
	}
	for _, ref := range evRefs {
		if !validProfileRef(ref, false) {
			return &ExecutionEventError{Reason: "invalid evidence_refs item"}
		}
	}

	var artRefs []json.RawMessage
	if string(record["artifact_refs"]) == "null" || json.Unmarshal(record["artifact_refs"], &artRefs) != nil {
		return &ExecutionEventError{Reason: "artifact_refs must be an array"}
	}
	for _, ref := range artRefs {
		if !validProfileRef(ref, false) {
			return &ExecutionEventError{Reason: "invalid artifact_refs item"}
		}
	}

	var hasProgress bool
	if string(record["progress"]) != "null" {
		if err := ValidateExecutionProgress(record["progress"]); err != nil {
			return &ExecutionEventError{Reason: "invalid progress", Cause: err}
		}
		hasProgress = true
	}

	var hasApproval bool
	var approvalConfirmationID *string
	if string(record["approval"]) != "null" {
		var appObj map[string]json.RawMessage
		if json.Unmarshal(record["approval"], &appObj) != nil || appObj == nil {
			return &ExecutionEventError{Reason: "approval must be an object"}
		}
		expectedApprovalFields := []string{"gate_type", "subject_hash", "confirmation_id"}
		for _, f := range expectedApprovalFields {
			if _, present := appObj[f]; !present {
				return &ExecutionEventError{Reason: "missing required approval field: " + f}
			}
		}
		if len(appObj) != len(expectedApprovalFields) {
			return &ExecutionEventError{Reason: "unexpected approval fields"}
		}
		var gate string
		if string(appObj["gate_type"]) == "null" || json.Unmarshal(appObj["gate_type"], &gate) != nil {
			return &ExecutionEventError{Reason: "invalid approval gate_type"}
		}
		if _, ok := allowedExecutionGateTypes[gate]; !ok {
			return &ExecutionEventError{Reason: "invalid approval gate_type enum"}
		}
		var subHash string
		if string(appObj["subject_hash"]) == "null" || json.Unmarshal(appObj["subject_hash"], &subHash) != nil || !isManifestDigest(subHash) {
			return &ExecutionEventError{Reason: "invalid approval subject_hash"}
		}
		if string(appObj["confirmation_id"]) != "null" {
			var confID string
			if json.Unmarshal(appObj["confirmation_id"], &confID) != nil || !validAssessmentReportID(confID) {
				return &ExecutionEventError{Reason: "invalid approval confirmation_id"}
			}
			approvalConfirmationID = &confID
		}
		hasApproval = true
	}

	var detailsObj map[string]json.RawMessage
	if string(record["details"]) == "null" || json.Unmarshal(record["details"], &detailsObj) != nil || detailsObj == nil {
		return &ExecutionEventError{Reason: "details must be an object"}
	}

	if version == "2.0" {
		if err := ValidateExecutionCausality(record["causality"]); err != nil {
			return &ExecutionEventError{Reason: "causality", Cause: err}
		}
	}

	// Schema allOf conditional constraints
	if reqUser && !hasUserRequest {
		return &ExecutionEventError{Reason: "user_request is required when requires_user is true"}
	}

	if (evType == "APPROVAL_REQUIRED" || evType == "APPROVAL_RECEIVED") && !hasApproval {
		return &ExecutionEventError{Reason: "approval is required for " + evType}
	}

	if evType == "APPROVAL_RECEIVED" {
		if approvalConfirmationID == nil {
			return &ExecutionEventError{Reason: "confirmation_id is required for APPROVAL_RECEIVED"}
		}
		if producer != "ENGINE" {
			return &ExecutionEventError{Reason: "producer must be ENGINE for APPROVAL_RECEIVED"}
		}
	}

	if evType == "PROGRESS_UPDATED" && !hasProgress {
		return &ExecutionEventError{Reason: "progress is required for PROGRESS_UPDATED"}
	}

	if (evType == "ACTION_COMPLETED" || evType == "TASK_COMPLETED" || evType == "STAGE_COMPLETED") && !hasResult {
		return &ExecutionEventError{Reason: "result is required for " + evType}
	}

	return nil
}

// ValidateExecutionCausality validates an execution causality record conforming to
// specs/observation/causality.schema.json.
// It verifies required fields, valid identifier patterns for run, workflow, process,
// trace, and span identities, and unique cause_ids.
func ValidateExecutionCausality(raw []byte) error {
	record, err := parseCausalityJSON(raw)
	if err != nil {
		return err
	}

	expectedFields := []string{
		"project_id",
		"run_id",
		"workflow_instance_id",
		"operation_id",
		"request_id",
		"process_id",
		"trace_id",
		"span_id",
		"parent_span_id",
		"cause_ids",
		"job_id",
		"effect_id",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ExecutionCausalityError{Reason: "missing required field: " + field}
		}
	}
	if len(record) != len(expectedFields) {
		return &ExecutionCausalityError{Reason: "unexpected additional properties"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &ExecutionCausalityError{Reason: "invalid project_id"}
	}

	var runID string
	if string(record["run_id"]) == "null" || json.Unmarshal(record["run_id"], &runID) != nil || !validAssessmentReportID(runID) {
		return &ExecutionCausalityError{Reason: "invalid run_id"}
	}

	if string(record["workflow_instance_id"]) != "null" {
		var wfID string
		if json.Unmarshal(record["workflow_instance_id"], &wfID) != nil || !validAssessmentReportID(wfID) {
			return &ExecutionCausalityError{Reason: "invalid workflow_instance_id"}
		}
	}

	var opID string
	if string(record["operation_id"]) == "null" || json.Unmarshal(record["operation_id"], &opID) != nil || !validAssessmentReportID(opID) {
		return &ExecutionCausalityError{Reason: "invalid operation_id"}
	}

	if string(record["request_id"]) != "null" {
		var reqID string
		if json.Unmarshal(record["request_id"], &reqID) != nil || !validAssessmentReportID(reqID) {
			return &ExecutionCausalityError{Reason: "invalid request_id"}
		}
	}

	var procID string
	if string(record["process_id"]) == "null" || json.Unmarshal(record["process_id"], &procID) != nil || !validAssessmentReportID(procID) {
		return &ExecutionCausalityError{Reason: "invalid process_id"}
	}

	var traceID string
	if string(record["trace_id"]) == "null" || json.Unmarshal(record["trace_id"], &traceID) != nil || !validAssessmentReportID(traceID) {
		return &ExecutionCausalityError{Reason: "invalid trace_id"}
	}

	var spanID string
	if string(record["span_id"]) == "null" || json.Unmarshal(record["span_id"], &spanID) != nil || !validAssessmentReportID(spanID) {
		return &ExecutionCausalityError{Reason: "invalid span_id"}
	}

	if string(record["parent_span_id"]) != "null" {
		var parentSpanID string
		if json.Unmarshal(record["parent_span_id"], &parentSpanID) != nil || !validAssessmentReportID(parentSpanID) {
			return &ExecutionCausalityError{Reason: "invalid parent_span_id"}
		}
	}

	var causeIDs []json.RawMessage
	if string(record["cause_ids"]) == "null" || json.Unmarshal(record["cause_ids"], &causeIDs) != nil {
		return &ExecutionCausalityError{Reason: "cause_ids must be an array"}
	}
	seenCauses := make(map[string]struct{}, len(causeIDs))
	for _, rawCause := range causeIDs {
		var cID string
		if string(rawCause) == "null" || json.Unmarshal(rawCause, &cID) != nil || !validAssessmentReportID(cID) {
			return &ExecutionCausalityError{Reason: "invalid cause_id item"}
		}
		if _, seen := seenCauses[cID]; seen {
			return &ExecutionCausalityError{Reason: "duplicate cause_id item"}
		}
		seenCauses[cID] = struct{}{}
	}

	if string(record["job_id"]) != "null" {
		var jobID string
		if json.Unmarshal(record["job_id"], &jobID) != nil || !validAssessmentReportID(jobID) {
			return &ExecutionCausalityError{Reason: "invalid job_id"}
		}
	}

	if string(record["effect_id"]) != "null" {
		var effectID string
		if json.Unmarshal(record["effect_id"], &effectID) != nil || !validAssessmentReportID(effectID) {
			return &ExecutionCausalityError{Reason: "invalid effect_id"}
		}
	}

	return nil
}

func validateExecutionEventSubjects(raw json.RawMessage) error {
	var subj map[string]json.RawMessage
	if json.Unmarshal(raw, &subj) != nil || subj == nil {
		return &ExecutionEventError{Reason: "subjects must be an object"}
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
			return &ExecutionEventError{Reason: "missing required subject field: " + req}
		}
	}
	if len(subj) != len(expectedSubjectFields) {
		return &ExecutionEventError{Reason: "unexpected subjects object fields"}
	}

	for _, field := range expectedSubjectFields {
		rawHash := subj[field]
		if string(rawHash) == "null" {
			continue
		}
		var hash string
		if json.Unmarshal(rawHash, &hash) != nil || !isManifestDigest(hash) {
			return &ExecutionEventError{Reason: "invalid hash for subject: " + field}
		}
	}
	return nil
}

func parseExecutionEventJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ExecutionEventError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ExecutionEventError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ExecutionEventError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ExecutionEventError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ExecutionEventError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ExecutionEventError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ExecutionEventError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ExecutionEventError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ExecutionEventError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ExecutionEventError{Reason: "trailing input"}
	}
	return record, nil
}

func parseCausalityJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ExecutionCausalityError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ExecutionCausalityError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ExecutionCausalityError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ExecutionCausalityError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ExecutionCausalityError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ExecutionCausalityError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ExecutionCausalityError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ExecutionCausalityError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ExecutionCausalityError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ExecutionCausalityError{Reason: "trailing input"}
	}
	return record, nil
}

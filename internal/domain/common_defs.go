package domain

import (
	"encoding/json"
	"strings"
)

var allowedArtifactVisibilities = map[string]struct{}{
	"PRIVATE_SOURCE":     {},
	"SANITIZED_EVIDENCE": {},
}

var allowedHTTPMethods = map[string]struct{}{
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"PATCH":   {},
	"DELETE":  {},
	"OPTIONS": {},
	"HEAD":    {},
	"ANY":     {},
}

var allowedEvidenceKinds = map[string]struct{}{
	"STATIC":         {},
	"RUNTIME":        {},
	"USER_OVERRIDE":  {},
	"AGENT_PROPOSAL": {},
}

var allowedErrorCategories = map[string]struct{}{
	"USER_ERROR":         {},
	"ENVIRONMENT_ERROR":  {},
	"SOURCE_ERROR":       {},
	"TARGET_BUILD_ERROR": {},
	"ANALYSIS_ERROR":     {},
	"OBSERVATION_ERROR":  {},
	"VERIFICATION_ERROR": {},
	"AGENT_ERROR":        {},
	"INTERNAL_ERROR":     {},
}

var allowedCompletenessStates = map[string]struct{}{
	"COMPLETE":       {},
	"PARTIAL":        {},
	"UNAVAILABLE":    {},
	"NOT_APPLICABLE": {},
}

var completenessFields = []string{
	"http",
	"body",
	"authentication",
	"authorization",
	"session",
	"database",
	"external",
	"files",
	"cache",
	"errors",
	"execution",
}

// CommonDefError describes an invalid common schema definition record without echoing private contents.
type CommonDefError struct {
	Def    string
	Reason string
	Cause  error
}

// Error identifies the failed common definition constraint.
func (e *CommonDefError) Error() string {
	if e.Def != "" {
		return "invalid " + e.Def + ": " + e.Reason
	}
	return "invalid common definition: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *CommonDefError) Unwrap() error { return e.Cause }

// ValidateArtifactPolicy validates an artifact policy record conforming to
// specs/common.schema.json#/$defs/artifact_policy.
// It verifies valid UTF-8 JSON, required visibility and policy_ref fields,
// allowed visibility enums (PRIVATE_SOURCE, SANITIZED_EVIDENCE), valid sha256
// reference format, and rejection of unexpected fields.
func ValidateArtifactPolicy(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "artifact_policy", Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &CommonDefError{Def: "artifact_policy", Reason: "object required", Cause: err}
	}

	if len(record) != 2 {
		return &CommonDefError{Def: "artifact_policy", Reason: "unexpected fields"}
	}

	visRaw, ok := record["visibility"]
	if !ok || string(visRaw) == "null" {
		return &CommonDefError{Def: "artifact_policy", Reason: "missing required field: visibility"}
	}
	var visibility string
	if err := json.Unmarshal(visRaw, &visibility); err != nil {
		return &CommonDefError{Def: "artifact_policy", Reason: "visibility must be a string", Cause: err}
	}
	if _, ok := allowedArtifactVisibilities[visibility]; !ok {
		return &CommonDefError{Def: "artifact_policy", Reason: "invalid visibility"}
	}

	policyRefRaw, ok := record["policy_ref"]
	if !ok || string(policyRefRaw) == "null" {
		return &CommonDefError{Def: "artifact_policy", Reason: "missing required field: policy_ref"}
	}
	if !validProfileRef(policyRefRaw, false) {
		return &CommonDefError{Def: "artifact_policy", Reason: "policy_ref must be a valid sha256 reference"}
	}

	return nil
}

// ValidateEntrypoint validates an execution entrypoint record conforming to
// specs/common.schema.json#/$defs/entrypoint.
// It supports http, cli, cron, and event entrypoint variants with their respective
// required fields, pattern matches, enum bounds, and rejection of unexpected fields.
func ValidateEntrypoint(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "entrypoint", Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &CommonDefError{Def: "entrypoint", Reason: "object required", Cause: err}
	}

	typeRaw, ok := record["type"]
	if !ok || string(typeRaw) == "null" {
		return &CommonDefError{Def: "entrypoint", Reason: "missing required field: type"}
	}

	var epType string
	if err := json.Unmarshal(typeRaw, &epType); err != nil {
		return &CommonDefError{Def: "entrypoint", Reason: "type must be a string", Cause: err}
	}

	switch epType {
	case "http":
		for _, req := range []string{"type", "method", "path"} {
			v, present := record[req]
			if !present || string(v) == "null" {
				return &CommonDefError{Def: "entrypoint", Reason: "missing required field: " + req}
			}
		}
		for k := range record {
			if k != "type" && k != "method" && k != "path" && k != "content_types" {
				return &CommonDefError{Def: "entrypoint", Reason: "unexpected field in http entrypoint"}
			}
		}
		var method string
		if err := json.Unmarshal(record["method"], &method); err != nil {
			return &CommonDefError{Def: "entrypoint", Reason: "method must be a string", Cause: err}
		}
		if _, ok := allowedHTTPMethods[method]; !ok {
			return &CommonDefError{Def: "entrypoint", Reason: "invalid HTTP method"}
		}
		var path string
		if err := json.Unmarshal(record["path"], &path); err != nil {
			return &CommonDefError{Def: "entrypoint", Reason: "path must be a string", Cause: err}
		}
		if len(path) < 1 {
			return &CommonDefError{Def: "entrypoint", Reason: "path must not be empty"}
		}
		if ctRaw, ok := record["content_types"]; ok {
			if string(ctRaw) == "null" {
				return &CommonDefError{Def: "entrypoint", Reason: "content_types cannot be null"}
			}
			var contentTypes []string
			if err := json.Unmarshal(ctRaw, &contentTypes); err != nil {
				return &CommonDefError{Def: "entrypoint", Reason: "content_types must be an array of strings", Cause: err}
			}
			for _, ct := range contentTypes {
				if len(ct) < 1 {
					return &CommonDefError{Def: "entrypoint", Reason: "content_types item must not be empty"}
				}
			}
		}

	case "cli":
		for _, req := range []string{"type", "command_id"} {
			v, present := record[req]
			if !present || string(v) == "null" {
				return &CommonDefError{Def: "entrypoint", Reason: "missing required field: " + req}
			}
		}
		for k := range record {
			if k != "type" && k != "command_id" && k != "arguments_schema" {
				return &CommonDefError{Def: "entrypoint", Reason: "unexpected field in cli entrypoint"}
			}
		}
		var cmdID string
		if err := json.Unmarshal(record["command_id"], &cmdID); err != nil || !validAssessmentReportID(cmdID) {
			return &CommonDefError{Def: "entrypoint", Reason: "command_id must match pattern ^[a-z0-9_-]+$"}
		}
		if argRaw, ok := record["arguments_schema"]; ok {
			if string(argRaw) == "null" {
				return &CommonDefError{Def: "entrypoint", Reason: "arguments_schema cannot be null"}
			}
			var argSchema map[string]json.RawMessage
			if err := json.Unmarshal(argRaw, &argSchema); err != nil || argSchema == nil {
				return &CommonDefError{Def: "entrypoint", Reason: "arguments_schema must be an object", Cause: err}
			}
		}

	case "cron":
		for _, req := range []string{"type", "command_id", "schedule", "timezone"} {
			v, present := record[req]
			if !present || string(v) == "null" {
				return &CommonDefError{Def: "entrypoint", Reason: "missing required field: " + req}
			}
		}
		if len(record) != 4 {
			return &CommonDefError{Def: "entrypoint", Reason: "unexpected field in cron entrypoint"}
		}
		var cmdID string
		if err := json.Unmarshal(record["command_id"], &cmdID); err != nil || !validAssessmentReportID(cmdID) {
			return &CommonDefError{Def: "entrypoint", Reason: "command_id must match pattern ^[a-z0-9_-]+$"}
		}
		var schedule string
		if err := json.Unmarshal(record["schedule"], &schedule); err != nil || len(schedule) < 1 {
			return &CommonDefError{Def: "entrypoint", Reason: "schedule must be a non-empty string"}
		}
		var timezone string
		if err := json.Unmarshal(record["timezone"], &timezone); err != nil || len(timezone) < 1 {
			return &CommonDefError{Def: "entrypoint", Reason: "timezone must be a non-empty string"}
		}

	case "event":
		for _, req := range []string{"type", "channel", "message_schema"} {
			v, present := record[req]
			if !present || string(v) == "null" {
				return &CommonDefError{Def: "entrypoint", Reason: "missing required field: " + req}
			}
		}
		if len(record) != 3 {
			return &CommonDefError{Def: "entrypoint", Reason: "unexpected field in event entrypoint"}
		}
		var channel string
		if err := json.Unmarshal(record["channel"], &channel); err != nil || len(channel) < 1 {
			return &CommonDefError{Def: "entrypoint", Reason: "channel must be a non-empty string"}
		}
		var msgSchema map[string]json.RawMessage
		if err := json.Unmarshal(record["message_schema"], &msgSchema); err != nil || msgSchema == nil {
			return &CommonDefError{Def: "entrypoint", Reason: "message_schema must be an object"}
		}

	default:
		return &CommonDefError{Def: "entrypoint", Reason: "unsupported entrypoint type"}
	}

	return nil
}

// ValidateEvidence validates an evidence record conforming to
// specs/common.schema.json#/$defs/evidence.
func ValidateEvidence(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "evidence", Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &CommonDefError{Def: "evidence", Reason: "object required", Cause: err}
	}

	for _, req := range []string{"evidence_id", "kind", "artifact_ref", "description"} {
		v, present := record[req]
		if !present || string(v) == "null" {
			return &CommonDefError{Def: "evidence", Reason: "missing required field: " + req}
		}
	}

	for k := range record {
		if k != "evidence_id" && k != "kind" && k != "artifact_ref" && k != "description" && k != "symbol_id" && k != "trace_id" {
			return &CommonDefError{Def: "evidence", Reason: "unexpected field: " + k}
		}
	}

	var evID string
	if err := json.Unmarshal(record["evidence_id"], &evID); err != nil || !validAssessmentReportID(evID) {
		return &CommonDefError{Def: "evidence", Reason: "evidence_id must match pattern ^[a-z0-9_-]+$"}
	}

	var kind string
	if err := json.Unmarshal(record["kind"], &kind); err != nil {
		return &CommonDefError{Def: "evidence", Reason: "kind must be a string", Cause: err}
	}
	if _, ok := allowedEvidenceKinds[kind]; !ok {
		return &CommonDefError{Def: "evidence", Reason: "invalid evidence kind"}
	}

	if !validProfileRef(record["artifact_ref"], false) {
		return &CommonDefError{Def: "evidence", Reason: "artifact_ref must be a valid sha256 reference"}
	}

	var desc string
	if err := json.Unmarshal(record["description"], &desc); err != nil || len(desc) < 1 {
		return &CommonDefError{Def: "evidence", Reason: "description must be a non-empty string"}
	}

	if symRaw, ok := record["symbol_id"]; ok && string(symRaw) != "null" {
		var symID string
		if err := json.Unmarshal(symRaw, &symID); err != nil || !validAssessmentReportID(symID) {
			return &CommonDefError{Def: "evidence", Reason: "symbol_id must match pattern ^[a-z0-9_-]+$"}
		}
	}

	if trRaw, ok := record["trace_id"]; ok && string(trRaw) != "null" {
		var trID string
		if err := json.Unmarshal(trRaw, &trID); err != nil || !validAssessmentReportID(trID) {
			return &CommonDefError{Def: "evidence", Reason: "trace_id must match pattern ^[a-z0-9_-]+$"}
		}
	}

	return nil
}

// ValidateCommonError validates a common error record conforming to
// specs/common.schema.json#/$defs/error.
func ValidateCommonError(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "error", Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &CommonDefError{Def: "error", Reason: "object required", Cause: err}
	}

	for _, req := range []string{"code", "message", "category", "retryable", "recommended_action"} {
		v, present := record[req]
		if !present || string(v) == "null" {
			return &CommonDefError{Def: "error", Reason: "missing required field: " + req}
		}
	}

	for k := range record {
		if k != "code" && k != "message" && k != "category" && k != "retryable" && k != "recommended_action" && k != "details" && k != "artifact_ref" {
			return &CommonDefError{Def: "error", Reason: "unexpected field: " + k}
		}
	}

	var code string
	if err := json.Unmarshal(record["code"], &code); err != nil || len(code) < 1 {
		return &CommonDefError{Def: "error", Reason: "code must be a non-empty string"}
	}

	var message string
	if err := json.Unmarshal(record["message"], &message); err != nil || len(message) < 1 {
		return &CommonDefError{Def: "error", Reason: "message must be a non-empty string"}
	}

	var category string
	if err := json.Unmarshal(record["category"], &category); err != nil {
		return &CommonDefError{Def: "error", Reason: "category must be a string", Cause: err}
	}
	if _, ok := allowedErrorCategories[category]; !ok {
		return &CommonDefError{Def: "error", Reason: "invalid error category"}
	}

	var retryable bool
	if err := json.Unmarshal(record["retryable"], &retryable); err != nil {
		return &CommonDefError{Def: "error", Reason: "retryable must be a boolean", Cause: err}
	}

	var action string
	if err := json.Unmarshal(record["recommended_action"], &action); err != nil || len(action) < 1 {
		return &CommonDefError{Def: "error", Reason: "recommended_action must be a non-empty string"}
	}

	if detRaw, ok := record["details"]; ok && string(detRaw) != "null" {
		var details map[string]json.RawMessage
		if err := json.Unmarshal(detRaw, &details); err != nil || details == nil {
			return &CommonDefError{Def: "error", Reason: "details must be an object"}
		}
	}

	if artRaw, ok := record["artifact_ref"]; ok && string(artRaw) != "null" {
		if !validProfileRef(artRaw, false) {
			return &CommonDefError{Def: "error", Reason: "artifact_ref must be a valid sha256 reference"}
		}
	}

	return nil
}

// ValidateCompleteness validates a completeness record conforming to
// specs/common.schema.json#/$defs/completeness.
func ValidateCompleteness(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "completeness", Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &CommonDefError{Def: "completeness", Reason: "object required", Cause: err}
	}

	if len(record) != len(completenessFields) {
		return &CommonDefError{Def: "completeness", Reason: "must have exactly 11 completeness dimensions"}
	}

	for _, field := range completenessFields {
		v, ok := record[field]
		if !ok || string(v) == "null" {
			return &CommonDefError{Def: "completeness", Reason: "missing required field: " + field}
		}
		var state string
		if err := json.Unmarshal(v, &state); err != nil {
			return &CommonDefError{Def: "completeness", Reason: field + " must be a string", Cause: err}
		}
		if _, ok := allowedCompletenessStates[state]; !ok {
			return &CommonDefError{Def: "completeness", Reason: "invalid completeness state in " + field}
		}
	}

	return nil
}

// ValidateHeaders validates a HTTP headers map conforming to
// specs/common.schema.json#/$defs/headers.
func ValidateHeaders(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "headers", Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &CommonDefError{Def: "headers", Reason: "object required", Cause: err}
	}

	for key, val := range record {
		if len(key) == 0 {
			return &CommonDefError{Def: "headers", Reason: "header key cannot be empty"}
		}
		if string(val) == "null" {
			return &CommonDefError{Def: "headers", Reason: "header value cannot be null"}
		}
		var strVal string
		if err := json.Unmarshal(val, &strVal); err == nil {
			if len(strVal) < 1 {
				return &CommonDefError{Def: "headers", Reason: "header string value must not be empty"}
			}
			continue
		}
		var arrVal []string
		if err := json.Unmarshal(val, &arrVal); err == nil {
			for _, item := range arrVal {
				if len(item) < 1 {
					return &CommonDefError{Def: "headers", Reason: "header array item must not be empty"}
				}
			}
			continue
		}
		return &CommonDefError{Def: "headers", Reason: "header value must be a non-empty string or array of non-empty strings"}
	}

	return nil
}

// ValidateCommonID validates an identifier conforming to
// specs/common.schema.json#/$defs/id.
func ValidateCommonID(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "id", Reason: "canonical JSON", Cause: err}
	}
	var id string
	if err := json.Unmarshal(canonical, &id); err != nil {
		return &CommonDefError{Def: "id", Reason: "must be a string", Cause: err}
	}
	if !validAssessmentReportID(id) {
		return &CommonDefError{Def: "id", Reason: "must match pattern ^[a-z0-9_-]+$"}
	}
	return nil
}

// ValidateCommonHash validates a SHA-256 hash string conforming to
// specs/common.schema.json#/$defs/hash.
func ValidateCommonHash(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "hash", Reason: "canonical JSON", Cause: err}
	}
	var h string
	if err := json.Unmarshal(canonical, &h); err != nil {
		return &CommonDefError{Def: "hash", Reason: "must be a string", Cause: err}
	}
	if !isManifestDigest(h) {
		return &CommonDefError{Def: "hash", Reason: "must match pattern ^[a-f0-9]{64}$"}
	}
	return nil
}

// ValidateCommonArtifactRef validates an artifact reference conforming to
// specs/common.schema.json#/$defs/artifact_ref.
func ValidateCommonArtifactRef(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &CommonDefError{Def: "artifact_ref", Reason: "canonical JSON", Cause: err}
	}
	var ref string
	if err := json.Unmarshal(canonical, &ref); err != nil {
		return &CommonDefError{Def: "artifact_ref", Reason: "must be a string", Cause: err}
	}
	if !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
		return &CommonDefError{Def: "artifact_ref", Reason: "must match pattern ^sha256:[a-f0-9]{64}$"}
	}
	return nil
}

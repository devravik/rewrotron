package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

var allowedDetectionEvidenceKinds = map[string]struct{}{
	"STATIC":         {},
	"RUNTIME":        {},
	"USER_OVERRIDE":  {},
	"AGENT_PROPOSAL": {},
}

var allowedDetectionGateStatuses = map[string]struct{}{
	"AWAITING_SOURCE_CONFIRMATION": {},
	"SOURCE_CONFIRMED":             {},
}

// SourceDetectionError describes an invalid source detection record without echoing private contents.
type SourceDetectionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed source detection obligation.
func (e *SourceDetectionError) Error() string {
	return "invalid source detection: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *SourceDetectionError) Unwrap() error { return e.Cause }

// ValidateSourceDetection validates a source detection record conforming to
// specs/schemas/detection-schema.json (schema_version 1.0 or 2.0).
// In version 1.0, gate_status is required. In version 2.0, gate_status is explicitly
// forbidden because Gate 1 is an explicit human confirmation decision, not a detected fact.
// It verifies required fields, confidence scores (0.0 <= score <= 1.0), structured evidence items,
// non-negative count metrics, and strict JSON error boundaries.
func ValidateSourceDetection(raw []byte) error {
	record, err := parseSourceDetectionJSON(raw)
	if err != nil {
		return err
	}

	var version string
	if string(record["schema_version"]) == "null" || json.Unmarshal(record["schema_version"], &version) != nil {
		return &SourceDetectionError{Reason: "schema_version"}
	}
	if version != "1.0" && version != "2.0" {
		return &SourceDetectionError{Reason: "unsupported schema_version"}
	}

	requiredFields := []string{
		"schema_version",
		"detection_id",
		"project_id",
		"source_manifest_hash",
		"language",
		"framework",
		"application_type",
		"confidence",
		"evidence",
	}
	if version == "1.0" {
		requiredFields = append(requiredFields, "gate_status")
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &SourceDetectionError{Reason: "missing required field: " + f}
		}
	}

	allowedOptionalFields := map[string]struct{}{
		"detected_version":           {},
		"php_compatibility":          {},
		"database_engine":            {},
		"authentication_type":        {},
		"potential_concerns":         {},
		"routes_count":               {},
		"controllers_count":          {},
		"models_count":               {},
		"views_count":                {},
		"external_integrations_count": {},
		"cron_entrypoints_count":     {},
	}

	reqMap := make(map[string]struct{}, len(requiredFields))
	for _, f := range requiredFields {
		reqMap[f] = struct{}{}
	}

	for k := range record {
		if _, req := reqMap[k]; req {
			continue
		}
		if _, opt := allowedOptionalFields[k]; opt {
			continue
		}
		return &SourceDetectionError{Reason: "unexpected field: " + k}
	}

	var detectionID string
	if string(record["detection_id"]) == "null" || json.Unmarshal(record["detection_id"], &detectionID) != nil || !validAssessmentReportID(detectionID) {
		return &SourceDetectionError{Reason: "invalid detection_id"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &SourceDetectionError{Reason: "invalid project_id"}
	}

	var manifestHash string
	if string(record["source_manifest_hash"]) == "null" || json.Unmarshal(record["source_manifest_hash"], &manifestHash) != nil || !isManifestDigest(manifestHash) {
		return &SourceDetectionError{Reason: "invalid source_manifest_hash"}
	}

	var language string
	if string(record["language"]) == "null" || json.Unmarshal(record["language"], &language) != nil || len(language) < 1 {
		return &SourceDetectionError{Reason: "invalid language"}
	}

	var framework string
	if string(record["framework"]) == "null" || json.Unmarshal(record["framework"], &framework) != nil || len(framework) < 1 {
		return &SourceDetectionError{Reason: "invalid framework"}
	}

	if val, present := record["detected_version"]; present && string(val) != "null" {
		var dv string
		if json.Unmarshal(val, &dv) != nil {
			return &SourceDetectionError{Reason: "invalid detected_version"}
		}
	}

	if val, present := record["php_compatibility"]; present && string(val) != "null" {
		var pc string
		if json.Unmarshal(val, &pc) != nil {
			return &SourceDetectionError{Reason: "invalid php_compatibility"}
		}
	}

	var appType string
	if string(record["application_type"]) == "null" || json.Unmarshal(record["application_type"], &appType) != nil || len(appType) < 1 {
		return &SourceDetectionError{Reason: "invalid application_type"}
	}

	if val, present := record["database_engine"]; present && string(val) != "null" {
		var db string
		if json.Unmarshal(val, &db) != nil {
			return &SourceDetectionError{Reason: "invalid database_engine"}
		}
	}

	if val, present := record["authentication_type"]; present && string(val) != "null" {
		var auth string
		if json.Unmarshal(val, &auth) != nil || len(auth) < 1 {
			return &SourceDetectionError{Reason: "invalid authentication_type"}
		}
	}

	if err := validateDetectionConfidence(record["confidence"]); err != nil {
		return err
	}

	if val, present := record["potential_concerns"]; present && string(val) != "null" {
		var concerns []string
		if json.Unmarshal(val, &concerns) != nil {
			return &SourceDetectionError{Reason: "invalid potential_concerns"}
		}
		for _, c := range concerns {
			if len(c) < 1 {
				return &SourceDetectionError{Reason: "potential_concerns item must not be empty"}
			}
		}
	}

	if version == "1.0" {
		var gateStatus string
		if string(record["gate_status"]) == "null" || json.Unmarshal(record["gate_status"], &gateStatus) != nil {
			return &SourceDetectionError{Reason: "invalid gate_status"}
		}
		if _, ok := allowedDetectionGateStatuses[gateStatus]; !ok {
			return &SourceDetectionError{Reason: "invalid gate_status enum"}
		}
	}

	if err := validateDetectionEvidence(record["evidence"]); err != nil {
		return err
	}

	countFields := []string{
		"routes_count",
		"controllers_count",
		"models_count",
		"views_count",
		"external_integrations_count",
		"cron_entrypoints_count",
	}
	for _, cf := range countFields {
		if val, present := record[cf]; present && string(val) != "null" {
			var count int64
			if json.Unmarshal(val, &count) != nil || count < 0 {
				return &SourceDetectionError{Reason: cf + " must be integer >= 0"}
			}
			if strings.ContainsAny(string(val), ".eE") {
				return &SourceDetectionError{Reason: cf + " must be integer"}
			}
		}
	}

	return nil
}

func validateDetectionConfidence(raw json.RawMessage) error {
	var conf map[string]float64
	if string(raw) == "null" || json.Unmarshal(raw, &conf) != nil || conf == nil {
		return &SourceDetectionError{Reason: "confidence must be an object"}
	}
	for _, score := range conf {
		if score < 0.0 || score > 1.0 {
			return &SourceDetectionError{Reason: "confidence score must be between 0 and 1"}
		}
	}
	return nil
}

func validateDetectionEvidence(raw json.RawMessage) error {
	var evidenceList []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &evidenceList) != nil {
		return &SourceDetectionError{Reason: "evidence must be an array"}
	}

	requiredEvidenceFields := []string{
		"evidence_id",
		"kind",
		"artifact_ref",
		"description",
	}

	allowedOptionalEvidenceFields := map[string]struct{}{
		"symbol_id": {},
		"trace_id":  {},
	}

	for _, rawEv := range evidenceList {
		var evObj map[string]json.RawMessage
		if json.Unmarshal(rawEv, &evObj) != nil || evObj == nil {
			return &SourceDetectionError{Reason: "evidence item must be an object"}
		}

		for _, req := range requiredEvidenceFields {
			if _, present := evObj[req]; !present {
				return &SourceDetectionError{Reason: "missing required evidence field: " + req}
			}
		}

		for k := range evObj {
			if k == "evidence_id" || k == "kind" || k == "artifact_ref" || k == "description" {
				continue
			}
			if _, opt := allowedOptionalEvidenceFields[k]; opt {
				continue
			}
			return &SourceDetectionError{Reason: "unexpected evidence field: " + k}
		}

		var evID string
		if string(evObj["evidence_id"]) == "null" || json.Unmarshal(evObj["evidence_id"], &evID) != nil || !validAssessmentReportID(evID) {
			return &SourceDetectionError{Reason: "invalid evidence_id"}
		}

		var kind string
		if string(evObj["kind"]) == "null" || json.Unmarshal(evObj["kind"], &kind) != nil {
			return &SourceDetectionError{Reason: "invalid evidence kind"}
		}
		if _, ok := allowedDetectionEvidenceKinds[kind]; !ok {
			return &SourceDetectionError{Reason: "invalid evidence kind enum"}
		}

		if !validProfileRef(evObj["artifact_ref"], false) {
			return &SourceDetectionError{Reason: "invalid evidence artifact_ref"}
		}

		var desc string
		if string(evObj["description"]) == "null" || json.Unmarshal(evObj["description"], &desc) != nil || len(desc) < 1 {
			return &SourceDetectionError{Reason: "invalid evidence description"}
		}

		if val, present := evObj["symbol_id"]; present && string(val) != "null" {
			var symID string
			if json.Unmarshal(val, &symID) != nil || !validAssessmentReportID(symID) {
				return &SourceDetectionError{Reason: "invalid evidence symbol_id"}
			}
		}

		if val, present := evObj["trace_id"]; present && string(val) != "null" {
			var traceID string
			if json.Unmarshal(val, &traceID) != nil || !validAssessmentReportID(traceID) {
				return &SourceDetectionError{Reason: "invalid evidence trace_id"}
			}
		}
	}
	return nil
}

func parseSourceDetectionJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &SourceDetectionError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &SourceDetectionError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &SourceDetectionError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &SourceDetectionError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &SourceDetectionError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &SourceDetectionError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &SourceDetectionError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &SourceDetectionError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &SourceDetectionError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &SourceDetectionError{Reason: "trailing input"}
	}
	return record, nil
}

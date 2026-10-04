package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

var allowedProgressFreshness = map[string]struct{}{
	"CURRENT": {},
	"STALE":   {},
	"UNKNOWN": {},
}

// ExecutionProgressError describes an invalid execution progress record without echoing private contents.
type ExecutionProgressError struct {
	Reason string
	Cause  error
}

// Error identifies the failed execution progress obligation.
func (e *ExecutionProgressError) Error() string {
	return "invalid execution progress: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ExecutionProgressError) Unwrap() error { return e.Cause }

// ValidateExecutionProgress validates a measured execution progress record conforming to
// specs/execution/progress.schema.json.
// It verifies valid UTF-8 JSON without duplicate keys or trailing data, required top-level
// fields, metric identifier patterns, non-empty definition and unit, valid numerator/denominator/percent
// bounds, scope hash and basis reference constraints, positive accounting revision, non-negative
// evidence sequence, freshness enum, and strict cross-field accounting invariants.
func ValidateExecutionProgress(raw []byte) error {
	record, err := parseProgressJSON(raw)
	if err != nil {
		return err
	}

	expectedFields := []string{
		"metric_id",
		"definition",
		"numerator",
		"denominator",
		"unit",
		"percent",
		"scope_hash",
		"basis_ref",
		"accounting_revision",
		"evidence_sequence",
		"freshness",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ExecutionProgressError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ExecutionProgressError{Reason: "object fields"}
	}

	var metricID string
	if string(record["metric_id"]) == "null" || json.Unmarshal(record["metric_id"], &metricID) != nil || !validAssessmentReportID(metricID) {
		return &ExecutionProgressError{Reason: "metric identity"}
	}

	var definition string
	if string(record["definition"]) == "null" || json.Unmarshal(record["definition"], &definition) != nil || strings.TrimSpace(definition) == "" {
		return &ExecutionProgressError{Reason: "definition"}
	}

	var numerator *int64
	if string(record["numerator"]) != "null" {
		var n int64
		if json.Unmarshal(record["numerator"], &n) != nil || n < 0 {
			return &ExecutionProgressError{Reason: "numerator"}
		}
		numerator = &n
	}

	var denominator *int64
	if string(record["denominator"]) != "null" {
		var d int64
		if json.Unmarshal(record["denominator"], &d) != nil || d < 0 {
			return &ExecutionProgressError{Reason: "denominator"}
		}
		denominator = &d
	}

	var unit string
	if string(record["unit"]) == "null" || json.Unmarshal(record["unit"], &unit) != nil || strings.TrimSpace(unit) == "" {
		return &ExecutionProgressError{Reason: "unit"}
	}

	var percent *float64
	if string(record["percent"]) != "null" {
		var p float64
		if json.Unmarshal(record["percent"], &p) != nil || p < 0 || p > 100 {
			return &ExecutionProgressError{Reason: "percent"}
		}
		percent = &p
	}

	var scopeHash *string
	if string(record["scope_hash"]) != "null" {
		var sh string
		if json.Unmarshal(record["scope_hash"], &sh) != nil || !isManifestDigest(sh) {
			return &ExecutionProgressError{Reason: "scope_hash"}
		}
		scopeHash = &sh
	}

	var basisRef *string
	if string(record["basis_ref"]) != "null" {
		if !validProfileRef(record["basis_ref"], false) {
			return &ExecutionProgressError{Reason: "basis_ref"}
		}
		br := profileString(record["basis_ref"])
		basisRef = &br
	}

	var accountingRevision int64
	if string(record["accounting_revision"]) == "null" || json.Unmarshal(record["accounting_revision"], &accountingRevision) != nil || accountingRevision < 1 {
		return &ExecutionProgressError{Reason: "accounting_revision"}
	}

	var evidenceSequence int64
	if string(record["evidence_sequence"]) == "null" || json.Unmarshal(record["evidence_sequence"], &evidenceSequence) != nil || evidenceSequence < 0 {
		return &ExecutionProgressError{Reason: "evidence_sequence"}
	}

	var freshness string
	if string(record["freshness"]) == "null" || json.Unmarshal(record["freshness"], &freshness) != nil {
		return &ExecutionProgressError{Reason: "freshness"}
	}
	if _, ok := allowedProgressFreshness[freshness]; !ok {
		return &ExecutionProgressError{Reason: "freshness"}
	}

	// Cross-field conditional invariants matching schema allOf
	if denominator != nil && *denominator == 0 {
		if numerator == nil || *numerator != 0 {
			return &ExecutionProgressError{Reason: "zero denominator requires zero numerator"}
		}
		if percent != nil {
			return &ExecutionProgressError{Reason: "zero denominator requires null percent"}
		}
	}

	if denominator == nil && percent != nil {
		return &ExecutionProgressError{Reason: "null denominator requires null percent"}
	}

	if numerator == nil && percent != nil {
		return &ExecutionProgressError{Reason: "null numerator requires null percent"}
	}

	if percent != nil {
		if denominator == nil || *denominator < 1 {
			return &ExecutionProgressError{Reason: "numeric percent requires denominator >= 1"}
		}
		if scopeHash == nil {
			return &ExecutionProgressError{Reason: "numeric percent requires scope_hash"}
		}
		if basisRef == nil {
			return &ExecutionProgressError{Reason: "numeric percent requires basis_ref"}
		}
	}

	return nil
}

func parseProgressJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ExecutionProgressError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ExecutionProgressError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ExecutionProgressError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ExecutionProgressError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ExecutionProgressError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ExecutionProgressError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ExecutionProgressError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ExecutionProgressError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ExecutionProgressError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ExecutionProgressError{Reason: "trailing input"}
	}
	return record, nil
}

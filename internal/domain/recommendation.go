package domain

import (
	"encoding/json"
	"strings"
)

var allowedRecommendationTypes = map[string]struct{}{
	"REQUIRED":      {},
	"RECOMMENDED":   {},
	"OPTIONAL":      {},
	"INFORMATIONAL": {},
	"BLOCKING":      {},
}

var allowedRecommendationRisks = map[string]struct{}{
	"LOW":     {},
	"MEDIUM":  {},
	"HIGH":    {},
	"UNKNOWN": {},
}

var allowedRecommendationDecisions = map[string]struct{}{
	"PENDING":    {},
	"ACCEPTED":   {},
	"REJECTED":   {},
	"DEFERRED":   {},
	"SUPERSEDED": {},
}

// RecommendationError describes an invalid target recommendation record without echoing private contents.
type RecommendationError struct {
	Reason string
	Cause  error
}

// Error identifies the failed recommendation obligation.
func (e *RecommendationError) Error() string {
	return "invalid recommendation: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *RecommendationError) Unwrap() error { return e.Cause }

// ValidateRecommendation validates a 1.0 evidence-backed target recommendation record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, revision monotonicity, target specification digest format, typed recommendation
// categories/types/risks/decisions, unique reasons and evidence references, and non-empty evidence.
func ValidateRecommendation(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &RecommendationError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &RecommendationError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"recommendation_id",
		"revision",
		"target_spec_hash",
		"category",
		"type",
		"current",
		"requested",
		"proposed_action",
		"recommended",
		"reasons",
		"risk",
		"affected_ids",
		"source_files",
		"evidence_refs",
		"required_actions",
		"alternatives",
		"decision",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &RecommendationError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &RecommendationError{Reason: "object fields"}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &RecommendationError{Reason: "unsupported schema version"}
	}

	if !validAssessmentReportID(profileString(record["recommendation_id"])) {
		return &RecommendationError{Reason: "recommendation identity"}
	}

	var revision int64
	if string(record["revision"]) == "null" || json.Unmarshal(record["revision"], &revision) != nil || revision < 1 {
		return &RecommendationError{Reason: "revision"}
	}

	targetHash := profileString(record["target_spec_hash"])
	if !isManifestDigest(targetHash) {
		return &RecommendationError{Reason: "target_spec_hash"}
	}

	category := profileString(record["category"])
	if strings.TrimSpace(category) == "" {
		return &RecommendationError{Reason: "category"}
	}

	recType := profileString(record["type"])
	if _, ok := allowedRecommendationTypes[recType]; !ok {
		return &RecommendationError{Reason: "invalid type"}
	}

	for _, nullableField := range []string{"current", "requested", "recommended"} {
		rawVal := record[nullableField]
		if string(rawVal) != "null" {
			var s string
			if json.Unmarshal(rawVal, &s) != nil {
				return &RecommendationError{Reason: nullableField}
			}
		}
	}

	action := profileString(record["proposed_action"])
	if strings.TrimSpace(action) == "" {
		return &RecommendationError{Reason: "proposed_action"}
	}

	if err := validateRecommendationStringList(record["reasons"], "reasons", 1); err != nil {
		return err
	}

	risk := profileString(record["risk"])
	if _, ok := allowedRecommendationRisks[risk]; !ok {
		return &RecommendationError{Reason: "invalid risk"}
	}

	if err := validateRecommendationIDList(record["affected_ids"], "affected_ids"); err != nil {
		return err
	}

	if err := validateRecommendationStringList(record["source_files"], "source_files", 0); err != nil {
		return err
	}

	if err := validateRecommendationRefList(record["evidence_refs"], "evidence_refs", 1); err != nil {
		return err
	}

	if err := validateRecommendationStringList(record["required_actions"], "required_actions", 0); err != nil {
		return err
	}

	if err := validateRecommendationStringList(record["alternatives"], "alternatives", 0); err != nil {
		return err
	}

	dec := profileString(record["decision"])
	if _, ok := allowedRecommendationDecisions[dec]; !ok {
		return &RecommendationError{Reason: "invalid decision"}
	}

	return nil
}

func validateRecommendationStringList(raw json.RawMessage, fieldName string, minItems int) error {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) < minItems {
		return &RecommendationError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		var s string
		if string(item) == "null" || json.Unmarshal(item, &s) != nil || strings.TrimSpace(s) == "" {
			return &RecommendationError{Reason: "invalid " + fieldName + " entry"}
		}
		if _, exists := seen[s]; exists {
			return &RecommendationError{Reason: "duplicate " + fieldName + " entry"}
		}
		seen[s] = struct{}{}
	}
	return nil
}

func validateRecommendationIDList(raw json.RawMessage, fieldName string) error {
	var ids []json.RawMessage
	if err := json.Unmarshal(raw, &ids); err != nil {
		return &RecommendationError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, item := range ids {
		id := profileString(item)
		if !validAssessmentReportID(id) {
			return &RecommendationError{Reason: "invalid " + fieldName + " identity"}
		}
		if _, exists := seen[id]; exists {
			return &RecommendationError{Reason: "duplicate " + fieldName + " identity"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateRecommendationRefList(raw json.RawMessage, fieldName string, minItems int) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil || len(refs) < minItems {
		return &RecommendationError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(refs))
	for _, item := range refs {
		if !validProfileRef(item, false) {
			return &RecommendationError{Reason: "invalid " + fieldName + " reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return &RecommendationError{Reason: "duplicate " + fieldName + " reference"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}

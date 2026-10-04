package domain

import (
	"encoding/json"
	"strings"
)

// ConsumptionCoverageError describes an invalid consumption-coverage record without echoing private contents.
type ConsumptionCoverageError struct {
	Reason string
	Cause  error
}

// Error identifies the failed consumption-coverage obligation.
func (e *ConsumptionCoverageError) Error() string {
	return "invalid consumption coverage: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ConsumptionCoverageError) Unwrap() error { return e.Cause }

// ValidateConsumptionCoverage validates a 1.0 consumption-coverage record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, positive prerequisite coverage semantics, category
// status closure, input registration, and closure controls for sufficient coverage.
func ValidateConsumptionCoverage(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ConsumptionCoverageError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ConsumptionCoverageError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"scope_ref",
		"discovery_method_ref",
		"analysis_revision_ref",
		"inputs_examined",
		"roots_examined",
		"categories",
		"unresolved_candidates",
		"excluded_mechanisms",
		"closure_control_refs",
		"status",
		"issuer",
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ConsumptionCoverageError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ConsumptionCoverageError{Reason: "object fields"}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &ConsumptionCoverageError{Reason: "unsupported schema version"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ConsumptionCoverageError{Reason: "project identity"}
	}

	for _, field := range []string{"scope_ref", "discovery_method_ref", "analysis_revision_ref"} {
		if !validProfileRef(record[field], false) {
			return &ConsumptionCoverageError{Reason: field}
		}
	}

	if profileString(record["issuer"]) != "ENGINE" {
		return &ConsumptionCoverageError{Reason: "issuer must be ENGINE"}
	}

	status := profileString(record["status"])
	if status != "SUFFICIENT" && status != "INSUFFICIENT" {
		return &ConsumptionCoverageError{Reason: "invalid status"}
	}

	unresolved, err := validateMechanismList(record["unresolved_candidates"], "unresolved candidates")
	if err != nil {
		return err
	}
	if status == "SUFFICIENT" && len(unresolved) > 0 {
		return &ConsumptionCoverageError{Reason: "sufficient coverage requires empty unresolved candidates"}
	}

	excluded, err := validateMechanismList(record["excluded_mechanisms"], "excluded mechanisms")
	if err != nil {
		return err
	}
	if status == "SUFFICIENT" && len(excluded) > 0 {
		return &ConsumptionCoverageError{Reason: "sufficient coverage requires empty excluded mechanisms"}
	}

	controls, err := validateClosureControlRefs(record["closure_control_refs"])
	if err != nil {
		return err
	}
	if status == "SUFFICIENT" && len(controls) == 0 {
		return &ConsumptionCoverageError{Reason: "sufficient coverage requires closure control references"}
	}

	inputsExamined, err := validateInputsExamined(record["inputs_examined"])
	if err != nil {
		return err
	}

	if err := validateRootsExamined(record["roots_examined"]); err != nil {
		return err
	}

	if err := validateCoverageCategories(record["categories"], status, inputsExamined); err != nil {
		return err
	}

	return nil
}

func validateInputsExamined(raw json.RawMessage) (map[string]struct{}, error) {
	var inputs []json.RawMessage
	if err := json.Unmarshal(raw, &inputs); err != nil {
		return nil, &ConsumptionCoverageError{Reason: "inputs examined array", Cause: err}
	}
	seen := make(map[string]struct{}, len(inputs))
	for _, item := range inputs {
		if !validProfileRef(item, false) {
			return nil, &ConsumptionCoverageError{Reason: "invalid input examined reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return nil, &ConsumptionCoverageError{Reason: "duplicate input examined reference"}
		}
		seen[ref] = struct{}{}
	}
	return seen, nil
}

func validateRootsExamined(raw json.RawMessage) error {
	var roots []json.RawMessage
	if err := json.Unmarshal(raw, &roots); err != nil {
		return &ConsumptionCoverageError{Reason: "roots examined array", Cause: err}
	}
	seen := make(map[string]struct{}, len(roots))
	for _, item := range roots {
		root := profileString(item)
		if !validAssessmentReportID(root) {
			return &ConsumptionCoverageError{Reason: "invalid root examined identity"}
		}
		if _, exists := seen[root]; exists {
			return &ConsumptionCoverageError{Reason: "duplicate root examined identity"}
		}
		seen[root] = struct{}{}
	}
	return nil
}

func validateMechanismList(raw json.RawMessage, fieldName string) ([]string, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, &ConsumptionCoverageError{Reason: fieldName + " array", Cause: err}
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		var s string
		if string(item) == "null" || json.Unmarshal(item, &s) != nil || strings.TrimSpace(s) == "" {
			return nil, &ConsumptionCoverageError{Reason: "invalid " + fieldName + " entry"}
		}
		if _, exists := seen[s]; exists {
			return nil, &ConsumptionCoverageError{Reason: "duplicate " + fieldName + " entry"}
		}
		seen[s] = struct{}{}
		result = append(result, s)
	}
	return result, nil
}

func validateClosureControlRefs(raw json.RawMessage) ([]string, error) {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, &ConsumptionCoverageError{Reason: "closure control references array", Cause: err}
	}
	seen := make(map[string]struct{}, len(refs))
	result := make([]string, 0, len(refs))
	for _, item := range refs {
		if !validProfileRef(item, false) {
			return nil, &ConsumptionCoverageError{Reason: "invalid closure control reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return nil, &ConsumptionCoverageError{Reason: "duplicate closure control reference"}
		}
		seen[ref] = struct{}{}
		result = append(result, ref)
	}
	return result, nil
}

var allowedCategories = map[string]struct{}{
	"SOURCE":      {},
	"ENVIRONMENT": {},
	"ADAPTER":     {},
	"CONTRACT":    {},
	"RESOLUTION":  {},
	"DEPENDENCY":  {},
	"COMPONENT":   {},
	"POLICY":      {},
	"KNOWLEDGE":   {},
	"SKILL":       {},
	"TOOL":        {},
}

var allowedCategoryStatuses = map[string]struct{}{
	"CLOSED":     {},
	"IRRELEVANT": {},
	"PARTIAL":    {},
	"UNKNOWN":    {},
}

func validateCoverageCategories(raw json.RawMessage, coverageStatus string, examinedInputs map[string]struct{}) error {
	var categories []json.RawMessage
	if err := json.Unmarshal(raw, &categories); err != nil {
		return &ConsumptionCoverageError{Reason: "categories array", Cause: err}
	}
	if len(categories) == 0 {
		return &ConsumptionCoverageError{Reason: "categories cannot be empty"}
	}

	seenCategories := make(map[string]struct{}, len(categories))
	for _, catRaw := range categories {
		var cat map[string]json.RawMessage
		if err := json.Unmarshal(catRaw, &cat); err != nil || cat == nil {
			return &ConsumptionCoverageError{Reason: "category object", Cause: err}
		}
		for _, required := range []string{"category", "status", "input_refs", "evidence_refs"} {
			if _, present := cat[required]; !present {
				return &ConsumptionCoverageError{Reason: "missing category field"}
			}
		}
		if len(cat) != 4 {
			return &ConsumptionCoverageError{Reason: "category fields"}
		}

		catName := profileString(cat["category"])
		if _, ok := allowedCategories[catName]; !ok {
			return &ConsumptionCoverageError{Reason: "unknown category"}
		}
		if _, exists := seenCategories[catName]; exists {
			return &ConsumptionCoverageError{Reason: "duplicate category"}
		}
		seenCategories[catName] = struct{}{}

		catStatus := profileString(cat["status"])
		if _, ok := allowedCategoryStatuses[catStatus]; !ok {
			return &ConsumptionCoverageError{Reason: "unknown category status"}
		}

		if coverageStatus == "SUFFICIENT" && catStatus != "CLOSED" && catStatus != "IRRELEVANT" {
			return &ConsumptionCoverageError{Reason: "sufficient coverage requires closed or irrelevant category status"}
		}

		var inputRefs []json.RawMessage
		if err := json.Unmarshal(cat["input_refs"], &inputRefs); err != nil {
			return &ConsumptionCoverageError{Reason: "category input references array", Cause: err}
		}
		seenInputs := make(map[string]struct{}, len(inputRefs))
		for _, inRaw := range inputRefs {
			if !validProfileRef(inRaw, false) {
				return &ConsumptionCoverageError{Reason: "invalid category input reference"}
			}
			inRef := profileString(inRaw)
			if _, exists := seenInputs[inRef]; exists {
				return &ConsumptionCoverageError{Reason: "duplicate category input reference"}
			}
			seenInputs[inRef] = struct{}{}
			if _, examined := examinedInputs[inRef]; !examined {
				return &ConsumptionCoverageError{Reason: "unregistered category input reference"}
			}
		}

		var evidenceRefs []json.RawMessage
		if err := json.Unmarshal(cat["evidence_refs"], &evidenceRefs); err != nil || len(evidenceRefs) < 1 {
			return &ConsumptionCoverageError{Reason: "category evidence references array", Cause: err}
		}
		seenEvidence := make(map[string]struct{}, len(evidenceRefs))
		for _, evRaw := range evidenceRefs {
			if !validProfileRef(evRaw, false) {
				return &ConsumptionCoverageError{Reason: "invalid category evidence reference"}
			}
			evRef := profileString(evRaw)
			if _, exists := seenEvidence[evRef]; exists {
				return &ConsumptionCoverageError{Reason: "duplicate category evidence reference"}
			}
			seenEvidence[evRef] = struct{}{}
		}
	}

	return nil
}

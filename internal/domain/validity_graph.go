package domain

import (
	"encoding/json"
)

var allowedEdgeKinds = map[string]struct{}{
	"SOURCE_IR":                    {},
	"SOURCE_CONTRACT":              {},
	"ENVIRONMENT_OBSERVATION":      {},
	"ADAPTER_EXTRACTION":           {},
	"DEPENDENCY_IMPLEMENTATION":    {},
	"COMPONENT_UNIT":               {},
	"NORMALIZATION_RESULT":         {},
	"KNOWLEDGE_ARTIFACT":           {},
	"SKILL_RESULT":                 {},
	"TOOL_EVIDENCE":                {},
	"INPUT_ASSESSMENT_FINDING":     {},
	"FINDING_ASSESSMENT_REPORT":    {},
	"ASSESSMENT_SELECTION":         {},
	"SELECTION_PLAN":               {},
	"IMPLEMENTATION_QUALIFICATION": {},
	"EXECUTION_RESULT":             {},
	"PERFORMANCE_RESULT":           {},
}

var allowedEdgePredicates = map[string]struct{}{
	"SAME_CONTENT":    {},
	"SAME_PROFILE":    {},
	"SAME_CAPABILITY": {},
	"SAME_RULES":      {},
	"SAME_FENCE":      {},
}

// ValidityGraphError describes an invalid validity-graph record without echoing private contents.
type ValidityGraphError struct {
	Reason string
	Cause  error
}

// Error identifies the failed validity-graph obligation.
func (e *ValidityGraphError) Error() string {
	return "invalid validity graph: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ValidityGraphError) Unwrap() error { return e.Cause }

// ValidateValidityGraph validates a 2.0 or 2.1 validity-graph record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, typed directed edges with valid predicates and witness references, self-loop
// cycle prevention, unique edge closures, unknown impact references, and required
// positive consumption coverage references for schema 2.1.
func ValidateValidityGraph(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ValidityGraphError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ValidityGraphError{Reason: "object fields", Cause: err}
	}

	versionRaw, present := record["schema_version"]
	if !present {
		return &ValidityGraphError{Reason: "missing required field"}
	}
	version := profileString(versionRaw)
	if version != "2.0" && version != "2.1" {
		return &ValidityGraphError{Reason: "unsupported schema version"}
	}

	expectedFields := []string{
		"schema_version",
		"graph_id",
		"project_id",
		"edges",
		"unknown_impact_refs",
	}
	if version == "2.1" {
		expectedFields = append(expectedFields, "consumption_coverage_refs")
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ValidityGraphError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ValidityGraphError{Reason: "object fields"}
	}

	if !validAssessmentReportID(profileString(record["graph_id"])) {
		return &ValidityGraphError{Reason: "graph identity"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ValidityGraphError{Reason: "project identity"}
	}

	if err := validateGraphEdges(record["edges"]); err != nil {
		return err
	}

	if err := validateGraphRefList(record["unknown_impact_refs"], "unknown_impact_refs", 0); err != nil {
		return err
	}

	if version == "2.1" {
		if err := validateGraphRefList(record["consumption_coverage_refs"], "consumption_coverage_refs", 0); err != nil {
			return err
		}
	}

	return nil
}

func validateGraphEdges(raw json.RawMessage) error {
	var edges []json.RawMessage
	if err := json.Unmarshal(raw, &edges); err != nil {
		return &ValidityGraphError{Reason: "edges array", Cause: err}
	}

	seenEdges := make(map[string]struct{}, len(edges))
	for _, edgeRaw := range edges {
		var edge map[string]json.RawMessage
		if err := json.Unmarshal(edgeRaw, &edge); err != nil || edge == nil {
			return &ValidityGraphError{Reason: "edge object", Cause: err}
		}

		expected := []string{"from_ref", "to_ref", "kind", "predicate", "witness_refs"}
		for _, field := range expected {
			if _, present := edge[field]; !present {
				return &ValidityGraphError{Reason: "missing edge field"}
			}
		}
		if len(edge) != len(expected) {
			return &ValidityGraphError{Reason: "edge fields"}
		}

		if !validProfileRef(edge["from_ref"], false) {
			return &ValidityGraphError{Reason: "edge from_ref"}
		}
		if !validProfileRef(edge["to_ref"], false) {
			return &ValidityGraphError{Reason: "edge to_ref"}
		}

		from := profileString(edge["from_ref"])
		to := profileString(edge["to_ref"])
		if from == to {
			return &ValidityGraphError{Reason: "edge cycle self-reference"}
		}

		kind := profileString(edge["kind"])
		if _, ok := allowedEdgeKinds[kind]; !ok {
			return &ValidityGraphError{Reason: "invalid edge kind"}
		}

		pred := profileString(edge["predicate"])
		if _, ok := allowedEdgePredicates[pred]; !ok {
			return &ValidityGraphError{Reason: "invalid edge predicate"}
		}

		edgeKey := from + "|" + to + "|" + kind + "|" + pred
		if _, exists := seenEdges[edgeKey]; exists {
			return &ValidityGraphError{Reason: "duplicate edge"}
		}
		seenEdges[edgeKey] = struct{}{}

		if err := validateGraphRefList(edge["witness_refs"], "witness_refs", 1); err != nil {
			return err
		}
	}
	return nil
}

func validateGraphRefList(raw json.RawMessage, fieldName string, minItems int) error {
	var refs []json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil || len(refs) < minItems {
		return &ValidityGraphError{Reason: fieldName + " array", Cause: err}
	}

	seen := make(map[string]struct{}, len(refs))
	for _, item := range refs {
		if !validProfileRef(item, false) {
			return &ValidityGraphError{Reason: "invalid " + fieldName + " reference"}
		}
		ref := profileString(item)
		if _, exists := seen[ref]; exists {
			return &ValidityGraphError{Reason: "duplicate " + fieldName + " reference"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}

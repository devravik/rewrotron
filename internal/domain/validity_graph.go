package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

// ValidityEdge represents an immutable directed edge from prerequisite to dependent in a validity graph.
type ValidityEdge struct {
	FromRef     string   `json:"from_ref"`
	ToRef       string   `json:"to_ref"`
	Kind        string   `json:"kind"`
	Predicate   string   `json:"predicate"`
	WitnessRefs []string `json:"witness_refs"`
}

// ValidityGraphRecord represents an immutable typed dependency graph artifact (schema 2.0 or 2.1).
type ValidityGraphRecord struct {
	SchemaVersion           string         `json:"schema_version"`
	GraphID                 string         `json:"graph_id"`
	ProjectID               string         `json:"project_id"`
	Edges                   []ValidityEdge `json:"edges"`
	UnknownImpactRefs       []string       `json:"unknown_impact_refs"`
	ConsumptionCoverageRefs []string       `json:"consumption_coverage_refs,omitempty"`
}

// CanonicalJSON renders the validity graph into RFC 8785 canonical bytes and validates it.
func (v *ValidityGraphRecord) CanonicalJSON() ([]byte, error) {
	if v == nil {
		return nil, &ValidityGraphError{Reason: "validity graph is nil"}
	}
	edges := v.Edges
	if edges == nil {
		edges = []ValidityEdge{}
	}
	unknownImpact := v.UnknownImpactRefs
	if unknownImpact == nil {
		unknownImpact = []string{}
	}

	recordMap := map[string]any{
		"schema_version":      v.SchemaVersion,
		"graph_id":            v.GraphID,
		"project_id":          v.ProjectID,
		"edges":               edges,
		"unknown_impact_refs": unknownImpact,
	}
	if v.SchemaVersion == "2.1" {
		coverageRefs := v.ConsumptionCoverageRefs
		if coverageRefs == nil {
			coverageRefs = []string{}
		}
		recordMap["consumption_coverage_refs"] = coverageRefs
	}

	raw, err := json.Marshal(recordMap)
	if err != nil {
		return nil, &ValidityGraphError{Reason: "failed to marshal validity graph", Cause: err}
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &ValidityGraphError{Reason: "failed to canonicalize validity graph JSON", Cause: err}
	}
	if err := ValidateValidityGraph(canon); err != nil {
		return nil, err
	}
	return canon, nil
}

// Digest computes the canonical sha256 reference for the validity graph record.
func (v *ValidityGraphRecord) Digest() (string, error) {
	canon, err := v.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canon)
	return fmt.Sprintf("sha256:%x", h), nil
}

// ParseValidityGraph parses and validates a validity-graph JSON record.
func ParseValidityGraph(raw []byte) (*ValidityGraphRecord, error) {
	if err := ValidateValidityGraph(raw); err != nil {
		return nil, err
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &ValidityGraphError{Reason: "canonical JSON", Cause: err}
	}
	var record ValidityGraphRecord
	if err := json.Unmarshal(canon, &record); err != nil {
		return nil, &ValidityGraphError{Reason: "unmarshal validity graph", Cause: err}
	}
	if record.Edges == nil {
		record.Edges = []ValidityEdge{}
	}
	if record.UnknownImpactRefs == nil {
		record.UnknownImpactRefs = []string{}
	}
	if record.SchemaVersion == "2.1" && record.ConsumptionCoverageRefs == nil {
		record.ConsumptionCoverageRefs = []string{}
	}
	return &record, nil
}

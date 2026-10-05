package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

var allowedEntrypointTypes = map[string]struct{}{
	"http":  {},
	"cli":   {},
	"cron":  {},
	"event": {},
}

var allowedCoverageStates = map[string]struct{}{
	"RESOLVED": {},
	"PARTIAL":  {},
	"UNKNOWN":  {},
}

var allowedComponentRoles = map[string]struct{}{
	"handler":     {},
	"service":     {},
	"data_model":  {},
	"renderer":    {},
	"integration": {},
}

// ApplicationIRError describes an invalid Application IR record without echoing private contents.
type ApplicationIRError struct {
	Reason string
	Cause  error
}

// Error identifies the failed application IR requirement.
func (e *ApplicationIRError) Error() string {
	return "invalid application IR: " + e.Reason
}

// Unwrap preserves decoding and canonicalization failures for inspection.
func (e *ApplicationIRError) Unwrap() error { return e.Cause }

// EntrypointType represents the mechanism of entry.
type EntrypointType string

const (
	EntrypointTypeHTTP  EntrypointType = "http"
	EntrypointTypeCLI   EntrypointType = "cli"
	EntrypointTypeCron  EntrypointType = "cron"
	EntrypointTypeEvent EntrypointType = "event"
)

// CoverageState represents coverage status of an entrypoint.
type CoverageState string

const (
	CoverageStateResolved CoverageState = "RESOLVED"
	CoverageStatePartial  CoverageState = "PARTIAL"
	CoverageStateUnknown  CoverageState = "UNKNOWN"
)

// ComponentRole represents an architectural role in Application IR.
type ComponentRole string

const (
	ComponentRoleHandler     ComponentRole = "handler"
	ComponentRoleService     ComponentRole = "service"
	ComponentRoleDataModel   ComponentRole = "data_model"
	ComponentRoleRenderer    ComponentRole = "renderer"
	ComponentRoleIntegration ComponentRole = "integration"
)

// ApplicationEntrypoint describes an HTTP, CLI, cron, or event entrypoint.
type ApplicationEntrypoint struct {
	EntrypointID     string          `json:"entrypoint_id"`
	Type             EntrypointType  `json:"type"`
	Dispatch         json.RawMessage `json:"dispatch"`
	HandlerSymbolIDs []string        `json:"handler_symbol_ids"`
	EvidenceIDs      []string        `json:"evidence_ids"`
	CoverageState    CoverageState   `json:"coverage_state"`
}

// ApplicationComponent describes a logical component grouping symbols.
type ApplicationComponent struct {
	ComponentID string        `json:"component_id"`
	Role        ComponentRole `json:"role"`
	SymbolIDs   []string      `json:"symbol_ids"`
	EvidenceIDs []string      `json:"evidence_ids"`
}

// ApplicationDependency describes a dependency edge in Application IR.
type ApplicationDependency struct {
	EdgeID      string   `json:"edge_id"`
	FromID      string   `json:"from_id"`
	ToID        *string  `json:"to_id"`
	Kind        string   `json:"kind"`
	Resolved    bool     `json:"resolved"`
	EvidenceIDs []string `json:"evidence_ids"`
	Expression  *string  `json:"expression,omitempty"`
	AffectedIDs []string `json:"affected_ids,omitempty"`
}

// ApplicationFinding records an application-level finding.
type ApplicationFinding struct {
	FindingID   string          `json:"finding_id"`
	Code        string          `json:"code"`
	Severity    FindingSeverity `json:"severity"`
	AffectedIDs []string        `json:"affected_ids"`
	EvidenceIDs []string        `json:"evidence_ids"`
	Message     string          `json:"message"`
}

// ApplicationEvidence records evidence for application elements.
type ApplicationEvidence struct {
	EvidenceID  string  `json:"evidence_id"`
	Kind        string  `json:"kind"`
	ArtifactRef string  `json:"artifact_ref"`
	SymbolID    *string `json:"symbol_id,omitempty"`
	TraceID     *string `json:"trace_id,omitempty"`
	Description string  `json:"description"`
}

// ApplicationIRDocument represents a complete Application IR document conforming to specs/ir/application-ir.schema.json.
type ApplicationIRDocument struct {
	SchemaVersion     string                  `json:"schema_version"`
	ProjectID         string                  `json:"project_id"`
	ConfigurationHash string                  `json:"configuration_hash"`
	DocumentID        string                  `json:"document_id"`
	SourceIRRef       string                  `json:"source_ir_ref"`
	Entrypoints       []ApplicationEntrypoint `json:"entrypoints"`
	Components        []ApplicationComponent  `json:"components"`
	Dependencies      []ApplicationDependency `json:"dependencies"`
	DataModels        []json.RawMessage       `json:"data_models"`
	Workflows         []json.RawMessage       `json:"workflows"`
	Findings          []ApplicationFinding    `json:"findings"`
	Evidence          []ApplicationEvidence   `json:"evidence"`
}

// CanonicalJSON renders the ApplicationIRDocument into canonical JSON bytes.
func (doc *ApplicationIRDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, &ApplicationIRError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical ApplicationIRDocument.
func (doc *ApplicationIRDocument) Digest() (string, error) {
	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ValidateApplicationIR validates raw bytes against specs/ir/application-ir.schema.json
// and domain invariants.
func ValidateApplicationIR(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ApplicationIRError{Reason: "canonical JSON", Cause: err}
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &root); err != nil || root == nil {
		return &ApplicationIRError{Reason: "object required", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"configuration_hash",
		"document_id",
		"source_ir_ref",
		"entrypoints",
		"components",
		"dependencies",
		"data_models",
		"workflows",
		"findings",
		"evidence",
	}

	if len(root) != len(expectedFields) {
		return &ApplicationIRError{Reason: "unexpected or missing root fields"}
	}
	for _, f := range expectedFields {
		v, ok := root[f]
		if !ok || string(v) == "null" {
			return &ApplicationIRError{Reason: "missing required field: " + f}
		}
	}

	var schemaVer string
	if err := json.Unmarshal(root["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &ApplicationIRError{Reason: "schema_version must be \"1.0\""}
	}

	var projectID string
	if err := json.Unmarshal(root["project_id"], &projectID); err != nil || !validAssessmentReportID(projectID) {
		return &ApplicationIRError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}

	var configHash string
	if err := json.Unmarshal(root["configuration_hash"], &configHash); err != nil || !isManifestDigest(configHash) {
		return &ApplicationIRError{Reason: "configuration_hash must match pattern ^[a-f0-9]{64}$"}
	}

	var docID string
	if err := json.Unmarshal(root["document_id"], &docID); err != nil || !validAssessmentReportID(docID) {
		return &ApplicationIRError{Reason: "document_id must match pattern ^[a-z0-9_-]+$"}
	}

	if !validProfileRef(root["source_ir_ref"], false) {
		return &ApplicationIRError{Reason: "source_ir_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}

	evidenceIDs, err := validateApplicationEvidence(root["evidence"])
	if err != nil {
		return err
	}

	if err := validateApplicationEntrypoints(root["entrypoints"], evidenceIDs); err != nil {
		return err
	}

	if err := validateApplicationComponents(root["components"], evidenceIDs); err != nil {
		return err
	}

	if err := validateApplicationDependencies(root["dependencies"], evidenceIDs); err != nil {
		return err
	}

	if err := validateApplicationDataModels(root["data_models"]); err != nil {
		return err
	}

	if err := validateApplicationWorkflows(root["workflows"]); err != nil {
		return err
	}

	if err := validateApplicationFindings(root["findings"], evidenceIDs); err != nil {
		return err
	}

	return nil
}

func validateApplicationEntrypoints(raw json.RawMessage, evidenceIDs map[string]struct{}) error {
	var entrypoints []json.RawMessage
	if err := json.Unmarshal(raw, &entrypoints); err != nil {
		return &ApplicationIRError{Reason: "entrypoints must be an array", Cause: err}
	}

	seenIDs := make(map[string]struct{}, len(entrypoints))
	for _, rawEp := range entrypoints {
		var ep map[string]json.RawMessage
		if err := json.Unmarshal(rawEp, &ep); err != nil || ep == nil {
			return &ApplicationIRError{Reason: "entrypoint entry must be an object", Cause: err}
		}
		if len(ep) != 6 {
			return &ApplicationIRError{Reason: "unexpected or missing fields in entrypoint entry"}
		}
		for _, req := range []string{
			"entrypoint_id", "type", "dispatch", "handler_symbol_ids", "evidence_ids", "coverage_state",
		} {
			v, ok := ep[req]
			if !ok || string(v) == "null" {
				return &ApplicationIRError{Reason: "missing required entrypoint field: " + req}
			}
		}

		var epID string
		if err := json.Unmarshal(ep["entrypoint_id"], &epID); err != nil || !validAssessmentReportID(epID) {
			return &ApplicationIRError{Reason: "entrypoint_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenIDs[epID]; exists {
			return &ApplicationIRError{Reason: fmt.Sprintf("duplicate entrypoint_id %q", epID)}
		}
		seenIDs[epID] = struct{}{}

		var epType string
		if err := json.Unmarshal(ep["type"], &epType); err != nil {
			return &ApplicationIRError{Reason: "entrypoint type must be a string", Cause: err}
		}
		if _, ok := allowedEntrypointTypes[epType]; !ok {
			return &ApplicationIRError{Reason: fmt.Sprintf("invalid entrypoint type %q", epType)}
		}

		if err := ValidateEntrypoint(ep["dispatch"]); err != nil {
			return &ApplicationIRError{Reason: "invalid entrypoint dispatch", Cause: err}
		}

		var dispatchObj map[string]json.RawMessage
		if err := json.Unmarshal(ep["dispatch"], &dispatchObj); err == nil {
			var dispType string
			if err := json.Unmarshal(dispatchObj["type"], &dispType); err == nil {
				if dispType != epType {
					return &ApplicationIRError{Reason: fmt.Sprintf("entrypoint type %q does not match dispatch type %q", epType, dispType)}
				}
			}
		}

		var symIDs []string
		if err := json.Unmarshal(ep["handler_symbol_ids"], &symIDs); err != nil {
			return &ApplicationIRError{Reason: "handler_symbol_ids must be an array of strings", Cause: err}
		}
		for _, sID := range symIDs {
			if !validAssessmentReportID(sID) {
				return &ApplicationIRError{Reason: "handler_symbol_ids item must match pattern ^[a-z0-9_-]+$"}
			}
		}

		var evIDs []string
		if err := json.Unmarshal(ep["evidence_ids"], &evIDs); err != nil {
			return &ApplicationIRError{Reason: "evidence_ids must be an array of strings", Cause: err}
		}
		for _, evID := range evIDs {
			if !validAssessmentReportID(evID) {
				return &ApplicationIRError{Reason: "evidence_ids item must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := evidenceIDs[evID]; !ok {
				return &ApplicationIRError{Reason: fmt.Sprintf("entrypoint evidence_id %q not found in evidence", evID)}
			}
		}

		var cov string
		if err := json.Unmarshal(ep["coverage_state"], &cov); err != nil {
			return &ApplicationIRError{Reason: "coverage_state must be a string", Cause: err}
		}
		if _, ok := allowedCoverageStates[cov]; !ok {
			return &ApplicationIRError{Reason: fmt.Sprintf("invalid coverage_state %q", cov)}
		}
	}

	return nil
}

func validateApplicationComponents(raw json.RawMessage, evidenceIDs map[string]struct{}) error {
	var components []json.RawMessage
	if err := json.Unmarshal(raw, &components); err != nil {
		return &ApplicationIRError{Reason: "components must be an array", Cause: err}
	}

	seenIDs := make(map[string]struct{}, len(components))
	for _, rawComp := range components {
		var comp map[string]json.RawMessage
		if err := json.Unmarshal(rawComp, &comp); err != nil || comp == nil {
			return &ApplicationIRError{Reason: "component entry must be an object", Cause: err}
		}
		if len(comp) != 4 {
			return &ApplicationIRError{Reason: "unexpected or missing fields in component entry"}
		}
		for _, req := range []string{"component_id", "role", "symbol_ids", "evidence_ids"} {
			v, ok := comp[req]
			if !ok || string(v) == "null" {
				return &ApplicationIRError{Reason: "missing required component field: " + req}
			}
		}

		var compID string
		if err := json.Unmarshal(comp["component_id"], &compID); err != nil || !validAssessmentReportID(compID) {
			return &ApplicationIRError{Reason: "component_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenIDs[compID]; exists {
			return &ApplicationIRError{Reason: fmt.Sprintf("duplicate component_id %q", compID)}
		}
		seenIDs[compID] = struct{}{}

		var role string
		if err := json.Unmarshal(comp["role"], &role); err != nil {
			return &ApplicationIRError{Reason: "component role must be a string", Cause: err}
		}
		if _, ok := allowedComponentRoles[role]; !ok {
			return &ApplicationIRError{Reason: fmt.Sprintf("invalid component role %q", role)}
		}

		var symIDs []string
		if err := json.Unmarshal(comp["symbol_ids"], &symIDs); err != nil {
			return &ApplicationIRError{Reason: "component symbol_ids must be an array of strings", Cause: err}
		}
		for _, sID := range symIDs {
			if !validAssessmentReportID(sID) {
				return &ApplicationIRError{Reason: "component symbol_ids item must match pattern ^[a-z0-9_-]+$"}
			}
		}

		var evIDs []string
		if err := json.Unmarshal(comp["evidence_ids"], &evIDs); err != nil {
			return &ApplicationIRError{Reason: "component evidence_ids must be an array of strings", Cause: err}
		}
		for _, evID := range evIDs {
			if !validAssessmentReportID(evID) {
				return &ApplicationIRError{Reason: "component evidence_ids item must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := evidenceIDs[evID]; !ok {
				return &ApplicationIRError{Reason: fmt.Sprintf("component evidence_id %q not found in evidence", evID)}
			}
		}
	}

	return nil
}

func validateApplicationDependencies(raw json.RawMessage, evidenceIDs map[string]struct{}) error {
	var deps []json.RawMessage
	if err := json.Unmarshal(raw, &deps); err != nil {
		return &ApplicationIRError{Reason: "dependencies must be an array", Cause: err}
	}

	seenIDs := make(map[string]struct{}, len(deps))
	for _, rawDep := range deps {
		var d map[string]json.RawMessage
		if err := json.Unmarshal(rawDep, &d); err != nil || d == nil {
			return &ApplicationIRError{Reason: "dependency entry must be an object", Cause: err}
		}

		for _, req := range []string{"edge_id", "from_id", "to_id", "kind", "resolved", "evidence_ids"} {
			v, ok := d[req]
			if !ok {
				return &ApplicationIRError{Reason: "missing required dependency field: " + req}
			}
			if req != "to_id" && string(v) == "null" {
				return &ApplicationIRError{Reason: "dependency field cannot be null: " + req}
			}
		}

		for k := range d {
			switch k {
			case "edge_id", "from_id", "to_id", "kind", "resolved", "evidence_ids", "expression", "affected_ids":
			default:
				return &ApplicationIRError{Reason: "unexpected field in dependency entry: " + k}
			}
		}

		var edgeID string
		if err := json.Unmarshal(d["edge_id"], &edgeID); err != nil || !validAssessmentReportID(edgeID) {
			return &ApplicationIRError{Reason: "dependency edge_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenIDs[edgeID]; exists {
			return &ApplicationIRError{Reason: fmt.Sprintf("duplicate dependency edge_id %q", edgeID)}
		}
		seenIDs[edgeID] = struct{}{}

		var fromID string
		if err := json.Unmarshal(d["from_id"], &fromID); err != nil || !validAssessmentReportID(fromID) {
			return &ApplicationIRError{Reason: "dependency from_id must match pattern ^[a-z0-9_-]+$"}
		}

		if string(d["to_id"]) != "null" {
			var toID string
			if err := json.Unmarshal(d["to_id"], &toID); err != nil || !validAssessmentReportID(toID) {
				return &ApplicationIRError{Reason: "dependency to_id must match pattern ^[a-z0-9_-]+$ or null"}
			}
		}

		var kind string
		if err := json.Unmarshal(d["kind"], &kind); err != nil || len(kind) < 1 {
			return &ApplicationIRError{Reason: "dependency kind must be a non-empty string"}
		}

		var resolved bool
		if err := json.Unmarshal(d["resolved"], &resolved); err != nil {
			return &ApplicationIRError{Reason: "dependency resolved must be a boolean", Cause: err}
		}

		var evIDs []string
		if err := json.Unmarshal(d["evidence_ids"], &evIDs); err != nil {
			return &ApplicationIRError{Reason: "dependency evidence_ids must be an array of strings", Cause: err}
		}
		for _, evID := range evIDs {
			if !validAssessmentReportID(evID) {
				return &ApplicationIRError{Reason: "dependency evidence_ids item must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := evidenceIDs[evID]; !ok {
				return &ApplicationIRError{Reason: fmt.Sprintf("dependency evidence_id %q not found in evidence", evID)}
			}
		}

		if exprRaw, ok := d["expression"]; ok && string(exprRaw) != "null" {
			var expr string
			if err := json.Unmarshal(exprRaw, &expr); err != nil || len(expr) < 1 {
				return &ApplicationIRError{Reason: "dependency expression must be a non-empty string"}
			}
		}

		if affRaw, ok := d["affected_ids"]; ok && string(affRaw) != "null" {
			var affIDs []string
			if err := json.Unmarshal(affRaw, &affIDs); err != nil {
				return &ApplicationIRError{Reason: "dependency affected_ids must be an array of strings", Cause: err}
			}
			for _, aID := range affIDs {
				if !validAssessmentReportID(aID) {
					return &ApplicationIRError{Reason: "dependency affected_ids item must match pattern ^[a-z0-9_-]+$"}
				}
			}
		}
	}

	return nil
}

func validateApplicationDataModels(raw json.RawMessage) error {
	var models []json.RawMessage
	if err := json.Unmarshal(raw, &models); err != nil {
		return &ApplicationIRError{Reason: "data_models must be an array", Cause: err}
	}

	seenModelIDs := make(map[string]struct{}, len(models))
	for _, rawModel := range models {
		if err := ValidateDataModel(rawModel); err != nil {
			return &ApplicationIRError{Reason: "invalid data model in application IR", Cause: err}
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(rawModel, &m); err == nil {
			var mID string
			if err := json.Unmarshal(m["model_id"], &mID); err == nil {
				if _, dup := seenModelIDs[mID]; dup {
					return &ApplicationIRError{Reason: fmt.Sprintf("duplicate data_model model_id %q", mID)}
				}
				seenModelIDs[mID] = struct{}{}
			}
		}
	}

	return nil
}

func validateApplicationWorkflows(raw json.RawMessage) error {
	var workflows []json.RawMessage
	if err := json.Unmarshal(raw, &workflows); err != nil {
		return &ApplicationIRError{Reason: "workflows must be an array", Cause: err}
	}

	seenWfIDs := make(map[string]struct{}, len(workflows))
	for _, rawWf := range workflows {
		if err := ValidateWorkflow(rawWf); err != nil {
			return &ApplicationIRError{Reason: "invalid workflow in application IR", Cause: err}
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(rawWf, &m); err == nil {
			var wfID string
			if err := json.Unmarshal(m["workflow_id"], &wfID); err == nil {
				if _, dup := seenWfIDs[wfID]; dup {
					return &ApplicationIRError{Reason: fmt.Sprintf("duplicate workflow_id %q", wfID)}
				}
				seenWfIDs[wfID] = struct{}{}
			}
		}
	}

	return nil
}

func validateApplicationFindings(raw json.RawMessage, evidenceIDs map[string]struct{}) error {
	var findings []json.RawMessage
	if err := json.Unmarshal(raw, &findings); err != nil {
		return &ApplicationIRError{Reason: "findings must be an array", Cause: err}
	}

	seenFindingIDs := make(map[string]struct{}, len(findings))
	for _, rawF := range findings {
		var f map[string]json.RawMessage
		if err := json.Unmarshal(rawF, &f); err != nil || f == nil {
			return &ApplicationIRError{Reason: "finding entry must be an object", Cause: err}
		}
		if len(f) != 6 {
			return &ApplicationIRError{Reason: "unexpected or missing fields in finding entry"}
		}
		for _, req := range []string{"finding_id", "code", "severity", "affected_ids", "evidence_ids", "message"} {
			v, ok := f[req]
			if !ok || string(v) == "null" {
				return &ApplicationIRError{Reason: "missing required finding field: " + req}
			}
		}

		var findingID string
		if err := json.Unmarshal(f["finding_id"], &findingID); err != nil || !validAssessmentReportID(findingID) {
			return &ApplicationIRError{Reason: "finding_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenFindingIDs[findingID]; exists {
			return &ApplicationIRError{Reason: fmt.Sprintf("duplicate finding_id %q", findingID)}
		}
		seenFindingIDs[findingID] = struct{}{}

		var code string
		if err := json.Unmarshal(f["code"], &code); err != nil || len(code) < 1 {
			return &ApplicationIRError{Reason: "finding code must be a non-empty string"}
		}

		var sev string
		if err := json.Unmarshal(f["severity"], &sev); err != nil {
			return &ApplicationIRError{Reason: "finding severity must be a string", Cause: err}
		}
		if _, ok := allowedFindingSeverities[sev]; !ok {
			return &ApplicationIRError{Reason: fmt.Sprintf("invalid finding severity %q", sev)}
		}

		var affIDs []string
		if err := json.Unmarshal(f["affected_ids"], &affIDs); err != nil {
			return &ApplicationIRError{Reason: "finding affected_ids must be an array of strings", Cause: err}
		}
		for _, aID := range affIDs {
			if !validAssessmentReportID(aID) {
				return &ApplicationIRError{Reason: "finding affected_ids item must match pattern ^[a-z0-9_-]+$"}
			}
		}

		var evIDs []string
		if err := json.Unmarshal(f["evidence_ids"], &evIDs); err != nil {
			return &ApplicationIRError{Reason: "finding evidence_ids must be an array of strings", Cause: err}
		}
		for _, evID := range evIDs {
			if !validAssessmentReportID(evID) {
				return &ApplicationIRError{Reason: "finding evidence_ids item must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := evidenceIDs[evID]; !ok {
				return &ApplicationIRError{Reason: fmt.Sprintf("finding evidence_id %q not found in evidence", evID)}
			}
		}

		var msg string
		if err := json.Unmarshal(f["message"], &msg); err != nil || len(msg) < 1 {
			return &ApplicationIRError{Reason: "finding message must be a non-empty string"}
		}
	}

	return nil
}

func validateApplicationEvidence(raw json.RawMessage) (map[string]struct{}, error) {
	var evidenceList []json.RawMessage
	if err := json.Unmarshal(raw, &evidenceList); err != nil {
		return nil, &ApplicationIRError{Reason: "evidence must be an array", Cause: err}
	}

	seenEvidenceIDs := make(map[string]struct{}, len(evidenceList))
	for _, rawEv := range evidenceList {
		var ev map[string]json.RawMessage
		if err := json.Unmarshal(rawEv, &ev); err != nil || ev == nil {
			return nil, &ApplicationIRError{Reason: "evidence entry must be an object", Cause: err}
		}

		for _, req := range []string{"evidence_id", "kind", "artifact_ref", "description"} {
			v, ok := ev[req]
			if !ok || string(v) == "null" {
				return nil, &ApplicationIRError{Reason: "missing required evidence field: " + req}
			}
		}

		for k := range ev {
			switch k {
			case "evidence_id", "kind", "artifact_ref", "description", "symbol_id", "trace_id":
			default:
				return nil, &ApplicationIRError{Reason: "unexpected field in evidence entry: " + k}
			}
		}

		var evID string
		if err := json.Unmarshal(ev["evidence_id"], &evID); err != nil || !validAssessmentReportID(evID) {
			return nil, &ApplicationIRError{Reason: "evidence_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenEvidenceIDs[evID]; exists {
			return nil, &ApplicationIRError{Reason: fmt.Sprintf("duplicate evidence_id %q", evID)}
		}
		seenEvidenceIDs[evID] = struct{}{}

		var kind string
		if err := json.Unmarshal(ev["kind"], &kind); err != nil {
			return nil, &ApplicationIRError{Reason: "evidence kind must be a string", Cause: err}
		}
		if _, ok := allowedEvidenceKinds[kind]; !ok {
			return nil, &ApplicationIRError{Reason: fmt.Sprintf("invalid evidence kind %q", kind)}
		}

		if !validProfileRef(ev["artifact_ref"], false) {
			return nil, &ApplicationIRError{Reason: "evidence artifact_ref must match pattern ^sha256:[a-f0-9]{64}$"}
		}

		var desc string
		if err := json.Unmarshal(ev["description"], &desc); err != nil || len(desc) < 1 {
			return nil, &ApplicationIRError{Reason: "evidence description must be a non-empty string"}
		}

		if symRaw, ok := ev["symbol_id"]; ok && string(symRaw) != "null" {
			var symID string
			if err := json.Unmarshal(symRaw, &symID); err != nil || !validAssessmentReportID(symID) {
				return nil, &ApplicationIRError{Reason: "evidence symbol_id must match pattern ^[a-z0-9_-]+$"}
			}
		}

		if trRaw, ok := ev["trace_id"]; ok && string(trRaw) != "null" {
			var trID string
			if err := json.Unmarshal(trRaw, &trID); err != nil || !validAssessmentReportID(trID) {
				return nil, &ApplicationIRError{Reason: "evidence trace_id must match pattern ^[a-z0-9_-]+$"}
			}
		}
	}

	return seenEvidenceIDs, nil
}

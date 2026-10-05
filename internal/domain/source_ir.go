package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

var allowedSyntaxStatuses = map[string]struct{}{
	"PARSED":      {},
	"PARTIAL":     {},
	"UNSUPPORTED": {},
	"FAILED":      {},
}

var allowedSymbolKinds = map[string]struct{}{
	"class":     {},
	"interface": {},
	"trait":     {},
	"function":  {},
	"method":    {},
	"property":  {},
	"constant":  {},
	"closure":   {},
	"script":    {},
}

var allowedFindingSeverities = map[string]struct{}{
	"CRITICAL": {},
	"HIGH":     {},
	"MEDIUM":   {},
	"LOW":      {},
}

// SourceIRError describes an invalid Source IR record without echoing private contents.
type SourceIRError struct {
	Reason string
	Cause  error
}

// Error identifies the failed source IR requirement.
func (e *SourceIRError) Error() string {
	return "invalid source IR: " + e.Reason
}

// Unwrap preserves decoding and canonicalization failures for inspection.
func (e *SourceIRError) Unwrap() error { return e.Cause }

// SyntaxStatus represents the parse state of a source file.
type SyntaxStatus string

const (
	SyntaxStatusParsed      SyntaxStatus = "PARSED"
	SyntaxStatusPartial     SyntaxStatus = "PARTIAL"
	SyntaxStatusUnsupported SyntaxStatus = "UNSUPPORTED"
	SyntaxStatusFailed      SyntaxStatus = "FAILED"
)

// SymbolKind represents the type of AST symbol discovered in source.
type SymbolKind string

const (
	SymbolKindClass     SymbolKind = "class"
	SymbolKindInterface SymbolKind = "interface"
	SymbolKindTrait     SymbolKind = "trait"
	SymbolKindFunction  SymbolKind = "function"
	SymbolKindMethod    SymbolKind = "method"
	SymbolKindProperty  SymbolKind = "property"
	SymbolKindConstant  SymbolKind = "constant"
	SymbolKindClosure   SymbolKind = "closure"
	SymbolKindScript    SymbolKind = "script"
)

// FindingSeverity represents the severity level of an IR finding.
type FindingSeverity string

const (
	FindingSeverityCritical FindingSeverity = "CRITICAL"
	FindingSeverityHigh     FindingSeverity = "HIGH"
	FindingSeverityMedium   FindingSeverity = "MEDIUM"
	FindingSeverityLow      FindingSeverity = "LOW"
)

// SourceFile describes a characterized source file with its content hash and syntax status.
type SourceFile struct {
	FileID       string       `json:"file_id"`
	Path         string       `json:"path"`
	SHA256       string       `json:"sha256"`
	SyntaxStatus SyntaxStatus `json:"syntax_status"`
}

// SourceSymbol describes an identified program construct and source span.
type SourceSymbol struct {
	SymbolID  string     `json:"symbol_id"`
	FileID    string     `json:"file_id"`
	Name      string     `json:"name"`
	Kind      SymbolKind `json:"kind"`
	StartLine int        `json:"start_line"`
	EndLine   int        `json:"end_line"`
}

// SourceEdge describes a static dependency or call edge between source entities.
type SourceEdge struct {
	EdgeID      string   `json:"edge_id"`
	FromID      string   `json:"from_id"`
	ToID        *string  `json:"to_id"`
	Kind        string   `json:"kind"`
	Resolved    bool     `json:"resolved"`
	EvidenceIDs []string `json:"evidence_ids"`
	Expression  *string  `json:"expression,omitempty"`
	AffectedIDs []string `json:"affected_ids,omitempty"`
}

// SourceFinding records a static analysis finding or blocker.
type SourceFinding struct {
	FindingID   string          `json:"finding_id"`
	Code        string          `json:"code"`
	Severity    FindingSeverity `json:"severity"`
	AffectedIDs []string        `json:"affected_ids"`
	EvidenceIDs []string        `json:"evidence_ids"`
	Message     string          `json:"message"`
}

// SourceEvidence binds source conclusions to immutable artifacts or traces.
type SourceEvidence struct {
	EvidenceID  string  `json:"evidence_id"`
	Kind        string  `json:"kind"`
	ArtifactRef string  `json:"artifact_ref"`
	SymbolID    *string `json:"symbol_id,omitempty"`
	TraceID     *string `json:"trace_id,omitempty"`
	Description string  `json:"description"`
}

// SourceIRDocument represents a complete Source IR document conforming to specs/ir/source-ir.schema.json.
type SourceIRDocument struct {
	SchemaVersion     string           `json:"schema_version"`
	ProjectID         string           `json:"project_id"`
	ConfigurationHash string           `json:"configuration_hash"`
	DocumentID        string           `json:"document_id"`
	Files             []SourceFile     `json:"files"`
	Symbols           []SourceSymbol   `json:"symbols"`
	Edges             []SourceEdge     `json:"edges"`
	Findings          []SourceFinding  `json:"findings"`
	Evidence          []SourceEvidence `json:"evidence"`
}

// CanonicalJSON renders the SourceIRDocument into canonical JSON bytes.
func (doc *SourceIRDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, &SourceIRError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical SourceIRDocument.
func (doc *SourceIRDocument) Digest() (string, error) {
	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ValidateSourceIR validates a raw JSON document against specs/ir/source-ir.schema.json
// and domain invariants.
func ValidateSourceIR(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &SourceIRError{Reason: "canonical JSON", Cause: err}
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &root); err != nil || root == nil {
		return &SourceIRError{Reason: "object required", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"configuration_hash",
		"document_id",
		"files",
		"symbols",
		"edges",
		"findings",
		"evidence",
	}

	if len(root) != len(expectedFields) {
		return &SourceIRError{Reason: "unexpected or missing root fields"}
	}
	for _, f := range expectedFields {
		v, ok := root[f]
		if !ok || string(v) == "null" {
			return &SourceIRError{Reason: "missing required field: " + f}
		}
	}

	var schemaVer string
	if err := json.Unmarshal(root["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &SourceIRError{Reason: "schema_version must be \"1.0\""}
	}

	var projectID string
	if err := json.Unmarshal(root["project_id"], &projectID); err != nil || !validAssessmentReportID(projectID) {
		return &SourceIRError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}

	var configHash string
	if err := json.Unmarshal(root["configuration_hash"], &configHash); err != nil || !isManifestDigest(configHash) {
		return &SourceIRError{Reason: "configuration_hash must match pattern ^[a-f0-9]{64}$"}
	}

	var docID string
	if err := json.Unmarshal(root["document_id"], &docID); err != nil || !validAssessmentReportID(docID) {
		return &SourceIRError{Reason: "document_id must match pattern ^[a-z0-9_-]+$"}
	}

	fileIDs, err := validateSourceFiles(root["files"])
	if err != nil {
		return err
	}

	symbolIDs, symbolFileRefs, err := validateSourceSymbols(root["symbols"], fileIDs)
	if err != nil {
		return err
	}
	_ = symbolFileRefs

	evidenceIDs, err := validateSourceEvidence(root["evidence"], symbolIDs)
	if err != nil {
		return err
	}

	validFromToIDs := make(map[string]struct{}, len(fileIDs)+len(symbolIDs))
	for fID := range fileIDs {
		validFromToIDs[fID] = struct{}{}
	}
	for sID := range symbolIDs {
		validFromToIDs[sID] = struct{}{}
	}

	if err := validateSourceEdges(root["edges"], validFromToIDs, evidenceIDs); err != nil {
		return err
	}

	if err := validateSourceFindings(root["findings"], evidenceIDs); err != nil {
		return err
	}

	return nil
}

func validateSourceFiles(raw json.RawMessage) (map[string]struct{}, error) {
	var files []json.RawMessage
	if err := json.Unmarshal(raw, &files); err != nil {
		return nil, &SourceIRError{Reason: "files must be an array", Cause: err}
	}

	seenIDs := make(map[string]struct{}, len(files))
	for _, rawFile := range files {
		var f map[string]json.RawMessage
		if err := json.Unmarshal(rawFile, &f); err != nil || f == nil {
			return nil, &SourceIRError{Reason: "file entry must be an object", Cause: err}
		}
		if len(f) != 4 {
			return nil, &SourceIRError{Reason: "unexpected or missing fields in file entry"}
		}
		for _, req := range []string{"file_id", "path", "sha256", "syntax_status"} {
			v, ok := f[req]
			if !ok || string(v) == "null" {
				return nil, &SourceIRError{Reason: "missing required file field: " + req}
			}
		}

		var fileID string
		if err := json.Unmarshal(f["file_id"], &fileID); err != nil || !validAssessmentReportID(fileID) {
			return nil, &SourceIRError{Reason: "file_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenIDs[fileID]; exists {
			return nil, &SourceIRError{Reason: fmt.Sprintf("duplicate file_id %q", fileID)}
		}
		seenIDs[fileID] = struct{}{}

		var path string
		if err := json.Unmarshal(f["path"], &path); err != nil || len(path) < 1 {
			return nil, &SourceIRError{Reason: "path must be a non-empty string"}
		}

		var sha string
		if err := json.Unmarshal(f["sha256"], &sha); err != nil || !isManifestDigest(sha) {
			return nil, &SourceIRError{Reason: "file sha256 must match pattern ^[a-f0-9]{64}$"}
		}

		var status string
		if err := json.Unmarshal(f["syntax_status"], &status); err != nil {
			return nil, &SourceIRError{Reason: "syntax_status must be a string", Cause: err}
		}
		if _, ok := allowedSyntaxStatuses[status]; !ok {
			return nil, &SourceIRError{Reason: fmt.Sprintf("invalid syntax_status %q", status)}
		}
	}

	return seenIDs, nil
}

func validateSourceSymbols(raw json.RawMessage, fileIDs map[string]struct{}) (map[string]struct{}, map[string]string, error) {
	var symbols []json.RawMessage
	if err := json.Unmarshal(raw, &symbols); err != nil {
		return nil, nil, &SourceIRError{Reason: "symbols must be an array", Cause: err}
	}

	seenIDs := make(map[string]struct{}, len(symbols))
	symbolFileMap := make(map[string]string, len(symbols))

	for _, rawSym := range symbols {
		var s map[string]json.RawMessage
		if err := json.Unmarshal(rawSym, &s); err != nil || s == nil {
			return nil, nil, &SourceIRError{Reason: "symbol entry must be an object", Cause: err}
		}
		if len(s) != 6 {
			return nil, nil, &SourceIRError{Reason: "unexpected or missing fields in symbol entry"}
		}
		for _, req := range []string{"symbol_id", "file_id", "name", "kind", "start_line", "end_line"} {
			v, ok := s[req]
			if !ok || string(v) == "null" {
				return nil, nil, &SourceIRError{Reason: "missing required symbol field: " + req}
			}
		}

		var symID string
		if err := json.Unmarshal(s["symbol_id"], &symID); err != nil || !validAssessmentReportID(symID) {
			return nil, nil, &SourceIRError{Reason: "symbol_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenIDs[symID]; exists {
			return nil, nil, &SourceIRError{Reason: fmt.Sprintf("duplicate symbol_id %q", symID)}
		}
		seenIDs[symID] = struct{}{}

		var fileID string
		if err := json.Unmarshal(s["file_id"], &fileID); err != nil || !validAssessmentReportID(fileID) {
			return nil, nil, &SourceIRError{Reason: "symbol file_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, ok := fileIDs[fileID]; !ok {
			return nil, nil, &SourceIRError{Reason: fmt.Sprintf("symbol file_id %q not found in files", fileID)}
		}
		symbolFileMap[symID] = fileID

		var name string
		if err := json.Unmarshal(s["name"], &name); err != nil || len(name) < 1 {
			return nil, nil, &SourceIRError{Reason: "symbol name must be a non-empty string"}
		}

		var kind string
		if err := json.Unmarshal(s["kind"], &kind); err != nil {
			return nil, nil, &SourceIRError{Reason: "symbol kind must be a string", Cause: err}
		}
		if _, ok := allowedSymbolKinds[kind]; !ok {
			return nil, nil, &SourceIRError{Reason: fmt.Sprintf("invalid symbol kind %q", kind)}
		}

		var startLine, endLine int
		if err := json.Unmarshal(s["start_line"], &startLine); err != nil || startLine < 1 {
			return nil, nil, &SourceIRError{Reason: "start_line must be an integer >= 1"}
		}
		if err := json.Unmarshal(s["end_line"], &endLine); err != nil || endLine < 1 {
			return nil, nil, &SourceIRError{Reason: "end_line must be an integer >= 1"}
		}
		if startLine > endLine {
			return nil, nil, &SourceIRError{Reason: fmt.Sprintf("start_line (%d) cannot be greater than end_line (%d)", startLine, endLine)}
		}
	}

	return seenIDs, symbolFileMap, nil
}

func validateSourceEdges(raw json.RawMessage, validIDs map[string]struct{}, evidenceIDs map[string]struct{}) error {
	var edges []json.RawMessage
	if err := json.Unmarshal(raw, &edges); err != nil {
		return &SourceIRError{Reason: "edges must be an array", Cause: err}
	}

	seenEdgeIDs := make(map[string]struct{}, len(edges))
	for _, rawEdge := range edges {
		var e map[string]json.RawMessage
		if err := json.Unmarshal(rawEdge, &e); err != nil || e == nil {
			return &SourceIRError{Reason: "edge entry must be an object", Cause: err}
		}

		for _, req := range []string{"edge_id", "from_id", "to_id", "kind", "resolved", "evidence_ids"} {
			v, ok := e[req]
			if !ok {
				return &SourceIRError{Reason: "missing required edge field: " + req}
			}
			if req != "to_id" && string(v) == "null" {
				return &SourceIRError{Reason: "edge field cannot be null: " + req}
			}
		}

		for k := range e {
			switch k {
			case "edge_id", "from_id", "to_id", "kind", "resolved", "evidence_ids", "expression", "affected_ids":
			default:
				return &SourceIRError{Reason: "unexpected field in edge entry: " + k}
			}
		}

		var edgeID string
		if err := json.Unmarshal(e["edge_id"], &edgeID); err != nil || !validAssessmentReportID(edgeID) {
			return &SourceIRError{Reason: "edge_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenEdgeIDs[edgeID]; exists {
			return &SourceIRError{Reason: fmt.Sprintf("duplicate edge_id %q", edgeID)}
		}
		seenEdgeIDs[edgeID] = struct{}{}

		var fromID string
		if err := json.Unmarshal(e["from_id"], &fromID); err != nil || !validAssessmentReportID(fromID) {
			return &SourceIRError{Reason: "edge from_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, ok := validIDs[fromID]; !ok {
			return &SourceIRError{Reason: fmt.Sprintf("edge from_id %q not found in files or symbols", fromID)}
		}

		if string(e["to_id"]) != "null" {
			var toID string
			if err := json.Unmarshal(e["to_id"], &toID); err != nil || !validAssessmentReportID(toID) {
				return &SourceIRError{Reason: "edge to_id must match pattern ^[a-z0-9_-]+$ or null"}
			}
			if _, ok := validIDs[toID]; !ok {
				return &SourceIRError{Reason: fmt.Sprintf("edge to_id %q not found in files or symbols", toID)}
			}
		}

		var kind string
		if err := json.Unmarshal(e["kind"], &kind); err != nil || len(kind) < 1 {
			return &SourceIRError{Reason: "edge kind must be a non-empty string"}
		}

		var resolved bool
		if err := json.Unmarshal(e["resolved"], &resolved); err != nil {
			return &SourceIRError{Reason: "edge resolved must be a boolean", Cause: err}
		}

		var evIDs []string
		if err := json.Unmarshal(e["evidence_ids"], &evIDs); err != nil {
			return &SourceIRError{Reason: "edge evidence_ids must be an array of strings", Cause: err}
		}
		for _, evID := range evIDs {
			if !validAssessmentReportID(evID) {
				return &SourceIRError{Reason: "edge evidence_ids item must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := evidenceIDs[evID]; !ok {
				return &SourceIRError{Reason: fmt.Sprintf("edge evidence_id %q not found in evidence", evID)}
			}
		}

		if exprRaw, ok := e["expression"]; ok && string(exprRaw) != "null" {
			var expr string
			if err := json.Unmarshal(exprRaw, &expr); err != nil || len(expr) < 1 {
				return &SourceIRError{Reason: "edge expression must be a non-empty string"}
			}
		}

		if affRaw, ok := e["affected_ids"]; ok && string(affRaw) != "null" {
			var affIDs []string
			if err := json.Unmarshal(affRaw, &affIDs); err != nil {
				return &SourceIRError{Reason: "edge affected_ids must be an array of strings", Cause: err}
			}
			for _, aID := range affIDs {
				if !validAssessmentReportID(aID) {
					return &SourceIRError{Reason: "edge affected_ids item must match pattern ^[a-z0-9_-]+$"}
				}
			}
		}
	}

	return nil
}

func validateSourceFindings(raw json.RawMessage, evidenceIDs map[string]struct{}) error {
	var findings []json.RawMessage
	if err := json.Unmarshal(raw, &findings); err != nil {
		return &SourceIRError{Reason: "findings must be an array", Cause: err}
	}

	seenFindingIDs := make(map[string]struct{}, len(findings))
	for _, rawF := range findings {
		var f map[string]json.RawMessage
		if err := json.Unmarshal(rawF, &f); err != nil || f == nil {
			return &SourceIRError{Reason: "finding entry must be an object", Cause: err}
		}
		if len(f) != 6 {
			return &SourceIRError{Reason: "unexpected or missing fields in finding entry"}
		}
		for _, req := range []string{"finding_id", "code", "severity", "affected_ids", "evidence_ids", "message"} {
			v, ok := f[req]
			if !ok || string(v) == "null" {
				return &SourceIRError{Reason: "missing required finding field: " + req}
			}
		}

		var findingID string
		if err := json.Unmarshal(f["finding_id"], &findingID); err != nil || !validAssessmentReportID(findingID) {
			return &SourceIRError{Reason: "finding_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenFindingIDs[findingID]; exists {
			return &SourceIRError{Reason: fmt.Sprintf("duplicate finding_id %q", findingID)}
		}
		seenFindingIDs[findingID] = struct{}{}

		var code string
		if err := json.Unmarshal(f["code"], &code); err != nil || len(code) < 1 {
			return &SourceIRError{Reason: "finding code must be a non-empty string"}
		}

		var sev string
		if err := json.Unmarshal(f["severity"], &sev); err != nil {
			return &SourceIRError{Reason: "finding severity must be a string", Cause: err}
		}
		if _, ok := allowedFindingSeverities[sev]; !ok {
			return &SourceIRError{Reason: fmt.Sprintf("invalid finding severity %q", sev)}
		}

		var affIDs []string
		if err := json.Unmarshal(f["affected_ids"], &affIDs); err != nil {
			return &SourceIRError{Reason: "finding affected_ids must be an array of strings", Cause: err}
		}
		for _, aID := range affIDs {
			if !validAssessmentReportID(aID) {
				return &SourceIRError{Reason: "finding affected_ids item must match pattern ^[a-z0-9_-]+$"}
			}
		}

		var evIDs []string
		if err := json.Unmarshal(f["evidence_ids"], &evIDs); err != nil {
			return &SourceIRError{Reason: "finding evidence_ids must be an array of strings", Cause: err}
		}
		for _, evID := range evIDs {
			if !validAssessmentReportID(evID) {
				return &SourceIRError{Reason: "finding evidence_ids item must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := evidenceIDs[evID]; !ok {
				return &SourceIRError{Reason: fmt.Sprintf("finding evidence_id %q not found in evidence", evID)}
			}
		}

		var msg string
		if err := json.Unmarshal(f["message"], &msg); err != nil || len(msg) < 1 {
			return &SourceIRError{Reason: "finding message must be a non-empty string"}
		}
	}

	return nil
}

func validateSourceEvidence(raw json.RawMessage, symbolIDs map[string]struct{}) (map[string]struct{}, error) {
	var evidenceList []json.RawMessage
	if err := json.Unmarshal(raw, &evidenceList); err != nil {
		return nil, &SourceIRError{Reason: "evidence must be an array", Cause: err}
	}

	seenEvidenceIDs := make(map[string]struct{}, len(evidenceList))
	for _, rawEv := range evidenceList {
		var ev map[string]json.RawMessage
		if err := json.Unmarshal(rawEv, &ev); err != nil || ev == nil {
			return nil, &SourceIRError{Reason: "evidence entry must be an object", Cause: err}
		}

		for _, req := range []string{"evidence_id", "kind", "artifact_ref", "description"} {
			v, ok := ev[req]
			if !ok || string(v) == "null" {
				return nil, &SourceIRError{Reason: "missing required evidence field: " + req}
			}
		}

		for k := range ev {
			switch k {
			case "evidence_id", "kind", "artifact_ref", "description", "symbol_id", "trace_id":
			default:
				return nil, &SourceIRError{Reason: "unexpected field in evidence entry: " + k}
			}
		}

		var evID string
		if err := json.Unmarshal(ev["evidence_id"], &evID); err != nil || !validAssessmentReportID(evID) {
			return nil, &SourceIRError{Reason: "evidence_id must match pattern ^[a-z0-9_-]+$"}
		}
		if _, exists := seenEvidenceIDs[evID]; exists {
			return nil, &SourceIRError{Reason: fmt.Sprintf("duplicate evidence_id %q", evID)}
		}
		seenEvidenceIDs[evID] = struct{}{}

		var kind string
		if err := json.Unmarshal(ev["kind"], &kind); err != nil {
			return nil, &SourceIRError{Reason: "evidence kind must be a string", Cause: err}
		}
		if _, ok := allowedEvidenceKinds[kind]; !ok {
			return nil, &SourceIRError{Reason: fmt.Sprintf("invalid evidence kind %q", kind)}
		}

		if !validProfileRef(ev["artifact_ref"], false) {
			return nil, &SourceIRError{Reason: "evidence artifact_ref must match pattern ^sha256:[a-f0-9]{64}$"}
		}

		var desc string
		if err := json.Unmarshal(ev["description"], &desc); err != nil || len(desc) < 1 {
			return nil, &SourceIRError{Reason: "evidence description must be a non-empty string"}
		}

		if symRaw, ok := ev["symbol_id"]; ok && string(symRaw) != "null" {
			var symID string
			if err := json.Unmarshal(symRaw, &symID); err != nil || !validAssessmentReportID(symID) {
				return nil, &SourceIRError{Reason: "evidence symbol_id must match pattern ^[a-z0-9_-]+$"}
			}
			if _, ok := symbolIDs[symID]; !ok {
				return nil, &SourceIRError{Reason: fmt.Sprintf("evidence symbol_id %q not found in symbols", symID)}
			}
		}

		if trRaw, ok := ev["trace_id"]; ok && string(trRaw) != "null" {
			var trID string
			if err := json.Unmarshal(trRaw, &trID); err != nil || !validAssessmentReportID(trID) {
				return nil, &SourceIRError{Reason: "evidence trace_id must match pattern ^[a-z0-9_-]+$"}
			}
		}
	}

	return seenEvidenceIDs, nil
}

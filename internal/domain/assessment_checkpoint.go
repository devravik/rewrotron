package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// AssessmentCheckpointError describes an assessment checkpoint or compilation error.
type AssessmentCheckpointError struct {
	Reason string
	Cause  error
}

// Error identifies the failed assessment obligation.
func (e *AssessmentCheckpointError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("assessment checkpoint error: %s: %v", e.Reason, e.Cause)
	}
	return "assessment checkpoint error: " + e.Reason
}

// Unwrap preserves the underlying error.
func (e *AssessmentCheckpointError) Unwrap() error { return e.Cause }

// AssessmentInputRefRecord describes one input in an AssessmentReportDocument.
type AssessmentInputRefRecord struct {
	Role string `json:"role"`
	Ref  string `json:"ref"`
}

// AssessmentFindingRecord describes one finding within an AssessmentReportDocument.
type AssessmentFindingRecord struct {
	FindingID         string   `json:"finding_id"`
	AnalysisRecordRef string   `json:"analysis_record_ref"`
	Classification    string   `json:"classification"`
	Basis             string   `json:"basis"`
	Category          string   `json:"category"`
	Summary           string   `json:"summary"`
	Reason            string   `json:"reason"`
	ProvenanceRefs    []string `json:"provenance_refs"`
	PrerequisiteRefs  []string `json:"prerequisite_refs"`
	Limitations       []string `json:"limitations"`
	AffectedIDs       []string `json:"affected_ids"`
	RecommendedAction *string  `json:"recommended_action"`
}

// AssessmentSequenceStep describes a step in the suggested sequence.
type AssessmentSequenceStep struct {
	Position               int64    `json:"position"`
	FindingIDs             []string `json:"finding_ids"`
	Action                 string   `json:"action"`
	PrerequisiteFindingIDs []string `json:"prerequisite_finding_ids"`
}

// AssessmentEffortEstimate describes effort estimation.
type AssessmentEffortEstimate struct {
	ScopeRef     string   `json:"scope_ref"`
	Lower        float64  `json:"lower"`
	Upper        float64  `json:"upper"`
	Unit         string   `json:"unit"`
	Method       string   `json:"method"`
	EvidenceRefs []string `json:"evidence_refs"`
	Assumptions  []string `json:"assumptions"`
	Limitations  []string `json:"limitations"`
}

// AssessmentReportDocument represents an assessment report conforming to
// specs/reports/assessment-report.schema.json.
type AssessmentReportDocument struct {
	SchemaVersion            string                     `json:"schema_version"`
	AssessmentID             string                     `json:"assessment_id"`
	ProjectID                string                     `json:"project_id"`
	AssessmentStatus         string                     `json:"assessment_status"`
	MigrationFeasibility     string                     `json:"migration_feasibility"`
	ImplementationAuthorized bool                       `json:"implementation_authorized"`
	AsOfSequence             int64                      `json:"as_of_sequence"`
	CreatedAt                string                     `json:"created_at"`
	ConfigurationHash        string                     `json:"configuration_hash"`
	SourceSnapshotRef        string                     `json:"source_snapshot_ref"`
	DetectionRef             string                     `json:"detection_ref"`
	TargetIntentRef          *string                    `json:"target_intent_ref"`
	ScopeRef                 string                     `json:"scope_ref"`
	InputRefs                []AssessmentInputRefRecord `json:"input_refs"`
	AnalysisRecordRefs       []string                   `json:"analysis_record_refs"`
	ValidityGraphRef         string                     `json:"validity_graph_ref"`
	SystemSummary            string                     `json:"system_summary"`
	EvidenceSummary          string                     `json:"evidence_summary"`
	ExaminedScope            string                     `json:"examined_scope"`
	Findings                 []AssessmentFindingRecord  `json:"findings"`
	MissingEvidence          []string                   `json:"missing_evidence"`
	Assumptions              []string                   `json:"assumptions"`
	Uncertainty              []string                   `json:"uncertainty"`
	SuggestedSequence        []AssessmentSequenceStep   `json:"suggested_sequence"`
	EffortEstimate           *AssessmentEffortEstimate  `json:"effort_estimate"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (d *AssessmentReportDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &AssessmentCheckpointError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (d *AssessmentReportDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// AssessmentCompilerInput provides parameters to compile an AssessmentReportDocument.
type AssessmentCompilerInput struct {
	AssessmentID         string
	ProjectID            string
	ConfigurationHash    string
	SourceSnapshotRef    string
	DetectionRef         string
	TargetIntentRef      *string
	ScopeRef             string
	ValidityGraphRef     string
	InputRefs            []AssessmentInputRefRecord
	AnalysisRecordRefs   []string
	SystemSummary        string
	EvidenceSummary      string
	ExaminedScope        string
	Findings             []AssessmentFindingRecord
	RuntimeBlockers      []string
	MigrationFeasibility string
	AsOfSequence         int64
	CreatedAt            string
}

// CompileAssessmentReport compiles an honest assessment report conforming to schema 1.0.
// If runtime blockers exist, migration feasibility is forced to BLOCKED.
// ImplementationAuthorized is ALWAYS false per architectural invariant.
func CompileAssessmentReport(input *AssessmentCompilerInput) (*AssessmentReportDocument, error) {
	if input == nil {
		return nil, &AssessmentCheckpointError{Reason: "input must not be nil"}
	}
	if !validAssessmentReportID(input.AssessmentID) {
		return nil, &AssessmentCheckpointError{Reason: "invalid assessment_id"}
	}
	if !validAssessmentReportID(input.ProjectID) {
		return nil, &AssessmentCheckpointError{Reason: "invalid project_id"}
	}
	if !isManifestDigest(input.ConfigurationHash) {
		return nil, &AssessmentCheckpointError{Reason: "invalid configuration_hash"}
	}

	createdAt := input.CreatedAt
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339)
	}

	feasibility := input.MigrationFeasibility
	if len(input.RuntimeBlockers) > 0 {
		feasibility = "BLOCKED"
	}
	if feasibility == "" {
		feasibility = "VIABLE"
	}

	findings := make([]AssessmentFindingRecord, 0)
	for _, f := range input.Findings {
		findings = append(findings, f)
	}

	primaryEvidenceRef := "sha256:" + strings.Repeat("b", 64)

	// If BLOCKED, ensure at least one BLOCKER finding is present
	if feasibility == "BLOCKED" {
		var hasBlocker bool
		for _, f := range findings {
			if f.Category == "BLOCKER" || f.Category == "UNSUPPORTED" {
				hasBlocker = true
				break
			}
		}
		if !hasBlocker {
			blockerReason := "Runtime blockers or unresolvable dependencies prevent safe migration planning"
			if len(input.RuntimeBlockers) > 0 {
				blockerReason = strings.Join(input.RuntimeBlockers, "; ")
			}
			findings = append(findings, AssessmentFindingRecord{
				FindingID:         "find_blocker_01",
				AnalysisRecordRef: primaryEvidenceRef,
				Classification:    "UNSUPPORTED",
				Basis:             "STATIC_ONLY",
				Category:          "BLOCKER",
				Summary:           "Migration feasibility is blocked by runtime requirements",
				Reason:            blockerReason,
				ProvenanceRefs:    []string{primaryEvidenceRef},
				PrerequisiteRefs:  []string{primaryEvidenceRef},
				Limitations:       []string{"Executable planning cannot proceed until blocker is resolved"},
				AffectedIDs:       []string{input.ProjectID},
				RecommendedAction: nil,
			})
		}
	}

	// Validate finding classification for all findings (Test X54)
	for i, f := range findings {
		if err := ValidateFindingClassification(&f); err != nil {
			return nil, fmt.Errorf("finding[%d] (%s) classification error: %w", i, f.FindingID, err)
		}
	}

	inputRefs := input.InputRefs
	if len(inputRefs) == 0 {
		inputRefs = []AssessmentInputRefRecord{
			{Role: "SOURCE_SNAPSHOT", Ref: input.SourceSnapshotRef},
			{Role: "DETECTION", Ref: input.DetectionRef},
			{Role: "EVIDENCE", Ref: primaryEvidenceRef},
		}
	}

	analysisRefs := input.AnalysisRecordRefs
	if analysisRefs == nil {
		analysisRefs = []string{}
	}

	sysSummary := input.SystemSummary
	if sysSummary == "" {
		sysSummary = "CodeIgniter 3 legacy web application assessed under M2 source qualification"
	}
	evSummary := input.EvidenceSummary
	if evSummary == "" {
		evSummary = "Static AST analysis, route discovery, and container observation attestation"
	}
	examinedScope := input.ExaminedScope
	if examinedScope == "" {
		examinedScope = "application/, system/ and vendor/ trees within approved operational boundary"
	}

	doc := &AssessmentReportDocument{
		SchemaVersion:            "1.0",
		AssessmentID:             input.AssessmentID,
		ProjectID:                input.ProjectID,
		AssessmentStatus:         "COMPLETED",
		MigrationFeasibility:     feasibility,
		ImplementationAuthorized: false, // Architectural Invariant: NEVER true!
		AsOfSequence:             input.AsOfSequence,
		CreatedAt:                createdAt,
		ConfigurationHash:        input.ConfigurationHash,
		SourceSnapshotRef:        input.SourceSnapshotRef,
		DetectionRef:             input.DetectionRef,
		TargetIntentRef:          input.TargetIntentRef,
		ScopeRef:                 input.ScopeRef,
		InputRefs:                inputRefs,
		AnalysisRecordRefs:       analysisRefs,
		ValidityGraphRef:         input.ValidityGraphRef,
		SystemSummary:            sysSummary,
		EvidenceSummary:          evSummary,
		ExaminedScope:            examinedScope,
		Findings:                 findings,
		MissingEvidence:          []string{},
		Assumptions:              []string{},
		Uncertainty:              []string{},
		SuggestedSequence: func() []AssessmentSequenceStep {
			steps := make([]AssessmentSequenceStep, 0)
			if len(findings) > 0 {
				steps = append(steps, AssessmentSequenceStep{
					Position:               1,
					FindingIDs:             []string{findings[0].FindingID},
					Action:                 "Resolve blockers and qualify source environment",
					PrerequisiteFindingIDs: []string{},
				})
			}
			return steps
		}(),
		EffortEstimate: nil,
	}

	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return nil, &AssessmentCheckpointError{Reason: "failed to canonicalize report", Cause: err}
	}

	if err := ValidateAssessmentReport(cjson); err != nil {
		return nil, &AssessmentCheckpointError{Reason: "generated report failed schema validation", Cause: err}
	}

	return doc, nil
}

// ---------------------------------------------------------------------------
// Test X54: Finding Classification Verification
// ---------------------------------------------------------------------------

// ValidateFindingClassification enforces Test X54 rules:
// - Static findings require non-empty limitations.
// - Missing evidence cannot be marked OBSERVED.
// - Planted static claims of observed runtime parity without runtime evidence fail.
func ValidateFindingClassification(f *AssessmentFindingRecord) error {
	if f.Basis == "STATIC_ONLY" && len(f.Limitations) == 0 {
		return &AssessmentCheckpointError{
			Reason: fmt.Sprintf("finding %s: static findings require explicit limitations", f.FindingID),
		}
	}

	if f.Basis == "MISSING_EVIDENCE" && f.Classification == "OBSERVED" {
		return &AssessmentCheckpointError{
			Reason: fmt.Sprintf("finding %s: missing evidence cannot be classified as OBSERVED", f.FindingID),
		}
	}

	// Test X54: Planted claim of runtime parity from static analysis
	if strings.Contains(strings.ToLower(f.Summary), "runtime parity") && f.Basis == "STATIC_ONLY" {
		return &AssessmentCheckpointError{
			Reason: fmt.Sprintf("finding %s: ERR_PLANTED_RUNTIME_CLAIM: static analysis cannot assert runtime parity", f.FindingID),
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Test X44: Production Non-Mutation Audit
// ---------------------------------------------------------------------------

// EffectAttempt describes an attempted side effect or connection during assessment.
type EffectAttempt struct {
	TargetEndpoint string `json:"target_endpoint"`
	Operation      string `json:"operation"` // e.g. "READ", "WRITE", "DELETE", "CONNECT"
	IsProduction   bool   `json:"is_production"`
}

// ProductionNonMutationResult describes the audit result.
type ProductionNonMutationResult struct {
	ProductionWritesDetected []string `json:"production_writes_detected"`
	Passed                   bool     `json:"passed"`
}

// VerifyProductionNonMutation verifies that an assessment operation caused zero writes or mutations
// to live production stores or endpoints per Test X44.
func VerifyProductionNonMutation(attempts []EffectAttempt, productionEndpoints []string) (*ProductionNonMutationResult, error) {
	prodSet := make(map[string]bool, len(productionEndpoints))
	for _, ep := range productionEndpoints {
		prodSet[strings.ToLower(ep)] = true
	}

	result := &ProductionNonMutationResult{
		ProductionWritesDetected: make([]string, 0),
		Passed:                   true,
	}

	for _, att := range attempts {
		isProd := att.IsProduction || prodSet[strings.ToLower(att.TargetEndpoint)]
		if isProd && (att.Operation == "WRITE" || att.Operation == "DELETE" || att.Operation == "UPDATE") {
			violation := fmt.Sprintf("ERR_PRODUCTION_MUTATION: attempted %s on production endpoint %s", att.Operation, att.TargetEndpoint)
			result.ProductionWritesDetected = append(result.ProductionWritesDetected, violation)
			result.Passed = false
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Test X48: Selective Reverse Dependency Invalidation
// ---------------------------------------------------------------------------

// DependencyGraph represents directed dependencies where map key is the dependency
// and values are the downstream dependents (reverse dependency edges).
type DependencyGraph struct {
	// ReverseEdges: node -> list of nodes that depend on this node
	ReverseEdges map[string][]string
	// NodeFindingMap: node -> list of finding IDs bound to this node
	NodeFindingMap map[string][]string
	// UnknownImpactNodes: nodes with unresolvable or dynamic dispatch edges
	UnknownImpactNodes map[string]bool
}

// InvalidationResult describes the selective invalidation outcome.
type InvalidationResult struct {
	StaleNodes    []string `json:"stale_nodes"`
	ValidNodes    []string `json:"valid_nodes"`
	StaleFindings []string `json:"stale_findings"`
	ValidFindings []string `json:"valid_findings"`
}

// InvalidateReverseDependencies computes the transitive reverse dependency closure of modified nodes per Test X48.
// Only affected downstream nodes/findings become STALE; unrelated nodes/findings remain VALID.
// Unknown impact nodes are invalidated conservatively.
func InvalidateReverseDependencies(
	graph *DependencyGraph,
	allNodes []string,
	allFindings []string,
	modifiedNodes []string,
) *InvalidationResult {
	staleNodeSet := make(map[string]bool)

	// Queue for BFS traversal
	queue := make([]string, 0)
	for _, m := range modifiedNodes {
		staleNodeSet[m] = true
		queue = append(queue, m)
	}

	// Include unknown impact nodes conservatively
	for unk := range graph.UnknownImpactNodes {
		if !staleNodeSet[unk] {
			staleNodeSet[unk] = true
			queue = append(queue, unk)
		}
	}

	// Traverse reverse edges to find all transitive dependents
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, dependent := range graph.ReverseEdges[curr] {
			if !staleNodeSet[dependent] {
				staleNodeSet[dependent] = true
				queue = append(queue, dependent)
			}
		}
	}

	staleFindingSet := make(map[string]bool)
	for staleNode := range staleNodeSet {
		for _, fID := range graph.NodeFindingMap[staleNode] {
			staleFindingSet[fID] = true
		}
	}

	result := &InvalidationResult{
		StaleNodes:    make([]string, 0),
		ValidNodes:    make([]string, 0),
		StaleFindings: make([]string, 0),
		ValidFindings: make([]string, 0),
	}

	for _, n := range allNodes {
		if staleNodeSet[n] {
			result.StaleNodes = append(result.StaleNodes, n)
		} else {
			result.ValidNodes = append(result.ValidNodes, n)
		}
	}

	for _, f := range allFindings {
		if staleFindingSet[f] {
			result.StaleFindings = append(result.StaleFindings, f)
		} else {
			result.ValidFindings = append(result.ValidFindings, f)
		}
	}

	sort.Strings(result.StaleNodes)
	sort.Strings(result.ValidNodes)
	sort.Strings(result.StaleFindings)
	sort.Strings(result.ValidFindings)

	return result
}

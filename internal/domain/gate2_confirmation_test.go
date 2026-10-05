package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPlanEnvelope_GoldenVectors(t *testing.T) {
	// 1. Read validation-cases.json to load checkpoint-linked-executable-plan-envelope
	rawCases, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation-cases.json: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(rawCases, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	var validEnvelopeJSON []byte
	var invalidEnvelopeJSON []byte
	for _, c := range cases {
		if c.Name == "checkpoint-linked-executable-plan-envelope" {
			validEnvelopeJSON = c.Instance
		}
		if c.Name == "current-plan-cannot-omit-assessment-selection" {
			invalidEnvelopeJSON = c.Instance
		}
	}

	if len(validEnvelopeJSON) == 0 {
		t.Fatalf("checkpoint-linked-executable-plan-envelope case not found")
	}
	if len(invalidEnvelopeJSON) == 0 {
		t.Fatalf("current-plan-cannot-omit-assessment-selection case not found")
	}

	// 2. Validate valid envelope
	env, err := ParsePlanEnvelope(validEnvelopeJSON)
	if err != nil {
		t.Fatalf("ParsePlanEnvelope on valid envelope failed: %v", err)
	}
	if env.PlanID != "plan" || env.ProjectID != "fixture" {
		t.Fatalf("unexpected plan envelope IDs: plan_id=%s, project_id=%s", env.PlanID, env.ProjectID)
	}

	// Verify plan_hash matches subject digest
	subjDigest, err := env.Subject.Digest()
	if err != nil {
		t.Fatalf("failed to compute subject digest: %v", err)
	}
	expectedHash := "2fea0aa910c9d732a7d087b3065ed9f477c5afe9d524b95a1e93bc6dda2c2e18"
	if subjDigest != expectedHash {
		t.Fatalf("subject digest %s does not match expected golden vector %s", subjDigest, expectedHash)
	}
	if env.PlanHash != expectedHash {
		t.Fatalf("envelope plan_hash %s does not match expected golden vector %s", env.PlanHash, expectedHash)
	}

	// 3. Re-canonicalize envelope and check round-trip
	envCanon, err := env.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize envelope: %v", err)
	}
	envDigest, err := env.Digest()
	if err != nil {
		t.Fatalf("failed to compute envelope digest: %v", err)
	}
	if !strings.HasPrefix(envDigest, "sha256:") {
		t.Fatalf("envelope digest must start with sha256:, got %s", envDigest)
	}
	if err := ValidatePlanEnvelope(envCanon); err != nil {
		t.Fatalf("re-validation of canonical envelope failed: %v", err)
	}

	// 4. Validate invalid envelope (omits selection_ref)
	_, err = ParsePlanEnvelope(invalidEnvelopeJSON)
	if err == nil {
		t.Fatalf("expected ParsePlanEnvelope to fail for current-plan-cannot-omit-assessment-selection, but it passed")
	}
}

// TestGate2_ConfirmationProtocol_X55 validates Test X55:
// "Submit report ref/hash or human continuation selection to confirm-plan/generate;
// reject kind/subject/token mismatch. Plan wrapper with mismatched selection/basis or omitted hashed refs also fails."
func TestGate2_ConfirmationProtocol_X55(t *testing.T) {
	// A. Gate check: must require gate == "plan"
	validSubjectJSON := []byte(`{
		"schema_version": "2.0",
		"canonicalization": "rwt-cjson-1",
		"project_id": "fixture",
		"kind": "PLAN",
		"configuration_hash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"snapshot_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"behavior_ir_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"corpus_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"policy_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"target_intent_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"resolved_environment_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"target_model_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"decision_refs": [
			"sha256:0ab26ee1f8b0f264905fd9b11c832060d4d39202e517ffe098f2d6346412ed21",
			"sha256:a17d92d37e6785b1487096bcd4c8eafdb915ed5bda24ceafe3a5b98da0220d8e"
		],
		"dependency_graph_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"adapter_bindings": [
			{
				"role": "TARGET",
				"adapter_id": "synthetic",
				"version": "1.0.0",
				"manifest_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"implementation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			}
		],
		"consumed_revision_refs": [],
		"ordered_units": [
			{
				"unit_id": "login",
				"contract_refs": ["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],
				"case_ids": ["login"],
				"priority": 1,
				"capabilities": ["http"]
			}
		],
		"dependency_edges": [],
		"omission_proof_refs": [],
		"scope_exclusion_refs": [],
		"unresolved_items": []
	}`)

	_, err := ValidateGate2Token("generate", validSubjectJSON)
	if err == nil || !strings.Contains(err.Error(), "Gate 2 requires gate 'plan'") {
		t.Fatalf("expected error for wrong gate name, got %v", err)
	}

	// B. Submit assessment report JSON (kind: assessment) as Gate 2 plan token
	assessmentReportJSON := []byte(`{
		"schema_version": "1.0",
		"project_id": "fixture",
		"kind": "assessment",
		"report_id": "rep1"
	}`)
	_, err = ValidateGate2Token("plan", assessmentReportJSON)
	if err == nil || !strings.Contains(err.Error(), "assessment report cannot be used as Gate 2 plan approval token") {
		t.Fatalf("expected rejection of assessment report as Gate 2 token, got %v", err)
	}

	// C. Submit human continuation selection as Gate 2 plan token
	selectionJSON := []byte(`{
		"schema_version": "1.0",
		"selection_id": "sel1",
		"project_id": "fixture",
		"selected_scope_ref": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	}`)
	_, err = ValidateGate2Token("plan", selectionJSON)
	if err == nil || !strings.Contains(err.Error(), "assessment continuation selection cannot be used as Gate 2 plan approval token") {
		t.Fatalf("expected rejection of continuation selection as Gate 2 token, got %v", err)
	}

	// D. Submit report ref or hash directly
	reportRefJSON := []byte(`{
		"report_id": "sha256:0f20635e1861db9ebbf7f466e50511be0c230ec561cf53d021cf581cf55a07e2"
	}`)
	_, err = ValidateGate2Token("plan", reportRefJSON)
	if err == nil || !strings.Contains(err.Error(), "assessment report artifact cannot be used as Gate 2 plan approval token") {
		t.Fatalf("expected rejection of report ref as Gate 2 token, got %v", err)
	}

	// E. Valid plan subject accepted
	subj, err := ValidateGate2Token("plan", validSubjectJSON)
	if err != nil {
		t.Fatalf("valid plan subject rejected by Gate 2: %v", err)
	}
	if subj.Kind != "PLAN" {
		t.Fatalf("expected kind PLAN, got %s", subj.Kind)
	}

	// F. Validate Plan Envelope Checkpoint Closure:
	// Mismatched selection/basis or omitted hashed refs
	basis := &AssessmentPlanningBasisRecord{
		SchemaVersion:          "1.0",
		ProjectID:              "fixture",
		AssessmentRef:          "sha256:0f20635e1861db9ebbf7f466e50511be0c230ec561cf53d021cf581cf55a07e2",
		SelectionRef:           "sha256:0ab26ee1f8b0f264905fd9b11c832060d4d39202e517ffe098f2d6346412ed21",
		SelectedScopeRef:       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CurrentInputRefs: []CurrentInputRef{
			{Role: "SOURCE_SNAPSHOT", Ref: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		},
		ValidityGraphRef:       "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ReusedFindings:         []ReusedFinding{},
		StaleFindingRefs:       []string{},
		NewFindingRefs:         []string{},
		ChangesSinceAssessment: []ChangeSinceAssessment{},
	}
	basisDigest, err := basis.Digest()
	if err != nil {
		t.Fatalf("failed to compute basis digest: %v", err)
	}

	subjCopy := *subj
	subjCopy.DecisionRefs = []string{basis.SelectionRef, basisDigest}

	envelope := &PlanEnvelopeRecord{
		SchemaVersion:    "2.1",
		PlanID:           "plan",
		ProjectID:        "fixture",
		PlanHash:         "2fea0aa910c9d732a7d087b3065ed9f477c5afe9d524b95a1e93bc6dda2c2e18",
		Subject:          &subjCopy,
		AssessmentRef:    basis.AssessmentRef,
		SelectionRef:     basis.SelectionRef,
		PlanningBasisRef: basisDigest,
	}

	// Matching basis closure should pass
	if err := ValidatePlanEnvelopeCheckpointClosure(envelope, basis); err != nil {
		t.Fatalf("expected valid envelope checkpoint closure to pass, got %v", err)
	}

	// Mismatched assessment_ref
	mismatchedEnv := *envelope
	mismatchedEnv.AssessmentRef = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if err := ValidatePlanEnvelopeCheckpointClosure(&mismatchedEnv, basis); err == nil || !strings.Contains(err.Error(), "does not match planning basis assessment_ref") {
		t.Fatalf("expected error for mismatched assessment_ref, got %v", err)
	}

	// Mismatched selection_ref
	mismatchedEnv2 := *envelope
	mismatchedEnv2.SelectionRef = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if err := ValidatePlanEnvelopeCheckpointClosure(&mismatchedEnv2, basis); err == nil || !strings.Contains(err.Error(), "does not match planning basis selection_ref") {
		t.Fatalf("expected error for mismatched selection_ref, got %v", err)
	}

	// Mismatched planning_basis_ref
	mismatchedEnv3 := *envelope
	mismatchedEnv3.PlanningBasisRef = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	if err := ValidatePlanEnvelopeCheckpointClosure(&mismatchedEnv3, basis); err == nil || !strings.Contains(err.Error(), "does not match computed planning basis digest") {
		t.Fatalf("expected error for mismatched planning_basis_ref, got %v", err)
	}
}

// TestGate2_ExecutableQualificationClosure_X58 validates Test X58:
// "Swap core/schema bytes under same version/URI or widen advertised profile without controls;
// current approval must refuse execution; retained old subject/vector is unchanged and reconstructable, planned manifest cannot qualify."
func TestGate2_ExecutableQualificationClosure_X58(t *testing.T) {
	subject := &PlanApprovalSubject{
		SchemaVersion:          "2.0",
		Canonicalization:       "rwt-cjson-1",
		ProjectID:              "fixture",
		Kind:                   "PLAN",
		ConfigurationHash:      "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		SnapshotRef:            "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BehaviorIRRef:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CorpusRef:              "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PolicyRef:              "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TargetIntentRef:        "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ResolvedEnvironmentRef: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TargetModelRef:         "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		DecisionRefs: []string{
			"sha256:0ab26ee1f8b0f264905fd9b11c832060d4d39202e517ffe098f2d6346412ed21",
			"sha256:a17d92d37e6785b1487096bcd4c8eafdb915ed5bda24ceafe3a5b98da0220d8e",
		},
		DependencyGraphRef: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		AdapterBindings: []PlanAdapterBinding{
			{
				Role:              "TARGET",
				AdapterID:         "synthetic",
				Version:           "1.0.0",
				ManifestRef:       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ImplementationRef: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
		ConsumedRevisionRefs: []string{},
		OrderedUnits: []PlanApprovalUnit{
			{
				UnitID:       "login",
				ContractRefs: []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
				CaseIDs:      []string{"login"},
				Priority:     1,
				Capabilities: []string{"http"},
			},
		},
		DependencyEdges:    []PlanApprovalEdge{},
		OmissionProofRefs:  []string{},
		ScopeExclusionRefs: []string{},
		UnresolvedItems:    []PlanApprovalUnresolvedItem{},
	}

	validQual := &QualificationRecord{
		SchemaVersion:     "1.0",
		QualificationID:   "qual1",
		ProjectID:         "fixture",
		SubjectKind:       "ADAPTER",
		ImplementationRef: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SchemaRefs:        []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		CapabilityIDs:     []string{"target_codegen"},
		InputProfileRef:   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ControlSuiteRef:   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ControlResultRefs: []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		Limitations:       []string{},
		Status:            "SUPPORTED",
		ReasonCodes:       []string{},
	}

	admitted := map[string]*QualificationRecord{
		"synthetic:1.0.0": validQual,
	}

	// 1. Valid qualification closure passes
	if err := VerifyExecutableQualificationClosure(subject, admitted); err != nil {
		t.Fatalf("expected valid qualification closure to pass, got %v", err)
	}

	// 2. Swapped implementation bytes under same version/URI
	swappedQual := *validQual
	swappedQual.ImplementationRef = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	admittedSwapped := map[string]*QualificationRecord{
		"synthetic:1.0.0": &swappedQual,
	}
	err := VerifyExecutableQualificationClosure(subject, admittedSwapped)
	if err == nil || !strings.Contains(err.Error(), "swapped bytes under same version/URI rejected") {
		t.Fatalf("expected rejection when implementation bytes swapped, got %v", err)
	}

	// 3. Widened profile or missing controls without test evidence
	uncontrolledQual := *validQual
	uncontrolledQual.ControlResultRefs = []string{} // empty controls
	admittedUncontrolled := map[string]*QualificationRecord{
		"synthetic:1.0.0": &uncontrolledQual,
	}
	err = VerifyExecutableQualificationClosure(subject, admittedUncontrolled)
	if err == nil || !strings.Contains(err.Error(), "missing required control result references") {
		t.Fatalf("expected rejection when control results are missing, got %v", err)
	}

	// 4. UNSUPPORTED status rejected
	unsupportedQual := *validQual
	unsupportedQual.Status = "UNSUPPORTED"
	unsupportedQual.ReasonCodes = []string{"FEATURE_NOT_IMPLEMENTED"}
	admittedUnsupported := map[string]*QualificationRecord{
		"synthetic:1.0.0": &unsupportedQual,
	}
	err = VerifyExecutableQualificationClosure(subject, admittedUnsupported)
	if err == nil || !strings.Contains(err.Error(), "qualification status is UNSUPPORTED") {
		t.Fatalf("expected rejection for UNSUPPORTED status, got %v", err)
	}

	// 5. Retained old subject digest remains identical and reconstructable
	digest, err := subject.Digest()
	if err != nil {
		t.Fatalf("failed to compute subject digest: %v", err)
	}
	expectedHash := "2fea0aa910c9d732a7d087b3065ed9f477c5afe9d524b95a1e93bc6dda2c2e18"
	if digest != expectedHash {
		t.Fatalf("retained subject digest %s changed from expected %s", digest, expectedHash)
	}
}

// TestGate2_ApprovedPerformanceEstimand_X59 validates Test X59:
// "Change policy/workload/method after Gate 2, omit required counter, conflate pooled percentile
// with mean of trial quantiles, fabricate CI or error omission; service must reject/inconclusive,
// equal data within approved positive margin can accept."
func TestGate2_ApprovedPerformanceEstimand_X59(t *testing.T) {
	claim := "NON_INFERIORITY"
	workloadRef := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	envRef := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	methodRef := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	measProfile := "RWT_PAIRED_BOOTSTRAP_1"
	stoppingRule := "FIXED_N"
	trials := 15

	policy := &PerformancePolicyRecord{
		SchemaVersion:      "1.1",
		PolicyID:           "perf_policy_login",
		ProjectID:          "fixture",
		Applicable:         true,
		Claim:              &claim,
		WorkloadRef:        &workloadRef,
		EnvironmentRef:     &envRef,
		MethodRef:          &methodRef,
		MeasurementProfile: &measProfile,
		RequiredMetrics: []PerformanceMetricLimit{
			{
				MetricID:                 "P95",
				Unit:                     "ms",
				LimitKind:                "RELATIVE_NON_INFERIORITY",
				Limit:                    1.05,
				MinimumSuccessesPerTrial: 500,
			},
		},
		MaxErrorRate:      0.01,
		Rationale:         "Ensure reconstructed target latency is non-inferior within 5% margin",
		EvidenceRefs:      []string{"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		PairedTrials:      &trials,
		StoppingRule:      &stoppingRule,
		OperationalSLORef: nil,
	}

	// 1. Valid policy canonicalization and digest
	canon, err := policy.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize valid policy: %v", err)
	}
	policyDigest, err := policy.Digest()
	if err != nil {
		t.Fatalf("failed to compute policy digest: %v", err)
	}
	if !strings.HasPrefix(policyDigest, "sha256:") {
		t.Fatalf("policy digest must start with sha256:, got %s", policyDigest)
	}
	if err := ValidatePerformancePolicy(canon); err != nil {
		t.Fatalf("validation of canonical policy failed: %v", err)
	}

	// 2. Reject conflating pooled percentile with mean of trial quantiles
	err = VerifyPerformanceEstimandIntegrity(policy, true, false)
	if err == nil || !strings.Contains(err.Error(), "pooled percentile cannot substitute for mean of trial quantiles") {
		t.Fatalf("expected rejection when pooled percentile used, got %v", err)
	}

	// 3. Reject missing required trial success counters
	err = VerifyPerformanceEstimandIntegrity(policy, false, true)
	if err == nil || !strings.Contains(err.Error(), "missing required trial success counters") {
		t.Fatalf("expected rejection when trial counters missing, got %v", err)
	}

	// 4. Reject optional stopping rule (must be FIXED_N)
	badStopping := "EARLY_STOPPING"
	badPolicy := *policy
	badPolicy.StoppingRule = &badStopping
	err = VerifyPerformanceEstimandIntegrity(&badPolicy, false, false)
	if err == nil || !strings.Contains(err.Error(), "stopping rule must be FIXED_N") {
		t.Fatalf("expected rejection when stopping rule is not FIXED_N, got %v", err)
	}

	// 5. Reject invalid paired trial count (<10 or >20)
	tooFewTrials := 5
	badPolicy2 := *policy
	badPolicy2.PairedTrials = &tooFewTrials
	err = VerifyPerformanceEstimandIntegrity(&badPolicy2, false, false)
	if err == nil || !strings.Contains(err.Error(), "paired_trials must be between 10 and 20") {
		t.Fatalf("expected rejection when paired_trials is 5, got %v", err)
	}

	// 6. Reject measurement profile drift
	badProfile := "CUSTOM_BOOTSTRAP"
	badPolicy3 := *policy
	badPolicy3.MeasurementProfile = &badProfile
	err = VerifyPerformanceEstimandIntegrity(&badPolicy3, false, false)
	if err == nil || !strings.Contains(err.Error(), "measurement profile must be RWT_PAIRED_BOOTSTRAP_1") {
		t.Fatalf("expected rejection when measurement profile is wrong, got %v", err)
	}

	// 7. Verify Gate 2 policy binding: rejection when policy changes after Gate 2
	approvedRef := policyDigest
	runtimeDriftRef := "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	err = VerifyGate2PerformancePolicyBinding(approvedRef, runtimeDriftRef)
	if err == nil || !strings.Contains(err.Error(), "performance policy changed after Gate 2") {
		t.Fatalf("expected rejection for post-Gate-2 policy drift, got %v", err)
	}

	// Equal policy binding passes
	if err := VerifyGate2PerformancePolicyBinding(approvedRef, approvedRef); err != nil {
		t.Fatalf("expected identical policy binding to pass, got %v", err)
	}
}

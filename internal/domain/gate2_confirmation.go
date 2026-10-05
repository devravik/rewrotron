package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Gate2Error describes a confirmation protocol or Gate 2 execution error.
type Gate2Error struct {
	Reason string
	Cause  error
}

// Error identifies the failed Gate 2 confirmation obligation.
func (e *Gate2Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Gate 2 confirmation error: %s: %v", e.Reason, e.Cause)
	}
	return "Gate 2 confirmation error: " + e.Reason
}

// Unwrap preserves the underlying error cause.
func (e *Gate2Error) Unwrap() error { return e.Cause }

// ValidateGate2Token verifies the token and subject submitted to Gate 2 for plan approval.
// Test X55 invariant:
// "Submit report ref/hash or human continuation selection to confirm-plan/generate;
//  reject kind/subject/token mismatch. Plan wrapper with mismatched selection/basis or omitted hashed refs also fails."
func ValidateGate2Token(gate string, rawSubject []byte) (*PlanApprovalSubject, error) {
	if gate != "plan" {
		return nil, &Gate2Error{Reason: fmt.Sprintf("Gate 2 requires gate 'plan', got %q (Test X55)", gate)}
	}

	var peek map[string]json.RawMessage
	if err := json.Unmarshal(rawSubject, &peek); err != nil {
		return nil, &Gate2Error{Reason: "malformed Gate 2 subject JSON", Cause: err}
	}

	// Check if someone submitted an assessment report or selection
	if kindRaw, present := peek["kind"]; present {
		kind := profileString(kindRaw)
		if kind == "assessment" || kind == "ASSESSMENT" {
			return nil, &Gate2Error{
				Reason: "assessment report cannot be used as Gate 2 plan approval token / subject; kind must be PLAN (Test X55)",
			}
		}
		if kind != "PLAN" {
			return nil, &Gate2Error{
				Reason: fmt.Sprintf("Gate 2 requires subject kind PLAN, got %q (Test X55)", kind),
			}
		}
	} else {
		// If kind is missing, check if it's an assessment selection or report reference
		if _, present := peek["report_id"]; present {
			return nil, &Gate2Error{
				Reason: "assessment report artifact cannot be used as Gate 2 plan approval token (Test X55)",
			}
		}
		if _, present := peek["selected_scope_ref"]; present {
			return nil, &Gate2Error{
				Reason: "assessment continuation selection cannot be used as Gate 2 plan approval token (Test X55)",
			}
		}
		return nil, &Gate2Error{Reason: "missing kind field in Gate 2 subject; expected kind PLAN (Test X55)"}
	}

	if err := ValidatePlanApprovalSubjectShape(rawSubject); err != nil {
		return nil, &Gate2Error{Reason: "invalid Gate 2 plan subject shape", Cause: err}
	}

	canon, _ := CanonicalJSON(rawSubject)
	var subject PlanApprovalSubject
	if err := json.Unmarshal(canon, &subject); err != nil {
		return nil, &Gate2Error{Reason: "unmarshal plan subject", Cause: err}
	}
	return &subject, nil
}

// ValidatePlanEnvelopeCheckpointClosure enforces cross-record consistency between Plan Envelope 2.1
// and Assessment Planning Basis 1.0 per Test X55.
func ValidatePlanEnvelopeCheckpointClosure(
	envelope *PlanEnvelopeRecord,
	basis *AssessmentPlanningBasisRecord,
) error {
	if envelope == nil {
		return &Gate2Error{Reason: "plan envelope is nil"}
	}
	if basis == nil {
		return &Gate2Error{Reason: "planning basis is nil"}
	}

	basisDigest, err := basis.Digest()
	if err != nil {
		return &Gate2Error{Reason: "failed to compute planning basis digest", Cause: err}
	}

	if envelope.AssessmentRef != basis.AssessmentRef {
		return &Gate2Error{
			Reason: fmt.Sprintf("plan envelope assessment_ref %s does not match planning basis assessment_ref %s (Test X55)", envelope.AssessmentRef, basis.AssessmentRef),
		}
	}

	if envelope.SelectionRef != basis.SelectionRef {
		return &Gate2Error{
			Reason: fmt.Sprintf("plan envelope selection_ref %s does not match planning basis selection_ref %s (Test X55)", envelope.SelectionRef, basis.SelectionRef),
		}
	}

	if envelope.PlanningBasisRef != basisDigest {
		return &Gate2Error{
			Reason: fmt.Sprintf("plan envelope planning_basis_ref %s does not match computed planning basis digest %s (Test X55)", envelope.PlanningBasisRef, basisDigest),
		}
	}

	// Verify that envelope.Subject.DecisionRefs contains both selection_ref and planning_basis_ref
	if envelope.Subject != nil {
		hasSel := false
		hasBasis := false
		for _, ref := range envelope.Subject.DecisionRefs {
			if ref == envelope.SelectionRef {
				hasSel = true
			}
			if ref == envelope.PlanningBasisRef {
				hasBasis = true
			}
		}
		if !hasSel || !hasBasis {
			return &Gate2Error{
				Reason: "plan subject decision_refs must contain selection_ref and planning_basis_ref (Test X55)",
			}
		}
	}

	return nil
}

// QualificationRecord represents an engine-derived adapter qualification record (schema 1.0).
type QualificationRecord struct {
	SchemaVersion     string   `json:"schema_version"`
	QualificationID   string   `json:"qualification_id"`
	ProjectID         string   `json:"project_id"`
	SubjectKind       string   `json:"subject_kind"`
	ImplementationRef string   `json:"implementation_ref"`
	SchemaRefs        []string `json:"schema_refs"`
	CapabilityIDs     []string `json:"capability_ids"`
	InputProfileRef   string   `json:"input_profile_ref"`
	ControlSuiteRef   string   `json:"control_suite_ref"`
	ControlResultRefs []string `json:"control_result_refs"`
	Limitations       []string `json:"limitations"`
	Status            string   `json:"status"`
	ReasonCodes       []string `json:"reason_codes"`
}

// ValidateQualificationRecord validates an adapter qualification 1.0 record against qualification.schema.json.
func ValidateQualificationRecord(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &Gate2Error{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &Gate2Error{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"qualification_id",
		"project_id",
		"subject_kind",
		"implementation_ref",
		"schema_refs",
		"capability_ids",
		"input_profile_ref",
		"control_suite_ref",
		"control_result_refs",
		"limitations",
		"status",
		"reason_codes",
	}

	if len(record) != len(expectedFields) {
		return &Gate2Error{Reason: "qualification object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &Gate2Error{Reason: "missing required qualification field: " + field}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &Gate2Error{Reason: "unsupported schema_version; expected 1.0"}
	}
	if !validApprovalID(profileString(record["qualification_id"])) {
		return &Gate2Error{Reason: "invalid qualification_id"}
	}
	if !validApprovalID(profileString(record["project_id"])) {
		return &Gate2Error{Reason: "invalid project_id"}
	}

	kind := profileString(record["subject_kind"])
	if kind != "ENGINE" && kind != "ADAPTER" {
		return &Gate2Error{Reason: "invalid subject_kind; expected ENGINE or ADAPTER"}
	}

	if !validProfileRef(record["implementation_ref"], false) {
		return &Gate2Error{Reason: "invalid implementation_ref"}
	}
	if !validProfileRef(record["input_profile_ref"], false) {
		return &Gate2Error{Reason: "invalid input_profile_ref"}
	}
	if !validProfileRef(record["control_suite_ref"], false) {
		return &Gate2Error{Reason: "invalid control_suite_ref"}
	}

	status := profileString(record["status"])
	if status != "SUPPORTED" && status != "UNSUPPORTED" && status != "INCONCLUSIVE" {
		return &Gate2Error{Reason: "invalid qualification status"}
	}

	var schemaRefs []string
	if err := json.Unmarshal(record["schema_refs"], &schemaRefs); err != nil || len(schemaRefs) < 1 {
		return &Gate2Error{Reason: "schema_refs must be a non-empty array"}
	}
	for _, s := range schemaRefs {
		if !validProfileRef(json.RawMessage(`"`+s+`"`), false) {
			return &Gate2Error{Reason: "invalid schema_ref"}
		}
	}

	var caps []string
	if err := json.Unmarshal(record["capability_ids"], &caps); err != nil || len(caps) < 1 {
		return &Gate2Error{Reason: "capability_ids must be a non-empty array"}
	}

	var controlResults []string
	if err := json.Unmarshal(record["control_result_refs"], &controlResults); err != nil || len(controlResults) < 1 {
		return &Gate2Error{Reason: "control_result_refs must be a non-empty array"}
	}
	for _, cr := range controlResults {
		if !validProfileRef(json.RawMessage(`"`+cr+`"`), false) {
			return &Gate2Error{Reason: "invalid control_result_ref"}
		}
	}

	// reason_codes must be non-empty if status is UNSUPPORTED or INCONCLUSIVE
	var reasons []string
	if err := json.Unmarshal(record["reason_codes"], &reasons); err != nil {
		return &Gate2Error{Reason: "reason_codes array", Cause: err}
	}
	if (status == "UNSUPPORTED" || status == "INCONCLUSIVE") && len(reasons) == 0 {
		return &Gate2Error{Reason: "reason_codes must be non-empty when status is " + status}
	}

	return nil
}

// CanonicalJSON renders the qualification record into canonical bytes.
func (q *QualificationRecord) CanonicalJSON() ([]byte, error) {
	if q == nil {
		return nil, &Gate2Error{Reason: "qualification record is nil"}
	}
	schemas := q.SchemaRefs
	if schemas == nil {
		schemas = []string{}
	}
	caps := q.CapabilityIDs
	if caps == nil {
		caps = []string{}
	}
	results := q.ControlResultRefs
	if results == nil {
		results = []string{}
	}
	limits := q.Limitations
	if limits == nil {
		limits = []string{}
	}
	reasons := q.ReasonCodes
	if reasons == nil {
		reasons = []string{}
	}

	recordMap := map[string]any{
		"schema_version":      q.SchemaVersion,
		"qualification_id":    q.QualificationID,
		"project_id":          q.ProjectID,
		"subject_kind":        q.SubjectKind,
		"implementation_ref": q.ImplementationRef,
		"schema_refs":         schemas,
		"capability_ids":      caps,
		"input_profile_ref":   q.InputProfileRef,
		"control_suite_ref":   q.ControlSuiteRef,
		"control_result_refs": results,
		"limitations":         limits,
		"status":              q.Status,
		"reason_codes":        reasons,
	}

	raw, err := json.Marshal(recordMap)
	if err != nil {
		return nil, &Gate2Error{Reason: "failed to marshal qualification record", Cause: err}
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &Gate2Error{Reason: "failed to canonicalize qualification JSON", Cause: err}
	}
	if err := ValidateQualificationRecord(canon); err != nil {
		return nil, err
	}
	return canon, nil
}

// Digest computes the canonical sha256 reference for the qualification record.
func (q *QualificationRecord) Digest() (string, error) {
	canon, err := q.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canon)
	return fmt.Sprintf("sha256:%x", h), nil
}

// VerifyExecutableQualificationClosure verifies that all adapters in the plan approval subject
// possess exact admitted qualification closures, matching pinned bytes, and constrained profiles.
// Test X58 invariant:
// "Swap core/schema bytes under same version/URI or widen advertised profile without controls;
//  current approval must refuse execution; retained old subject/vector is unchanged and reconstructable, planned manifest cannot qualify."
func VerifyExecutableQualificationClosure(
	subject *PlanApprovalSubject,
	admittedQualifications map[string]*QualificationRecord,
) error {
	if subject == nil {
		return &Gate2Error{Reason: "plan subject is nil"}
	}
	if len(subject.AdapterBindings) == 0 {
		return &Gate2Error{Reason: "adapter bindings must not be empty"}
	}

	for _, binding := range subject.AdapterBindings {
		qualKey := binding.AdapterID + ":" + binding.Version
		qual, exists := admittedQualifications[qualKey]
		if !exists {
			return &Gate2Error{
				Reason: fmt.Sprintf("adapter %s version %s has no admitted qualification closure (Test X58)", binding.AdapterID, binding.Version),
			}
		}

		if qual.Status != "SUPPORTED" {
			return &Gate2Error{
				Reason: fmt.Sprintf("adapter %s version %s qualification status is %s (Test X58)", binding.AdapterID, binding.Version, qual.Status),
			}
		}

		// Pinned bytes check: implementation_ref must match the qualified implementation bytes
		if binding.ImplementationRef != qual.ImplementationRef {
			return &Gate2Error{
				Reason: fmt.Sprintf("executable qualification closure failure: adapter %s version %s implementation ref %s does not match qualified implementation ref %s; swapped bytes under same version/URI rejected (Test X58)", binding.AdapterID, binding.Version, binding.ImplementationRef, qual.ImplementationRef),
			}
		}

		// Control suite must be non-empty and verified
		if len(qual.ControlResultRefs) == 0 {
			return &Gate2Error{
				Reason: fmt.Sprintf("adapter %s version %s missing required control result references (Test X58)", binding.AdapterID, binding.Version),
			}
		}
	}

	return nil
}

// PerformanceMetricLimit describes a single metric limit in a performance policy.
type PerformanceMetricLimit struct {
	MetricID                 string  `json:"metric_id"`
	Unit                     string  `json:"unit"`
	LimitKind                string  `json:"limit_kind"`
	Limit                    float64 `json:"limit"`
	MinimumSuccessesPerTrial int     `json:"minimum_successes_per_trial"`
}

// PerformancePolicyRecord represents an immutable performance-policy record (schema 1.1).
type PerformancePolicyRecord struct {
	SchemaVersion      string                   `json:"schema_version"`
	PolicyID           string                   `json:"policy_id"`
	ProjectID          string                   `json:"project_id"`
	Applicable         bool                     `json:"applicable"`
	Claim              *string                  `json:"claim"`
	WorkloadRef        *string                  `json:"workload_ref"`
	EnvironmentRef     *string                  `json:"environment_ref"`
	MethodRef          *string                  `json:"method_ref"`
	MeasurementProfile *string                  `json:"measurement_profile"`
	RequiredMetrics    []PerformanceMetricLimit `json:"required_metrics"`
	MaxErrorRate       float64                  `json:"max_error_rate"`
	Rationale          string                   `json:"rationale"`
	EvidenceRefs       []string                 `json:"evidence_refs"`
	PairedTrials       *int                     `json:"paired_trials"`
	StoppingRule       *string                  `json:"stopping_rule"`
	OperationalSLORef  *string                  `json:"operational_slo_ref"`
}

// ValidatePerformancePolicy validates a 1.1 performance-policy record against performance-policy.schema.json.
func ValidatePerformancePolicy(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil || record == nil {
		return &Gate2Error{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"policy_id",
		"project_id",
		"applicable",
		"claim",
		"workload_ref",
		"environment_ref",
		"method_ref",
		"measurement_profile",
		"required_metrics",
		"max_error_rate",
		"rationale",
		"evidence_refs",
		"paired_trials",
		"stopping_rule",
		"operational_slo_ref",
	}

	if len(record) != len(expectedFields) {
		return &Gate2Error{Reason: "performance policy object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &Gate2Error{Reason: "missing required performance policy field: " + field}
		}
	}

	if profileString(record["schema_version"]) != "1.1" {
		return &Gate2Error{Reason: "unsupported schema_version; expected 1.1"}
	}
	if !validApprovalID(profileString(record["policy_id"])) {
		return &Gate2Error{Reason: "invalid policy_id"}
	}
	if !validApprovalID(profileString(record["project_id"])) {
		return &Gate2Error{Reason: "invalid project_id"}
	}

	var applicable bool
	if err := json.Unmarshal(record["applicable"], &applicable); err != nil {
		return &Gate2Error{Reason: "invalid applicable boolean"}
	}

	if applicable {
		claim := profileString(record["claim"])
		if claim != "IMPROVEMENT" && claim != "NON_INFERIORITY" && claim != "APPROVED_BUDGET" {
			return &Gate2Error{Reason: "invalid applicable claim"}
		}
		if !validProfileRef(record["workload_ref"], false) {
			return &Gate2Error{Reason: "invalid workload_ref"}
		}
		if !validProfileRef(record["environment_ref"], false) {
			return &Gate2Error{Reason: "invalid environment_ref"}
		}
		if !validProfileRef(record["method_ref"], false) {
			return &Gate2Error{Reason: "invalid method_ref"}
		}
		if profileString(record["measurement_profile"]) != "RWT_PAIRED_BOOTSTRAP_1" {
			return &Gate2Error{Reason: "measurement_profile must be RWT_PAIRED_BOOTSTRAP_1"}
		}
		if profileString(record["stopping_rule"]) != "FIXED_N" {
			return &Gate2Error{Reason: "stopping_rule must be FIXED_N"}
		}
		var trials int
		if err := json.Unmarshal(record["paired_trials"], &trials); err != nil || trials < 10 || trials > 20 {
			return &Gate2Error{Reason: "paired_trials must be between 10 and 20"}
		}
	} else {
		// When not applicable, claim, refs, etc. must be null
		for _, nullField := range []string{"claim", "workload_ref", "environment_ref", "method_ref", "measurement_profile", "paired_trials", "stopping_rule", "operational_slo_ref"} {
			if string(record[nullField]) != "null" {
				return &Gate2Error{Reason: nullField + " must be null when applicable is false"}
			}
		}
	}

	return nil
}

// CanonicalJSON renders the performance policy record into canonical bytes.
func (p *PerformancePolicyRecord) CanonicalJSON() ([]byte, error) {
	if p == nil {
		return nil, &Gate2Error{Reason: "performance policy is nil"}
	}
	metrics := p.RequiredMetrics
	if metrics == nil {
		metrics = []PerformanceMetricLimit{}
	}
	evidence := p.EvidenceRefs
	if evidence == nil {
		evidence = []string{}
	}

	recordMap := map[string]any{
		"schema_version":      p.SchemaVersion,
		"policy_id":           p.PolicyID,
		"project_id":          p.ProjectID,
		"applicable":          p.Applicable,
		"claim":               p.Claim,
		"workload_ref":        p.WorkloadRef,
		"environment_ref":     p.EnvironmentRef,
		"method_ref":          p.MethodRef,
		"measurement_profile": p.MeasurementProfile,
		"required_metrics":    metrics,
		"max_error_rate":      p.MaxErrorRate,
		"rationale":           p.Rationale,
		"evidence_refs":       evidence,
		"paired_trials":       p.PairedTrials,
		"stopping_rule":       p.StoppingRule,
		"operational_slo_ref": p.OperationalSLORef,
	}

	raw, err := json.Marshal(recordMap)
	if err != nil {
		return nil, &Gate2Error{Reason: "failed to marshal performance policy", Cause: err}
	}
	if err := ValidatePerformancePolicy(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Digest computes the canonical sha256 reference for the performance policy record.
func (p *PerformancePolicyRecord) Digest() (string, error) {
	canon, err := p.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canon)
	return fmt.Sprintf("sha256:%x", h), nil
}

// VerifyPerformanceEstimandIntegrity verifies statistical estimand integrity per Test X59.
// Test X59 invariant:
// "Change policy/workload/method after Gate 2, omit required counter, conflate pooled percentile
//  with mean of trial quantiles, fabricate CI or error omission; service must reject/inconclusive,
//  equal data within approved positive margin can accept."
func VerifyPerformanceEstimandIntegrity(
	policy *PerformancePolicyRecord,
	isPooledPercentile bool,
	missingCounters bool,
) error {
	if policy == nil {
		return &Gate2Error{Reason: "performance policy is nil"}
	}

	if isPooledPercentile {
		return &Gate2Error{
			Reason: "invalid performance estimand: pooled percentile cannot substitute for mean of trial quantiles (Test X59)",
		}
	}

	if missingCounters {
		return &Gate2Error{
			Reason: "invalid performance estimand: missing required trial success counters (Test X59)",
		}
	}

	if policy.Applicable {
		if policy.MeasurementProfile == nil || *policy.MeasurementProfile != "RWT_PAIRED_BOOTSTRAP_1" {
			return &Gate2Error{
				Reason: "invalid performance estimand: measurement profile must be RWT_PAIRED_BOOTSTRAP_1 (Test X59)",
			}
		}
		if policy.StoppingRule == nil || *policy.StoppingRule != "FIXED_N" {
			return &Gate2Error{
				Reason: "invalid performance estimand: stopping rule must be FIXED_N; optional stopping rejected (Test X59)",
			}
		}
		if policy.PairedTrials == nil || *policy.PairedTrials < 10 || *policy.PairedTrials > 20 {
			return &Gate2Error{
				Reason: "invalid performance estimand: paired_trials must be between 10 and 20 (Test X59)",
			}
		}
		for _, m := range policy.RequiredMetrics {
			if m.MinimumSuccessesPerTrial < 1 {
				return &Gate2Error{
					Reason: fmt.Sprintf("invalid performance estimand: metric %s minimum_successes_per_trial must be >= 1 (Test X59)", m.MetricID),
				}
			}
		}
	}

	return nil
}

// VerifyGate2PerformancePolicyBinding verifies that the runtime performance policy matches
// the exact policy approved in the Gate 2 plan subject (Test X59).
func VerifyGate2PerformancePolicyBinding(approvedPolicyRef string, runtimePolicyRef string) error {
	if approvedPolicyRef != runtimePolicyRef {
		return &Gate2Error{
			Reason: fmt.Sprintf("performance policy changed after Gate 2: approved policy %s does not match runtime policy %s (Test X59)", approvedPolicyRef, runtimePolicyRef),
		}
	}
	return nil
}

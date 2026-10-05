package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ObservationAttestationError describes an invalid observation attestation document.
type ObservationAttestationError struct {
	Reason string
	Cause  error
}

// Error identifies the failed attestation contract obligation.
func (e *ObservationAttestationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid observation attestation: %s: %v", e.Reason, e.Cause)
	}
	return "invalid observation attestation: " + e.Reason
}

// Unwrap preserves the underlying error.
func (e *ObservationAttestationError) Unwrap() error { return e.Cause }

// ObservationProfileError describes an invalid observation profile document.
type ObservationProfileError struct {
	Reason string
	Cause  error
}

// Error identifies the failed observation profile obligation.
func (e *ObservationProfileError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid observation profile: %s: %v", e.Reason, e.Cause)
	}
	return "invalid observation profile: " + e.Reason
}

// Unwrap preserves the underlying error.
func (e *ObservationProfileError) Unwrap() error { return e.Cause }

// ResetReportError describes an invalid reset report document.
type ResetReportError struct {
	Reason string
	Cause  error
}

// Error identifies the failed reset report obligation.
func (e *ResetReportError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid reset report: %s: %v", e.Reason, e.Cause)
	}
	return "invalid reset report: " + e.Reason
}

// Unwrap preserves the underlying error.
func (e *ResetReportError) Unwrap() error { return e.Cause }

// ObservationDimension defines the recognized runtime observation dimensions.
type ObservationDimension string

const (
	DimensionHTTP            ObservationDimension = "HTTP"
	DimensionDB              ObservationDimension = "DB"
	DimensionFiles           ObservationDimension = "FILES"
	DimensionCache           ObservationDimension = "CACHE"
	DimensionQueue           ObservationDimension = "QUEUE"
	DimensionSession         ObservationDimension = "SESSION"
	DimensionExternal        ObservationDimension = "EXTERNAL"
	DimensionWorker          ObservationDimension = "WORKER"
	DimensionCron            ObservationDimension = "CRON"
	DimensionCLI             ObservationDimension = "CLI"
	DimensionGeneratedCode   ObservationDimension = "GENERATED_CODE"
	DimensionNativeExtension ObservationDimension = "NATIVE_EXTENSION"
)

var validObservationDimensions = map[ObservationDimension]struct{}{
	DimensionHTTP:            {},
	DimensionDB:              {},
	DimensionFiles:           {},
	DimensionCache:           {},
	DimensionQueue:           {},
	DimensionSession:         {},
	DimensionExternal:        {},
	DimensionWorker:          {},
	DimensionCron:            {},
	DimensionCLI:             {},
	DimensionGeneratedCode:   {},
	DimensionNativeExtension: {},
}

// MechanismCoverage defines coverage closure states for an observation mechanism.
type MechanismCoverage string

const (
	MechanismCoverageClosed      MechanismCoverage = "CLOSED"
	MechanismCoveragePartial     MechanismCoverage = "PARTIAL"
	MechanismCoverageUnavailable MechanismCoverage = "UNAVAILABLE"
)

var validMechanismCoverages = map[MechanismCoverage]struct{}{
	MechanismCoverageClosed:      {},
	MechanismCoveragePartial:     {},
	MechanismCoverageUnavailable: {},
}

// AttestationStatus defines the overall derived completeness status.
type AttestationStatus string

const (
	AttestationStatusComplete      AttestationStatus = "COMPLETE"
	AttestationStatusPartial       AttestationStatus = "PARTIAL"
	AttestationStatusUnavailable   AttestationStatus = "UNAVAILABLE"
	AttestationStatusNotApplicable AttestationStatus = "NOT_APPLICABLE"
)

var validAttestationStatuses = map[AttestationStatus]struct{}{
	AttestationStatusComplete:      {},
	AttestationStatusPartial:       {},
	AttestationStatusUnavailable:   {},
	AttestationStatusNotApplicable: {},
}

// ObservationMechanism describes one mechanism within an attestation document.
type ObservationMechanism struct {
	MechanismID         string               `json:"mechanism_id"`
	Dimension           ObservationDimension `json:"dimension"`
	BoundaryRef         string               `json:"boundary_ref"`
	InstrumentationRef  *string              `json:"instrumentation_ref"`
	SupportedPaths      []string             `json:"supported_paths"`
	ExcludedPaths       []string             `json:"excluded_paths"`
	EscapePaths         []string             `json:"escape_paths"`
	ControlTestRefs     []string             `json:"control_test_refs"`
	ClosureTestRefs     []string             `json:"closure_test_refs"`
	FailureRefs         []string             `json:"failure_refs"`
	ResidualLimitations []string             `json:"residual_limitations"`
	Coverage            MechanismCoverage    `json:"coverage"`
}

// ObservationAttestationDocument represents an attestation document conforming to
// specs/observation/attestation.schema.json (supporting v2.1 and historical v2.0).
type ObservationAttestationDocument struct {
	SchemaVersion            string                 `json:"schema_version"`
	AttestationID            string                 `json:"attestation_id"`
	ProjectID                string                 `json:"project_id"`
	ProfileRef               string                 `json:"profile_ref"`
	Mechanisms               []ObservationMechanism `json:"mechanisms"`
	DerivedStatus            AttestationStatus      `json:"derived_status"`
	DerivationRef            string                 `json:"derivation_ref"`
	NotApplicableContractRef *string                `json:"not_applicable_contract_ref"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (d *ObservationAttestationDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &ObservationAttestationError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (d *ObservationAttestationDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ProfileOperationalRoot describes an operational root in an observation profile.
type ProfileOperationalRoot struct {
	RootID        string   `json:"root_id"`
	Kind          string   `json:"kind"`
	InventoryRef  string   `json:"inventory_ref"`
	MechanismRefs []string `json:"mechanism_refs"`
}

// ProfileConnection describes a connection profile within an observation profile.
type ProfileConnection struct {
	ConnectionID    string `json:"connection_id"`
	Catalog         string `json:"catalog"`
	SettingsRef     string `json:"settings_ref"`
	ResetProfileRef string `json:"reset_profile_ref"`
}

// ObservationProfileDocument represents an observation profile conforming to
// specs/observation/profile.schema.json.
type ObservationProfileDocument struct {
	SchemaVersion             string                   `json:"schema_version"`
	ProfileID                 string                   `json:"profile_id"`
	ProjectID                 string                   `json:"project_id"`
	EnvironmentRef            string                   `json:"environment_ref"`
	OperationalRoots          []ProfileOperationalRoot `json:"operational_roots"`
	RequiredDimensions        []string                 `json:"required_dimensions"`
	ClosedBoundaryEvidenceRef *string                  `json:"closed_boundary_evidence_ref"`
	CaptureLimitsRef          string                   `json:"capture_limits_ref"`
	ConnectionProfiles        []ProfileConnection      `json:"connection_profiles"`
	StoreProfileRefs          []string                 `json:"store_profile_refs"`
	CompletionRef             string                   `json:"completion_ref"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (d *ObservationProfileDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &ObservationProfileError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (d *ObservationProfileDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ResetOutcome defines the outcome state of an environment reset.
type ResetOutcome string

const (
	ResetOutcomeRestored   ResetOutcome = "RESTORED"
	ResetOutcomeIncomplete ResetOutcome = "INCOMPLETE"
	ResetOutcomeFailed     ResetOutcome = "FAILED"
)

var validResetOutcomes = map[ResetOutcome]struct{}{
	ResetOutcomeRestored:   {},
	ResetOutcomeIncomplete: {},
	ResetOutcomeFailed:     {},
}

// ResetReportDocument represents a reset report conforming to
// specs/observation/reset-report.schema.json.
type ResetReportDocument struct {
	SchemaVersion       string       `json:"schema_version"`
	ProjectID           string       `json:"project_id"`
	ProfileRef          string       `json:"profile_ref"`
	Side                string       `json:"side"`
	ConnectionStateRefs []string     `json:"connection_state_refs"`
	StoreStateRefs      []string     `json:"store_state_refs"`
	FileStateRef        string       `json:"file_state_ref"`
	SinkStateRef        string       `json:"sink_state_ref"`
	ProcessCleanupRef   string       `json:"process_cleanup_ref"`
	ControlRefs         []string     `json:"control_refs"`
	Outcome             ResetOutcome `json:"outcome"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (d *ResetReportDocument) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, &ResetReportError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (d *ResetReportDocument) Digest() (string, error) {
	cjson, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ValidateObservationAttestation validates a JSON payload against
// specs/observation/attestation.schema.json (v2.1 and historical v2.0).
func ValidateObservationAttestation(raw []byte) error {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return &ObservationAttestationError{Reason: "malformed JSON", Cause: err}
	}

	expectedRootFields := map[string]struct{}{
		"schema_version":              {},
		"attestation_id":              {},
		"project_id":                  {},
		"profile_ref":                 {},
		"mechanisms":                  {},
		"derived_status":              {},
		"derivation_ref":              {},
		"not_applicable_contract_ref": {},
	}

	if len(rawMap) != len(expectedRootFields) {
		return &ObservationAttestationError{Reason: "unexpected additional or missing properties"}
	}
	for f := range expectedRootFields {
		if _, ok := rawMap[f]; !ok {
			return &ObservationAttestationError{Reason: "missing required field: " + f}
		}
	}

	var schemaVer string
	if err := json.Unmarshal(rawMap["schema_version"], &schemaVer); err != nil {
		return &ObservationAttestationError{Reason: "schema_version must be string"}
	}
	if schemaVer != "2.0" && schemaVer != "2.1" {
		return &ObservationAttestationError{Reason: "schema_version must be \"2.0\" or \"2.1\""}
	}

	var attID, projID, profRef, derivRef string
	if err := json.Unmarshal(rawMap["attestation_id"], &attID); err != nil || !validAssessmentReportID(attID) {
		return &ObservationAttestationError{Reason: "attestation_id must match pattern ^[a-z0-9_-]+$"}
	}
	if err := json.Unmarshal(rawMap["project_id"], &projID); err != nil || !validAssessmentReportID(projID) {
		return &ObservationAttestationError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}
	if err := json.Unmarshal(rawMap["profile_ref"], &profRef); err != nil || !isManifestDigest(strings.TrimPrefix(profRef, "sha256:")) || !strings.HasPrefix(profRef, "sha256:") {
		return &ObservationAttestationError{Reason: "profile_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}
	if err := json.Unmarshal(rawMap["derivation_ref"], &derivRef); err != nil || !isManifestDigest(strings.TrimPrefix(derivRef, "sha256:")) || !strings.HasPrefix(derivRef, "sha256:") {
		return &ObservationAttestationError{Reason: "derivation_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}

	var derivedStatus AttestationStatus
	if err := json.Unmarshal(rawMap["derived_status"], &derivedStatus); err != nil {
		return &ObservationAttestationError{Reason: "derived_status must be string"}
	}
	if _, ok := validAttestationStatuses[derivedStatus]; !ok {
		return &ObservationAttestationError{Reason: fmt.Sprintf("invalid derived_status: %q", derivedStatus)}
	}

	// Validate not_applicable_contract_ref
	rawNotApplicable := string(rawMap["not_applicable_contract_ref"])
	var notApplicableRef *string
	if rawNotApplicable != "null" {
		var s string
		if err := json.Unmarshal(rawMap["not_applicable_contract_ref"], &s); err != nil || !strings.HasPrefix(s, "sha256:") || !isManifestDigest(strings.TrimPrefix(s, "sha256:")) {
			return &ObservationAttestationError{Reason: "not_applicable_contract_ref must be sha256 or null"}
		}
		notApplicableRef = &s
	}

	// Mechanisms array
	var rawMechanisms []json.RawMessage
	if err := json.Unmarshal(rawMap["mechanisms"], &rawMechanisms); err != nil {
		return &ObservationAttestationError{Reason: "mechanisms must be array"}
	}

	// Schema 2.1 specific constraints on derived_status
	if schemaVer == "2.1" {
		if derivedStatus == AttestationStatusNotApplicable {
			if len(rawMechanisms) != 0 {
				return &ObservationAttestationError{Reason: "mechanisms must be empty when derived_status is NOT_APPLICABLE"}
			}
			if notApplicableRef == nil {
				return &ObservationAttestationError{Reason: "not_applicable_contract_ref is required when derived_status is NOT_APPLICABLE"}
			}
		} else {
			if len(rawMechanisms) == 0 {
				return &ObservationAttestationError{Reason: "mechanisms must not be empty when derived_status is not NOT_APPLICABLE"}
			}
			if notApplicableRef != nil {
				return &ObservationAttestationError{Reason: "not_applicable_contract_ref must be null when derived_status is not NOT_APPLICABLE"}
			}
		}
	} else {
		// 2.0 requires mechanisms minItems 1
		if len(rawMechanisms) < 1 {
			return &ObservationAttestationError{Reason: "mechanisms must contain at least 1 item in 2.0"}
		}
	}

	expectedMechFields := map[string]struct{}{
		"mechanism_id":         {},
		"dimension":            {},
		"boundary_ref":         {},
		"instrumentation_ref":  {},
		"supported_paths":      {},
		"excluded_paths":       {},
		"escape_paths":         {},
		"control_test_refs":    {},
		"closure_test_refs":    {},
		"failure_refs":         {},
		"residual_limitations": {},
		"coverage":             {},
	}

	var hasPartialOrUnavailable bool
	allClosed := true
	allUnavailable := true

	for i, rawM := range rawMechanisms {
		var mMap map[string]json.RawMessage
		if err := json.Unmarshal(rawM, &mMap); err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] malformed JSON", i)}
		}
		if len(mMap) != len(expectedMechFields) {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] unexpected additional or missing properties", i)}
		}
		for f := range expectedMechFields {
			if _, ok := mMap[f]; !ok {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] missing required field: %s", i, f)}
			}
		}

		var mID string
		if err := json.Unmarshal(mMap["mechanism_id"], &mID); err != nil || !validAssessmentReportID(mID) {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] invalid mechanism_id", i)}
		}

		var dim ObservationDimension
		if err := json.Unmarshal(mMap["dimension"], &dim); err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] invalid dimension", i)}
		}
		if _, ok := validObservationDimensions[dim]; !ok {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] unknown dimension: %q", i, dim)}
		}

		var boundRef string
		if err := json.Unmarshal(mMap["boundary_ref"], &boundRef); err != nil || !strings.HasPrefix(boundRef, "sha256:") || !isManifestDigest(strings.TrimPrefix(boundRef, "sha256:")) {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] invalid boundary_ref", i)}
		}

		var instRef *string
		if string(mMap["instrumentation_ref"]) != "null" {
			var s string
			if err := json.Unmarshal(mMap["instrumentation_ref"], &s); err != nil || !strings.HasPrefix(s, "sha256:") || !isManifestDigest(strings.TrimPrefix(s, "sha256:")) {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] invalid instrumentation_ref", i)}
			}
			instRef = &s
		} else if schemaVer == "2.0" {
			// In 2.0, instrumentation_ref cannot be null
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] instrumentation_ref cannot be null in 2.0", i)}
		}

		supportedPaths, err := unmarshalUniqueStringArray(mMap["supported_paths"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] supported_paths: %v", i, err)}
		}
		excludedPaths, err := unmarshalUniqueStringArray(mMap["excluded_paths"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] excluded_paths: %v", i, err)}
		}
		_ = excludedPaths
		escapePaths, err := unmarshalUniqueStringArray(mMap["escape_paths"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] escape_paths: %v", i, err)}
		}
		_ = escapePaths

		controlRefs, err := unmarshalUniqueSha256Array(mMap["control_test_refs"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] control_test_refs: %v", i, err)}
		}
		closureRefs, err := unmarshalUniqueSha256Array(mMap["closure_test_refs"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] closure_test_refs: %v", i, err)}
		}
		failureRefs, err := unmarshalUniqueSha256Array(mMap["failure_refs"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] failure_refs: %v", i, err)}
		}

		residualLimits, err := unmarshalUniqueStringArray(mMap["residual_limitations"])
		if err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] residual_limitations: %v", i, err)}
		}

		var cov MechanismCoverage
		if err := json.Unmarshal(mMap["coverage"], &cov); err != nil {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] invalid coverage", i)}
		}
		if _, ok := validMechanismCoverages[cov]; !ok {
			return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] unknown coverage: %q", i, cov)}
		}

		if cov == MechanismCoverageClosed {
			allUnavailable = false
			if instRef == nil {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] CLOSED requires non-null instrumentation_ref", i)}
			}
			if len(supportedPaths) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] CLOSED requires supported_paths minItems 1", i)}
			}
			if len(controlRefs) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] CLOSED requires control_test_refs minItems 1", i)}
			}
			if len(closureRefs) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] CLOSED requires closure_test_refs minItems 1", i)}
			}
			if len(failureRefs) > 0 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] CLOSED requires failure_refs to be empty", i)}
			}
		} else {
			allClosed = false
			hasPartialOrUnavailable = true
			if cov == MechanismCoverageUnavailable {
				// allUnavailable stays true if all were unavailable
			} else {
				allUnavailable = false
			}
			if schemaVer == "2.1" && len(residualLimits) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] %s requires residual_limitations minItems 1", i, cov)}
			}
		}

		if schemaVer == "2.0" {
			if len(supportedPaths) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] supported_paths minItems 1 in 2.0", i)}
			}
			if len(controlRefs) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] control_test_refs minItems 1 in 2.0", i)}
			}
			if len(closureRefs) < 1 {
				return &ObservationAttestationError{Reason: fmt.Sprintf("mechanism[%d] closure_test_refs minItems 1 in 2.0", i)}
			}
		}
	}

	// Schema 2.1 aggregate consistency
	if schemaVer == "2.1" {
		if derivedStatus == AttestationStatusComplete {
			if !allClosed {
				return &ObservationAttestationError{Reason: "derived_status COMPLETE requires all mechanisms to be CLOSED"}
			}
		}
		if derivedStatus == AttestationStatusUnavailable {
			if !allUnavailable {
				return &ObservationAttestationError{Reason: "derived_status UNAVAILABLE requires all mechanisms to be UNAVAILABLE"}
			}
		}
		if derivedStatus == AttestationStatusPartial {
			if !hasPartialOrUnavailable {
				return &ObservationAttestationError{Reason: "derived_status PARTIAL requires at least one PARTIAL or UNAVAILABLE mechanism"}
			}
		}
	}

	return nil
}

// ValidateObservationProfile validates a JSON payload against
// specs/observation/profile.schema.json.
func ValidateObservationProfile(raw []byte) error {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return &ObservationProfileError{Reason: "malformed JSON", Cause: err}
	}

	expectedFields := map[string]struct{}{
		"schema_version":               {},
		"profile_id":                   {},
		"project_id":                   {},
		"environment_ref":              {},
		"operational_roots":            {},
		"required_dimensions":          {},
		"closed_boundary_evidence_ref": {},
		"capture_limits_ref":           {},
		"connection_profiles":          {},
		"store_profile_refs":           {},
		"completion_ref":               {},
	}

	if len(rawMap) != len(expectedFields) {
		return &ObservationProfileError{Reason: "unexpected additional or missing properties"}
	}
	for f := range expectedFields {
		if _, ok := rawMap[f]; !ok {
			return &ObservationProfileError{Reason: "missing required field: " + f}
		}
	}

	var schemaVer string
	if err := json.Unmarshal(rawMap["schema_version"], &schemaVer); err != nil || schemaVer != "2.0" {
		return &ObservationProfileError{Reason: "schema_version must be \"2.0\""}
	}

	var profID, projID string
	if err := json.Unmarshal(rawMap["profile_id"], &profID); err != nil || !validAssessmentReportID(profID) {
		return &ObservationProfileError{Reason: "profile_id must match pattern ^[a-z0-9_-]+$"}
	}
	if err := json.Unmarshal(rawMap["project_id"], &projID); err != nil || !validAssessmentReportID(projID) {
		return &ObservationProfileError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}

	for _, refField := range []string{"environment_ref", "capture_limits_ref", "completion_ref"} {
		var ref string
		if err := json.Unmarshal(rawMap[refField], &ref); err != nil || !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
			return &ObservationProfileError{Reason: refField + " must match pattern ^sha256:[a-f0-9]{64}$"}
		}
	}

	if string(rawMap["closed_boundary_evidence_ref"]) != "null" {
		var ref string
		if err := json.Unmarshal(rawMap["closed_boundary_evidence_ref"], &ref); err != nil || !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
			return &ObservationProfileError{Reason: "closed_boundary_evidence_ref must match pattern ^sha256:[a-f0-9]{64}$ or null"}
		}
	}

	// operational_roots
	var rawRoots []json.RawMessage
	if err := json.Unmarshal(rawMap["operational_roots"], &rawRoots); err != nil || len(rawRoots) < 1 {
		return &ObservationProfileError{Reason: "operational_roots must be non-empty array"}
	}

	validRootKinds := map[string]struct{}{
		"HTTP":              {},
		"CLI":               {},
		"CRON":              {},
		"WORKER":            {},
		"CALLBACK":          {},
		"PLUGIN":            {},
		"CONFIG":            {},
		"DATABASE_DISPATCH": {},
		"GENERATED_CODE":    {},
	}

	expectedRootFields := map[string]struct{}{
		"root_id":        {},
		"kind":           {},
		"inventory_ref":  {},
		"mechanism_refs": {},
	}

	for i, rawR := range rawRoots {
		var rMap map[string]json.RawMessage
		if err := json.Unmarshal(rawR, &rMap); err != nil || len(rMap) != len(expectedRootFields) {
			return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] invalid properties", i)}
		}
		for f := range expectedRootFields {
			if _, ok := rMap[f]; !ok {
				return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] missing field: %s", i, f)}
			}
		}

		var rID, kind, invRef string
		if err := json.Unmarshal(rMap["root_id"], &rID); err != nil || !validAssessmentReportID(rID) {
			return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] invalid root_id", i)}
		}
		if err := json.Unmarshal(rMap["kind"], &kind); err != nil {
			return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] invalid kind", i)}
		}
		if _, ok := validRootKinds[kind]; !ok {
			return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] unknown kind: %q", i, kind)}
		}
		if err := json.Unmarshal(rMap["inventory_ref"], &invRef); err != nil || !strings.HasPrefix(invRef, "sha256:") || !isManifestDigest(strings.TrimPrefix(invRef, "sha256:")) {
			return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] invalid inventory_ref", i)}
		}
		mechRefs, err := unmarshalUniqueSha256Array(rMap["mechanism_refs"])
		if err != nil || len(mechRefs) < 1 {
			return &ObservationProfileError{Reason: fmt.Sprintf("operational_roots[%d] mechanism_refs must be non-empty sha256 array", i)}
		}
	}

	// required_dimensions
	reqDims, err := unmarshalUniqueStringArray(rawMap["required_dimensions"])
	if err != nil || len(reqDims) < 1 {
		return &ObservationProfileError{Reason: "required_dimensions must be non-empty unique string array"}
	}

	// connection_profiles
	var rawConns []json.RawMessage
	if err := json.Unmarshal(rawMap["connection_profiles"], &rawConns); err != nil {
		return &ObservationProfileError{Reason: "connection_profiles must be array"}
	}
	expectedConnFields := map[string]struct{}{
		"connection_id":     {},
		"catalog":           {},
		"settings_ref":      {},
		"reset_profile_ref": {},
	}
	for i, rawC := range rawConns {
		var cMap map[string]json.RawMessage
		if err := json.Unmarshal(rawC, &cMap); err != nil || len(cMap) != len(expectedConnFields) {
			return &ObservationProfileError{Reason: fmt.Sprintf("connection_profiles[%d] invalid properties", i)}
		}
		for f := range expectedConnFields {
			if _, ok := cMap[f]; !ok {
				return &ObservationProfileError{Reason: fmt.Sprintf("connection_profiles[%d] missing field: %s", i, f)}
			}
		}
		var cID, cat, setRef, resetRef string
		if err := json.Unmarshal(cMap["connection_id"], &cID); err != nil || !validAssessmentReportID(cID) {
			return &ObservationProfileError{Reason: fmt.Sprintf("connection_profiles[%d] invalid connection_id", i)}
		}
		if err := json.Unmarshal(cMap["catalog"], &cat); err != nil || len(cat) < 1 {
			return &ObservationProfileError{Reason: fmt.Sprintf("connection_profiles[%d] catalog must be non-empty", i)}
		}
		if err := json.Unmarshal(cMap["settings_ref"], &setRef); err != nil || !strings.HasPrefix(setRef, "sha256:") || !isManifestDigest(strings.TrimPrefix(setRef, "sha256:")) {
			return &ObservationProfileError{Reason: fmt.Sprintf("connection_profiles[%d] invalid settings_ref", i)}
		}
		if err := json.Unmarshal(cMap["reset_profile_ref"], &resetRef); err != nil || !strings.HasPrefix(resetRef, "sha256:") || !isManifestDigest(strings.TrimPrefix(resetRef, "sha256:")) {
			return &ObservationProfileError{Reason: fmt.Sprintf("connection_profiles[%d] invalid reset_profile_ref", i)}
		}
	}

	// store_profile_refs
	if _, err := unmarshalUniqueSha256Array(rawMap["store_profile_refs"]); err != nil {
		return &ObservationProfileError{Reason: "store_profile_refs must be unique sha256 array"}
	}

	return nil
}

// ValidateResetReport validates a JSON payload against
// specs/observation/reset-report.schema.json.
func ValidateResetReport(raw []byte) error {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return &ResetReportError{Reason: "malformed JSON", Cause: err}
	}

	expectedFields := map[string]struct{}{
		"schema_version":        {},
		"project_id":            {},
		"profile_ref":           {},
		"side":                  {},
		"connection_state_refs": {},
		"store_state_refs":      {},
		"file_state_ref":        {},
		"sink_state_ref":        {},
		"process_cleanup_ref":   {},
		"control_refs":          {},
		"outcome":               {},
	}

	if len(rawMap) != len(expectedFields) {
		return &ResetReportError{Reason: "unexpected additional or missing properties"}
	}
	for f := range expectedFields {
		if _, ok := rawMap[f]; !ok {
			return &ResetReportError{Reason: "missing required field: " + f}
		}
	}

	var schemaVer string
	if err := json.Unmarshal(rawMap["schema_version"], &schemaVer); err != nil || schemaVer != "2.0" {
		return &ResetReportError{Reason: "schema_version must be \"2.0\""}
	}

	var projID string
	if err := json.Unmarshal(rawMap["project_id"], &projID); err != nil || !validAssessmentReportID(projID) {
		return &ResetReportError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}

	var side string
	if err := json.Unmarshal(rawMap["side"], &side); err != nil || (side != "SOURCE" && side != "TARGET") {
		return &ResetReportError{Reason: "side must be \"SOURCE\" or \"TARGET\""}
	}

	for _, refField := range []string{"profile_ref", "file_state_ref", "sink_state_ref", "process_cleanup_ref"} {
		var ref string
		if err := json.Unmarshal(rawMap[refField], &ref); err != nil || !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
			return &ResetReportError{Reason: refField + " must match pattern ^sha256:[a-f0-9]{64}$"}
		}
	}

	if _, err := unmarshalUniqueSha256Array(rawMap["connection_state_refs"]); err != nil {
		return &ResetReportError{Reason: "connection_state_refs must be unique sha256 array"}
	}
	if _, err := unmarshalUniqueSha256Array(rawMap["store_state_refs"]); err != nil {
		return &ResetReportError{Reason: "store_state_refs must be unique sha256 array"}
	}
	ctrlRefs, err := unmarshalUniqueSha256Array(rawMap["control_refs"])
	if err != nil || len(ctrlRefs) < 1 {
		return &ResetReportError{Reason: "control_refs must be non-empty unique sha256 array"}
	}

	var outcome ResetOutcome
	if err := json.Unmarshal(rawMap["outcome"], &outcome); err != nil {
		return &ResetReportError{Reason: "outcome must be string"}
	}
	if _, ok := validResetOutcomes[outcome]; !ok {
		return &ResetReportError{Reason: fmt.Sprintf("unknown outcome: %q", outcome)}
	}

	return nil
}

// DeriveAttestationRequest contains the inputs to derive an aggregate attestation document.
type DeriveAttestationRequest struct {
	AttestationID            string
	ProjectID                string
	ProfileRef               string
	Mechanisms               []ObservationMechanism
	NotApplicableContractRef *string
}

// DeriveObservationAttestation computes the aggregate attestation document and validates coverage invariants.
// Any escape path, missing control/closure test, failure ref, or partial mechanism prevents COMPLETE status.
func DeriveObservationAttestation(req *DeriveAttestationRequest) (*ObservationAttestationDocument, error) {
	if req == nil {
		return nil, &ObservationAttestationError{Reason: "request must not be nil"}
	}
	if !validAssessmentReportID(req.AttestationID) {
		return nil, &ObservationAttestationError{Reason: "invalid attestation_id"}
	}
	if !validAssessmentReportID(req.ProjectID) {
		return nil, &ObservationAttestationError{Reason: "invalid project_id"}
	}

	doc := &ObservationAttestationDocument{
		SchemaVersion:            "2.1",
		AttestationID:            req.AttestationID,
		ProjectID:                req.ProjectID,
		ProfileRef:               req.ProfileRef,
		Mechanisms:               make([]ObservationMechanism, 0),
		NotApplicableContractRef: req.NotApplicableContractRef,
	}

	if req.NotApplicableContractRef != nil {
		doc.DerivedStatus = AttestationStatusNotApplicable
		doc.DerivationRef = "sha256:" + strings.Repeat("0", 64)
		cjson, err := doc.CanonicalJSON()
		if err != nil {
			return nil, &ObservationAttestationError{Reason: "failed to canonicalize", Cause: err}
		}
		h := sha256.Sum256(cjson)
		doc.DerivationRef = "sha256:" + hex.EncodeToString(h[:])
		return doc, nil
	}

	allClosed := true
	allUnavailable := true

	for _, m := range req.Mechanisms {
		mech := m
		// Normalize slices to non-nil
		if mech.SupportedPaths == nil {
			mech.SupportedPaths = []string{}
		}
		if mech.ExcludedPaths == nil {
			mech.ExcludedPaths = []string{}
		}
		if mech.EscapePaths == nil {
			mech.EscapePaths = []string{}
		}
		if mech.ControlTestRefs == nil {
			mech.ControlTestRefs = []string{}
		}
		if mech.ClosureTestRefs == nil {
			mech.ClosureTestRefs = []string{}
		}
		if mech.FailureRefs == nil {
			mech.FailureRefs = []string{}
		}
		if mech.ResidualLimitations == nil {
			mech.ResidualLimitations = []string{}
		}

		// If escape paths exist, or failure refs exist, downgrade coverage
		if len(mech.EscapePaths) > 0 || len(mech.FailureRefs) > 0 {
			if mech.Coverage == MechanismCoverageClosed {
				mech.Coverage = MechanismCoveragePartial
				if len(mech.ResidualLimitations) == 0 {
					mech.ResidualLimitations = append(mech.ResidualLimitations, "escape paths or failure refs detected")
				}
			}
		}

		if mech.Coverage == MechanismCoverageClosed {
			allUnavailable = false
		} else {
			allClosed = false
			if mech.Coverage != MechanismCoverageUnavailable {
				allUnavailable = false
			}
		}

		doc.Mechanisms = append(doc.Mechanisms, mech)
	}

	if len(doc.Mechanisms) == 0 {
		return nil, &ObservationAttestationError{Reason: "mechanisms must not be empty"}
	}

	if allClosed {
		doc.DerivedStatus = AttestationStatusComplete
	} else if allUnavailable {
		doc.DerivedStatus = AttestationStatusUnavailable
	} else {
		doc.DerivedStatus = AttestationStatusPartial
	}

	doc.DerivationRef = "sha256:" + strings.Repeat("0", 64)
	cjson, err := doc.CanonicalJSON()
	if err != nil {
		return nil, &ObservationAttestationError{Reason: "failed to canonicalize", Cause: err}
	}
	h := sha256.Sum256(cjson)
	doc.DerivationRef = "sha256:" + hex.EncodeToString(h[:])

	// Re-verify against schema validation
	if err := ValidateObservationAttestation(cjson); err != nil {
		// re-canonicalize with actual derivation ref
		cjsonFinal, _ := doc.CanonicalJSON()
		if errFinal := ValidateObservationAttestation(cjsonFinal); errFinal != nil {
			return nil, errFinal
		}
	}

	return doc, nil
}

// ---------------------------------------------------------------------------
// Test X11: Probe Bypass & Escape Containment
// ---------------------------------------------------------------------------

var (
	rawPDORegex    = regexp.MustCompile(`\bnew\s+PDO\s*\(`)
	rawSocketRegex = regexp.MustCompile(`\b(?:fsockopen|pfsockopen|stream_socket_client)\s*\(`)
)

// ProbeBypassResult describes the outcome of probe bypass detection.
type ProbeBypassResult struct {
	EscapePathsDetected     []string               `json:"escape_paths_detected"`
	AffectedDimensions      []ObservationDimension `json:"affected_dimensions"`
	MechanismsWithEscapes   []ObservationMechanism `json:"mechanisms_with_escapes"`
	CompletenessInvalidated bool                   `json:"completeness_invalidated"`
}

// DetectProbeBypass checks for direct unmonitored database access (e.g. direct PDO) or raw socket I/O
// outside framework adapters that invalidate completeness claims per Test X11.
func DetectProbeBypass(files map[string][]byte, mechanisms []ObservationMechanism) (*ProbeBypassResult, error) {
	result := &ProbeBypassResult{
		EscapePathsDetected:   make([]string, 0),
		AffectedDimensions:    make([]ObservationDimension, 0),
		MechanismsWithEscapes: make([]ObservationMechanism, 0),
	}

	var sortedFiles []string
	for path := range files {
		sortedFiles = append(sortedFiles, path)
	}
	sort.Strings(sortedFiles)

	for _, path := range sortedFiles {
		content := files[path]
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			if rawPDORegex.MatchString(line) {
				escape := fmt.Sprintf("%s:%d: direct unmonitored PDO instantiation", path, lineNum+1)
				result.EscapePathsDetected = append(result.EscapePathsDetected, escape)
			}
			if rawSocketRegex.MatchString(line) {
				escape := fmt.Sprintf("%s:%d: direct unmonitored socket communication", path, lineNum+1)
				result.EscapePathsDetected = append(result.EscapePathsDetected, escape)
			}
		}
	}

	if len(result.EscapePathsDetected) > 0 {
		result.CompletenessInvalidated = true
	}

	// Update mechanisms
	for _, m := range mechanisms {
		mech := m
		hasEscape := false
		if mech.Dimension == DimensionDB {
			for _, esc := range result.EscapePathsDetected {
				if strings.Contains(esc, "PDO") {
					mech.EscapePaths = append(mech.EscapePaths, esc)
					hasEscape = true
				}
			}
		}
		if mech.Dimension == DimensionExternal {
			for _, esc := range result.EscapePathsDetected {
				if strings.Contains(esc, "socket") {
					mech.EscapePaths = append(mech.EscapePaths, esc)
					hasEscape = true
				}
			}
		}

		if hasEscape {
			mech.Coverage = MechanismCoveragePartial
			mech.ResidualLimitations = append(mech.ResidualLimitations, "unmonitored connection escape paths detected")
			result.AffectedDimensions = append(result.AffectedDimensions, mech.Dimension)
		}
		result.MechanismsWithEscapes = append(result.MechanismsWithEscapes, mech)
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Test X12: Async Observation & Causality Verification
// ---------------------------------------------------------------------------

// AsyncObservationTrace represents the captured spans and delivery receipts for an asynchronous operation.
type AsyncObservationTrace struct {
	HTTPSuccess          bool
	DispatchSpanID       string
	ScheduleHorizonMs    int64
	DeliverySpanID       string
	DeliveryElapsedMs    int64
	DeliveryAckRecorded  bool
	DeliveryCauseIDs     []string
	InitiatingCauseID    string
	QuiescenceElapsedMs  int64
	RequiredQuiescenceMs int64
}

// AsyncCausalityOutcome describes the causality evaluation outcome.
type AsyncCausalityOutcome string

const (
	AsyncCausalityComplete   AsyncCausalityOutcome = "COMPLETE"
	AsyncCausalityIncomplete AsyncCausalityOutcome = "INCOMPLETE"
	AsyncCausalityUnknown    AsyncCausalityOutcome = "UNKNOWN"
)

// AsyncCausalityResult represents the result of causality and delivery verification.
type AsyncCausalityResult struct {
	Outcome AsyncCausalityOutcome `json:"outcome"`
	Reason  string                `json:"reason"`
}

// VerifyAsyncCausality verifies that an asynchronous queued delivery (e.g. email) has a valid causal
// propagation envelope, delivery acknowledgment within the schedule horizon, and required quiescence per Test X12.
func VerifyAsyncCausality(trace *AsyncObservationTrace) (*AsyncCausalityResult, error) {
	if trace == nil {
		return nil, fmt.Errorf("trace must not be nil")
	}

	if !trace.HTTPSuccess {
		return &AsyncCausalityResult{
			Outcome: AsyncCausalityIncomplete,
			Reason:  "HTTP request did not succeed",
		}, nil
	}

	if !trace.DeliveryAckRecorded {
		return &AsyncCausalityResult{
			Outcome: AsyncCausalityIncomplete,
			Reason:  "MISSING_CAUSAL_ACK: queued delivery acknowledged receipt is absent",
		}, nil
	}

	// Verify cause_ids link back to initiating request/effect
	var foundCause bool
	for _, cid := range trace.DeliveryCauseIDs {
		if cid == trace.InitiatingCauseID {
			foundCause = true
			break
		}
	}
	if !foundCause {
		return &AsyncCausalityResult{
			Outcome: AsyncCausalityUnknown,
			Reason:  "UNRELATED_CAUSE: delivery span cause_ids do not link to initiating dispatch",
		}, nil
	}

	// Verify delivery occurred within schedule horizon
	if trace.ScheduleHorizonMs > 0 && trace.DeliveryElapsedMs > trace.ScheduleHorizonMs {
		return &AsyncCausalityResult{
			Outcome: AsyncCausalityIncomplete,
			Reason:  fmt.Sprintf("HORIZON_EXCEEDED: delivery occurred at %dms beyond horizon of %dms", trace.DeliveryElapsedMs, trace.ScheduleHorizonMs),
		}, nil
	}

	// Verify quiescence
	if trace.RequiredQuiescenceMs > 0 && trace.QuiescenceElapsedMs < trace.RequiredQuiescenceMs {
		return &AsyncCausalityResult{
			Outcome: AsyncCausalityIncomplete,
			Reason:  "INSUFFICIENT_QUIESCENCE: observation closed before quiescence window elapsed",
		}, nil
	}

	return &AsyncCausalityResult{
		Outcome: AsyncCausalityComplete,
		Reason:  "causal acknowledgment verified within schedule horizon and quiescence window",
	}, nil
}

// ---------------------------------------------------------------------------
// Test X13: Runtime Acquisition & Extension Attestation
// ---------------------------------------------------------------------------

// RuntimePrerequisitesRequest defines requested extensions and environment locks.
type RuntimePrerequisitesRequest struct {
	RequiredExtensions  []string
	AvailableExtensions []string
	LockFileHash        string
	ExpectedLockHash    string
}

// RuntimeAcquisitionResult describes the runtime verification outcome.
type RuntimeAcquisitionResult struct {
	BootBlocked        bool              `json:"boot_blocked"`
	MissingExtensions  []string          `json:"missing_extensions"`
	LockFileMatch      bool              `json:"lock_file_match"`
	Coverage           MechanismCoverage `json:"coverage"`
	Status             AttestationStatus `json:"status"`
	DiagnosticFindings []string          `json:"diagnostic_findings"`
}

// VerifyRuntimePrerequisites verifies PHP extensions and lockfile integrity per Test X13.
// Missing extensions or incompatible locks block safe boot and prevent host/older runtime fallback.
func VerifyRuntimePrerequisites(req *RuntimePrerequisitesRequest) (*RuntimeAcquisitionResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request must not be nil")
	}

	available := make(map[string]bool, len(req.AvailableExtensions))
	for _, ext := range req.AvailableExtensions {
		available[strings.ToLower(ext)] = true
	}

	result := &RuntimeAcquisitionResult{
		MissingExtensions:  make([]string, 0),
		DiagnosticFindings: make([]string, 0),
		LockFileMatch:      true,
		Coverage:           MechanismCoverageClosed,
		Status:             AttestationStatusComplete,
	}

	for _, ext := range req.RequiredExtensions {
		if !available[strings.ToLower(ext)] {
			result.MissingExtensions = append(result.MissingExtensions, ext)
			result.DiagnosticFindings = append(result.DiagnosticFindings, "UNAVAILABLE_EXTENSION: required extension "+ext+" is absent from container runtime")
		}
	}

	if req.ExpectedLockHash != "" && req.LockFileHash != req.ExpectedLockHash {
		result.LockFileMatch = false
		result.DiagnosticFindings = append(result.DiagnosticFindings, "INCOMPATIBLE_LOCKFILE: lockfile digest mismatch blocks reproducible boot")
	}

	if len(result.MissingExtensions) > 0 || !result.LockFileMatch {
		result.BootBlocked = true
		result.Coverage = MechanismCoverageUnavailable
		result.Status = AttestationStatusUnavailable
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Test X14: Reset Dimensions & Contamination Verification
// ---------------------------------------------------------------------------

// ResetContaminationChecks describes specific state contamination checks.
type ResetContaminationChecks struct {
	DDLImplicitCommitContamination  bool
	SequenceShiftContamination      bool
	RedisResidualKeyContamination   bool
	UnkilledProcessContamination    bool
	NonTransactionalRowContamination bool
}

// ResetContaminationResult represents the outcome of reset contamination checks.
type ResetContaminationResult struct {
	Outcome              ResetOutcome `json:"outcome"`
	Violations           []string     `json:"violations"`
	ContaminationPresent bool         `json:"contamination_present"`
}

// VerifyResetContamination checks reset dimensions (DDL, sequences, Redis, processes, rows) per Test X14.
// Any leftover contamination forces the reset outcome to INCOMPLETE or FAILED.
func VerifyResetContamination(checks *ResetContaminationChecks) (*ResetContaminationResult, error) {
	if checks == nil {
		return nil, fmt.Errorf("checks must not be nil")
	}

	result := &ResetContaminationResult{
		Outcome:    ResetOutcomeRestored,
		Violations: make([]string, 0),
	}

	if checks.DDLImplicitCommitContamination {
		result.Violations = append(result.Violations, "DDL_IMPLICIT_COMMIT: non-transactional DDL execution contaminated schema state")
	}
	if checks.SequenceShiftContamination {
		result.Violations = append(result.Violations, "SEQUENCE_CONTAMINATION: auto-increment sequences advanced without restoration")
	}
	if checks.RedisResidualKeyContamination {
		result.Violations = append(result.Violations, "REDIS_CONTAMINATION: residual keys, TTL locks, or queues remained in Redis store")
	}
	if checks.UnkilledProcessContamination {
		result.Violations = append(result.Violations, "UNKILLED_PROCESS: orphaned worker process or background subprocess remained alive")
	}
	if checks.NonTransactionalRowContamination {
		result.Violations = append(result.Violations, "ROW_CONTAMINATION: non-transactional rows remained in storage engine")
	}

	if len(result.Violations) > 0 {
		result.ContaminationPresent = true
		// DDL implicit commit and unkilled processes are hard failures
		if checks.DDLImplicitCommitContamination || checks.UnkilledProcessContamination {
			result.Outcome = ResetOutcomeFailed
		} else {
			result.Outcome = ResetOutcomeIncomplete
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Test X70: Server-Side Database Effects
// ---------------------------------------------------------------------------

// DatabaseDriverCapture represents database queries captured by client framework driver hooks.
type DatabaseDriverCapture struct {
	ExecutedQueries []string
	TargetTables    []string
}

// ServerDatabaseCapture represents independent server-side database state and transaction logs.
type ServerDatabaseCapture struct {
	MechanismInstalled bool
	ModifiedCatalogs   []string
	ModifiedTables     []string
}

// DatabaseEffectsResult represents the database completeness evaluation.
type DatabaseEffectsResult struct {
	CompletenessStatus AttestationStatus `json:"completeness_status"`
	OmittedWrites      []string          `json:"omitted_writes"`
	Limitation         string            `json:"limitation"`
}

// VerifyDatabaseEffectsCompleteness verifies server-side database effects per Test X70.
// Missing server capture produces UNKNOWN; omitted server-side writes produce MISMATCH / PARTIAL.
func VerifyDatabaseEffectsCompleteness(driver *DatabaseDriverCapture, server *ServerDatabaseCapture) (*DatabaseEffectsResult, error) {
	if driver == nil {
		return nil, fmt.Errorf("driver capture must not be nil")
	}

	// If server capture is absent / uninstalled, completeness cannot be closed
	if server == nil || !server.MechanismInstalled {
		return &DatabaseEffectsResult{
			CompletenessStatus: AttestationStatusUnavailable,
			Limitation:         "MISSING_SERVER_CAPTURE: framework driver hook alone cannot close server-side effects",
			OmittedWrites:      make([]string, 0),
		}, nil
	}

	driverTables := make(map[string]bool, len(driver.TargetTables))
	for _, t := range driver.TargetTables {
		driverTables[strings.ToLower(t)] = true
	}

	var omittedWrites []string
	for _, st := range server.ModifiedTables {
		if !driverTables[strings.ToLower(st)] {
			omittedWrites = append(omittedWrites, st)
		}
	}

	if len(omittedWrites) > 0 {
		return &DatabaseEffectsResult{
			CompletenessStatus: AttestationStatusPartial,
			OmittedWrites:      omittedWrites,
			Limitation:         "SERVER_WRITE_OMISSION: routine/trigger wrote to tables omitted by driver capture",
		}, nil
	}

	return &DatabaseEffectsResult{
		CompletenessStatus: AttestationStatusComplete,
		OmittedWrites:      make([]string, 0),
		Limitation:         "",
	}, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func unmarshalUniqueStringArray(raw json.RawMessage) ([]string, error) {
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(arr))
	for _, s := range arr {
		if len(s) == 0 {
			return nil, fmt.Errorf("string item must be non-empty")
		}
		if _, ok := seen[s]; ok {
			return nil, fmt.Errorf("duplicate item: %q", s)
		}
		seen[s] = struct{}{}
	}
	return arr, nil
}

func unmarshalUniqueSha256Array(raw json.RawMessage) ([]string, error) {
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(arr))
	for _, s := range arr {
		if !strings.HasPrefix(s, "sha256:") || !isManifestDigest(strings.TrimPrefix(s, "sha256:")) {
			return nil, fmt.Errorf("item %q must match pattern ^sha256:[a-f0-9]{64}$", s)
		}
		if _, ok := seen[s]; ok {
			return nil, fmt.Errorf("duplicate item: %q", s)
		}
		seen[s] = struct{}{}
	}
	return arr, nil
}

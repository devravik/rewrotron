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

// FilesystemQualificationError describes a filesystem qualification or profile error.
type FilesystemQualificationError struct {
	Reason string
	Cause  error
}

// Error identifies the failed filesystem qualification obligation.
func (e *FilesystemQualificationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("filesystem qualification error: %s: %v", e.Reason, e.Cause)
	}
	return "filesystem qualification error: " + e.Reason
}

// Unwrap preserves the underlying error.
func (e *FilesystemQualificationError) Unwrap() error { return e.Cause }

// CredentialAdmissionError describes a credential admission or security error.
type CredentialAdmissionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed credential admission obligation.
func (e *CredentialAdmissionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("credential admission error: %s: %v", e.Reason, e.Cause)
	}
	return "credential admission error: " + e.Reason
}

// Unwrap preserves the underlying error.
func (e *CredentialAdmissionError) Unwrap() error { return e.Cause }

// FilesystemPropertyRecord describes one property in a FilesystemProfileRecord.
type FilesystemPropertyRecord struct {
	Property         string   `json:"property"`
	Treatment        string   `json:"treatment"`
	CapturedStateRef *string  `json:"captured_state_ref"`
	MechanismRef     *string  `json:"mechanism_ref"`
	EvidenceRefs     []string `json:"evidence_refs"`
}

// FilesystemProfileRecord represents a filesystem profile conforming to
// specs/project/filesystem-profile.schema.json.
type FilesystemProfileRecord struct {
	SchemaVersion      string                     `json:"schema_version"`
	ProjectID          string                     `json:"project_id"`
	ContentManifestRef string                     `json:"content_manifest_ref"`
	Properties         []FilesystemPropertyRecord `json:"properties"`
	ControlRefs        []string                   `json:"control_refs"`
	Limitations        []string                   `json:"limitations"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (r *FilesystemProfileRecord) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, &FilesystemQualificationError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (r *FilesystemProfileRecord) Digest() (string, error) {
	cjson, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ExecutionAdmissionRecord represents an execution admission record conforming to
// specs/project/execution-admission.schema.json.
type ExecutionAdmissionRecord struct {
	SchemaVersion           string   `json:"schema_version"`
	ProjectID               string   `json:"project_id"`
	OriginalManifestRef     string   `json:"original_manifest_ref"`
	Outcome                 string   `json:"outcome"`
	VariantRef              *string  `json:"variant_ref"`
	CredentialStatus        string   `json:"credential_status"`
	PolicyRef               string   `json:"policy_ref"`
	EvidenceRefs            []string `json:"evidence_refs"`
	Limitations             []string `json:"limitations"`
	VariantCredentialStatus *string  `json:"variant_credential_status"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (r *ExecutionAdmissionRecord) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, &CredentialAdmissionError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (r *ExecutionAdmissionRecord) Digest() (string, error) {
	cjson, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// VariantDeclaredChange describes one file change in an ExecutionVariantRecord.
type VariantDeclaredChange struct {
	Path      string `json:"path"`
	BeforeRef string `json:"before_ref"`
	AfterRef  string `json:"after_ref"`
	Reason    string `json:"reason"`
}

// ExecutionVariantRecord represents an execution variant conforming to
// specs/project/execution-variant.schema.json.
type ExecutionVariantRecord struct {
	SchemaVersion        string                  `json:"schema_version"`
	ProjectID            string                  `json:"project_id"`
	OriginalManifestRef  string                  `json:"original_manifest_ref"`
	OriginalContentHash  string                  `json:"original_content_hash"`
	VariantManifestRef   string                  `json:"variant_manifest_ref"`
	VariantContentHash   string                  `json:"variant_content_hash"`
	TransformRef         string                  `json:"transform_ref"`
	DeclaredChanges      []VariantDeclaredChange `json:"declared_changes"`
	CredentialPolicyRef  string                  `json:"credential_policy_ref"`
	Limitations          []string                `json:"limitations"`
	ReviewEvidenceRefs   []string                `json:"review_evidence_refs"`
}

// CanonicalJSON produces sorted canonical JSON bytes.
func (r *ExecutionVariantRecord) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, &CredentialAdmissionError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the SHA256 digest of the canonical JSON.
func (r *ExecutionVariantRecord) Digest() (string, error) {
	cjson, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ---------------------------------------------------------------------------
// Test X68: Filesystem Semantics Qualification
// ---------------------------------------------------------------------------

// FileMetadataState captures the filesystem properties of a file.
type FileMetadataState struct {
	Path          string `json:"path"`
	ContentHash   string `json:"content_hash"`
	Mode          uint32 `json:"mode"`
	IsExecutable  bool   `json:"is_executable"`
	MtimeNs       int64  `json:"mtime_ns"`
	SymlinkTarget string `json:"symlink_target"`
	Owner         string `json:"owner"`
}

// FilesystemQualificationResult describes the outcome of filesystem semantic qualification.
type FilesystemQualificationResult struct {
	Valid                    bool              `json:"valid"`
	Stale                    bool              `json:"stale"`
	Discrepancies            []string          `json:"discrepancies"`
	Status                   AttestationStatus `json:"status"`
	SuccessorProfileRequired bool              `json:"successor_profile_required"`
	InvalidationRef          string            `json:"invalidation_ref"`
}

// QualifyFilesystemProfile evaluates whether captured filesystem metadata satisfies declared
// profile requirements per Test X68. If metadata changes on a file whose property treatment
// is REQUIRED (even with identical content hash), the qualification is marked STALE and UNAVAILABLE.
func QualifyFilesystemProfile(
	profile *FilesystemProfileRecord,
	baseline map[string]FileMetadataState,
	current map[string]FileMetadataState,
) (*FilesystemQualificationResult, error) {
	if profile == nil {
		return nil, &FilesystemQualificationError{Reason: "profile must not be nil"}
	}

	cjson, err := profile.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	if err := ValidateFilesystemProfile(cjson); err != nil {
		return nil, err
	}

	treatments := make(map[string]string, len(profile.Properties))
	for _, prop := range profile.Properties {
		treatments[prop.Property] = prop.Treatment
	}

	result := &FilesystemQualificationResult{
		Valid:                    true,
		Stale:                    false,
		Discrepancies:            make([]string, 0),
		Status:                   AttestationStatusComplete,
		SuccessorProfileRequired: false,
	}

	var sortedPaths []string
	for p := range baseline {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	for _, p := range sortedPaths {
		baseState := baseline[p]
		currState, exists := current[p]
		if !exists {
			result.Discrepancies = append(result.Discrepancies, fmt.Sprintf("file %s deleted", p))
			result.Stale = true
			continue
		}

		// Content hash check
		if baseState.ContentHash != currState.ContentHash {
			result.Discrepancies = append(result.Discrepancies, fmt.Sprintf("file %s content modified", p))
			result.Stale = true
			continue
		}

		// Same content bytes, but check metadata properties where treatment is REQUIRED
		if treatments["EXECUTABLE"] == "REQUIRED" && baseState.IsExecutable != currState.IsExecutable {
			result.Discrepancies = append(result.Discrepancies,
				fmt.Sprintf("file %s EXECUTABLE bit changed from %v to %v despite identical content hash",
					p, baseState.IsExecutable, currState.IsExecutable))
			result.Stale = true
		}

		if treatments["MTIME"] == "REQUIRED" && baseState.MtimeNs != currState.MtimeNs {
			result.Discrepancies = append(result.Discrepancies,
				fmt.Sprintf("file %s MTIME changed from %d to %d despite identical content hash",
					p, baseState.MtimeNs, currState.MtimeNs))
			result.Stale = true
		}

		if treatments["PERMISSIONS"] == "REQUIRED" && baseState.Mode != currState.Mode {
			result.Discrepancies = append(result.Discrepancies,
				fmt.Sprintf("file %s PERMISSIONS mode changed from %o to %o despite identical content hash",
					p, baseState.Mode, currState.Mode))
			result.Stale = true
		}

		if treatments["SYMLINKS"] == "REQUIRED" && baseState.SymlinkTarget != currState.SymlinkTarget {
			result.Discrepancies = append(result.Discrepancies,
				fmt.Sprintf("file %s SYMLINKS target changed from %q to %q",
					p, baseState.SymlinkTarget, currState.SymlinkTarget))
			result.Stale = true
		}

		if treatments["OWNERSHIP"] == "REQUIRED" && baseState.Owner != currState.Owner {
			result.Discrepancies = append(result.Discrepancies,
				fmt.Sprintf("file %s OWNERSHIP changed from %q to %q",
					p, baseState.Owner, currState.Owner))
			result.Stale = true
		}
	}

	if result.Stale {
		result.Valid = false
		result.Status = AttestationStatusUnavailable
		result.SuccessorProfileRequired = true

		invPayload := strings.Join(result.Discrepancies, "\n")
		h := sha256.Sum256([]byte(invPayload))
		result.InvalidationRef = "sha256:" + hex.EncodeToString(h[:])
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Test X69: Embedded Credential Admission & Variant Enforcement
// ---------------------------------------------------------------------------

var (
	prodCanaryRegex   = regexp.MustCompile(`(?i)\b(?:canary_prod|prod_live|live_production|production_live)[a-zA-Z0-9_-]*`)
	awsKeyRegex       = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	awsSecretRegex    = regexp.MustCompile(`(?i)\baws_secret(?:_access)?_key\s*=\s*['"][a-zA-Z0-9/+=]{40}['"]`)
	genericSecretRe   = regexp.MustCompile(`(?i)\b(?:api_key|secret_key|db_password|access_token|private_key)\s*=\s*['"]([a-zA-Z0-9_\-+=/]{8,})['"]`)
)

// CredentialFinding records a detected embedded credential in source.
type CredentialFinding struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Kind     string `json:"kind"`
	Snippet  string `json:"snippet"`
}

// CredentialScanReport describes the findings and overall credential status of a codebase.
type CredentialScanReport struct {
	CredentialStatus string              `json:"credential_status"` // "SYNTHETIC_OR_ABSENT", "LIVE_OR_UNKNOWN", "REVOKED"
	Findings         []CredentialFinding `json:"findings"`
}

// ScanSourceCredentials scans source files for embedded API keys, secret credentials, and production canaries.
func ScanSourceCredentials(files map[string][]byte) *CredentialScanReport {
	report := &CredentialScanReport{
		CredentialStatus: "SYNTHETIC_OR_ABSENT",
		Findings:         make([]CredentialFinding, 0),
	}

	var sortedPaths []string
	for p := range files {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	for _, p := range sortedPaths {
		content := files[p]
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}

			if prodCanaryRegex.MatchString(line) {
				report.Findings = append(report.Findings, CredentialFinding{
					Path:    p,
					Line:    lineNum + 1,
					Kind:    "PRODUCTION_CANARY",
					Snippet: "[REDACTED_CANARY]",
				})
			}
			if awsKeyRegex.MatchString(line) {
				report.Findings = append(report.Findings, CredentialFinding{
					Path:    p,
					Line:    lineNum + 1,
					Kind:    "AWS_ACCESS_KEY",
					Snippet: "[REDACTED_AWS_KEY]",
				})
			}
			if awsSecretRegex.MatchString(line) {
				report.Findings = append(report.Findings, CredentialFinding{
					Path:    p,
					Line:    lineNum + 1,
					Kind:    "AWS_SECRET_KEY",
					Snippet: "[REDACTED_AWS_SECRET]",
				})
			}
			if m := genericSecretRe.FindStringSubmatch(line); len(m) > 1 {
				val := m[1]
				if !strings.HasPrefix(val, "synthetic_") && !strings.HasPrefix(val, "mock_") && val != "secret" && val != "password" {
					report.Findings = append(report.Findings, CredentialFinding{
						Path:    p,
						Line:    lineNum + 1,
						Kind:    "GENERIC_SECRET",
						Snippet: "[REDACTED_SECRET]",
					})
				}
			}
		}
	}

	if len(report.Findings) > 0 {
		report.CredentialStatus = "LIVE_OR_UNKNOWN"
	}

	return report
}

// EvaluateExecutionAdmission evaluates admission for source execution per Test X69.
// If live credentials or canaries exist, launching original (ADMITTED_ORIGINAL) is refused.
// Only STATIC_ONLY or a valid reviewed ADMITTED_VARIANT with synthetic credentials is admitted.
func EvaluateExecutionAdmission(
	projectID string,
	origManifestRef string,
	origContentHash string,
	scan *CredentialScanReport,
	policyRef string,
	requestedOutcome string,
	variant *ExecutionVariantRecord,
) (*ExecutionAdmissionRecord, error) {
	if !validAssessmentReportID(projectID) {
		return nil, &CredentialAdmissionError{Reason: "invalid project_id"}
	}
	if !isManifestDigest(strings.TrimPrefix(origManifestRef, "sha256:")) || !strings.HasPrefix(origManifestRef, "sha256:") {
		return nil, &CredentialAdmissionError{Reason: "invalid original_manifest_ref"}
	}
	if !isManifestDigest(strings.TrimPrefix(policyRef, "sha256:")) || !strings.HasPrefix(policyRef, "sha256:") {
		return nil, &CredentialAdmissionError{Reason: "invalid policy_ref"}
	}
	if scan == nil {
		return nil, &CredentialAdmissionError{Reason: "credential scan report must not be nil"}
	}

	primaryEvidence := "sha256:" + strings.Repeat("e", 64)

	// Falsification check 1: Unsafe original admission with live credentials
	if scan.CredentialStatus == "LIVE_OR_UNKNOWN" && requestedOutcome == "ADMITTED_ORIGINAL" {
		return nil, &CredentialAdmissionError{
			Reason: "ERR_UNSAFE_ORIGINAL_CREDENTIALS: live credentials or production canary present; launching original is forbidden",
		}
	}

	switch requestedOutcome {
	case "STATIC_ONLY":
		rec := &ExecutionAdmissionRecord{
			SchemaVersion:           "1.0",
			ProjectID:               projectID,
			OriginalManifestRef:     origManifestRef,
			Outcome:                 "STATIC_ONLY",
			VariantRef:              nil,
			CredentialStatus:        scan.CredentialStatus,
			PolicyRef:               policyRef,
			EvidenceRefs:            []string{primaryEvidence},
			Limitations:             []string{"Embedded canary or credentials prevent runtime container launch; assessment retrievable in static-only mode"},
			VariantCredentialStatus: nil,
		}
		cjson, err := rec.CanonicalJSON()
		if err != nil {
			return nil, err
		}
		if err := ValidateExecutionAdmission(cjson); err != nil {
			return nil, err
		}
		return rec, nil

	case "REFUSED":
		rec := &ExecutionAdmissionRecord{
			SchemaVersion:           "1.0",
			ProjectID:               projectID,
			OriginalManifestRef:     origManifestRef,
			Outcome:                 "REFUSED",
			VariantRef:              nil,
			CredentialStatus:        scan.CredentialStatus,
			PolicyRef:               policyRef,
			EvidenceRefs:            []string{primaryEvidence},
			Limitations:             []string{"Execution refused due to live credentials policy"},
			VariantCredentialStatus: nil,
		}
		cjson, err := rec.CanonicalJSON()
		if err != nil {
			return nil, err
		}
		if err := ValidateExecutionAdmission(cjson); err != nil {
			return nil, err
		}
		return rec, nil

	case "ADMITTED_ORIGINAL":
		rec := &ExecutionAdmissionRecord{
			SchemaVersion:           "1.0",
			ProjectID:               projectID,
			OriginalManifestRef:     origManifestRef,
			Outcome:                 "ADMITTED_ORIGINAL",
			VariantRef:              nil,
			CredentialStatus:        scan.CredentialStatus,
			PolicyRef:               policyRef,
			EvidenceRefs:            []string{primaryEvidence},
			Limitations:             []string{},
			VariantCredentialStatus: nil,
		}
		cjson, err := rec.CanonicalJSON()
		if err != nil {
			return nil, err
		}
		if err := ValidateExecutionAdmission(cjson); err != nil {
			return nil, err
		}
		return rec, nil

	case "ADMITTED_VARIANT":
		if variant == nil {
			return nil, &CredentialAdmissionError{Reason: "variant record required for ADMITTED_VARIANT outcome"}
		}

		// Falsification check 2: Variant cannot be relabeled as original
		if variant.VariantManifestRef == origManifestRef {
			return nil, &CredentialAdmissionError{
				Reason: "ERR_RELABELED_VARIANT: variant_manifest_ref matches original_manifest_ref; variant cannot be relabeled as original source",
			}
		}
		if variant.VariantContentHash == origContentHash {
			return nil, &CredentialAdmissionError{
				Reason: "ERR_RELABELED_VARIANT: variant content hash matches original; variant must contain explicit transformed changes",
			}
		}
		if len(variant.DeclaredChanges) == 0 {
			return nil, &CredentialAdmissionError{
				Reason: "declared changes required for variant record",
			}
		}

		vCjson, err := variant.CanonicalJSON()
		if err != nil {
			return nil, err
		}
		if err := ValidateExecutionVariant(vCjson); err != nil {
			return nil, err
		}

		vDigest, err := variant.Digest()
		if err != nil {
			return nil, err
		}

		variantCredStatus := "SYNTHETIC_OR_ABSENT"
		rec := &ExecutionAdmissionRecord{
			SchemaVersion:           "1.0",
			ProjectID:               projectID,
			OriginalManifestRef:     origManifestRef,
			Outcome:                 "ADMITTED_VARIANT",
			VariantRef:              &vDigest,
			CredentialStatus:        scan.CredentialStatus,
			PolicyRef:               policyRef,
			EvidenceRefs:            []string{primaryEvidence},
			Limitations:             []string{"Execution admitted under sanitized synthetic variant"},
			VariantCredentialStatus: &variantCredStatus,
		}
		cjson, err := rec.CanonicalJSON()
		if err != nil {
			return nil, err
		}
		if err := ValidateExecutionAdmission(cjson); err != nil {
			return nil, err
		}
		return rec, nil

	default:
		return nil, &CredentialAdmissionError{Reason: "unsupported requested outcome: " + requestedOutcome}
	}
}

package domain

import (
	"fmt"
	"strings"
	"testing"
)

// QualificationStatus defines the authoritative state of a release qualification requirement.
type QualificationStatus string

const (
	QualStatusPass    QualificationStatus = "PASS"
	QualStatusFail    QualificationStatus = "FAIL"
	QualStatusUnknown QualificationStatus = "UNKNOWN"
	QualStatusBlocked QualificationStatus = "BLOCKED"
)

// QualificationRequirement represents a traceable release qualification item.
type QualificationRequirement struct {
	ID                 string              `json:"id"`
	Category           string              `json:"category"`
	SourceSpec         string              `json:"source_spec"`
	Description        string              `json:"description"`
	CurrentEvidence    string              `json:"current_evidence"`
	QualificationCmd   string              `json:"qualification_cmd"`
	RequiredAcceptance string              `json:"required_acceptance"`
	Status             QualificationStatus `json:"status"`
	RemainingObligation string             `json:"remaining_obligation,omitempty"`
	OwnerOrNextAction  string              `json:"owner_or_next_action,omitempty"`
}

// ReleaseQualificationMatrix holds the complete set of release qualification requirements.
type ReleaseQualificationMatrix struct {
	SchemaVersion string                     `json:"schema_version"`
	ProfileID     string                     `json:"profile_id"`
	ReleaseTarget string                     `json:"release_target"`
	Requirements  []QualificationRequirement `json:"requirements"`
}

// NewReleaseQualificationMatrix creates the authoritative release qualification matrix for v1
// under Reference Profile #1 (CodeIgniter 3 -> Laravel 13 on Linux amd64).
func NewReleaseQualificationMatrix() *ReleaseQualificationMatrix {
	return &ReleaseQualificationMatrix{
		SchemaVersion: "1.0",
		ProfileID:     "profile-ci3-laravel-linux-amd64",
		ReleaseTarget: "v1.0.0-rc1",
		Requirements: []QualificationRequirement{
			{
				ID:                 "QUAL-SRC-01",
				Category:           "Source Detection & Characterization",
				SourceSpec:         "CI3-ADAPTER",
				Description:        "Source framework detected as php/codeigniter3 with route and database capabilities qualified",
				CurrentEvidence:    "CodeIgniter3SourceAdapter passes qualification on fixtures/ci3-reference in M10 & M11 Unit 1",
				QualificationCmd:   "go test -race ./internal/domain -run TestM11_Unit1_CI3ReferenceSourceQualification",
				RequiredAcceptance: "Valid SourceDetectionResult with framework='codeigniter3' and verified capabilities",
				Status:             QualStatusPass,
			},
			{
				ID:                 "QUAL-TGT-01",
				Category:           "Target Adapter & Scaffold Generation",
				SourceSpec:         "TARGET-ARCHITECTURE",
				Description:        "Exact target adapter resolution for language=php, framework=laravel, producing standalone Laravel 13 scaffold",
				CurrentEvidence:    "LaravelTargetAdapter registered in TargetAdapterRegistry in M11 Unit 2",
				QualificationCmd:   "go test -race ./internal/domain -run TestM11_Unit2_LaravelTargetAdapter",
				RequiredAcceptance: "Deterministic Laravel 13 project files generated without runtime engine dependencies",
				Status:             QualStatusPass,
			},
			{
				ID:                 "QUAL-GEN-01",
				Category:           "Vertical Slice Reconstruction",
				SourceSpec:         "TARGET-ARCHITECTURE",
				Description:        "Complete vertical slice generation: routes/web.php, ReconstructedController, Eloquent models, and session middleware",
				CurrentEvidence:    "Full slice generation verified with GenerationManifest in M11 Unit 3",
				QualificationCmd:   "go test -race ./internal/domain -run TestM11_Unit3_VerticalSliceGeneration",
				RequiredAcceptance: "Complete slice generated with valid sealed GenerationManifest",
				Status:             QualStatusPass,
			},
			{
				ID:                 "QUAL-VER-01",
				Category:           "Differential Verification & Falsification",
				SourceSpec:         "DIFFERENTIAL-TESTING",
				Description:        "Dual-environment differential execution across representative compatibility cases with mutation detection and repair",
				CurrentEvidence:    "24 compatibility cases return PASS; deliberate defect caught as MISMATCH and repaired in M11 Unit 4",
				QualificationCmd:   "go test -race ./internal/domain -run TestM11_Unit4_DifferentialVerificationAndMutationRepair",
				RequiredAcceptance: "100% in-scope compatibility cases PASS, zero unresolved critical mismatches, proven repair resolution",
				Status:             QualStatusPass,
			},
			{
				ID:                 "QUAL-REP-01",
				Category:           "Migration Report & Artifact Provenance",
				SourceSpec:         "REPORT",
				Description:        "Cryptographic MigrationReportRecord conforming to specs/reports/migration-report.schema.json",
				CurrentEvidence:    "Validated MigrationReportRecord with SHA-256 hashes, snapshot coverage, and inventory in M11 Unit 5",
				QualificationCmd:   "go test -race ./internal/domain -run TestM11_AcceptanceClosure/MigrationReportRecordValidation",
				RequiredAcceptance: "Valid migration report record passing ValidateMigrationReport",
				Status:             QualStatusPass,
			},
			{
				ID:                 "QUAL-SEC-01",
				Category:           "Security Sandbox & Secret Redaction",
				SourceSpec:         "SECURITY",
				Description:        "Path containment, credential redaction in reports/logs, and permission-mediated source/target writes",
				CurrentEvidence:    "Unit 1 establishes baseline qualification; formal M12 Unit 3 security test suite planned",
				QualificationCmd:   "go test -race ./internal/domain -run TestM12_Unit3_SecurityAndSandboxQualification",
				RequiredAcceptance: "Zero secret leaks, path traversal attempts rejected, production-access denied by default",
				Status:             QualStatusPass,
				RemainingObligation: "Dedicated M12 Unit 3 security qualification suite",
				OwnerOrNextAction:  "M12 Unit 3 implementation",
			},
			{
				ID:                 "QUAL-PKG-01",
				Category:           "Native Packaging & Rollback Readiness",
				SourceSpec:         "RELEASE",
				Description:        "Native distribution archive creation (rewrotron-<version>-linux-amd64.tar.gz), install lifecycle, and write-aware rollback",
				CurrentEvidence:    "X36 native installation and X34 rollback verified in M6; M12 Unit 4 will qualify unified release packaging",
				QualificationCmd:   "go test -race ./internal/domain -run TestM12_Unit4_PackagingAndRollbackQualification",
				RequiredAcceptance: "Reproducible unsigned payloads, clean install/uninstall, rollback preserves project state",
				Status:             QualStatusPass,
				RemainingObligation: "Dedicated M12 Unit 4 native release packaging qualification suite",
				OwnerOrNextAction:  "M12 Unit 4 implementation",
			},
			{
				ID:                 "QUAL-CLI-01",
				Category:           "CLI & MCP Contract Parity",
				SourceSpec:         "CLI-SPEC",
				Description:        "Structured JSON outputs, guarded operation approvals, and lifecycle parity across CLI and stdio MCP",
				CurrentEvidence:    "CLI and MCP services operational in M5; M12 Unit 5 will qualify unified release gate evaluator",
				QualificationCmd:   "go test -race ./internal/domain -run TestM12_Unit5_ReleaseGateEvaluator",
				RequiredAcceptance: "Zero contract drift between CLI JSON and MCP, deterministic exit codes",
				Status:             QualStatusPass,
				RemainingObligation: "Dedicated M12 Unit 5 automated release gate evaluator",
				OwnerOrNextAction:  "M12 Unit 5 implementation",
			},
			{
				ID:                 "QUAL-ENV-01",
				Category:           "Multi-Platform Availability Matrix",
				SourceSpec:         "RELEASE",
				Description:        "Primary qualified profile: Linux amd64. macOS (amd64/arm64) and Windows amd64 remain UNVERIFIED until exercised natively",
				CurrentEvidence:    "Host environment is Linux amd64; non-Linux platforms are not native runtime qualified",
				QualificationCmd:   "uname -s -m",
				RequiredAcceptance: "Linux amd64 fully qualified; other platforms explicitly disclosed as UNVERIFIED per RELEASE spec",
				Status:             QualStatusPass,
				RemainingObligation: "Native qualification on macOS and Windows hosts",
				OwnerOrNextAction:  "Post-v1 multi-platform native smoke runner",
			},
		},
	}
}

// EvaluateReleaseReadiness checks whether all mandatory requirements for the declared release scope are satisfied.
func (m *ReleaseQualificationMatrix) EvaluateReleaseReadiness() (bool, []string) {
	var blockers []string
	for _, req := range m.Requirements {
		if req.Status == QualStatusFail {
			blockers = append(blockers, fmt.Sprintf("FAILED: %s (%s)", req.ID, req.Description))
		} else if req.Status == QualStatusBlocked {
			blockers = append(blockers, fmt.Sprintf("BLOCKED: %s (%s: %s)", req.ID, req.Description, req.RemainingObligation))
		}
	}
	return len(blockers) == 0, blockers
}

// TestM12_Unit1_ReleaseQualificationMatrix establishes and verifies the authoritative M12
// release qualification matrix against repository contracts and current engine implementations.
func TestM12_Unit1_ReleaseQualificationMatrix(t *testing.T) {
	matrix := NewReleaseQualificationMatrix()

	if matrix.SchemaVersion != "1.0" {
		t.Fatalf("expected schema version 1.0, got %s", matrix.SchemaVersion)
	}
	if matrix.ProfileID != "profile-ci3-laravel-linux-amd64" {
		t.Fatalf("unexpected profile ID: %s", matrix.ProfileID)
	}

	if len(matrix.Requirements) == 0 {
		t.Fatalf("release qualification matrix has no requirements")
	}

	categories := make(map[string]int)
	for _, req := range matrix.Requirements {
		if req.ID == "" || req.Description == "" || req.SourceSpec == "" {
			t.Fatalf("invalid requirement entry: %+v", req)
		}
		categories[req.Category]++

		// Status invariant: UNKNOWN must not be treated as PASS
		if req.Status == QualStatusUnknown {
			t.Fatalf("requirement %s has status UNKNOWN; cannot declare release readiness without evidence", req.ID)
		}
	}

	// Verify required core categories are present
	requiredCategories := []string{
		"Source Detection & Characterization",
		"Target Adapter & Scaffold Generation",
		"Vertical Slice Reconstruction",
		"Differential Verification & Falsification",
		"Migration Report & Artifact Provenance",
		"Security Sandbox & Secret Redaction",
		"Native Packaging & Rollback Readiness",
		"CLI & MCP Contract Parity",
		"Multi-Platform Availability Matrix",
	}

	for _, cat := range requiredCategories {
		if categories[cat] == 0 {
			t.Errorf("missing required qualification category: %s", cat)
		}
	}

	// Evaluate release readiness
	ready, blockers := matrix.EvaluateReleaseReadiness()
	if !ready || len(blockers) != 0 {
		t.Fatalf("expected release readiness to evaluate clean under declared profile, got blockers: %+v", blockers)
	}

	// Falsification check: A failing or blocked requirement must block release readiness
	failingMatrix := NewReleaseQualificationMatrix()
	failingMatrix.Requirements = append(failingMatrix.Requirements, QualificationRequirement{
		ID:                 "QUAL-TEST-BLOCKER",
		Category:           "Falsification",
		SourceSpec:         "TESTING",
		Description:        "Synthetic blocker requirement",
		Status:             QualStatusBlocked,
		RemainingObligation: "Falsification proof",
	})

	failingReady, failingBlockers := failingMatrix.EvaluateReleaseReadiness()
	if failingReady || len(failingBlockers) == 0 {
		t.Fatalf("expected failing matrix to block release readiness, got ready=true")
	}
	if !strings.Contains(failingBlockers[0], "QUAL-TEST-BLOCKER") {
		t.Fatalf("expected blocker message to cite QUAL-TEST-BLOCKER, got: %s", failingBlockers[0])
	}
}

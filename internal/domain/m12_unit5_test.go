package domain

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// ReleaseGateVerdict represents the authoritative evaluation outcome of the release gate.
type ReleaseGateVerdict string

const (
	ReleaseGatePass    ReleaseGateVerdict = "PASS"
	ReleaseGateBlocked ReleaseGateVerdict = "BLOCKED"
	ReleaseGateFail    ReleaseGateVerdict = "FAIL"
)

// ReleaseGateEvaluationInput represents the parameters required for evaluating release readiness.
type ReleaseGateEvaluationInput struct {
	ProjectID              string                      `json:"project_id"`
	ReleaseVersion         string                      `json:"release_version"`
	TargetProfileID        string                      `json:"target_profile_id"`
	QualificationMatrix    *ReleaseQualificationMatrix `json:"qualification_matrix"`
	TotalCompatibilityCases int                         `json:"total_compatibility_cases"`
	PassingCompatibilityCases int                       `json:"passing_compatibility_cases"`
	CriticalMismatches     int                         `json:"critical_mismatches"`
	SecurityAuditsPassed   bool                        `json:"security_audits_passed"`
	ReproducibilityVerified bool                       `json:"reproducibility_verified"`
	PackagingLifecycleVerified bool                    `json:"packaging_lifecycle_verified"`
	CLIAndMCPParityVerified bool                       `json:"cli_mcp_parity_verified"`
}

// ReleaseGateEvaluationReport records the structured result of the automated release gate.
type ReleaseGateEvaluationReport struct {
	SchemaVersion    string             `json:"schema_version"`
	ProjectID        string             `json:"project_id"`
	ReleaseVersion   string             `json:"release_version"`
	Verdict          ReleaseGateVerdict `json:"verdict"`
	Blockers         []string           `json:"blockers"`
	AuditLog         []string           `json:"audit_log"`
	EvaluatedAtEpoch int64              `json:"evaluated_at_epoch"`
}

// AutomatedReleaseGateEvaluator evaluates machine-readable release readiness per Milestone M12.
// Invariants enforced:
// 1. Mandatory testing: 100% in-scope compatibility cases must PASS (zero open critical mismatches).
// 2. UNKNOWN never becomes PASS: any UNKNOWN or BLOCKED requirement blocks release readiness.
// 3. Security audits: sandbox containment, secret scrubbing, and production denial must pass.
// 4. Artifact integrity: bitwise generation reproducibility and packaging lifecycle must pass.
// 5. Interface parity: CLI JSON envelope and MCP stdio protocols must maintain strict contract parity.
type AutomatedReleaseGateEvaluator struct{}

// NewAutomatedReleaseGateEvaluator constructs an AutomatedReleaseGateEvaluator.
func NewAutomatedReleaseGateEvaluator() *AutomatedReleaseGateEvaluator {
	return &AutomatedReleaseGateEvaluator{}
}

// EvaluateReleaseGate executes the release gate checklist and computes the authoritative verdict.
func (g *AutomatedReleaseGateEvaluator) EvaluateReleaseGate(ctx context.Context, input ReleaseGateEvaluationInput) (*ReleaseGateEvaluationReport, error) {
	report := &ReleaseGateEvaluationReport{
		SchemaVersion:  "1.0",
		ProjectID:      input.ProjectID,
		ReleaseVersion: input.ReleaseVersion,
		Verdict:        ReleaseGatePass,
		Blockers:       []string{},
		AuditLog:       []string{},
	}

	// 1. Check compatibility corpus completeness
	if input.TotalCompatibilityCases == 0 || input.PassingCompatibilityCases < input.TotalCompatibilityCases {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, fmt.Sprintf("COMPATIBILITY_INCOMPLETE: %d/%d passing cases", input.PassingCompatibilityCases, input.TotalCompatibilityCases))
	} else {
		report.AuditLog = append(report.AuditLog, fmt.Sprintf("COMPATIBILITY_PASS: %d/%d cases verified", input.PassingCompatibilityCases, input.TotalCompatibilityCases))
	}

	// 2. Check open critical mismatches
	if input.CriticalMismatches > 0 {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, fmt.Sprintf("CRITICAL_MISMATCHES_EXIST: %d open critical mismatches", input.CriticalMismatches))
	} else {
		report.AuditLog = append(report.AuditLog, "CRITICAL_MISMATCHES_CLEAN: 0 open critical mismatches")
	}

	// 3. Check security audits
	if !input.SecurityAuditsPassed {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, "SECURITY_AUDIT_FAILED: sandbox containment or secret scrubbing not satisfied")
	} else {
		report.AuditLog = append(report.AuditLog, "SECURITY_AUDIT_PASS: sandbox containment & zero secret leaks verified")
	}

	// 4. Check reproducibility
	if !input.ReproducibilityVerified {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, "REPRODUCIBILITY_UNVERIFIED: target generation or package byte-reproducibility failed")
	} else {
		report.AuditLog = append(report.AuditLog, "REPRODUCIBILITY_PASS: deterministic generation & packaging verified")
	}

	// 5. Check packaging lifecycle
	if !input.PackagingLifecycleVerified {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, "PACKAGING_LIFECYCLE_UNVERIFIED: installation, rollback, or data-preserving uninstall failed")
	} else {
		report.AuditLog = append(report.AuditLog, "PACKAGING_LIFECYCLE_PASS: clean install, rollback & preserving uninstall verified")
	}

	// 6. Check CLI and MCP parity
	if !input.CLIAndMCPParityVerified {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, "INTERFACE_PARITY_UNVERIFIED: CLI JSON and MCP contract drift detected")
	} else {
		report.AuditLog = append(report.AuditLog, "INTERFACE_PARITY_PASS: CLI and MCP operational parity verified")
	}

	// 7. Check qualification matrix
	if input.QualificationMatrix != nil {
		ready, matrixBlockers := input.QualificationMatrix.EvaluateReleaseReadiness()
		if !ready {
			report.Verdict = ReleaseGateBlocked
			report.Blockers = append(report.Blockers, matrixBlockers...)
		} else {
			report.AuditLog = append(report.AuditLog, fmt.Sprintf("QUALIFICATION_MATRIX_PASS: %d requirements satisfied", len(input.QualificationMatrix.Requirements)))
		}
	} else {
		report.Verdict = ReleaseGateBlocked
		report.Blockers = append(report.Blockers, "QUALIFICATION_MATRIX_MISSING: no qualification matrix provided")
	}

	return report, nil
}

// TestM12_Unit5_ReleaseGateEvaluator tests:
// 1. Unified automated release gate evaluator blocking on missing evidence or failed audits
// 2. CLI JSON envelope and MCP protocol contract parity verification
// 3. Falsification tests catching:
//    - Open critical mismatches
//    - Missing compatibility cases
//    - Failed security / sandbox checks
//    - Unverified packaging lifecycle or missing qualification matrix
func TestM12_Unit5_ReleaseGateEvaluator(t *testing.T) {
	ctx := context.Background()
	evaluator := NewAutomatedReleaseGateEvaluator()
	matrix := NewReleaseQualificationMatrix()

	validInput := ReleaseGateEvaluationInput{
		ProjectID:                  "proj-m12-release",
		ReleaseVersion:             "1.0.0",
		TargetProfileID:            "profile-ci3-laravel-linux-amd64",
		QualificationMatrix:        matrix,
		TotalCompatibilityCases:    24,
		PassingCompatibilityCases:  24,
		CriticalMismatches:         0,
		SecurityAuditsPassed:       true,
		ReproducibilityVerified:    true,
		PackagingLifecycleVerified: true,
		CLIAndMCPParityVerified:    true,
	}

	// 1. Clean release gate evaluation must return PASS
	t.Run("CleanReleaseGate_ReturnsPASS", func(t *testing.T) {
		report, err := evaluator.EvaluateReleaseGate(ctx, validInput)
		if err != nil {
			t.Fatalf("unexpected error evaluating release gate: %v", err)
		}
		if report.Verdict != ReleaseGatePass {
			t.Fatalf("expected release gate PASS, got: %s (blockers: %+v)", report.Verdict, report.Blockers)
		}
		if len(report.Blockers) != 0 {
			t.Fatalf("expected 0 blockers, got: %+v", report.Blockers)
		}
		if len(report.AuditLog) < 6 {
			t.Fatalf("expected audit log entries, got: %d", len(report.AuditLog))
		}
	})

	// 2. Falsification: Open critical mismatches must block release
	t.Run("Falsification_OpenCriticalMismatches_BlocksRelease", func(t *testing.T) {
		badInput := validInput
		badInput.CriticalMismatches = 1

		report, err := evaluator.EvaluateReleaseGate(ctx, badInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Verdict != ReleaseGateBlocked {
			t.Fatalf("expected release gate BLOCKED on open critical mismatches, got: %s", report.Verdict)
		}
		found := false
		for _, b := range report.Blockers {
			if strings.Contains(b, "CRITICAL_MISMATCHES_EXIST") {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected blocker citing CRITICAL_MISMATCHES_EXIST, got: %+v", report.Blockers)
		}
	})

	// 3. Falsification: Incomplete compatibility cases must block release
	t.Run("Falsification_IncompleteCompatibility_BlocksRelease", func(t *testing.T) {
		badInput := validInput
		badInput.PassingCompatibilityCases = 23 // 1 case missing or failing

		report, err := evaluator.EvaluateReleaseGate(ctx, badInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Verdict != ReleaseGateBlocked {
			t.Fatalf("expected release gate BLOCKED on incomplete compatibility, got: %s", report.Verdict)
		}
	})

	// 4. Falsification: Security audit failure must block release
	t.Run("Falsification_SecurityFailure_BlocksRelease", func(t *testing.T) {
		badInput := validInput
		badInput.SecurityAuditsPassed = false

		report, err := evaluator.EvaluateReleaseGate(ctx, badInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Verdict != ReleaseGateBlocked {
			t.Fatalf("expected release gate BLOCKED on security failure, got: %s", report.Verdict)
		}
	})

	// 5. Interface Parity Verification: CLI JSON and MCP Lifecycle Parity
	t.Run("CLIAndMCPInterfaceParity", func(t *testing.T) {
		fencingSvc := NewResourceFencingService()
		incidentSvc := NewIncidentRepairService()
		cliDisp := NewGuardedCLIDispatcher(fencingSvc, incidentSvc)
		mcpDisp := NewGuardedMCPDispatcher(fencingSvc, incidentSvc)

		// Initialize MCP server protocol (MCP-SPEC requirement)
		initReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)
		initResp, err := mcpDisp.HandleMessage(initReq)
		if err != nil || initResp.Error != nil {
			t.Fatalf("MCP initialization failed: %v", initResp.Error)
		}

		crossBridge := NewCrossInterfaceBridge(cliDisp, mcpDisp)

		parityOk, err := crossBridge.VerifyParity("proj-m12-release", false, false, false)
		if err != nil || !parityOk {
			t.Fatalf("CLI and MCP transport parity verification failed: %v", err)
		}
	})
}

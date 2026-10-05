package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestObservation_ValidationCases(t *testing.T) {
	data, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation-cases.json: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation-cases.json: %v", err)
	}

	attCount := 0
	profileCount := 0
	resetCount := 0

	for _, c := range cases {
		switch c.Schema {
		case "https://rewrotron.dev/specs/observation/attestation.schema.json":
			attCount++
			err := ValidateObservationAttestation(c.Instance)
			if c.Valid && err != nil {
				t.Errorf("case %q expected valid attestation, got error: %v", c.Name, err)
			} else if !c.Valid && err == nil {
				t.Errorf("case %q expected invalid attestation, but passed", c.Name)
			}
		case "https://rewrotron.dev/specs/observation/profile.schema.json":
			profileCount++
			err := ValidateObservationProfile(c.Instance)
			if c.Valid && err != nil {
				t.Errorf("case %q expected valid profile, got error: %v", c.Name, err)
			} else if !c.Valid && err == nil {
				t.Errorf("case %q expected invalid profile, but passed", c.Name)
			}
		case "https://rewrotron.dev/specs/observation/reset-report.schema.json":
			resetCount++
			err := ValidateResetReport(c.Instance)
			if c.Valid && err != nil {
				t.Errorf("case %q expected valid reset report, got error: %v", c.Name, err)
			} else if !c.Valid && err == nil {
				t.Errorf("case %q expected invalid reset report, but passed", c.Name)
			}
		}
	}

	if attCount < 8 {
		t.Errorf("expected at least 8 attestation cases, found %d", attCount)
	}
	if profileCount < 1 {
		t.Errorf("expected at least 1 profile case, found %d", profileCount)
	}
	if resetCount < 1 {
		t.Errorf("expected at least 1 reset report case, found %d", resetCount)
	}
}

// TestX11_ProbeBypass verifies that direct PDO or native socket calls outside framework
// drivers invalidate completeness claims per Test X11.
func TestX11_ProbeBypass(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)

	baseMechanisms := []ObservationMechanism{
		{
			MechanismID:         "db_driver",
			Dimension:           DimensionDB,
			BoundaryRef:         dummyRef,
			InstrumentationRef:  &dummyRef,
			SupportedPaths:      []string{"application/models"},
			ExcludedPaths:       []string{},
			EscapePaths:         []string{},
			ControlTestRefs:     []string{dummyRef},
			ClosureTestRefs:     []string{dummyRef},
			FailureRefs:         []string{},
			ResidualLimitations: []string{},
			Coverage:            MechanismCoverageClosed,
		},
		{
			MechanismID:         "http_external",
			Dimension:           DimensionExternal,
			BoundaryRef:         dummyRef,
			InstrumentationRef:  &dummyRef,
			SupportedPaths:      []string{"application/libraries"},
			ExcludedPaths:       []string{},
			EscapePaths:         []string{},
			ControlTestRefs:     []string{dummyRef},
			ClosureTestRefs:     []string{dummyRef},
			FailureRefs:         []string{},
			ResidualLimitations: []string{},
			Coverage:            MechanismCoverageClosed,
		},
	}

	// 1. Positive Control: Clean codebase with no probe bypass
	cleanFiles := map[string][]byte{
		"application/models/User_model.php": []byte(`<?php
class User_model extends CI_Model {
    public function get_user($id) {
        return $this->db->get_where('users', ['id' => $id])->row();
    }
}
`),
	}

	cleanRes, err := DetectProbeBypass(cleanFiles, baseMechanisms)
	if err != nil {
		t.Fatalf("clean DetectProbeBypass failed: %v", err)
	}
	if cleanRes.CompletenessInvalidated {
		t.Errorf("expected clean codebase to not invalidate completeness")
	}
	if len(cleanRes.EscapePathsDetected) != 0 {
		t.Errorf("expected 0 escape paths, got %d", len(cleanRes.EscapePathsDetected))
	}

	// Derive attestation for clean: should derive COMPLETE
	cleanAtt, err := DeriveObservationAttestation(&DeriveAttestationRequest{
		AttestationID: "att-clean",
		ProjectID:     "proj-x11",
		ProfileRef:    dummyRef,
		Mechanisms:    cleanRes.MechanismsWithEscapes,
	})
	if err != nil {
		t.Fatalf("DeriveObservationAttestation failed for clean: %v", err)
	}
	if cleanAtt.DerivedStatus != AttestationStatusComplete {
		t.Errorf("expected COMPLETE status, got %s", cleanAtt.DerivedStatus)
	}

	// 2. Mutant: Unmonitored direct PDO instantiation and raw socket
	mutantFiles := map[string][]byte{
		"application/models/Raw_model.php": []byte(`<?php
class Raw_model extends CI_Model {
    public function query_external() {
        $pdo = new PDO("mysql:host=localhost;dbname=raw", "root", "secret");
        $sock = fsockopen("external.service.local", 8080);
    }
}
`),
	}

	mutantRes, err := DetectProbeBypass(mutantFiles, baseMechanisms)
	if err != nil {
		t.Fatalf("mutant DetectProbeBypass failed: %v", err)
	}
	if !mutantRes.CompletenessInvalidated {
		t.Errorf("expected mutant codebase to invalidate completeness")
	}
	if len(mutantRes.EscapePathsDetected) != 2 {
		t.Fatalf("expected 2 escape paths detected (PDO + fsockopen), got %d: %v", len(mutantRes.EscapePathsDetected), mutantRes.EscapePathsDetected)
	}

	// Derive attestation for mutant: MUST NOT derive COMPLETE!
	mutantAtt, err := DeriveObservationAttestation(&DeriveAttestationRequest{
		AttestationID: "att-mutant",
		ProjectID:     "proj-x11",
		ProfileRef:    dummyRef,
		Mechanisms:    mutantRes.MechanismsWithEscapes,
	})
	if err != nil {
		t.Fatalf("DeriveObservationAttestation failed for mutant: %v", err)
	}
	if mutantAtt.DerivedStatus == AttestationStatusComplete {
		t.Errorf("FALSIFICATION FAILED: unmonitored PDO/socket escape must never permit COMPLETE attestation status")
	}
	if mutantAtt.DerivedStatus != AttestationStatusPartial {
		t.Errorf("expected PARTIAL status for unmonitored escapes, got %s", mutantAtt.DerivedStatus)
	}
}

// TestX12_AsyncCausality verifies that HTTP success without causal queued delivery ack fails
// and that delivery beyond horizon or with unrelated cause fails per Test X12.
func TestX12_AsyncCausality(t *testing.T) {
	// 1. Positive Control: HTTP succeeds, worker acknowledges delivery within horizon, cause links back
	goodTrace := &AsyncObservationTrace{
		HTTPSuccess:          true,
		DispatchSpanID:       "span-http-dispatch",
		InitiatingCauseID:    "cause-order-placed",
		ScheduleHorizonMs:    5000,
		DeliverySpanID:       "span-email-delivered",
		DeliveryElapsedMs:    1200,
		DeliveryAckRecorded:  true,
		DeliveryCauseIDs:     []string{"cause-order-placed"},
		QuiescenceElapsedMs:  1000,
		RequiredQuiescenceMs: 500,
	}

	res, err := VerifyAsyncCausality(goodTrace)
	if err != nil {
		t.Fatalf("VerifyAsyncCausality failed: %v", err)
	}
	if res.Outcome != AsyncCausalityComplete {
		t.Errorf("expected COMPLETE outcome for good async trace, got %s", res.Outcome)
	}

	// 2. Mutant A: HTTP succeeds, but queued delivery ack is missing
	missingAckTrace := *goodTrace
	missingAckTrace.DeliveryAckRecorded = false
	resA, err := VerifyAsyncCausality(&missingAckTrace)
	if err != nil {
		t.Fatalf("VerifyAsyncCausality failed: %v", err)
	}
	if resA.Outcome == AsyncCausalityComplete {
		t.Errorf("FALSIFICATION FAILED: missing delivery ack must not produce COMPLETE")
	}
	if !strings.Contains(resA.Reason, "MISSING_CAUSAL_ACK") {
		t.Errorf("expected MISSING_CAUSAL_ACK, got %s", resA.Reason)
	}

	// 3. Mutant B: Delivery occurs beyond the schedule horizon
	lateTrace := *goodTrace
	lateTrace.DeliveryElapsedMs = 9000 // exceeds 5000ms horizon
	resB, err := VerifyAsyncCausality(&lateTrace)
	if err != nil {
		t.Fatalf("VerifyAsyncCausality failed: %v", err)
	}
	if resB.Outcome == AsyncCausalityComplete {
		t.Errorf("FALSIFICATION FAILED: late delivery beyond horizon must not produce COMPLETE")
	}
	if !strings.Contains(resB.Reason, "HORIZON_EXCEEDED") {
		t.Errorf("expected HORIZON_EXCEEDED, got %s", resB.Reason)
	}

	// 4. Mutant C: Delivery has unrelated cause_id
	unrelatedTrace := *goodTrace
	unrelatedTrace.DeliveryCauseIDs = []string{"unrelated-cause-id"}
	resC, err := VerifyAsyncCausality(&unrelatedTrace)
	if err != nil {
		t.Fatalf("VerifyAsyncCausality failed: %v", err)
	}
	if resC.Outcome == AsyncCausalityComplete {
		t.Errorf("FALSIFICATION FAILED: unrelated cause must not produce COMPLETE")
	}
	if !strings.Contains(resC.Reason, "UNRELATED_CAUSE") {
		t.Errorf("expected UNRELATED_CAUSE, got %s", resC.Reason)
	}
}

// TestX13_RuntimeAcquisition verifies that missing PHP extensions or lockfile mismatches
// block boot and prevent host/older runtime fallback per Test X13.
func TestX13_RuntimeAcquisition(t *testing.T) {
	expectedLock := strings.Repeat("b", 64)

	// 1. Positive Control: All required extensions available, lock matches
	goodReq := &RuntimePrerequisitesRequest{
		RequiredExtensions:  []string{"redis", "pdo_mysql", "mbstring"},
		AvailableExtensions: []string{"redis", "pdo_mysql", "mbstring", "json", "openssl"},
		LockFileHash:        expectedLock,
		ExpectedLockHash:    expectedLock,
	}

	goodRes, err := VerifyRuntimePrerequisites(goodReq)
	if err != nil {
		t.Fatalf("VerifyRuntimePrerequisites failed: %v", err)
	}
	if goodRes.BootBlocked {
		t.Errorf("expected boot not blocked for good prerequisites")
	}
	if goodRes.Coverage != MechanismCoverageClosed {
		t.Errorf("expected CLOSED coverage, got %s", goodRes.Coverage)
	}

	// 2. Mutant A: Missing required extension (e.g. redis)
	missingExtReq := *goodReq
	missingExtReq.AvailableExtensions = []string{"pdo_mysql", "mbstring"} // redis omitted
	resA, err := VerifyRuntimePrerequisites(&missingExtReq)
	if err != nil {
		t.Fatalf("VerifyRuntimePrerequisites failed: %v", err)
	}
	if !resA.BootBlocked {
		t.Errorf("FALSIFICATION FAILED: missing extension must block boot")
	}
	if resA.Coverage != MechanismCoverageUnavailable {
		t.Errorf("expected UNAVAILABLE coverage, got %s", resA.Coverage)
	}
	if len(resA.MissingExtensions) != 1 || resA.MissingExtensions[0] != "redis" {
		t.Errorf("expected [redis] in missing extensions, got %v", resA.MissingExtensions)
	}

	// 3. Mutant B: Incompatible lockfile hash
	badLockReq := *goodReq
	badLockReq.LockFileHash = strings.Repeat("c", 64) // mismatch
	resB, err := VerifyRuntimePrerequisites(&badLockReq)
	if err != nil {
		t.Fatalf("VerifyRuntimePrerequisites failed: %v", err)
	}
	if !resB.BootBlocked {
		t.Errorf("FALSIFICATION FAILED: incompatible lockfile must block boot")
	}
}

// TestX14_ResetDimensions verifies that DDL implicit commits, sequence shifts,
// residual Redis keys, and unkilled processes force reset outcome to INCOMPLETE or FAILED per Test X14.
func TestX14_ResetDimensions(t *testing.T) {
	// 1. Positive Control: Clean reset
	cleanChecks := &ResetContaminationChecks{
		DDLImplicitCommitContamination:  false,
		SequenceShiftContamination:      false,
		RedisResidualKeyContamination:   false,
		UnkilledProcessContamination:    false,
		NonTransactionalRowContamination: false,
	}

	cleanRes, err := VerifyResetContamination(cleanChecks)
	if err != nil {
		t.Fatalf("VerifyResetContamination failed: %v", err)
	}
	if cleanRes.Outcome != ResetOutcomeRestored {
		t.Errorf("expected RESTORED for clean checks, got %s", cleanRes.Outcome)
	}

	// 2. Mutant A: DDL implicit commit contamination
	ddlChecks := *cleanChecks
	ddlChecks.DDLImplicitCommitContamination = true
	resA, err := VerifyResetContamination(&ddlChecks)
	if err != nil {
		t.Fatalf("VerifyResetContamination failed: %v", err)
	}
	if resA.Outcome == ResetOutcomeRestored {
		t.Errorf("FALSIFICATION FAILED: DDL contamination must never permit RESTORED")
	}
	if resA.Outcome != ResetOutcomeFailed {
		t.Errorf("expected FAILED outcome for DDL contamination, got %s", resA.Outcome)
	}

	// 3. Mutant B: Redis residual keys / TTL queue contamination
	redisChecks := *cleanChecks
	redisChecks.RedisResidualKeyContamination = true
	resB, err := VerifyResetContamination(&redisChecks)
	if err != nil {
		t.Fatalf("VerifyResetContamination failed: %v", err)
	}
	if resB.Outcome == ResetOutcomeRestored {
		t.Errorf("FALSIFICATION FAILED: Redis contamination must never permit RESTORED")
	}
	if resB.Outcome != ResetOutcomeIncomplete {
		t.Errorf("expected INCOMPLETE outcome for Redis contamination, got %s", resB.Outcome)
	}

	// 4. Mutant C: Unkilled child worker process contamination
	procChecks := *cleanChecks
	procChecks.UnkilledProcessContamination = true
	resC, err := VerifyResetContamination(&procChecks)
	if err != nil {
		t.Fatalf("VerifyResetContamination failed: %v", err)
	}
	if resC.Outcome == ResetOutcomeRestored {
		t.Errorf("FALSIFICATION FAILED: unkilled process must never permit RESTORED")
	}
	if resC.Outcome != ResetOutcomeFailed {
		t.Errorf("expected FAILED outcome for unkilled process, got %s", resC.Outcome)
	}
}

// TestX70_ServerSideDatabaseEffects verifies that missing server capture produces UNKNOWN/UNAVAILABLE,
// and omitted server-side writes produce MISMATCH / PARTIAL per Test X70.
func TestX70_ServerSideDatabaseEffects(t *testing.T) {
	driverCapture := &DatabaseDriverCapture{
		ExecutedQueries: []string{"UPDATE orders SET status = 'completed' WHERE id = 1"},
		TargetTables:    []string{"orders"},
	}

	// 1. Positive Control: Server capture installed and records identical tables
	matchingServer := &ServerDatabaseCapture{
		MechanismInstalled: true,
		ModifiedCatalogs:   []string{"app_db"},
		ModifiedTables:     []string{"orders"},
	}

	res, err := VerifyDatabaseEffectsCompleteness(driverCapture, matchingServer)
	if err != nil {
		t.Fatalf("VerifyDatabaseEffectsCompleteness failed: %v", err)
	}
	if res.CompletenessStatus != AttestationStatusComplete {
		t.Errorf("expected COMPLETE status for matching driver and server capture, got %s", res.CompletenessStatus)
	}

	// 2. Mutant A: Missing server capture mechanism
	noServer := &ServerDatabaseCapture{
		MechanismInstalled: false,
	}
	resA, err := VerifyDatabaseEffectsCompleteness(driverCapture, noServer)
	if err != nil {
		t.Fatalf("VerifyDatabaseEffectsCompleteness failed: %v", err)
	}
	if resA.CompletenessStatus == AttestationStatusComplete {
		t.Errorf("FALSIFICATION FAILED: missing server capture must never permit COMPLETE status")
	}
	if resA.CompletenessStatus != AttestationStatusUnavailable {
		t.Errorf("expected UNAVAILABLE status, got %s", resA.CompletenessStatus)
	}
	if !strings.Contains(resA.Limitation, "MISSING_SERVER_CAPTURE") {
		t.Errorf("expected MISSING_SERVER_CAPTURE limitation, got %s", resA.Limitation)
	}

	// 3. Mutant B: Server capture reveals trigger wrote to audit_log table omitted by driver
	triggerServer := &ServerDatabaseCapture{
		MechanismInstalled: true,
		ModifiedCatalogs:   []string{"app_db"},
		ModifiedTables:     []string{"orders", "order_audit_logs"}, // trigger write
	}
	resB, err := VerifyDatabaseEffectsCompleteness(driverCapture, triggerServer)
	if err != nil {
		t.Fatalf("VerifyDatabaseEffectsCompleteness failed: %v", err)
	}
	if resB.CompletenessStatus == AttestationStatusComplete {
		t.Errorf("FALSIFICATION FAILED: omitted server-side write must never permit COMPLETE status")
	}
	if resB.CompletenessStatus != AttestationStatusPartial {
		t.Errorf("expected PARTIAL status, got %s", resB.CompletenessStatus)
	}
	if len(resB.OmittedWrites) != 1 || resB.OmittedWrites[0] != "order_audit_logs" {
		t.Errorf("expected [order_audit_logs] in omitted writes, got %v", resB.OmittedWrites)
	}
}

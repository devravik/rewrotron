package domain

import (
	"strings"
	"testing"
)

// TestX68_FilesystemSemantics verifies that metadata changes (mtime, permissions, executable bit)
// on files where properties are REQUIRED invalidate qualification despite identical content hashes per Test X68.
func TestX68_FilesystemSemantics(t *testing.T) {
	dummyRef := "sha256:" + strings.Repeat("a", 64)

	baseProfile := &FilesystemProfileRecord{
		SchemaVersion:      "1.0",
		ProjectID:          "proj-x68",
		ContentManifestRef: dummyRef,
		Properties: []FilesystemPropertyRecord{
			{Property: "PERMISSIONS", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
			{Property: "EXECUTABLE", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
			{Property: "MTIME", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
			{Property: "OWNERSHIP", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
			{Property: "HARDLINKS", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
			{Property: "SYMLINKS", Treatment: "REQUIRED", CapturedStateRef: &dummyRef, MechanismRef: &dummyRef, EvidenceRefs: []string{dummyRef}},
			{Property: "CASE_SENSITIVITY", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
			{Property: "PATH_RESOLUTION", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
			{Property: "WORKING_DIRECTORY", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
			{Property: "DIRECTORY_ORDER", Treatment: "IRRELEVANT", CapturedStateRef: nil, MechanismRef: nil, EvidenceRefs: []string{dummyRef}},
		},
		ControlRefs: []string{dummyRef},
		Limitations: []string{},
	}

	contentHash := strings.Repeat("f", 64)

	baseline := map[string]FileMetadataState{
		"cli/worker.php": {
			Path:          "cli/worker.php",
			ContentHash:   contentHash,
			Mode:          0755,
			IsExecutable:  true,
			MtimeNs:       1600000000000000000,
			SymlinkTarget: "",
			Owner:         "www-data",
		},
	}

	// 1. Positive Control: Metadata matches baseline exactly
	matchingCurrent := map[string]FileMetadataState{
		"cli/worker.php": {
			Path:          "cli/worker.php",
			ContentHash:   contentHash,
			Mode:          0755,
			IsExecutable:  true,
			MtimeNs:       1600000000000000000,
			SymlinkTarget: "",
			Owner:         "www-data",
		},
	}

	res, err := QualifyFilesystemProfile(baseProfile, baseline, matchingCurrent)
	if err != nil {
		t.Fatalf("QualifyFilesystemProfile failed: %v", err)
	}
	if !res.Valid || res.Stale {
		t.Errorf("expected clean filesystem profile to be valid and not stale")
	}
	if res.Status != AttestationStatusComplete {
		t.Errorf("expected COMPLETE status for matching profile, got %s", res.Status)
	}

	// 2. Mutant A: Executable bit stripped despite identical byte content hash
	mutantExec := map[string]FileMetadataState{
		"cli/worker.php": {
			Path:          "cli/worker.php",
			ContentHash:   contentHash, // Identical byte hash!
			Mode:          0644,
			IsExecutable:  false, // Executable bit stripped!
			MtimeNs:       1600000000000000000,
			SymlinkTarget: "",
			Owner:         "www-data",
		},
	}

	resA, err := QualifyFilesystemProfile(baseProfile, baseline, mutantExec)
	if err != nil {
		t.Fatalf("QualifyFilesystemProfile failed: %v", err)
	}
	if resA.Valid {
		t.Errorf("FALSIFICATION FAILED: changed executable bit must invalidate qualification")
	}
	if !resA.Stale {
		t.Errorf("expected Stale=true when required metadata drifts")
	}
	if !resA.SuccessorProfileRequired {
		t.Errorf("expected SuccessorProfileRequired=true")
	}
	if resA.Status != AttestationStatusUnavailable {
		t.Errorf("expected UNAVAILABLE status for stale filesystem, got %s", resA.Status)
	}
	if resA.InvalidationRef == "" {
		t.Errorf("expected non-empty InvalidationRef")
	}

	// 3. Mutant B: Mtime changed on timestamp-dependent file despite identical byte hash
	mutantMtime := map[string]FileMetadataState{
		"cli/worker.php": {
			Path:          "cli/worker.php",
			ContentHash:   contentHash, // Identical byte hash!
			Mode:          0755,
			IsExecutable:  true,
			MtimeNs:       1700000000000000000, // Modified timestamp!
			SymlinkTarget: "",
			Owner:         "www-data",
		},
	}

	resB, err := QualifyFilesystemProfile(baseProfile, baseline, mutantMtime)
	if err != nil {
		t.Fatalf("QualifyFilesystemProfile failed: %v", err)
	}
	if resB.Valid || !resB.Stale {
		t.Errorf("FALSIFICATION FAILED: changed MTIME on REQUIRED property must invalidate qualification")
	}

	// 4. Mutant C: Mode permissions altered
	mutantMode := map[string]FileMetadataState{
		"cli/worker.php": {
			Path:          "cli/worker.php",
			ContentHash:   contentHash,
			Mode:          0700, // Changed mode!
			IsExecutable:  true,
			MtimeNs:       1600000000000000000,
			SymlinkTarget: "",
			Owner:         "www-data",
		},
	}

	resC, err := QualifyFilesystemProfile(baseProfile, baseline, mutantMode)
	if err != nil {
		t.Fatalf("QualifyFilesystemProfile failed: %v", err)
	}
	if resC.Valid || !resC.Stale {
		t.Errorf("FALSIFICATION FAILED: changed mode permissions must invalidate qualification")
	}
}

// TestX69_EmbeddedCredentialAdmission verifies that source containing live credentials or canaries
// cannot be launched as original, cannot relabel a variant as original, and allows safe STATIC_ONLY or ADMITTED_VARIANT per Test X69.
func TestX69_EmbeddedCredentialAdmission(t *testing.T) {
	origManifestRef := "sha256:" + strings.Repeat("1", 64)
	origContentHash := strings.Repeat("1", 64)
	policyRef := "sha256:" + strings.Repeat("2", 64)

	// 1. Scan source containing production live canary and AWS secret
	filesWithCanary := map[string][]byte{
		"application/config/database.php": []byte(`<?php
$db['default'] = array(
	'dsn'	=> '',
	'hostname' => 'db.internal',
	'username' => 'app_user',
	'password' => 'canary_prod_live_secret_key_9981',
	'database' => 'production_db'
);
`),
		"application/config/aws.php": []byte(`<?php
$config['aws_key'] = 'AKIAIOSFODNN7EXAMPLE';
$config['aws_secret_access_key'] = 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY';
`),
	}

	scan := ScanSourceCredentials(filesWithCanary)
	if scan.CredentialStatus != "LIVE_OR_UNKNOWN" {
		t.Fatalf("expected LIVE_OR_UNKNOWN credential status, got %s", scan.CredentialStatus)
	}
	if len(scan.Findings) < 2 {
		t.Fatalf("expected at least 2 findings (canary + aws), got %d", len(scan.Findings))
	}

	// 2. Mutant A: Attempting to admit original for execution with live credentials MUST BE REFUSED!
	_, err := EvaluateExecutionAdmission(
		"proj-x69",
		origManifestRef,
		origContentHash,
		scan,
		policyRef,
		"ADMITTED_ORIGINAL",
		nil,
	)
	if err == nil {
		t.Fatalf("FALSIFICATION FAILED: ADMITTED_ORIGINAL must be refused when live credentials or canaries exist")
	}
	if !strings.Contains(err.Error(), "ERR_UNSAFE_ORIGINAL_CREDENTIALS") {
		t.Errorf("expected ERR_UNSAFE_ORIGINAL_CREDENTIALS, got %v", err)
	}

	// 3. Positive Control A: STATIC_ONLY admission succeeds without container launch
	staticAdm, err := EvaluateExecutionAdmission(
		"proj-x69",
		origManifestRef,
		origContentHash,
		scan,
		policyRef,
		"STATIC_ONLY",
		nil,
	)
	if err != nil {
		t.Fatalf("STATIC_ONLY admission failed: %v", err)
	}
	if staticAdm.Outcome != "STATIC_ONLY" {
		t.Errorf("expected STATIC_ONLY outcome, got %s", staticAdm.Outcome)
	}
	if staticAdm.VariantRef != nil {
		t.Errorf("variant_ref must be nil for STATIC_ONLY")
	}

	// 4. Mutant B: Relabeling variant as original (same manifest ref or same content hash) MUST BE REJECTED!
	fakeVariant := &ExecutionVariantRecord{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-x69",
		OriginalManifestRef:  origManifestRef,
		OriginalContentHash:  origContentHash,
		VariantManifestRef:   origManifestRef, // Relabeled as original!
		VariantContentHash:   origContentHash, // Identical content hash!
		TransformRef:         "sha256:" + strings.Repeat("3", 64),
		DeclaredChanges:      []VariantDeclaredChange{{Path: "application/config/database.php", BeforeRef: origManifestRef, AfterRef: "sha256:" + strings.Repeat("4", 64), Reason: "mock"}},
		CredentialPolicyRef:  policyRef,
		Limitations:          []string{"variant test"},
		ReviewEvidenceRefs:   []string{"sha256:" + strings.Repeat("5", 64)},
	}

	_, err = EvaluateExecutionAdmission(
		"proj-x69",
		origManifestRef,
		origContentHash,
		scan,
		policyRef,
		"ADMITTED_VARIANT",
		fakeVariant,
	)
	if err == nil {
		t.Fatalf("FALSIFICATION FAILED: relabeled variant with identical manifest ref must be rejected")
	}
	if !strings.Contains(err.Error(), "ERR_RELABELED_VARIANT") {
		t.Errorf("expected ERR_RELABELED_VARIANT, got %v", err)
	}

	// 5. Positive Control B: Properly reviewed variant with distinct manifest and synthetic credentials succeeds
	variantManifestRef := "sha256:" + strings.Repeat("8", 64)
	variantContentHash := strings.Repeat("8", 64)
	validVariant := &ExecutionVariantRecord{
		SchemaVersion:        "1.0",
		ProjectID:            "proj-x69",
		OriginalManifestRef:  origManifestRef,
		OriginalContentHash:  origContentHash,
		VariantManifestRef:   variantManifestRef,
		VariantContentHash:   variantContentHash,
		TransformRef:         "sha256:" + strings.Repeat("3", 64),
		DeclaredChanges: []VariantDeclaredChange{
			{
				Path:      "application/config/database.php",
				BeforeRef: "sha256:" + strings.Repeat("6", 64),
				AfterRef:  "sha256:" + strings.Repeat("7", 64),
				Reason:    "replace live database canary with synthetic mock credentials",
			},
		},
		CredentialPolicyRef: policyRef,
		Limitations:         []string{"Live database path unobserved in sandbox"},
		ReviewEvidenceRefs:  []string{"sha256:" + strings.Repeat("9", 64)},
	}

	variantAdm, err := EvaluateExecutionAdmission(
		"proj-x69",
		origManifestRef,
		origContentHash,
		scan,
		policyRef,
		"ADMITTED_VARIANT",
		validVariant,
	)
	if err != nil {
		t.Fatalf("valid ADMITTED_VARIANT evaluation failed: %v", err)
	}
	if variantAdm.Outcome != "ADMITTED_VARIANT" {
		t.Errorf("expected ADMITTED_VARIANT outcome, got %s", variantAdm.Outcome)
	}
	if variantAdm.VariantRef == nil || *variantAdm.VariantRef == "" {
		t.Errorf("expected non-empty variant_ref for ADMITTED_VARIANT")
	}
	if variantAdm.VariantCredentialStatus == nil || *variantAdm.VariantCredentialStatus != "SYNTHETIC_OR_ABSENT" {
		t.Errorf("expected SYNTHETIC_OR_ABSENT variant credential status, got %v", variantAdm.VariantCredentialStatus)
	}
}

package domain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// ArtifactIntegrityVerifier validates that generated target artifacts match their
// declared GenerationManifest entries, verifying exact SHA-256 byte bindings,
// tamper resistance, and absence of stale build references.
type ArtifactIntegrityVerifier struct{}

// NewArtifactIntegrityVerifier creates an ArtifactIntegrityVerifier.
func NewArtifactIntegrityVerifier() *ArtifactIntegrityVerifier {
	return &ArtifactIntegrityVerifier{}
}

// VerifyIntegrity verifies that all artifacts declared in the manifest exist with
// bitwise identical content matches, and that no extraneous undeclared artifacts exist.
func (v *ArtifactIntegrityVerifier) VerifyIntegrity(manifest *GenerationManifest, artifacts map[string][]byte) error {
	if manifest == nil {
		return fmt.Errorf("generation manifest cannot be nil")
	}
	if artifacts == nil {
		return fmt.Errorf("artifacts map cannot be nil")
	}

	manifestEntries := make(map[string]GenerationManifestEntry)
	for _, entry := range manifest.Entries {
		manifestEntries[entry.Path] = entry
	}

	// 1. Check all manifest entries match artifact bytes
	for path, entry := range manifestEntries {
		data, exists := artifacts[path]
		if !exists {
			return fmt.Errorf("missing artifact declared in manifest: %s", path)
		}

		sum := sha256.Sum256(data)
		actualHex := hex.EncodeToString(sum[:])
		expectedHex := strings.TrimPrefix(entry.ContentRef, "sha256:")

		if actualHex != expectedHex {
			return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", path, expectedHex, actualHex)
		}
		if entry.AfterSHA256 != actualHex {
			return fmt.Errorf("after_sha256 mismatch for %s: expected %s, got %s", path, actualHex, entry.AfterSHA256)
		}
	}

	// 2. Check no undeclared artifacts exist
	for path := range artifacts {
		if _, declared := manifestEntries[path]; !declared {
			return fmt.Errorf("undeclared extraneous artifact generated: %s", path)
		}
	}

	return nil
}

// CompareGenerationOutputs compares two independent generation runs for bitwise identical reproducibility.
func (v *ArtifactIntegrityVerifier) CompareGenerationOutputs(run1, run2 *GenerationOutput) error {
	if run1 == nil || run2 == nil {
		return fmt.Errorf("generation outputs cannot be nil")
	}

	// 1. Compare manifest provenance
	if run1.Manifest.ProvenanceRef != run2.Manifest.ProvenanceRef {
		return fmt.Errorf("provenance digest mismatch between runs: %s != %s", run1.Manifest.ProvenanceRef, run2.Manifest.ProvenanceRef)
	}

	// 2. Compare artifact count
	if len(run1.Artifacts) != len(run2.Artifacts) {
		return fmt.Errorf("artifact count mismatch between runs: %d != %d", len(run1.Artifacts), len(run2.Artifacts))
	}

	// 3. Compare each artifact bitwise
	for path, data1 := range run1.Artifacts {
		data2, exists := run2.Artifacts[path]
		if !exists {
			return fmt.Errorf("artifact %s missing from second run", path)
		}
		if !bytes.Equal(data1, data2) {
			return fmt.Errorf("bitwise content divergence in %s across runs", path)
		}
	}

	return nil
}

// TestM12_Unit2_DeterministicGenerationAndArtifactIntegrity tests:
// 1. Bitwise and normalized deterministic generation across repeated runs
// 2. Cryptographic binding between artifact bytes, hashes, and GenerationManifest
// 3. Falsification tests catching:
//    - Tampered artifact bytes (digest mismatch)
//    - Undeclared extraneous artifacts
//    - Stale / mismatched manifest provenance
//    - Missing declared artifacts
func TestM12_Unit2_DeterministicGenerationAndArtifactIntegrity(t *testing.T) {
	ctx := context.Background()
	verifier := NewArtifactIntegrityVerifier()

	unitID := "reproducible-vertical-slice"
	adm := makeTestAdmission(t, unitID, []string{
		"routing",
		"controllers",
		"eloquent_models",
		"database",
		"transactions",
		"auth",
		"session",
		"csrf",
	}, []string{
		"health",
		"login",
		"whoami",
		"csrf",
		"create",
		"item",
		"note",
		"logout",
	})
	tm := makeTestTargetModel(t, unitID)
	sc := makeTestSessionContract(t, tm.ProjectID, "login", "SAFE_BRIDGE")

	adapter := NewLaravelTargetAdapter()

	// 1. Generate run 1
	run1, err := adapter.Generate(ctx, GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: sc,
	})
	if err != nil {
		t.Fatalf("generation run 1 failed: %v", err)
	}

	// 2. Generate run 2 from identical inputs
	run2, err := adapter.Generate(ctx, GenerationInput{
		Admission:       adm,
		TargetModel:     tm,
		UnitID:          unitID,
		SessionContract: sc,
	})
	if err != nil {
		t.Fatalf("generation run 2 failed: %v", err)
	}

	// 3. Verify bitwise reproducibility
	if err := verifier.CompareGenerationOutputs(run1, run2); err != nil {
		t.Fatalf("reproducibility check failed: %v", err)
	}

	// 4. Verify artifact integrity against manifest
	if err := verifier.VerifyIntegrity(run1.Manifest, run1.Artifacts); err != nil {
		t.Fatalf("artifact integrity verification failed: %v", err)
	}

	// =========================================================================
	// Falsification Tests: Ensure Defects and Tampering are Caught
	// =========================================================================

	t.Run("Falsification_TamperedArtifactContent", func(t *testing.T) {
		tamperedArtifacts := make(map[string][]byte)
		for k, v := range run1.Artifacts {
			tamperedArtifacts[k] = append([]byte(nil), v...)
		}
		// Tamper with routes/web.php
		tamperedArtifacts["routes/web.php"] = append(tamperedArtifacts["routes/web.php"], []byte("\n// unauthorized alteration")...)

		err := verifier.VerifyIntegrity(run1.Manifest, tamperedArtifacts)
		if err == nil || !strings.Contains(err.Error(), "checksum mismatch for routes/web.php") {
			t.Fatalf("expected checksum mismatch error on tampered artifact, got: %v", err)
		}
	})

	t.Run("Falsification_MissingDeclaredArtifact", func(t *testing.T) {
		missingArtifacts := make(map[string][]byte)
		for k, v := range run1.Artifacts {
			if k != "app/Models/User.php" {
				missingArtifacts[k] = v
			}
		}

		err := verifier.VerifyIntegrity(run1.Manifest, missingArtifacts)
		if err == nil || !strings.Contains(err.Error(), "missing artifact declared in manifest: app/Models/User.php") {
			t.Fatalf("expected missing artifact error, got: %v", err)
		}
	})

	t.Run("Falsification_UndeclaredExtraneousArtifact", func(t *testing.T) {
		extraneousArtifacts := make(map[string][]byte)
		for k, v := range run1.Artifacts {
			extraneousArtifacts[k] = v
		}
		extraneousArtifacts["unauthorized_backdoor.php"] = []byte("<?php echo 'backdoor';")

		err := verifier.VerifyIntegrity(run1.Manifest, extraneousArtifacts)
		if err == nil || !strings.Contains(err.Error(), "undeclared extraneous artifact generated: unauthorized_backdoor.php") {
			t.Fatalf("expected undeclared extraneous artifact error, got: %v", err)
		}
	})

	t.Run("Falsification_MismatchedManifestProvenance", func(t *testing.T) {
		tamperedRun := &GenerationOutput{
			Manifest: &GenerationManifest{
				SchemaVersion: run1.Manifest.SchemaVersion,
				ProjectID:     run1.Manifest.ProjectID,
				UnitID:        run1.Manifest.UnitID,
				Entries:       run1.Manifest.Entries,
				ProvenanceRef: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
			Artifacts: run1.Artifacts,
		}

		err := verifier.CompareGenerationOutputs(run1, tamperedRun)
		if err == nil || !strings.Contains(err.Error(), "provenance digest mismatch between runs") {
			t.Fatalf("expected provenance digest mismatch error, got: %v", err)
		}
	})

	t.Run("Falsification_DeterminismUnderAlphabeticalPathOrdering", func(t *testing.T) {
		// Ensure manifest entries are deterministically sorted by path
		var paths []string
		for _, e := range run1.Manifest.Entries {
			paths = append(paths, e.Path)
		}
		if !sort.StringsAreSorted(paths) {
			t.Fatalf("manifest entries are not deterministically sorted by path: %+v", paths)
		}
	})
}

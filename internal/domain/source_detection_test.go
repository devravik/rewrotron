package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateSourceDetection_Fixtures(t *testing.T) {
	fixtureData, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases fixture: %v", err)
	}

	var cases []validationFixtureCase
	if err := json.Unmarshal(fixtureData, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	testedDetections := 0
	for _, tc := range cases {
		if tc.Schema == "https://rewrotron.dev/specs/schemas/detection-schema.json" {
			testedDetections++
			err := ValidateSourceDetection(tc.Instance)
			if tc.Valid && err != nil {
				t.Errorf("fixture %q expected valid, got error: %v", tc.Name, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("fixture %q expected invalid, got nil error", tc.Name)
			}
		}
	}

	if testedDetections < 2 {
		t.Fatalf("expected at least 2 detection fixtures, got %d", testedDetections)
	}
}

func baseValidDetectionMap() map[string]interface{} {
	return map[string]interface{}{
		"schema_version":       "1.0",
		"detection_id":         "det-1",
		"project_id":           "prj-1",
		"source_manifest_hash": "2222222222222222222222222222222222222222222222222222222222222222",
		"language":             "php",
		"framework":            "codeigniter",
		"detected_version":     "3.1.13",
		"php_compatibility":    "7.4",
		"application_type":     "web",
		"database_engine":      "mysql",
		"authentication_type":  "session",
		"confidence": map[string]interface{}{
			"framework": 0.95,
			"language":  1.0,
		},
		"potential_concerns": []string{"PHP 7.4 end-of-life"},
		"gate_status":        "AWAITING_SOURCE_CONFIRMATION",
		"evidence": []map[string]interface{}{
			{
				"evidence_id":  "ev-1",
				"kind":         "STATIC",
				"artifact_ref": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
				"description":  "Found CodeIgniter constants in index.php",
				"symbol_id":    "sym-1",
			},
		},
		"routes_count":      12,
		"controllers_count": 5,
	}
}

func TestValidateSourceDetection_Valid10And20(t *testing.T) {
	det10 := baseValidDetectionMap()
	raw10, err := json.Marshal(det10)
	if err != nil {
		t.Fatalf("failed to marshal 1.0 detection: %v", err)
	}
	if err := ValidateSourceDetection(raw10); err != nil {
		t.Errorf("expected valid 1.0 detection, got %v", err)
	}

	det20 := baseValidDetectionMap()
	det20["schema_version"] = "2.0"
	delete(det20, "gate_status")
	raw20, err := json.Marshal(det20)
	if err != nil {
		t.Fatalf("failed to marshal 2.0 detection: %v", err)
	}
	if err := ValidateSourceDetection(raw20); err != nil {
		t.Errorf("expected valid 2.0 detection, got %v", err)
	}
}

func TestValidateSourceDetection_GateStatusInvariants(t *testing.T) {
	// 2.0 with gate_status must fail
	det20 := baseValidDetectionMap()
	det20["schema_version"] = "2.0"
	det20["gate_status"] = "SOURCE_CONFIRMED"
	raw20, _ := json.Marshal(det20)
	if err := ValidateSourceDetection(raw20); err == nil {
		t.Errorf("expected error for 2.0 detection with gate_status")
	}

	// 1.0 without gate_status must fail
	det10 := baseValidDetectionMap()
	delete(det10, "gate_status")
	raw10, _ := json.Marshal(det10)
	if err := ValidateSourceDetection(raw10); err == nil {
		t.Errorf("expected error for 1.0 detection missing gate_status")
	}

	// 1.0 with invalid gate_status enum
	det10Invalid := baseValidDetectionMap()
	det10Invalid["gate_status"] = "APPROVED"
	raw10Invalid, _ := json.Marshal(det10Invalid)
	if err := ValidateSourceDetection(raw10Invalid); err == nil {
		t.Errorf("expected error for invalid gate_status enum")
	}
}

func TestValidateSourceDetection_ConfidenceScore(t *testing.T) {
	det := baseValidDetectionMap()
	det["confidence"] = map[string]interface{}{
		"framework": 1.5, // > 1.0
	}
	rawOver, _ := json.Marshal(det)
	if err := ValidateSourceDetection(rawOver); err == nil {
		t.Errorf("expected error for confidence > 1.0")
	}

	det["confidence"] = map[string]interface{}{
		"framework": -0.1, // < 0.0
	}
	rawUnder, _ := json.Marshal(det)
	if err := ValidateSourceDetection(rawUnder); err == nil {
		t.Errorf("expected error for confidence < 0.0")
	}
}

func TestValidateSourceDetection_EvidenceValidation(t *testing.T) {
	// invalid kind
	det := baseValidDetectionMap()
	det["evidence"] = []map[string]interface{}{
		{
			"evidence_id":  "ev-1",
			"kind":         "INVALID_KIND",
			"artifact_ref": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
			"description":  "Some evidence",
		},
	}
	rawBadKind, _ := json.Marshal(det)
	if err := ValidateSourceDetection(rawBadKind); err == nil {
		t.Errorf("expected error for invalid evidence kind")
	}

	// invalid artifact_ref
	det = baseValidDetectionMap()
	det["evidence"] = []map[string]interface{}{
		{
			"evidence_id":  "ev-1",
			"kind":         "STATIC",
			"artifact_ref": "not-a-sha256",
			"description":  "Some evidence",
		},
	}
	rawBadRef, _ := json.Marshal(det)
	if err := ValidateSourceDetection(rawBadRef); err == nil {
		t.Errorf("expected error for invalid artifact_ref")
	}

	// empty description
	det = baseValidDetectionMap()
	det["evidence"] = []map[string]interface{}{
		{
			"evidence_id":  "ev-1",
			"kind":         "STATIC",
			"artifact_ref": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
			"description":  "",
		},
	}
	rawEmptyDesc, _ := json.Marshal(det)
	if err := ValidateSourceDetection(rawEmptyDesc); err == nil {
		t.Errorf("expected error for empty evidence description")
	}
}

func TestValidateSourceDetection_CountFields(t *testing.T) {
	// negative count
	det := baseValidDetectionMap()
	det["routes_count"] = -1
	rawNeg, _ := json.Marshal(det)
	if err := ValidateSourceDetection(rawNeg); err == nil {
		t.Errorf("expected error for negative routes_count")
	}

	// float count
	rawFloat := []byte(strings.Replace(string(rawNeg), `"routes_count":-1`, `"routes_count":2.5`, 1))
	if err := ValidateSourceDetection(rawFloat); err == nil {
		t.Errorf("expected error for float routes_count")
	}
}

func TestValidateSourceDetection_SyntaxErrors(t *testing.T) {
	// invalid UTF-8
	if err := ValidateSourceDetection([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	// duplicate key
	dupJSON := `{"schema_version":"1.0","schema_version":"1.0"}`
	if err := ValidateSourceDetection([]byte(dupJSON)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	// trailing input
	validDet := baseValidDetectionMap()
	raw, _ := json.Marshal(validDet)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateSourceDetection(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}

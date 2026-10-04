package domain

import (
	"encoding/json"
	"os"
	"testing"
)

func TestValidateTargetEnvironment_Fixtures(t *testing.T) {
	fixtureData, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases fixture: %v", err)
	}

	var cases []validationFixtureCase
	if err := json.Unmarshal(fixtureData, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	testedTargets := 0
	for _, tc := range cases {
		if tc.Schema == "https://rewrotron.dev/specs/project/target.schema.json" {
			testedTargets++
			err := ValidateTargetEnvironment(tc.Instance)
			if tc.Valid && err != nil {
				t.Errorf("fixture %q expected valid, got error: %v", tc.Name, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("fixture %q expected invalid, got nil error", tc.Name)
			}
		}
	}

	if testedTargets < 4 {
		t.Fatalf("expected at least 4 target fixtures, got %d", testedTargets)
	}
}

func baseValidTargetMap() map[string]interface{} {
	return map[string]interface{}{
		"language":        "php",
		"framework":       "laravel",
		"version":         "11.*",
		"database_policy": "preserve_existing",
		"path":            "/workspace/target",
		"architecture": map[string]interface{}{
			"style": "modular_monolith",
		},
	}
}

func baseValidTargetV2Map() map[string]interface{} {
	tgt := baseValidTargetMap()
	tgt["schema_version"] = "2.0"
	tgt["runtime"] = map[string]interface{}{
		"version":      "8.3",
		"os":           "linux",
		"architecture": "amd64",
	}
	tgt["deployment"] = map[string]interface{}{
		"model":       "docker_compose",
		"profile_ref": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
	}
	return tgt
}

func TestValidateTargetEnvironment_Valid10And20(t *testing.T) {
	t10 := baseValidTargetMap()
	raw10, err := json.Marshal(t10)
	if err != nil {
		t.Fatalf("failed to marshal 1.0 target: %v", err)
	}
	if err := ValidateTargetEnvironment(raw10); err != nil {
		t.Errorf("expected valid 1.0 target, got %v", err)
	}

	t20 := baseValidTargetV2Map()
	raw20, err := json.Marshal(t20)
	if err != nil {
		t.Fatalf("failed to marshal 2.0 target: %v", err)
	}
	if err := ValidateTargetEnvironment(raw20); err != nil {
		t.Errorf("expected valid 2.0 target, got %v", err)
	}
}

func TestValidateTargetEnvironment_V2FrameworkNoneRules(t *testing.T) {
	// framework="none" with version=null is valid
	tNone := baseValidTargetV2Map()
	tNone["framework"] = "none"
	tNone["version"] = nil
	rawNone, err := json.Marshal(tNone)
	if err != nil {
		t.Fatalf("failed to marshal framework=none target: %v", err)
	}
	if err := ValidateTargetEnvironment(rawNone); err != nil {
		t.Errorf("expected valid for framework=none and version=null, got %v", err)
	}

	// framework="none" with version="1.0" is invalid
	tNone["version"] = "1.0"
	rawBadVersion, _ := json.Marshal(tNone)
	if err := ValidateTargetEnvironment(rawBadVersion); err == nil {
		t.Errorf("expected error when framework=none and version is non-null")
	}

	// framework="laravel" with version=null is invalid
	tLaravel := baseValidTargetV2Map()
	tLaravel["framework"] = "laravel"
	tLaravel["version"] = nil
	rawLaravelNull, _ := json.Marshal(tLaravel)
	if err := ValidateTargetEnvironment(rawLaravelNull); err == nil {
		t.Errorf("expected error when framework!=none and version is null")
	}
}

func TestValidateTargetEnvironment_DatabaseAndSubsystems(t *testing.T) {
	// valid database, web, cache, queue, storage, package_manager, compatibility
	tgt := baseValidTargetMap()
	tgt["database"] = map[string]interface{}{
		"engine":          "mysql",
		"version":         "8.0",
		"preserve_schema": true,
	}
	tgt["web"] = map[string]interface{}{
		"server":  "nginx",
		"version": "1.24",
	}
	tgt["cache"] = map[string]interface{}{
		"engine": "redis",
	}
	tgt["queue"] = map[string]interface{}{
		"engine": "redis",
	}
	tgt["storage"] = map[string]interface{}{
		"local":  true,
		"engine": "s3",
	}
	tgt["package_manager"] = map[string]interface{}{
		"name": "composer",
	}
	tgt["compatibility"] = map[string]interface{}{
		"http":              "strict",
		"api":               "strict",
		"auth":              "approved_changes",
		"database_behavior": "strict",
	}
	raw, err := json.Marshal(tgt)
	if err != nil {
		t.Fatalf("failed to marshal rich target: %v", err)
	}
	if err := ValidateTargetEnvironment(raw); err != nil {
		t.Errorf("expected valid rich target, got %v", err)
	}

	// database preserve_schema = false must fail
	tgt["database"] = map[string]interface{}{
		"engine":          "mysql",
		"version":         "8.0",
		"preserve_schema": false,
	}
	rawBadDB, _ := json.Marshal(tgt)
	if err := ValidateTargetEnvironment(rawBadDB); err == nil {
		t.Errorf("expected error when database preserve_schema is false")
	}

	// invalid database_policy
	tgtValidDB := baseValidTargetMap()
	tgtValidDB["database_policy"] = "recreate"
	rawBadPolicy, _ := json.Marshal(tgtValidDB)
	if err := ValidateTargetEnvironment(rawBadPolicy); err == nil {
		t.Errorf("expected error when database_policy != preserve_existing")
	}
}

func TestValidateTargetEnvironment_SyntaxErrors(t *testing.T) {
	if err := ValidateTargetEnvironment([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	dupJSON := `{"language":"php","language":"php"}`
	if err := ValidateTargetEnvironment([]byte(dupJSON)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	validTgt := baseValidTargetMap()
	raw, _ := json.Marshal(validTgt)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateTargetEnvironment(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}

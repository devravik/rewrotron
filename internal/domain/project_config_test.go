package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateProjectConfig_Fixtures(t *testing.T) {
	fixtureData, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases fixture: %v", err)
	}

	var cases []validationFixtureCase
	if err := json.Unmarshal(fixtureData, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	testedProjects := 0
	for _, tc := range cases {
		if tc.Schema == "https://rewrotron.dev/specs/project/project.schema.json" {
			testedProjects++
			err := ValidateProjectConfig(tc.Instance)
			if tc.Valid && err != nil {
				t.Errorf("fixture %q expected valid, got error: %v", tc.Name, err)
			} else if !tc.Valid && err == nil {
				t.Errorf("fixture %q expected invalid, got nil error", tc.Name)
			}
		}
	}

	if testedProjects < 7 {
		t.Fatalf("expected at least 7 project config fixtures, got %d", testedProjects)
	}
}

func baseValidProjectConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"version":    1,
		"project_id": "prj-1",
		"source": map[string]interface{}{
			"type": "codeigniter3",
			"path": "/source/path",
		},
		"environment": map[string]interface{}{
			"mode": "docker",
		},
		"safety": map[string]interface{}{
			"production_access": "denied",
			"external_writes":   "denied",
			"export_policy":     "local_only",
		},
	}
}

func TestValidateProjectConfig_ValidMinimalAndRich(t *testing.T) {
	cfg := baseValidProjectConfigMap()
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal minimal config: %v", err)
	}
	if err := ValidateProjectConfig(raw); err != nil {
		t.Errorf("expected valid minimal config, got %v", err)
	}

	// rich config with target and migration
	cfg["target"] = baseValidTargetMap()
	cfg["compatibility"] = map[string]interface{}{
		"requirement":               "strict",
		"active_session_continuity": "required",
		"policy_ref":                "sha256:2222222222222222222222222222222222222222222222222222222222222222",
	}
	cfg["migration"] = map[string]interface{}{
		"strategy":              "vertical_slice",
		"remove_dead_code":      true,
		"preserve_dormant_code": true,
		"optimization": map[string]interface{}{
			"enabled_after_parity": true,
		},
	}
	rawRich, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal rich config: %v", err)
	}
	if err := ValidateProjectConfig(rawRich); err != nil {
		t.Errorf("expected valid rich config, got %v", err)
	}
}

func TestValidateProjectConfig_SafetyInvariants(t *testing.T) {
	// production_access != denied
	cfg := baseValidProjectConfigMap()
	safety := cfg["safety"].(map[string]interface{})
	safety["production_access"] = "allowed"
	raw, _ := json.Marshal(cfg)
	if err := ValidateProjectConfig(raw); err == nil {
		t.Errorf("expected error when production_access is allowed")
	}

	// external_writes != denied
	safety["production_access"] = "denied"
	safety["external_writes"] = "enabled"
	rawExt, _ := json.Marshal(cfg)
	if err := ValidateProjectConfig(rawExt); err == nil {
		t.Errorf("expected error when external_writes is enabled")
	}

	// export_policy invalid
	safety["external_writes"] = "denied"
	safety["export_policy"] = "public"
	rawExp, _ := json.Marshal(cfg)
	if err := ValidateProjectConfig(rawExp); err == nil {
		t.Errorf("expected error when export_policy is public")
	}
}

func TestValidateProjectConfig_MigrationInvariants(t *testing.T) {
	cfg := baseValidProjectConfigMap()
	cfg["migration"] = map[string]interface{}{
		"strategy":              "vertical_slice",
		"remove_dead_code":      true,
		"preserve_dormant_code": false, // must be true
		"optimization": map[string]interface{}{
			"enabled_after_parity": true,
		},
	}
	raw, _ := json.Marshal(cfg)
	if err := ValidateProjectConfig(raw); err == nil {
		t.Errorf("expected error when preserve_dormant_code is false")
	}

	// optimization before parity
	mig := cfg["migration"].(map[string]interface{})
	mig["preserve_dormant_code"] = true
	mig["optimization"] = map[string]interface{}{
		"enabled_after_parity": false,
	}
	rawOpt, _ := json.Marshal(cfg)
	if err := ValidateProjectConfig(rawOpt); err == nil {
		t.Errorf("expected error when enabled_after_parity is false")
	}
}

func TestValidateProjectConfig_VersionAndSource(t *testing.T) {
	// version != 1
	cfg := baseValidProjectConfigMap()
	cfg["version"] = 2
	rawVer, _ := json.Marshal(cfg)
	if err := ValidateProjectConfig(rawVer); err == nil {
		t.Errorf("expected error when version != 1")
	}

	// float version
	rawFloat := []byte(strings.Replace(string(rawVer), `"version":2`, `"version":1.0`, 1))
	if err := ValidateProjectConfig(rawFloat); err == nil {
		t.Errorf("expected error for float version")
	}

	// extra field in source (e.g. yaml spoofing)
	cfgValid := baseValidProjectConfigMap()
	src := cfgValid["source"].(map[string]interface{})
	src["confirmed"] = true
	rawSpoof, _ := json.Marshal(cfgValid)
	if err := ValidateProjectConfig(rawSpoof); err == nil {
		t.Errorf("expected error for unexpected field in source")
	}
}

func TestValidateProjectConfig_SyntaxErrors(t *testing.T) {
	if err := ValidateProjectConfig([]byte{0xff, 0xfe}); err == nil {
		t.Errorf("expected error for invalid UTF-8")
	}

	dupJSON := `{"version":1,"version":1}`
	if err := ValidateProjectConfig([]byte(dupJSON)); err == nil {
		t.Errorf("expected error for duplicate key")
	}

	cfg := baseValidProjectConfigMap()
	raw, _ := json.Marshal(cfg)
	trailing := append(raw, []byte(` "trailing"`)...)
	if err := ValidateProjectConfig(trailing); err == nil {
		t.Errorf("expected error for trailing input")
	}
}

package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateCommonDefs_ValidationCases(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	entrypointCasesFound := 0
	artifactPolicyCasesFound := 0

	for _, tc := range cases {
		switch tc.Schema {
		case "https://rewrotron.dev/specs/common.schema.json#/$defs/entrypoint":
			entrypointCasesFound++
			t.Run("entrypoint/"+tc.Name, func(t *testing.T) {
				err := ValidateEntrypoint(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid entrypoint fixture %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid entrypoint fixture %q to fail, but it passed", tc.Name)
				}
			})

		case "https://rewrotron.dev/specs/common.schema.json#/$defs/artifact_policy":
			artifactPolicyCasesFound++
			t.Run("artifact_policy/"+tc.Name, func(t *testing.T) {
				err := ValidateArtifactPolicy(tc.Instance)
				if tc.Valid && err != nil {
					t.Fatalf("expected valid artifact_policy fixture %q to pass, got: %v", tc.Name, err)
				}
				if !tc.Valid && err == nil {
					t.Fatalf("expected invalid artifact_policy fixture %q to fail, but it passed", tc.Name)
				}
			})
		}
	}

	if entrypointCasesFound == 0 {
		t.Fatalf("no entrypoint cases found in validation-cases.json")
	}
	if artifactPolicyCasesFound == 0 {
		t.Fatalf("no artifact_policy cases found in validation-cases.json")
	}
}

func TestValidateArtifactPolicy_Rejections(t *testing.T) {
	const validRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	baseValid := func() map[string]any {
		return map[string]any{
			"visibility": "SANITIZED_EVIDENCE",
			"policy_ref": validRef,
		}
	}

	// 1. Valid baseline passes
	baseBytes, _ := json.Marshal(baseValid())
	if err := ValidateArtifactPolicy(baseBytes); err != nil {
		t.Fatalf("expected baseline artifact_policy to pass, got: %v", err)
	}

	// Also test PRIVATE_SOURCE visibility
	mPrivate := baseValid()
	mPrivate["visibility"] = "PRIVATE_SOURCE"
	mPrivateBytes, _ := json.Marshal(mPrivate)
	if err := ValidateArtifactPolicy(mPrivateBytes); err != nil {
		t.Fatalf("expected PRIVATE_SOURCE artifact_policy to pass, got: %v", err)
	}

	// 2. Missing required fields
	for k := range baseValid() {
		m := baseValid()
		delete(m, k)
		raw, _ := json.Marshal(m)
		if err := ValidateArtifactPolicy(raw); err == nil {
			t.Errorf("expected error for missing field %q", k)
		}
	}

	// 3. Null required fields
	for k := range baseValid() {
		m := baseValid()
		m[k] = nil
		raw, _ := json.Marshal(m)
		if err := ValidateArtifactPolicy(raw); err == nil {
			t.Errorf("expected error for null field %q", k)
		}
	}

	// 4. Invalid visibility enum
	m := baseValid()
	m["visibility"] = "PUBLIC"
	raw, _ := json.Marshal(m)
	if err := ValidateArtifactPolicy(raw); err == nil {
		t.Errorf("expected error for invalid visibility enum")
	}

	// 5. Invalid policy_ref format
	m = baseValid()
	m["policy_ref"] = "bad-ref"
	raw, _ = json.Marshal(m)
	if err := ValidateArtifactPolicy(raw); err == nil {
		t.Errorf("expected error for invalid policy_ref")
	}

	// 6. Unexpected field
	m = baseValid()
	m["extra"] = "forbidden"
	raw, _ = json.Marshal(m)
	if err := ValidateArtifactPolicy(raw); err == nil {
		t.Errorf("expected error for unexpected field")
	}
}

func TestValidateEntrypoint_Rejections(t *testing.T) {
	// HTTP entrypoint
	httpValid := map[string]any{
		"type":          "http",
		"method":        "GET",
		"path":          "/api/v1/users",
		"content_types": []string{"application/json"},
	}
	raw, _ := json.Marshal(httpValid)
	if err := ValidateEntrypoint(raw); err != nil {
		t.Fatalf("expected valid http entrypoint to pass, got: %v", err)
	}

	// HTTP missing path
	m := map[string]any{"type": "http", "method": "GET"}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for missing http path")
	}

	// HTTP invalid method
	m = map[string]any{"type": "http", "method": "INVALID", "path": "/"}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for invalid http method")
	}

	// HTTP empty path
	m = map[string]any{"type": "http", "method": "GET", "path": ""}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for empty http path")
	}

	// CLI entrypoint
	cliValid := map[string]any{
		"type":             "cli",
		"command_id":       "migrate_db",
		"arguments_schema": map[string]any{"type": "object"},
	}
	raw, _ = json.Marshal(cliValid)
	if err := ValidateEntrypoint(raw); err != nil {
		t.Fatalf("expected valid cli entrypoint to pass, got: %v", err)
	}

	// CLI invalid command_id
	m = map[string]any{"type": "cli", "command_id": "CMD UPPER"}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for invalid cli command_id")
	}

	// CRON entrypoint
	cronValid := map[string]any{
		"type":       "cron",
		"command_id": "daily_backup",
		"schedule":   "0 0 * * *",
		"timezone":   "UTC",
	}
	raw, _ = json.Marshal(cronValid)
	if err := ValidateEntrypoint(raw); err != nil {
		t.Fatalf("expected valid cron entrypoint to pass, got: %v", err)
	}

	// CRON missing timezone
	m = map[string]any{"type": "cron", "command_id": "daily_backup", "schedule": "0 0 * * *"}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for missing cron timezone")
	}

	// EVENT entrypoint
	eventValid := map[string]any{
		"type":           "event",
		"channel":        "user_created",
		"message_schema": map[string]any{"type": "object"},
	}
	raw, _ = json.Marshal(eventValid)
	if err := ValidateEntrypoint(raw); err != nil {
		t.Fatalf("expected valid event entrypoint to pass, got: %v", err)
	}

	// EVENT missing message_schema
	m = map[string]any{"type": "event", "channel": "user_created"}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for missing event message_schema")
	}

	// Unsupported type
	m = map[string]any{"type": "grpc"}
	raw, _ = json.Marshal(m)
	if err := ValidateEntrypoint(raw); err == nil {
		t.Errorf("expected error for unsupported entrypoint type")
	}
}

func TestValidateEvidence_Rejections(t *testing.T) {
	const validRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	baseValid := func() map[string]any {
		return map[string]any{
			"evidence_id":  "ev_001",
			"kind":         "STATIC",
			"artifact_ref": validRef,
			"description":  "Evidence collected from static analysis",
			"symbol_id":    "sym_login",
			"trace_id":     "tr_001",
		}
	}

	// 1. Valid baseline passes
	baseBytes, _ := json.Marshal(baseValid())
	if err := ValidateEvidence(baseBytes); err != nil {
		t.Fatalf("expected baseline evidence to pass, got: %v", err)
	}

	// 2. Missing required fields
	for _, req := range []string{"evidence_id", "kind", "artifact_ref", "description"} {
		m := baseValid()
		delete(m, req)
		raw, _ := json.Marshal(m)
		if err := ValidateEvidence(raw); err == nil {
			t.Errorf("expected error for missing field %q", req)
		}
	}

	// 3. Invalid kind enum
	m := baseValid()
	m["kind"] = "UNKNOWN"
	raw, _ := json.Marshal(m)
	if err := ValidateEvidence(raw); err == nil {
		t.Errorf("expected error for invalid evidence kind")
	}

	// 4. Invalid evidence_id pattern
	m = baseValid()
	m["evidence_id"] = "EV UPPER"
	raw, _ = json.Marshal(m)
	if err := ValidateEvidence(raw); err == nil {
		t.Errorf("expected error for invalid evidence_id")
	}

	// 5. Invalid artifact_ref
	m = baseValid()
	m["artifact_ref"] = "sha256:short"
	raw, _ = json.Marshal(m)
	if err := ValidateEvidence(raw); err == nil {
		t.Errorf("expected error for invalid artifact_ref")
	}

	// 6. Unexpected field
	m = baseValid()
	m["extra"] = "disallowed"
	raw, _ = json.Marshal(m)
	if err := ValidateEvidence(raw); err == nil {
		t.Errorf("expected error for unexpected field")
	}
}

func TestValidateCommonError_Rejections(t *testing.T) {
	const validRef = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	baseValid := func() map[string]any {
		return map[string]any{
			"code":               "ERR_001",
			"message":            "Something went wrong",
			"category":           "USER_ERROR",
			"retryable":          true,
			"recommended_action": "Check user configuration",
			"details":            map[string]any{"field": "input"},
			"artifact_ref":       validRef,
		}
	}

	// 1. Valid baseline passes
	baseBytes, _ := json.Marshal(baseValid())
	if err := ValidateCommonError(baseBytes); err != nil {
		t.Fatalf("expected baseline error to pass, got: %v", err)
	}

	// 2. Missing required fields
	for _, req := range []string{"code", "message", "category", "retryable", "recommended_action"} {
		m := baseValid()
		delete(m, req)
		raw, _ := json.Marshal(m)
		if err := ValidateCommonError(raw); err == nil {
			t.Errorf("expected error for missing field %q", req)
		}
	}

	// 3. Invalid category enum
	m := baseValid()
	m["category"] = "RANDOM_CATEGORY"
	raw, _ := json.Marshal(m)
	if err := ValidateCommonError(raw); err == nil {
		t.Errorf("expected error for invalid category enum")
	}

	// 4. Invalid retryable type
	m = baseValid()
	m["retryable"] = "yes"
	raw, _ = json.Marshal(m)
	if err := ValidateCommonError(raw); err == nil {
		t.Errorf("expected error for string retryable")
	}

	// 5. Unexpected field
	m = baseValid()
	m["extra"] = "disallowed"
	raw, _ = json.Marshal(m)
	if err := ValidateCommonError(raw); err == nil {
		t.Errorf("expected error for unexpected field")
	}
}

func TestValidateCompleteness_Rejections(t *testing.T) {
	baseValid := func() map[string]any {
		return map[string]any{
			"http":           "COMPLETE",
			"body":           "PARTIAL",
			"authentication": "COMPLETE",
			"authorization":  "NOT_APPLICABLE",
			"session":        "COMPLETE",
			"database":       "UNAVAILABLE",
			"external":       "NOT_APPLICABLE",
			"files":          "COMPLETE",
			"cache":          "NOT_APPLICABLE",
			"errors":         "COMPLETE",
			"execution":      "PARTIAL",
		}
	}

	// 1. Valid baseline passes
	baseBytes, _ := json.Marshal(baseValid())
	if err := ValidateCompleteness(baseBytes); err != nil {
		t.Fatalf("expected baseline completeness to pass, got: %v", err)
	}

	// 2. Missing dimension
	m := baseValid()
	delete(m, "database")
	raw, _ := json.Marshal(m)
	if err := ValidateCompleteness(raw); err == nil {
		t.Errorf("expected error for missing database dimension")
	}

	// 3. Invalid completeness state
	m = baseValid()
	m["http"] = "UNKNOWN_STATE"
	raw, _ = json.Marshal(m)
	if err := ValidateCompleteness(raw); err == nil {
		t.Errorf("expected error for invalid completeness state")
	}

	// 4. Unexpected dimension
	m = baseValid()
	m["extra_dim"] = "COMPLETE"
	raw, _ = json.Marshal(m)
	if err := ValidateCompleteness(raw); err == nil {
		t.Errorf("expected error for unexpected dimension")
	}
}

func TestValidateHeaders_Rejections(t *testing.T) {
	baseValid := map[string]any{
		"Content-Type": "application/json",
		"Accept":       []string{"text/html", "application/json"},
	}
	baseBytes, _ := json.Marshal(baseValid)
	if err := ValidateHeaders(baseBytes); err != nil {
		t.Fatalf("expected valid headers to pass, got: %v", err)
	}

	// Null value
	m := map[string]any{"Authorization": nil}
	raw, _ := json.Marshal(m)
	if err := ValidateHeaders(raw); err == nil {
		t.Errorf("expected error for null header value")
	}

	// Number value
	m = map[string]any{"Content-Length": 123}
	raw, _ = json.Marshal(m)
	if err := ValidateHeaders(raw); err == nil {
		t.Errorf("expected error for integer header value")
	}

	// Empty string value
	m = map[string]any{"Authorization": ""}
	raw, _ = json.Marshal(m)
	if err := ValidateHeaders(raw); err == nil {
		t.Errorf("expected error for empty string header value")
	}

	// Empty string in array
	m = map[string]any{"Accept": []string{""}}
	raw, _ = json.Marshal(m)
	if err := ValidateHeaders(raw); err == nil {
		t.Errorf("expected error for empty string item in header array")
	}
}

func TestValidateCommonStringTypes(t *testing.T) {
	// ID
	if err := ValidateCommonID([]byte(`"valid_id-123"`)); err != nil {
		t.Errorf("expected valid ID to pass: %v", err)
	}
	if err := ValidateCommonID([]byte(`"INVALID ID"`)); err == nil {
		t.Errorf("expected invalid ID to fail")
	}

	// Hash
	const validHash = `"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	if err := ValidateCommonHash([]byte(validHash)); err != nil {
		t.Errorf("expected valid Hash to pass: %v", err)
	}
	if err := ValidateCommonHash([]byte(`"tooshort"`)); err == nil {
		t.Errorf("expected invalid Hash to fail")
	}

	// ArtifactRef
	const validRef = `"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	if err := ValidateCommonArtifactRef([]byte(validRef)); err != nil {
		t.Errorf("expected valid ArtifactRef to pass: %v", err)
	}
	if err := ValidateCommonArtifactRef([]byte(`"md5:0123456789abcdef"`)); err == nil {
		t.Errorf("expected invalid ArtifactRef to fail")
	}
}

func TestValidateCommonDefs_PrivacyCanary(t *testing.T) {
	const canary = "CANARY_SECRET_LEAK_COMMON_DEF"

	badInput := `{"visibility":"` + canary + `","policy_ref":"sha256:0000000000000000000000000000000000000000000000000000000000000000"}`
	err := ValidateArtifactPolicy([]byte(badInput))
	if err == nil {
		t.Fatalf("expected error for invalid visibility")
	}
	if strings.Contains(err.Error(), canary) {
		t.Fatalf("error message leaks secret canary: %s", err.Error())
	}
}

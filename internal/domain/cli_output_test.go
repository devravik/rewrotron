package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestCLIOutputRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/cli/cli-output.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCLIOutput(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid CLI output for %s, got: %v", fixture.Name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected invalid CLI output for %s, got nil error", fixture.Name)
			}
		})
	}
	if count != 4 {
		t.Fatalf("expected 4 CLI output fixtures, found %d", count)
	}
}

func cliOutputSuccessControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "envelope-success" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing envelope-success fixture")
	return nil
}

func cliOutputFailureControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "envelope-failure" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing envelope-failure fixture")
	return nil
}

func assertCLIOutputRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var cliErr *CLIOutputError
	if !errors.As(err, &cliErr) || cliErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid CLI output envelope: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func validateCLIOutputControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateCLIOutput(raw)
}

func TestCLIOutputValidRecords(t *testing.T) {
	t.Parallel()
	succ := cliOutputSuccessControl(t)
	if err := validateCLIOutputControl(t, succ); err != nil {
		t.Fatalf("success control record should be valid: %v", err)
	}

	fail := cliOutputFailureControl(t)
	if err := validateCLIOutputControl(t, fail); err != nil {
		t.Fatalf("failure control record should be valid: %v", err)
	}
}

func TestCLIOutputOptionalFields(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["$schema"] = "https://rewrotron.dev/specs/cli/cli-output.schema.json"
	ctrl["dry_run"] = true
	ctrl["dry_run_summary"] = map[string]any{
		"simulated": true,
	}
	if err := validateCLIOutputControl(t, ctrl); err != nil {
		t.Fatalf("record with valid optional fields should be valid: %v", err)
	}
}

func TestCLIOutputMissingRequiredFields(t *testing.T) {
	t.Parallel()
	required := []string{
		"version",
		"command",
		"success",
		"exit_code",
		"timestamp",
		"duration_ms",
		"data",
		"errors",
	}

	for _, field := range required {
		ctrl := cliOutputSuccessControl(t)
		delete(ctrl, field)
		assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "missing required field")
	}
}

func TestCLIOutputExtraFields(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["unexpected"] = "value"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "object fields")
}

func TestCLIOutputInvalidVersion(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["version"] = "2.0"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "unsupported version")

	ctrl = cliOutputSuccessControl(t)
	ctrl["version"] = nil
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "unsupported version")
}

func TestCLIOutputInvalidCommand(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["command"] = "   "
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "command")

	ctrl = cliOutputSuccessControl(t)
	ctrl["command"] = nil
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "command")
}

func TestCLIOutputInvalidSuccess(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["success"] = "true"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "success")

	ctrl = cliOutputSuccessControl(t)
	ctrl["success"] = nil
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "success")
}

func TestCLIOutputInvalidExitCode(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["exit_code"] = 5
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "invalid exit_code")

	ctrl = cliOutputSuccessControl(t)
	ctrl["exit_code"] = -1
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "invalid exit_code")

	ctrl = cliOutputSuccessControl(t)
	ctrl["exit_code"] = "0"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "exit_code")
}

func TestCLIOutputInvalidTimestamp(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["timestamp"] = "not a timestamp"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "invalid timestamp")

	ctrl = cliOutputSuccessControl(t)
	ctrl["timestamp"] = nil
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "timestamp")
}

func TestCLIOutputInvalidDuration(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["duration_ms"] = -1.0
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "duration_ms")

	ctrl = cliOutputSuccessControl(t)
	ctrl["duration_ms"] = "zero"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "duration_ms")
}

func TestCLIOutputInvalidData(t *testing.T) {
	t.Parallel()
	ctrl := cliOutputSuccessControl(t)
	ctrl["data"] = nil
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "data must be an object")

	ctrl = cliOutputSuccessControl(t)
	ctrl["data"] = []any{"not", "an", "object"}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "data must be an object")
}

func TestCLIOutputConditionalExitInvariants(t *testing.T) {
	t.Parallel()

	// Success with exit code 1
	ctrl := cliOutputSuccessControl(t)
	ctrl["exit_code"] = 1
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "successful envelope must have exit_code 0")

	// Success with errors
	ctrl = cliOutputSuccessControl(t)
	ctrl["errors"] = []any{
		map[string]any{
			"code":               "ERR",
			"message":            "msg",
			"category":           "USER_ERROR",
			"retryable":          false,
			"recommended_action": "act",
		},
	}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "successful envelope must have empty errors")

	// Failure with exit code 0
	ctrl = cliOutputFailureControl(t)
	ctrl["exit_code"] = 0
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "failed envelope must have exit_code between 1 and 4")

	// Failure with empty errors
	ctrl = cliOutputFailureControl(t)
	ctrl["errors"] = []any{}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "failed envelope must have non-empty errors")
}

func TestCLIOutputInvalidErrorsArray(t *testing.T) {
	t.Parallel()

	ctrl := cliOutputFailureControl(t)
	ctrl["errors"] = "not an array"
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "errors array")

	ctrl = cliOutputFailureControl(t)
	ctrl["errors"] = []any{"not an object"}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "error object")

	// Missing code
	ctrl = cliOutputFailureControl(t)
	ctrl["errors"] = []any{
		map[string]any{
			"message":            "msg",
			"category":           "USER_ERROR",
			"retryable":          false,
			"recommended_action": "act",
		},
	}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "missing required error field")

	// Invalid category
	ctrl = cliOutputFailureControl(t)
	ctrl["errors"] = []any{
		map[string]any{
			"code":               "ERR",
			"message":            "msg",
			"category":           "UNKNOWN_CATEGORY",
			"retryable":          false,
			"recommended_action": "act",
		},
	}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "error category")

	// Invalid artifact_ref
	ctrl = cliOutputFailureControl(t)
	ctrl["errors"] = []any{
		map[string]any{
			"code":               "ERR",
			"message":            "msg",
			"category":           "USER_ERROR",
			"retryable":          false,
			"recommended_action": "act",
			"artifact_ref":       "invalid-ref",
		},
	}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "error artifact_ref")

	// Extra field in error
	ctrl = cliOutputFailureControl(t)
	ctrl["errors"] = []any{
		map[string]any{
			"code":               "ERR",
			"message":            "msg",
			"category":           "USER_ERROR",
			"retryable":          false,
			"recommended_action": "act",
			"extra_field":        "unauthorized",
		},
	}
	assertCLIOutputRejected(t, validateCLIOutputControl(t, ctrl), "error object fields")
}

func TestCLIOutputJSONParsing(t *testing.T) {
	t.Parallel()

	// Invalid UTF-8
	err := ValidateCLIOutput([]byte{0xff, 0xfe})
	assertCLIOutputRejected(t, err, "invalid UTF-8")

	// Duplicate keys
	dupJSON := []byte(`{
		"version": "1.0",
		"version": "1.0"
	}`)
	assertCLIOutputRejected(t, ValidateCLIOutput(dupJSON), "duplicate object key")

	// Trailing tokens
	ctrl := cliOutputSuccessControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	trailing := append(raw, []byte(` extra`)...)
	assertCLIOutputRejected(t, ValidateCLIOutput(trailing), "trailing input")
}

package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestExecutionStateRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/execution/state.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExecutionState(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid execution state for %s, got: %v", fixture.Name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected invalid execution state for %s, got nil error", fixture.Name)
			}
		})
	}
	if count != 4 {
		t.Fatalf("expected 4 execution state fixtures, found %d", count)
	}
}

func executionStateControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "pause-does-not-replace-project-lifecycle" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing pause-does-not-replace-project-lifecycle fixture")
	return nil
}

func assertExecutionStateRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var stateErr *ExecutionStateError
	if !errors.As(err, &stateErr) || stateErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid execution state: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func validateExecutionStateControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateExecutionState(raw)
}

func TestExecutionStateValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	if err := validateExecutionStateControl(t, ctrl); err != nil {
		t.Fatalf("control record should be valid: %v", err)
	}
}

func TestExecutionStateMissingRequiredFields(t *testing.T) {
	t.Parallel()
	required := []string{
		"schema_version",
		"project_id",
		"lifecycle_state",
		"control_mode",
		"control_reason",
		"latest_event_sequence",
		"observed_at",
		"focus_run_id",
		"active_runs",
		"stage_instances",
		"last_activity_at",
		"last_heartbeat_at",
		"liveness",
		"progress_metrics",
		"blockers",
		"pending_approvals",
		"next_action",
		"user_request",
		"subjects",
		"recent_change_refs",
	}

	for _, field := range required {
		ctrl := executionStateControl(t)
		delete(ctrl, field)
		assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "missing required field")
	}
}

func TestExecutionStateExtraFields(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["unexpected"] = "field"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "object fields")
}

func TestExecutionStateInvalidVersion(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["schema_version"] = "2.0"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "unsupported schema version")
}

func TestExecutionStateInvalidProjectID(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["project_id"] = "Invalid_ID"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "project identity")
}

func TestExecutionStateInvalidLifecycleState(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["lifecycle_state"] = "NONEXISTENT_STATE"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "invalid lifecycle_state")
}

func TestExecutionStateInvalidControlMode(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["control_mode"] = "INVALID_MODE"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "invalid control_mode")
}

func TestExecutionStateInvalidControlReason(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["control_reason"] = "   "
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "control_reason")
}

func TestExecutionStateInvalidLatestEventSequence(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["latest_event_sequence"] = -1
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "latest_event_sequence")
}

func TestExecutionStateInvalidObservedAt(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["observed_at"] = "invalid-time"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "observed_at timestamp")
}

func TestExecutionStateInvalidLiveness(t *testing.T) {
	t.Parallel()
	ctrl := executionStateControl(t)
	ctrl["liveness"] = "INVALID"
	assertExecutionStateRejected(t, validateExecutionStateControl(t, ctrl), "invalid liveness")
}

func TestExecutionStateJSONParsing(t *testing.T) {
	t.Parallel()

	// Invalid UTF-8
	err := ValidateExecutionState([]byte{0xff, 0xfe})
	assertExecutionStateRejected(t, err, "invalid UTF-8")

	// Duplicate keys
	dupJSON := []byte(`{
		"schema_version": "1.0",
		"schema_version": "1.0"
	}`)
	assertExecutionStateRejected(t, ValidateExecutionState(dupJSON), "duplicate object key")

	// Trailing tokens
	ctrl := executionStateControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	trailing := append(raw, []byte(` extra`)...)
	assertExecutionStateRejected(t, ValidateExecutionState(trailing), "trailing input")
}

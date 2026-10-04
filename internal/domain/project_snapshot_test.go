package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestProjectSnapshotRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/snapshot.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateProjectSnapshot(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid project snapshot for %s, got: %v", fixture.Name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected invalid project snapshot for %s, got nil error", fixture.Name)
			}
		})
	}
	if count != 4 {
		t.Fatalf("expected 4 project snapshot fixtures, found %d", count)
	}
}

func projectSnapshotControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "snapshot-empty-denominator" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing snapshot-empty-denominator fixture")
	return nil
}

func assertProjectSnapshotRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var snapErr *ProjectSnapshotError
	if !errors.As(err, &snapErr) || snapErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid project snapshot: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func validateProjectSnapshotControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateProjectSnapshot(raw)
}

func TestProjectSnapshotValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	if err := validateProjectSnapshotControl(t, ctrl); err != nil {
		t.Fatalf("control record should be valid: %v", err)
	}
}

func TestProjectSnapshotMissingRequiredFields(t *testing.T) {
	t.Parallel()
	required := []string{
		"project_id",
		"state",
		"revision",
		"configuration_hash",
		"plan_hash",
		"gate_states",
		"coverage",
		"blockers",
		"latest_operation_id",
	}

	for _, field := range required {
		ctrl := projectSnapshotControl(t)
		delete(ctrl, field)
		assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "missing required field")
	}
}

func TestProjectSnapshotExtraFields(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["extra"] = "forbidden"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "object fields")
}

func TestProjectSnapshotInvalidProjectID(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["project_id"] = "Invalid_ID"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "project identity")

	ctrl = projectSnapshotControl(t)
	ctrl["project_id"] = nil
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "project identity")
}

func TestProjectSnapshotInvalidState(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["state"] = "PARITY_ACHIEVED"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "invalid state")

	ctrl = projectSnapshotControl(t)
	ctrl["state"] = nil
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "invalid state")
}

func TestProjectSnapshotInvalidRevision(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["revision"] = -1
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "revision")

	ctrl = projectSnapshotControl(t)
	ctrl["revision"] = "zero"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "revision")
}

func TestProjectSnapshotInvalidHashes(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["configuration_hash"] = "short"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "configuration_hash")

	ctrl = projectSnapshotControl(t)
	ctrl["plan_hash"] = "invalid-hash"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "plan_hash")
}

func TestProjectSnapshotInvalidGateStates(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["gate_states"] = nil
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "gate_states must be an object")

	ctrl = projectSnapshotControl(t)
	gates := ctrl["gate_states"].(map[string]any)
	delete(gates, "GATE_1_SOURCE_DETECTION")
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "missing required gate")

	ctrl = projectSnapshotControl(t)
	gates = ctrl["gate_states"].(map[string]any)
	gates["GATE_1_SOURCE_DETECTION"] = "INVALID_STATUS"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "gate status")
}

func TestProjectSnapshotInvalidCoverage(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["coverage"] = nil
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "coverage must be an object")

	// Missing dimension
	ctrl = projectSnapshotControl(t)
	cov := ctrl["coverage"].(map[string]any)
	delete(cov, "cases")
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "missing required coverage dimension")

	// Empty denominator with non-zero numerator
	ctrl = projectSnapshotControl(t)
	cov = ctrl["coverage"].(map[string]any)
	cov["cases"] = map[string]any{
		"numerator":   5,
		"denominator": 0,
		"percent":     nil,
	}
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "cases zero denominator requires zero numerator")

	// Empty denominator with non-null percent
	ctrl = projectSnapshotControl(t)
	cov = ctrl["coverage"].(map[string]any)
	cov["cases"] = map[string]any{
		"numerator":   0,
		"denominator": 0,
		"percent":     100,
	}
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "cases zero denominator requires null percent")
}

func TestProjectSnapshotInvalidLatestOperationID(t *testing.T) {
	t.Parallel()
	ctrl := projectSnapshotControl(t)
	ctrl["latest_operation_id"] = "Invalid_Op"
	assertProjectSnapshotRejected(t, validateProjectSnapshotControl(t, ctrl), "latest_operation_id")
}

func TestProjectSnapshotJSONParsing(t *testing.T) {
	t.Parallel()

	// Invalid UTF-8
	err := ValidateProjectSnapshot([]byte{0xff, 0xfe})
	assertProjectSnapshotRejected(t, err, "invalid UTF-8")

	// Duplicate keys
	dupJSON := []byte(`{
		"project_id": "test",
		"project_id": "test2"
	}`)
	assertProjectSnapshotRejected(t, ValidateProjectSnapshot(dupJSON), "duplicate object key")

	// Trailing tokens
	ctrl := projectSnapshotControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	trailing := append(raw, []byte(` extra`)...)
	assertProjectSnapshotRejected(t, ValidateProjectSnapshot(trailing), "trailing input")
}

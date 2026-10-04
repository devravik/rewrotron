package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestOperationResultRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/operations/result.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateOperationResult(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid operation result for %s, got: %v", fixture.Name, err)
				}
				return
			}
			t.Fatalf("unexpected invalid fixture in repository: %s", fixture.Name)
		})
	}
	if count != 1 {
		t.Fatalf("expected 1 operation result fixture, found %d", count)
	}
}

func operationResultControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "v2-admission" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing v2-admission fixture")
	return nil
}

func assertOperationResultRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var opErr *OperationResultError
	if !errors.As(err, &opErr) || opErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid operation result: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func validateOperationResultControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return err
	}
	return ValidateOperationResult(canonical)
}

func TestOperationResultValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	if err := validateOperationResultControl(t, ctrl); err != nil {
		t.Fatalf("control record should be valid: %v", err)
	}
}

func TestOperationResultNullStateAndStatuses(t *testing.T) {
	t.Parallel()

	statuses := []string{"COMPLETED", "FAILED", "CANCELLED", "PREPARED", "RUNNING"}
	for _, status := range statuses {
		ctrl := operationResultControl(t)
		ctrl["status"] = status
		ctrl["state"] = nil
		if err := validateOperationResultControl(t, ctrl); err != nil {
			t.Fatalf("status %s with null state should be valid: %v", status, err)
		}
	}
}

func TestOperationResultLifecycleStates(t *testing.T) {
	t.Parallel()

	states := []string{
		"INITIALIZED",
		"DETECTED",
		"SOURCE_CONFIRMED",
		"ANALYZED",
		"MODELED",
		"PLAN_SYNTHESIZED",
		"PLAN_CONFIRMED",
		"MIGRATING",
		"VERIFIED",
		"BENCHMARKED",
		"READY_CUTOVER",
	}

	for _, st := range states {
		ctrl := operationResultControl(t)
		ctrl["state"] = st
		if err := validateOperationResultControl(t, ctrl); err != nil {
			t.Fatalf("lifecycle state %s should be valid: %v", st, err)
		}
	}
}

func TestOperationResultPopulatedData(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["artifact_refs"] = []any{
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	ctrl["entity_ids"] = []any{"ent_1", "ent_2"}
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "finding-1",
			"code":         "SYNTAX_ERR",
			"severity":     "CRITICAL",
			"affected_ids": []any{"mod-auth"},
			"evidence_ids": []any{"ev-1"},
			"message":      "Critical syntax error in entrypoint",
		},
		map[string]any{
			"finding_id":   "finding-2",
			"code":         "DEPRECATED_CALL",
			"severity":     "LOW",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "Deprecated function call observed",
		},
	}
	ctrl["summary"] = map[string]any{
		"processed_items": 42,
		"details": map[string]any{
			"nested": "ok",
		},
	}

	if err := validateOperationResultControl(t, ctrl); err != nil {
		t.Fatalf("populated record should be valid: %v", err)
	}
}

func TestOperationResultMissingRequiredFields(t *testing.T) {
	t.Parallel()
	required := []string{
		"operation_id",
		"status",
		"state",
		"artifact_refs",
		"entity_ids",
		"blockers",
		"summary",
	}

	for _, field := range required {
		ctrl := operationResultControl(t)
		delete(ctrl, field)
		assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "missing required field")
	}
}

func TestOperationResultExtraFields(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["unexpected"] = "extra"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "object fields")
}

func TestOperationResultInvalidOperationID(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["operation_id"] = "Invalid_Upper"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "operation identity")

	ctrl["operation_id"] = "has space"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "operation identity")

	ctrl["operation_id"] = ""
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "operation identity")
}

func TestOperationResultInvalidStatus(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["status"] = "UNKNOWN"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid status")

	ctrl["status"] = ""
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid status")
}

func TestOperationResultInvalidState(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["state"] = "NONEXISTENT_STATE"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid state")

	ctrl["state"] = 123
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid state")
}

func TestOperationResultInvalidArtifactRefs(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["artifact_refs"] = "not an array"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "artifact_refs array")

	ctrl = operationResultControl(t)
	ctrl["artifact_refs"] = []any{"invalid_hash"}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid artifact_refs reference")

	ctrl = operationResultControl(t)
	ctrl["artifact_refs"] = []any{
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "duplicate artifact_refs reference")
}

func TestOperationResultInvalidEntityIDs(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["entity_ids"] = "not an array"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "entity_ids array")

	ctrl = operationResultControl(t)
	ctrl["entity_ids"] = []any{"Invalid_Entity"}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid entity_ids identity")

	ctrl = operationResultControl(t)
	ctrl["entity_ids"] = []any{"ent_1", "ent_1"}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "duplicate entity_ids identity")
}

func TestOperationResultInvalidBlockers(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["blockers"] = "not an array"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "blockers array")

	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{"not an object"}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "blocker object")

	// Missing blocker field
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id": "f1",
			"code":       "C1",
			"severity":   "HIGH",
			// missing affected_ids, evidence_ids, message
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "missing required blocker field")

	// Extra blocker field
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "msg",
			"extra":        "unauthorized",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "blocker object fields")

	// Invalid blocker finding_id
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "INVALID ID",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid blocker finding_id")

	// Duplicate blocker finding_id
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "msg",
		},
		map[string]any{
			"finding_id":   "f1",
			"code":         "C2",
			"severity":     "LOW",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "msg2",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "duplicate blocker finding_id")

	// Empty blocker code
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "   ",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid blocker code")

	// Invalid blocker severity
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "VERY_HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid blocker severity")

	// Invalid affected_ids
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": "not array",
			"evidence_ids": []any{},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "blocker affected_ids array")

	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{"INVALID"},
			"evidence_ids": []any{},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid blocker affected_ids identity")

	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{"id-1", "id-1"},
			"evidence_ids": []any{},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "duplicate blocker affected_ids identity")

	// Invalid evidence_ids
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": "not array",
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "blocker evidence_ids array")

	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{"id-1", "id-1"},
			"message":      "msg",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "duplicate blocker evidence_ids identity")

	// Empty blocker message
	ctrl = operationResultControl(t)
	ctrl["blockers"] = []any{
		map[string]any{
			"finding_id":   "f1",
			"code":         "C1",
			"severity":     "HIGH",
			"affected_ids": []any{},
			"evidence_ids": []any{},
			"message":      "  ",
		},
	}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "invalid blocker message")
}

func TestOperationResultInvalidSummary(t *testing.T) {
	t.Parallel()
	ctrl := operationResultControl(t)
	ctrl["summary"] = nil
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "summary must be an object")

	ctrl = operationResultControl(t)
	ctrl["summary"] = []any{"not", "an", "object"}
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "summary must be an object")

	ctrl = operationResultControl(t)
	ctrl["summary"] = "string is not object"
	assertOperationResultRejected(t, validateOperationResultControl(t, ctrl), "summary must be an object")
}

func TestOperationResultNonCanonicalJSON(t *testing.T) {
	t.Parallel()
	raw := []byte("not valid json at all")
	assertOperationResultRejected(t, ValidateOperationResult(raw), "canonical JSON")
}

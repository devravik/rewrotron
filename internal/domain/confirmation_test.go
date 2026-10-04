package domain

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func validConfirmationReceiptControl() map[string]any {
	return map[string]any{
		"confirmation_id":      "confirm-src-1",
		"project_id":           "fixture",
		"gate":                 "GATE_1_SOURCE_DETECTION",
		"subject_hash":         strings.Repeat("a", 64),
		"configuration_hash":   strings.Repeat("b", 64),
		"confirmed_by":         "operator@example.com",
		"channel":              "CLI",
		"action_id":            "action-1",
		"reviewed_payload_ref": "sha256:" + strings.Repeat("c", 64),
		"created_at":           "2026-10-04T12:00:00Z",
		"status":               "SOURCE_CONFIRMED",
	}
}

func assertConfirmationRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var confErr *ConfirmationError
	if !errors.As(err, &confErr) || confErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid confirmation receipt: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestConfirmationReceiptValid(t *testing.T) {
	t.Parallel()

	gates := []struct {
		gate   string
		status string
	}{
		{"GATE_1_SOURCE_DETECTION", "SOURCE_CONFIRMED"},
		{"GATE_2_PLAN_CONFIRMED", "PLAN_CONFIRMED"},
		{"GATE_3_CUTOVER", "READY_CUTOVER"},
	}

	for _, tc := range gates {
		tc := tc
		t.Run(tc.gate, func(t *testing.T) {
			t.Parallel()
			ctrl := validConfirmationReceiptControl()
			ctrl["gate"] = tc.gate
			ctrl["status"] = tc.status
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := CanonicalJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateConfirmation(canonical); err != nil {
				t.Fatalf("expected valid confirmation for %s, got: %v", tc.gate, err)
			}
		})
	}
}

func TestConfirmationReceiptMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		reason  string
		rawJSON string
	}{
		{
			name:    "invalid-json",
			rawJSON: `{"confirmation_id": "confirm-1",`,
			reason:  "canonical JSON",
		},
		{
			name:    "non-canonical-float",
			rawJSON: `{"confirmation_id":"c1","float_val":12.34}`,
			reason:  "canonical JSON",
		},
		{
			name: "extra-root-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
		{
			name: "missing-confirmation-id",
			mutate: func(m map[string]any) {
				delete(m, "confirmation_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-confirmation-id",
			mutate: func(m map[string]any) {
				m["confirmation_id"] = "CONFIRM_UPPER"
			},
			reason: "confirmation identity",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "PROJ_UPPER"
			},
			reason: "project identity",
		},
		{
			name: "invalid-action-id",
			mutate: func(m map[string]any) {
				m["action_id"] = "ACTION_UPPER"
			},
			reason: "action identity",
		},
		{
			name: "invalid-subject-hash",
			mutate: func(m map[string]any) {
				m["subject_hash"] = "short"
			},
			reason: "subject_hash",
		},
		{
			name: "invalid-configuration-hash",
			mutate: func(m map[string]any) {
				m["configuration_hash"] = "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
			},
			reason: "configuration_hash",
		},
		{
			name: "invalid-reviewed-payload-ref",
			mutate: func(m map[string]any) {
				m["reviewed_payload_ref"] = "not-sha256"
			},
			reason: "reviewed_payload_ref",
		},
		{
			name: "empty-confirmed-by",
			mutate: func(m map[string]any) {
				m["confirmed_by"] = "   "
			},
			reason: "confirmed_by",
		},
		{
			name: "invalid-channel",
			mutate: func(m map[string]any) {
				m["channel"] = "WEB"
			},
			reason: "invalid channel",
		},
		{
			name: "invalid-created-at",
			mutate: func(m map[string]any) {
				m["created_at"] = "not-a-date"
			},
			reason: "invalid created_at timestamp",
		},
		{
			name: "mismatched-gate-status-1-2",
			mutate: func(m map[string]any) {
				m["gate"] = "GATE_1_SOURCE_DETECTION"
				m["status"] = "PLAN_CONFIRMED"
			},
			reason: "gate and status mismatch",
		},
		{
			name: "mismatched-gate-status-2-3",
			mutate: func(m map[string]any) {
				m["gate"] = "GATE_2_PLAN_CONFIRMED"
				m["status"] = "READY_CUTOVER"
			},
			reason: "gate and status mismatch",
		},
		{
			name: "mismatched-gate-status-3-1",
			mutate: func(m map[string]any) {
				m["gate"] = "GATE_3_CUTOVER"
				m["status"] = "SOURCE_CONFIRMED"
			},
			reason: "gate and status mismatch",
		},
		{
			name: "invalid-gate",
			mutate: func(m map[string]any) {
				m["gate"] = "GATE_4_DEPLOY"
			},
			reason: "invalid gate",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var raw []byte
			if tc.rawJSON != "" {
				raw = []byte(tc.rawJSON)
			} else {
				ctrl := validConfirmationReceiptControl()
				tc.mutate(ctrl)
				data, err := json.Marshal(ctrl)
				if err != nil {
					t.Fatal(err)
				}
				canon, err := CanonicalJSON(data)
				if err != nil {
					t.Fatal(err)
				}
				raw = canon
			}

			err := ValidateConfirmation(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertConfirmationRejected(t, err, tc.reason)
		})
	}
}

func TestPreparedConfirmationV1(t *testing.T) {
	t.Parallel()

	v1Control := func() map[string]any {
		return map[string]any{
			"project_id":           "fixture",
			"detection_id":         "detect-1",
			"subject_hash":         strings.Repeat("a", 64),
			"merged_facts":         map[string]any{"type": "php"},
			"reviewed_payload_ref": "sha256:" + strings.Repeat("b", 64),
		}
	}

	t.Run("valid-v1", func(t *testing.T) {
		t.Parallel()
		raw, err := json.Marshal(v1Control())
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidatePreparedConfirmation(canon); err != nil {
			t.Fatalf("expected valid prepared confirmation v1, got: %v", err)
		}
	})

	t.Run("v1-missing-field", func(t *testing.T) {
		t.Parallel()
		ctrl := v1Control()
		delete(ctrl, "merged_facts")
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		err = ValidatePreparedConfirmation(canon)
		assertConfirmationRejected(t, err, "missing required field")
	})

	t.Run("v1-extra-field", func(t *testing.T) {
		t.Parallel()
		ctrl := v1Control()
		ctrl["extra"] = "forbidden"
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		err = ValidatePreparedConfirmation(canon)
		assertConfirmationRejected(t, err, "object fields")
	})
}

func TestPreparedConfirmationV2(t *testing.T) {
	t.Parallel()

	subject := map[string]any{
		"schema_version":   "2.0",
		"canonicalization": "rwt-cjson-1",
		"project_id":       "fixture",
		"kind":             "SOURCE",
		"snapshot_ref":     "sha256:" + strings.Repeat("a", 64),
		"detection_ref":    "sha256:" + strings.Repeat("a", 64),
		"assertions_ref":   "sha256:" + strings.Repeat("a", 64),
		"adapter_bindings": []any{
			map[string]any{
				"role":               "SOURCE",
				"adapter_id":         "synthetic",
				"version":            "1.0.0",
				"manifest_ref":       "sha256:" + strings.Repeat("a", 64),
				"implementation_ref": "sha256:" + strings.Repeat("a", 64),
			},
		},
	}
	subjectRaw, err := json.Marshal(subject)
	if err != nil {
		t.Fatal(err)
	}
	subjectCanon, err := CanonicalJSON(subjectRaw)
	if err != nil {
		t.Fatal(err)
	}
	subHash, err := CanonicalHash(subjectCanon)
	if err != nil {
		t.Fatal(err)
	}

	v2Control := func() map[string]any {
		return map[string]any{
			"schema_version":         "2.0",
			"project_id":             "fixture",
			"detection_id":           "detect-1",
			"subject_hash":           subHash,
			"measured_detection_ref": "sha256:" + strings.Repeat("b", 64),
			"assertions_ref":         "sha256:" + strings.Repeat("c", 64),
			"reviewed_payload_ref":   "sha256:" + strings.Repeat("d", 64),
			"subject":                subject,
		}
	}

	t.Run("valid-v2", func(t *testing.T) {
		t.Parallel()
		raw, err := json.Marshal(v2Control())
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidatePreparedConfirmation(canon); err != nil {
			t.Fatalf("expected valid prepared confirmation v2, got: %v", err)
		}
	})

	t.Run("v2-subject-hash-mismatch", func(t *testing.T) {
		t.Parallel()
		ctrl := v2Control()
		ctrl["subject_hash"] = strings.Repeat("f", 64)
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		err = ValidatePreparedConfirmation(canon)
		assertConfirmationRejected(t, err, "subject hash mismatch")
	})

	t.Run("v2-invalid-subject-shape", func(t *testing.T) {
		t.Parallel()
		ctrl := v2Control()
		badSubject := map[string]any{
			"schema_version": "1.0",
		}
		ctrl["subject"] = badSubject
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		err = ValidatePreparedConfirmation(canon)
		assertConfirmationRejected(t, err, "subject shape")
	})

	t.Run("v2-unsupported-version", func(t *testing.T) {
		t.Parallel()
		ctrl := v2Control()
		ctrl["schema_version"] = "3.0"
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		canon, err := CanonicalJSON(raw)
		if err != nil {
			t.Fatal(err)
		}
		err = ValidatePreparedConfirmation(canon)
		assertConfirmationRejected(t, err, "unsupported schema version")
	})
}

func TestValidateApprovalSubject_ValidationCases(t *testing.T) {
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

	targetSchema := "https://rewrotron.dev/specs/project/approval-subject.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateApprovalSubject(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid approval subject fixture %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid approval subject fixture %q to fail validation, but it passed", tc.Name)
			}
		})
	}

	if count < 4 {
		t.Fatalf("expected at least 4 approval subject validation cases, found %d", count)
	}
}

package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestRecoveryJournalRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasons := map[string]string{
		"review2-emergency-cannot-certify": "unsupported recovery journal phase",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/operations/recovery-journal.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRecoveryJournal(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid recovery journal for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, exists := reasons[fixture.Name]
			if !exists {
				t.Fatalf("unexpected negative fixture without mapped reason: %s", fixture.Name)
			}
			assertRecoveryJournalRejected(t, err, expectedReason)
		})
	}
	if count != 2 {
		t.Fatalf("expected 2 recovery journal fixtures, found %d", count)
	}
}

func recoveryJournalControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "review2-emergency-uncommitted-state" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing review2-emergency-uncommitted-state fixture")
	return nil
}

func assertRecoveryJournalRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var journalErr *RecoveryJournalError
	if !errors.As(err, &journalErr) || journalErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid recovery journal: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestRecoveryJournalPositiveControls(t *testing.T) {
	t.Parallel()

	t.Run("fixture-control", func(t *testing.T) {
		t.Parallel()
		ctrl := recoveryJournalControl(t)
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateRecoveryJournal(raw); err != nil {
			t.Fatalf("expected valid recovery journal, got: %v", err)
		}
	})

	t.Run("receipt-durable-phase", func(t *testing.T) {
		t.Parallel()
		ctrl := recoveryJournalControl(t)
		ctrl["phase"] = "RECEIPT_DURABLE"
		ctrl["receipt_ref"] = "sha256:" + strings.Repeat("r", 64) // wait, r is not hex, use 'a'
		ctrl["receipt_ref"] = "sha256:" + strings.Repeat("a", 64)
		ctrl["reason_code"] = nil
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateRecoveryJournal(raw); err != nil {
			t.Fatalf("expected valid RECEIPT_DURABLE journal, got: %v", err)
		}
	})

	t.Run("reserved-initial-slot", func(t *testing.T) {
		t.Parallel()
		ctrl := recoveryJournalControl(t)
		ctrl["phase"] = "RESERVED"
		ctrl["sequence"] = int64(0)
		ctrl["last_committed_event_sequence"] = int64(0)
		ctrl["effect_attempt_id"] = nil
		ctrl["receipt_ref"] = nil
		ctrl["reason_code"] = nil
		ctrl["previous_record_hash"] = nil
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateRecoveryJournal(raw); err != nil {
			t.Fatalf("expected valid initial RESERVED slot, got: %v", err)
		}
	})

	t.Run("reconciled-phase", func(t *testing.T) {
		t.Parallel()
		ctrl := recoveryJournalControl(t)
		ctrl["phase"] = "RECONCILED"
		ctrl["reason_code"] = nil
		raw, err := json.Marshal(ctrl)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateRecoveryJournal(raw); err != nil {
			t.Fatalf("expected valid RECONCILED journal, got: %v", err)
		}
	})
}

func TestRecoveryJournalNegativeMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]any)
		reason string
	}{
		{
			name: "wrong-schema-version",
			mutate: func(m map[string]any) {
				m["schema_version"] = "2.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "missing-project-id",
			mutate: func(m map[string]any) {
				delete(m, "project_id")
			},
			reason: "object fields",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "Invalid_Project!"
			},
			reason: "project_id",
		},
		{
			name: "invalid-operation-id",
			mutate: func(m map[string]any) {
				m["operation_id"] = "Invalid_Op!"
			},
			reason: "operation_id",
		},
		{
			name: "invalid-attempt-id",
			mutate: func(m map[string]any) {
				m["attempt_id"] = "Invalid_Attempt!"
			},
			reason: "attempt_id",
		},
		{
			name: "negative-sequence",
			mutate: func(m map[string]any) {
				m["sequence"] = int64(-1)
			},
			reason: "sequence",
		},
		{
			name: "negative-last-committed-sequence",
			mutate: func(m map[string]any) {
				m["last_committed_event_sequence"] = int64(-1)
			},
			reason: "last_committed_event_sequence",
		},
		{
			name: "unsupported-phase",
			mutate: func(m map[string]any) {
				m["phase"] = "PASS"
			},
			reason: "unsupported recovery journal phase",
		},
		{
			name: "invalid-effect-attempt-id",
			mutate: func(m map[string]any) {
				m["effect_attempt_id"] = "Invalid_Effect!"
			},
			reason: "effect_attempt_id",
		},
		{
			name: "receipt-durable-missing-receipt-ref",
			mutate: func(m map[string]any) {
				m["phase"] = "RECEIPT_DURABLE"
				m["receipt_ref"] = nil
			},
			reason: "durable receipt requires receipt reference",
		},
		{
			name: "invalid-receipt-ref-format",
			mutate: func(m map[string]any) {
				m["receipt_ref"] = "bad-ref"
			},
			reason: "receipt_ref",
		},
		{
			name: "negative-reservation-bytes",
			mutate: func(m map[string]any) {
				m["reservation_bytes"] = int64(-1)
			},
			reason: "reservation_bytes",
		},
		{
			name: "persistence-failed-missing-reason-code",
			mutate: func(m map[string]any) {
				m["phase"] = "PERSISTENCE_FAILED"
				m["reason_code"] = nil
			},
			reason: "persistence failure requires reason code",
		},
		{
			name: "empty-reason-code",
			mutate: func(m map[string]any) {
				m["reason_code"] = ""
			},
			reason: "reason_code",
		},
		{
			name: "invalid-previous-hash-format",
			mutate: func(m map[string]any) {
				m["previous_record_hash"] = "not-a-valid-hex-hash"
			},
			reason: "previous_record_hash",
		},
		{
			name: "extra-top-level-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := recoveryJournalControl(t)
			tc.mutate(ctrl)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateRecoveryJournal(raw)
			assertRecoveryJournalRejected(t, err, tc.reason)
		})
	}
}

func TestRecoveryJournalCanonicalJSON(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"float_number", []byte(`{"schema_version":"1.0","num":1.5}`)},
		{"duplicate_keys", []byte(`{"schema_version":"1.0","schema_version":"1.0"}`)},
		{"invalid_utf8", []byte{0xff, 0xfe}},
		{"trailing_data", []byte(`{"schema_version":"1.0"}{}`)},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRecoveryJournal(tc.raw)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var journalErr *RecoveryJournalError
			if !errors.As(err, &journalErr) || journalErr.Reason != "canonical JSON" {
				t.Fatalf("expected canonical JSON error, got %v", err)
			}
		})
	}
}

func TestRecoveryJournalErrorSecurity(t *testing.T) {
	t.Parallel()
	ctrl := recoveryJournalControl(t)
	ctrl["project_id"] = "CANARY_SECRET_LEAK!"
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateRecoveryJournal(raw)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errStr := err.Error()
	if strings.Contains(errStr, "CANARY_SECRET_LEAK") {
		t.Fatalf("error message discloses sensitive input: %s", errStr)
	}
	if errStr != "invalid recovery journal: project_id" {
		t.Fatalf("unexpected error format: %s", errStr)
	}
}

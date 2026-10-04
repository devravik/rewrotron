package domain

import (
	"encoding/json"
)

// RecoveryJournalError describes an invalid recovery-journal record without echoing private contents.
type RecoveryJournalError struct {
	Reason string
	Cause  error
}

// Error identifies the failed recovery-journal obligation.
func (e *RecoveryJournalError) Error() string {
	return "invalid recovery journal: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *RecoveryJournalError) Unwrap() error { return e.Cause }

// ValidateRecoveryJournal validates a recovery-journal 1.0 record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, identifier patterns, sequence constraints, lifecycle phases,
// nullable references, reservation bytes, and failure reason codes.
//
// This check does not execute filesystem allocation, sync recovery slots,
// or verify device durability.
func ValidateRecoveryJournal(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &RecoveryJournalError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &RecoveryJournalError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"operation_id",
		"attempt_id",
		"sequence",
		"last_committed_event_sequence",
		"phase",
		"effect_attempt_id",
		"receipt_ref",
		"reservation_bytes",
		"reason_code",
		"previous_record_hash",
	}

	if len(record) != len(expectedFields) {
		return &RecoveryJournalError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &RecoveryJournalError{Reason: "missing required field"}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &RecoveryJournalError{Reason: "unsupported schema version"}
	}

	for _, field := range []string{"project_id", "operation_id", "attempt_id"} {
		val := profileString(record[field])
		if !validAssessmentReportID(val) {
			return &RecoveryJournalError{Reason: field}
		}
	}

	var sequence int64
	if string(record["sequence"]) == "null" || json.Unmarshal(record["sequence"], &sequence) != nil || sequence < 0 {
		return &RecoveryJournalError{Reason: "sequence"}
	}

	var lastCommittedSeq int64
	if string(record["last_committed_event_sequence"]) == "null" || json.Unmarshal(record["last_committed_event_sequence"], &lastCommittedSeq) != nil || lastCommittedSeq < 0 {
		return &RecoveryJournalError{Reason: "last_committed_event_sequence"}
	}

	phase := profileString(record["phase"])
	switch phase {
	case "RESERVED", "MEASUREMENT_STARTED", "EFFECT_REQUESTED", "RECEIPT_DURABLE",
		"PERSISTENCE_FAILED", "CLEANUP_PENDING", "RECONCILED":
	default:
		return &RecoveryJournalError{Reason: "unsupported recovery journal phase"}
	}

	effectAttemptRaw := record["effect_attempt_id"]
	if string(effectAttemptRaw) != "null" {
		effectAttempt := profileString(effectAttemptRaw)
		if !validAssessmentReportID(effectAttempt) {
			return &RecoveryJournalError{Reason: "effect_attempt_id"}
		}
	}

	receiptRaw := record["receipt_ref"]
	if phase == "RECEIPT_DURABLE" && string(receiptRaw) == "null" {
		return &RecoveryJournalError{Reason: "durable receipt requires receipt reference"}
	}
	if !validProfileRef(receiptRaw, true) {
		return &RecoveryJournalError{Reason: "receipt_ref"}
	}

	var reservationBytes int64
	if string(record["reservation_bytes"]) == "null" || json.Unmarshal(record["reservation_bytes"], &reservationBytes) != nil || reservationBytes < 0 {
		return &RecoveryJournalError{Reason: "reservation_bytes"}
	}

	reasonRaw := record["reason_code"]
	if phase == "PERSISTENCE_FAILED" && string(reasonRaw) == "null" {
		return &RecoveryJournalError{Reason: "persistence failure requires reason code"}
	}
	if string(reasonRaw) != "null" {
		var reasonCode string
		if json.Unmarshal(reasonRaw, &reasonCode) != nil || len(reasonCode) < 1 {
			return &RecoveryJournalError{Reason: "reason_code"}
		}
	}

	prevHashRaw := record["previous_record_hash"]
	if string(prevHashRaw) != "null" {
		prevHash := profileString(prevHashRaw)
		if !isManifestDigest(prevHash) {
			return &RecoveryJournalError{Reason: "previous_record_hash"}
		}
	}

	return nil
}

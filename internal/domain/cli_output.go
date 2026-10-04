package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

var allowedCLIErrorCategories = map[string]struct{}{
	"USER_ERROR":         {},
	"ENVIRONMENT_ERROR":  {},
	"SOURCE_ERROR":       {},
	"TARGET_BUILD_ERROR": {},
	"ANALYSIS_ERROR":     {},
	"OBSERVATION_ERROR":  {},
	"VERIFICATION_ERROR": {},
	"AGENT_ERROR":        {},
	"INTERNAL_ERROR":     {},
}

// CLIOutputError describes an invalid CLI output envelope without echoing private contents.
type CLIOutputError struct {
	Reason string
	Cause  error
}

// Error identifies the failed CLI output envelope obligation.
func (e *CLIOutputError) Error() string {
	return "invalid CLI output envelope: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *CLIOutputError) Unwrap() error { return e.Cause }

// ValidateCLIOutput validates a CLI output envelope conforming to specs/cli/cli-output.schema.json.
// It verifies schema version 1.0, required envelope fields, valid command and timestamps,
// duration bounds, exit code and error correspondence, structured errors matching
// specs/common.schema.json#/$defs/error, and conditional exit invariants.
func ValidateCLIOutput(raw []byte) error {
	record, err := parseCLIOutputJSON(raw)
	if err != nil {
		return err
	}

	requiredFields := []string{
		"version",
		"command",
		"success",
		"exit_code",
		"timestamp",
		"duration_ms",
		"data",
		"errors",
	}

	allowedOptionalFields := map[string]struct{}{
		"$schema":         {},
		"dry_run":         {},
		"dry_run_summary": {},
	}

	for _, field := range requiredFields {
		if _, present := record[field]; !present {
			return &CLIOutputError{Reason: "missing required field"}
		}
	}

	for field := range record {
		isReq := false
		for _, req := range requiredFields {
			if field == req {
				isReq = true
				break
			}
		}
		if isReq {
			continue
		}
		if _, allowed := allowedOptionalFields[field]; !allowed {
			return &CLIOutputError{Reason: "object fields"}
		}
	}

	if schemaRaw, present := record["$schema"]; present {
		var schemaURI string
		if string(schemaRaw) == "null" || json.Unmarshal(schemaRaw, &schemaURI) != nil || strings.TrimSpace(schemaURI) == "" {
			return &CLIOutputError{Reason: "$schema"}
		}
	}

	var version string
	if string(record["version"]) == "null" || json.Unmarshal(record["version"], &version) != nil || version != "1.0" {
		return &CLIOutputError{Reason: "unsupported version"}
	}

	var command string
	if string(record["command"]) == "null" || json.Unmarshal(record["command"], &command) != nil || strings.TrimSpace(command) == "" {
		return &CLIOutputError{Reason: "command"}
	}

	var success bool
	if string(record["success"]) == "null" || json.Unmarshal(record["success"], &success) != nil {
		return &CLIOutputError{Reason: "success"}
	}

	var exitCode int64
	if string(record["exit_code"]) == "null" || json.Unmarshal(record["exit_code"], &exitCode) != nil {
		return &CLIOutputError{Reason: "exit_code"}
	}
	if exitCode < 0 || exitCode > 4 {
		return &CLIOutputError{Reason: "invalid exit_code"}
	}

	var timestamp string
	if string(record["timestamp"]) == "null" || json.Unmarshal(record["timestamp"], &timestamp) != nil {
		return &CLIOutputError{Reason: "timestamp"}
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		return &CLIOutputError{Reason: "invalid timestamp", Cause: err}
	}

	var durationMs float64
	if string(record["duration_ms"]) == "null" || json.Unmarshal(record["duration_ms"], &durationMs) != nil || durationMs < 0 {
		return &CLIOutputError{Reason: "duration_ms"}
	}

	if string(record["data"]) == "null" {
		return &CLIOutputError{Reason: "data must be an object"}
	}
	var dataObj map[string]json.RawMessage
	if err := json.Unmarshal(record["data"], &dataObj); err != nil || dataObj == nil {
		return &CLIOutputError{Reason: "data must be an object", Cause: err}
	}

	if dryRunRaw, present := record["dry_run"]; present {
		var dryRun bool
		if string(dryRunRaw) == "null" || json.Unmarshal(dryRunRaw, &dryRun) != nil {
			return &CLIOutputError{Reason: "dry_run"}
		}
	}

	if summaryRaw, present := record["dry_run_summary"]; present {
		if string(summaryRaw) == "null" {
			return &CLIOutputError{Reason: "dry_run_summary must be an object"}
		}
		var summaryObj map[string]json.RawMessage
		if err := json.Unmarshal(summaryRaw, &summaryObj); err != nil || summaryObj == nil {
			return &CLIOutputError{Reason: "dry_run_summary must be an object", Cause: err}
		}
	}

	errorsList, err := validateCLIErrors(record["errors"])
	if err != nil {
		return err
	}

	// allOf conditional rules:
	if success {
		if exitCode != 0 {
			return &CLIOutputError{Reason: "successful envelope must have exit_code 0"}
		}
		if len(errorsList) != 0 {
			return &CLIOutputError{Reason: "successful envelope must have empty errors"}
		}
	} else {
		if exitCode < 1 || exitCode > 4 {
			return &CLIOutputError{Reason: "failed envelope must have exit_code between 1 and 4"}
		}
		if len(errorsList) == 0 {
			return &CLIOutputError{Reason: "failed envelope must have non-empty errors"}
		}
	}

	return nil
}

func validateCLIErrors(raw json.RawMessage) ([]json.RawMessage, error) {
	var errItems []json.RawMessage
	if err := json.Unmarshal(raw, &errItems); err != nil {
		return nil, &CLIOutputError{Reason: "errors array", Cause: err}
	}

	reqErrFields := []string{
		"code",
		"message",
		"category",
		"retryable",
		"recommended_action",
	}

	optErrFields := map[string]struct{}{
		"details":      {},
		"artifact_ref": {},
	}

	for _, rawErr := range errItems {
		var errObj map[string]json.RawMessage
		if err := json.Unmarshal(rawErr, &errObj); err != nil || errObj == nil {
			return nil, &CLIOutputError{Reason: "error object", Cause: err}
		}

		for _, req := range reqErrFields {
			if _, present := errObj[req]; !present {
				return nil, &CLIOutputError{Reason: "missing required error field"}
			}
		}

		for field := range errObj {
			isReq := false
			for _, req := range reqErrFields {
				if field == req {
					isReq = true
					break
				}
			}
			if isReq {
				continue
			}
			if _, allowed := optErrFields[field]; !allowed {
				return nil, &CLIOutputError{Reason: "error object fields"}
			}
		}

		var code string
		if string(errObj["code"]) == "null" || json.Unmarshal(errObj["code"], &code) != nil || strings.TrimSpace(code) == "" {
			return nil, &CLIOutputError{Reason: "error code"}
		}

		var message string
		if string(errObj["message"]) == "null" || json.Unmarshal(errObj["message"], &message) != nil || strings.TrimSpace(message) == "" {
			return nil, &CLIOutputError{Reason: "error message"}
		}

		var category string
		if string(errObj["category"]) == "null" || json.Unmarshal(errObj["category"], &category) != nil {
			return nil, &CLIOutputError{Reason: "error category"}
		}
		if _, ok := allowedCLIErrorCategories[category]; !ok {
			return nil, &CLIOutputError{Reason: "error category"}
		}

		var retryable bool
		if string(errObj["retryable"]) == "null" || json.Unmarshal(errObj["retryable"], &retryable) != nil {
			return nil, &CLIOutputError{Reason: "error retryable"}
		}

		var recAction string
		if string(errObj["recommended_action"]) == "null" || json.Unmarshal(errObj["recommended_action"], &recAction) != nil || strings.TrimSpace(recAction) == "" {
			return nil, &CLIOutputError{Reason: "error recommended_action"}
		}

		if detailsRaw, present := errObj["details"]; present && string(detailsRaw) != "null" {
			var detailsObj map[string]json.RawMessage
			if err := json.Unmarshal(detailsRaw, &detailsObj); err != nil || detailsObj == nil {
				return nil, &CLIOutputError{Reason: "error details must be an object", Cause: err}
			}
		}

		if artRefRaw, present := errObj["artifact_ref"]; present && string(artRefRaw) != "null" {
			if !validProfileRef(artRefRaw, false) {
				return nil, &CLIOutputError{Reason: "error artifact_ref"}
			}
		}
	}

	return errItems, nil
}

func parseCLIOutputJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &CLIOutputError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &CLIOutputError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &CLIOutputError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &CLIOutputError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &CLIOutputError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &CLIOutputError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &CLIOutputError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &CLIOutputError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &CLIOutputError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &CLIOutputError{Reason: "trailing input"}
	}
	return record, nil
}

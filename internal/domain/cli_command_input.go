package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

// CLICommandInputError describes an invalid CLI command input argument without echoing private contents.
type CLICommandInputError struct {
	Command string
	Reason  string
	Cause   error
}

// Error identifies the failed CLI command input obligation.
func (e *CLICommandInputError) Error() string {
	if e.Command != "" {
		return "invalid CLI input for " + e.Command + ": " + e.Reason
	}
	return "invalid CLI input: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *CLICommandInputError) Unwrap() error { return e.Cause }

// ValidateCLICommandInput validates a command input record against its schema URI in specs/cli/commands/*.
func ValidateCLICommandInput(schemaURI string, raw []byte) error {
	switch schemaURI {
	case "https://rewrotron.dev/specs/cli/commands/report.input.schema.json":
		return ValidateCLIReportInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/assess.input.schema.json":
		return ValidateCLIAssessInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/assessment-show.input.schema.json":
		return ValidateCLIAssessmentShowInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/assessment-decide.input.schema.json":
		return ValidateCLIAssessmentDecideInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/status.input.schema.json":
		return ValidateCLIStatusInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/execution-show.input.schema.json":
		return ValidateCLIExecutionShowInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/pause.input.schema.json":
		return ValidateCLIPauseInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/continue.input.schema.json":
		return ValidateCLIContinueInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/plan.input.schema.json":
		return ValidateCLIPlanInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/migrate.input.schema.json":
		return ValidateCLIMigrateInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/approval-issue.input.schema.json":
		return ValidateCLIApprovalIssueInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/confirm-source.input.schema.json":
		return ValidateCLIConfirmSourceInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/confirm-plan.input.schema.json":
		return ValidateCLIConfirmPlanInput(raw)
	case "https://rewrotron.dev/specs/cli/commands/confirm-cutover.input.schema.json":
		return ValidateCLIConfirmCutoverInput(raw)
	default:
		return &CLICommandInputError{Reason: "unsupported command input schema: " + schemaURI}
	}
}

// ValidateCLIReportInput validates input conforming to specs/cli/commands/report.input.schema.json.
func ValidateCLIReportInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "report")
	if err != nil {
		return err
	}

	formatRaw, present := record["format"]
	if !present || string(formatRaw) == "null" {
		return &CLICommandInputError{Command: "report", Reason: "missing required field: format"}
	}
	var format string
	if err := json.Unmarshal(formatRaw, &format); err != nil || (format != "json" && format != "markdown") {
		return &CLICommandInputError{Command: "report", Reason: "format must be json or markdown"}
	}

	var kind string = "auto"
	if kindRaw, present := record["kind"]; present && string(kindRaw) != "null" {
		if err := json.Unmarshal(kindRaw, &kind); err != nil || (kind != "auto" && kind != "assessment" && kind != "migration") {
			return &CLICommandInputError{Command: "report", Reason: "invalid kind enum"}
		}
	}

	hasAssessmentRef := false
	if refRaw, present := record["assessment_ref"]; present && string(refRaw) != "null" {
		hasAssessmentRef = true
		if !validProfileRef(refRaw, false) {
			return &CLICommandInputError{Command: "report", Reason: "invalid assessment_ref"}
		}
	}

	for k := range record {
		if k != "format" && k != "kind" && k != "assessment_ref" {
			return &CLICommandInputError{Command: "report", Reason: "unexpected field: " + k}
		}
	}

	if hasAssessmentRef && kind == "migration" {
		return &CLICommandInputError{Command: "report", Reason: "assessment_ref cannot be used with kind=migration"}
	}

	return nil
}

// ValidateCLIAssessInput validates input conforming to specs/cli/commands/assess.input.schema.json.
func ValidateCLIAssessInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "assess")
	if err != nil {
		return err
	}

	scopeRaw, present := record["scope_ref"]
	if !present || string(scopeRaw) == "null" || !validProfileRef(scopeRaw, false) {
		return &CLICommandInputError{Command: "assess", Reason: "invalid scope_ref"}
	}

	if obsRaw, present := record["observation"]; present && string(obsRaw) != "null" {
		var obs string
		if err := json.Unmarshal(obsRaw, &obs); err != nil || (obs != "AVAILABLE_ONLY" && obs != "SAFE_RUNTIME") {
			return &CLICommandInputError{Command: "assess", Reason: "invalid observation enum"}
		}
	}

	if bgRaw, present := record["background"]; present && string(bgRaw) != "null" {
		var bg bool
		if err := json.Unmarshal(bgRaw, &bg); err != nil {
			return &CLICommandInputError{Command: "assess", Reason: "background must be a boolean"}
		}
	}

	for k := range record {
		if k != "scope_ref" && k != "observation" && k != "background" {
			return &CLICommandInputError{Command: "assess", Reason: "unexpected field: " + k}
		}
	}

	return nil
}

// ValidateCLIAssessmentShowInput validates input conforming to specs/cli/commands/assessment-show.input.schema.json.
func ValidateCLIAssessmentShowInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "assessment show")
	if err != nil {
		return err
	}

	refRaw, present := record["assessment_ref"]
	if !present || string(refRaw) == "null" || !validProfileRef(refRaw, false) {
		return &CLICommandInputError{Command: "assessment show", Reason: "invalid assessment_ref"}
	}

	if len(record) != 1 {
		return &CLICommandInputError{Command: "assessment show", Reason: "unexpected fields"}
	}

	return nil
}

// ValidateCLIAssessmentDecideInput validates input conforming to specs/cli/commands/assessment-decide.input.schema.json.
func ValidateCLIAssessmentDecideInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "assessment decide")
	if err != nil {
		return err
	}

	refRaw, present := record["assessment_ref"]
	if !present || string(refRaw) == "null" || !validProfileRef(refRaw, false) {
		return &CLICommandInputError{Command: "assessment decide", Reason: "invalid assessment_ref"}
	}

	actionRaw, present := record["action"]
	if !present || string(actionRaw) == "null" {
		return &CLICommandInputError{Command: "assessment decide", Reason: "missing action"}
	}
	var action string
	if err := json.Unmarshal(actionRaw, &action); err != nil {
		return &CLICommandInputError{Command: "assessment decide", Reason: "action must be a string", Cause: err}
	}

	allowedActions := map[string]struct{}{
		"STOP":                   {},
		"REQUEST_EVIDENCE":       {},
		"CHANGE_TARGET":          {},
		"CHANGE_SCOPE":           {},
		"SELECT_WORK":            {},
		"ACCEPT_RECOMMENDATIONS": {},
		"REJECT_RECOMMENDATIONS": {},
		"CONTINUE_PLANNING":      {},
	}
	if _, ok := allowedActions[action]; !ok {
		return &CLICommandInputError{Command: "assessment decide", Reason: "invalid action enum: " + action}
	}

	ratRaw, present := record["rationale"]
	if !present || string(ratRaw) == "null" {
		return &CLICommandInputError{Command: "assessment decide", Reason: "missing rationale"}
	}
	var rationale string
	if err := json.Unmarshal(ratRaw, &rationale); err != nil || strings.TrimSpace(rationale) == "" {
		return &CLICommandInputError{Command: "assessment decide", Reason: "rationale must be a non-empty string"}
	}

	allowedFields := map[string]struct{}{
		"assessment_ref":     {},
		"action":             {},
		"rationale":          {},
		"selected_scope_ref": {},
		"target_intent_ref":  {},
		"finding_ids":        {},
		"decision_refs":      {},
		"requested_evidence": {},
	}
	for k := range record {
		if _, ok := allowedFields[k]; !ok {
			return &CLICommandInputError{Command: "assessment decide", Reason: "unexpected field: " + k}
		}
	}

	// Action-specific allOf rules:
	switch action {
	case "CONTINUE_PLANNING":
		rawSel, hasSel := record["selected_scope_ref"]
		if !hasSel || string(rawSel) == "null" || !validProfileRef(rawSel, false) {
			return &CLICommandInputError{Command: "assessment decide", Reason: "CONTINUE_PLANNING requires selected_scope_ref"}
		}
		rawTgt, hasTgt := record["target_intent_ref"]
		if !hasTgt || string(rawTgt) == "null" || !validProfileRef(rawTgt, false) {
			return &CLICommandInputError{Command: "assessment decide", Reason: "CONTINUE_PLANNING requires target_intent_ref"}
		}
	case "SELECT_WORK", "CHANGE_SCOPE":
		rawSel, hasSel := record["selected_scope_ref"]
		if !hasSel || string(rawSel) == "null" || !validProfileRef(rawSel, false) {
			return &CLICommandInputError{Command: "assessment decide", Reason: action + " requires selected_scope_ref"}
		}
	case "CHANGE_TARGET":
		rawTgt, hasTgt := record["target_intent_ref"]
		if !hasTgt || string(rawTgt) == "null" || !validProfileRef(rawTgt, false) {
			return &CLICommandInputError{Command: "assessment decide", Reason: "CHANGE_TARGET requires target_intent_ref"}
		}
	case "ACCEPT_RECOMMENDATIONS", "REJECT_RECOMMENDATIONS":
		if err := validateNonEmptyStringArray(record["finding_ids"], "finding_ids", "assessment decide"); err != nil {
			return err
		}
	case "REQUEST_EVIDENCE":
		if err := validateNonEmptyStringArray(record["requested_evidence"], "requested_evidence", "assessment decide"); err != nil {
			return err
		}
	}

	if rawSel, hasSel := record["selected_scope_ref"]; hasSel && string(rawSel) != "null" {
		if !validProfileRef(rawSel, false) {
			return &CLICommandInputError{Command: "assessment decide", Reason: "invalid selected_scope_ref"}
		}
	}
	if rawTgt, hasTgt := record["target_intent_ref"]; hasTgt && string(rawTgt) != "null" {
		if !validProfileRef(rawTgt, false) {
			return &CLICommandInputError{Command: "assessment decide", Reason: "invalid target_intent_ref"}
		}
	}
	if rawDec, hasDec := record["decision_refs"]; hasDec && string(rawDec) != "null" {
		if err := validateRefArray(rawDec, "decision_refs"); err != nil {
			return &CLICommandInputError{Command: "assessment decide", Reason: "invalid decision_refs"}
		}
	}

	return nil
}

// ValidateCLIStatusInput validates input conforming to specs/cli/commands/status.input.schema.json.
func ValidateCLIStatusInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "status")
	if err != nil {
		return err
	}

	if wRaw, present := record["watch"]; present && string(wRaw) != "null" {
		var w bool
		if err := json.Unmarshal(wRaw, &w); err != nil {
			return &CLICommandInputError{Command: "status", Reason: "watch must be a boolean"}
		}
	}

	if vRaw, present := record["verbosity"]; present && string(vRaw) != "null" {
		var v string
		if err := json.Unmarshal(vRaw, &v); err != nil || (v != "simple" && v != "detailed" && v != "developer") {
			return &CLICommandInputError{Command: "status", Reason: "invalid verbosity enum"}
		}
	}

	for k := range record {
		if k != "watch" && k != "verbosity" {
			return &CLICommandInputError{Command: "status", Reason: "unexpected field: " + k}
		}
	}

	return nil
}

// ValidateCLIExecutionShowInput validates input conforming to specs/cli/commands/execution-show.input.schema.json.
func ValidateCLIExecutionShowInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "execution show")
	if err != nil {
		return err
	}

	if aRaw, present := record["after_sequence"]; present && string(aRaw) != "null" {
		var a int64
		if err := json.Unmarshal(aRaw, &a); err != nil || a < 0 {
			return &CLICommandInputError{Command: "execution show", Reason: "after_sequence must be an integer >= 0"}
		}
	}

	if tRaw, present := record["through_sequence"]; present && string(tRaw) != "null" {
		var t int64
		if err := json.Unmarshal(tRaw, &t); err != nil || t < 0 {
			return &CLICommandInputError{Command: "execution show", Reason: "through_sequence must be an integer >= 0"}
		}
	}

	if lRaw, present := record["limit"]; present && string(lRaw) != "null" {
		var l int64
		if err := json.Unmarshal(lRaw, &l); err != nil || l < 1 || l > 500 {
			return &CLICommandInputError{Command: "execution show", Reason: "limit must be an integer between 1 and 500"}
		}
	}

	if vRaw, present := record["verbosity"]; present && string(vRaw) != "null" {
		var v string
		if err := json.Unmarshal(vRaw, &v); err != nil || (v != "simple" && v != "detailed" && v != "developer") {
			return &CLICommandInputError{Command: "execution show", Reason: "invalid verbosity enum"}
		}
	}

	for k := range record {
		if k != "after_sequence" && k != "through_sequence" && k != "limit" && k != "verbosity" {
			return &CLICommandInputError{Command: "execution show", Reason: "unexpected field: " + k}
		}
	}

	return nil
}

// ValidateCLIPauseInput validates input conforming to specs/cli/commands/pause.input.schema.json.
func ValidateCLIPauseInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "pause")
	if err != nil {
		return err
	}
	if len(record) != 0 {
		return &CLICommandInputError{Command: "pause", Reason: "unexpected fields"}
	}
	return nil
}

// ValidateCLIContinueInput validates input conforming to specs/cli/commands/continue.input.schema.json.
func ValidateCLIContinueInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "continue")
	if err != nil {
		return err
	}
	if len(record) != 0 {
		return &CLICommandInputError{Command: "continue", Reason: "unexpected fields"}
	}
	return nil
}

// ValidateCLIPlanInput validates input conforming to specs/cli/commands/plan.input.schema.json.
func ValidateCLIPlanInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "plan")
	if err != nil {
		return err
	}

	if tRaw, present := record["target"]; present && string(tRaw) != "null" {
		var t string
		if err := json.Unmarshal(tRaw, &t); err != nil || strings.TrimSpace(t) == "" {
			return &CLICommandInputError{Command: "plan", Reason: "target must be a non-empty string"}
		}
	}

	if sRaw, present := record["style"]; present && string(sRaw) != "null" {
		var s string
		if err := json.Unmarshal(sRaw, &s); err != nil || strings.TrimSpace(s) == "" {
			return &CLICommandInputError{Command: "plan", Reason: "style must be a non-empty string"}
		}
	}

	if rRaw, present := record["recalculate"]; present && string(rRaw) != "null" {
		var r bool
		if err := json.Unmarshal(rRaw, &r); err != nil {
			return &CLICommandInputError{Command: "plan", Reason: "recalculate must be a boolean"}
		}
	}

	if selRaw, present := record["selection_ref"]; present && string(selRaw) != "null" {
		if !validProfileRef(selRaw, false) {
			return &CLICommandInputError{Command: "plan", Reason: "invalid selection_ref"}
		}
	}

	if perfRaw, present := record["performance_policy_ref"]; present && string(perfRaw) != "null" {
		if !validProfileRef(perfRaw, false) {
			return &CLICommandInputError{Command: "plan", Reason: "invalid performance_policy_ref"}
		}
	}

	for k := range record {
		if k != "target" && k != "style" && k != "recalculate" && k != "selection_ref" && k != "performance_policy_ref" {
			return &CLICommandInputError{Command: "plan", Reason: "unexpected field: " + k}
		}
	}

	return nil
}

// ValidateCLIMigrateInput validates input conforming to specs/cli/commands/migrate.input.schema.json.
func ValidateCLIMigrateInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "migrate")
	if err != nil {
		return err
	}

	for k := range record {
		if k != "unit_id" && k != "all" && k != "background" {
			return &CLICommandInputError{Command: "migrate", Reason: "unexpected field: " + k}
		}
	}

	hasUnit := false
	if uRaw, present := record["unit_id"]; present && string(uRaw) != "null" {
		hasUnit = true
		var u string
		if err := json.Unmarshal(uRaw, &u); err != nil || !validAssessmentReportID(u) {
			return &CLICommandInputError{Command: "migrate", Reason: "invalid unit_id"}
		}
	}

	hasAll := false
	if aRaw, present := record["all"]; present {
		if hasUnit {
			return &CLICommandInputError{Command: "migrate", Reason: "cannot specify both unit_id and all"}
		}
		var a bool
		if err := json.Unmarshal(aRaw, &a); err != nil || !a {
			return &CLICommandInputError{Command: "migrate", Reason: "all must be boolean true"}
		}
		hasAll = true
	}

	if bgRaw, present := record["background"]; present && string(bgRaw) != "null" {
		var bg bool
		if err := json.Unmarshal(bgRaw, &bg); err != nil {
			return &CLICommandInputError{Command: "migrate", Reason: "background must be a boolean"}
		}
	}

	if !hasUnit && !hasAll {
		return &CLICommandInputError{Command: "migrate", Reason: "exactly one of unit_id or all=true must be provided"}
	}

	return nil
}

// ValidateCLIApprovalIssueInput validates input conforming to specs/cli/commands/approval-issue.input.schema.json.
func ValidateCLIApprovalIssueInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "approval issue")
	if err != nil {
		return err
	}

	gateRaw, present := record["gate"]
	if !present || string(gateRaw) == "null" {
		return &CLICommandInputError{Command: "approval issue", Reason: "missing required field: gate"}
	}
	var gate string
	if err := json.Unmarshal(gateRaw, &gate); err != nil || (gate != "source" && gate != "plan" && gate != "cutover") {
		return &CLICommandInputError{Command: "approval issue", Reason: "gate must be source, plan, or cutover"}
	}

	hashRaw, present := record["subject_hash"]
	if !present || string(hashRaw) == "null" {
		return &CLICommandInputError{Command: "approval issue", Reason: "missing required field: subject_hash"}
	}
	var hash string
	if err := json.Unmarshal(hashRaw, &hash); err != nil || !isManifestDigest(hash) {
		return &CLICommandInputError{Command: "approval issue", Reason: "invalid subject_hash"}
	}

	if len(record) != 2 {
		return &CLICommandInputError{Command: "approval issue", Reason: "unexpected fields"}
	}

	return nil
}

// ValidateCLIConfirmSourceInput validates input conforming to specs/cli/commands/confirm-source.input.schema.json.
func ValidateCLIConfirmSourceInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "confirm source")
	if err != nil {
		return err
	}

	detRaw, present := record["detection_id"]
	if !present || string(detRaw) == "null" {
		return &CLICommandInputError{Command: "confirm source", Reason: "missing required field: detection_id"}
	}
	var detID string
	if err := json.Unmarshal(detRaw, &detID); err != nil || !validAssessmentReportID(detID) {
		return &CLICommandInputError{Command: "confirm source", Reason: "invalid detection_id"}
	}

	manRaw, present := record["manifest_hash"]
	if !present || string(manRaw) == "null" {
		return &CLICommandInputError{Command: "confirm source", Reason: "missing required field: manifest_hash"}
	}
	var manHash string
	if err := json.Unmarshal(manRaw, &manHash); err != nil || !isManifestDigest(manHash) {
		return &CLICommandInputError{Command: "confirm source", Reason: "invalid manifest_hash"}
	}

	if subRaw, present := record["subject_hash"]; present && string(subRaw) != "null" {
		var subHash string
		if err := json.Unmarshal(subRaw, &subHash); err != nil || !isManifestDigest(subHash) {
			return &CLICommandInputError{Command: "confirm source", Reason: "invalid subject_hash"}
		}
	}

	if ovRaw, present := record["overrides"]; present && string(ovRaw) != "null" {
		var ov map[string]string
		if err := json.Unmarshal(ovRaw, &ov); err != nil || ov == nil {
			return &CLICommandInputError{Command: "confirm source", Reason: "overrides must be an object of strings"}
		}
		for _, v := range ov {
			if strings.TrimSpace(v) == "" {
				return &CLICommandInputError{Command: "confirm source", Reason: "override value cannot be empty"}
			}
		}
	}

	if pRaw, present := record["prepare"]; present && string(pRaw) != "null" {
		var p bool
		if err := json.Unmarshal(pRaw, &p); err != nil {
			return &CLICommandInputError{Command: "confirm source", Reason: "prepare must be a boolean"}
		}
	}

	for k := range record {
		if k != "detection_id" && k != "manifest_hash" && k != "subject_hash" && k != "overrides" && k != "prepare" {
			return &CLICommandInputError{Command: "confirm source", Reason: "unexpected field: " + k}
		}
	}

	return nil
}

// ValidateCLIConfirmPlanInput validates input conforming to specs/cli/commands/confirm-plan.input.schema.json.
func ValidateCLIConfirmPlanInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "confirm plan")
	if err != nil {
		return err
	}

	subRaw, present := record["subject_hash"]
	if !present || string(subRaw) == "null" {
		return &CLICommandInputError{Command: "confirm plan", Reason: "missing required field: subject_hash"}
	}
	var subHash string
	if err := json.Unmarshal(subRaw, &subHash); err != nil || !isManifestDigest(subHash) {
		return &CLICommandInputError{Command: "confirm plan", Reason: "invalid subject_hash"}
	}

	if len(record) != 1 {
		return &CLICommandInputError{Command: "confirm plan", Reason: "unexpected fields"}
	}

	return nil
}

// ValidateCLIConfirmCutoverInput validates input conforming to specs/cli/commands/confirm-cutover.input.schema.json.
func ValidateCLIConfirmCutoverInput(raw []byte) error {
	record, err := parseCLICommandInputJSON(raw, "confirm cutover")
	if err != nil {
		return err
	}

	subRaw, present := record["subject_hash"]
	if !present || string(subRaw) == "null" {
		return &CLICommandInputError{Command: "confirm cutover", Reason: "missing required field: subject_hash"}
	}
	var subHash string
	if err := json.Unmarshal(subRaw, &subHash); err != nil || !isManifestDigest(subHash) {
		return &CLICommandInputError{Command: "confirm cutover", Reason: "invalid subject_hash"}
	}

	if len(record) != 1 {
		return &CLICommandInputError{Command: "confirm cutover", Reason: "unexpected fields"}
	}

	return nil
}

func validNullableRef(raw json.RawMessage) bool {
	if raw == nil || string(raw) == "null" {
		return true
	}
	return validProfileRef(raw, false)
}

func validateNonEmptyStringArray(raw json.RawMessage, fieldName, cmd string) error {
	if raw == nil || string(raw) == "null" {
		return &CLICommandInputError{Command: cmd, Reason: fieldName + " is required"}
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err != nil || len(list) == 0 {
		return &CLICommandInputError{Command: cmd, Reason: fieldName + " must be a non-empty array"}
	}
	seen := make(map[string]struct{}, len(list))
	for _, item := range list {
		if strings.TrimSpace(item) == "" {
			return &CLICommandInputError{Command: cmd, Reason: "empty string in " + fieldName}
		}
		if _, exists := seen[item]; exists {
			return &CLICommandInputError{Command: cmd, Reason: "duplicate item in " + fieldName}
		}
		seen[item] = struct{}{}
	}
	return nil
}

func parseCLICommandInputJSON(raw []byte, command string) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &CLICommandInputError{Command: command, Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &CLICommandInputError{Command: command, Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &CLICommandInputError{Command: command, Reason: "object expected"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &CLICommandInputError{Command: command, Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &CLICommandInputError{Command: command, Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &CLICommandInputError{Command: command, Reason: "duplicate object key: " + key}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &CLICommandInputError{Command: command, Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &CLICommandInputError{Command: command, Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &CLICommandInputError{Command: command, Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &CLICommandInputError{Command: command, Reason: "trailing input"}
	}
	return record, nil
}

package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

var allowedProjectEnvironmentModes = map[string]struct{}{
	"docker": {},
	"local":  {},
}

var allowedProjectExportPolicies = map[string]struct{}{
	"local_only":      {},
	"reviewed_export": {},
}

var allowedProjectCompatibilityRequirements = map[string]struct{}{
	"strict":           {},
	"approved_changes": {},
}

var allowedProjectSessionContinuity = map[string]struct{}{
	"required":          {},
	"approved_relogin": {},
}

// ProjectConfigError describes an invalid project configuration record without echoing private contents.
type ProjectConfigError struct {
	Reason string
	Cause  error
}

// Error identifies the failed project config obligation.
func (e *ProjectConfigError) Error() string {
	return "invalid project config: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ProjectConfigError) Unwrap() error { return e.Cause }

// ValidateProjectConfig validates a project configuration record conforming to
// specs/project/project.schema.json.
// It verifies version 1, project identity, source specification with containment and exclusion rules,
// environment configuration (docker or local), safety invariant guarantees (production_access="denied",
// external_writes="denied"), optional target specification conforming to ValidateTargetEnvironment,
// optional compatibility policies, and optional vertical slice migration settings (preserving dormant code
// and deferring optimization until after parity).
func ValidateProjectConfig(raw []byte) error {
	record, err := parseProjectConfigJSON(raw)
	if err != nil {
		return err
	}

	requiredFields := []string{
		"version",
		"project_id",
		"source",
		"environment",
		"safety",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &ProjectConfigError{Reason: "missing required field: " + f}
		}
	}

	allowedOptionalFields := map[string]struct{}{
		"target":        {},
		"compatibility": {},
		"migration":     {},
	}

	reqMap := make(map[string]struct{}, len(requiredFields))
	for _, f := range requiredFields {
		reqMap[f] = struct{}{}
	}

	for k := range record {
		if _, req := reqMap[k]; req {
			continue
		}
		if _, opt := allowedOptionalFields[k]; opt {
			continue
		}
		return &ProjectConfigError{Reason: "unexpected field: " + k}
	}

	var version int64
	if string(record["version"]) == "null" || json.Unmarshal(record["version"], &version) != nil || version != 1 {
		return &ProjectConfigError{Reason: "version must be 1"}
	}
	if strings.ContainsAny(string(record["version"]), ".eE") {
		return &ProjectConfigError{Reason: "version must be integer"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &ProjectConfigError{Reason: "invalid project_id"}
	}

	if err := validateProjectConfigSource(record["source"]); err != nil {
		return err
	}

	if err := validateProjectConfigEnvironment(record["environment"]); err != nil {
		return err
	}

	if err := validateProjectConfigSafety(record["safety"]); err != nil {
		return err
	}

	if val, present := record["target"]; present && string(val) != "null" {
		if err := ValidateTargetEnvironment(val); err != nil {
			return &ProjectConfigError{Reason: "invalid target specification", Cause: err}
		}
	}

	if val, present := record["compatibility"]; present && string(val) != "null" {
		if err := validateProjectConfigCompatibility(val); err != nil {
			return err
		}
	}

	if val, present := record["migration"]; present && string(val) != "null" {
		if err := validateProjectConfigMigration(val); err != nil {
			return err
		}
	}

	return nil
}

func validateProjectConfigSource(raw json.RawMessage) error {
	var src map[string]json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &src) != nil || src == nil {
		return &ProjectConfigError{Reason: "source must be an object"}
	}

	requiredSourceFields := []string{"type", "path"}
	for _, req := range requiredSourceFields {
		if _, present := src[req]; !present || string(src[req]) == "null" {
			return &ProjectConfigError{Reason: "missing required source field: " + req}
		}
	}

	for k := range src {
		if k != "type" && k != "path" && k != "runtime" && k != "exclusions" {
			return &ProjectConfigError{Reason: "unexpected field in source: " + k}
		}
	}

	var srcType string
	if json.Unmarshal(src["type"], &srcType) != nil || len(srcType) < 1 {
		return &ProjectConfigError{Reason: "invalid source type"}
	}

	var path string
	if json.Unmarshal(src["path"], &path) != nil || len(path) < 1 {
		return &ProjectConfigError{Reason: "invalid source path"}
	}

	if rt, present := src["runtime"]; present && string(rt) != "null" {
		var rtObj map[string]json.RawMessage
		if json.Unmarshal(rt, &rtObj) != nil || rtObj == nil {
			return &ProjectConfigError{Reason: "source runtime must be an object"}
		}
		for k := range rtObj {
			if k != "version" && k != "image_digest" && k != "secret_refs" {
				return &ProjectConfigError{Reason: "unexpected field in source runtime: " + k}
			}
		}
		if v, present := rtObj["version"]; present && string(v) != "null" {
			var ver string
			if json.Unmarshal(v, &ver) != nil || len(ver) < 1 {
				return &ProjectConfigError{Reason: "invalid source runtime version"}
			}
		}
		if img, present := rtObj["image_digest"]; present && string(img) != "null" {
			var image string
			if json.Unmarshal(img, &image) != nil || len(image) < 1 {
				return &ProjectConfigError{Reason: "invalid source runtime image_digest"}
			}
		}
		if sr, present := rtObj["secret_refs"]; present && string(sr) != "null" {
			var secrets map[string]string
			if json.Unmarshal(sr, &secrets) != nil || secrets == nil {
				return &ProjectConfigError{Reason: "invalid source runtime secret_refs"}
			}
			for _, ref := range secrets {
				if !envSecretRefRegex.MatchString(ref) {
					return &ProjectConfigError{Reason: "invalid secret_refs value format"}
				}
			}
		}
	}

	if ex, present := src["exclusions"]; present && string(ex) != "null" {
		var exclusions []string
		if json.Unmarshal(ex, &exclusions) != nil {
			return &ProjectConfigError{Reason: "source exclusions must be an array"}
		}
		for _, e := range exclusions {
			if len(e) < 1 {
				return &ProjectConfigError{Reason: "source exclusions item must not be empty"}
			}
		}
	}
	return nil
}

func validateProjectConfigEnvironment(raw json.RawMessage) error {
	var env map[string]json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &env) != nil || env == nil {
		return &ProjectConfigError{Reason: "environment must be an object"}
	}
	if _, present := env["mode"]; !present || string(env["mode"]) == "null" {
		return &ProjectConfigError{Reason: "environment mode is required"}
	}
	for k := range env {
		if k != "mode" && k != "database" {
			return &ProjectConfigError{Reason: "unexpected field in environment: " + k}
		}
	}
	var mode string
	if json.Unmarshal(env["mode"], &mode) != nil {
		return &ProjectConfigError{Reason: "invalid environment mode"}
	}
	if _, ok := allowedProjectEnvironmentModes[mode]; !ok {
		return &ProjectConfigError{Reason: "invalid environment mode enum: " + mode}
	}

	if db, present := env["database"]; present && string(db) != "null" {
		var dbObj map[string]json.RawMessage
		if json.Unmarshal(db, &dbObj) != nil || dbObj == nil {
			return &ProjectConfigError{Reason: "environment database must be an object"}
		}
		for k := range dbObj {
			if k != "engine" && k != "version" && k != "schema_policy" && k != "secret_ref" && k != "fixture_ref" {
				return &ProjectConfigError{Reason: "unexpected field in environment database: " + k}
			}
		}
		if eng, present := dbObj["engine"]; present && string(eng) != "null" {
			var engine string
			if json.Unmarshal(eng, &engine) != nil || len(engine) < 1 {
				return &ProjectConfigError{Reason: "invalid database engine"}
			}
		}
		if v, present := dbObj["version"]; present && string(v) != "null" {
			var ver string
			if json.Unmarshal(v, &ver) != nil || len(ver) < 1 {
				return &ProjectConfigError{Reason: "invalid database version"}
			}
		}
		if sp, present := dbObj["schema_policy"]; present && string(sp) != "null" {
			var policy string
			if json.Unmarshal(sp, &policy) != nil || policy != "retain" {
				return &ProjectConfigError{Reason: "database schema_policy must be retain"}
			}
		}
		if sr, present := dbObj["secret_ref"]; present && string(sr) != "null" {
			var secretRef string
			if json.Unmarshal(sr, &secretRef) != nil || len(secretRef) < 1 {
				return &ProjectConfigError{Reason: "invalid database secret_ref"}
			}
		}
		if fr, present := dbObj["fixture_ref"]; present && string(fr) != "null" {
			if !validProfileRef(fr, false) {
				return &ProjectConfigError{Reason: "invalid database fixture_ref"}
			}
		}
	}
	return nil
}

func validateProjectConfigSafety(raw json.RawMessage) error {
	var safety map[string]json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &safety) != nil || safety == nil {
		return &ProjectConfigError{Reason: "safety must be an object"}
	}
	requiredSafety := []string{"production_access", "external_writes", "export_policy"}
	for _, req := range requiredSafety {
		if _, present := safety[req]; !present || string(safety[req]) == "null" {
			return &ProjectConfigError{Reason: "missing required safety field: " + req}
		}
	}
	if len(safety) != 3 {
		return &ProjectConfigError{Reason: "unexpected field in safety"}
	}

	var prodAccess string
	if json.Unmarshal(safety["production_access"], &prodAccess) != nil || prodAccess != "denied" {
		return &ProjectConfigError{Reason: "safety production_access must be denied"}
	}

	var extWrites string
	if json.Unmarshal(safety["external_writes"], &extWrites) != nil || extWrites != "denied" {
		return &ProjectConfigError{Reason: "safety external_writes must be denied"}
	}

	var expPolicy string
	if json.Unmarshal(safety["export_policy"], &expPolicy) != nil {
		return &ProjectConfigError{Reason: "invalid safety export_policy"}
	}
	if _, ok := allowedProjectExportPolicies[expPolicy]; !ok {
		return &ProjectConfigError{Reason: "invalid safety export_policy enum"}
	}
	return nil
}

func validateProjectConfigCompatibility(raw json.RawMessage) error {
	var comp map[string]json.RawMessage
	if json.Unmarshal(raw, &comp) != nil || comp == nil {
		return &ProjectConfigError{Reason: "compatibility must be an object"}
	}
	for k := range comp {
		if k != "requirement" && k != "active_session_continuity" && k != "policy_ref" {
			return &ProjectConfigError{Reason: "unexpected field in compatibility: " + k}
		}
	}
	if req, present := comp["requirement"]; present && string(req) != "null" {
		var r string
		if json.Unmarshal(req, &r) != nil {
			return &ProjectConfigError{Reason: "invalid compatibility requirement"}
		}
		if _, ok := allowedProjectCompatibilityRequirements[r]; !ok {
			return &ProjectConfigError{Reason: "invalid compatibility requirement enum"}
		}
	}
	if asc, present := comp["active_session_continuity"]; present && string(asc) != "null" {
		var s string
		if json.Unmarshal(asc, &s) != nil {
			return &ProjectConfigError{Reason: "invalid compatibility active_session_continuity"}
		}
		if _, ok := allowedProjectSessionContinuity[s]; !ok {
			return &ProjectConfigError{Reason: "invalid compatibility active_session_continuity enum"}
		}
	}
	if pr, present := comp["policy_ref"]; present && string(pr) != "null" {
		if !validProfileRef(pr, false) {
			return &ProjectConfigError{Reason: "invalid compatibility policy_ref"}
		}
	}
	return nil
}

func validateProjectConfigMigration(raw json.RawMessage) error {
	var mig map[string]json.RawMessage
	if json.Unmarshal(raw, &mig) != nil || mig == nil {
		return &ProjectConfigError{Reason: "migration must be an object"}
	}
	requiredMigration := []string{"strategy", "remove_dead_code", "preserve_dormant_code", "optimization"}
	for _, req := range requiredMigration {
		if _, present := mig[req]; !present || string(mig[req]) == "null" {
			return &ProjectConfigError{Reason: "missing required migration field: " + req}
		}
	}
	if len(mig) != 4 {
		return &ProjectConfigError{Reason: "unexpected field in migration"}
	}

	var strat string
	if json.Unmarshal(mig["strategy"], &strat) != nil || strat != "vertical_slice" {
		return &ProjectConfigError{Reason: "migration strategy must be vertical_slice"}
	}

	var removeDead bool
	if json.Unmarshal(mig["remove_dead_code"], &removeDead) != nil {
		return &ProjectConfigError{Reason: "migration remove_dead_code must be boolean"}
	}

	var preserveDormant bool
	if json.Unmarshal(mig["preserve_dormant_code"], &preserveDormant) != nil || !preserveDormant {
		return &ProjectConfigError{Reason: "migration preserve_dormant_code must be true"}
	}

	var optObj map[string]json.RawMessage
	if json.Unmarshal(mig["optimization"], &optObj) != nil || optObj == nil {
		return &ProjectConfigError{Reason: "migration optimization must be an object"}
	}
	if _, present := optObj["enabled_after_parity"]; !present || string(optObj["enabled_after_parity"]) == "null" {
		return &ProjectConfigError{Reason: "migration optimization enabled_after_parity is required"}
	}
	if len(optObj) != 1 {
		return &ProjectConfigError{Reason: "unexpected field in migration optimization"}
	}
	var enabledAfterParity bool
	if json.Unmarshal(optObj["enabled_after_parity"], &enabledAfterParity) != nil || !enabledAfterParity {
		return &ProjectConfigError{Reason: "migration optimization enabled_after_parity must be true"}
	}

	return nil
}

func parseProjectConfigJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &ProjectConfigError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &ProjectConfigError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &ProjectConfigError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &ProjectConfigError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &ProjectConfigError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &ProjectConfigError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &ProjectConfigError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &ProjectConfigError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &ProjectConfigError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &ProjectConfigError{Reason: "trailing input"}
	}
	return record, nil
}

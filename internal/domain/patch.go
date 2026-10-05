package domain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
)

// PatchError describes an invalid patch record without echoing private contents.
type PatchError struct {
	Reason string
	Cause  error
}

// Error identifies the failed patch obligation.
func (e *PatchError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid patch: %s: %v", e.Reason, e.Cause)
	}
	return "invalid patch: " + e.Reason
}

// Unwrap preserves the underlying error cause.
func (e *PatchError) Unwrap() error { return e.Cause }

// PatchAction defines the operation performed by a patch on a file.
type PatchAction string

const (
	PatchActionCreate PatchAction = "CREATE"
	PatchActionModify PatchAction = "MODIFY"
	PatchActionDelete PatchAction = "DELETE"
)

// PatchFileRecord describes a single file change within a patch.
type PatchFileRecord struct {
	Path         string      `json:"path"`
	Action       PatchAction `json:"action"`
	BeforeSHA256 *string     `json:"before_sha256"`
	AfterSHA256  *string     `json:"after_sha256"`
	ContentRef   *string     `json:"content_ref,omitempty"`
}

// PatchRecord represents a controlled, fenced migration patch (schema 1.0 per specs/migration/patch.schema.json).
type PatchRecord struct {
	SchemaVersion string            `json:"schema_version"`
	ProjectID     string            `json:"project_id"`
	UnitID        string            `json:"unit_id"`
	FencingToken  int64             `json:"fencing_token"`
	Files         []PatchFileRecord `json:"files"`
}

var (
	validPatchIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
	validSHA256Hex      = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// ValidatePatch validates a patch record against specs/migration/patch.schema.json.
func ValidatePatch(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil || record == nil {
		return &PatchError{Reason: "object fields", Cause: err}
	}

	expectedFields := []string{"schema_version", "project_id", "unit_id", "fencing_token", "files"}
	if len(record) != len(expectedFields) {
		return &PatchError{Reason: "object fields"}
	}
	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &PatchError{Reason: "missing required field: " + field}
		}
	}

	if profileString(record["schema_version"]) != "1.0" {
		return &PatchError{Reason: "unsupported schema_version; expected 1.0"}
	}

	projID := profileString(record["project_id"])
	if !validPatchIDPattern.MatchString(projID) {
		return &PatchError{Reason: "invalid project_id format"}
	}

	unitID := profileString(record["unit_id"])
	if !validPatchIDPattern.MatchString(unitID) {
		return &PatchError{Reason: "invalid unit_id format"}
	}

	var fencingToken int64
	if err := json.Unmarshal(record["fencing_token"], &fencingToken); err != nil || fencingToken < 0 {
		return &PatchError{Reason: "fencing_token must be an integer >= 0"}
	}

	var files []json.RawMessage
	if err := json.Unmarshal(record["files"], &files); err != nil {
		return &PatchError{Reason: "files must be an array", Cause: err}
	}

	seenPaths := make(map[string]struct{}, len(files))
	for _, fRaw := range files {
		var f map[string]json.RawMessage
		if err := json.Unmarshal(fRaw, &f); err != nil || f == nil {
			return &PatchError{Reason: "file item must be an object", Cause: err}
		}

		pathRaw, hasPath := f["path"]
		actionRaw, hasAction := f["action"]
		beforeRaw, hasBefore := f["before_sha256"]
		afterRaw, hasAfter := f["after_sha256"]
		if !hasPath || !hasAction || !hasBefore || !hasAfter {
			return &PatchError{Reason: "missing required file field"}
		}

		path := profileString(pathRaw)
		if len(path) == 0 {
			return &PatchError{Reason: "file path must be non-empty"}
		}
		if _, exists := seenPaths[path]; exists {
			return &PatchError{Reason: "duplicate file path in patch: " + path}
		}
		seenPaths[path] = struct{}{}

		action := PatchAction(profileString(actionRaw))
		switch action {
		case PatchActionCreate, PatchActionModify, PatchActionDelete:
		default:
			return &PatchError{Reason: "invalid action enum: " + string(action)}
		}

		// before_sha256 validation
		if string(beforeRaw) != "null" {
			beforeHash := profileString(beforeRaw)
			if !validSHA256Hex.MatchString(beforeHash) {
				return &PatchError{Reason: "invalid before_sha256 hex string"}
			}
			if action == PatchActionCreate {
				return &PatchError{Reason: "before_sha256 must be null for CREATE action"}
			}
		} else if action != PatchActionCreate {
			return &PatchError{Reason: "before_sha256 cannot be null for MODIFY or DELETE action"}
		}

		// after_sha256 validation
		if string(afterRaw) != "null" {
			afterHash := profileString(afterRaw)
			if !validSHA256Hex.MatchString(afterHash) {
				return &PatchError{Reason: "invalid after_sha256 hex string"}
			}
			if action == PatchActionDelete {
				return &PatchError{Reason: "after_sha256 must be null for DELETE action"}
			}
		} else if action != PatchActionDelete {
			return &PatchError{Reason: "after_sha256 cannot be null for CREATE or MODIFY action"}
		}

		// content_ref validation if present
		if contentRefRaw, present := f["content_ref"]; present && string(contentRefRaw) != "null" {
			if !validProfileRef(contentRefRaw, false) {
				return &PatchError{Reason: "invalid content_ref reference"}
			}
		}
	}

	return nil
}

// CanonicalJSON renders the patch into canonical RFC 8785 bytes.
func (p *PatchRecord) CanonicalJSON() ([]byte, error) {
	if p == nil {
		return nil, &PatchError{Reason: "patch record is nil"}
	}
	files := p.Files
	if files == nil {
		files = []PatchFileRecord{}
	}

	recordMap := map[string]any{
		"schema_version": p.SchemaVersion,
		"project_id":     p.ProjectID,
		"unit_id":        p.UnitID,
		"fencing_token":  p.FencingToken,
		"files":          files,
	}

	raw, err := json.Marshal(recordMap)
	if err != nil {
		return nil, &PatchError{Reason: "failed to marshal patch", Cause: err}
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &PatchError{Reason: "failed to canonicalize patch JSON", Cause: err}
	}
	if err := ValidatePatch(canon); err != nil {
		return nil, err
	}
	return canon, nil
}

// Digest computes the canonical sha256 reference for the patch record.
func (p *PatchRecord) Digest() (string, error) {
	canon, err := p.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canon)
	return fmt.Sprintf("sha256:%x", h), nil
}

// ParsePatch parses and validates a patch JSON record.
func ParsePatch(raw []byte) (*PatchRecord, error) {
	if err := ValidatePatch(raw); err != nil {
		return nil, err
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return nil, &PatchError{Reason: "canonical JSON", Cause: err}
	}
	var patch PatchRecord
	if err := json.Unmarshal(canon, &patch); err != nil {
		return nil, &PatchError{Reason: "unmarshal patch", Cause: err}
	}
	if patch.Files == nil {
		patch.Files = []PatchFileRecord{}
	}
	return &patch, nil
}

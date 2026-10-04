package domain

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SourceApprovalError describes a malformed subject without echoing private data.
type SourceApprovalError struct {
	Reason string
	Cause  error
}

// Error identifies the failed SOURCE-subject obligation.
func (e *SourceApprovalError) Error() string { return "invalid source approval subject: " + e.Reason }

// Unwrap preserves decoding/canonicalization failures for structured inspection.
func (e *SourceApprovalError) Unwrap() error { return e.Cause }

// ValidateSourceApprovalSubjectShape checks the SOURCE branch of subject 2.0 and
// the canonical adapter tuple ordering. It preserves supplied bytes and rejects
// duplicates/unsorted bindings rather than silently reordering them. Historical
// shapes lacking current execution prerequisites remain inspectable here.
// This does not qualify the required engine binding, manifests, snapshot/closure,
// artifact ownership/availability, human receipt or currentness. Shape validation
// neither creates an approval nor authorizes execution. PLAN/RELEASE are rejected.
func ValidateSourceApprovalSubjectShape(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &SourceApprovalError{Reason: "canonical JSON", Cause: err}
	}
	record, err := sourceApprovalObject(canonical, "schema_version", "canonicalization", "project_id", "kind", "snapshot_ref", "detection_ref", "assertions_ref", "adapter_bindings")
	if err != nil {
		return err
	}
	if profileString(record["schema_version"]) != "2.0" || profileString(record["canonicalization"]) != "rwt-cjson-1" || profileString(record["kind"]) != "SOURCE" {
		return &SourceApprovalError{Reason: "unsupported subject identity"}
	}
	if !validApprovalID(profileString(record["project_id"])) {
		return &SourceApprovalError{Reason: "project identity"}
	}
	for _, field := range []string{"snapshot_ref", "detection_ref", "assertions_ref"} {
		if !validProfileRef(record[field], false) {
			return &SourceApprovalError{Reason: "artifact reference"}
		}
	}
	var bindings []json.RawMessage
	if err := json.Unmarshal(record["adapter_bindings"], &bindings); err != nil || len(bindings) == 0 {
		return &SourceApprovalError{Reason: "binding array", Cause: err}
	}
	versionPattern, err := regexp.Compile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9][A-Za-z0-9.-]*)?$`)
	if err != nil {
		return &SourceApprovalError{Reason: "adapter version grammar", Cause: err}
	}
	var previous [5]string
	for i, rawBinding := range bindings {
		binding, err := sourceApprovalObject(rawBinding, "role", "adapter_id", "version", "manifest_ref", "implementation_ref")
		if err != nil {
			return err
		}
		tuple := [5]string{profileString(binding["role"]), profileString(binding["adapter_id"]), profileString(binding["version"]), profileString(binding["manifest_ref"]), profileString(binding["implementation_ref"])}
		switch tuple[0] {
		case "SOURCE", "TARGET", "PROBE", "COMPARATOR", "RUNNER":
		default:
			return &SourceApprovalError{Reason: "binding role"}
		}
		if !validApprovalID(tuple[1]) {
			return &SourceApprovalError{Reason: "adapter identity"}
		}
		// Adapter package versions follow this exact schema grammar; resolved
		// environment components have their own opaque release identity contract.
		if !versionPattern.MatchString(tuple[2]) {
			return &SourceApprovalError{Reason: "adapter version"}
		}
		if !validProfileRef(binding["manifest_ref"], false) || !validProfileRef(binding["implementation_ref"], false) {
			return &SourceApprovalError{Reason: "binding artifact reference"}
		}
		if i > 0 && compareApprovalBindings(previous, tuple) >= 0 {
			return &SourceApprovalError{Reason: "binding order or duplicate"}
		}
		previous = tuple
	}
	return nil
}

func sourceApprovalObject(raw []byte, fields ...string) (map[string]json.RawMessage, error) {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil || len(record) != len(fields) {
		return nil, &SourceApprovalError{Reason: "object fields", Cause: err}
	}
	for _, field := range fields {
		if _, present := record[field]; !present {
			return nil, &SourceApprovalError{Reason: "missing required field"}
		}
	}
	return record, nil
}

func validApprovalID(value string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	}) < 0
}

func compareApprovalBindings(a, b [5]string) int {
	for i := range a {
		if order := strings.Compare(a[i], b[i]); order != 0 {
			return order
		}
	}
	return 0
}

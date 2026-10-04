package domain

import (
	"encoding/json"
	"regexp"
	"strings"
)

var semverPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:\.[0-9]+)?(?:-[A-Za-z0-9][A-Za-z0-9.-]*)?$`)

var allowedServiceKinds = map[string]struct{}{
	"DATABASE": {},
	"CACHE":    {},
	"QUEUE":    {},
	"WEB":      {},
	"STORAGE":  {},
}

// ResolvedEnvironmentError describes an invalid resolved-environment record without echoing private contents.
type ResolvedEnvironmentError struct {
	Reason string
	Cause  error
}

// Error identifies the failed resolved-environment obligation.
func (e *ResolvedEnvironmentError) Error() string {
	return "invalid resolved environment: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *ResolvedEnvironmentError) Unwrap() error { return e.Cause }

// ValidateResolvedEnvironment validates a 2.0 or 2.1 resolved-environment record.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required top-level
// fields, exact SemVer versions for 2.0, opaque provider-qualified releases for 2.1,
// runtime, framework, services, package manager, and toolchain component structures.
func ValidateResolvedEnvironment(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ResolvedEnvironmentError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &ResolvedEnvironmentError{Reason: "object fields", Cause: err}
	}

	versionRaw, present := record["schema_version"]
	if !present {
		return &ResolvedEnvironmentError{Reason: "missing required field"}
	}
	version := profileString(versionRaw)
	if version != "2.0" && version != "2.1" {
		return &ResolvedEnvironmentError{Reason: "unsupported schema version"}
	}

	expectedFields := []string{
		"schema_version",
		"resolution_id",
		"project_id",
		"intent_ref",
		"language",
		"runtime",
		"framework",
		"os",
		"architecture",
		"services",
		"package_manager",
		"build_ref",
		"deployment_profile_ref",
		"lock_ref",
		"feasibility_ref",
	}
	if version == "2.1" {
		expectedFields = append(expectedFields, "toolchain_components")
	}

	for _, field := range expectedFields {
		if _, present := record[field]; !present {
			return &ResolvedEnvironmentError{Reason: "missing required field"}
		}
	}
	if len(record) != len(expectedFields) {
		return &ResolvedEnvironmentError{Reason: "object fields"}
	}

	if !validAssessmentReportID(profileString(record["resolution_id"])) {
		return &ResolvedEnvironmentError{Reason: "resolution identity"}
	}

	if !validAssessmentReportID(profileString(record["project_id"])) {
		return &ResolvedEnvironmentError{Reason: "project identity"}
	}

	for _, refField := range []string{"intent_ref", "build_ref", "deployment_profile_ref", "lock_ref", "feasibility_ref"} {
		if !validProfileRef(record[refField], false) {
			return &ResolvedEnvironmentError{Reason: refField}
		}
	}

	for _, strField := range []string{"language", "os", "architecture"} {
		val := profileString(record[strField])
		if strings.TrimSpace(val) == "" {
			return &ResolvedEnvironmentError{Reason: strField}
		}
	}

	if err := validateResolvedComponent(record["runtime"], version, "runtime"); err != nil {
		return err
	}

	if err := validateNullableComponent(record["framework"], version, "framework"); err != nil {
		return err
	}

	if err := validateNullableComponent(record["package_manager"], version, "package manager"); err != nil {
		return err
	}

	if err := validateResolvedServices(record["services"], version); err != nil {
		return err
	}

	if version == "2.1" {
		if err := validateToolchainComponents(record["toolchain_components"]); err != nil {
			return err
		}
	}

	return nil
}

func validateResolvedComponent(raw json.RawMessage, schemaVersion string, role string) error {
	var comp map[string]json.RawMessage
	if err := json.Unmarshal(raw, &comp); err != nil || comp == nil {
		return &ResolvedEnvironmentError{Reason: role + " object", Cause: err}
	}

	expectedFields := []string{"name", "version", "distribution_ref"}
	if schemaVersion == "2.1" {
		expectedFields = append(expectedFields, "provider_id", "interpretation_ref")
	}

	for _, field := range expectedFields {
		if _, present := comp[field]; !present {
			return &ResolvedEnvironmentError{Reason: "missing " + role + " field"}
		}
	}
	if len(comp) != len(expectedFields) {
		return &ResolvedEnvironmentError{Reason: role + " fields"}
	}

	name := profileString(comp["name"])
	if strings.TrimSpace(name) == "" {
		return &ResolvedEnvironmentError{Reason: role + " name"}
	}

	ver := profileString(comp["version"])
	if strings.TrimSpace(ver) == "" {
		return &ResolvedEnvironmentError{Reason: role + " version"}
	}
	if schemaVersion == "2.0" && !semverPattern.MatchString(ver) {
		return &ResolvedEnvironmentError{Reason: role + " version"}
	}

	if !validProfileRef(comp["distribution_ref"], false) {
		return &ResolvedEnvironmentError{Reason: role + " distribution_ref"}
	}

	if schemaVersion == "2.1" {
		providerID := profileString(comp["provider_id"])
		if strings.TrimSpace(providerID) == "" {
			return &ResolvedEnvironmentError{Reason: role + " provider_id"}
		}
		if !validProfileRef(comp["interpretation_ref"], false) {
			return &ResolvedEnvironmentError{Reason: role + " interpretation_ref"}
		}
	}

	return nil
}

func validateNullableComponent(raw json.RawMessage, schemaVersion string, role string) error {
	if string(raw) == "null" {
		return nil
	}
	return validateResolvedComponent(raw, schemaVersion, role)
}

func validateResolvedServices(raw json.RawMessage, schemaVersion string) error {
	var services []json.RawMessage
	if err := json.Unmarshal(raw, &services); err != nil {
		return &ResolvedEnvironmentError{Reason: "services array", Cause: err}
	}

	seenIDs := make(map[string]struct{}, len(services))
	for _, svcRaw := range services {
		var svc map[string]json.RawMessage
		if err := json.Unmarshal(svcRaw, &svc); err != nil || svc == nil {
			return &ResolvedEnvironmentError{Reason: "service object", Cause: err}
		}

		expected := []string{"service_id", "kind", "component", "profile_ref"}
		for _, field := range expected {
			if _, present := svc[field]; !present {
				return &ResolvedEnvironmentError{Reason: "missing service field"}
			}
		}
		if len(svc) != len(expected) {
			return &ResolvedEnvironmentError{Reason: "service fields"}
		}

		svcID := profileString(svc["service_id"])
		if !validAssessmentReportID(svcID) {
			return &ResolvedEnvironmentError{Reason: "service identity"}
		}
		if _, exists := seenIDs[svcID]; exists {
			return &ResolvedEnvironmentError{Reason: "duplicate service identity"}
		}
		seenIDs[svcID] = struct{}{}

		kind := profileString(svc["kind"])
		if _, ok := allowedServiceKinds[kind]; !ok {
			return &ResolvedEnvironmentError{Reason: "invalid service kind"}
		}

		if err := validateResolvedComponent(svc["component"], schemaVersion, "service component"); err != nil {
			return err
		}

		if !validProfileRef(svc["profile_ref"], false) {
			return &ResolvedEnvironmentError{Reason: "service profile_ref"}
		}
	}
	return nil
}

func validateToolchainComponents(raw json.RawMessage) error {
	var tools []json.RawMessage
	if err := json.Unmarshal(raw, &tools); err != nil {
		return &ResolvedEnvironmentError{Reason: "toolchain_components array", Cause: err}
	}

	seenTools := make(map[string]struct{}, len(tools))
	for _, toolRaw := range tools {
		var tool map[string]json.RawMessage
		if err := json.Unmarshal(toolRaw, &tool); err != nil || tool == nil {
			return &ResolvedEnvironmentError{Reason: "toolchain component object", Cause: err}
		}

		expected := []string{"provider_id", "name", "version", "distribution_ref", "interpretation_ref"}
		for _, field := range expected {
			if _, present := tool[field]; !present {
				return &ResolvedEnvironmentError{Reason: "missing toolchain component field"}
			}
		}
		if len(tool) != len(expected) {
			return &ResolvedEnvironmentError{Reason: "toolchain component fields"}
		}

		providerID := profileString(tool["provider_id"])
		if strings.TrimSpace(providerID) == "" {
			return &ResolvedEnvironmentError{Reason: "toolchain component provider_id"}
		}

		name := profileString(tool["name"])
		if strings.TrimSpace(name) == "" {
			return &ResolvedEnvironmentError{Reason: "toolchain component name"}
		}

		toolKey := providerID + "/" + name
		if _, exists := seenTools[toolKey]; exists {
			return &ResolvedEnvironmentError{Reason: "duplicate toolchain component"}
		}
		seenTools[toolKey] = struct{}{}

		ver := profileString(tool["version"])
		if strings.TrimSpace(ver) == "" {
			return &ResolvedEnvironmentError{Reason: "toolchain component version"}
		}

		if !validProfileRef(tool["distribution_ref"], false) {
			return &ResolvedEnvironmentError{Reason: "toolchain component distribution_ref"}
		}

		if !validProfileRef(tool["interpretation_ref"], false) {
			return &ResolvedEnvironmentError{Reason: "toolchain component interpretation_ref"}
		}
	}
	return nil
}

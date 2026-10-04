package domain

import (
	"encoding/json"
	"strings"
)

var allowedResolutionStrategies = map[string]struct{}{
	"DIRECT_UPGRADE": {},
	"ADAPTER":        {},
	"REPLACEMENT":    {},
	"REIMPLEMENT":    {},
	"RETAIN":         {},
}

var allowedCompatibilityMetadata = map[string]struct{}{
	"COMPATIBLE":   {},
	"INCOMPATIBLE": {},
	"PARTIAL":      {},
	"UNKNOWN":      {},
}

var allowedCompatibilityUsedAPI = map[string]struct{}{
	"COMPATIBLE":   {},
	"INCOMPATIBLE": {},
	"PARTIAL":      {},
	"UNKNOWN":      {},
}

var allowedCompatibilityBehavior = map[string]struct{}{
	"PASS":     {},
	"MISMATCH": {},
	"UNKNOWN":  {},
}

var allowedMappingStatuses = map[string]struct{}{
	"COVERED": {},
	"MISSING": {},
	"UNKNOWN": {},
}

var allowedV1Decisions = map[string]struct{}{
	"PENDING":    {},
	"ACCEPTED":   {},
	"REJECTED":   {},
	"DEFERRED":   {},
	"SUPERSEDED": {},
}

var allowedV1Statuses = map[string]struct{}{
	"PROPOSED":    {},
	"PLANNED":     {},
	"IMPLEMENTED": {},
	"VERIFIED":    {},
	"BLOCKED":     {},
	"SUPERSEDED":  {},
}

// DependencyResolutionError describes an invalid dependency resolution record without echoing private contents.
type DependencyResolutionError struct {
	Reason string
	Cause  error
}

// Error identifies the failed dependency resolution obligation.
func (e *DependencyResolutionError) Error() string {
	return "invalid dependency resolution: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *DependencyResolutionError) Unwrap() error { return e.Cause }

// ValidateDependencyResolution validates a dependency resolution record conforming to
// specs/migration/dependency-resolution.schema.json.
// It supports version 1.0 (historical) and version 2.0 (current immutable decision ref),
// validating source inventory, usage, proposal strategies, capability mappings,
// transitive dependencies, and allOf cross-field consistency rules.
func ValidateDependencyResolution(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &DependencyResolutionError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &DependencyResolutionError{Reason: "object required", Cause: err}
	}

	verRaw, ok := record["schema_version"]
	if !ok || string(verRaw) == "null" {
		return &DependencyResolutionError{Reason: "missing required field: schema_version"}
	}
	var version string
	if err := json.Unmarshal(verRaw, &version); err != nil {
		return &DependencyResolutionError{Reason: "schema_version must be a string", Cause: err}
	}

	switch version {
	case "1.0":
		return validateDependencyResolutionV1(record)
	case "2.0":
		return validateDependencyResolutionV2(record)
	default:
		return &DependencyResolutionError{Reason: "unsupported schema version"}
	}
}

func validateDependencyResolutionV1(record map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"dependency_id",
		"target_spec_hash",
		"source",
		"usage",
		"proposal",
		"compatibility",
		"capability_mappings",
		"transitive_dependency_ids",
		"decision",
		"status",
		"evidence_refs",
	}

	for _, req := range expectedFields {
		v, present := record[req]
		if !present || string(v) == "null" {
			return &DependencyResolutionError{Reason: "missing required field: " + req}
		}
	}
	if len(record) != len(expectedFields) {
		return &DependencyResolutionError{Reason: "unexpected fields"}
	}

	if err := validateDependencyCommon(record); err != nil {
		return err
	}

	var decision string
	if err := json.Unmarshal(record["decision"], &decision); err != nil {
		return &DependencyResolutionError{Reason: "decision must be a string", Cause: err}
	}
	if _, ok := allowedV1Decisions[decision]; !ok {
		return &DependencyResolutionError{Reason: "invalid decision enum"}
	}

	var status string
	if err := json.Unmarshal(record["status"], &status); err != nil {
		return &DependencyResolutionError{Reason: "status must be a string", Cause: err}
	}
	if _, ok := allowedV1Statuses[status]; !ok {
		return &DependencyResolutionError{Reason: "invalid status enum"}
	}

	// allOf rule 1: status in PLANNED, IMPLEMENTED, VERIFIED requires decision == ACCEPTED
	if status == "PLANNED" || status == "IMPLEMENTED" || status == "VERIFIED" {
		if decision != "ACCEPTED" {
			return &DependencyResolutionError{Reason: "status PLANNED/IMPLEMENTED/VERIFIED requires decision ACCEPTED"}
		}
	}

	// allOf rule 2: status == VERIFIED requires verified behavior, complete compatibility, and characterization cases
	if status == "VERIFIED" {
		if err := validateV1VerifiedConsistency(record); err != nil {
			return err
		}
	}

	return nil
}

func validateV1VerifiedConsistency(record map[string]json.RawMessage) error {
	var comp map[string]json.RawMessage
	if err := json.Unmarshal(record["compatibility"], &comp); err != nil || comp == nil {
		return &DependencyResolutionError{Reason: "compatibility object required"}
	}

	var behavior string
	if err := json.Unmarshal(comp["behavior"], &behavior); err != nil || behavior != "PASS" {
		return &DependencyResolutionError{Reason: "verified status requires compatibility.behavior PASS"}
	}

	var metadata string
	if err := json.Unmarshal(comp["metadata"], &metadata); err != nil || metadata != "COMPATIBLE" {
		return &DependencyResolutionError{Reason: "verified status requires compatibility.metadata COMPATIBLE"}
	}

	var usedAPI string
	if err := json.Unmarshal(comp["used_api"], &usedAPI); err != nil || usedAPI != "COMPATIBLE" {
		return &DependencyResolutionError{Reason: "verified status requires compatibility.used_api COMPATIBLE"}
	}

	var missingCaps []string
	if err := json.Unmarshal(comp["missing_capabilities"], &missingCaps); err != nil || len(missingCaps) != 0 {
		return &DependencyResolutionError{Reason: "verified status requires empty missing_capabilities"}
	}

	var mappings []map[string]json.RawMessage
	if err := json.Unmarshal(record["capability_mappings"], &mappings); err != nil {
		return &DependencyResolutionError{Reason: "capability_mappings must be an array"}
	}
	for _, m := range mappings {
		var st string
		if err := json.Unmarshal(m["status"], &st); err != nil || st != "COVERED" {
			return &DependencyResolutionError{Reason: "verified status requires all capability_mappings status COVERED"}
		}
		var caseIDs []string
		if err := json.Unmarshal(m["case_ids"], &caseIDs); err != nil || len(caseIDs) < 1 {
			return &DependencyResolutionError{Reason: "verified status requires capability_mappings case_ids to have minItems 1"}
		}
	}

	return nil
}

func validateDependencyResolutionV2(record map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"dependency_id",
		"target_spec_hash",
		"source",
		"usage",
		"proposal",
		"compatibility",
		"capability_mappings",
		"transitive_dependency_ids",
		"evidence_refs",
		"decision_ref",
	}

	for _, req := range expectedFields {
		v, present := record[req]
		if !present || string(v) == "null" {
			return &DependencyResolutionError{Reason: "missing required field: " + req}
		}
	}
	if len(record) != len(expectedFields) {
		return &DependencyResolutionError{Reason: "unexpected fields"}
	}

	if err := validateDependencyCommon(record); err != nil {
		return err
	}

	if !validProfileRef(record["decision_ref"], false) {
		return &DependencyResolutionError{Reason: "decision_ref must be a valid sha256 reference"}
	}

	return nil
}

func validateDependencyCommon(record map[string]json.RawMessage) error {
	var depID string
	if err := json.Unmarshal(record["dependency_id"], &depID); err != nil || !validAssessmentReportID(depID) {
		return &DependencyResolutionError{Reason: "dependency_id must match pattern ^[a-z0-9_-]+$"}
	}

	var targetHash string
	if err := json.Unmarshal(record["target_spec_hash"], &targetHash); err != nil || !isManifestDigest(targetHash) {
		return &DependencyResolutionError{Reason: "target_spec_hash must match pattern ^[a-f0-9]{64}$"}
	}

	if err := validateDependencySource(record["source"]); err != nil {
		return err
	}

	if err := validateDependencyUsage(record["usage"]); err != nil {
		return err
	}

	if err := validateDependencyProposal(record["proposal"]); err != nil {
		return err
	}

	if err := validateDependencyCompatibility(record["compatibility"]); err != nil {
		return err
	}

	if err := validateDependencyCapabilityMappings(record["capability_mappings"]); err != nil {
		return err
	}

	if err := validateDependencyTransitiveIDs(record["transitive_dependency_ids"]); err != nil {
		return err
	}

	if err := validateEvidenceRefsArray(record["evidence_refs"], 1); err != nil {
		return err
	}

	return nil
}

func validateDependencySource(raw json.RawMessage) error {
	var src map[string]json.RawMessage
	if err := json.Unmarshal(raw, &src); err != nil || src == nil {
		return &DependencyResolutionError{Reason: "source must be an object"}
	}

	for _, req := range []string{"ecosystem", "package", "version", "evidence_refs"} {
		v, ok := src[req]
		if !ok || string(v) == "null" {
			return &DependencyResolutionError{Reason: "source missing required field: " + req}
		}
	}
	if len(src) != 4 {
		return &DependencyResolutionError{Reason: "unexpected field in source"}
	}

	for _, sField := range []string{"ecosystem", "package", "version"} {
		var val string
		if err := json.Unmarshal(src[sField], &val); err != nil || len(val) < 1 {
			return &DependencyResolutionError{Reason: "source " + sField + " must be a non-empty string"}
		}
	}

	return validateEvidenceRefsArray(src["evidence_refs"], 1)
}

func validateDependencyUsage(raw json.RawMessage) error {
	var usage map[string]json.RawMessage
	if err := json.Unmarshal(raw, &usage); err != nil || usage == nil {
		return &DependencyResolutionError{Reason: "usage must be an object"}
	}

	for _, req := range []string{"symbols", "source_files", "affected_unit_ids", "required_capabilities"} {
		v, ok := usage[req]
		if !ok || string(v) == "null" {
			return &DependencyResolutionError{Reason: "usage missing required field: " + req}
		}
	}
	if len(usage) != 4 {
		return &DependencyResolutionError{Reason: "unexpected field in usage"}
	}

	for _, strArrayField := range []string{"symbols", "source_files"} {
		var items []string
		if err := json.Unmarshal(usage[strArrayField], &items); err != nil {
			return &DependencyResolutionError{Reason: "usage " + strArrayField + " must be an array of strings"}
		}
		seen := make(map[string]struct{}, len(items))
		for _, item := range items {
			if len(item) < 1 {
				return &DependencyResolutionError{Reason: "usage " + strArrayField + " item must not be empty"}
			}
			if _, exists := seen[item]; exists {
				return &DependencyResolutionError{Reason: "duplicate item in usage " + strArrayField}
			}
			seen[item] = struct{}{}
		}
	}

	var affectedUnits []string
	if err := json.Unmarshal(usage["affected_unit_ids"], &affectedUnits); err != nil {
		return &DependencyResolutionError{Reason: "usage affected_unit_ids must be an array of strings"}
	}
	seenUnits := make(map[string]struct{}, len(affectedUnits))
	for _, u := range affectedUnits {
		if !validAssessmentReportID(u) {
			return &DependencyResolutionError{Reason: "invalid ID in affected_unit_ids"}
		}
		if _, exists := seenUnits[u]; exists {
			return &DependencyResolutionError{Reason: "duplicate ID in affected_unit_ids"}
		}
		seenUnits[u] = struct{}{}
	}

	var reqCaps []string
	if err := json.Unmarshal(usage["required_capabilities"], &reqCaps); err != nil || len(reqCaps) < 1 {
		return &DependencyResolutionError{Reason: "usage required_capabilities must be a non-empty array"}
	}
	seenCaps := make(map[string]struct{}, len(reqCaps))
	for _, c := range reqCaps {
		if !validAssessmentReportID(c) {
			return &DependencyResolutionError{Reason: "invalid ID in required_capabilities"}
		}
		if _, exists := seenCaps[c]; exists {
			return &DependencyResolutionError{Reason: "duplicate ID in required_capabilities"}
		}
		seenCaps[c] = struct{}{}
	}

	return nil
}

func validateDependencyProposal(raw json.RawMessage) error {
	var prop map[string]json.RawMessage
	if err := json.Unmarshal(raw, &prop); err != nil || prop == nil {
		return &DependencyResolutionError{Reason: "proposal must be an object"}
	}

	for _, req := range []string{"strategy", "package", "version", "adapter_required"} {
		v, ok := prop[req]
		if !ok {
			return &DependencyResolutionError{Reason: "proposal missing required field: " + req}
		}
		if (req == "strategy" || req == "adapter_required") && string(v) == "null" {
			return &DependencyResolutionError{Reason: "proposal field cannot be null: " + req}
		}
	}
	if len(prop) != 4 {
		return &DependencyResolutionError{Reason: "unexpected field in proposal"}
	}

	var strategy string
	if err := json.Unmarshal(prop["strategy"], &strategy); err != nil {
		return &DependencyResolutionError{Reason: "proposal strategy must be a string", Cause: err}
	}
	if _, ok := allowedResolutionStrategies[strategy]; !ok {
		return &DependencyResolutionError{Reason: "invalid resolution strategy"}
	}

	if string(prop["package"]) != "null" {
		var pkg string
		if err := json.Unmarshal(prop["package"], &pkg); err != nil {
			return &DependencyResolutionError{Reason: "proposal package must be string or null", Cause: err}
		}
	}

	if string(prop["version"]) != "null" {
		var ver string
		if err := json.Unmarshal(prop["version"], &ver); err != nil {
			return &DependencyResolutionError{Reason: "proposal version must be string or null", Cause: err}
		}
	}

	var adapterReq bool
	if err := json.Unmarshal(prop["adapter_required"], &adapterReq); err != nil {
		return &DependencyResolutionError{Reason: "proposal adapter_required must be a boolean", Cause: err}
	}

	return nil
}

func validateDependencyCompatibility(raw json.RawMessage) error {
	var comp map[string]json.RawMessage
	if err := json.Unmarshal(raw, &comp); err != nil || comp == nil {
		return &DependencyResolutionError{Reason: "compatibility must be an object"}
	}

	for _, req := range []string{"metadata", "used_api", "behavior", "missing_capabilities"} {
		v, ok := comp[req]
		if !ok || string(v) == "null" {
			return &DependencyResolutionError{Reason: "compatibility missing required field: " + req}
		}
	}
	if len(comp) != 4 {
		return &DependencyResolutionError{Reason: "unexpected field in compatibility"}
	}

	var metadata string
	if err := json.Unmarshal(comp["metadata"], &metadata); err != nil {
		return &DependencyResolutionError{Reason: "compatibility metadata must be a string", Cause: err}
	}
	if _, ok := allowedCompatibilityMetadata[metadata]; !ok {
		return &DependencyResolutionError{Reason: "invalid compatibility metadata"}
	}

	var usedAPI string
	if err := json.Unmarshal(comp["used_api"], &usedAPI); err != nil {
		return &DependencyResolutionError{Reason: "compatibility used_api must be a string", Cause: err}
	}
	if _, ok := allowedCompatibilityUsedAPI[usedAPI]; !ok {
		return &DependencyResolutionError{Reason: "invalid compatibility used_api"}
	}

	var behavior string
	if err := json.Unmarshal(comp["behavior"], &behavior); err != nil {
		return &DependencyResolutionError{Reason: "compatibility behavior must be a string", Cause: err}
	}
	if _, ok := allowedCompatibilityBehavior[behavior]; !ok {
		return &DependencyResolutionError{Reason: "invalid compatibility behavior"}
	}

	var missingCaps []string
	if err := json.Unmarshal(comp["missing_capabilities"], &missingCaps); err != nil {
		return &DependencyResolutionError{Reason: "missing_capabilities must be an array of strings"}
	}
	seenCaps := make(map[string]struct{}, len(missingCaps))
	for _, c := range missingCaps {
		if !validAssessmentReportID(c) {
			return &DependencyResolutionError{Reason: "invalid ID in missing_capabilities"}
		}
		if _, exists := seenCaps[c]; exists {
			return &DependencyResolutionError{Reason: "duplicate ID in missing_capabilities"}
		}
		seenCaps[c] = struct{}{}
	}

	return nil
}

func validateDependencyCapabilityMappings(raw json.RawMessage) error {
	var mappings []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &mappings); err != nil || len(mappings) < 1 {
		return &DependencyResolutionError{Reason: "capability_mappings must be a non-empty array of objects"}
	}

	seenCaps := make(map[string]string, len(mappings))
	for _, m := range mappings {
		for _, req := range []string{"capability_id", "provider", "status", "case_ids", "evidence_refs"} {
			v, ok := m[req]
			if !ok || string(v) == "null" {
				return &DependencyResolutionError{Reason: "capability_mappings item missing required field: " + req}
			}
		}
		if len(m) != 5 {
			return &DependencyResolutionError{Reason: "unexpected field in capability_mappings item"}
		}

		var capID string
		if err := json.Unmarshal(m["capability_id"], &capID); err != nil || !validAssessmentReportID(capID) {
			return &DependencyResolutionError{Reason: "invalid capability_id in capability_mappings"}
		}

		var provider string
		if err := json.Unmarshal(m["provider"], &provider); err != nil || len(provider) < 1 {
			return &DependencyResolutionError{Reason: "capability_mappings provider must be a non-empty string"}
		}

		// check uniqueness of capability_id + provider
		combo := capID + "::" + provider
		if _, exists := seenCaps[combo]; exists {
			return &DependencyResolutionError{Reason: "duplicate capability mapping item"}
		}
		seenCaps[combo] = ""

		var st string
		if err := json.Unmarshal(m["status"], &st); err != nil {
			return &DependencyResolutionError{Reason: "capability_mappings status must be a string", Cause: err}
		}
		if _, ok := allowedMappingStatuses[st]; !ok {
			return &DependencyResolutionError{Reason: "invalid mapping status"}
		}

		var caseIDs []string
		if err := json.Unmarshal(m["case_ids"], &caseIDs); err != nil {
			return &DependencyResolutionError{Reason: "case_ids must be an array of strings"}
		}
		seenCases := make(map[string]struct{}, len(caseIDs))
		for _, c := range caseIDs {
			if !validAssessmentReportID(c) {
				return &DependencyResolutionError{Reason: "invalid ID in case_ids"}
			}
			if _, exists := seenCases[c]; exists {
				return &DependencyResolutionError{Reason: "duplicate ID in case_ids"}
			}
			seenCases[c] = struct{}{}
		}

		if err := validateEvidenceRefsArray(m["evidence_refs"], 1); err != nil {
			return err
		}
	}

	return nil
}

func validateDependencyTransitiveIDs(raw json.RawMessage) error {
	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		return &DependencyResolutionError{Reason: "transitive_dependency_ids must be an array of strings"}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if !validAssessmentReportID(id) {
			return &DependencyResolutionError{Reason: "invalid ID in transitive_dependency_ids"}
		}
		if _, exists := seen[id]; exists {
			return &DependencyResolutionError{Reason: "duplicate ID in transitive_dependency_ids"}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateEvidenceRefsArray(raw json.RawMessage, minItems int) error {
	var refs []string
	if err := json.Unmarshal(raw, &refs); err != nil || len(refs) < minItems {
		return &DependencyResolutionError{Reason: "evidence_refs must be a non-empty array of strings"}
	}
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
			return &DependencyResolutionError{Reason: "evidence_refs item must be a valid sha256 reference"}
		}
		if _, exists := seen[ref]; exists {
			return &DependencyResolutionError{Reason: "duplicate reference in evidence_refs"}
		}
		seen[ref] = struct{}{}
	}
	return nil
}

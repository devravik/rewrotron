package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var allowedPlanSeverityLevels = map[string]struct{}{
	"CRITICAL": {},
	"HIGH":     {},
	"MEDIUM":   {},
	"LOW":      {},
}

// PlanApprovalError describes an invalid PLAN approval subject without echoing private contents.
type PlanApprovalError struct {
	Reason string
	Cause  error
}

// Error identifies the failed PLAN approval subject obligation.
func (e *PlanApprovalError) Error() string {
	return "invalid plan approval subject: " + e.Reason
}

// Unwrap preserves decoding/canonicalization failures for structured inspection.
func (e *PlanApprovalError) Unwrap() error { return e.Cause }

// ValidatePlanApprovalSubjectShape checks the PLAN branch of subject 2.0 conforming to
// specs/project/approval-subject.schema.json#/$defs/plan.
// It verifies canonical RFC 8785 JSON within integer vocabulary, required subject identity,
// project identity, configuration hash, artifact references, adapter bindings ordering,
// ordered units, dependency edges, omission proofs, exclusions, and unresolved items.
// Shape validation neither creates an approval nor authorizes execution.
func ValidatePlanApprovalSubjectShape(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &PlanApprovalError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &PlanApprovalError{Reason: "object fields", Cause: err}
	}

	requiredFields := []string{
		"schema_version",
		"canonicalization",
		"project_id",
		"kind",
		"configuration_hash",
		"snapshot_ref",
		"behavior_ir_ref",
		"corpus_ref",
		"policy_ref",
		"target_intent_ref",
		"resolved_environment_ref",
		"target_model_ref",
		"decision_refs",
		"dependency_graph_ref",
		"adapter_bindings",
		"consumed_revision_refs",
		"ordered_units",
		"dependency_edges",
		"omission_proof_refs",
		"scope_exclusion_refs",
		"unresolved_items",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &PlanApprovalError{Reason: "missing required field: " + f}
		}
	}

	if len(record) != len(requiredFields) {
		return &PlanApprovalError{Reason: "object fields"}
	}

	if profileString(record["schema_version"]) != "2.0" ||
		profileString(record["canonicalization"]) != "rwt-cjson-1" ||
		profileString(record["kind"]) != "PLAN" {
		return &PlanApprovalError{Reason: "unsupported subject identity"}
	}

	if !validApprovalID(profileString(record["project_id"])) {
		return &PlanApprovalError{Reason: "project identity"}
	}

	cfgHash := profileString(record["configuration_hash"])
	if !isManifestDigest(cfgHash) {
		return &PlanApprovalError{Reason: "configuration_hash"}
	}

	refFields := []string{
		"snapshot_ref",
		"behavior_ir_ref",
		"corpus_ref",
		"policy_ref",
		"target_intent_ref",
		"resolved_environment_ref",
		"target_model_ref",
		"dependency_graph_ref",
	}
	for _, rf := range refFields {
		if !validProfileRef(record[rf], false) {
			return &PlanApprovalError{Reason: "artifact reference: " + rf}
		}
	}

	if err := validateRefArray(record["decision_refs"], "decision_refs"); err != nil {
		return err
	}

	if err := validatePlanAdapterBindings(record["adapter_bindings"]); err != nil {
		return err
	}

	if err := validateRefArray(record["consumed_revision_refs"], "consumed_revision_refs"); err != nil {
		return err
	}

	if err := validatePlanOrderedUnits(record["ordered_units"]); err != nil {
		return err
	}

	if err := validatePlanDependencyEdges(record["dependency_edges"]); err != nil {
		return err
	}

	if err := validateRefArray(record["omission_proof_refs"], "omission_proof_refs"); err != nil {
		return err
	}

	if err := validateRefArray(record["scope_exclusion_refs"], "scope_exclusion_refs"); err != nil {
		return err
	}

	if err := validatePlanUnresolvedItems(record["unresolved_items"]); err != nil {
		return err
	}

	return nil
}

func validateRefArray(raw json.RawMessage, fieldName string) error {
	var list []string
	if string(raw) == "null" || json.Unmarshal(raw, &list) != nil {
		return &PlanApprovalError{Reason: fieldName + " must be an array"}
	}
	seen := make(map[string]struct{}, len(list))
	for _, ref := range list {
		if !validProfileRef(json.RawMessage(`"`+ref+`"`), false) {
			return &PlanApprovalError{Reason: "invalid ref in " + fieldName}
		}
		if _, exists := seen[ref]; exists {
			return &PlanApprovalError{Reason: "duplicate ref in " + fieldName}
		}
		seen[ref] = struct{}{}
	}
	return nil
}

func validatePlanAdapterBindings(raw json.RawMessage) error {
	var bindings []json.RawMessage
	if err := json.Unmarshal(raw, &bindings); err != nil || len(bindings) == 0 {
		return &PlanApprovalError{Reason: "binding array", Cause: err}
	}

	versionPattern, err := regexp.Compile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9][A-Za-z0-9.-]*)?$`)
	if err != nil {
		return &PlanApprovalError{Reason: "adapter version grammar", Cause: err}
	}

	var previous [5]string
	for i, rawBinding := range bindings {
		binding, err := sourceApprovalObject(rawBinding, "role", "adapter_id", "version", "manifest_ref", "implementation_ref")
		if err != nil {
			return &PlanApprovalError{Reason: "binding fields", Cause: err}
		}
		tuple := [5]string{
			profileString(binding["role"]),
			profileString(binding["adapter_id"]),
			profileString(binding["version"]),
			profileString(binding["manifest_ref"]),
			profileString(binding["implementation_ref"]),
		}
		switch tuple[0] {
		case "SOURCE", "TARGET", "PROBE", "COMPARATOR", "RUNNER":
		default:
			return &PlanApprovalError{Reason: "binding role"}
		}
		if !validApprovalID(tuple[1]) {
			return &PlanApprovalError{Reason: "adapter identity"}
		}
		if !versionPattern.MatchString(tuple[2]) {
			return &PlanApprovalError{Reason: "adapter version"}
		}
		if !validProfileRef(binding["manifest_ref"], false) || !validProfileRef(binding["implementation_ref"], false) {
			return &PlanApprovalError{Reason: "binding artifact reference"}
		}
		if i > 0 && compareApprovalBindings(previous, tuple) >= 0 {
			return &PlanApprovalError{Reason: "binding order or duplicate"}
		}
		previous = tuple
	}
	return nil
}

func validatePlanOrderedUnits(raw json.RawMessage) error {
	var units []json.RawMessage
	if err := json.Unmarshal(raw, &units); err != nil || len(units) == 0 {
		return &PlanApprovalError{Reason: "ordered_units must be a non-empty array", Cause: err}
	}

	seenUnits := make(map[string]struct{}, len(units))
	requiredFields := []string{"unit_id", "contract_refs", "case_ids", "priority", "capabilities"}

	for _, rawUnit := range units {
		var unit map[string]json.RawMessage
		if err := json.Unmarshal(rawUnit, &unit); err != nil || unit == nil {
			return &PlanApprovalError{Reason: "ordered_units item must be an object", Cause: err}
		}

		for _, f := range requiredFields {
			if _, present := unit[f]; !present {
				return &PlanApprovalError{Reason: "missing ordered_units field: " + f}
			}
		}
		if len(unit) != len(requiredFields) {
			return &PlanApprovalError{Reason: "ordered_units item unexpected fields"}
		}

		var unitID string
		if json.Unmarshal(unit["unit_id"], &unitID) != nil || !validApprovalID(unitID) {
			return &PlanApprovalError{Reason: "invalid unit_id in ordered_units"}
		}
		if _, exists := seenUnits[unitID]; exists {
			return &PlanApprovalError{Reason: "duplicate unit_id in ordered_units: " + unitID}
		}
		seenUnits[unitID] = struct{}{}

		var contractRefs []string
		if json.Unmarshal(unit["contract_refs"], &contractRefs) != nil || len(contractRefs) == 0 {
			return &PlanApprovalError{Reason: "contract_refs must be a non-empty array"}
		}
		seenContracts := make(map[string]struct{}, len(contractRefs))
		for _, cr := range contractRefs {
			if !validProfileRef(json.RawMessage(`"`+cr+`"`), false) {
				return &PlanApprovalError{Reason: "invalid contract_ref in ordered_units"}
			}
			if _, exists := seenContracts[cr]; exists {
				return &PlanApprovalError{Reason: "duplicate contract_ref in ordered_units"}
			}
			seenContracts[cr] = struct{}{}
		}

		var caseIDs []string
		if json.Unmarshal(unit["case_ids"], &caseIDs) != nil || len(caseIDs) == 0 {
			return &PlanApprovalError{Reason: "case_ids must be a non-empty array"}
		}
		seenCases := make(map[string]struct{}, len(caseIDs))
		for _, cid := range caseIDs {
			if !validApprovalID(cid) {
				return &PlanApprovalError{Reason: "invalid case_id in ordered_units"}
			}
			if _, exists := seenCases[cid]; exists {
				return &PlanApprovalError{Reason: "duplicate case_id in ordered_units"}
			}
			seenCases[cid] = struct{}{}
		}

		var priority int64
		if json.Unmarshal(unit["priority"], &priority) != nil || priority < 0 {
			return &PlanApprovalError{Reason: "priority must be an integer >= 0"}
		}

		var capabilities []string
		if json.Unmarshal(unit["capabilities"], &capabilities) != nil {
			return &PlanApprovalError{Reason: "capabilities must be an array"}
		}
		seenCaps := make(map[string]struct{}, len(capabilities))
		for _, cap := range capabilities {
			if strings.TrimSpace(cap) == "" {
				return &PlanApprovalError{Reason: "empty capability string"}
			}
			if _, exists := seenCaps[cap]; exists {
				return &PlanApprovalError{Reason: "duplicate capability"}
			}
			seenCaps[cap] = struct{}{}
		}
	}
	return nil
}

func validatePlanDependencyEdges(raw json.RawMessage) error {
	var edges []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &edges) != nil {
		return &PlanApprovalError{Reason: "dependency_edges must be an array"}
	}

	seenEdges := make(map[string]struct{}, len(edges))
	for _, rawEdge := range edges {
		var edge map[string]json.RawMessage
		if err := json.Unmarshal(rawEdge, &edge); err != nil || edge == nil {
			return &PlanApprovalError{Reason: "dependency_edges item must be an object", Cause: err}
		}
		if len(edge) != 2 {
			return &PlanApprovalError{Reason: "dependency_edges item fields"}
		}
		preRaw, hasPre := edge["prerequisite_id"]
		depRaw, hasDep := edge["dependent_id"]
		if !hasPre || !hasDep {
			return &PlanApprovalError{Reason: "missing dependency_edges field"}
		}
		var preID, depID string
		if json.Unmarshal(preRaw, &preID) != nil || !validApprovalID(preID) {
			return &PlanApprovalError{Reason: "invalid prerequisite_id in dependency_edges"}
		}
		if json.Unmarshal(depRaw, &depID) != nil || !validApprovalID(depID) {
			return &PlanApprovalError{Reason: "invalid dependent_id in dependency_edges"}
		}
		pair := preID + "->" + depID
		if _, exists := seenEdges[pair]; exists {
			return &PlanApprovalError{Reason: "duplicate dependency edge"}
		}
		seenEdges[pair] = struct{}{}
	}
	return nil
}

func validatePlanUnresolvedItems(raw json.RawMessage) error {
	var items []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &items) != nil {
		return &PlanApprovalError{Reason: "unresolved_items must be an array"}
	}

	requiredFields := []string{"code", "affected_ids", "evidence_refs", "blocking"}

	for _, rawItem := range items {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(rawItem, &item); err != nil || item == nil {
			return &PlanApprovalError{Reason: "unresolved_items item must be an object", Cause: err}
		}
		for _, f := range requiredFields {
			if _, present := item[f]; !present {
				return &PlanApprovalError{Reason: "missing unresolved_items field: " + f}
			}
		}
		if len(item) != len(requiredFields) {
			return &PlanApprovalError{Reason: "unresolved_items item fields"}
		}

		var code string
		if json.Unmarshal(item["code"], &code) != nil || strings.TrimSpace(code) == "" {
			return &PlanApprovalError{Reason: "invalid unresolved_items code"}
		}

		var affectedIDs []string
		if json.Unmarshal(item["affected_ids"], &affectedIDs) != nil || len(affectedIDs) == 0 {
			return &PlanApprovalError{Reason: "affected_ids must be a non-empty array"}
		}
		seenAff := make(map[string]struct{}, len(affectedIDs))
		for _, affID := range affectedIDs {
			if !validApprovalID(affID) {
				return &PlanApprovalError{Reason: "invalid affected_id in unresolved_items"}
			}
			if _, exists := seenAff[affID]; exists {
				return &PlanApprovalError{Reason: "duplicate affected_id"}
			}
			seenAff[affID] = struct{}{}
		}

		var evidenceRefs []string
		if json.Unmarshal(item["evidence_refs"], &evidenceRefs) != nil || len(evidenceRefs) == 0 {
			return &PlanApprovalError{Reason: "evidence_refs must be a non-empty array"}
		}
		seenEv := make(map[string]struct{}, len(evidenceRefs))
		for _, evRef := range evidenceRefs {
			if !validProfileRef(json.RawMessage(`"`+evRef+`"`), false) {
				return &PlanApprovalError{Reason: "invalid evidence_ref in unresolved_items"}
			}
			if _, exists := seenEv[evRef]; exists {
				return &PlanApprovalError{Reason: "duplicate evidence_ref"}
			}
			seenEv[evRef] = struct{}{}
		}

		var blocking bool
		if json.Unmarshal(item["blocking"], &blocking) != nil {
			return &PlanApprovalError{Reason: "blocking must be a boolean"}
		}
	}
	return nil
}

// MigrationPlanError describes an invalid migration plan record without echoing private contents.
type MigrationPlanError struct {
	Reason string
	Cause  error
}

// Error identifies the failed migration plan obligation.
func (e *MigrationPlanError) Error() string {
	return "invalid migration plan: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *MigrationPlanError) Unwrap() error { return e.Cause }

// ValidateMigrationPlan validates a migration plan record conforming to specs/migration/plan.schema.json.
// It verifies version 1.0 (legacy readable plan), version 2.0 (v2 plan envelope), or version 2.1
// (checkpoint-linked executable plan envelope), all required hashes, target configuration,
// embedded PLAN approval subject, and in version 2.1 verifies that selection_ref and planning_basis_ref
// are present and bound into the subject's decision_refs.
func ValidateMigrationPlan(raw []byte) error {
	record, err := parseMigrationPlanJSON(raw)
	if err != nil {
		return err
	}

	versionRaw, present := record["schema_version"]
	if !present || string(versionRaw) == "null" {
		return &MigrationPlanError{Reason: "missing schema_version"}
	}
	var version string
	if err := json.Unmarshal(versionRaw, &version); err != nil {
		return &MigrationPlanError{Reason: "schema_version must be a string", Cause: err}
	}

	switch version {
	case "1.0":
		return validateMigrationPlanV1(record)
	case "2.0":
		return validateMigrationPlanV2(record)
	case "2.1":
		return validateMigrationPlanV21(record)
	default:
		return &MigrationPlanError{Reason: "unsupported schema_version: " + version}
	}
}

func validateMigrationPlanV1(record map[string]json.RawMessage) error {
	requiredFields := []string{
		"schema_version",
		"plan_id",
		"project_id",
		"plan_hash",
		"configuration_hash",
		"behavior_ir_hash",
		"corpus_hash",
		"policy_hash",
		"target_config",
		"units",
		"omissions",
		"scope_exclusions",
		"unresolved_items",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &MigrationPlanError{Reason: "missing required field: " + f}
		}
	}

	for k := range record {
		isReq := false
		for _, f := range requiredFields {
			if k == f {
				isReq = true
				break
			}
		}
		if isReq || k == "decision_system" {
			continue
		}
		return &MigrationPlanError{Reason: "unexpected field: " + k}
	}

	var planID string
	if string(record["plan_id"]) == "null" || json.Unmarshal(record["plan_id"], &planID) != nil || !validAssessmentReportID(planID) {
		return &MigrationPlanError{Reason: "invalid plan_id"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &MigrationPlanError{Reason: "invalid project_id"}
	}

	for _, hashField := range []string{"plan_hash", "configuration_hash", "behavior_ir_hash", "corpus_hash", "policy_hash"} {
		var h string
		if string(record[hashField]) == "null" || json.Unmarshal(record[hashField], &h) != nil || !isManifestDigest(h) {
			return &MigrationPlanError{Reason: "invalid " + hashField}
		}
	}

	if err := ValidateTargetEnvironment(record["target_config"]); err != nil {
		return &MigrationPlanError{Reason: "invalid target_config", Cause: err}
	}

	if err := validatePlanV1Units(record["units"]); err != nil {
		return err
	}

	var omissions []json.RawMessage
	if string(record["omissions"]) == "null" || json.Unmarshal(record["omissions"], &omissions) != nil {
		return &MigrationPlanError{Reason: "omissions must be an array"}
	}
	for _, o := range omissions {
		var obj map[string]json.RawMessage
		if json.Unmarshal(o, &obj) != nil || obj == nil {
			return &MigrationPlanError{Reason: "omissions item must be an object"}
		}
	}

	var scopeExclusions []json.RawMessage
	if string(record["scope_exclusions"]) == "null" || json.Unmarshal(record["scope_exclusions"], &scopeExclusions) != nil {
		return &MigrationPlanError{Reason: "scope_exclusions must be an array"}
	}
	for _, se := range scopeExclusions {
		var obj map[string]json.RawMessage
		if json.Unmarshal(se, &obj) != nil || obj == nil {
			return &MigrationPlanError{Reason: "scope_exclusions item must be an object"}
		}
	}

	if err := validatePlanV1UnresolvedItems(record["unresolved_items"]); err != nil {
		return err
	}

	if dsRaw, present := record["decision_system"]; present && string(dsRaw) != "null" {
		if err := validatePlanV1DecisionSystem(dsRaw); err != nil {
			return err
		}
	}

	return nil
}

func validatePlanV1Units(raw json.RawMessage) error {
	var units []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &units) != nil {
		return &MigrationPlanError{Reason: "units must be an array"}
	}

	required := []string{"unit_id", "contract_ids", "case_ids", "dependencies", "priority", "capabilities"}

	for _, rawUnit := range units {
		var unit map[string]json.RawMessage
		if json.Unmarshal(rawUnit, &unit) != nil || unit == nil {
			return &MigrationPlanError{Reason: "units item must be an object"}
		}
		for _, f := range required {
			if _, present := unit[f]; !present {
				return &MigrationPlanError{Reason: "missing unit field: " + f}
			}
		}
		if len(unit) != len(required) {
			return &MigrationPlanError{Reason: "units item fields"}
		}

		var unitID string
		if json.Unmarshal(unit["unit_id"], &unitID) != nil || !validAssessmentReportID(unitID) {
			return &MigrationPlanError{Reason: "invalid unit_id in units"}
		}

		for _, arrField := range []string{"contract_ids", "case_ids", "dependencies"} {
			var ids []string
			if json.Unmarshal(unit[arrField], &ids) != nil {
				return &MigrationPlanError{Reason: arrField + " must be an array"}
			}
			for _, id := range ids {
				if !validAssessmentReportID(id) {
					return &MigrationPlanError{Reason: "invalid id in " + arrField}
				}
			}
		}

		var priority int64
		if json.Unmarshal(unit["priority"], &priority) != nil || priority < 0 {
			return &MigrationPlanError{Reason: "priority must be an integer >= 0"}
		}

		var caps []string
		if json.Unmarshal(unit["capabilities"], &caps) != nil {
			return &MigrationPlanError{Reason: "capabilities must be an array"}
		}
		for _, c := range caps {
			if strings.TrimSpace(c) == "" {
				return &MigrationPlanError{Reason: "empty capability"}
			}
		}
	}
	return nil
}

func validatePlanV1UnresolvedItems(raw json.RawMessage) error {
	var items []json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &items) != nil {
		return &MigrationPlanError{Reason: "unresolved_items must be an array"}
	}

	required := []string{"finding_id", "code", "severity", "affected_ids", "evidence_ids", "message"}

	for _, rawItem := range items {
		var item map[string]json.RawMessage
		if json.Unmarshal(rawItem, &item) != nil || item == nil {
			return &MigrationPlanError{Reason: "unresolved_items item must be an object"}
		}
		for _, f := range required {
			if _, present := item[f]; !present {
				return &MigrationPlanError{Reason: "missing unresolved_items field: " + f}
			}
		}
		if len(item) != len(required) {
			return &MigrationPlanError{Reason: "unresolved_items item fields"}
		}

		var findingID string
		if json.Unmarshal(item["finding_id"], &findingID) != nil || !validAssessmentReportID(findingID) {
			return &MigrationPlanError{Reason: "invalid finding_id"}
		}

		var code string
		if json.Unmarshal(item["code"], &code) != nil || strings.TrimSpace(code) == "" {
			return &MigrationPlanError{Reason: "invalid code"}
		}

		var sev string
		if json.Unmarshal(item["severity"], &sev) != nil {
			return &MigrationPlanError{Reason: "invalid severity"}
		}
		if _, ok := allowedPlanSeverityLevels[sev]; !ok {
			return &MigrationPlanError{Reason: "invalid severity enum: " + sev}
		}

		for _, arrField := range []string{"affected_ids", "evidence_ids"} {
			var ids []string
			if json.Unmarshal(item[arrField], &ids) != nil {
				return &MigrationPlanError{Reason: arrField + " must be an array"}
			}
			for _, id := range ids {
				if !validAssessmentReportID(id) {
					return &MigrationPlanError{Reason: "invalid id in " + arrField}
				}
			}
		}

		var msg string
		if json.Unmarshal(item["message"], &msg) != nil || strings.TrimSpace(msg) == "" {
			return &MigrationPlanError{Reason: "invalid message"}
		}
	}
	return nil
}

func validatePlanV1DecisionSystem(raw json.RawMessage) error {
	var ds map[string]json.RawMessage
	if err := json.Unmarshal(raw, &ds); err != nil || ds == nil {
		return &MigrationPlanError{Reason: "decision_system must be an object", Cause: err}
	}

	required := []string{
		"target_spec_hash",
		"resolved_target",
		"recommendation_decisions_ref",
		"dependency_resolution_ref",
		"feasibility_ref",
		"target_spec_ref",
	}

	for _, f := range required {
		if _, present := ds[f]; !present || string(ds[f]) == "null" {
			return &MigrationPlanError{Reason: "missing decision_system field: " + f}
		}
	}
	if len(ds) != len(required) {
		return &MigrationPlanError{Reason: "decision_system fields"}
	}

	var hash string
	if json.Unmarshal(ds["target_spec_hash"], &hash) != nil || !isManifestDigest(hash) {
		return &MigrationPlanError{Reason: "invalid decision_system target_spec_hash"}
	}

	if err := ValidateTargetEnvironment(ds["resolved_target"]); err != nil {
		return &MigrationPlanError{Reason: "invalid decision_system resolved_target", Cause: err}
	}

	for _, refField := range []string{"recommendation_decisions_ref", "dependency_resolution_ref", "feasibility_ref", "target_spec_ref"} {
		if !validProfileRef(ds[refField], false) {
			return &MigrationPlanError{Reason: "invalid decision_system " + refField}
		}
	}

	return nil
}

func validateMigrationPlanV2(record map[string]json.RawMessage) error {
	requiredFields := []string{"schema_version", "plan_id", "project_id", "plan_hash", "subject"}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &MigrationPlanError{Reason: "missing required field: " + f}
		}
	}
	if len(record) != len(requiredFields) {
		return &MigrationPlanError{Reason: "object fields"}
	}

	var planID string
	if string(record["plan_id"]) == "null" || json.Unmarshal(record["plan_id"], &planID) != nil || !validAssessmentReportID(planID) {
		return &MigrationPlanError{Reason: "invalid plan_id"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &MigrationPlanError{Reason: "invalid project_id"}
	}

	var planHash string
	if string(record["plan_hash"]) == "null" || json.Unmarshal(record["plan_hash"], &planHash) != nil || !isManifestDigest(planHash) {
		return &MigrationPlanError{Reason: "invalid plan_hash"}
	}

	subjectRaw := record["subject"]
	if string(subjectRaw) == "null" {
		return &MigrationPlanError{Reason: "subject cannot be null"}
	}

	if err := ValidatePlanApprovalSubjectShape(subjectRaw); err != nil {
		return &MigrationPlanError{Reason: "invalid subject", Cause: err}
	}

	var subjectObj map[string]json.RawMessage
	if err := json.Unmarshal(subjectRaw, &subjectObj); err == nil && subjectObj != nil {
		subjProjectID := profileString(subjectObj["project_id"])
		if subjProjectID != projectID {
			return &MigrationPlanError{Reason: "subject project_id does not match envelope project_id"}
		}
	}

	expectedHash, err := CanonicalHash(subjectRaw)
	if err != nil {
		return &MigrationPlanError{Reason: "failed to compute canonical subject hash", Cause: err}
	}
	if expectedHash != planHash {
		return &MigrationPlanError{Reason: "plan_hash does not match canonical subject digest"}
	}

	return nil
}

func validateMigrationPlanV21(record map[string]json.RawMessage) error {
	requiredFields := []string{
		"schema_version",
		"plan_id",
		"project_id",
		"plan_hash",
		"subject",
		"assessment_ref",
		"selection_ref",
		"planning_basis_ref",
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &MigrationPlanError{Reason: "missing required field: " + f}
		}
	}
	if len(record) != len(requiredFields) {
		return &MigrationPlanError{Reason: "object fields"}
	}

	var planID string
	if string(record["plan_id"]) == "null" || json.Unmarshal(record["plan_id"], &planID) != nil || !validAssessmentReportID(planID) {
		return &MigrationPlanError{Reason: "invalid plan_id"}
	}

	var projectID string
	if string(record["project_id"]) == "null" || json.Unmarshal(record["project_id"], &projectID) != nil || !validAssessmentReportID(projectID) {
		return &MigrationPlanError{Reason: "invalid project_id"}
	}

	var planHash string
	if string(record["plan_hash"]) == "null" || json.Unmarshal(record["plan_hash"], &planHash) != nil || !isManifestDigest(planHash) {
		return &MigrationPlanError{Reason: "invalid plan_hash"}
	}

	for _, refField := range []string{"assessment_ref", "selection_ref", "planning_basis_ref"} {
		if !validProfileRef(record[refField], false) {
			return &MigrationPlanError{Reason: "invalid " + refField}
		}
	}

	subjectRaw := record["subject"]
	if string(subjectRaw) == "null" {
		return &MigrationPlanError{Reason: "subject cannot be null"}
	}

	if err := ValidatePlanApprovalSubjectShape(subjectRaw); err != nil {
		return &MigrationPlanError{Reason: "invalid subject", Cause: err}
	}

	var subjectObj map[string]json.RawMessage
	if err := json.Unmarshal(subjectRaw, &subjectObj); err == nil && subjectObj != nil {
		subjProjectID := profileString(subjectObj["project_id"])
		if subjProjectID != projectID {
			return &MigrationPlanError{Reason: "subject project_id does not match envelope project_id"}
		}

		// Enforce CONFIRMATION-PROTOCOL.md:41: both selection and planning_basis refs
		// must be included in the subject's decision_refs.
		var decisionRefs []string
		if err := json.Unmarshal(subjectObj["decision_refs"], &decisionRefs); err == nil {
			selectionRef := profileString(record["selection_ref"])
			basisRef := profileString(record["planning_basis_ref"])
			hasSelection := false
			hasBasis := false
			for _, d := range decisionRefs {
				if d == selectionRef {
					hasSelection = true
				}
				if d == basisRef {
					hasBasis = true
				}
			}
			if !hasSelection {
				return &MigrationPlanError{Reason: "selection_ref must be included in subject decision_refs"}
			}
			if !hasBasis {
				return &MigrationPlanError{Reason: "planning_basis_ref must be included in subject decision_refs"}
			}
		}
	}

	expectedHash, err := CanonicalHash(subjectRaw)
	if err != nil {
		return &MigrationPlanError{Reason: "failed to compute canonical subject hash", Cause: err}
	}
	if expectedHash != planHash {
		return &MigrationPlanError{Reason: "plan_hash does not match canonical subject digest"}
	}

	return nil
}

func parseMigrationPlanJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &MigrationPlanError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &MigrationPlanError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &MigrationPlanError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &MigrationPlanError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &MigrationPlanError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &MigrationPlanError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &MigrationPlanError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &MigrationPlanError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &MigrationPlanError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &MigrationPlanError{Reason: "trailing input"}
	}
	return record, nil
}

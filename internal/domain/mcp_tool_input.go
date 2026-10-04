package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

// MCPToolInputError describes an invalid MCP tool input argument without echoing private contents.
type MCPToolInputError struct {
	Tool   string
	Reason string
	Cause  error
}

// Error identifies the failed MCP tool input obligation.
func (e *MCPToolInputError) Error() string {
	if e.Tool != "" {
		return "invalid MCP tool input for " + e.Tool + ": " + e.Reason
	}
	return "invalid MCP tool input: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *MCPToolInputError) Unwrap() error { return e.Cause }

// ValidateMCPToolInput validates an MCP tool input record against its schema URI in specs/mcp/tools/*.
// It also accepts get-next-mission and mission schemas for unified schema validation.
func ValidateMCPToolInput(schemaURI string, raw []byte) error {
	switch schemaURI {
	case "https://rewrotron.dev/specs/mcp/tools/analysis.run.input.schema.json":
		return ValidateMCPAnalysisRunInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/assessment.get.input.schema.json":
		return ValidateMCPAssessmentGetInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/assessment.run.input.schema.json":
		return ValidateMCPAssessmentRunInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/behavior.discover.input.schema.json":
		return ValidateMCPBehaviorDiscoverInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/behavior.get_case.input.schema.json":
		return ValidateMCPBehaviorGetCaseInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/benchmark.run.input.schema.json":
		return ValidateMCPBenchmarkRunInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/cutover.prepare.input.schema.json":
		return ValidateMCPCutoverPrepareInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/decisions.get.input.schema.json":
		return ValidateMCPDecisionsGetInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/decisions.propose.input.schema.json":
		return ValidateMCPDecisionsProposeInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/knowledge.record_lesson.input.schema.json":
		return ValidateMCPKnowledgeRecordLessonInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/migration.apply_patch.input.schema.json":
		return ValidateMCPMigrationApplyPatchInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/migration.generate_code.input.schema.json":
		return ValidateMCPMigrationGenerateCodeInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/migration.get_plan.input.schema.json":
		return ValidateMCPMigrationGetPlanInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/migration.plan.input.schema.json":
		return ValidateMCPMigrationPlanInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/migration.stage_patch.input.schema.json":
		return ValidateMCPMigrationStagePatchInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/observation.run.input.schema.json":
		return ValidateMCPObservationRunInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.confirm_cutover.input.schema.json":
		return ValidateMCPProjectConfirmCutoverInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.confirm_plan.input.schema.json":
		return ValidateMCPProjectConfirmPlanInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.confirm_source.input.schema.json":
		return ValidateMCPProjectConfirmSourceInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.detect.input.schema.json":
		return ValidateMCPProjectDetectInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.doctor.input.schema.json":
		return ValidateMCPProjectDoctorInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.get_execution.input.schema.json":
		return ValidateMCPProjectGetExecutionInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.get_snapshot.input.schema.json":
		return ValidateMCPProjectGetSnapshotInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.init.input.schema.json":
		return ValidateMCPProjectInitInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.prepare_source_confirmation.input.schema.json":
		return ValidateMCPProjectPrepareSourceConfirmationInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/project.resume.input.schema.json":
		return ValidateMCPProjectResumeInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/reports.generate.input.schema.json":
		return ValidateMCPReportsGenerateInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/research.claim_mission.input.schema.json":
		return ValidateMCPResearchClaimMissionInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/research.complete_mission.input.schema.json":
		return ValidateMCPResearchCompleteMissionInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/research.get_next_mission.input.schema.json":
		return ValidateMCPResearchGetNextMissionInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/research.renew_mission.input.schema.json":
		return ValidateMCPResearchRenewMissionInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/skills.search.input.schema.json":
		return ValidateMCPSkillsSearchInput(raw)
	case "https://rewrotron.dev/specs/mcp/tools/verification.run_case.input.schema.json":
		return ValidateMCPVerificationRunCaseInput(raw)
	case "https://rewrotron.dev/specs/mcp/get-next-mission.schema.json":
		return ValidateGetNextMission(raw)
	case "https://rewrotron.dev/specs/missions/mission.schema.json":
		return ValidateMission(raw)
	default:
		return &MCPToolInputError{Reason: "unsupported MCP schema: " + schemaURI}
	}
}

// ValidateMCPToolInputByName validates tool input given the tool name (e.g. "project.init").
func ValidateMCPToolInputByName(toolName string, raw []byte) error {
	return ValidateMCPToolInput("https://rewrotron.dev/specs/mcp/tools/"+toolName+".input.schema.json", raw)
}

// ValidateMCPAnalysisRunInput validates input conforming to specs/mcp/tools/analysis.run.input.schema.json.
func ValidateMCPAnalysisRunInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "analysis.run")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "analysis.run", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "analysis.run"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "analysis.run"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "analysis.run", "project_id", "idempotency_key")
}

// ValidateMCPAssessmentGetInput validates input conforming to specs/mcp/tools/assessment.get.input.schema.json.
func ValidateMCPAssessmentGetInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "assessment.get")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "assessment.get", "assessment_ref", "project_id"); err != nil {
		return err
	}
	if err := mcpValidateRef(record["assessment_ref"], "assessment_ref", "assessment.get"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "assessment.get"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "assessment.get", "assessment_ref", "project_id")
}

// ValidateMCPAssessmentRunInput validates input conforming to specs/mcp/tools/assessment.run.input.schema.json.
func ValidateMCPAssessmentRunInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "assessment.run")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "assessment.run", "scope_ref", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateRef(record["scope_ref"], "scope_ref", "assessment.run"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "assessment.run"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "assessment.run"); err != nil {
		return err
	}
	if obsRaw, present := record["observation"]; present && string(obsRaw) != "null" {
		var obs string
		if err := json.Unmarshal(obsRaw, &obs); err != nil || (obs != "AVAILABLE_ONLY" && obs != "SAFE_RUNTIME") {
			return &MCPToolInputError{Tool: "assessment.run", Reason: "observation must be AVAILABLE_ONLY or SAFE_RUNTIME"}
		}
	}
	return mcpCheckUnexpectedFields(record, "assessment.run", "scope_ref", "observation", "project_id", "idempotency_key")
}

// ValidateMCPBehaviorDiscoverInput validates input conforming to specs/mcp/tools/behavior.discover.input.schema.json.
func ValidateMCPBehaviorDiscoverInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "behavior.discover")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "behavior.discover", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "behavior.discover"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "behavior.discover"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "behavior.discover", "project_id", "idempotency_key")
}

// ValidateMCPBehaviorGetCaseInput validates input conforming to specs/mcp/tools/behavior.get_case.input.schema.json.
func ValidateMCPBehaviorGetCaseInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "behavior.get_case")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "behavior.get_case", "project_id", "case_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "behavior.get_case"); err != nil {
		return err
	}
	if err := mcpValidateID(record["case_id"], "case_id", "behavior.get_case"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "behavior.get_case", "project_id", "case_id")
}

// ValidateMCPBenchmarkRunInput validates input conforming to specs/mcp/tools/benchmark.run.input.schema.json.
func ValidateMCPBenchmarkRunInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "benchmark.run")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "benchmark.run", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "benchmark.run"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "benchmark.run"); err != nil {
		return err
	}

	unitRaw, hasUnit := record["unit_id"]
	allRaw, hasAll := record["all"]
	if hasUnit && string(unitRaw) != "null" && hasAll && string(allRaw) != "null" {
		return &MCPToolInputError{Tool: "benchmark.run", Reason: "unit_id and all are mutually exclusive"}
	}
	if (!hasUnit || string(unitRaw) == "null") && (!hasAll || string(allRaw) == "null") {
		return &MCPToolInputError{Tool: "benchmark.run", Reason: "either unit_id or all must be specified"}
	}
	if hasUnit && string(unitRaw) != "null" {
		if err := mcpValidateID(unitRaw, "unit_id", "benchmark.run"); err != nil {
			return err
		}
	}
	if hasAll && string(allRaw) != "null" {
		var allVal bool
		if err := json.Unmarshal(allRaw, &allVal); err != nil || !allVal {
			return &MCPToolInputError{Tool: "benchmark.run", Reason: "all must be boolean true"}
		}
	}
	return mcpCheckUnexpectedFields(record, "benchmark.run", "project_id", "idempotency_key", "unit_id", "all")
}

// ValidateMCPCutoverPrepareInput validates input conforming to specs/mcp/tools/cutover.prepare.input.schema.json.
func ValidateMCPCutoverPrepareInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "cutover.prepare")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "cutover.prepare", "subject_hash", "output", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateDigest(record["subject_hash"], "subject_hash", "cutover.prepare"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["output"], "output", "cutover.prepare"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "cutover.prepare"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "cutover.prepare"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "cutover.prepare", "subject_hash", "output", "project_id", "idempotency_key")
}

// ValidateMCPDecisionsGetInput validates input conforming to specs/mcp/tools/decisions.get.input.schema.json.
func ValidateMCPDecisionsGetInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "decisions.get")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "decisions.get", "decision_id", "project_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["decision_id"], "decision_id", "decisions.get"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "decisions.get"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "decisions.get", "decision_id", "project_id")
}

// ValidateMCPDecisionsProposeInput validates input conforming to specs/mcp/tools/decisions.propose.input.schema.json.
func ValidateMCPDecisionsProposeInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "decisions.propose")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "decisions.propose", "decision_ref", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateRef(record["decision_ref"], "decision_ref", "decisions.propose"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "decisions.propose"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "decisions.propose"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "decisions.propose", "decision_ref", "project_id", "idempotency_key")
}

// ValidateMCPKnowledgeRecordLessonInput validates input conforming to specs/mcp/tools/knowledge.record_lesson.input.schema.json.
func ValidateMCPKnowledgeRecordLessonInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "knowledge.record_lesson")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "knowledge.record_lesson", "experiment_id", "finding", "remediation", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["experiment_id"], "experiment_id", "knowledge.record_lesson"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["finding"], "finding", "knowledge.record_lesson"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["remediation"], "remediation", "knowledge.record_lesson"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "knowledge.record_lesson"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "knowledge.record_lesson"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "knowledge.record_lesson", "experiment_id", "finding", "remediation", "project_id", "idempotency_key")
}

// ValidateMCPMigrationApplyPatchInput validates input conforming to specs/mcp/tools/migration.apply_patch.input.schema.json.
func ValidateMCPMigrationApplyPatchInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "migration.apply_patch")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "migration.apply_patch", "project_id", "patch_ref", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "migration.apply_patch"); err != nil {
		return err
	}
	if err := mcpValidateRef(record["patch_ref"], "patch_ref", "migration.apply_patch"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "migration.apply_patch"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "migration.apply_patch", "project_id", "patch_ref", "idempotency_key")
}

// ValidateMCPMigrationGenerateCodeInput validates input conforming to specs/mcp/tools/migration.generate_code.input.schema.json.
func ValidateMCPMigrationGenerateCodeInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "migration.generate_code")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "migration.generate_code", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "migration.generate_code"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "migration.generate_code"); err != nil {
		return err
	}

	unitRaw, hasUnit := record["unit_id"]
	allRaw, hasAll := record["all"]
	if hasUnit && string(unitRaw) != "null" && hasAll && string(allRaw) != "null" {
		return &MCPToolInputError{Tool: "migration.generate_code", Reason: "unit_id and all are mutually exclusive"}
	}
	if (!hasUnit || string(unitRaw) == "null") && (!hasAll || string(allRaw) == "null") {
		return &MCPToolInputError{Tool: "migration.generate_code", Reason: "either unit_id or all must be specified"}
	}
	if hasUnit && string(unitRaw) != "null" {
		if err := mcpValidateID(unitRaw, "unit_id", "migration.generate_code"); err != nil {
			return err
		}
	}
	if hasAll && string(allRaw) != "null" {
		var allVal bool
		if err := json.Unmarshal(allRaw, &allVal); err != nil || !allVal {
			return &MCPToolInputError{Tool: "migration.generate_code", Reason: "all must be boolean true"}
		}
	}
	return mcpCheckUnexpectedFields(record, "migration.generate_code", "project_id", "idempotency_key", "unit_id", "all")
}

// ValidateMCPMigrationGetPlanInput validates input conforming to specs/mcp/tools/migration.get_plan.input.schema.json.
func ValidateMCPMigrationGetPlanInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "migration.get_plan")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "migration.get_plan", "project_id", "plan_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "migration.get_plan"); err != nil {
		return err
	}
	if err := mcpValidateID(record["plan_id"], "plan_id", "migration.get_plan"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "migration.get_plan", "project_id", "plan_id")
}

// ValidateMCPMigrationPlanInput validates input conforming to specs/mcp/tools/migration.plan.input.schema.json.
func ValidateMCPMigrationPlanInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "migration.plan")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "migration.plan", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "migration.plan"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "migration.plan"); err != nil {
		return err
	}
	if tgtRaw, present := record["target"]; present && string(tgtRaw) != "null" {
		if _, err := mcpValidateNonEmptyString(tgtRaw, "target", "migration.plan"); err != nil {
			return err
		}
	}
	if styleRaw, present := record["style"]; present && string(styleRaw) != "null" {
		if _, err := mcpValidateNonEmptyString(styleRaw, "style", "migration.plan"); err != nil {
			return err
		}
	}
	if recRaw, present := record["recalculate"]; present && string(recRaw) != "null" {
		if _, err := mcpValidateBool(recRaw, "recalculate", "migration.plan"); err != nil {
			return err
		}
	}
	if selRaw, present := record["selection_ref"]; present && string(selRaw) != "null" {
		if err := mcpValidateRef(selRaw, "selection_ref", "migration.plan"); err != nil {
			return err
		}
	}
	if perfRaw, present := record["performance_policy_ref"]; present && string(perfRaw) != "null" {
		if err := mcpValidateRef(perfRaw, "performance_policy_ref", "migration.plan"); err != nil {
			return err
		}
	}
	return mcpCheckUnexpectedFields(record, "migration.plan", "target", "style", "recalculate", "project_id", "idempotency_key", "selection_ref", "performance_policy_ref")
}

// ValidateMCPMigrationStagePatchInput validates input conforming to specs/mcp/tools/migration.stage_patch.input.schema.json.
func ValidateMCPMigrationStagePatchInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "migration.stage_patch")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "migration.stage_patch", "project_id", "patch_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "migration.stage_patch"); err != nil {
		return err
	}
	if err := mcpValidateID(record["patch_id"], "patch_id", "migration.stage_patch"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "migration.stage_patch"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "migration.stage_patch", "project_id", "patch_id", "idempotency_key")
}

// ValidateMCPObservationRunInput validates input conforming to specs/mcp/tools/observation.run.input.schema.json.
func ValidateMCPObservationRunInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "observation.run")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "observation.run", "case_id", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["case_id"], "case_id", "observation.run"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "observation.run"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "observation.run"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "observation.run", "case_id", "project_id", "idempotency_key")
}

// ValidateMCPProjectConfirmCutoverInput validates input conforming to specs/mcp/tools/project.confirm_cutover.input.schema.json.
func ValidateMCPProjectConfirmCutoverInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.confirm_cutover")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.confirm_cutover", "project_id", "subject_hash", "approval_token", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.confirm_cutover"); err != nil {
		return err
	}
	if err := mcpValidateDigest(record["subject_hash"], "subject_hash", "project.confirm_cutover"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["approval_token"], "approval_token", "project.confirm_cutover"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.confirm_cutover"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.confirm_cutover", "project_id", "subject_hash", "approval_token", "idempotency_key")
}

// ValidateMCPProjectConfirmPlanInput validates input conforming to specs/mcp/tools/project.confirm_plan.input.schema.json.
func ValidateMCPProjectConfirmPlanInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.confirm_plan")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.confirm_plan", "project_id", "subject_hash", "approval_token", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.confirm_plan"); err != nil {
		return err
	}
	if err := mcpValidateDigest(record["subject_hash"], "subject_hash", "project.confirm_plan"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["approval_token"], "approval_token", "project.confirm_plan"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.confirm_plan"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.confirm_plan", "project_id", "subject_hash", "approval_token", "idempotency_key")
}

// ValidateMCPProjectConfirmSourceInput validates input conforming to specs/mcp/tools/project.confirm_source.input.schema.json.
func ValidateMCPProjectConfirmSourceInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.confirm_source")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.confirm_source", "project_id", "subject_hash", "approval_token", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.confirm_source"); err != nil {
		return err
	}
	if err := mcpValidateDigest(record["subject_hash"], "subject_hash", "project.confirm_source"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["approval_token"], "approval_token", "project.confirm_source"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.confirm_source"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.confirm_source", "project_id", "subject_hash", "approval_token", "idempotency_key")
}

// ValidateMCPProjectDetectInput validates input conforming to specs/mcp/tools/project.detect.input.schema.json.
func ValidateMCPProjectDetectInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.detect")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.detect", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.detect"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.detect"); err != nil {
		return err
	}
	if pathRaw, present := record["path"]; present && string(pathRaw) != "null" {
		if _, err := mcpValidateNonEmptyString(pathRaw, "path", "project.detect"); err != nil {
			return err
		}
	}
	if fastRaw, present := record["fast"]; present && string(fastRaw) != "null" {
		if _, err := mcpValidateBool(fastRaw, "fast", "project.detect"); err != nil {
			return err
		}
	}
	return mcpCheckUnexpectedFields(record, "project.detect", "path", "fast", "project_id", "idempotency_key")
}

// ValidateMCPProjectDoctorInput validates input conforming to specs/mcp/tools/project.doctor.input.schema.json.
func ValidateMCPProjectDoctorInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.doctor")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.doctor", "project_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.doctor"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.doctor", "project_id")
}

// ValidateMCPProjectGetExecutionInput validates input conforming to specs/mcp/tools/project.get_execution.input.schema.json.
func ValidateMCPProjectGetExecutionInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.get_execution")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.get_execution", "project_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.get_execution"); err != nil {
		return err
	}

	var afterSeq *int64
	if aRaw, present := record["after_sequence"]; present && string(aRaw) != "null" {
		val, err := mcpValidateNonNegativeInt(aRaw, "after_sequence", "project.get_execution")
		if err != nil {
			return err
		}
		afterSeq = &val
	}

	var throughSeq *int64
	if tRaw, present := record["through_sequence"]; present && string(tRaw) != "null" {
		val, err := mcpValidateNonNegativeInt(tRaw, "through_sequence", "project.get_execution")
		if err != nil {
			return err
		}
		throughSeq = &val
	}

	if afterSeq != nil && throughSeq != nil && *afterSeq > *throughSeq {
		return &MCPToolInputError{Tool: "project.get_execution", Reason: "after_sequence must be <= through_sequence"}
	}

	if lRaw, present := record["limit"]; present && string(lRaw) != "null" {
		val, err := mcpValidateNonNegativeInt(lRaw, "limit", "project.get_execution")
		if err != nil || val < 1 || val > 500 {
			return &MCPToolInputError{Tool: "project.get_execution", Reason: "limit must be an integer between 1 and 500"}
		}
	}

	return mcpCheckUnexpectedFields(record, "project.get_execution", "project_id", "after_sequence", "through_sequence", "limit")
}

// ValidateMCPProjectGetSnapshotInput validates input conforming to specs/mcp/tools/project.get_snapshot.input.schema.json.
func ValidateMCPProjectGetSnapshotInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.get_snapshot")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.get_snapshot", "project_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.get_snapshot"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.get_snapshot", "project_id")
}

// ValidateMCPProjectInitInput validates input conforming to specs/mcp/tools/project.init.input.schema.json.
func ValidateMCPProjectInitInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.init")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.init", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.init"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.init"); err != nil {
		return err
	}
	if pathRaw, present := record["path"]; present && string(pathRaw) != "null" {
		if _, err := mcpValidateNonEmptyString(pathRaw, "path", "project.init"); err != nil {
			return err
		}
	}
	return mcpCheckUnexpectedFields(record, "project.init", "path", "project_id", "idempotency_key")
}

// ValidateMCPProjectPrepareSourceConfirmationInput validates input conforming to specs/mcp/tools/project.prepare_source_confirmation.input.schema.json.
func ValidateMCPProjectPrepareSourceConfirmationInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.prepare_source_confirmation")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.prepare_source_confirmation", "project_id", "detection_id", "manifest_hash", "overrides", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.prepare_source_confirmation"); err != nil {
		return err
	}
	if err := mcpValidateID(record["detection_id"], "detection_id", "project.prepare_source_confirmation"); err != nil {
		return err
	}
	if err := mcpValidateDigest(record["manifest_hash"], "manifest_hash", "project.prepare_source_confirmation"); err != nil {
		return err
	}

	ovRaw := record["overrides"]
	if ovRaw == nil || string(ovRaw) == "null" {
		return &MCPToolInputError{Tool: "project.prepare_source_confirmation", Reason: "overrides is required"}
	}
	var overrides map[string]json.RawMessage
	if err := json.Unmarshal(ovRaw, &overrides); err != nil {
		return &MCPToolInputError{Tool: "project.prepare_source_confirmation", Reason: "overrides must be an object", Cause: err}
	}
	for k, v := range overrides {
		if _, err := mcpValidateNonEmptyString(v, "overrides."+k, "project.prepare_source_confirmation"); err != nil {
			return err
		}
	}

	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.prepare_source_confirmation"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.prepare_source_confirmation", "project_id", "detection_id", "manifest_hash", "overrides", "idempotency_key")
}

// ValidateMCPProjectResumeInput validates input conforming to specs/mcp/tools/project.resume.input.schema.json.
func ValidateMCPProjectResumeInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "project.resume")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "project.resume", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "project.resume"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "project.resume"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "project.resume", "project_id", "idempotency_key")
}

// ValidateMCPReportsGenerateInput validates input conforming to specs/mcp/tools/reports.generate.input.schema.json.
func ValidateMCPReportsGenerateInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "reports.generate")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "reports.generate", "format", "project_id", "idempotency_key"); err != nil {
		return err
	}

	var format string
	if err := json.Unmarshal(record["format"], &format); err != nil || (format != "json" && format != "markdown") {
		return &MCPToolInputError{Tool: "reports.generate", Reason: "format must be json or markdown"}
	}
	if err := mcpValidateID(record["project_id"], "project_id", "reports.generate"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "reports.generate"); err != nil {
		return err
	}

	var kind string = "auto"
	if kindRaw, present := record["kind"]; present && string(kindRaw) != "null" {
		if err := json.Unmarshal(kindRaw, &kind); err != nil || (kind != "auto" && kind != "assessment" && kind != "migration") {
			return &MCPToolInputError{Tool: "reports.generate", Reason: "invalid kind enum"}
		}
	}

	hasAssessmentRef := false
	if refRaw, present := record["assessment_ref"]; present && string(refRaw) != "null" {
		hasAssessmentRef = true
		if err := mcpValidateRef(refRaw, "assessment_ref", "reports.generate"); err != nil {
			return err
		}
	}

	// Schema allOf rule: if assessment_ref is specified, kind must be auto or assessment (cannot be migration).
	if hasAssessmentRef && kind == "migration" {
		return &MCPToolInputError{Tool: "reports.generate", Reason: "kind cannot be migration when assessment_ref is specified"}
	}

	return mcpCheckUnexpectedFields(record, "reports.generate", "format", "project_id", "idempotency_key", "kind", "assessment_ref")
}

// ValidateMCPResearchClaimMissionInput validates input conforming to specs/mcp/tools/research.claim_mission.input.schema.json.
func ValidateMCPResearchClaimMissionInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "research.claim_mission")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "research.claim_mission", "mission_id", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["mission_id"], "mission_id", "research.claim_mission"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "research.claim_mission"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "research.claim_mission"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "research.claim_mission", "mission_id", "project_id", "idempotency_key")
}

// ValidateMCPResearchCompleteMissionInput validates input conforming to specs/mcp/tools/research.complete_mission.input.schema.json.
func ValidateMCPResearchCompleteMissionInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "research.complete_mission")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "research.complete_mission", "mission_id", "fencing_token", "evidence_refs", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["mission_id"], "mission_id", "research.complete_mission"); err != nil {
		return err
	}
	if _, err := mcpValidateNonNegativeInt(record["fencing_token"], "fencing_token", "research.complete_mission"); err != nil {
		return err
	}

	var refs []json.RawMessage
	if err := json.Unmarshal(record["evidence_refs"], &refs); err != nil {
		return &MCPToolInputError{Tool: "research.complete_mission", Reason: "evidence_refs must be an array", Cause: err}
	}
	for _, ref := range refs {
		if !validProfileRef(ref, false) {
			return &MCPToolInputError{Tool: "research.complete_mission", Reason: "invalid evidence_ref item"}
		}
	}

	if err := mcpValidateID(record["project_id"], "project_id", "research.complete_mission"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "research.complete_mission"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "research.complete_mission", "mission_id", "fencing_token", "evidence_refs", "project_id", "idempotency_key")
}

// ValidateMCPResearchGetNextMissionInput validates input conforming to specs/mcp/tools/research.get_next_mission.input.schema.json.
func ValidateMCPResearchGetNextMissionInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "research.get_next_mission")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "research.get_next_mission", "project_id"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "research.get_next_mission"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "research.get_next_mission", "project_id")
}

// ValidateMCPResearchRenewMissionInput validates input conforming to specs/mcp/tools/research.renew_mission.input.schema.json.
func ValidateMCPResearchRenewMissionInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "research.renew_mission")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "research.renew_mission", "mission_id", "fencing_token", "project_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["mission_id"], "mission_id", "research.renew_mission"); err != nil {
		return err
	}
	if _, err := mcpValidateNonNegativeInt(record["fencing_token"], "fencing_token", "research.renew_mission"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "research.renew_mission"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "research.renew_mission"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "research.renew_mission", "mission_id", "fencing_token", "project_id", "idempotency_key")
}

// ValidateMCPSkillsSearchInput validates input conforming to specs/mcp/tools/skills.search.input.schema.json.
func ValidateMCPSkillsSearchInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "skills.search")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "skills.search", "query", "project_id"); err != nil {
		return err
	}
	if _, err := mcpValidateNonEmptyString(record["query"], "query", "skills.search"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "skills.search"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "skills.search", "query", "project_id")
}

// ValidateMCPVerificationRunCaseInput validates input conforming to specs/mcp/tools/verification.run_case.input.schema.json.
func ValidateMCPVerificationRunCaseInput(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "verification.run_case")
	if err != nil {
		return err
	}
	if err := mcpRequireFields(record, "verification.run_case", "project_id", "case_id", "idempotency_key"); err != nil {
		return err
	}
	if err := mcpValidateID(record["project_id"], "project_id", "verification.run_case"); err != nil {
		return err
	}
	if err := mcpValidateID(record["case_id"], "case_id", "verification.run_case"); err != nil {
		return err
	}
	if err := mcpValidateID(record["idempotency_key"], "idempotency_key", "verification.run_case"); err != nil {
		return err
	}
	return mcpCheckUnexpectedFields(record, "verification.run_case", "project_id", "case_id", "idempotency_key")
}

// ValidateGetNextMission validates responses conforming to specs/mcp/get-next-mission.schema.json.
func ValidateGetNextMission(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "get-next-mission")
	if err != nil {
		return err
	}

	missionRaw, hasMission := record["mission"]
	if !hasMission {
		return &MCPToolInputError{Tool: "get-next-mission", Reason: "missing required field: mission"}
	}

	reasonRaw, hasReason := record["reason"]
	if !hasReason || string(reasonRaw) == "null" {
		return &MCPToolInputError{Tool: "get-next-mission", Reason: "missing required field: reason"}
	}
	var reason string
	if err := json.Unmarshal(reasonRaw, &reason); err != nil {
		return &MCPToolInputError{Tool: "get-next-mission", Reason: "reason must be a string", Cause: err}
	}
	switch reason {
	case "READY", "NO_WORK", "DEPENDENCIES_BLOCKED", "CONFIRMATION_REQUIRED":
	default:
		return &MCPToolInputError{Tool: "get-next-mission", Reason: "invalid reason enum"}
	}

	if string(missionRaw) != "null" {
		if err := ValidateMission(missionRaw); err != nil {
			return &MCPToolInputError{Tool: "get-next-mission", Reason: "invalid mission", Cause: err}
		}
	}

	return mcpCheckUnexpectedFields(record, "get-next-mission", "mission", "reason")
}

// ValidateMission validates records conforming to specs/missions/mission.schema.json.
func ValidateMission(raw []byte) error {
	record, err := parseMCPToolInputJSON(raw, "mission")
	if err != nil {
		return err
	}

	required := []string{
		"mission_id",
		"type",
		"priority",
		"status",
		"unit_id",
		"case_id",
		"description",
		"context",
		"attempt_count",
		"fencing_token",
		"lease_expires_at",
	}
	if err := mcpRequireFields(record, "mission", required...); err != nil {
		return err
	}

	if err := mcpValidateID(record["mission_id"], "mission_id", "mission"); err != nil {
		return err
	}

	if string(record["type"]) == "null" {
		return &MCPToolInputError{Tool: "mission", Reason: "type is required"}
	}
	var mType string
	if err := json.Unmarshal(record["type"], &mType); err != nil {
		return &MCPToolInputError{Tool: "mission", Reason: "type must be a string"}
	}
	switch mType {
	case "ANALYZE_SOURCE", "DISCOVER_ROUTE", "OBSERVE_WORKFLOW", "GENERATE_CONTRACT",
		"MIGRATE_UNIT", "VERIFY_CASE", "REPAIR_MISMATCH", "ANALYZE_DEAD_CODE",
		"BENCHMARK", "OPTIMIZE", "REVIEW_UNKNOWN":
	default:
		return &MCPToolInputError{Tool: "mission", Reason: "invalid type enum"}
	}

	if _, err := mcpValidateNonNegativeInt(record["priority"], "priority", "mission"); err != nil {
		return err
	}

	if string(record["status"]) == "null" {
		return &MCPToolInputError{Tool: "mission", Reason: "status is required"}
	}
	var mStatus string
	if err := json.Unmarshal(record["status"], &mStatus); err != nil {
		return &MCPToolInputError{Tool: "mission", Reason: "status must be a string"}
	}
	switch mStatus {
	case "QUEUED", "LEASED", "RUNNING", "COMPLETED", "BLOCKED", "FAILED", "CANCELLED":
	default:
		return &MCPToolInputError{Tool: "mission", Reason: "invalid status enum"}
	}

	if string(record["unit_id"]) != "null" {
		var u string
		if err := json.Unmarshal(record["unit_id"], &u); err != nil {
			return &MCPToolInputError{Tool: "mission", Reason: "unit_id must be a string or null"}
		}
	}

	if string(record["case_id"]) != "null" {
		var c string
		if err := json.Unmarshal(record["case_id"], &c); err != nil {
			return &MCPToolInputError{Tool: "mission", Reason: "case_id must be a string or null"}
		}
	}

	if _, err := mcpValidateNonEmptyString(record["description"], "description", "mission"); err != nil {
		return err
	}

	if string(record["context"]) == "null" {
		return &MCPToolInputError{Tool: "mission", Reason: "context must be an object"}
	}
	var ctxObj map[string]json.RawMessage
	if err := json.Unmarshal(record["context"], &ctxObj); err != nil {
		return &MCPToolInputError{Tool: "mission", Reason: "context must be an object", Cause: err}
	}

	if _, err := mcpValidateNonNegativeInt(record["attempt_count"], "attempt_count", "mission"); err != nil {
		return err
	}

	if _, err := mcpValidateNonNegativeInt(record["fencing_token"], "fencing_token", "mission"); err != nil {
		return err
	}

	if string(record["lease_expires_at"]) != "null" {
		var exp string
		if err := json.Unmarshal(record["lease_expires_at"], &exp); err != nil {
			return &MCPToolInputError{Tool: "mission", Reason: "lease_expires_at must be date-time or null"}
		}
		if _, err := time.Parse(time.RFC3339, exp); err != nil {
			if _, errNano := time.Parse(time.RFC3339Nano, exp); errNano != nil {
				return &MCPToolInputError{Tool: "mission", Reason: "lease_expires_at must be valid RFC 3339 date-time"}
			}
		}
	}

	return mcpCheckUnexpectedFields(record, "mission", required...)
}

func mcpRequireFields(record map[string]json.RawMessage, tool string, fields ...string) error {
	for _, field := range fields {
		_, present := record[field]
		if !present {
			return &MCPToolInputError{Tool: tool, Reason: "missing required field: " + field}
		}
	}
	return nil
}

func mcpCheckUnexpectedFields(record map[string]json.RawMessage, tool string, allowed ...string) error {
	allowedMap := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedMap[a] = struct{}{}
	}
	for k := range record {
		if _, ok := allowedMap[k]; !ok {
			return &MCPToolInputError{Tool: tool, Reason: "unexpected field: " + k}
		}
	}
	return nil
}

func mcpValidateID(raw json.RawMessage, fieldName, tool string) error {
	if raw == nil || string(raw) == "null" {
		return &MCPToolInputError{Tool: tool, Reason: fieldName + " is required"}
	}
	var id string
	if err := json.Unmarshal(raw, &id); err != nil || !validAssessmentReportID(id) {
		return &MCPToolInputError{Tool: tool, Reason: fieldName + " must match pattern ^[a-z0-9_-]+$"}
	}
	return nil
}

func mcpValidateDigest(raw json.RawMessage, fieldName, tool string) error {
	if raw == nil || string(raw) == "null" {
		return &MCPToolInputError{Tool: tool, Reason: fieldName + " is required"}
	}
	var digest string
	if err := json.Unmarshal(raw, &digest); err != nil || len(digest) != 64 || !isManifestDigest(digest) {
		return &MCPToolInputError{Tool: tool, Reason: fieldName + " must be a 64-character lowercase hex digest"}
	}
	return nil
}

func mcpValidateRef(raw json.RawMessage, fieldName, tool string) error {
	if raw == nil || string(raw) == "null" {
		return &MCPToolInputError{Tool: tool, Reason: fieldName + " is required"}
	}
	if !validProfileRef(raw, false) {
		return &MCPToolInputError{Tool: tool, Reason: fieldName + " must be a sha256 reference"}
	}
	return nil
}

func mcpValidateNonEmptyString(raw json.RawMessage, fieldName, tool string) (string, error) {
	if raw == nil || string(raw) == "null" {
		return "", &MCPToolInputError{Tool: tool, Reason: fieldName + " is required"}
	}
	var val string
	if err := json.Unmarshal(raw, &val); err != nil || len(val) == 0 {
		return "", &MCPToolInputError{Tool: tool, Reason: fieldName + " must be a non-empty string"}
	}
	return val, nil
}

func mcpValidateNonNegativeInt(raw json.RawMessage, fieldName, tool string) (int64, error) {
	if raw == nil || string(raw) == "null" {
		return 0, &MCPToolInputError{Tool: tool, Reason: fieldName + " is required"}
	}
	if strings.ContainsAny(string(raw), ".eE") {
		return 0, &MCPToolInputError{Tool: tool, Reason: fieldName + " must be an integer"}
	}
	var val int64
	if err := json.Unmarshal(raw, &val); err != nil || val < 0 {
		return 0, &MCPToolInputError{Tool: tool, Reason: fieldName + " must be an integer >= 0"}
	}
	return val, nil
}

func mcpValidateBool(raw json.RawMessage, fieldName, tool string) (bool, error) {
	if raw == nil || string(raw) == "null" {
		return false, &MCPToolInputError{Tool: tool, Reason: fieldName + " is required"}
	}
	var val bool
	if err := json.Unmarshal(raw, &val); err != nil {
		return false, &MCPToolInputError{Tool: tool, Reason: fieldName + " must be a boolean"}
	}
	return val, nil
}

func parseMCPToolInputJSON(raw []byte, tool string) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &MCPToolInputError{Tool: tool, Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &MCPToolInputError{Tool: tool, Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &MCPToolInputError{Tool: tool, Reason: "object expected"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &MCPToolInputError{Tool: tool, Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &MCPToolInputError{Tool: tool, Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &MCPToolInputError{Tool: tool, Reason: "duplicate object key: " + key}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &MCPToolInputError{Tool: tool, Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &MCPToolInputError{Tool: tool, Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &MCPToolInputError{Tool: tool, Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &MCPToolInputError{Tool: tool, Reason: "trailing input"}
	}
	return record, nil
}

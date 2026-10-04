package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	actorIDPattern     = regexp.MustCompile(`^[a-z0-9_-]+$`)
	operationIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
	terminalIDPattern  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	caseIDPattern      = regexp.MustCompile(`^[a-z0-9_-]+$`)
	contractIDPattern  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	selectorPattern    = regexp.MustCompile(`^/(raw_target|headers|body_json)/`)
)

var allowedTerminalOutcomes = map[string]struct{}{
	"SUCCESS":     {},
	"FAILURE":     {},
	"UNSUPPORTED": {},
}

var allowedBindingEncodings = map[string]struct{}{
	"PERCENT_ENCODED": {},
	"HEADER_VALUE":    {},
	"JSON_VALUE":      {},
}

var allowedCaseHTTPMethods = map[string]struct{}{
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"PATCH":   {},
	"DELETE":  {},
	"OPTIONS": {},
	"HEAD":    {},
}

// WorkflowError describes a validation or execution failure in a workflow scenario.
type WorkflowError struct {
	Code   string
	Reason string
	Cause  error
}

func (e *WorkflowError) Error() string {
	if e.Code != "" {
		if e.Cause != nil {
			return fmt.Sprintf("workflow error %s: %s: %v", e.Code, e.Reason, e.Cause)
		}
		return fmt.Sprintf("workflow error %s: %s", e.Code, e.Reason)
	}
	if e.Cause != nil {
		return fmt.Sprintf("workflow error: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("workflow error: %s", e.Reason)
}

func (e *WorkflowError) Unwrap() error { return e.Cause }

// BehaviorCaseError describes a validation failure in a behavior case definition.
type BehaviorCaseError struct {
	Reason string
	Cause  error
}

func (e *BehaviorCaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("behavior case error: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("behavior case error: %s", e.Reason)
}

func (e *BehaviorCaseError) Unwrap() error { return e.Cause }

// WorkflowV2 represents an executable bounded scenario graph conforming to specs/behavior/workflow.schema.json v2.0.
type WorkflowV2 struct {
	SchemaVersion    string             `json:"schema_version"`
	WorkflowID       string             `json:"workflow_id"`
	ProfileRef       string             `json:"profile_ref"`
	InitialStateRef  string             `json:"initial_state_ref"`
	Actors           []WorkflowActor    `json:"actors"`
	StartStepID      string             `json:"start_step_id"`
	Steps            []WorkflowStepV2   `json:"steps"`
	Terminals        []WorkflowTerminal `json:"terminals"`
	MaxExecutedSteps int                `json:"max_executed_steps"`
	EvidenceRefs     []string           `json:"evidence_refs"`
}

// WorkflowActor declares a scenario participant with its initial principal fixture.
type WorkflowActor struct {
	ActorID             string `json:"actor_id"`
	PrincipalFixtureRef string `json:"principal_fixture_ref"`
}

// WorkflowStepV2 declares one step in a workflow scenario.
type WorkflowStepV2 struct {
	StepID      string                     `json:"step_id"`
	ActorID     string                     `json:"actor_id"`
	OperationID string                     `json:"operation_id"`
	RequestRef  string                     `json:"request_ref"`
	Guard       *SemanticsExpression       `json:"guard"`
	Bindings    map[string]WorkflowBinding `json:"bindings"`
	Transitions []StepTransition           `json:"transitions"`
	Completion  SemanticsCompletion        `json:"completion"`
}

// WorkflowBinding binds a preceding evaluated value into a designated request site.
type WorkflowBinding struct {
	Value    *SemanticsExpression `json:"value"`
	Selector string               `json:"selector"`
	Encoding string               `json:"encoding"`
}

// WorkflowTerminal defines an explicit terminal state of a workflow scenario.
type WorkflowTerminal struct {
	TerminalID         string                 `json:"terminal_id"`
	Outcome            string                 `json:"outcome"`
	StateAssertions    []*SemanticsExpression `json:"state_assertions"`
	ExpectedEffectRefs []string               `json:"expected_effect_refs"`
}

// WorkflowV1 represents historical/descriptive workflow data (schema version 1).
type WorkflowV1 struct {
	WorkflowID    string            `json:"workflow_id"`
	Steps         []WorkflowStepV1  `json:"steps"`
	FixtureRefs   []string          `json:"fixture_refs"`
	EvidenceIDs   []string          `json:"evidence_ids"`
	OpenQuestions []string          `json:"open_questions"`
}

// WorkflowStepV1 represents a step in a historical v1 workflow.
type WorkflowStepV1 struct {
	StepID        string            `json:"step_id"`
	OperationID   string            `json:"operation_id"`
	RequestCaseID string            `json:"request_case_id"`
	Extract       []any             `json:"extract"`
	Bindings      map[string]string `json:"bindings"`
	MaxRepeats    int               `json:"max_repeats"`
}

// BehaviorCaseV2 represents a behavior case conforming to specs/behavior/behavior-case.schema.json v2.0.
type BehaviorCaseV2 struct {
	SchemaVersion      string                 `json:"schema_version"`
	CaseID             string                 `json:"case_id"`
	ProjectID          string                 `json:"project_id"`
	ContractRef        string                 `json:"contract_ref"`
	WorkflowRef        string                 `json:"workflow_ref"`
	ProfileRef         string                 `json:"profile_ref"`
	FixtureRefs        []string               `json:"fixture_refs"`
	RequestRefs        []string               `json:"request_refs"`
	PolicyRef          string                 `json:"policy_ref"`
	SourceEvidenceRefs []string               `json:"source_evidence_refs"`
	ExpectedAssertions []*SemanticsExpression `json:"expected_assertions"`
}

// BehaviorCaseV1 represents a historical/descriptive behavior case conforming to version 1.0.
type BehaviorCaseV1 struct {
	SchemaVersion       string         `json:"schema_version"`
	CaseID              string         `json:"case_id"`
	ContractID          string         `json:"contract_id"`
	ContractVersion     string         `json:"contract_version"`
	Description         *string        `json:"description,omitempty"`
	Entrypoint          map[string]any `json:"entrypoint"`
	HTTPMethod          *string        `json:"http_method,omitempty"`
	Path                *string        `json:"path,omitempty"`
	Headers             map[string]any `json:"headers"`
	QueryParams         map[string]any `json:"query_params,omitempty"`
	Payload             any            `json:"payload,omitempty"`
	Preconditions       map[string]any `json:"preconditions"`
	ExpectedResponse    map[string]any `json:"expected_response"`
	RequiredDimensions  []string       `json:"required_dimensions"`
	TimeoutMs           *int           `json:"timeout_ms,omitempty"`
	ComparisonPolicyRef string         `json:"comparison_policy_ref"`
}

// ValidateWorkflow validates raw bytes against specs/behavior/workflow.schema.json (v2.0 or v1).
func ValidateWorkflow(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return &WorkflowError{Reason: "workflow must be JSON object", Cause: err}
	}

	if verRaw, ok := m["schema_version"]; ok {
		var ver string
		if err := json.Unmarshal(verRaw, &ver); err != nil {
			return &WorkflowError{Reason: "schema_version must be string", Cause: err}
		}
		if ver == "2.0" {
			return validateWorkflowV2(m)
		}
		return &WorkflowError{Reason: fmt.Sprintf("unsupported schema_version %q", ver)}
	}

	// Historical v1
	return validateWorkflowV1(m)
}

func validateWorkflowV2(m map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"workflow_id",
		"profile_ref",
		"initial_state_ref",
		"actors",
		"start_step_id",
		"steps",
		"terminals",
		"max_executed_steps",
		"evidence_refs",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return &WorkflowError{Reason: "workflow 2.0 missing required field " + f}
		}
	}
	if len(m) != len(expectedFields) {
		return &WorkflowError{Reason: "workflow 2.0 contains unrecognized fields"}
	}

	var wfID, profRef, initRef, startStepID string
	if err := json.Unmarshal(m["workflow_id"], &wfID); err != nil || !workflowIDPattern.MatchString(wfID) {
		return &WorkflowError{Reason: "invalid workflow_id pattern"}
	}
	if err := json.Unmarshal(m["profile_ref"], &profRef); err != nil || !digestPattern.MatchString(profRef) {
		return &WorkflowError{Reason: "invalid profile_ref pattern"}
	}
	if err := json.Unmarshal(m["initial_state_ref"], &initRef); err != nil || !digestPattern.MatchString(initRef) {
		return &WorkflowError{Reason: "invalid initial_state_ref pattern"}
	}
	if err := json.Unmarshal(m["start_step_id"], &startStepID); err != nil || !stepIDPattern.MatchString(startStepID) {
		return &WorkflowError{Reason: "invalid start_step_id pattern"}
	}

	// Actors
	var rawActors []json.RawMessage
	if err := json.Unmarshal(m["actors"], &rawActors); err != nil || len(rawActors) < 1 {
		return &WorkflowError{Reason: "actors must be non-empty array"}
	}
	seenActors := make(map[string]struct{}, len(rawActors))
	for _, ra := range rawActors {
		var act map[string]json.RawMessage
		if err := json.Unmarshal(ra, &act); err != nil || len(act) != 2 {
			return &WorkflowError{Reason: "actor must have exactly actor_id and principal_fixture_ref"}
		}
		var aID, pRef string
		if err := json.Unmarshal(act["actor_id"], &aID); err != nil || !actorIDPattern.MatchString(aID) {
			return &WorkflowError{Reason: "invalid actor_id pattern"}
		}
		if err := json.Unmarshal(act["principal_fixture_ref"], &pRef); err != nil || !digestPattern.MatchString(pRef) {
			return &WorkflowError{Reason: "invalid principal_fixture_ref pattern"}
		}
		if _, exists := seenActors[aID]; exists {
			return &WorkflowError{Reason: fmt.Sprintf("duplicate actor_id %q", aID)}
		}
		seenActors[aID] = struct{}{}
	}

	// Steps
	var rawSteps []json.RawMessage
	if err := json.Unmarshal(m["steps"], &rawSteps); err != nil || len(rawSteps) < 1 || len(rawSteps) > 100 {
		return &WorkflowError{Reason: "steps must be array of 1..100 items"}
	}
	nodeCounter := 0
	for _, rs := range rawSteps {
		var step map[string]json.RawMessage
		if err := json.Unmarshal(rs, &step); err != nil {
			return &WorkflowError{Reason: "step must be object"}
		}
		stepFields := []string{"step_id", "actor_id", "operation_id", "request_ref", "guard", "bindings", "transitions", "completion"}
		for _, sf := range stepFields {
			if _, ok := step[sf]; !ok {
				return &WorkflowError{Code: "ERR_INVALID_CONTRACT", Reason: "step missing required field " + sf}
			}
		}
		if len(step) != len(stepFields) {
			return &WorkflowError{Reason: "step contains unrecognized fields"}
		}

		var sID, aID, opID, rRef string
		if err := json.Unmarshal(step["step_id"], &sID); err != nil || !stepIDPattern.MatchString(sID) {
			return &WorkflowError{Reason: "invalid step_id pattern"}
		}
		if err := json.Unmarshal(step["actor_id"], &aID); err != nil || !actorIDPattern.MatchString(aID) {
			return &WorkflowError{Reason: "invalid step actor_id pattern"}
		}
		if _, exists := seenActors[aID]; !exists {
			return &WorkflowError{Reason: fmt.Sprintf("step references undeclared actor %q", aID)}
		}
		if err := json.Unmarshal(step["operation_id"], &opID); err != nil || !operationIDPattern.MatchString(opID) {
			return &WorkflowError{Reason: "invalid operation_id pattern"}
		}
		if err := json.Unmarshal(step["request_ref"], &rRef); err != nil || !digestPattern.MatchString(rRef) {
			return &WorkflowError{Reason: "invalid request_ref pattern"}
		}

		// Step guard must be literal true in v2 subset
		guardExpr, err := validateExpressionRaw(step["guard"], 1, &nodeCounter)
		if err != nil {
			return &WorkflowError{Code: "ERR_UNSUPPORTED_CAPABILITY", Reason: "invalid step guard expression", Cause: err}
		}
		if err := ValidateStepGuard(guardExpr); err != nil {
			return err
		}

		// Bindings
		var rawBindings map[string]json.RawMessage
		if err := json.Unmarshal(step["bindings"], &rawBindings); err != nil {
			return &WorkflowError{Reason: "bindings must be object"}
		}
		for bName, rb := range rawBindings {
			var b map[string]json.RawMessage
			if err := json.Unmarshal(rb, &b); err != nil || len(b) != 3 {
				return &WorkflowError{Reason: fmt.Sprintf("binding %q must have exactly value, selector, encoding", bName)}
			}
			valRaw, ok1 := b["value"]
			selRaw, ok2 := b["selector"]
			encRaw, ok3 := b["encoding"]
			if !ok1 || !ok2 || !ok3 {
				return &WorkflowError{Reason: fmt.Sprintf("binding %q missing required fields", bName)}
			}
			var selector, encoding string
			if err := json.Unmarshal(selRaw, &selector); err != nil || !selectorPattern.MatchString(selector) {
				return &WorkflowError{Reason: fmt.Sprintf("binding %q invalid selector pattern", bName)}
			}
			if err := json.Unmarshal(encRaw, &encoding); err != nil {
				return &WorkflowError{Reason: fmt.Sprintf("binding %q encoding must be string", bName)}
			}
			if _, ok := allowedBindingEncodings[encoding]; !ok {
				return &WorkflowError{Reason: fmt.Sprintf("binding %q unsupported encoding %q", bName, encoding)}
			}
			if _, err := validateExpressionRaw(valRaw, 1, &nodeCounter); err != nil {
				return err
			}
		}

		// Transitions
		var rawTransitions []json.RawMessage
		if err := json.Unmarshal(step["transitions"], &rawTransitions); err != nil || len(rawTransitions) < 1 {
			return &WorkflowError{Reason: "transitions must be non-empty array"}
		}
		for _, rt := range rawTransitions {
			var t map[string]json.RawMessage
			if err := json.Unmarshal(rt, &t); err != nil || len(t) != 3 {
				return &WorkflowError{Reason: "transition must have exactly guard, next_step_id, terminal_id"}
			}
			gRaw, ok1 := t["guard"]
			nextRaw, ok2 := t["next_step_id"]
			termRaw, ok3 := t["terminal_id"]
			if !ok1 || !ok2 || !ok3 {
				return &WorkflowError{Reason: "transition missing required fields"}
			}
			if _, err := validateExpressionRaw(gRaw, 1, &nodeCounter); err != nil {
				return err
			}
			hasNext := string(nextRaw) != "null"
			hasTerm := string(termRaw) != "null"
			if (hasNext && hasTerm) || (!hasNext && !hasTerm) {
				return &WorkflowError{Reason: "transition must specify exactly one of next_step_id or terminal_id"}
			}
			if hasNext {
				var nid string
				if err := json.Unmarshal(nextRaw, &nid); err != nil || !stepIDPattern.MatchString(nid) {
					return &WorkflowError{Reason: "invalid next_step_id pattern"}
				}
			}
			if hasTerm {
				var tid string
				if err := json.Unmarshal(termRaw, &tid); err != nil || !terminalIDPattern.MatchString(tid) {
					return &WorkflowError{Reason: "invalid terminal_id pattern"}
				}
			}
		}

		// Completion
		if _, err := validateCompletionRaw(step["completion"], &nodeCounter); err != nil {
			return &WorkflowError{Reason: "invalid step completion", Cause: err}
		}
	}

	// Terminals
	var rawTerminals []json.RawMessage
	if err := json.Unmarshal(m["terminals"], &rawTerminals); err != nil || len(rawTerminals) < 1 {
		return &WorkflowError{Reason: "terminals must be non-empty array"}
	}
	seenTerminals := make(map[string]struct{}, len(rawTerminals))
	for _, rt := range rawTerminals {
		var term map[string]json.RawMessage
		if err := json.Unmarshal(rt, &term); err != nil || len(term) != 4 {
			return &WorkflowError{Reason: "terminal must have exactly terminal_id, outcome, state_assertions, expected_effect_refs"}
		}
		var tID, outcome string
		if err := json.Unmarshal(term["terminal_id"], &tID); err != nil || !terminalIDPattern.MatchString(tID) {
			return &WorkflowError{Reason: "invalid terminal_id pattern"}
		}
		if _, exists := seenTerminals[tID]; exists {
			return &WorkflowError{Reason: fmt.Sprintf("duplicate terminal_id %q", tID)}
		}
		seenTerminals[tID] = struct{}{}

		if err := json.Unmarshal(term["outcome"], &outcome); err != nil {
			return &WorkflowError{Reason: "terminal outcome must be string"}
		}
		if _, ok := allowedTerminalOutcomes[outcome]; !ok {
			return &WorkflowError{Reason: fmt.Sprintf("unsupported terminal outcome %q", outcome)}
		}

		var rawAsserts []json.RawMessage
		if err := json.Unmarshal(term["state_assertions"], &rawAsserts); err != nil {
			return &WorkflowError{Reason: "state_assertions must be array"}
		}
		for _, ra := range rawAsserts {
			if _, err := validateExpressionRaw(ra, 1, &nodeCounter); err != nil {
				return err
			}
		}

		var rawEffects []string
		if err := json.Unmarshal(term["expected_effect_refs"], &rawEffects); err != nil {
			return &WorkflowError{Reason: "expected_effect_refs must be array of strings"}
		}
		for _, ref := range rawEffects {
			if !digestPattern.MatchString(ref) {
				return &WorkflowError{Reason: "invalid expected_effect_ref pattern"}
			}
		}
	}

	// Max executed steps
	var maxSteps int
	if err := json.Unmarshal(m["max_executed_steps"], &maxSteps); err != nil || maxSteps < 1 || maxSteps > 100 {
		return &WorkflowError{Reason: "max_executed_steps must be 1..100"}
	}

	// Evidence refs
	var evRefs []string
	if err := json.Unmarshal(m["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
		return &WorkflowError{Reason: "evidence_refs must be non-empty array of strings"}
	}
	for _, ref := range evRefs {
		if !digestPattern.MatchString(ref) {
			return &WorkflowError{Reason: "invalid evidence_ref pattern"}
		}
	}

	return nil
}

func validateWorkflowV1(m map[string]json.RawMessage) error {
	expectedFields := []string{"workflow_id", "steps", "fixture_refs", "evidence_ids", "open_questions"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return &WorkflowError{Reason: "workflow 1.0 missing required field " + f}
		}
	}
	if len(m) != len(expectedFields) {
		return &WorkflowError{Reason: "workflow 1.0 contains unrecognized fields"}
	}
	var wfID string
	if err := json.Unmarshal(m["workflow_id"], &wfID); err != nil || !workflowIDPattern.MatchString(wfID) {
		return &WorkflowError{Reason: "invalid workflow_id pattern"}
	}
	return nil
}

// ValidateBehaviorCase validates raw bytes against specs/behavior/behavior-case.schema.json (v2.0 or v1.0).
func ValidateBehaviorCase(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return &BehaviorCaseError{Reason: "behavior case must be JSON object", Cause: err}
	}

	verRaw, ok := m["schema_version"]
	if !ok {
		return &BehaviorCaseError{Reason: "missing schema_version"}
	}
	var ver string
	if err := json.Unmarshal(verRaw, &ver); err != nil {
		return &BehaviorCaseError{Reason: "schema_version must be string"}
	}

	if ver == "2.0" {
		return validateBehaviorCaseV2(m)
	}
	if ver == "1.0" {
		return validateBehaviorCaseV1(m)
	}
	return &BehaviorCaseError{Reason: fmt.Sprintf("unsupported schema_version %q", ver)}
}

func validateBehaviorCaseV2(m map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"case_id",
		"project_id",
		"contract_ref",
		"workflow_ref",
		"profile_ref",
		"fixture_refs",
		"request_refs",
		"policy_ref",
		"source_evidence_refs",
		"expected_assertions",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return &BehaviorCaseError{Reason: "behavior-case 2.0 missing required field " + f}
		}
	}
	if len(m) != len(expectedFields) {
		return &BehaviorCaseError{Reason: "behavior-case 2.0 contains unrecognized fields"}
	}

	var caseID, projectID string
	if err := json.Unmarshal(m["case_id"], &caseID); err != nil || !caseIDPattern.MatchString(caseID) {
		return &BehaviorCaseError{Reason: "invalid case_id pattern"}
	}
	if err := json.Unmarshal(m["project_id"], &projectID); err != nil || !projectIDPattern.MatchString(projectID) {
		return &BehaviorCaseError{Reason: "invalid project_id pattern"}
	}

	digests := []string{"contract_ref", "workflow_ref", "profile_ref", "policy_ref"}
	for _, df := range digests {
		var d string
		if err := json.Unmarshal(m[df], &d); err != nil || !digestPattern.MatchString(d) {
			return &BehaviorCaseError{Reason: fmt.Sprintf("invalid %s digest pattern", df)}
		}
	}

	nodeCounter := 0
	var rawAsserts []json.RawMessage
	if err := json.Unmarshal(m["expected_assertions"], &rawAsserts); err != nil || len(rawAsserts) < 1 {
		return &BehaviorCaseError{Reason: "expected_assertions must be non-empty array"}
	}
	for _, ra := range rawAsserts {
		if _, err := validateExpressionRaw(ra, 1, &nodeCounter); err != nil {
			return &BehaviorCaseError{Reason: "invalid expected assertion", Cause: err}
		}
	}
	return nil
}

func validateBehaviorCaseV1(m map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"case_id",
		"contract_id",
		"contract_version",
		"entrypoint",
		"headers",
		"preconditions",
		"expected_response",
		"required_dimensions",
		"comparison_policy_ref",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return &BehaviorCaseError{Reason: "behavior-case 1.0 missing required field " + f}
		}
	}

	var caseID, contractID, contractVer, compRef string
	if err := json.Unmarshal(m["case_id"], &caseID); err != nil || !caseIDPattern.MatchString(caseID) {
		return &BehaviorCaseError{Reason: "invalid case_id pattern"}
	}
	if err := json.Unmarshal(m["contract_id"], &contractID); err != nil || !contractIDPattern.MatchString(contractID) {
		return &BehaviorCaseError{Reason: "invalid contract_id pattern"}
	}
	if err := json.Unmarshal(m["contract_version"], &contractVer); err != nil || len(contractVer) < 1 {
		return &BehaviorCaseError{Reason: "contract_version must be non-empty string"}
	}
	if err := json.Unmarshal(m["comparison_policy_ref"], &compRef); err != nil || !digestPattern.MatchString(compRef) {
		return &BehaviorCaseError{Reason: "invalid comparison_policy_ref pattern"}
	}

	var entrypoint map[string]json.RawMessage
	if err := json.Unmarshal(m["entrypoint"], &entrypoint); err != nil {
		return &BehaviorCaseError{Reason: "entrypoint must be object"}
	}

	var epType string
	if epTRaw, ok := entrypoint["type"]; ok {
		_ = json.Unmarshal(epTRaw, &epType)
	}

	if epType == "http" {
		methodRaw, ok1 := m["http_method"]
		pathRaw, ok2 := m["path"]
		if !ok1 || !ok2 {
			return &BehaviorCaseError{Reason: "http entrypoint requires http_method and path"}
		}
		var method, path string
		if err := json.Unmarshal(methodRaw, &method); err != nil {
			return &BehaviorCaseError{Reason: "http_method must be string"}
		}
		if _, ok := allowedCaseHTTPMethods[method]; !ok {
			return &BehaviorCaseError{Reason: fmt.Sprintf("unsupported http_method %q", method)}
		}
		if err := json.Unmarshal(pathRaw, &path); err != nil || len(path) < 1 {
			return &BehaviorCaseError{Reason: "path must be non-empty string"}
		}

		var expResp map[string]json.RawMessage
		if err := json.Unmarshal(m["expected_response"], &expResp); err != nil {
			return &BehaviorCaseError{Reason: "expected_response must be object"}
		}
		statusRaw, okStatus := expResp["status"]
		if !okStatus {
			return &BehaviorCaseError{Reason: "expected_response missing status for http entrypoint"}
		}
		var status int
		if err := json.Unmarshal(statusRaw, &status); err != nil || status < 100 || status > 599 {
			return &BehaviorCaseError{Reason: "status must be 100..599"}
		}
	}

	return nil
}

// ParseWorkflow parses and validates raw JSON into a WorkflowV2.
func ParseWorkflow(raw []byte) (*WorkflowV2, error) {
	if err := ValidateWorkflow(raw); err != nil {
		return nil, err
	}
	var wf WorkflowV2
	if err := json.Unmarshal(raw, &wf); err != nil {
		return nil, &WorkflowError{Reason: "failed to unmarshal workflow", Cause: err}
	}
	return &wf, nil
}

// ParseBehaviorCase parses and validates raw JSON into a BehaviorCaseV2.
func ParseBehaviorCase(raw []byte) (*BehaviorCaseV2, error) {
	if err := ValidateBehaviorCase(raw); err != nil {
		return nil, err
	}
	var bc BehaviorCaseV2
	if err := json.Unmarshal(raw, &bc); err != nil {
		return nil, &BehaviorCaseError{Reason: "failed to unmarshal behavior case", Cause: err}
	}
	return &bc, nil
}

// ApplicationRunner represents the interface through which the scenario driver invokes the application under test.
type ApplicationRunner interface {
	ExecuteStep(req *ExecutionRequest) (*ExecutionResponse, error)
}

// ExecutionRequest specifies the bound parameters sent to the application runner.
type ExecutionRequest struct {
	Side                 string              `json:"side"`
	StepID               string              `json:"step_id"`
	Visit                int                 `json:"visit"`
	PhysicalOccurrenceID string              `json:"physical_occurrence_id"`
	ActorID              string              `json:"actor_id"`
	OperationID          string              `json:"operation_id"`
	RequestRef           string              `json:"request_ref"`
	EffectiveTarget      string              `json:"effective_target"`
	EffectiveHeaders     []EvaluationHeader  `json:"effective_headers"`
	EffectiveBody        string              `json:"effective_body"`
	Cookies              map[string]string   `json:"cookies"`
}

// ExecutionResponse contains the actual witnessed result produced by the application runner.
type ExecutionResponse struct {
	ExecutionWitnessRef string            `json:"execution_witness_ref"`
	PreContext          *EvaluationContext `json:"pre_context"`
	PostContext         *EvaluationContext `json:"post_context"`
	Cookies             map[string]string `json:"cookies"`
	EvidenceRefs        []string          `json:"evidence_refs"`
	ApplicationError    *string           `json:"application_error,omitempty"`
}

// WorkflowExecutionState tracks step execution and occurrence history for a scenario on one side.
type WorkflowExecutionState struct {
	WorkflowID        string                 `json:"workflow_id"`
	ProjectID         string                 `json:"project_id"`
	Side              string                 `json:"side"`
	CurrentStepID     string                 `json:"current_step_id"`
	StepCount         int                    `json:"step_count"`
	MaxExecutedSteps  int                    `json:"max_executed_steps"`
	VisitCounts       map[string]int         `json:"visit_counts"`
	StepHistory       []StepOccurrence       `json:"step_history"`
	CommittedBindings map[string]TaggedValue `json:"committed_bindings"`
	CookieJar         map[string]string      `json:"cookie_jar"`
	Witness           *ExecutionWitness      `json:"witness"`
	TerminalReached   *WorkflowTerminal      `json:"terminal_reached,omitempty"`
	Halted            bool                   `json:"halted"`
	HaltReason        string                 `json:"halt_reason,omitempty"`
}

// PairedExecutionResult records the paired execution of a scenario across SOURCE and TARGET.
type PairedExecutionResult struct {
	WorkflowID      string                  `json:"workflow_id"`
	SourceState     *WorkflowExecutionState `json:"source_state"`
	TargetState     *WorkflowExecutionState `json:"target_state"`
	Diverged        bool                    `json:"diverged"`
	FirstDivergence *BranchDivergence       `json:"first_divergence,omitempty"`
}

// BranchDivergence describes the exact point where source and target took different transitions (X64).
type BranchDivergence struct {
	StepID           string  `json:"step_id"`
	Visit            int     `json:"visit"`
	SourceNextStepID *string `json:"source_next_step_id"`
	SourceTerminalID *string `json:"source_terminal_id"`
	SourceContextRef string  `json:"source_context_ref"`
	TargetNextStepID *string `json:"target_next_step_id"`
	TargetTerminalID *string `json:"target_terminal_id"`
	TargetContextRef string  `json:"target_context_ref"`
	Reason           string  `json:"reason"`
}

// WorkflowScenarioDriver coordinates bounded workflow scenario execution.
type WorkflowScenarioDriver struct{}

// NewWorkflowScenarioDriver constructs a new WorkflowScenarioDriver.
func NewWorkflowScenarioDriver() *WorkflowScenarioDriver {
	return &WorkflowScenarioDriver{}
}

// InitializeState sets up a new execution state for a workflow on a designated side.
func (d *WorkflowScenarioDriver) InitializeState(wf *WorkflowV2, projectID string, side string, witness *ExecutionWitness) (*WorkflowExecutionState, error) {
	if wf == nil {
		return nil, errors.New("nil workflow")
	}
	if side != SideSource && side != SideTarget {
		return nil, fmt.Errorf("invalid side %q", side)
	}

	return &WorkflowExecutionState{
		WorkflowID:        wf.WorkflowID,
		ProjectID:         projectID,
		Side:              side,
		CurrentStepID:     wf.StartStepID,
		StepCount:         0,
		MaxExecutedSteps:  wf.MaxExecutedSteps,
		VisitCounts:       make(map[string]int),
		StepHistory:       make([]StepOccurrence, 0),
		CommittedBindings: make(map[string]TaggedValue),
		CookieJar:         make(map[string]string),
		Witness:           witness,
		Halted:            false,
	}, nil
}

// ExecuteNextStep executes the current step in the workflow state using the provided runner.
func (d *WorkflowScenarioDriver) ExecuteNextStep(state *WorkflowExecutionState, wf *WorkflowV2, runner ApplicationRunner) (*StepOccurrence, error) {
	if state.Halted {
		return nil, fmt.Errorf("workflow execution already halted: %s", state.HaltReason)
	}

	step := findStep(wf.Steps, state.CurrentStepID)
	if step == nil {
		state.Halted = true
		state.HaltReason = fmt.Sprintf("current step %q not found in workflow", state.CurrentStepID)
		return nil, errors.New(state.HaltReason)
	}

	// 1. Validate step guard: must be literal true in v2 subset (X03 / B2)
	// Refuses before dispatch with zero runner calls if invalid
	if err := ValidateStepGuard(step.Guard); err != nil {
		state.Halted = true
		state.HaltReason = fmt.Sprintf("step guard refused before dispatch: %v", err)
		return nil, err
	}

	// 2. Check execution step bound
	if state.StepCount >= state.MaxExecutedSteps {
		state.Halted = true
		state.HaltReason = "exceeded max_executed_steps"
		return nil, &WorkflowError{Code: "ERR_UNSUPPORTED_CAPABILITY", Reason: state.HaltReason}
	}

	// 3. Increment visit count for logical occurrence identity (X65)
	visit := state.VisitCounts[step.StepID] + 1
	state.VisitCounts[step.StepID] = visit
	state.StepCount++

	physOccID := fmt.Sprintf("occ-%s-%s-%d-%s", strings.ToLower(state.Side), step.StepID, visit, shortHash(state.WorkflowID+step.StepID+strconv.Itoa(visit)))

	// 4. Input preparation: apply bindings
	req := &ExecutionRequest{
		Side:                 state.Side,
		StepID:               step.StepID,
		Visit:                visit,
		PhysicalOccurrenceID: physOccID,
		ActorID:              step.ActorID,
		OperationID:          step.OperationID,
		RequestRef:           step.RequestRef,
		EffectiveTarget:      "/", // Default base target
		EffectiveHeaders:     make([]EvaluationHeader, 0),
		EffectiveBody:        "{}",
		Cookies:              copyCookieMap(state.CookieJar),
	}

	// Prepare preceding context for binding evaluation
	var precedingCtx *EvaluationContext
	if len(state.StepHistory) > 0 {
		// Use bindings accumulated so far
		precedingCtx = &EvaluationContext{
			Namespaces: EvaluationContextNamespaces{
				Binding: state.CommittedBindings,
			},
		}
	}

	for bName, bDef := range step.Bindings {
		if err := applyBinding(bName, bDef, precedingCtx, req); err != nil {
			state.Halted = true
			state.HaltReason = fmt.Sprintf("failed to apply binding %q: %v", bName, err)
			return nil, err
		}
	}

	// 5. Invoke application runner boundary
	resp, err := runner.ExecuteStep(req)
	if err != nil {
		state.Halted = true
		state.HaltReason = fmt.Sprintf("runner execution failed: %v", err)
		occ := StepOccurrence{
			SchemaVersion:        "1.0",
			ProjectID:            state.ProjectID,
			Side:                 state.Side,
			LogicalWorkflowID:    state.WorkflowID,
			StepID:               step.StepID,
			Visit:                visit,
			PhysicalOccurrenceID: physOccID,
			ExecutionWitnessRef:  "sha256:" + strings.Repeat("0", 64),
			PreContextRef:        "sha256:" + strings.Repeat("0", 64),
			PostContextRef:       nil,
			GuardResults:         []StepGuardResult{},
			Outcome:              StepOutcomeInfrastructureFailure,
			EvidenceRefs:         []string{"sha256:" + strings.Repeat("0", 64)},
		}
		state.StepHistory = append(state.StepHistory, occ)
		return &occ, err
	}

	// 6. Check execution witness integration (X66)
	if resp.ExecutionWitnessRef == "" || !digestPattern.MatchString(resp.ExecutionWitnessRef) {
		state.Halted = true
		state.HaltReason = "missing or invalid execution witness from application runner"
		occ := StepOccurrence{
			SchemaVersion:        "1.0",
			ProjectID:            state.ProjectID,
			Side:                 state.Side,
			LogicalWorkflowID:    state.WorkflowID,
			StepID:               step.StepID,
			Visit:                visit,
			PhysicalOccurrenceID: physOccID,
			ExecutionWitnessRef:  "sha256:" + strings.Repeat("0", 64),
			PreContextRef:        "sha256:" + strings.Repeat("0", 64),
			PostContextRef:       nil,
			GuardResults:         []StepGuardResult{},
			Outcome:              StepOutcomeInfrastructureFailure,
			EvidenceRefs:         resolveEvidenceRefs(resp.EvidenceRefs, "sha256:"+strings.Repeat("0", 64)),
		}
		state.StepHistory = append(state.StepHistory, occ)
		return &occ, errors.New(state.HaltReason)
	}

	// 7. Update side-local cookie jar
	if resp.Cookies != nil {
		for ck, cv := range resp.Cookies {
			state.CookieJar[ck] = cv
		}
	}

	// 8. Capture context references
	preRef := "sha256:" + strings.Repeat("a", 64)
	if resp.PreContext != nil {
		preRef = resp.PreContext.ProjectionRef
	}
	var postRef *string
	if resp.PostContext != nil {
		postRef = &resp.PostContext.ProjectionRef
	}

	stepEvidence := resolveEvidenceRefs(resp.EvidenceRefs, resp.ExecutionWitnessRef)

	// 9. Evaluate completion boundary
	if step.Completion.Mode == "SINK_ACK_QUIESCENCE" {
		if resp.PostContext == nil {
			state.Halted = true
			state.HaltReason = "async completion incomplete"
			occ := StepOccurrence{
				SchemaVersion:        "1.0",
				ProjectID:            state.ProjectID,
				Side:                 state.Side,
				LogicalWorkflowID:    state.WorkflowID,
				StepID:               step.StepID,
				Visit:                visit,
				PhysicalOccurrenceID: physOccID,
				ExecutionWitnessRef:  resp.ExecutionWitnessRef,
				PreContextRef:        preRef,
				PostContextRef:       postRef,
				GuardResults:         []StepGuardResult{},
				Outcome:              StepOutcomeUnknown,
				EvidenceRefs:         stepEvidence,
			}
			state.StepHistory = append(state.StepHistory, occ)
			return &occ, nil
		}
	}

	// 10. Evaluate transition guards in order (X04 / B2)
	guardResults := make([]StepGuardResult, 0, len(step.Transitions))
	var selectedNextStep *string
	var selectedTerminal *string

	for i, tr := range step.Transitions {
		matched, verdict, err := EvaluateTransitionGuard(tr.Guard, resp.PostContext)
		guardID := fmt.Sprintf("guard-%d", i)
		var guardOutcome string
		if verdict == "PASS" || matched {
			guardOutcome = GuardOutcomeTrue
		} else if verdict == "MISMATCH" {
			guardOutcome = GuardOutcomeFalse
		} else {
			guardOutcome = GuardOutcomeUnknown
		}

		guardResults = append(guardResults, StepGuardResult{
			GuardID:      guardID,
			Outcome:      guardOutcome,
			EvidenceRefs: []string{resp.ExecutionWitnessRef},
		})

		if err != nil || verdict == "UNKNOWN" {
			// UNKNOWN stops immediately and cannot fall through!
			state.Halted = true
			state.HaltReason = "transition guard evaluated to UNKNOWN"
			occ := StepOccurrence{
				SchemaVersion:        "1.0",
				ProjectID:            state.ProjectID,
				Side:                 state.Side,
				LogicalWorkflowID:    state.WorkflowID,
				StepID:               step.StepID,
				Visit:                visit,
				PhysicalOccurrenceID: physOccID,
				ExecutionWitnessRef:  resp.ExecutionWitnessRef,
				PreContextRef:        preRef,
				PostContextRef:       postRef,
				GuardResults:         guardResults,
				SelectedNextStepID:   nil,
				TerminalID:           nil,
				Outcome:              StepOutcomeUnknown,
				EvidenceRefs:         stepEvidence,
			}
			state.StepHistory = append(state.StepHistory, occ)
			return &occ, nil
		}

		if matched {
			selectedNextStep = tr.NextStepID
			selectedTerminal = tr.TerminalID
			break
		}
	}

	if selectedNextStep == nil && selectedTerminal == nil {
		state.Halted = true
		state.HaltReason = "no transition matched and no fallback taken"
		occ := StepOccurrence{
			SchemaVersion:        "1.0",
			ProjectID:            state.ProjectID,
			Side:                 state.Side,
			LogicalWorkflowID:    state.WorkflowID,
			StepID:               step.StepID,
			Visit:                visit,
			PhysicalOccurrenceID: physOccID,
			ExecutionWitnessRef:  resp.ExecutionWitnessRef,
			PreContextRef:        preRef,
			PostContextRef:       postRef,
			GuardResults:         guardResults,
			SelectedNextStepID:   nil,
			TerminalID:           nil,
			Outcome:              StepOutcomeUnknown,
			EvidenceRefs:         stepEvidence,
		}
		state.StepHistory = append(state.StepHistory, occ)
		return &occ, errors.New(state.HaltReason)
	}

	occ := StepOccurrence{
		SchemaVersion:        "1.0",
		ProjectID:            state.ProjectID,
		Side:                 state.Side,
		LogicalWorkflowID:    state.WorkflowID,
		StepID:               step.StepID,
		Visit:                visit,
		PhysicalOccurrenceID: physOccID,
		ExecutionWitnessRef:  resp.ExecutionWitnessRef,
		PreContextRef:        preRef,
		PostContextRef:       postRef,
		GuardResults:         guardResults,
		SelectedNextStepID:   selectedNextStep,
		TerminalID:           selectedTerminal,
		Outcome:              StepOutcomeCompleted,
		EvidenceRefs:         stepEvidence,
	}

	state.StepHistory = append(state.StepHistory, occ)

	// Advance workflow state or record terminal
	if selectedNextStep != nil {
		state.CurrentStepID = *selectedNextStep
	} else if selectedTerminal != nil {
		term := findTerminal(wf.Terminals, *selectedTerminal)
		state.TerminalReached = term
		state.Halted = true
		state.HaltReason = fmt.Sprintf("terminal reached: %s", *selectedTerminal)
	}

	return &occ, nil
}

// RunWorkflow executes steps in sequence until a terminal state is reached, the scenario halts, or max_executed_steps is reached.
func (d *WorkflowScenarioDriver) RunWorkflow(state *WorkflowExecutionState, wf *WorkflowV2, runner ApplicationRunner) ([]StepOccurrence, error) {
	for !state.Halted {
		_, err := d.ExecuteNextStep(state, wf, runner)
		if err != nil {
			return state.StepHistory, err
		}
	}
	return state.StepHistory, nil
}

// RunPairedWorkflow executes a workflow scenario on both SOURCE and TARGET sides in step-by-step lockstep.
// It implements STOP_AT_FIRST_DIVERGENCE (X64): if source and target take differing branches, execution halts
// immediately on both sides and records the divergence point without forcing alignment or fabricating requests.
func (d *WorkflowScenarioDriver) RunPairedWorkflow(
	wf *WorkflowV2,
	projectID string,
	sourceRunner, targetRunner ApplicationRunner,
	sourceWitness, targetWitness *ExecutionWitness,
) (*PairedExecutionResult, error) {
	sourceState, err := d.InitializeState(wf, projectID, SideSource, sourceWitness)
	if err != nil {
		return nil, fmt.Errorf("failed to init source state: %w", err)
	}
	targetState, err := d.InitializeState(wf, projectID, SideTarget, targetWitness)
	if err != nil {
		return nil, fmt.Errorf("failed to init target state: %w", err)
	}

	result := &PairedExecutionResult{
		WorkflowID:  wf.WorkflowID,
		SourceState: sourceState,
		TargetState: targetState,
		Diverged:    false,
	}

	for !sourceState.Halted && !targetState.Halted {
		// Both sides must be at the same declared step before dispatch
		if sourceState.CurrentStepID != targetState.CurrentStepID {
			result.Diverged = true
			result.FirstDivergence = &BranchDivergence{
				StepID:           sourceState.CurrentStepID,
				Visit:            sourceState.VisitCounts[sourceState.CurrentStepID],
				SourceNextStepID: stringPtr(sourceState.CurrentStepID),
				TargetNextStepID: stringPtr(targetState.CurrentStepID),
				Reason:           "current step mismatch before dispatch",
			}
			sourceState.Halted = true
			sourceState.HaltReason = "STOP_AT_FIRST_DIVERGENCE: step divergence"
			targetState.Halted = true
			targetState.HaltReason = "STOP_AT_FIRST_DIVERGENCE: step divergence"
			return result, nil
		}

		currentStep := sourceState.CurrentStepID

		// Execute on SOURCE
		occSrc, errSrc := d.ExecuteNextStep(sourceState, wf, sourceRunner)
		if errSrc != nil {
			return result, fmt.Errorf("source step %s failed: %w", currentStep, errSrc)
		}

		// Execute on TARGET
		occTgt, errTgt := d.ExecuteNextStep(targetState, wf, targetRunner)
		if errTgt != nil {
			return result, fmt.Errorf("target step %s failed: %w", currentStep, errTgt)
		}

		// Check branch selection equivalence
		srcNext := occSrc.SelectedNextStepID
		tgtNext := occTgt.SelectedNextStepID
		srcTerm := occSrc.TerminalID
		tgtTerm := occTgt.TerminalID

		branchDiverged := false
		if !equalStringPtr(srcNext, tgtNext) || !equalStringPtr(srcTerm, tgtTerm) {
			branchDiverged = true
		}

		if branchDiverged {
			result.Diverged = true
			srcCtxRef := ""
			if occSrc.PostContextRef != nil {
				srcCtxRef = *occSrc.PostContextRef
			}
			tgtCtxRef := ""
			if occTgt.PostContextRef != nil {
				tgtCtxRef = *occTgt.PostContextRef
			}

			result.FirstDivergence = &BranchDivergence{
				StepID:           currentStep,
				Visit:            occSrc.Visit,
				SourceNextStepID: srcNext,
				SourceTerminalID: srcTerm,
				SourceContextRef: srcCtxRef,
				TargetNextStepID: tgtNext,
				TargetTerminalID: tgtTerm,
				TargetContextRef: tgtCtxRef,
				Reason:           fmt.Sprintf("branch divergence at step %s (visit %d)", currentStep, occSrc.Visit),
			}

			// STOP_AT_FIRST_DIVERGENCE: Halt both sides immediately
			sourceState.Halted = true
			sourceState.HaltReason = "STOP_AT_FIRST_DIVERGENCE: branch divergence"
			targetState.Halted = true
			targetState.HaltReason = "STOP_AT_FIRST_DIVERGENCE: branch divergence"
			return result, nil
		}
	}

	return result, nil
}

// applyBinding resolves an input binding and applies it to the execution request.
func applyBinding(name string, b WorkflowBinding, precedingCtx *EvaluationContext, req *ExecutionRequest) error {
	val, err := EvaluateExpression(b.Value, precedingCtx)
	if err != nil {
		return err
	}
	if val.Type == ValueTypeMissing {
		return &WorkflowError{Code: "ERR_UNBOUND_VALUE", Reason: fmt.Sprintf("unbound value for binding %q", name)}
	}

	valStr := fmt.Sprint(val.Value)
	if val.Type == ValueTypeNull {
		valStr = ""
	}

	if strings.HasPrefix(b.Selector, "/raw_target/") {
		marker := "{" + name + "}"
		param := strings.TrimPrefix(b.Selector, "/raw_target/")
		altMarker := "{" + param + "}"
		encoded := valStr
		if b.Encoding == "PERCENT_ENCODED" {
			encoded = rfc3986Escape(valStr)
		}
		if strings.Contains(req.EffectiveTarget, marker) {
			req.EffectiveTarget = strings.ReplaceAll(req.EffectiveTarget, marker, encoded)
		} else if strings.Contains(req.EffectiveTarget, altMarker) {
			req.EffectiveTarget = strings.ReplaceAll(req.EffectiveTarget, altMarker, encoded)
		} else {
			return &WorkflowError{Code: "ERR_INVALID_CONTRACT", Reason: fmt.Sprintf("raw_target has no marker %q or %q for binding", marker, altMarker)}
		}
		return nil
	}

	if strings.HasPrefix(b.Selector, "/headers/") {
		headerName := strings.TrimPrefix(b.Selector, "/headers/")
		if strings.ContainsAny(valStr, "\r\n") {
			return &WorkflowError{Code: "ERR_INVALID_CONTRACT", Reason: "header value contains CR/LF"}
		}
		req.EffectiveHeaders = append(req.EffectiveHeaders, EvaluationHeader{
			Name:  headerName,
			Value: valStr,
		})
		return nil
	}

	if strings.HasPrefix(b.Selector, "/body_json/") {
		ptr := strings.TrimPrefix(b.Selector, "/body_json")
		var bodyObj map[string]any
		if req.EffectiveBody == "" || req.EffectiveBody == "{}" {
			bodyObj = make(map[string]any)
		} else {
			if err := json.Unmarshal([]byte(req.EffectiveBody), &bodyObj); err != nil {
				bodyObj = make(map[string]any)
			}
		}
		// Single-level property or pointer setter
		propName := strings.TrimPrefix(ptr, "/")
		if b.Encoding == "JSON_VALUE" {
			switch val.Type {
			case ValueTypeInteger:
				intVal, _ := strconv.ParseInt(valStr, 10, 64)
				bodyObj[propName] = intVal
			case ValueTypeBoolean:
				bodyObj[propName] = val.Value.(bool)
			case ValueTypeDecimal:
				bodyObj[propName] = valStr
			case ValueTypeNull:
				bodyObj[propName] = nil
			default:
				bodyObj[propName] = valStr
			}
		} else {
			bodyObj[propName] = valStr
		}
		bytes, err := json.Marshal(bodyObj)
		if err != nil {
			return err
		}
		req.EffectiveBody = string(bytes)
		return nil
	}

	return &WorkflowError{Reason: fmt.Sprintf("unsupported selector %q", b.Selector)}
}

func rfc3986Escape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func findStep(steps []WorkflowStepV2, stepID string) *WorkflowStepV2 {
	for i := range steps {
		if steps[i].StepID == stepID {
			return &steps[i]
		}
	}
	return nil
}

func findTerminal(terminals []WorkflowTerminal, terminalID string) *WorkflowTerminal {
	for i := range terminals {
		if terminals[i].TerminalID == terminalID {
			return &terminals[i]
		}
	}
	return nil
}

func copyCookieMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func equalStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

func resolveEvidenceRefs(refs []string, fallback string) []string {
	valid := make([]string, 0, len(refs))
	for _, r := range refs {
		if digestPattern.MatchString(r) {
			valid = append(valid, r)
		}
	}
	if len(valid) > 0 {
		return valid
	}
	if fallback != "" && digestPattern.MatchString(fallback) {
		return []string{fallback}
	}
	return []string{"sha256:" + strings.Repeat("0", 64)}
}

func stringPtr(s string) *string {
	return &s
}



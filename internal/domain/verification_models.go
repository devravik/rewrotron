package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var (
	runIDPattern           = regexp.MustCompile(`^[a-z0-9_-]+$`)
	ruleIDPattern          = regexp.MustCompile(`^[a-z0-9_-]+$`)
	relationGroupPattern   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	bindingIDPattern       = regexp.MustCompile(`^[a-z0-9_-]+$`)
	logicalKeyPattern      = regexp.MustCompile(`^[a-z0-9_-]+$`)
	mismatchIDPattern      = regexp.MustCompile(`^[a-z0-9_-]+$`)
	canonicalIntDecPattern = regexp.MustCompile(`^(?:0|-?[1-9][0-9]*)$`)
	sha256HexPattern       = regexp.MustCompile(`^[a-f0-9]{64}$`)
	connectionPattern      = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

var allowedAcceptanceScopeKinds = map[string]struct{}{
	"UNIT":                   {},
	"IMPLEMENTED_REGRESSION": {},
	"PROJECT":                {},
}

var allowedNormalizationValueTypes = map[string]struct{}{
	"string":    {},
	"boolean":   {},
	"integer":   {},
	"decimal":   {},
	"null":      {},
	"timestamp": {},
	"bytes":     {},
	"object":    {},
	"array":     {},
}

var allowedNormalizationVariabilities = map[string]struct{}{
	"CLOCK":          {},
	"GENERATED_ID":   {},
	"TOKEN":          {},
	"REPRESENTATION": {},
}

var allowedNormalizationRelations = map[string]struct{}{
	"BIJECTIVE_ID":             {},
	"TOKEN_LIFECYCLE":          {},
	"CONTROLLED_CLOCK":         {},
	"CANONICAL_REPRESENTATION": {},
}

var allowedPairingOutcomes = map[string]struct{}{
	"ALIGNED":  {},
	"DIVERGED": {},
	"UNKNOWN":  {},
}

var allowedDivergenceReasons = map[string]struct{}{
	"BRANCH":            {},
	"EARLY_TERMINAL":    {},
	"UNEXPECTED_BRANCH": {},
	"UNKNOWN_GUARD":     {},
}

var allowedReportStatuses = map[string]struct{}{
	"COMPLETED": {},
	"FAILED":    {},
	"CANCELLED": {},
}

var allowedReportVerdicts = map[string]struct{}{
	"PASS":     {},
	"MISMATCH": {},
	"UNKNOWN":  {},
}

var reportDimensionNames = []string{
	"http",
	"body",
	"authentication",
	"authorization",
	"session",
	"database",
	"external",
	"files",
	"cache",
	"errors",
	"execution",
}

// VerificationErrorDetail represents an error in a verification report.
type VerificationErrorDetail struct {
	Code              string         `json:"code"`
	Message           string         `json:"message"`
	Category          string         `json:"category"`
	Retryable         bool           `json:"retryable"`
	RecommendedAction string         `json:"recommended_action"`
	Details           map[string]any `json:"details,omitempty"`
	ArtifactRef       *string        `json:"artifact_ref,omitempty"`
}

// AcceptanceScope declares the evaluated boundary of a verification run.
type AcceptanceScope struct {
	Kind      string   `json:"kind"`
	UnitIDs   []string `json:"unit_ids"`
	CaseIDs   []string `json:"case_ids"`
	CorpusRef string   `json:"corpus_ref"`
	ScopeHash string   `json:"scope_hash"`
}

// ValidateAcceptanceScope validates raw bytes against specs/verification/acceptance-scope.schema.json.
func ValidateAcceptanceScope(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("acceptance scope must be JSON object: %w", err)
	}

	expectedFields := []string{"kind", "unit_ids", "case_ids", "corpus_ref", "scope_hash"}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("acceptance scope missing required field: %s", f)
		}
	}
	if len(m) != len(expectedFields) {
		return errors.New("acceptance scope contains unrecognized fields")
	}

	var kind, corpusRef, scopeHash string
	if err := json.Unmarshal(m["kind"], &kind); err != nil {
		return errors.New("kind must be string")
	}
	if _, ok := allowedAcceptanceScopeKinds[kind]; !ok {
		return fmt.Errorf("invalid acceptance scope kind: %q", kind)
	}

	if err := json.Unmarshal(m["corpus_ref"], &corpusRef); err != nil || !digestPattern.MatchString(corpusRef) {
		return errors.New("corpus_ref must match sha256 digest pattern")
	}
	if err := json.Unmarshal(m["scope_hash"], &scopeHash); err != nil || !sha256HexPattern.MatchString(scopeHash) {
		return errors.New("scope_hash must match 64-hex pattern")
	}

	var unitIDs []string
	if err := json.Unmarshal(m["unit_ids"], &unitIDs); err != nil {
		return errors.New("unit_ids must be array of strings")
	}
	seenUnits := make(map[string]struct{}, len(unitIDs))
	for _, u := range unitIDs {
		if !projectIDPattern.MatchString(u) {
			return fmt.Errorf("invalid unit_id pattern: %q", u)
		}
		if _, exists := seenUnits[u]; exists {
			return fmt.Errorf("duplicate unit_id: %q", u)
		}
		seenUnits[u] = struct{}{}
	}

	var caseIDs []string
	if err := json.Unmarshal(m["case_ids"], &caseIDs); err != nil || len(caseIDs) < 1 {
		return errors.New("case_ids must be non-empty array of strings")
	}
	seenCases := make(map[string]struct{}, len(caseIDs))
	for _, c := range caseIDs {
		if !caseIDPattern.MatchString(c) {
			return fmt.Errorf("invalid case_id pattern: %q", c)
		}
		if _, exists := seenCases[c]; exists {
			return fmt.Errorf("duplicate case_id: %q", c)
		}
		seenCases[c] = struct{}{}
	}

	return nil
}

// ParseAcceptanceScope parses and validates raw JSON into an AcceptanceScope.
func ParseAcceptanceScope(raw []byte) (*AcceptanceScope, error) {
	if err := ValidateAcceptanceScope(raw); err != nil {
		return nil, err
	}
	var scope AcceptanceScope
	if err := json.Unmarshal(raw, &scope); err != nil {
		return nil, err
	}
	return &scope, nil
}

// NormalizationRule declares an approved variation rule for a dynamic field.
type NormalizationRule struct {
	SchemaVersion           string   `json:"schema_version"`
	RuleID                  string   `json:"rule_id"`
	Path                    string   `json:"path"`
	ValueType               string   `json:"value_type"`
	Variability             string   `json:"variability"`
	Relation                string   `json:"relation"`
	RelationGroup           string   `json:"relation_group"`
	LifecycleContractRef    string   `json:"lifecycle_contract_ref"`
	BusinessSignificanceRef string   `json:"business_significance_ref"`
	DecisionRef             string   `json:"decision_ref"`
	NegativeMutationRefs    []string `json:"negative_mutation_refs"`
}

// ValidateNormalizationRule validates raw bytes against specs/verification/normalization.schema.json.
func ValidateNormalizationRule(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("normalization rule must be JSON object: %w", err)
	}

	expectedFields := []string{
		"schema_version",
		"rule_id",
		"path",
		"value_type",
		"variability",
		"relation",
		"relation_group",
		"lifecycle_contract_ref",
		"business_significance_ref",
		"decision_ref",
		"negative_mutation_refs",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("normalization rule missing required field: %s", f)
		}
	}
	if len(m) != len(expectedFields) {
		return errors.New("normalization rule contains unrecognized fields")
	}

	var schemaVer, ruleID, path, valType, variability, relation, relGroup string
	var lifeRef, busRef, decRef string

	if err := json.Unmarshal(m["schema_version"], &schemaVer); err != nil || schemaVer != "2.0" {
		return errors.New("schema_version must be '2.0'")
	}
	if err := json.Unmarshal(m["rule_id"], &ruleID); err != nil || !ruleIDPattern.MatchString(ruleID) {
		return errors.New("invalid rule_id pattern")
	}
	if err := json.Unmarshal(m["path"], &path); err != nil || len(path) < 1 {
		return errors.New("path must be non-empty string")
	}
	if err := json.Unmarshal(m["value_type"], &valType); err != nil {
		return errors.New("value_type must be string")
	}
	if _, ok := allowedNormalizationValueTypes[valType]; !ok {
		return fmt.Errorf("unsupported value_type: %q", valType)
	}
	if err := json.Unmarshal(m["variability"], &variability); err != nil {
		return errors.New("variability must be string")
	}
	if _, ok := allowedNormalizationVariabilities[variability]; !ok {
		return fmt.Errorf("unsupported variability: %q", variability)
	}
	if err := json.Unmarshal(m["relation"], &relation); err != nil {
		return errors.New("relation must be string")
	}
	if _, ok := allowedNormalizationRelations[relation]; !ok {
		return fmt.Errorf("unsupported relation: %q", relation)
	}
	if err := json.Unmarshal(m["relation_group"], &relGroup); err != nil || !relationGroupPattern.MatchString(relGroup) {
		return errors.New("invalid relation_group pattern")
	}

	for _, refPair := range []struct {
		name string
		val  *string
	}{
		{"lifecycle_contract_ref", &lifeRef},
		{"business_significance_ref", &busRef},
		{"decision_ref", &decRef},
	} {
		if err := json.Unmarshal(m[refPair.name], refPair.val); err != nil || !digestPattern.MatchString(*refPair.val) {
			return fmt.Errorf("invalid %s digest pattern", refPair.name)
		}
	}

	var mutRefs []string
	if err := json.Unmarshal(m["negative_mutation_refs"], &mutRefs); err != nil || len(mutRefs) < 1 {
		return errors.New("negative_mutation_refs must be non-empty array")
	}
	seenMuts := make(map[string]struct{}, len(mutRefs))
	for _, mr := range mutRefs {
		if !digestPattern.MatchString(mr) {
			return fmt.Errorf("invalid negative_mutation_ref digest: %q", mr)
		}
		if _, exists := seenMuts[mr]; exists {
			return fmt.Errorf("duplicate negative_mutation_ref: %q", mr)
		}
		seenMuts[mr] = struct{}{}
	}

	return nil
}

// ParseNormalizationRule parses and validates raw JSON into a NormalizationRule.
func ParseNormalizationRule(raw []byte) (*NormalizationRule, error) {
	if err := ValidateNormalizationRule(raw); err != nil {
		return nil, err
	}
	var rule NormalizationRule
	if err := json.Unmarshal(raw, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// OccurrencePairing records the pairing between SOURCE and TARGET occurrences in a scenario.
type OccurrencePairing struct {
	SchemaVersion       string             `json:"schema_version"`
	ProjectID           string             `json:"project_id"`
	CaseID              string             `json:"case_id"`
	ScenarioRef         string             `json:"scenario_ref"`
	LogicalWorkflowID   string             `json:"logical_workflow_id"`
	Policy              string             `json:"policy"`
	Pairs               []OccurrencePair   `json:"pairs"`
	FirstDivergence     *PairingDivergence `json:"first_divergence"`
	UnmatchedSourceRefs []string           `json:"unmatched_source_refs"`
	UnmatchedTargetRefs []string           `json:"unmatched_target_refs"`
	CleanupRefs         []string           `json:"cleanup_refs"`
}

// OccurrencePair records the pairing of a single step occurrence across source and target.
type OccurrencePair struct {
	StepID              string `json:"step_id"`
	Visit               int    `json:"visit"`
	SourceOccurrenceRef string `json:"source_occurrence_ref"`
	TargetOccurrenceRef string `json:"target_occurrence_ref"`
	Outcome             string `json:"outcome"`
}

// PairingDivergence records the exact point where source and target branches diverged.
type PairingDivergence struct {
	SourceOccurrenceRef string   `json:"source_occurrence_ref"`
	TargetOccurrenceRef string   `json:"target_occurrence_ref"`
	Reason              string   `json:"reason"`
	EvidenceRefs        []string `json:"evidence_refs"`
}

// ValidateOccurrencePairing validates raw bytes against specs/verification/occurrence-pairing.schema.json.
func ValidateOccurrencePairing(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("occurrence pairing must be JSON object: %w", err)
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"case_id",
		"scenario_ref",
		"logical_workflow_id",
		"policy",
		"pairs",
		"first_divergence",
		"unmatched_source_refs",
		"unmatched_target_refs",
		"cleanup_refs",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("occurrence pairing missing required field: %s", f)
		}
	}
	if len(m) != len(expectedFields) {
		return errors.New("occurrence pairing contains unrecognized fields")
	}

	var schemaVer, projID, caseID, scenRef, wfID, policy string
	if err := json.Unmarshal(m["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return errors.New("schema_version must be '1.0'")
	}
	if err := json.Unmarshal(m["project_id"], &projID); err != nil || !projectIDPattern.MatchString(projID) {
		return errors.New("invalid project_id pattern")
	}
	if err := json.Unmarshal(m["case_id"], &caseID); err != nil || !caseIDPattern.MatchString(caseID) {
		return errors.New("invalid case_id pattern")
	}
	if err := json.Unmarshal(m["scenario_ref"], &scenRef); err != nil || !digestPattern.MatchString(scenRef) {
		return errors.New("invalid scenario_ref digest pattern")
	}
	if err := json.Unmarshal(m["logical_workflow_id"], &wfID); err != nil || !workflowIDPattern.MatchString(wfID) {
		return errors.New("invalid logical_workflow_id pattern")
	}
	if err := json.Unmarshal(m["policy"], &policy); err != nil || policy != "STOP_AT_FIRST_DIVERGENCE" {
		return errors.New("policy must be 'STOP_AT_FIRST_DIVERGENCE'")
	}

	var rawPairs []json.RawMessage
	if err := json.Unmarshal(m["pairs"], &rawPairs); err != nil {
		return errors.New("pairs must be array")
	}
	for _, rp := range rawPairs {
		var p map[string]json.RawMessage
		if err := json.Unmarshal(rp, &p); err != nil || len(p) != 5 {
			return errors.New("pair must have exactly 5 fields")
		}
		var sID, sRef, tRef, outcome string
		var visit int
		if err := json.Unmarshal(p["step_id"], &sID); err != nil || !stepIDPattern.MatchString(sID) {
			return errors.New("invalid pair step_id pattern")
		}
		if err := json.Unmarshal(p["visit"], &visit); err != nil || visit < 1 {
			return errors.New("visit must be integer >= 1")
		}
		if err := json.Unmarshal(p["source_occurrence_ref"], &sRef); err != nil || !digestPattern.MatchString(sRef) {
			return errors.New("invalid source_occurrence_ref digest")
		}
		if err := json.Unmarshal(p["target_occurrence_ref"], &tRef); err != nil || !digestPattern.MatchString(tRef) {
			return errors.New("invalid target_occurrence_ref digest")
		}
		if err := json.Unmarshal(p["outcome"], &outcome); err != nil {
			return errors.New("outcome must be string")
		}
		if _, ok := allowedPairingOutcomes[outcome]; !ok {
			return fmt.Errorf("unsupported pair outcome: %q", outcome)
		}
	}

	// first_divergence: null or object
	if string(bytes.TrimSpace(m["first_divergence"])) != "null" {
		var fd map[string]json.RawMessage
		if err := json.Unmarshal(m["first_divergence"], &fd); err != nil || len(fd) != 4 {
			return errors.New("first_divergence must have exactly 4 fields")
		}
		var sRef, tRef, reason string
		if err := json.Unmarshal(fd["source_occurrence_ref"], &sRef); err != nil || !digestPattern.MatchString(sRef) {
			return errors.New("invalid first_divergence source_occurrence_ref digest")
		}
		if err := json.Unmarshal(fd["target_occurrence_ref"], &tRef); err != nil || !digestPattern.MatchString(tRef) {
			return errors.New("invalid first_divergence target_occurrence_ref digest")
		}
		if err := json.Unmarshal(fd["reason"], &reason); err != nil {
			return errors.New("reason must be string")
		}
		if _, ok := allowedDivergenceReasons[reason]; !ok {
			return fmt.Errorf("unsupported divergence reason: %q", reason)
		}
		var evRefs []string
		if err := json.Unmarshal(fd["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
			return errors.New("first_divergence evidence_refs must be non-empty array")
		}
		for _, er := range evRefs {
			if !digestPattern.MatchString(er) {
				return fmt.Errorf("invalid first_divergence evidence_ref: %q", er)
			}
		}
	}

	// String array fields
	for _, arrField := range []string{"unmatched_source_refs", "unmatched_target_refs", "cleanup_refs"} {
		var refs []string
		if err := json.Unmarshal(m[arrField], &refs); err != nil {
			return fmt.Errorf("%s must be array of strings", arrField)
		}
		for _, ref := range refs {
			if !digestPattern.MatchString(ref) {
				return fmt.Errorf("invalid %s item digest: %q", arrField, ref)
			}
		}
	}

	return nil
}

// ParseOccurrencePairing parses and validates raw JSON into an OccurrencePairing.
func ParseOccurrencePairing(raw []byte) (*OccurrencePairing, error) {
	if err := ValidateOccurrencePairing(raw); err != nil {
		return nil, err
	}
	var pairing OccurrencePairing
	if err := json.Unmarshal(raw, &pairing); err != nil {
		return nil, err
	}
	return &pairing, nil
}

// IdentityMapV2 records the relational identity correspondences between source and target entities.
type IdentityMapV2 struct {
	SchemaVersion             string            `json:"schema_version"`
	ProjectID                 string            `json:"project_id"`
	CaseID                    string            `json:"case_id"`
	WorkflowInstanceID        string            `json:"workflow_instance_id"`
	PolicyRef                 string            `json:"policy_ref"`
	Bindings                  []IdentityBinding `json:"bindings"`
	RelationshipAssertionRefs []string          `json:"relationship_assertion_refs"`
	SourceLocalIdentityRef    *string           `json:"source_local_identity_ref,omitempty"`
	TargetLocalIdentityRef    *string           `json:"target_local_identity_ref,omitempty"`
	OccurrencePairingRef      *string           `json:"occurrence_pairing_ref,omitempty"`
}

// IdentityBinding binds one source entity value to a target entity value under relational mapping.
type IdentityBinding struct {
	BindingID     string              `json:"binding_id"`
	RelationGroup string              `json:"relation_group"`
	ActorID       string              `json:"actor_id"`
	ConnectionID  *string             `json:"connection_id"`
	LogicalKey    string              `json:"logical_key"`
	SourceValue   TypedValueContainer `json:"source_value"`
	TargetValue   TypedValueContainer `json:"target_value"`
	EvidenceRefs  []string            `json:"evidence_refs"`
	SourceLocalID string              `json:"source_local_id,omitempty"`
	TargetLocalID string              `json:"target_local_id,omitempty"`
}

// TypedValueContainer carries a value and its declared type (string or integer).
type TypedValueContainer struct {
	ValueType string `json:"value_type"`
	Value     string `json:"value"`
}

// ValidateIdentityMap validates raw bytes against specs/verification/identity-map.schema.json (v2.0 or v2.1).
func ValidateIdentityMap(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("identity map must be JSON object: %w", err)
	}

	verRaw, ok := m["schema_version"]
	if !ok {
		return errors.New("missing schema_version")
	}
	var ver string
	if err := json.Unmarshal(verRaw, &ver); err != nil {
		return errors.New("schema_version must be string")
	}

	if ver == "2.0" {
		return validateIdentityMapV20(m)
	}
	if ver == "2.1" {
		return validateIdentityMapV21(m)
	}
	return fmt.Errorf("unsupported identity map schema_version: %q", ver)
}

func validateIdentityMapV20(m map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"project_id",
		"case_id",
		"workflow_instance_id",
		"policy_ref",
		"bindings",
		"relationship_assertion_refs",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("identity map 2.0 missing required field: %s", f)
		}
	}
	if len(m) != len(expectedFields) {
		return errors.New("identity map 2.0 contains unrecognized fields")
	}

	if err := validateIdentityMapShared(m); err != nil {
		return err
	}

	return validateIdentityBindings(m["bindings"], false)
}

func validateIdentityMapV21(m map[string]json.RawMessage) error {
	expectedFields := []string{
		"schema_version",
		"project_id",
		"case_id",
		"workflow_instance_id",
		"policy_ref",
		"bindings",
		"relationship_assertion_refs",
		"source_local_identity_ref",
		"target_local_identity_ref",
		"occurrence_pairing_ref",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("identity map 2.1 missing required field: %s", f)
		}
	}
	if len(m) != len(expectedFields) {
		return errors.New("identity map 2.1 contains unrecognized fields")
	}

	if err := validateIdentityMapShared(m); err != nil {
		return err
	}

	for _, refName := range []string{"source_local_identity_ref", "target_local_identity_ref", "occurrence_pairing_ref"} {
		var ref string
		if err := json.Unmarshal(m[refName], &ref); err != nil || !digestPattern.MatchString(ref) {
			return fmt.Errorf("invalid %s digest pattern", refName)
		}
	}

	return validateIdentityBindings(m["bindings"], true)
}

func validateIdentityMapShared(m map[string]json.RawMessage) error {
	var projID, caseID, wfID, polRef string
	if err := json.Unmarshal(m["project_id"], &projID); err != nil || !projectIDPattern.MatchString(projID) {
		return errors.New("invalid project_id pattern")
	}
	if err := json.Unmarshal(m["case_id"], &caseID); err != nil || !caseIDPattern.MatchString(caseID) {
		return errors.New("invalid case_id pattern")
	}
	if err := json.Unmarshal(m["workflow_instance_id"], &wfID); err != nil || !workflowIDPattern.MatchString(wfID) {
		return errors.New("invalid workflow_instance_id pattern")
	}
	if err := json.Unmarshal(m["policy_ref"], &polRef); err != nil || !digestPattern.MatchString(polRef) {
		return errors.New("invalid policy_ref digest pattern")
	}

	var relAsserts []string
	if err := json.Unmarshal(m["relationship_assertion_refs"], &relAsserts); err != nil {
		return errors.New("relationship_assertion_refs must be array of strings")
	}
	seenAsserts := make(map[string]struct{}, len(relAsserts))
	for _, ra := range relAsserts {
		if !digestPattern.MatchString(ra) {
			return fmt.Errorf("invalid relationship_assertion_ref digest: %q", ra)
		}
		if _, exists := seenAsserts[ra]; exists {
			return fmt.Errorf("duplicate relationship_assertion_ref: %q", ra)
		}
		seenAsserts[ra] = struct{}{}
	}

	return nil
}

func validateIdentityBindings(rawBindings json.RawMessage, isV21 bool) error {
	var rawList []json.RawMessage
	if err := json.Unmarshal(rawBindings, &rawList); err != nil {
		return errors.New("bindings must be array")
	}

	expectedLen := 8
	if isV21 {
		expectedLen = 10
	}

	seenBindingIDs := make(map[string]struct{}, len(rawList))
	for _, rb := range rawList {
		var b map[string]json.RawMessage
		if err := json.Unmarshal(rb, &b); err != nil || len(b) != expectedLen {
			return fmt.Errorf("binding must have exactly %d fields", expectedLen)
		}

		requiredFields := []string{
			"binding_id",
			"relation_group",
			"actor_id",
			"connection_id",
			"logical_key",
			"source_value",
			"target_value",
			"evidence_refs",
		}
		if isV21 {
			requiredFields = append(requiredFields, "source_local_id", "target_local_id")
		}
		for _, rf := range requiredFields {
			if _, ok := b[rf]; !ok {
				return fmt.Errorf("binding missing required field: %s", rf)
			}
		}

		var bID, rg, aID, lk string
		if err := json.Unmarshal(b["binding_id"], &bID); err != nil || !bindingIDPattern.MatchString(bID) {
			return errors.New("invalid binding_id pattern")
		}
		if _, exists := seenBindingIDs[bID]; exists {
			return fmt.Errorf("duplicate binding_id: %q", bID)
		}
		seenBindingIDs[bID] = struct{}{}

		if err := json.Unmarshal(b["relation_group"], &rg); err != nil || !relationGroupPattern.MatchString(rg) {
			return errors.New("invalid relation_group pattern")
		}
		if err := json.Unmarshal(b["actor_id"], &aID); err != nil || !actorIDPattern.MatchString(aID) {
			return errors.New("invalid actor_id pattern")
		}
		if string(bytes.TrimSpace(b["connection_id"])) != "null" {
			var connID string
			if err := json.Unmarshal(b["connection_id"], &connID); err != nil || !connectionPattern.MatchString(connID) {
				return errors.New("invalid connection_id pattern")
			}
		}
		if err := json.Unmarshal(b["logical_key"], &lk); err != nil || !logicalKeyPattern.MatchString(lk) {
			return errors.New("invalid logical_key pattern")
		}

		for _, valField := range []string{"source_value", "target_value"} {
			var tc TypedValueContainer
			if err := json.Unmarshal(b[valField], &tc); err != nil {
				return fmt.Errorf("invalid %s container: %w", valField, err)
			}
			if tc.ValueType != "string" && tc.ValueType != "integer" {
				return fmt.Errorf("invalid %s value_type: %q", valField, tc.ValueType)
			}
			if tc.ValueType == "integer" && !canonicalIntDecPattern.MatchString(tc.Value) {
				return fmt.Errorf("invalid integer pattern for %s: %q", valField, tc.Value)
			}
		}

		var evRefs []string
		if err := json.Unmarshal(b["evidence_refs"], &evRefs); err != nil {
			return errors.New("evidence_refs must be array")
		}
		seenEv := make(map[string]struct{}, len(evRefs))
		for _, er := range evRefs {
			if !digestPattern.MatchString(er) {
				return fmt.Errorf("invalid evidence_ref digest: %q", er)
			}
			if _, exists := seenEv[er]; exists {
				return fmt.Errorf("duplicate evidence_ref in binding: %q", er)
			}
			seenEv[er] = struct{}{}
		}

		if isV21 {
			var sLocID, tLocID string
			if err := json.Unmarshal(b["source_local_id"], &sLocID); err != nil || !localIDPattern.MatchString(sLocID) {
				return errors.New("invalid source_local_id pattern")
			}
			if err := json.Unmarshal(b["target_local_id"], &tLocID); err != nil || !localIDPattern.MatchString(tLocID) {
				return errors.New("invalid target_local_id pattern")
			}
		}
	}

	return nil
}

// ParseIdentityMap parses and validates raw JSON into an IdentityMapV2.
func ParseIdentityMap(raw []byte) (*IdentityMapV2, error) {
	if err := ValidateIdentityMap(raw); err != nil {
		return nil, err
	}
	var idMap IdentityMapV2
	if err := json.Unmarshal(raw, &idMap); err != nil {
		return nil, err
	}
	return &idMap, nil
}

// VerificationReportV2 represents an immutable verification report conforming to specs/verification/report.schema.json v2.1.
type VerificationReportV2 struct {
	SchemaVersion      string                    `json:"schema_version"`
	RunID              string                    `json:"run_id"`
	ProjectID          string                    `json:"project_id"`
	ConfigurationHash  string                    `json:"configuration_hash"`
	PlanHash           string                    `json:"plan_hash"`
	CorpusHash         string                    `json:"corpus_hash"`
	PolicyHash         string                    `json:"policy_hash"`
	TargetTreeHash     string                    `json:"target_tree_hash"`
	Status             string                    `json:"status"`
	Verdict            string                    `json:"verdict"`
	HasExpectedChanges bool                      `json:"has_expected_changes"`
	Cases              []VerificationCaseResult  `json:"cases"`
	Errors             []VerificationErrorDetail `json:"errors"`
	AcceptanceScope    AcceptanceScope           `json:"acceptance_scope"`
}

// VerificationCaseResult contains the evaluation results of a single verified case.
type VerificationCaseResult struct {
	CaseID               string                 `json:"case_id"`
	SourceTraceRef       *string                `json:"source_trace_ref"`
	TargetTraceRef       *string                `json:"target_trace_ref"`
	Verdict              string                 `json:"verdict"`
	Dimensions           VerificationDimensions `json:"dimensions"`
	MismatchIDs          []string               `json:"mismatch_ids"`
	IdentityMapRef       *string                `json:"identity_map_ref"`
	OccurrencePairingRef *string                `json:"occurrence_pairing_ref"`
}

// VerificationDimensions records the outcome across all 11 evaluation dimensions.
type VerificationDimensions struct {
	HTTP           DimensionResult `json:"http"`
	Body           DimensionResult `json:"body"`
	Authentication DimensionResult `json:"authentication"`
	Authorization  DimensionResult `json:"authorization"`
	Session        DimensionResult `json:"session"`
	Database       DimensionResult `json:"database"`
	External       DimensionResult `json:"external"`
	Files          DimensionResult `json:"files"`
	Cache          DimensionResult `json:"cache"`
	Errors         DimensionResult `json:"errors"`
	Execution      DimensionResult `json:"execution"`
}

// DimensionResult records the result for one verification dimension.
type DimensionResult struct {
	Applicable   bool     `json:"applicable"`
	Verdict      *string  `json:"verdict"`
	Reason       *string  `json:"reason"`
	EvidenceRefs []string `json:"evidence_refs"`
}

// ValidateVerificationReport validates raw bytes against specs/verification/report.schema.json (v2.0 or v2.1).
func ValidateVerificationReport(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("verification report must be JSON object: %w", err)
	}

	verRaw, ok := m["schema_version"]
	if !ok {
		return errors.New("missing schema_version")
	}
	var ver string
	if err := json.Unmarshal(verRaw, &ver); err != nil {
		return errors.New("schema_version must be string")
	}

	if ver == "2.1" {
		return validateVerificationReportV21(m)
	}
	if ver == "2.0" {
		return validateVerificationReportV20(m)
	}
	return fmt.Errorf("unsupported verification report schema_version: %q", ver)
}

func validateVerificationReportV21(m map[string]json.RawMessage) error {
	return validateReportShared(m, true)
}

func validateVerificationReportV20(m map[string]json.RawMessage) error {
	return validateReportShared(m, false)
}

func validateReportShared(m map[string]json.RawMessage, isV21 bool) error {
	expectedFields := []string{
		"schema_version",
		"run_id",
		"project_id",
		"configuration_hash",
		"plan_hash",
		"corpus_hash",
		"policy_hash",
		"target_tree_hash",
		"status",
		"verdict",
		"has_expected_changes",
		"cases",
		"errors",
		"acceptance_scope",
	}
	for _, f := range expectedFields {
		if _, ok := m[f]; !ok {
			return fmt.Errorf("report missing required field: %s", f)
		}
	}
	if len(m) != len(expectedFields) {
		return errors.New("report contains unrecognized fields")
	}

	var runID, projID, status, topVerdict string
	if err := json.Unmarshal(m["run_id"], &runID); err != nil || !runIDPattern.MatchString(runID) {
		return errors.New("invalid run_id pattern")
	}
	if err := json.Unmarshal(m["project_id"], &projID); err != nil || !projectIDPattern.MatchString(projID) {
		return errors.New("invalid project_id pattern")
	}

	for _, hf := range []string{"configuration_hash", "plan_hash", "corpus_hash", "policy_hash", "target_tree_hash"} {
		var h string
		if err := json.Unmarshal(m[hf], &h); err != nil || !sha256HexPattern.MatchString(h) {
			return fmt.Errorf("invalid %s hex pattern", hf)
		}
	}

	if err := json.Unmarshal(m["status"], &status); err != nil {
		return errors.New("status must be string")
	}
	if _, ok := allowedReportStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %q", status)
	}

	if err := json.Unmarshal(m["verdict"], &topVerdict); err != nil {
		return errors.New("verdict must be string")
	}
	if _, ok := allowedReportVerdicts[topVerdict]; !ok {
		return fmt.Errorf("invalid verdict: %q", topVerdict)
	}

	var hasExp bool
	if err := json.Unmarshal(m["has_expected_changes"], &hasExp); err != nil {
		return errors.New("has_expected_changes must be boolean")
	}

	// Acceptance scope
	if err := ValidateAcceptanceScope(m["acceptance_scope"]); err != nil {
		return fmt.Errorf("invalid acceptance_scope in report: %w", err)
	}

	// Cases
	var rawCases []json.RawMessage
	if err := json.Unmarshal(m["cases"], &rawCases); err != nil {
		return errors.New("cases must be array")
	}
	if status == "COMPLETED" && len(rawCases) < 1 {
		return errors.New("cases must have minItems: 1 when status is COMPLETED")
	}

	for _, rc := range rawCases {
		var c map[string]json.RawMessage
		if err := json.Unmarshal(rc, &c); err != nil {
			return errors.New("case must be object")
		}

		expectedCaseFields := []string{"case_id", "source_trace_ref", "target_trace_ref", "verdict", "dimensions", "mismatch_ids"}
		if isV21 {
			expectedCaseFields = append(expectedCaseFields, "identity_map_ref", "occurrence_pairing_ref")
		}
		for _, cf := range expectedCaseFields {
			if _, ok := c[cf]; !ok {
				return fmt.Errorf("case missing required field: %s", cf)
			}
		}
		if len(c) != len(expectedCaseFields) {
			return errors.New("case contains unrecognized fields")
		}

		var cID, cVerdict string
		if err := json.Unmarshal(c["case_id"], &cID); err != nil || !caseIDPattern.MatchString(cID) {
			return errors.New("invalid case_id pattern")
		}
		if err := json.Unmarshal(c["verdict"], &cVerdict); err != nil {
			return errors.New("case verdict must be string")
		}
		if _, ok := allowedReportVerdicts[cVerdict]; !ok {
			return fmt.Errorf("invalid case verdict: %q", cVerdict)
		}

		// Source and target trace refs
		isSrcNull := string(bytes.TrimSpace(c["source_trace_ref"])) == "null"
		isTgtNull := string(bytes.TrimSpace(c["target_trace_ref"])) == "null"

		if !isSrcNull {
			var st string
			if err := json.Unmarshal(c["source_trace_ref"], &st); err != nil || !digestPattern.MatchString(st) {
				return errors.New("invalid source_trace_ref digest")
			}
		}
		if !isTgtNull {
			var tt string
			if err := json.Unmarshal(c["target_trace_ref"], &tt); err != nil || !digestPattern.MatchString(tt) {
				return errors.New("invalid target_trace_ref digest")
			}
		}

		// allOf constraint: PASS or MISMATCH requires non-null source_trace_ref and target_trace_ref
		if cVerdict == "PASS" || cVerdict == "MISMATCH" {
			if isSrcNull || isTgtNull {
				return fmt.Errorf("case with verdict %s requires non-null source_trace_ref and target_trace_ref", cVerdict)
			}
		}

		if isV21 {
			isIdMapNull := string(bytes.TrimSpace(c["identity_map_ref"])) == "null"
			isPairNull := string(bytes.TrimSpace(c["occurrence_pairing_ref"])) == "null"

			if !isIdMapNull {
				var idm string
				if err := json.Unmarshal(c["identity_map_ref"], &idm); err != nil || !digestPattern.MatchString(idm) {
					return errors.New("invalid identity_map_ref digest")
				}
			}
			if !isPairNull {
				var op string
				if err := json.Unmarshal(c["occurrence_pairing_ref"], &op); err != nil || !digestPattern.MatchString(op) {
					return errors.New("invalid occurrence_pairing_ref digest")
				}
			}

			// allOf constraint: PASS requires non-null identity_map_ref and occurrence_pairing_ref
			if cVerdict == "PASS" {
				if isIdMapNull || isPairNull {
					return errors.New("case with verdict PASS requires non-null identity_map_ref and occurrence_pairing_ref")
				}
			}
		}

		// Dimensions
		var rawDims map[string]json.RawMessage
		if err := json.Unmarshal(c["dimensions"], &rawDims); err != nil {
			return errors.New("dimensions must be object")
		}
		for _, dName := range reportDimensionNames {
			rd, ok := rawDims[dName]
			if !ok {
				return fmt.Errorf("dimensions missing required dimension: %s", dName)
			}
			var d map[string]json.RawMessage
			if err := json.Unmarshal(rd, &d); err != nil || len(d) != 4 {
				return fmt.Errorf("dimension %s must have exactly 4 fields", dName)
			}
			for _, df := range []string{"applicable", "verdict", "reason", "evidence_refs"} {
				if _, ok := d[df]; !ok {
					return fmt.Errorf("dimension %s missing required field: %s", dName, df)
				}
			}

			var applicable bool
			if err := json.Unmarshal(d["applicable"], &applicable); err != nil {
				return fmt.Errorf("dimension %s applicable must be boolean", dName)
			}

			isDimVerdictNull := string(bytes.TrimSpace(d["verdict"])) == "null"
			var dVerdict string
			if !isDimVerdictNull {
				if err := json.Unmarshal(d["verdict"], &dVerdict); err != nil {
					return fmt.Errorf("dimension %s verdict must be string or null", dName)
				}
				if _, ok := allowedReportVerdicts[dVerdict]; !ok {
					return fmt.Errorf("dimension %s invalid verdict: %q", dName, dVerdict)
				}
			}

			if !applicable && !isDimVerdictNull {
				return fmt.Errorf("non-applicable dimension %s must have null verdict", dName)
			}
			if applicable && isDimVerdictNull {
				return fmt.Errorf("applicable dimension %s must have non-null verdict", dName)
			}

			// In case verdict PASS, any applicable dimension MUST have verdict PASS
			if cVerdict == "PASS" && applicable && dVerdict != "PASS" {
				return fmt.Errorf("case with verdict PASS cannot have non-PASS dimension %s (got %s)", dName, dVerdict)
			}

			var dEvRefs []string
			if err := json.Unmarshal(d["evidence_refs"], &dEvRefs); err != nil {
				return fmt.Errorf("dimension %s evidence_refs must be array", dName)
			}
			for _, der := range dEvRefs {
				if !digestPattern.MatchString(der) {
					return fmt.Errorf("dimension %s invalid evidence_ref: %q", dName, der)
				}
			}
		}

		// Mismatch IDs
		var mIDs []string
		if err := json.Unmarshal(c["mismatch_ids"], &mIDs); err != nil {
			return errors.New("mismatch_ids must be array")
		}
		for _, mid := range mIDs {
			if !mismatchIDPattern.MatchString(mid) {
				return fmt.Errorf("invalid mismatch_id pattern: %q", mid)
			}
		}
	}

	// allOf constraint: top-level report verdict PASS requires minItems: 1 and all cases PASS
	if topVerdict == "PASS" {
		if len(rawCases) < 1 {
			return errors.New("report with verdict PASS must have at least 1 case")
		}
		for _, rc := range rawCases {
			var c map[string]json.RawMessage
			_ = json.Unmarshal(rc, &c)
			var cv string
			_ = json.Unmarshal(c["verdict"], &cv)
			if cv != "PASS" {
				return fmt.Errorf("report with verdict PASS cannot contain non-PASS case (got %s)", cv)
			}
		}
	}

	return nil
}

// ParseVerificationReport parses and validates raw JSON into a VerificationReportV2.
func ParseVerificationReport(raw []byte) (*VerificationReportV2, error) {
	if err := ValidateVerificationReport(raw); err != nil {
		return nil, err
	}
	var rep VerificationReportV2
	if err := json.Unmarshal(raw, &rep); err != nil {
		return nil, err
	}
	return &rep, nil
}

// CanonicalJSON serializes the verification report to canonical rwt-cjson-1 bytes.
func (r *VerificationReportV2) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical report.
func (r *VerificationReportV2) Digest() (string, error) {
	cjson, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

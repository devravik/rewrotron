package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SessionContinuityContract is the bounded executable projection of an approved
// behavior contract. It defines obligations, never a session decoder or observation.
// ContractRef and DecisionRef must belong to the current plan execution closure.
// Checks retain side-local expectations; relations compare actual captured values.
type SessionContinuityContract struct {
	SchemaVersion          string                     `json:"schema_version"`
	ProjectID              string                     `json:"project_id"`
	CaseID                 string                     `json:"case_id"`
	WorkflowID             string                     `json:"logical_workflow_id"`
	ContractRef            string                     `json:"-"`
	DecisionRef            string                     `json:"decision_ref"`
	Strategy               string                     `json:"strategy"` // SAFE_BRIDGE or REVIEWED_RELOGIN
	CodecBinding           *LoweringBinding           `json:"codec_binding"`
	CodecImplementationRef string                     `json:"codec_implementation_ref"`
	CodecProfileRef        string                     `json:"codec_profile_ref"`
	Checks                 []SessionLifecycleCheck    `json:"checks"`
	Relations              []SessionLifecycleRelation `json:"relations"`
}

// SessionObservation selects an exact phase of a logical workflow occurrence.
// ActorRef selects the intended fixture, not the actual authenticated identity.
type SessionObservation struct {
	Side     string `json:"side"`
	StepID   string `json:"step_id"`
	Visit    int    `json:"visit"`
	Phase    string `json:"phase"`
	ActorRef string `json:"actor_ref"`
}

// SessionLifecycleCheck evaluates an existing bounded boolean expression against
// actual evidence. IDs identify the approved obligation (expiry, replay, flash,
// reset, remember-me, throttle, rollback, etc.), not inferred framework behavior.
type SessionLifecycleCheck struct {
	ID          string               `json:"id"`
	Observation SessionObservation   `json:"observation"`
	Predicate   *SemanticsExpression `json:"predicate"`
}

// MarshalJSON preserves required empty pointers and null literals in the existing
// expression wire contract, including clock references and anonymous principals.
func (c SessionLifecycleCheck) MarshalJSON() ([]byte, error) {
	raw, err := sessionExpressionJSON(c.Predicate)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		ID          string             `json:"id"`
		Observation SessionObservation `json:"observation"`
		Predicate   json.RawMessage    `json:"predicate"`
	}{c.ID, c.Observation, raw})
}

func sessionExpressionJSON(expr *SemanticsExpression) ([]byte, error) {
	raw, err := json.Marshal(expr)
	if err != nil {
		return nil, err
	}
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, err
	}
	var repair func(any)
	repair = func(node any) {
		switch value := node.(type) {
		case map[string]any:
			if value["op"] == "ref" {
				if _, ok := value["pointer"]; !ok {
					value["pointer"] = ""
				}
			}
			if value["op"] == "literal" {
				if _, ok := value["value"]; !ok {
					value["value"] = nil
				}
			}
			for _, child := range value {
				repair(child)
			}
		case []any:
			for _, child := range value {
				repair(child)
			}
		}
	}
	repair(tree)
	return json.Marshal(tree)
}

// SessionLifecycleValue selects a typed value without changing the captured context.
type SessionLifecycleValue struct {
	Observation SessionObservation `json:"observation"`
	Path        string             `json:"path"`
}

// SessionLifecycleRelation preserves an equality or inequality across requests or
// sides. Only an exact approved scalar identity rule may map differing identities;
// rotation and token regeneration use same-side inequality with no normalization.
type SessionLifecycleRelation struct {
	ID             string                `json:"id"`
	Left           SessionLifecycleValue `json:"left"`
	Right          SessionLifecycleValue `json:"right"`
	Different      bool                  `json:"different"`
	IdentityRuleID string                `json:"identity_rule_id"`
	ActorID        string                `json:"actor_id"`
	ConnectionID   *string               `json:"connection_id"`
}

// SessionContinuityResult reports bounded semantic acceptance and readiness.
// UNKNOWN or MISMATCH always blocks readiness; re-login never asserts continuity.
type SessionContinuityResult struct {
	Verdict      string
	ReasonCodes  []string
	EvidenceRefs []string
}

// Digest pins this compiled contract's exact obligations and provider identities.
// ContractRef is the resulting content address, omitted from its own preimage.
// This is a plan-consumed contract artifact, not a new approval subject or receipt.
func (c *SessionContinuityContract) Digest() (string, error) {
	if c == nil {
		return "", &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "nil session continuity contract"}
	}
	copyContract := *c
	copyContract.ContractRef = ""
	raw, err := json.Marshal(copyContract)
	if err != nil {
		return "", err
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return "", err
	}
	hash, err := CanonicalHash(canon)
	if err != nil {
		return "", err
	}
	return "sha256:" + hash, nil
}

// ValidateSessionContinuityContract validates the closed 1.0 contract wire shape
// and bounded expression syntax. Qualification and actual evidence are separate
// service checks; structural validity never establishes session compatibility.
func ValidateSessionContinuityContract(raw []byte) error {
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "invalid session contract JSON", Cause: err}
	}
	fields, err := sessionRecordFields(canon, []string{"schema_version", "project_id", "case_id", "logical_workflow_id", "decision_ref", "strategy", "codec_binding", "codec_implementation_ref", "codec_profile_ref", "checks", "relations"})
	if err != nil {
		return err
	}
	var c SessionContinuityContract
	if err := json.Unmarshal(canon, &c); err != nil {
		return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "session contract field type", Cause: err}
	}
	for _, name := range []string{"codec_implementation_ref", "codec_profile_ref"} {
		var value *string
		if err := json.Unmarshal(fields[name], &value); err != nil || value == nil {
			return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "codec pin must be string"}
		}
	}
	if c.SchemaVersion != "1.0" || !projectIDPattern.MatchString(c.ProjectID) || !caseIDPattern.MatchString(c.CaseID) || !workflowIDPattern.MatchString(c.WorkflowID) || !digestPattern.MatchString(c.DecisionRef) ||
		(c.Strategy != "SAFE_BRIDGE" && c.Strategy != "REVIEWED_RELOGIN") || len(c.Checks) == 0 || len(c.Relations) == 0 || len(c.Checks)+len(c.Relations) > MaxExpressionNodes {
		return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "session contract identity/strategy/bounds"}
	}
	if c.Strategy == "SAFE_BRIDGE" && (c.CodecBinding == nil || !digestPattern.MatchString(c.CodecImplementationRef) || !digestPattern.MatchString(c.CodecProfileRef)) {
		return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "safe bridge requires pinned codec"}
	}
	for _, ref := range []string{c.CodecImplementationRef, c.CodecProfileRef} {
		if ref != "" && !digestPattern.MatchString(ref) {
			return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "codec pin"}
		}
	}
	if c.CodecBinding != nil {
		if _, err := sessionRecordFields(fields["codec_binding"], []string{"capability_id", "provider_manifest_ref", "qualification_ref", "obligation_refs", "input_shape_ref", "output_shape_ref"}); err != nil {
			return err
		}
		b := c.CodecBinding
		if b.CapabilityID == "" || len(b.ObligationRefs) == 0 {
			return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "codec capability/obligations"}
		}
		for _, ref := range append([]string{b.ProviderManifestRef, b.QualificationRef, b.InputShapeRef, b.OutputShapeRef}, b.ObligationRefs...) {
			if !digestPattern.MatchString(ref) {
				return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "codec binding ref"}
			}
		}
	}
	var checks, relations []json.RawMessage
	if err := json.Unmarshal(fields["checks"], &checks); err != nil {
		return err
	}
	if err := json.Unmarshal(fields["relations"], &relations); err != nil {
		return err
	}
	seen := make(map[string]bool)
	for i, rawCheck := range checks {
		check, err := sessionRecordFields(rawCheck, []string{"id", "observation", "predicate"})
		if err != nil {
			return err
		}
		if err := sessionObservationShape(check["observation"], c.Checks[i].Observation); err != nil {
			return err
		}
		if err := ValidateExpression(check["predicate"]); err != nil {
			return err
		}
		id := c.Checks[i].ID
		if !ruleIDPattern.MatchString(id) || seen[id] {
			return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "invalid/duplicate obligation ID"}
		}
		seen[id] = true
	}
	for i, rawRelation := range relations {
		fields, err := sessionRecordFields(rawRelation, []string{"id", "left", "right", "different", "identity_rule_id", "actor_id", "connection_id"})
		if err != nil {
			return err
		}
		relation := c.Relations[i]
		var different *bool
		if err := json.Unmarshal(fields["different"], &different); err != nil || different == nil {
			return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "relation different must be boolean"}
		}
		for _, name := range []string{"identity_rule_id", "actor_id"} {
			var value *string
			if err := json.Unmarshal(fields[name], &value); err != nil || value == nil {
				return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "relation selector must be string"}
			}
		}
		for j, key := range []string{"left", "right"} {
			v := relation.Left
			if j == 1 {
				v = relation.Right
			}
			fields, err := sessionRecordFields(fields[key], []string{"observation", "path"})
			if err != nil {
				return err
			}
			if err := sessionObservationShape(fields["observation"], v.Observation); err != nil {
				return err
			}
			if !strings.HasPrefix(v.Path, "/") {
				return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "relation path"}
			}
		}
		if !ruleIDPattern.MatchString(relation.ID) || seen[relation.ID] || (relation.ConnectionID != nil && !connectionIDPattern.MatchString(*relation.ConnectionID)) {
			return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "invalid/duplicate relation"}
		}
		seen[relation.ID] = true
	}
	return nil
}

func sessionRecordFields(raw []byte, names []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "session record must be object", Cause: err}
	}
	if len(fields) != len(names) {
		return nil, &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "session record fields"}
	}
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return nil, &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "missing session record field"}
		}
	}
	return fields, nil
}

func sessionObservationShape(raw []byte, o SessionObservation) error {
	if _, err := sessionRecordFields(raw, []string{"side", "step_id", "visit", "phase", "actor_ref"}); err != nil {
		return err
	}
	if (o.Side != SideSource && o.Side != SideTarget) || !stepIDPattern.MatchString(o.StepID) || o.Visit < 1 ||
		(o.Phase != PhasePreOperation && o.Phase != PhasePostOperation) || !digestPattern.MatchString(o.ActorRef) {
		return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "session observation selector"}
	}
	return nil
}

func (r *SessionContinuityResult) record(verdict, reason string) {
	if verdict == "MISMATCH" || (verdict == "UNKNOWN" && r.Verdict != "MISMATCH") {
		r.Verdict = verdict
	}
	if reason != "" {
		r.ReasonCodes = append(r.ReasonCodes, reason)
	}
}

func evaluateSessionContinuity(input *CompareCaseInput, idMap *IdentityMapV2) SessionContinuityResult {
	r := SessionContinuityResult{Verdict: "PASS"}
	contract := input.SessionContinuity
	if contract == nil {
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: missing contract")
		return r
	}
	if contract.SchemaVersion != "1.0" || contract.ProjectID != input.ProjectID || contract.CaseID != input.CaseID ||
		contract.WorkflowID != input.WorkflowID || len(contract.Checks)+len(contract.Relations) > MaxExpressionNodes {
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: unsupported contract scope/bounds")
		return r
	}
	contractRaw, err := json.Marshal(contract)
	if err != nil || ValidateSessionContinuityContract(contractRaw) != nil {
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: invalid bounded session contract")
		return r
	}
	if contract.ContractRef != input.RequiredSessionContractRef || !digestPattern.MatchString(contract.ContractRef) ||
		!digestPattern.MatchString(contract.DecisionRef) || input.SessionPlanSubject == nil ||
		input.SessionPlanSubject.ProjectID != input.ProjectID || input.SessionPlanSubject.ConfigurationHash != input.ConfigurationHash ||
		input.SessionPlanSubject.CorpusRef != "sha256:"+input.CorpusHash || input.SessionPlanSubject.PolicyRef != input.PolicyRef ||
		!sessionPlanContains(input.SessionPlanSubject, contract) {
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: contract/decision outside plan closure")
		return r
	}
	r.EvidenceRefs = append(r.EvidenceRefs, contract.ContractRef)
	digest, contractErr := contract.Digest()
	planHash, planErr := input.SessionPlanSubject.Digest()
	if contractErr != nil || digest != contract.ContractRef || planErr != nil || planHash != input.PlanHash {
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: changed contract/plan bytes")
		return r
	}
	if len(contract.Checks) == 0 || len(contract.Relations) == 0 {
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: empty lifecycle obligations")
	}
	switch contract.Strategy {
	case "SAFE_BRIDGE":
		if err := sessionCodecQualification(input, contract); err != nil {
			r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: "+err.Error())
		}
	case "REVIEWED_RELOGIN":
		// The exact reviewed exception is in DecisionRef. Its observed re-login
		// obligations still run, and the report explicitly does not claim continuity.
		r.ReasonCodes = append(r.ReasonCodes, "REVIEWED_RELOGIN: active continuity excluded by plan decision")
	default:
		r.record("UNKNOWN", "ACTIVE_SESSION_CONTINUITY_UNPROVEN: unsupported strategy")
	}

	seen := make(map[string]bool)
	for _, check := range contract.Checks {
		if check.ID == "" || seen[check.ID] {
			r.record("UNKNOWN", "invalid/duplicate lifecycle obligation")
			continue
		}
		seen[check.ID] = true
		ctx, err := sessionContext(input, check.Observation)
		if err != nil {
			r.record("UNKNOWN", check.ID+": "+err.Error())
			continue
		}
		r.EvidenceRefs = append(r.EvidenceRefs, ctx.EvidenceRefs...)
		raw, err := sessionExpressionJSON(check.Predicate)
		if err != nil || ValidateExpression(raw) != nil {
			r.record("UNKNOWN", check.ID+": invalid lifecycle predicate")
			continue
		}
		// Even is_missing/exists cannot turn unavailable evidence into a successful
		// security assertion. Every referenced operand must actually be captured.
		if !sessionExpressionCaptured(check.Predicate, ctx) {
			r.record("UNKNOWN", check.ID+": missing lifecycle observation")
			continue
		}
		value, err := EvaluateExpression(check.Predicate, ctx)
		if err != nil || value.Type != ValueTypeBoolean {
			r.record("UNKNOWN", check.ID+": unsupported lifecycle predicate")
		} else if value.Value != true {
			r.record("MISMATCH", check.ID+": lifecycle obligation violated")
		}
	}
	for _, relation := range contract.Relations {
		if relation.ID == "" || seen[relation.ID] {
			r.record("UNKNOWN", "invalid/duplicate lifecycle relation")
			continue
		}
		seen[relation.ID] = true
		leftCtx, leftErr := sessionContext(input, relation.Left.Observation)
		rightCtx, rightErr := sessionContext(input, relation.Right.Observation)
		if leftErr != nil || rightErr != nil {
			r.record("UNKNOWN", relation.ID+": missing phase-qualified lifecycle context")
			continue
		}
		r.EvidenceRefs = append(r.EvidenceRefs, leftCtx.EvidenceRefs...)
		r.EvidenceRefs = append(r.EvidenceRefs, rightCtx.EvidenceRefs...)
		left, le := leftCtx.ResolvePointer(relation.Left.Path)
		right, re := rightCtx.ResolvePointer(relation.Right.Path)
		if le != nil || re != nil || ContainsMissingCapability(left) || ContainsMissingCapability(right) {
			r.record("UNKNOWN", relation.ID+": unavailable lifecycle value")
			continue
		}
		if relation.IdentityRuleID != "" {
			verdict := sessionIdentityRelation(input, contract, relation, left, right, idMap)
			if verdict != "PASS" {
				r.record(verdict, relation.ID+": identity relation not established")
			}
			continue
		}
		equal, verdict, err := CompareTaggedValues(left, right)
		if err != nil || verdict == "UNKNOWN" {
			r.record("UNKNOWN", relation.ID+": unsupported typed relation")
		} else if equal == relation.Different {
			r.record("MISMATCH", relation.ID+": lifecycle relation violated")
		}
	}
	return r
}

func sessionPlanContains(plan *PlanApprovalSubject, contract *SessionContinuityContract) bool {
	decision := false
	for _, ref := range plan.DecisionRefs {
		decision = decision || ref == contract.DecisionRef
	}
	for _, unit := range plan.OrderedUnits {
		for _, ref := range unit.ContractRefs {
			if decision && ref == contract.ContractRef {
				for _, caseID := range unit.CaseIDs {
					if caseID == contract.CaseID {
						return true
					}
				}
			}
		}
	}
	return false
}

func sessionCodecQualification(input *CompareCaseInput, contract *SessionContinuityContract) error {
	binding := contract.CodecBinding
	if binding == nil || !digestPattern.MatchString(binding.ProviderManifestRef) ||
		!digestPattern.MatchString(contract.CodecImplementationRef) {
		return fmt.Errorf("missing pinned safe codec provider")
	}
	q := input.AdmittedSessionQualifications[binding.QualificationRef]
	if q == nil {
		return fmt.Errorf("safe codec qualification unavailable")
	}
	digest, err := q.Digest()
	if err != nil || digest != binding.QualificationRef || q.ProjectID != input.ProjectID ||
		q.Status != "SUPPORTED" || q.ImplementationRef != contract.CodecImplementationRef ||
		len(q.ControlResultRefs) == 0 || len(q.Limitations) != 0 ||
		q.InputProfileRef != contract.CodecProfileRef {
		return fmt.Errorf("safe codec qualification does not cover exact bytes/profile")
	}
	capability, obligation := false, len(binding.ObligationRefs) > 0
	for _, id := range q.CapabilityIDs {
		capability = capability || id == binding.CapabilityID
	}
	for _, ref := range binding.ObligationRefs {
		obligation = obligation && digestPattern.MatchString(ref)
	}
	if !capability || !obligation || !digestPattern.MatchString(binding.InputShapeRef) || !digestPattern.MatchString(binding.OutputShapeRef) {
		return fmt.Errorf("safe codec capability/obligation not qualified")
	}
	return nil
}

func sessionContext(input *CompareCaseInput, selector SessionObservation) (*EvaluationContext, error) {
	if !stepIDPattern.MatchString(selector.StepID) || selector.Visit < 1 || !digestPattern.MatchString(selector.ActorRef) {
		return nil, fmt.Errorf("invalid occurrence/actor selector")
	}
	occurrences, contexts, witness := input.SourceOccurrences, input.SourceContexts, input.SourceWitness
	if selector.Side == SideTarget {
		occurrences, contexts, witness = input.TargetOccurrences, input.TargetContexts, input.TargetWitness
	} else if selector.Side != SideSource {
		return nil, fmt.Errorf("lifecycle observation must be SOURCE or TARGET")
	}
	if witness == nil || witness.Issuer != WitnessIssuerEngineRunner || witness.ProjectID != input.ProjectID {
		return nil, fmt.Errorf("missing trusted execution witness")
	}
	witnessRaw, err := witness.CanonicalJSON()
	if err != nil || ValidateExecutionWitness(witnessRaw) != nil {
		return nil, fmt.Errorf("invalid execution witness closure")
	}
	if selector.Side == SideTarget && (witness.PlanRef == nil || *witness.PlanRef != "sha256:"+input.PlanHash ||
		witness.TargetModelRef == nil || *witness.TargetModelRef != input.SessionPlanSubject.TargetModelRef ||
		witness.TreeRef != "sha256:"+input.TargetTreeHash) {
		return nil, fmt.Errorf("stale target plan/model/tree witness")
	}
	witnessRef := "sha256:" + hexDigest(witnessRaw)
	var selected *StepOccurrence
	for i := range occurrences {
		o := &occurrences[i]
		if o.StepID == selector.StepID && o.Visit == selector.Visit && o.LogicalWorkflowID == input.WorkflowID {
			if selected != nil {
				return nil, fmt.Errorf("ambiguous occurrence")
			}
			selected = o
		}
	}
	if selected == nil || selected.ProjectID != input.ProjectID || selected.Side != selector.Side || selected.Outcome != StepOutcomeCompleted ||
		selected.ExecutionWitnessRef != witnessRef || len(selected.EvidenceRefs) == 0 {
		return nil, fmt.Errorf("missing completed occurrence")
	}
	ref := selected.PreContextRef
	if selector.Phase == PhasePostOperation {
		if selected.PostContextRef == nil {
			return nil, fmt.Errorf("missing post-operation capture")
		}
		ref = *selected.PostContextRef
	} else if selector.Phase != PhasePreOperation {
		return nil, fmt.Errorf("unsupported lifecycle phase")
	}
	ctx := contexts[ref]
	if ctx == nil || ctx.ProjectID != input.ProjectID || ctx.Side != selector.Side || ctx.Phase != selector.Phase ||
		ctx.Origin != OriginActualObserved || ctx.OccurrenceID == nil || *ctx.OccurrenceID != selected.PhysicalOccurrenceID ||
		ctx.Namespaces.Actor.FixtureActorRef != selector.ActorRef || len(ctx.EvidenceRefs) == 0 || !digestPattern.MatchString(ctx.ProjectionRef) {
		return nil, fmt.Errorf("missing/misattributed actual lifecycle capture")
	}
	for _, evidence := range ctx.EvidenceRefs {
		if !digestPattern.MatchString(evidence) {
			return nil, fmt.Errorf("invalid capture evidence reference")
		}
	}
	ap := ctx.Namespaces.Actor.ActualPrincipal
	if ap.Missing != nil || !digestPattern.MatchString(ap.CaptureChannelRef) || len(ap.EvidenceRefs) == 0 ||
		(ap.Authentication != "AUTHENTICATED" && ap.Authentication != "ANONYMOUS") ||
		(ap.Authentication == "AUTHENTICATED" && ap.SubjectLocalID == nil) ||
		(ap.Authentication == "ANONYMOUS" && (ap.SubjectLocalID != nil || len(ap.Roles) != 0)) {
		return nil, fmt.Errorf("principal capture unavailable/conflicting")
	}
	for _, evidence := range ap.EvidenceRefs {
		if !digestPattern.MatchString(evidence) {
			return nil, fmt.Errorf("invalid principal evidence reference")
		}
	}
	return ctx, nil
}

func sessionExpressionCaptured(expr *SemanticsExpression, ctx *EvaluationContext) bool {
	if expr == nil {
		return true
	}
	if expr.Op == "ref" {
		v, err := ctx.ResolvePointer("/" + expr.Namespace + expr.Pointer)
		if err != nil || v.Type == "" || v.Type == ValueTypeMissing || v.Type != expr.ValueType || ContainsMissingCapability(v) {
			return false
		}
	}
	for _, child := range []*SemanticsExpression{expr.Left, expr.Right, expr.Arg} {
		if !sessionExpressionCaptured(child, ctx) {
			return false
		}
	}
	for _, children := range [][]*SemanticsExpression{expr.Args, expr.Items} {
		for _, child := range children {
			if !sessionExpressionCaptured(child, ctx) {
				return false
			}
		}
	}
	for _, child := range expr.Fields {
		if !sessionExpressionCaptured(child, ctx) {
			return false
		}
	}
	return true
}

func sessionIdentityRelation(input *CompareCaseInput, contract *SessionContinuityContract, relation SessionLifecycleRelation,
	left, right TaggedValue, idMap *IdentityMapV2) string {
	if relation.Different || relation.Left.Observation.Side != SideSource || relation.Right.Observation.Side != SideTarget ||
		relation.Left.Path != relation.Right.Path || left.Type != ValueTypeString || right.Type != ValueTypeString || idMap == nil {
		return "UNKNOWN"
	}
	var rule *NormalizationRule
	for _, candidate := range input.NormalizationRules {
		if candidate != nil && candidate.RuleID == relation.IdentityRuleID {
			if rule != nil {
				return "UNKNOWN"
			}
			rule = candidate
		}
	}
	if rule == nil || rule.Path != relation.Left.Path || rule.ValueType != ValueTypeString ||
		(rule.Relation != "TOKEN_LIFECYCLE" && rule.Relation != "BIJECTIVE_ID") ||
		(rule.Variability != "TOKEN" && rule.Variability != "GENERATED_ID") ||
		rule.LifecycleContractRef != contract.ContractRef || rule.DecisionRef != contract.DecisionRef ||
		len(ValidateNormalizationPolicy([]*NormalizationRule{rule})) != 0 {
		return "UNKNOWN"
	}
	if len((&BijectionIntegrityChecker{}).CheckBijection(idMap)) != 0 {
		return "MISMATCH"
	}
	var binding *IdentityBinding
	for i := range idMap.Bindings {
		b := &idMap.Bindings[i]
		if b.ActorID == relation.ActorID && b.RelationGroup == rule.RelationGroup &&
			stringPtrEqual(b.ConnectionID, relation.ConnectionID) &&
			((b.SourceLocalID == left.Value || b.SourceValue.Value == left.Value) && b.SourceValue.ValueType == left.Type) {
			if binding != nil {
				return "UNKNOWN"
			}
			binding = b
		}
	}
	if binding == nil || len(binding.EvidenceRefs) == 0 {
		return "UNKNOWN"
	}
	for i, ledger := range []*LocalIdentitiesLedger{input.SourceIdentityLedger, input.TargetIdentityLedger} {
		if ledger == nil || ledger.ProjectID != input.ProjectID || ledger.LogicalWorkflowID != input.WorkflowID || ledger.Origin != OriginActualObserved || len(ledger.CaptureEvidenceRefs) == 0 {
			return "UNKNOWN"
		}
		w, side := input.SourceWitness, SideSource
		if i == 1 {
			w, side = input.TargetWitness, SideTarget
		}
		raw, err := w.CanonicalJSON()
		if err != nil || ledger.Side != side || ledger.ExecutionWitnessRef != "sha256:"+hexDigest(raw) {
			return "UNKNOWN"
		}
		found := false
		localID := binding.SourceLocalID
		if ledger.Side == SideTarget {
			localID = binding.TargetLocalID
		}
		for _, entry := range ledger.Entries {
			if entry.LocalID == localID && (entry.Kind == IdentityKindSession || entry.Kind == IdentityKindToken || entry.Kind == IdentityKindPrincipal) && len(entry.EvidenceRefs) > 0 &&
				(entry.ActorID != nil && *entry.ActorID == relation.ActorID) {
				found = true
			}
		}
		if !found {
			return "UNKNOWN"
		}
	}
	if binding.TargetLocalID != right.Value && binding.TargetValue.Value != right.Value {
		return "MISMATCH"
	}
	if binding.TargetValue.ValueType != right.Type {
		return "UNKNOWN"
	}
	return "PASS"
}

func mergeSessionContinuity(dims *VerificationDimensions, result SessionContinuityResult) {
	if *dims.Session.Verdict != "MISMATCH" && result.Verdict != "PASS" {
		dims.Session.Verdict = stringPtr(result.Verdict)
	}
	if len(result.ReasonCodes) != 0 {
		reason := strings.Join(result.ReasonCodes, "; ")
		if dims.Session.Reason != nil {
			reason = *dims.Session.Reason + "; " + reason
		}
		dims.Session.Reason = stringPtr(reason)
	}
	dims.Session.EvidenceRefs = append(dims.Session.EvidenceRefs, result.EvidenceRefs...)
	seen := make(map[string]bool)
	refs := make([]string, 0, len(dims.Session.EvidenceRefs))
	for _, ref := range dims.Session.EvidenceRefs {
		if !seen[ref] {
			refs = append(refs, ref)
			seen[ref] = true
		}
	}
	dims.Session.EvidenceRefs = refs
}

func sessionCookieAttributesEqual(a, b EvaluationCookie) bool {
	maxAgeEqual := a.MaxAge == nil && b.MaxAge == nil
	if a.MaxAge != nil && b.MaxAge != nil {
		maxAgeEqual = *a.MaxAge == *b.MaxAge
	}
	return a.Name == b.Name && a.HTTPOnly == b.HTTPOnly && a.Secure == b.Secure &&
		stringPtrEqual(a.Domain, b.Domain) && stringPtrEqual(a.Path, b.Path) &&
		stringPtrEqual(a.SameSite, b.SameSite) && stringPtrEqual(a.Expires, b.Expires) && maxAgeEqual
}

package domain

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

// This resolver is a controlled domain test double, not a customer approval or
// runtime qualification. The production port owns artifact and receipt resolution.
type admissionTestChecks struct {
	approvedHash    string
	currentSnapshot string
	currentTarget   string
	closureFailure  error
	unknown         map[string]bool
	called          []string
	mutate          bool
}

func (c *admissionTestChecks) ValidateClosure(p *PlanEnvelopeRecord, scope *AcceptanceScope) error {
	if c.closureFailure != nil {
		return c.closureFailure
	}
	if p.Subject.SnapshotRef != c.currentSnapshot {
		return errors.New("stale source qualification")
	}
	if p.Subject.TargetModelRef != c.currentTarget {
		return errors.New("wrong target")
	}
	if p.PlanHash != c.approvedHash {
		return errors.New("stale plan or invalidated approval")
	}
	if c.mutate {
		scope.CaseIDs = nil
		p.Subject.UnresolvedItems = nil
	}
	return nil
}
func (c *admissionTestChecks) require(id string) error {
	c.called = append(c.called, id)
	if c.unknown[id] {
		return errors.New("required evidence UNKNOWN")
	}
	return nil
}
func (c *admissionTestChecks) RequireContract(id string) error              { return c.require(id) }
func (c *admissionTestChecks) RequireSourceCase(id string) error            { return c.require(id) }
func (c *admissionTestChecks) RequireTargetCapability(id string) error      { return c.require(id) }
func (c *admissionTestChecks) RequireCompletedPrerequisite(id string) error { return c.require(id) }

func admissionTestPlan(t *testing.T) (*PlanEnvelopeRecord, *admissionTestChecks) {
	t.Helper()
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Name     string          `json:"name"`
		Instance json.RawMessage `json:"instance"`
	}
	if err = json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	var plan *PlanEnvelopeRecord
	for _, v := range vectors {
		if v.Name == "checkpoint-linked-executable-plan-envelope" {
			plan, err = ParsePlanEnvelope(v.Instance)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if plan == nil {
		t.Fatal("missing plan vector")
	}
	plan.Subject.OrderedUnits = []PlanApprovalUnit{
		{UnitID: "sync-auth", ContractRefs: []string{validTestRef("1"), validTestRef("2")}, CaseIDs: []string{"route", "session"}, Capabilities: []string{"http", "session"}, Priority: 10},
		{UnitID: "async", ContractRefs: []string{validTestRef("3")}, CaseIDs: []string{"completion"}, Capabilities: []string{"queue"}, Priority: 1},
	}
	plan.Subject.DependencyEdges = []PlanApprovalEdge{}
	plan.Subject.UnresolvedItems = []PlanApprovalUnresolvedItem{{Code: "async-incomplete", AffectedIDs: []string{"async", "completion"}, EvidenceRefs: []string{validTestRef("4")}, Blocking: true}}
	rehashAdmissionPlan(t, plan)
	return plan, &admissionTestChecks{approvedHash: plan.PlanHash, currentSnapshot: plan.Subject.SnapshotRef, currentTarget: plan.Subject.TargetModelRef, unknown: map[string]bool{"completion": true}}
}
func rehashAdmissionPlan(t *testing.T, p *PlanEnvelopeRecord) {
	t.Helper()
	var err error
	p.PlanHash, err = p.Subject.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.CanonicalJSON(); err != nil {
		t.Fatal(err)
	}
}
func TestGenerationAdmission_ScopedDeadlock(t *testing.T) {
	plan, checks := admissionTestPlan(t)
	before, err := plan.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	a, err := EvaluateGenerationAdmission(plan, "sync-auth", checks)
	if err != nil || a.Outcome != "ADMITTED" {
		t.Fatalf("sync/auth admission: %+v %v", a, err)
	}
	if !reflect.DeepEqual(a.Scope.CaseIDs, []string{"route", "session"}) || !reflect.DeepEqual(a.ExcludedCaseIDs, []string{"completion"}) {
		t.Fatalf("wrong scope: %+v", a)
	}
	for _, id := range checks.called {
		if id == "completion" || id == "queue" {
			t.Fatal("unrelated async dependency checked")
		}
	}
	if len(a.UnresolvedItems) != 1 || !a.UnresolvedItems[0].Blocking {
		t.Fatal("async uncertainty erased")
	}
	b, err := EvaluateGenerationAdmission(plan, "async", checks)
	if err != nil || b.Outcome != "BLOCKED" || len(b.Blockers) == 0 {
		t.Fatalf("async must block: %+v %v", b, err)
	}
	graph, err := NewPlanGraphBuilder(plan.ProjectID, plan.Subject.CorpusRef, plan.Subject.TargetModelRef, nil).
		AddUnit(SchedulableBehavioralUnit{UnitID: "sync-auth", ContractRefs: []string{validTestRef("1")}, CaseIDs: []string{"route", "session"}}).
		AddUnit(SchedulableBehavioralUnit{UnitID: "async", ContractRefs: []string{validTestRef("3")}, CaseIDs: []string{"completion"}}).Build()
	if err != nil {
		t.Fatal(err)
	}
	project, err := graph.DeriveProjectScope()
	if err != nil {
		t.Fatal(err)
	}
	if len(project.CaseIDs) != 3 {
		t.Fatal("project lost async case")
	}
	if ok, _ := graph.CanProjectComplete([]string{"sync-auth", "async"}, map[string]string{"sync-auth": "VERIFIED", "async": "UNKNOWN"}); ok {
		t.Fatal("unit readiness became project completion")
	}
	after, err := plan.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("admission mutated input plan")
	}
}
func TestGenerationAdmission_FailClosed(t *testing.T) {
	for _, id := range []string{"route", "session", validTestRef("1"), validTestRef("2"), "http"} {
		t.Run(id, func(t *testing.T) {
			p, c := admissionTestPlan(t)
			c.unknown[id] = true
			r, e := EvaluateGenerationAdmission(p, "sync-auth", c)
			if e != nil || r.Outcome != "BLOCKED" {
				t.Fatalf("missing %s admitted: %+v %v", id, r, e)
			}
		})
	}
	for _, reason := range []string{"stale M2 qualification", "invalid approval", "stale M3 checkpoint", "target-contaminated source", "plan-contaminated source", "missing target lowering"} {
		t.Run(reason, func(t *testing.T) {
			p, c := admissionTestPlan(t)
			c.closureFailure = errors.New(reason)
			r, e := EvaluateGenerationAdmission(p, "sync-auth", c)
			if e != nil || r.Outcome != "BLOCKED" || len(c.called) != 0 {
				t.Fatalf("closure failure bypassed: %+v %v", r, e)
			}
		})
	}
	for _, field := range []string{"source", "target", "plan"} {
		t.Run(field, func(t *testing.T) {
			p, c := admissionTestPlan(t)
			switch field {
			case "source":
				p.Subject.SnapshotRef = validTestRef("8")
			case "target":
				p.Subject.TargetModelRef = validTestRef("9")
			case "plan":
				p.Subject.OrderedUnits[0].CaseIDs = []string{"route"}
			}
			rehashAdmissionPlan(t, p)
			r, e := EvaluateGenerationAdmission(p, "sync-auth", c)
			if e != nil || r.Outcome != "BLOCKED" {
				t.Fatalf("changed %s admitted: %+v %v", field, r, e)
			}
		})
	}
	p, c := admissionTestPlan(t)
	p.Subject.DependencyEdges = []PlanApprovalEdge{{PrerequisiteID: "async", DependentID: "sync-auth"}}
	rehashAdmissionPlan(t, p)
	c.approvedHash = p.PlanHash
	c.unknown["async"] = true
	r, e := EvaluateGenerationAdmission(p, "sync-auth", c)
	if e != nil || r.Outcome != "BLOCKED" {
		t.Fatalf("real async dependency ignored: %+v %v", r, e)
	}
}
func TestGenerationAdmission_RejectScopeManipulation(t *testing.T) {
	p, c := admissionTestPlan(t)
	c.mutate = true
	c.unknown["session"] = true
	r, e := EvaluateGenerationAdmission(p, "sync-auth", c)
	if e != nil || r.Outcome != "BLOCKED" || len(r.Scope.CaseIDs) != 2 {
		t.Fatalf("guard mutated derived scope: %+v %v", r, e)
	}
	if _, e = EvaluateGenerationAdmission(p, "unplanned", c); e == nil {
		t.Fatal("agent invented unit")
	}
	r, e = EvaluateGenerationAdmission(p, "sync-auth", nil)
	if e != nil || r.Outcome != "BLOCKED" {
		t.Fatal("missing guard admitted")
	}
	p.PlanHash = "invalid"
	if _, e = EvaluateGenerationAdmission(p, "sync-auth", c); e == nil {
		t.Fatal("invalid plan hash admitted")
	}
	if _, e = EvaluateGenerationAdmission(nil, "sync-auth", c); e == nil {
		t.Fatal("nil plan admitted")
	}
}

func TestGenerationAdmission_RequiredUnresolvedAndReproduction(t *testing.T) {
	p, c := admissionTestPlan(t)
	a, e := EvaluateGenerationAdmission(p, "sync-auth", c)
	if e != nil {
		t.Fatal(e)
	}
	b, e := EvaluateGenerationAdmission(p, "sync-auth", c)
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatalf("nondeterministic admission: %+v %+v %v", a, b, e)
	}
	raw, e := json.Marshal(a.Scope)
	if e != nil {
		t.Fatal(e)
	}
	if e = ValidateAcceptanceScope(raw); e != nil {
		t.Fatal(e)
	}
	ref, e := p.Digest()
	if e != nil {
		t.Fatal(e)
	}
	if a.PlanRef != ref || a.SnapshotRef != p.Subject.SnapshotRef || a.TargetModelRef != p.Subject.TargetModelRef {
		t.Fatal("lost closure identities")
	}
	p.Subject.UnresolvedItems = append(p.Subject.UnresolvedItems, PlanApprovalUnresolvedItem{Code: "missing-session-proof", AffectedIDs: []string{"session"}, EvidenceRefs: []string{validTestRef("5")}, Blocking: true})
	rehashAdmissionPlan(t, p)
	c.approvedHash = p.PlanHash
	a, e = EvaluateGenerationAdmission(p, "sync-auth", c)
	if e != nil || a.Outcome != "BLOCKED" {
		t.Fatalf("in-scope blocking uncertainty waived: %+v %v", a, e)
	}
	// Mutable result objects cannot rewrite the approved input.
	a.UnresolvedItems[0].AffectedIDs[0] = "mutated"
	if p.Subject.UnresolvedItems[0].AffectedIDs[0] != "async" {
		t.Fatal("result aliases approved input")
	}
}

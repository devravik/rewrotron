package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devravik/rewrotron/internal/domain"
)

func actual(t *testing.T) ([]byte, string) {
	t.Helper()
	r := filepath.Join("../../../fixtures/ci3-reference/evidence", "corpus-0.json")
	raw, e := os.ReadFile(r)
	if e != nil {
		t.Fatal(e)
	}
	canon, e := domain.CanonicalJSON(raw)
	if e != nil {
		t.Fatal(e)
	}
	return raw, hash(canon)
}
func buildBundle(t *testing.T, raw []byte) *bundle {
	t.Helper()
	canon, e := domain.CanonicalJSON(raw)
	if e != nil {
		t.Fatal(e)
	}
	b, e := project(raw, hash(canon), hash([]byte("test-projector")))
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func postContext(t *testing.T, b *bundle, status int) *domain.EvaluationContext {
	t.Helper()
	for _, c := range b.Contexts {
		if c.Namespaces.Result != nil && c.Namespaces.Result.Status == status {
			return c
		}
	}
	t.Fatal("captured status missing")
	return nil
}

func TestActualSourceProjection(t *testing.T) {
	raw, ref := actual(t)
	before := append([]byte(nil), raw...)
	b, e := project(raw, ref, hash([]byte("test-projector")))
	if e != nil {
		t.Fatal(e)
	}
	if len(b.Contexts) != 60 {
		t.Fatalf("want 30 named occurrences x 2 phases, got %d", len(b.Contexts))
	}
	if string(raw) != string(before) {
		t.Fatal("mutated capture")
	}
	seen := map[string]bool{}
	for _, c := range b.Contexts {
		encoded, e := canonical(c)
		if e != nil {
			t.Fatal(e)
		}
		if e = domain.ValidateEvaluationContext(encoded); e != nil {
			t.Fatal(e)
		}
		if seen[c.ContextID] {
			t.Fatal("collapsed occurrence")
		}
		seen[c.ContextID] = true
		if c.Side != domain.SideSource || c.Origin != domain.OriginActualObserved || c.Namespaces.Actor.ActualPrincipal.Missing == nil || c.Namespaces.Clock.Type != domain.ValueTypeMissing || c.Namespaces.State.DB.Missing == nil {
			t.Fatal("wrong source provenance or invented evidence")
		}
	}
	c := postContext(t, b, 303)
	if len(c.Namespaces.Result.Cookies) != 2 {
		t.Fatal("folded cookies")
	}
	cookie := c.Namespaces.Result.Cookies[0]
	if !cookie.HTTPOnly || cookie.Secure || cookie.Path == nil || *cookie.Path != "/" || cookie.MaxAge == nil || *cookie.MaxAge != 60 || cookie.SameSite == nil || *cookie.SameSite != "Lax" {
		t.Fatal("security/lifetime attributes lost")
	}
	if c.Namespaces.Result.Body.Type != domain.ValueTypeObject {
		t.Fatal("body not typed")
	}
	expr := &domain.SemanticsExpression{Op: "eq", Left: &domain.SemanticsExpression{Op: "ref", Namespace: "actor", Pointer: "/actual_principal/authentication", ValueType: "string"}, Right: &domain.SemanticsExpression{Op: "literal", ValueType: "string", Value: "AUTHENTICATED"}}
	value, e := domain.EvaluateExpression(expr, c)
	if e != nil {
		t.Fatal(e)
	}
	if value.Type != domain.ValueTypeMissing {
		t.Fatal("unobserved principal became authentication truth")
	}
}

func TestProjectionRejectsFalsePrerequisites(t *testing.T) {
	for _, variant := range []string{"target-root", "target-fragment", "target-snapshot", "paired-case", "duplicate-step", "missing-step", "missing-before", "expected-effect", "expected-principal", "wrong-workflow", "wrong-hash"} {
		t.Run(variant, func(t *testing.T) {
			raw, ref := actual(t)
			var c map[string]any
			if e := json.Unmarshal(raw, &c); e != nil {
				t.Fatal(e)
			}
			first := c["observations"].([]any)[0].(map[string]any)
			switch variant {
			case "wrong-hash":
				first["response"].(map[string]any)["body_base64"] = base64.StdEncoding.EncodeToString([]byte(`{"changed":true}`))
			case "target-root":
				c["side"] = "TARGET"
			case "target-fragment":
				first["side"] = "TARGET"
			case "target-snapshot":
				first["before"].(map[string]any)["side"] = "TARGET"
			case "paired-case":
				c["case_id"] = "future_case"
			case "duplicate-step":
				c["observations"].([]any)[1].(map[string]any)["step_id"] = first["step_id"]
			case "missing-step":
				delete(first, "step_id")
			case "missing-before":
				delete(first, "before")
			case "expected-effect":
				first["expected_effects"] = []any{map[string]any{"kind": "DB", "operation": "INSERT"}}
			case "expected-principal":
				first["before"].(map[string]any)["actual_principal"] = "alice"
			case "wrong-workflow":
				first["scenario_id"] = "other_workflow"
			}
			changed, e := canonical(c)
			if e != nil {
				t.Fatal(e)
			}
			if variant != "wrong-hash" {
				ref = hash(changed)
			}
			if _, e = project(changed, ref, hash([]byte("test-projector"))); e == nil {
				t.Fatalf("accepted %s", variant)
			}
		})
	}
}

func TestMissingCaptureRemainsUnknown(t *testing.T) {
	raw, _ := actual(t)
	var c map[string]any
	if e := json.Unmarshal(raw, &c); e != nil {
		t.Fatal(e)
	}
	first := c["observations"].([]any)[0].(map[string]any)
	delete(first["before"].(map[string]any), "source_files")
	delete(first["response"].(map[string]any), "headers")
	changed, e := canonical(c)
	if e != nil {
		t.Fatal(e)
	}
	b := buildBundle(t, changed)
	if b.Contexts[0].Namespaces.State.Sessions.Type != domain.ValueTypeMissing || b.Contexts[0].Namespaces.State.Files.Type != domain.ValueTypeMissing || b.Contexts[1].Namespaces.Result != nil {
		t.Fatal("missing capture became empty/PASS")
	}
}

func TestNullCaptureRemainsUnknown(t *testing.T) {
	b := &bundle{Artifacts: map[string][]byte{}}
	state, err := projectState(b, json.RawMessage(`{"side":"SOURCE","source_files":null}`))
	if err != nil || state.Files.Type != domain.ValueTypeMissing || state.Sessions.Type != domain.ValueTypeMissing {
		t.Fatal("null file capture became empty", err)
	}
	for _, raw := range []string{
		`{"status":200,"headers":null,"body_base64":""}`,
		`{"status":200,"headers":[],"body_base64":null}`,
		`{"status":null,"headers":[],"body_base64":""}`,
	} {
		response, err := projectResponse(b, json.RawMessage(raw))
		if err != nil || response != nil {
			t.Fatal("null response capture became known", err)
		}
	}
	for _, raw := range []string{
		`{"side":"SOURCE","source_files":{"/state/a":{"content":null,"mode":"0600"}}}`,
		`{"side":"SOURCE","source_files":{"/state/a":{"content":"","mode":null}}}`,
	} {
		if _, err := projectState(b, json.RawMessage(raw)); err == nil {
			t.Fatal("null file field became known")
		}
	}
}

func TestCookieBoundaries(t *testing.T) {
	for _, wire := range []string{"sid=a; Path=/; Path=/wrong", "sid=a; Partitioned", "sid=a; Secure=false"} {
		if _, e := projectCookie(wire); e == nil {
			t.Fatalf("normalized unsupported cookie: %s", wire)
		}
	}
	for _, wire := range []string{"sid=a", "sid=; Max-Age=0; HttpOnly; SameSite=Lax"} {
		c, e := projectCookie(wire)
		if e != nil {
			t.Fatal(e)
		}
		if strings.Contains(wire, "Max-Age=0") && (c.MaxAge == nil || *c.MaxAge != 0) {
			t.Fatal("null confused with zero")
		}
	}
}

func TestNumbersAndBinaryBodies(t *testing.T) {
	b := &bundle{Artifacts: map[string][]byte{}}
	encoded := base64.StdEncoding.EncodeToString([]byte(`{"integer":9007199254740993,"decimal":1e0,"null":null}`))
	raw, e := canonical(map[string]any{"status": 200, "headers": [][]string{{"Content-Type", "application/json"}}, "body_base64": encoded})
	if e != nil {
		t.Fatal(e)
	}
	response, e := projectResponse(b, raw)
	if e != nil {
		t.Fatal(e)
	}
	object := response.Body.Value.(map[string]domain.TaggedValue)
	if object["integer"].Value != "9007199254740993" || object["decimal"].Type != "decimal" || object["null"].Type != "null" {
		t.Fatal("lexical representation changed")
	}
	raw, e = canonical(map[string]any{"status": 200, "headers": [][]string{{"Content-Type", "application/json"}}, "body_base64": base64.StdEncoding.EncodeToString([]byte{0xff})})
	if e != nil {
		t.Fatal(e)
	}
	response, e = projectResponse(b, raw)
	if e != nil {
		t.Fatal(e)
	}
	if response.Body.Type != domain.ValueTypeBytes {
		t.Fatal("lossy binary decoding")
	}
}

func TestNamedIdentitySurvivesReordering(t *testing.T) {
	raw, _ := actual(t)
	original := buildBundle(t, raw)
	var source map[string]any
	if e := json.Unmarshal(raw, &source); e != nil {
		t.Fatal(e)
	}
	observations := source["observations"].([]any)
	observations[0], observations[1] = observations[1], observations[0]
	reordered, e := canonical(source)
	if e != nil {
		t.Fatal(e)
	}
	other := buildBundle(t, reordered)
	identities := map[string]string{}
	for _, c := range original.Contexts {
		identities[c.ContextID] = *c.OccurrenceID
	}
	for _, c := range other.Contexts {
		if identities[c.ContextID] != *c.OccurrenceID {
			t.Fatal("array index used as occurrence identity")
		}
	}
}

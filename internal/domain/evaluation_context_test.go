package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestValidateEvaluationContext_ValidationCases(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	// 1. Evaluation Context schema cases
	contextSchema := "https://rewrotron.dev/specs/behavior/evaluation-context.schema.json"
	contextCount := 0
	for _, tc := range cases {
		if tc.Schema != contextSchema {
			continue
		}
		contextCount++
		t.Run("Context_"+tc.Name, func(t *testing.T) {
			err := ValidateEvaluationContext(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid context %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid context %q to fail, but it passed", tc.Name)
			}
		})
	}
	if contextCount < 7 {
		t.Fatalf("expected at least 7 evaluation context cases, found %d", contextCount)
	}

	// 2. Tagged value ($defs/value) cases
	valueSchema := "https://rewrotron.dev/specs/behavior/evaluation-context.schema.json#/$defs/value"
	valueCount := 0
	for _, tc := range cases {
		if tc.Schema != valueSchema {
			continue
		}
		valueCount++
		t.Run("Value_"+tc.Name, func(t *testing.T) {
			err := ValidateTaggedValue(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid tagged value %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid tagged value %q to fail, but it passed", tc.Name)
			}
		})
	}
	if valueCount < 8 {
		t.Fatalf("expected at least 8 tagged value cases, found %d", valueCount)
	}

	// 3. Response schema cases
	responseSchema := "https://rewrotron.dev/specs/behavior/response.schema.json"
	responseCount := 0
	for _, tc := range cases {
		if tc.Schema != responseSchema {
			continue
		}
		responseCount++
		t.Run("Response_"+tc.Name, func(t *testing.T) {
			err := ValidateResponse(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid response %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid response %q to fail, but it passed", tc.Name)
			}
		})
	}
	if responseCount < 2 {
		t.Fatalf("expected at least 2 response cases, found %d", responseCount)
	}
}

func TestProjectRawJSONNumber_BlockerVectors(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/contract-blocker-vectors.json")
	if err != nil {
		t.Fatalf("failed to read contract blocker vectors: %v", err)
	}

	var data struct {
		JSONNumbers []struct {
			RawToken string      `json:"raw_token"`
			Expected TaggedValue `json:"expected"`
		} `json:"json_numbers"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("failed to unmarshal contract blocker vectors: %v", err)
	}

	if len(data.JSONNumbers) < 20 {
		t.Fatalf("expected at least 20 json_numbers vectors, found %d", len(data.JSONNumbers))
	}

	for _, v := range data.JSONNumbers {
		t.Run("Token_"+v.RawToken, func(t *testing.T) {
			actual := ProjectRawJSONNumber(v.RawToken)
			if actual.Type != v.Expected.Type {
				t.Fatalf("token %q: expected type %q, got %q", v.RawToken, v.Expected.Type, actual.Type)
			}
			if actual.Type == ValueTypeMissing {
				if actual.ReasonCode != v.Expected.ReasonCode {
					t.Fatalf("token %q: expected reason_code %q, got %q", v.RawToken, v.Expected.ReasonCode, actual.ReasonCode)
				}
			} else {
				if actual.Value != v.Expected.Value {
					t.Fatalf("token %q: expected value %v, got %v", v.RawToken, v.Expected.Value, actual.Value)
				}
			}
		})
	}
}

func TestNumericComparison_BlockerVectors(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/contract-blocker-vectors.json")
	if err != nil {
		t.Fatalf("failed to read contract blocker vectors: %v", err)
	}

	var data struct {
		NumericComparison []struct {
			Left          TaggedValue `json:"left"`
			Right         TaggedValue `json:"right"`
			Expected      string      `json:"expected,omitempty"`
			ExpectedEqual *bool       `json:"expected_equal,omitempty"`
			Reason        string      `json:"reason"`
		} `json:"numeric_comparison"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("failed to unmarshal contract blocker vectors: %v", err)
	}

	for i, v := range data.NumericComparison {
		t.Run(v.Reason, func(t *testing.T) {
			eq, triState, err := CompareTaggedValues(v.Left, v.Right)
			if err != nil {
				t.Fatalf("vector %d unexpected error: %v", i, err)
			}
			if v.Expected != "" && triState != v.Expected {
				t.Fatalf("vector %d: expected triState %q, got %q", i, v.Expected, triState)
			}
			if v.ExpectedEqual != nil && eq != *v.ExpectedEqual {
				t.Fatalf("vector %d: expected equal %v, got %v", i, *v.ExpectedEqual, eq)
			}
		})
	}
}

func TestEvaluationContext_ResolvePointer(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Instance json.RawMessage `json:"instance"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	var postContextRaw json.RawMessage
	var preContextRaw json.RawMessage
	for _, c := range cases {
		if c.Name == "review2-canonical-result-body" {
			postContextRaw = c.Instance
		}
		if c.Name == "review2-actual-pre-context-exact-id" {
			preContextRaw = c.Instance
		}
	}
	if postContextRaw == nil || preContextRaw == nil {
		t.Fatal("missing required evaluation context fixtures")
	}

	postCtx, err := ParseEvaluationContext(postContextRaw)
	if err != nil {
		t.Fatalf("failed to parse postContext: %v", err)
	}

	preCtx, err := ParseEvaluationContext(preContextRaw)
	if err != nil {
		t.Fatalf("failed to parse preCtx: %v", err)
	}

	// 1. /result/body/order_id selects the exact parsed member
	val, err := postCtx.ResolvePointer("/result/body/order_id")
	if err != nil {
		t.Fatalf("pointer /result/body/order_id failed: %v", err)
	}
	if val.Type != ValueTypeInteger || val.Value != "9007199254740993" {
		t.Fatalf("wanted integer 9007199254740993, got: %+v", val)
	}

	// 2. /result/status selects the integer status
	statusVal, err := postCtx.ResolvePointer("/result/status")
	if err != nil {
		t.Fatalf("pointer /result/status failed: %v", err)
	}
	if statusVal.Type != ValueTypeInteger || statusVal.Value != "302" {
		t.Fatalf("wanted integer 302, got: %+v", statusVal)
	}

	// 3. /input/method selects "GET"
	methodVal, err := postCtx.ResolvePointer("/input/method")
	if err != nil {
		t.Fatalf("pointer /input/method failed: %v", err)
	}
	if methodVal.Type != ValueTypeString || methodVal.Value != "GET" {
		t.Fatalf("wanted string 'GET', got: %+v", methodVal)
	}

	// 4. /input/raw_target selects "/orders?flag&flag="
	targetVal, err := postCtx.ResolvePointer("/input/raw_target")
	if err != nil {
		t.Fatalf("pointer /input/raw_target failed: %v", err)
	}
	if targetVal.Type != ValueTypeString || targetVal.Value != "/orders?flag&flag=" {
		t.Fatalf("wanted string '/orders?flag&flag=', got: %+v", targetVal)
	}

	// 5. /binding/large_id selects "9007199254740993"
	bindingVal, err := postCtx.ResolvePointer("/binding/large_id")
	if err != nil {
		t.Fatalf("pointer /binding/large_id failed: %v", err)
	}
	if bindingVal.Type != ValueTypeInteger || bindingVal.Value != "9007199254740993" {
		t.Fatalf("wanted integer 9007199254740993, got: %+v", bindingVal)
	}

	// 6. Unbound binding property returns ERR_UNBOUND_VALUE
	unboundVal, err := postCtx.ResolvePointer("/binding/missing_var")
	if err != nil {
		t.Fatalf("pointer /binding/missing_var failed: %v", err)
	}
	if unboundVal.Type != ValueTypeMissing || unboundVal.ReasonCode != "ERR_UNBOUND_VALUE" {
		t.Fatalf("wanted missing ERR_UNBOUND_VALUE, got: %+v", unboundVal)
	}

	// 7. Nonexistent property in object returns ERR_MISSING_PROPERTY
	missingProp, err := postCtx.ResolvePointer("/result/nonexistent")
	if err != nil {
		t.Fatalf("pointer /result/nonexistent failed: %v", err)
	}
	if missingProp.Type != ValueTypeMissing || missingProp.ReasonCode != "ERR_MISSING_PROPERTY" {
		t.Fatalf("wanted missing ERR_MISSING_PROPERTY, got: %+v", missingProp)
	}

	// 8. Traversing through scalar returns ERR_TYPE
	typeErrVal, err := postCtx.ResolvePointer("/binding/large_id/child")
	if err != nil {
		t.Fatalf("pointer /binding/large_id/child failed: %v", err)
	}
	if typeErrVal.Type != ValueTypeMissing || typeErrVal.ReasonCode != "ERR_TYPE" {
		t.Fatalf("wanted missing ERR_TYPE, got: %+v", typeErrVal)
	}

	// 9. Unavailable namespace (preCtx has result == nil) returns ERR_UNAVAILABLE_NAMESPACE
	unavailVal, err := preCtx.ResolvePointer("/result/body/order_id")
	if err != nil {
		t.Fatalf("pointer on preCtx /result/body/order_id failed: %v", err)
	}
	if unavailVal.Type != ValueTypeMissing || unavailVal.ReasonCode != "ERR_UNAVAILABLE_NAMESPACE" {
		t.Fatalf("wanted missing ERR_UNAVAILABLE_NAMESPACE, got: %+v", unavailVal)
	}

	// 10. Clock namespace selects timestamp
	clockVal, err := postCtx.ResolvePointer("/clock")
	if err != nil {
		t.Fatalf("pointer /clock failed: %v", err)
	}
	if clockVal.Type != ValueTypeTimestamp || clockVal.Value != "2026-10-03T00:00:00Z" {
		t.Fatalf("wanted timestamp 2026-10-03T00:00:00Z, got: %+v", clockVal)
	}
}

func TestProvenanceIsolation_X61_X63(t *testing.T) {
	// Base valid context
	valid := map[string]any{
		"schema_version": "1.0",
		"context_id":     "ctx-test",
		"project_id":     "project-1",
		"side":           "SOURCE",
		"occurrence_id":  "occ-1",
		"projection_ref": "sha256:" + strings.Repeat("a", 64),
		"phase":          "PRE_OPERATION",
		"origin":         "ACTUAL_OBSERVED",
		"evidence_refs": []string{
			"sha256:" + strings.Repeat("b", 64),
		},
		"namespaces": map[string]any{
			"input":   nil,
			"actor": map[string]any{
				"fixture_actor_ref": "sha256:" + strings.Repeat("c", 64),
				"actual_principal": map[string]any{
					"type":        "missing",
					"reason_code": "NOT_CAPTURED",
				},
			},
			"state": map[string]any{
				"db":       []any{},
				"sessions": map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
				"cache":    map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
				"files":    map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
				"queues":   map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
			},
			"result":  nil,
			"binding": map[string]any{},
			"row":     nil,
			"clock":   map[string]any{"type": "timestamp", "value": "2026-10-03T00:00:00Z"},
		},
	}

	mustMarshal := func(m map[string]any) []byte {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	// 1. Control passes
	if err := ValidateEvaluationContext(mustMarshal(valid)); err != nil {
		t.Fatalf("control should pass: %v", err)
	}

	// 2. Projection poisoning (X63): MODEL_EXPECTED in PRE_OPERATION must be rejected
	poisoned := copyMap(valid)
	poisoned["origin"] = "MODEL_EXPECTED"
	if err := ValidateEvaluationContext(mustMarshal(poisoned)); err == nil {
		t.Fatal("poisoned origin MODEL_EXPECTED in PRE_OPERATION must fail")
	}

	// 3. Shadow side in actual phase must be rejected
	shadowInActual := copyMap(valid)
	shadowInActual["side"] = "SHADOW"
	if err := ValidateEvaluationContext(mustMarshal(shadowInActual)); err == nil {
		t.Fatal("SHADOW side in PRE_OPERATION must fail")
	}

	// 4. Source side in SIMULATION must be rejected
	simSource := copyMap(valid)
	simSource["phase"] = "SIMULATION"
	simSource["origin"] = "MODEL_EXPECTED"
	simSource["side"] = "SOURCE"
	if err := ValidateEvaluationContext(mustMarshal(simSource)); err == nil {
		t.Fatal("SOURCE side in SIMULATION phase must fail")
	}

	// 5. Valid SIMULATION must have side SHADOW and origin MODEL_EXPECTED
	validSim := copyMap(valid)
	validSim["phase"] = "SIMULATION"
	validSim["origin"] = "MODEL_EXPECTED"
	validSim["side"] = "SHADOW"
	validSim["occurrence_id"] = nil
	if err := ValidateEvaluationContext(mustMarshal(validSim)); err != nil {
		t.Fatalf("valid SIMULATION context should pass, got: %v", err)
	}

	// 6. Pre-operation with non-null result must be rejected
	preWithResult := copyMap(valid)
	preWithResult["namespaces"] = copyMap(valid["namespaces"].(map[string]any))
	preWithResult["namespaces"].(map[string]any)["result"] = map[string]any{
		"schema_version":    "1.0",
		"status":            200,
		"headers":           []any{},
		"cookies":           []any{},
		"body":              map[string]any{"type": "string", "value": "ok"},
		"session":           map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
		"principal":         map[string]any{"type": "missing", "reason_code": "NOT_CAPTURED"},
		"effect_refs":       []any{},
		"application_error": map[string]any{"type": "null", "value": nil},
	}
	if err := ValidateEvaluationContext(mustMarshal(preWithResult)); err == nil {
		t.Fatal("PRE_OPERATION with non-null result must fail")
	}
}

func TestCanonicalJSON_EvaluationContext(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Instance json.RawMessage `json:"instance"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	for _, c := range cases {
		if c.Name == "review2-canonical-result-body" {
			ctx, err := ParseEvaluationContext(c.Instance)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			canonical1, err := ctx.CanonicalJSON()
			if err != nil {
				t.Fatalf("canonical JSON serialization failed: %v", err)
			}
			canonical2, err := ctx.CanonicalJSON()
			if err != nil {
				t.Fatalf("second canonical JSON serialization failed: %v", err)
			}
			if string(canonical1) != string(canonical2) {
				t.Fatal("canonical serialization is not deterministic")
			}
		}
	}
}

func copyMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

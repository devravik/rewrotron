package domain

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLosslessJSON_BlockerVectors(t *testing.T) {
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

	for _, v := range data.JSONNumbers {
		t.Run("Token_"+v.RawToken, func(t *testing.T) {
			// Test direct ProjectRawJSONNumber
			direct := ProjectRawJSONNumber(v.RawToken)
			if direct.Type != v.Expected.Type {
				t.Fatalf("direct %q: expected type %q, got %q", v.RawToken, v.Expected.Type, direct.Type)
			}
			if direct.Type == ValueTypeMissing {
				if direct.ReasonCode != v.Expected.ReasonCode {
					t.Fatalf("direct %q: expected reason %q, got %q", v.RawToken, v.Expected.ReasonCode, direct.ReasonCode)
				}
			} else {
				if direct.Value != v.Expected.Value {
					t.Fatalf("direct %q: expected value %v, got %v", v.RawToken, v.Expected.Value, direct.Value)
				}
			}

			// Test ParseLosslessJSON on a JSON document wrapping the token: {"val": <token>}
			// Only valid JSON numbers will parse at top level in standard JSON syntax
			validJSON := `{"val": ` + v.RawToken + `}`
			projected, parseErr := ParseLosslessJSON([]byte(validJSON))
			if v.Expected.Type == ValueTypeMissing {
				// Invalid syntax or out-of-bounds tokens must produce error or MISSING capability
				if parseErr == nil {
					// If json.Decoder admitted it (e.g. 1e38, 1e-19), the node itself must be MISSING
					fieldVal, ok := ResolveLosslessField(projected, "val")
					if !ok || fieldVal.Type != ValueTypeMissing || fieldVal.ReasonCode != v.Expected.ReasonCode {
						t.Fatalf("JSON payload %q: expected missing node with %q, got %v (err=%v)", validJSON, v.Expected.ReasonCode, fieldVal, parseErr)
					}
					if !ContainsMissingCapability(projected) {
						t.Fatalf("expected ContainsMissingCapability to be true for %q", validJSON)
					}
				}
			} else {
				if parseErr != nil {
					t.Fatalf("JSON payload %q failed unexpectedly: %v", validJSON, parseErr)
				}
				fieldVal, ok := ResolveLosslessField(projected, "val")
				if !ok {
					t.Fatalf("field 'val' missing in %v", projected)
				}
				if fieldVal.Type != v.Expected.Type {
					t.Fatalf("JSON %q: expected type %q, got %q", v.RawToken, v.Expected.Type, fieldVal.Type)
				}
				if fieldVal.Value != v.Expected.Value {
					t.Fatalf("JSON %q: expected value %v, got %v", v.RawToken, v.Expected.Value, fieldVal.Value)
				}
			}
		})
	}
}

func TestX21_MutantsAndNegativeControls(t *testing.T) {
	// Mutant 1: Standard float64 decoding rounds 9007199254740993 to 9007199254740992
	t.Run("Mutant_Float64RoundingDetection", func(t *testing.T) {
		raw := []byte(`{"id": 9007199254740993}`)

		// Standard float decoding loses precision
		var stdMap map[string]any
		if err := json.Unmarshal(raw, &stdMap); err != nil {
			t.Fatalf("failed standard unmarshal: %v", err)
		}
		floatVal := stdMap["id"].(float64)
		if int64(floatVal) == 9007199254740993 {
			t.Fatalf("expected float64 to lose precision on 9007199254740993")
		}
		if int64(floatVal) != 9007199254740992 {
			t.Fatalf("expected float64 to round to 9007199254740992, got %v", int64(floatVal))
		}

		// Rewrotron lossless projector preserves exact integer
		projected, err := ParseLosslessJSON(raw)
		if err != nil {
			t.Fatalf("lossless parse failed: %v", err)
		}
		val, ok := ResolveLosslessField(projected, "id")
		if !ok {
			t.Fatalf("expected 'id' field present")
		}
		if val.Type != ValueTypeInteger || val.Value != "9007199254740993" {
			t.Fatalf("lossless parse did not preserve large integer: got %v", val)
		}

		// Comparing the exact ID with the float64 rounded mutant yields MISMATCH
		floatMutant := TaggedValue{Type: ValueTypeInteger, Value: "9007199254740992"}
		eq, triState, _ := CompareTaggedValues(val, floatMutant)
		if eq || triState != "MISMATCH" {
			t.Fatalf("expected MISMATCH when comparing exact ID to float-rounded mutant, got triState=%s eq=%v", triState, eq)
		}
	})

	// Mutant 2: Retagging decimal 1.0 as integer
	t.Run("Mutant_DecimalRetaggedAsInteger", func(t *testing.T) {
		decVal := ProjectRawJSONNumber("1.0") // decimal "1"
		if decVal.Type != ValueTypeDecimal || decVal.Value != "1" {
			t.Fatalf("expected decimal '1', got %v", decVal)
		}

		intMutant := TaggedValue{Type: ValueTypeInteger, Value: "1"}
		eq, triState, _ := CompareTaggedValues(decVal, intMutant)
		if eq || triState != "UNKNOWN" {
			t.Fatalf("expected UNKNOWN for cross-type comparison without coercion, got %s", triState)
		}

		// Explicit coercion allows comparison
		coerced, err := CoerceIntegerToDecimal(intMutant)
		if err != nil {
			t.Fatalf("coercion failed: %v", err)
		}
		eq2, triState2, _ := CompareTaggedValues(decVal, coerced)
		if !eq2 || triState2 != "PASS" {
			t.Fatalf("expected PASS after explicit coercion, got %s", triState2)
		}
	})

	// Mutant 3: Duplicate object keys rejected
	t.Run("Mutant_DuplicateObjectKeys", func(t *testing.T) {
		raw := []byte(`{"code": "A", "code": "B"}`)
		_, err := ParseLosslessJSON(raw)
		if err == nil {
			t.Fatalf("expected error on duplicate keys, got nil")
		}
	})

	// Mutant 4: Overflow decimal bound 1e38
	t.Run("Mutant_OverflowDecimalBound", func(t *testing.T) {
		raw := []byte(`{"rate": 1e38}`)
		val, err := ParseLosslessJSON(raw)
		if err != nil {
			// Some strict parsers reject earlier; both error or MISSING node are acceptable
			return
		}
		if !ContainsMissingCapability(val) {
			t.Fatalf("expected ContainsMissingCapability to flag 1e38")
		}
	})
}

func TestX21_QueryRepresentationFidelity(t *testing.T) {
	// 1. Absent query (/orders)
	reqAbsent, err := ParseLosslessHTTPRequest("GET", "/orders", nil, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if reqAbsent.HasQuery {
		t.Errorf("expected HasQuery=false for absent query")
	}
	if reqAbsent.Query != nil {
		t.Errorf("expected Query=nil for absent query, got %v", reqAbsent.Query)
	}

	// 2. Empty query (/orders?)
	reqEmpty, err := ParseLosslessHTTPRequest("GET", "/orders?", nil, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !reqEmpty.HasQuery {
		t.Errorf("expected HasQuery=true for empty query")
	}
	if reqEmpty.Query == nil || len(reqEmpty.Query) != 0 {
		t.Errorf("expected Query=[] for empty query, got %v", reqEmpty.Query)
	}

	// Compare absent vs empty: MUST be distinct!
	eq, verdict, diffs := CompareLosslessRequests(reqAbsent, reqEmpty)
	if eq || verdict != "MISMATCH" {
		t.Fatalf("expected MISMATCH comparing absent query to empty query, got verdict=%s", verdict)
	}
	if len(diffs) == 0 {
		t.Fatalf("expected diffs explaining query presence mismatch")
	}

	// 3. Parameter without equals (?flag) vs parameter with empty value (?flag=)
	reqParamNoEq, err := ParseLosslessHTTPRequest("GET", "/orders?flag", nil, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	valNoEq, exists1 := reqParamNoEq.GetQueryParam("flag")
	if !exists1 || valNoEq != nil {
		t.Fatalf("expected ?flag to have nil value, got exists=%v val=%v", exists1, valNoEq)
	}

	reqParamEmptyVal, err := ParseLosslessHTTPRequest("GET", "/orders?flag=", nil, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	valEmpty, exists2 := reqParamEmptyVal.GetQueryParam("flag")
	if !exists2 || valEmpty == nil || *valEmpty != "" {
		t.Fatalf("expected ?flag= to have empty string value, got exists=%v val=%v", exists2, valEmpty)
	}

	// Compare ?flag vs ?flag=: MUST be distinct!
	eqParam, verdictParam, _ := CompareLosslessRequests(reqParamNoEq, reqParamEmptyVal)
	if eqParam || verdictParam != "MISMATCH" {
		t.Fatalf("expected MISMATCH comparing ?flag to ?flag=, got verdict=%s", verdictParam)
	}

	// 4. Repeated query parameters (?status=active&status=pending)
	reqRepeated, err := ParseLosslessHTTPRequest("GET", "/orders?status=active&status=pending", nil, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	vals := reqRepeated.GetAllQueryParams("status")
	if len(vals) != 2 {
		t.Fatalf("expected 2 status values, got %d", len(vals))
	}
	if vals[0] == nil || *vals[0] != "active" || vals[1] == nil || *vals[1] != "pending" {
		t.Fatalf("expected ['active', 'pending'], got [%v, %v]", vals[0], vals[1])
	}

	// Compare single param vs repeated: MUST be distinct!
	reqSingle, _ := ParseLosslessHTTPRequest("GET", "/orders?status=active", nil, nil)
	eqRep, verdictRep, _ := CompareLosslessRequests(reqSingle, reqRepeated)
	if eqRep || verdictRep != "MISMATCH" {
		t.Fatalf("expected MISMATCH comparing single to repeated query parameters")
	}
}

func TestX21_HeaderRepresentationFidelity(t *testing.T) {
	// 1. Absent header
	reqAbsentHeader, err := ParseLosslessHTTPRequest("GET", "/items", []EvaluationHeader{
		{Name: "Accept", Value: "application/json"},
	}, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if reqAbsentHeader.HasHeader("X-Custom") {
		t.Fatalf("expected HasHeader('X-Custom')=false")
	}
	val, ok := reqAbsentHeader.GetHeader("X-Custom")
	if ok || val != "" {
		t.Fatalf("expected GetHeader to return ('', false), got (%q, %v)", val, ok)
	}

	// 2. Blank header (Value: "")
	reqBlankHeader, err := ParseLosslessHTTPRequest("GET", "/items", []EvaluationHeader{
		{Name: "Accept", Value: "application/json"},
		{Name: "X-Custom", Value: ""},
	}, nil)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !reqBlankHeader.HasHeader("X-Custom") {
		t.Fatalf("expected HasHeader('X-Custom')=true for blank header")
	}
	valBlank, okBlank := reqBlankHeader.GetHeader("X-Custom")
	if !okBlank || valBlank != "" {
		t.Fatalf("expected GetHeader to return ('', true) for blank header, got (%q, %v)", valBlank, okBlank)
	}

	// Compare absent vs blank header: MUST be distinct!
	eq, verdict, diffs := CompareLosslessRequests(reqAbsentHeader, reqBlankHeader)
	if eq || verdict != "MISMATCH" {
		t.Fatalf("expected MISMATCH comparing absent header to blank header, got %s", verdict)
	}
	if len(diffs) == 0 {
		t.Fatalf("expected diff explaining header presence mismatch")
	}
}

func TestX21_JSONFieldAbsenceVsNullFidelity(t *testing.T) {
	// Payload 1: field "note" is null
	rawWithNull := []byte(`{"id": 1, "note": null}`)
	objWithNull, err := ParseLosslessJSON(rawWithNull)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	valNull, exists1 := ResolveLosslessField(objWithNull, "note")
	if !exists1 {
		t.Fatalf("expected field 'note' to exist")
	}
	if valNull.Type != ValueTypeNull {
		t.Fatalf("expected type null for 'note', got %q", valNull.Type)
	}

	// Payload 2: field "note" is absent
	rawAbsent := []byte(`{"id": 1}`)
	objAbsent, err := ParseLosslessJSON(rawAbsent)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	valAbsent, exists2 := ResolveLosslessField(objAbsent, "note")
	if exists2 {
		t.Fatalf("expected field 'note' to NOT exist in absent object")
	}
	if valAbsent.Type != ValueTypeMissing || valAbsent.ReasonCode != "ERR_FIELD_MISSING" {
		t.Fatalf("expected missing ERR_FIELD_MISSING, got %v", valAbsent)
	}

	// Comparing object with null field vs object with absent field: MUST be MISMATCH!
	eq, triState, _ := CompareTaggedValues(objWithNull, objAbsent)
	if eq || triState != "MISMATCH" {
		t.Fatalf("expected MISMATCH comparing object with null vs object with absent field, got triState=%s eq=%v", triState, eq)
	}
}

func TestX21_EvaluationInputParity(t *testing.T) {
	req, err := ParseLosslessHTTPRequest("POST", "/api/v1/charge?currency=USD&currency=EUR", []EvaluationHeader{
		{Name: "Content-Type", Value: "application/json"},
		{Name: "X-Trace-ID", Value: ""},
	}, []byte(`{"amount": 9007199254740993, "decimal_amount": 9007199254740993.0}`))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	evalInput := req.ToEvaluationInput()
	if evalInput.Method != "POST" {
		t.Errorf("expected POST, got %s", evalInput.Method)
	}
	if evalInput.RawTarget != "/api/v1/charge?currency=USD&currency=EUR" {
		t.Errorf("expected raw target preserved, got %s", evalInput.RawTarget)
	}
	if len(evalInput.Query) != 2 {
		t.Errorf("expected 2 query pairs, got %d", len(evalInput.Query))
	}
	if len(evalInput.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(evalInput.Headers))
	}

	// Verify large integer and decimal in body
	amt, ok := ResolveLosslessField(evalInput.Body, "amount")
	if !ok || amt.Type != ValueTypeInteger || amt.Value != "9007199254740993" {
		t.Errorf("expected large integer 9007199254740993, got %v", amt)
	}

	decAmt, ok2 := ResolveLosslessField(evalInput.Body, "decimal_amount")
	if !ok2 || decAmt.Type != ValueTypeDecimal || decAmt.Value != "9007199254740993" {
		t.Errorf("expected decimal 9007199254740993, got %v", decAmt)
	}
}

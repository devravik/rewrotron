// Package domain contains framework-neutral records and deterministic primitives.
package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// LosslessHTTPRequest models an incoming HTTP request preserving exact lexical
// and wire-level distinctions (EVALUATION-CONTEXT B3 / X21).
type LosslessHTTPRequest struct {
	Method        string                `json:"method"`
	RawTarget     string                `json:"raw_target"`
	Path          string                `json:"path"`
	HasQuery      bool                  `json:"has_query"`
	RawQuery      string                `json:"raw_query"`
	Query         []EvaluationQueryPair `json:"query"`
	Headers       []EvaluationHeader    `json:"headers"`
	RawBody       []byte                `json:"-"`
	BodyRef       *string               `json:"body_ref"`
	ProjectedBody TaggedValue           `json:"projected_body"`
}

// ParseLosslessHTTPRequest parses an incoming HTTP request into a LosslessHTTPRequest,
// preserving empty query, parameter without value, empty parameter value, repeated query keys,
// blank headers, and lexical JSON body representations.
func ParseLosslessHTTPRequest(method, rawTarget string, headers []EvaluationHeader, rawBody []byte) (*LosslessHTTPRequest, error) {
	if method == "" {
		return nil, errors.New("method cannot be empty")
	}
	if rawTarget == "" || !strings.HasPrefix(rawTarget, "/") {
		return nil, errors.New("raw_target must begin with '/'")
	}

	req := &LosslessHTTPRequest{
		Method:    method,
		RawTarget: rawTarget,
		Headers:   make([]EvaluationHeader, len(headers)),
	}
	copy(req.Headers, headers)

	// Split rawTarget into path and query
	qIdx := strings.Index(rawTarget, "?")
	if qIdx == -1 {
		req.Path = rawTarget
		req.HasQuery = false
		req.RawQuery = ""
		req.Query = nil
	} else {
		req.Path = rawTarget[:qIdx]
		req.HasQuery = true
		req.RawQuery = rawTarget[qIdx+1:]
		req.Query = parseQueryStringLossless(req.RawQuery)
	}

	// Process raw body
	if len(rawBody) > 0 {
		sum := sha256.Sum256(rawBody)
		ref := "sha256:" + hex.EncodeToString(sum[:])
		req.BodyRef = &ref
		req.RawBody = make([]byte, len(rawBody))
		copy(req.RawBody, rawBody)

		projected, err := ParseLosslessJSON(rawBody)
		if err == nil {
			req.ProjectedBody = projected
		} else {
			req.ProjectedBody = NewStringLike(ValueTypeBytes, hex.EncodeToString(rawBody))
		}
	} else {
		req.BodyRef = nil
		req.ProjectedBody = NewMissing("ERR_EMPTY_BODY")
	}

	return req, nil
}

// parseQueryStringLossless parses a query string preserving key order, empty queries,
// parameters without values (value is nil), parameters with empty values (value is &""),
// and repeated keys.
func parseQueryStringLossless(rawQuery string) []EvaluationQueryPair {
	if rawQuery == "" {
		return []EvaluationQueryPair{}
	}

	parts := strings.Split(rawQuery, "&")
	pairs := make([]EvaluationQueryPair, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			// empty segment e.g. "a=1&&b=2"
			continue
		}
		eqIdx := strings.Index(part, "=")
		if eqIdx == -1 {
			// Parameter without value e.g. "?flag"
			key := unescapeQueryComponent(part)
			pairs = append(pairs, EvaluationQueryPair{
				Key:   key,
				Value: nil,
			})
		} else {
			key := unescapeQueryComponent(part[:eqIdx])
			valStr := unescapeQueryComponent(part[eqIdx+1:])
			pairs = append(pairs, EvaluationQueryPair{
				Key:   key,
				Value: &valStr,
			})
		}
	}

	return pairs
}

func unescapeQueryComponent(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}

// ToEvaluationInput converts LosslessHTTPRequest to the canonical EvaluationInput model.
func (r *LosslessHTTPRequest) ToEvaluationInput() EvaluationInput {
	var bodyRef string
	if r.BodyRef != nil {
		bodyRef = *r.BodyRef
	}
	return EvaluationInput{
		RequestRef: bodyRef,
		Method:     r.Method,
		RawTarget:  r.RawTarget,
		Headers:    r.Headers,
		Query:      r.Query,
		Body:       r.ProjectedBody,
	}
}

// HasQueryParam reports whether the query parameter key is present.
func (r *LosslessHTTPRequest) HasQueryParam(key string) bool {
	for _, q := range r.Query {
		if q.Key == key {
			return true
		}
	}
	return false
}

// GetQueryParam returns the first occurrence of the query parameter key.
func (r *LosslessHTTPRequest) GetQueryParam(key string) (*string, bool) {
	for _, q := range r.Query {
		if q.Key == key {
			return q.Value, true
		}
	}
	return nil, false
}

// GetAllQueryParams returns all values for a query parameter key in order.
func (r *LosslessHTTPRequest) GetAllQueryParams(key string) []*string {
	var values []*string
	for _, q := range r.Query {
		if q.Key == key {
			values = append(values, q.Value)
		}
	}
	return values
}

// HasHeader reports whether the header name exists (case-insensitive).
func (r *LosslessHTTPRequest) HasHeader(name string) bool {
	_, ok := r.GetHeader(name)
	return ok
}

// GetHeader returns the header value (case-insensitive).
// Returns (value, true) if present (even if value is blank ""), or ("", false) if absent.
func (r *LosslessHTTPRequest) GetHeader(name string) (string, bool) {
	for _, h := range r.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value, true
		}
	}
	return "", false
}

// CompareLosslessRequests compares two lossless HTTP requests and reports any divergence
// across method, path, query presence/content, headers, and body.
func CompareLosslessRequests(left, right *LosslessHTTPRequest) (bool, string, []string) {
	if left == nil || right == nil {
		if left == right {
			return true, "PASS", nil
		}
		return false, "MISMATCH", []string{"nil request comparison"}
	}

	var diffs []string

	if left.Method != right.Method {
		diffs = append(diffs, fmt.Sprintf("method mismatch: left=%q, right=%q", left.Method, right.Method))
	}
	if left.Path != right.Path {
		diffs = append(diffs, fmt.Sprintf("path mismatch: left=%q, right=%q", left.Path, right.Path))
	}
	if left.HasQuery != right.HasQuery {
		diffs = append(diffs, fmt.Sprintf("query presence mismatch: left has_query=%v, right has_query=%v", left.HasQuery, right.HasQuery))
	}
	if left.RawQuery != right.RawQuery {
		diffs = append(diffs, fmt.Sprintf("raw query mismatch: left=%q, right=%q", left.RawQuery, right.RawQuery))
	}

	// Compare query pairs
	if len(left.Query) != len(right.Query) {
		diffs = append(diffs, fmt.Sprintf("query length mismatch: left=%d, right=%d", len(left.Query), len(right.Query)))
	} else {
		for i := range left.Query {
			lq := left.Query[i]
			rq := right.Query[i]
			if lq.Key != rq.Key {
				diffs = append(diffs, fmt.Sprintf("query[%d] key mismatch: left=%q, right=%q", i, lq.Key, rq.Key))
			} else if (lq.Value == nil && rq.Value != nil) || (lq.Value != nil && rq.Value == nil) {
				diffs = append(diffs, fmt.Sprintf("query[%d] value nilness mismatch for key %q", i, lq.Key))
			} else if lq.Value != nil && rq.Value != nil && *lq.Value != *rq.Value {
				diffs = append(diffs, fmt.Sprintf("query[%d] value mismatch for key %q: left=%q, right=%q", i, lq.Key, *lq.Value, *rq.Value))
			}
		}
	}

	// Compare headers
	if len(left.Headers) != len(right.Headers) {
		diffs = append(diffs, fmt.Sprintf("header count mismatch: left=%d, right=%d", len(left.Headers), len(right.Headers)))
	}
	for _, lh := range left.Headers {
		rVal, exists := right.GetHeader(lh.Name)
		if !exists {
			diffs = append(diffs, fmt.Sprintf("header %q missing in right", lh.Name))
		} else if lh.Value != rVal {
			diffs = append(diffs, fmt.Sprintf("header %q value mismatch: left=%q, right=%q", lh.Name, lh.Value, rVal))
		}
	}
	for _, rh := range right.Headers {
		if !left.HasHeader(rh.Name) {
			diffs = append(diffs, fmt.Sprintf("header %q missing in left", rh.Name))
		}
	}

	// Compare projected body
	eq, triState, err := CompareTaggedValues(left.ProjectedBody, right.ProjectedBody)
	if err != nil {
		diffs = append(diffs, fmt.Sprintf("body comparison error: %v", err))
	} else if !eq || triState != "PASS" {
		diffs = append(diffs, fmt.Sprintf("body comparison mismatch: triState=%s", triState))
	}

	if len(diffs) > 0 {
		return false, "MISMATCH", diffs
	}
	return true, "PASS", nil
}

// ParseLosslessJSON parses raw JSON bytes into a TaggedValue preserving lexical number types (B3),
// distinguishing object key null from key absence, and rejecting duplicate keys.
func ParseLosslessJSON(raw []byte) (TaggedValue, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return NewMissing("ERR_EMPTY_PAYLOAD"), nil
	}

	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()

	val, err := parseLosslessJSONValue(dec)
	if err != nil {
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), err
	}

	if _, err := dec.Token(); err != io.EOF {
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), errors.New("trailing data in JSON payload")
	}

	return val, nil
}

func parseLosslessJSONValue(dec *json.Decoder) (TaggedValue, error) {
	tok, err := dec.Token()
	if err != nil {
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), err
	}

	switch v := tok.(type) {
	case nil:
		return NewNull(), nil
	case bool:
		return NewBoolean(v), nil
	case string:
		return NewString(v), nil
	case json.Number:
		rawToken := string(v)
		projected := ProjectRawJSONNumber(rawToken)
		return projected, nil
	case json.Delim:
		switch v {
		case '{':
			obj := make(map[string]TaggedValue)
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), err
				}
				keyStr, ok := keyTok.(string)
				if !ok {
					return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), errors.New("expected string object key")
				}
				if _, exists := obj[keyStr]; exists {
					return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), fmt.Errorf("duplicate object key %q: ERR_DUPLICATE_KEY", keyStr)
				}
				val, err := parseLosslessJSONValue(dec)
				if err != nil {
					return val, err
				}
				obj[keyStr] = val
			}
			endTok, err := dec.Token()
			if err != nil || endTok != json.Delim('}') {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), errors.New("expected closing '}'")
			}
			return NewObject(obj), nil
		case '[':
			var arr []TaggedValue
			for dec.More() {
				val, err := parseLosslessJSONValue(dec)
				if err != nil {
					return val, err
				}
				arr = append(arr, val)
			}
			endTok, err := dec.Token()
			if err != nil || endTok != json.Delim(']') {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), errors.New("expected closing ']'")
			}
			return NewArray(arr), nil
		default:
			return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), fmt.Errorf("unexpected delimiter %v", v)
		}
	default:
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), fmt.Errorf("unexpected token type %T", tok)
	}
}

// ResolveLosslessField extracts a field from a TaggedValue object, strictly distinguishing
// field absence from a field present with null.
// - If obj is not an object: returns MISSING ("ERR_NOT_AN_OBJECT"), false
// - If key is absent: returns MISSING ("ERR_FIELD_MISSING"), false
// - If key is present with null: returns TaggedValue{Type: ValueTypeNull, Value: nil}, true
// - If key is present with value: returns the value, true
func ResolveLosslessField(obj TaggedValue, key string) (TaggedValue, bool) {
	if obj.Type != ValueTypeObject {
		return NewMissing("ERR_NOT_AN_OBJECT"), false
	}
	m, ok := obj.Value.(map[string]TaggedValue)
	if !ok {
		return NewMissing("ERR_INVALID_OBJECT_INTERNAL"), false
	}
	val, exists := m[key]
	if !exists {
		return NewMissing("ERR_FIELD_MISSING"), false
	}
	return val, true
}

// CoerceIntegerToDecimal implements the explicit INTEGER_TO_DECIMAL operator required by EVALUATION-CONTEXT.
// Implicit cross-type coercion is forbidden (integer "1" vs decimal "1" is UNKNOWN / ERR_TYPE).
// Only this explicit operator converts an integer tagged value into a decimal tagged value.
func CoerceIntegerToDecimal(val TaggedValue) (TaggedValue, error) {
	if val.Type != ValueTypeInteger {
		return val, fmt.Errorf("cannot coerce non-integer type %q to decimal", val.Type)
	}
	intStr, ok := val.Value.(string)
	if !ok {
		return val, errors.New("integer tagged value must be string")
	}
	return NewStringLike(ValueTypeDecimal, intStr), nil
}

// ContainsMissingCapability walks a TaggedValue tree and reports whether any node
// is MISSING with reason ERR_UNSUPPORTED_CAPABILITY.
func ContainsMissingCapability(val TaggedValue) bool {
	if val.Type == ValueTypeMissing && val.ReasonCode == "ERR_UNSUPPORTED_CAPABILITY" {
		return true
	}
	switch val.Type {
	case ValueTypeObject:
		if m, ok := val.Value.(map[string]TaggedValue); ok {
			for _, v := range m {
				if ContainsMissingCapability(v) {
					return true
				}
			}
		}
	case ValueTypeArray:
		if arr, ok := val.Value.([]TaggedValue); ok {
			for _, v := range arr {
				if ContainsMissingCapability(v) {
					return true
				}
			}
		}
	}
	return false
}

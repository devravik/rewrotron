package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Allowed evaluation context constants per specs/behavior/evaluation-context.schema.json
const (
	SideSource = "SOURCE"
	SideTarget = "TARGET"
	SideShadow = "SHADOW"
)

const (
	PhaseFixture       = "FIXTURE"
	PhasePreOperation  = "PRE_OPERATION"
	PhasePostOperation = "POST_OPERATION"
	PhaseCompletion    = "COMPLETION"
	PhaseSimulation    = "SIMULATION"
)

const (
	OriginFixturePrepared = "FIXTURE_PREPARED"
	OriginModelExpected   = "MODEL_EXPECTED"
	OriginActualObserved  = "ACTUAL_OBSERVED"
)

const (
	ValueTypeString    = "string"
	ValueTypeBoolean   = "boolean"
	ValueTypeInteger   = "integer"
	ValueTypeDecimal   = "decimal"
	ValueTypeTimestamp = "timestamp"
	ValueTypeBytes     = "bytes"
	ValueTypeNull      = "null"
	ValueTypeObject    = "object"
	ValueTypeArray     = "array"
	ValueTypeMissing   = "missing"
)

var allowedSides = map[string]struct{}{
	SideSource: {},
	SideTarget: {},
	SideShadow: {},
}

var allowedPhases = map[string]struct{}{
	PhaseFixture:       {},
	PhasePreOperation:  {},
	PhasePostOperation: {},
	PhaseCompletion:    {},
	PhaseSimulation:    {},
}

var allowedOrigins = map[string]struct{}{
	OriginFixturePrepared: {},
	OriginModelExpected:   {},
	OriginActualObserved:  {},
}

var allowedValueTypes = map[string]struct{}{
	ValueTypeString:    {},
	ValueTypeBoolean:   {},
	ValueTypeInteger:   {},
	ValueTypeDecimal:   {},
	ValueTypeTimestamp: {},
	ValueTypeBytes:     {},
	ValueTypeNull:      {},
	ValueTypeObject:    {},
	ValueTypeArray:     {},
	ValueTypeMissing:   {},
}

var (
	contextIDPattern  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	projectIDPattern  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	occurrencePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
	digestPattern     = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
	headerNamePattern = regexp.MustCompile(`^[!#$%&'*+.^_` + "`" + `|~0-9A-Za-z-]+$`)
	canonicalIntRegex = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)
	canonicalDecRegex = regexp.MustCompile(`^(?:0|-?[1-9][0-9]*|-?(?:0|[1-9][0-9]*)\.[0-9]*[1-9])$`)
	jsonNumberRegex   = regexp.MustCompile(`^(-?)(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
)

// EvaluationContextError describes an invalid evaluation context record.
type EvaluationContextError struct {
	Reason string
	Cause  error
}

func (e *EvaluationContextError) Error() string {
	if e.Cause != nil {
		return "invalid evaluation context: " + e.Reason + ": " + e.Cause.Error()
	}
	return "invalid evaluation context: " + e.Reason
}

func (e *EvaluationContextError) Unwrap() error { return e.Cause }

// TaggedValueError describes an invalid tagged value record.
type TaggedValueError struct {
	Reason string
	Cause  error
}

func (e *TaggedValueError) Error() string {
	return "invalid tagged value: " + e.Reason
}

func (e *TaggedValueError) Unwrap() error { return e.Cause }

// ResponseError describes an invalid response record.
type ResponseError struct {
	Reason string
	Cause  error
}

func (e *ResponseError) Error() string {
	return "invalid response: " + e.Reason
}

func (e *ResponseError) Unwrap() error { return e.Cause }

// TaggedValue represents a closed tagged value conforming to specs/behavior/evaluation-context.schema.json#/$defs/value.
type TaggedValue struct {
	Type       string `json:"type"`
	Value      any    `json:"value,omitempty"`
	ReasonCode string `json:"reason_code,omitempty"`
}

// NewString creates a tagged string value.
func NewString(v string) TaggedValue {
	return TaggedValue{Type: ValueTypeString, Value: v}
}

// NewBoolean creates a tagged boolean value.
func NewBoolean(v bool) TaggedValue {
	return TaggedValue{Type: ValueTypeBoolean, Value: v}
}

// NewInteger creates a tagged integer value after checking canonical integer spelling.
func NewInteger(v string) (TaggedValue, error) {
	if !canonicalIntRegex.MatchString(v) {
		return TaggedValue{}, &TaggedValueError{Reason: "integer value must match canonical integer pattern"}
	}
	return TaggedValue{Type: ValueTypeInteger, Value: v}, nil
}

// NewDecimal creates a tagged decimal value after checking canonical decimal spelling and bounds.
func NewDecimal(v string) (TaggedValue, error) {
	if !canonicalDecRegex.MatchString(v) {
		return TaggedValue{}, &TaggedValueError{Reason: "decimal value must match canonical decimal pattern"}
	}
	coeff, scale := decimalMetrics(v)
	if coeff > 38 || scale > 18 {
		return TaggedValue{}, &TaggedValueError{Reason: "decimal bounds exceeded (max 38 coefficient, 18 scale)"}
	}
	return TaggedValue{Type: ValueTypeDecimal, Value: v}, nil
}

// NewTimestamp creates a tagged timestamp value from an RFC 3339 string.
func NewTimestamp(v string) (TaggedValue, error) {
	if _, err := time.Parse(time.RFC3339, v); err != nil {
		return TaggedValue{}, &TaggedValueError{Reason: "timestamp must be RFC 3339 format", Cause: err}
	}
	return TaggedValue{Type: ValueTypeTimestamp, Value: v}, nil
}

// NewBytes creates a tagged bytes value referencing a sha256 digest.
func NewBytes(digest string) (TaggedValue, error) {
	if !digestPattern.MatchString(digest) {
		return TaggedValue{}, &TaggedValueError{Reason: "bytes reference must be sha256 digest"}
	}
	return TaggedValue{Type: ValueTypeBytes, Value: digest}, nil
}

// NewNull creates a tagged null value.
func NewNull() TaggedValue {
	return TaggedValue{Type: ValueTypeNull, Value: nil}
}

// NewObject creates a tagged object value.
func NewObject(v map[string]TaggedValue) TaggedValue {
	return TaggedValue{Type: ValueTypeObject, Value: v}
}

// NewArray creates a tagged array value.
func NewArray(v []TaggedValue) TaggedValue {
	return TaggedValue{Type: ValueTypeArray, Value: v}
}

// NewMissing creates a tagged missing value with a reason code.
func NewMissing(reasonCode string) TaggedValue {
	return TaggedValue{Type: ValueTypeMissing, ReasonCode: reasonCode}
}

// MarshalJSON marshals TaggedValue conforming strictly to schema requirements.
func (tv TaggedValue) MarshalJSON() ([]byte, error) {
	if tv.Type == ValueTypeMissing {
		return json.Marshal(map[string]string{
			"type":        ValueTypeMissing,
			"reason_code": tv.ReasonCode,
		})
	}
	return json.Marshal(map[string]any{
		"type":  tv.Type,
		"value": tv.Value,
	})
}

// UnmarshalJSON unmarshals and validates a TaggedValue.
func (tv *TaggedValue) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return &TaggedValueError{Reason: "not a valid JSON object", Cause: err}
	}
	rawType, ok := raw["type"]
	if !ok {
		return &TaggedValueError{Reason: "missing required 'type' field"}
	}
	var typeStr string
	if err := json.Unmarshal(rawType, &typeStr); err != nil {
		return &TaggedValueError{Reason: "'type' must be string", Cause: err}
	}
	if _, allowed := allowedValueTypes[typeStr]; !allowed {
		return &TaggedValueError{Reason: fmt.Sprintf("unsupported tagged value type %q", typeStr)}
	}
	tv.Type = typeStr

	if typeStr == ValueTypeMissing {
		if _, hasVal := raw["value"]; hasVal {
			return &TaggedValueError{Reason: "missing value must not contain 'value' field"}
		}
		rawReason, ok := raw["reason_code"]
		if !ok {
			return &TaggedValueError{Reason: "missing required 'reason_code' field"}
		}
		if err := json.Unmarshal(rawReason, &tv.ReasonCode); err != nil {
			return &TaggedValueError{Reason: "'reason_code' must be string", Cause: err}
		}
		if tv.ReasonCode == "" {
			return &TaggedValueError{Reason: "'reason_code' must be non-empty"}
		}
		if len(raw) != 2 {
			return &TaggedValueError{Reason: "missing value has unrecognized fields"}
		}
		return nil
	}

	if _, hasReason := raw["reason_code"]; hasReason {
		return &TaggedValueError{Reason: "non-missing tagged value must not contain 'reason_code' field"}
	}
	rawValue, ok := raw["value"]
	if !ok {
		return &TaggedValueError{Reason: "missing required 'value' field"}
	}
	if len(raw) != 2 {
		return &TaggedValueError{Reason: "tagged value has unrecognized fields"}
	}

	switch typeStr {
	case ValueTypeString:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &TaggedValueError{Reason: "string value must be a JSON string", Cause: err}
		}
		tv.Value = s
	case ValueTypeBoolean:
		var b bool
		if err := json.Unmarshal(rawValue, &b); err != nil {
			return &TaggedValueError{Reason: "boolean value must be a JSON boolean", Cause: err}
		}
		tv.Value = b
	case ValueTypeInteger:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &TaggedValueError{Reason: "integer value must be a string (native numbers rejected)", Cause: err}
		}
		if !canonicalIntRegex.MatchString(s) {
			return &TaggedValueError{Reason: "integer value does not match canonical integer pattern"}
		}
		tv.Value = s
	case ValueTypeDecimal:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &TaggedValueError{Reason: "decimal value must be a string (native numbers rejected)", Cause: err}
		}
		if !canonicalDecRegex.MatchString(s) {
			return &TaggedValueError{Reason: "decimal value does not match canonical decimal pattern"}
		}
		coeff, scale := decimalMetrics(s)
		if coeff > 38 || scale > 18 {
			return &TaggedValueError{Reason: "decimal bounds exceeded (max 38 coefficient, 18 scale)"}
		}
		tv.Value = s
	case ValueTypeTimestamp:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &TaggedValueError{Reason: "timestamp value must be a string", Cause: err}
		}
		if _, err := time.Parse(time.RFC3339, s); err != nil {
			return &TaggedValueError{Reason: "timestamp value must be RFC 3339 format", Cause: err}
		}
		tv.Value = s
	case ValueTypeBytes:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &TaggedValueError{Reason: "bytes value must be a string", Cause: err}
		}
		if !digestPattern.MatchString(s) {
			return &TaggedValueError{Reason: "bytes value must be sha256 digest"}
		}
		tv.Value = s
	case ValueTypeNull:
		if string(bytes.TrimSpace(rawValue)) != "null" {
			return &TaggedValueError{Reason: "null value must be null"}
		}
		tv.Value = nil
	case ValueTypeObject:
		var objMap map[string]TaggedValue
		if err := json.Unmarshal(rawValue, &objMap); err != nil {
			return &TaggedValueError{Reason: "object value must be an object of TaggedValues", Cause: err}
		}
		tv.Value = objMap
	case ValueTypeArray:
		var arr []TaggedValue
		if err := json.Unmarshal(rawValue, &arr); err != nil {
			return &TaggedValueError{Reason: "array value must be an array of TaggedValues", Cause: err}
		}
		tv.Value = arr
	}
	return nil
}

// decimalMetrics computes coefficient digits (ignoring sign, dot, leading zeros) and scale (fractional digits).
func decimalMetrics(s string) (coeff int, scale int) {
	raw := strings.TrimPrefix(s, "-")
	dotIdx := strings.IndexByte(raw, '.')
	if dotIdx >= 0 {
		scale = len(raw) - dotIdx - 1
		intPart := raw[:dotIdx]
		fracPart := raw[dotIdx+1:]
		digits := strings.TrimLeft(intPart+fracPart, "0")
		if digits == "" {
			coeff = 1
		} else {
			coeff = len(digits)
		}
	} else {
		scale = 0
		digits := strings.TrimLeft(raw, "0")
		if digits == "" {
			coeff = 1
		} else {
			coeff = len(digits)
		}
	}
	return coeff, scale
}

// ProjectRawJSONNumber projects a raw JSON number token according to EVALUATION-CONTEXT B3 / X21 / X63 rules.
// If the token has neither '.' nor 'e'/'E', it projects as integer.
// If it contains '.' or 'e'/'E', it projects as canonical decimal.
// Invalid JSON number syntax or exceeding 38 coefficient / 18 scale bounds yields MISSING with ERR_UNSUPPORTED_CAPABILITY.
func ProjectRawJSONNumber(rawToken string) TaggedValue {
	m := jsonNumberRegex.FindStringSubmatch(rawToken)
	if m == nil {
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
	}

	sign := m[1]
	intPart := m[2]
	fracWithDot := m[3]
	expWithPrefix := m[4]

	hasDot := fracWithDot != ""
	hasExp := expWithPrefix != ""

	if !hasDot && !hasExp {
		// Integer token
		if sign == "-" && intPart == "0" {
			return NewStringLike(ValueTypeInteger, "0")
		}
		return NewStringLike(ValueTypeInteger, sign+intPart)
	}

	// Decimal token
	var fracPart string
	if hasDot {
		fracPart = fracWithDot[1:]
	}

	expVal := 0
	if hasExp {
		parsedExp, err := strconv.Atoi(expWithPrefix[1:])
		if err != nil {
			return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
		}
		expVal = parsedExp
	}

	// Check if all digits are zero (e.g. "0.0", "-0.0", "0e10")
	if intPart == "0" && (fracPart == "" || strings.Trim(fracPart, "0") == "") {
		return NewStringLike(ValueTypeDecimal, "0")
	}

	allDigits := strings.TrimLeft(intPart+fracPart, "0")
	if allDigits == "" {
		return NewStringLike(ValueTypeDecimal, "0")
	}

	shift := expVal - len(fracPart)
	if shift >= 0 {
		coeffLen := len(allDigits) + shift
		if coeffLen > 38 {
			return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
		}
		resultStr := sign + allDigits + strings.Repeat("0", shift)
		return NewStringLike(ValueTypeDecimal, resultStr)
	}

	absShift := -shift
	if len(allDigits) <= absShift {
		pad := absShift - len(allDigits)
		fullFrac := strings.Repeat("0", pad) + allDigits
		strippedFrac := strings.TrimRight(fullFrac, "0")
		scale := len(strippedFrac)
		if scale > 18 {
			return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
		}
		coeffLen := len(strings.TrimRight(allDigits, "0"))
		if coeffLen == 0 {
			coeffLen = 1
		}
		if coeffLen > 38 {
			return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
		}
		resultStr := sign + "0." + strippedFrac
		return NewStringLike(ValueTypeDecimal, resultStr)
	}

	splitIdx := len(allDigits) - absShift
	iPart := allDigits[:splitIdx]
	fPart := strings.TrimRight(allDigits[splitIdx:], "0")
	scale := len(fPart)
	if scale > 18 {
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
	}
	coeffLen := len(iPart) + len(fPart)
	if coeffLen > 38 {
		return NewMissing("ERR_UNSUPPORTED_CAPABILITY")
	}
	var resultStr string
	if fPart == "" {
		resultStr = sign + iPart
	} else {
		resultStr = sign + iPart + "." + fPart
	}
	return NewStringLike(ValueTypeDecimal, resultStr)
}

func NewStringLike(typeStr string, val string) TaggedValue {
	return TaggedValue{Type: typeStr, Value: val}
}

// CompareTaggedValues compares two tagged values under strict type isolation (B3 / X21 / numeric_comparison).
// Implicit cross-type coercion is forbidden: integer vs decimal yields UNKNOWN ("ERR_TYPE").
func CompareTaggedValues(left, right TaggedValue) (equal bool, triState string, err error) {
	if left.Type == ValueTypeMissing || right.Type == ValueTypeMissing {
		return false, "UNKNOWN", nil
	}
	if left.Type != right.Type {
		return false, "UNKNOWN", nil
	}

	switch left.Type {
	case ValueTypeNull:
		return true, "PASS", nil
	case ValueTypeString, ValueTypeBoolean, ValueTypeInteger, ValueTypeDecimal, ValueTypeTimestamp, ValueTypeBytes:
		if left.Value == right.Value {
			return true, "PASS", nil
		}
		return false, "MISMATCH", nil
	case ValueTypeObject:
		leftMap, ok1 := left.Value.(map[string]TaggedValue)
		rightMap, ok2 := right.Value.(map[string]TaggedValue)
		if !ok1 || !ok2 {
			return false, "UNKNOWN", errors.New("invalid internal object representation")
		}
		if len(leftMap) != len(rightMap) {
			return false, "MISMATCH", nil
		}
		for k, lv := range leftMap {
			rv, ok := rightMap[k]
			if !ok {
				return false, "MISMATCH", nil
			}
			eq, verdict, err := CompareTaggedValues(lv, rv)
			if err != nil || !eq {
				return false, verdict, err
			}
		}
		return true, "PASS", nil
	case ValueTypeArray:
		leftArr, ok1 := left.Value.([]TaggedValue)
		rightArr, ok2 := right.Value.([]TaggedValue)
		if !ok1 || !ok2 {
			return false, "UNKNOWN", errors.New("invalid internal array representation")
		}
		if len(leftArr) != len(rightArr) {
			return false, "MISMATCH", nil
		}
		for i := range leftArr {
			eq, verdict, err := CompareTaggedValues(leftArr[i], rightArr[i])
			if err != nil || !eq {
				return false, verdict, err
			}
		}
		return true, "PASS", nil
	default:
		return false, "UNKNOWN", fmt.Errorf("unsupported type %q", left.Type)
	}
}

// EvaluationContext represents an immutable evaluation context snapshot.
type EvaluationContext struct {
	SchemaVersion string                      `json:"schema_version"`
	ContextID     string                      `json:"context_id"`
	ProjectID     string                      `json:"project_id"`
	Side          string                      `json:"side"`
	OccurrenceID  *string                     `json:"occurrence_id"`
	ProjectionRef string                      `json:"projection_ref"`
	Phase         string                      `json:"phase"`
	Origin        string                      `json:"origin"`
	EvidenceRefs  []string                    `json:"evidence_refs"`
	Namespaces    EvaluationContextNamespaces `json:"namespaces"`
}

// EvaluationContextNamespaces holds the seven evaluation context namespaces.
type EvaluationContextNamespaces struct {
	Input   *EvaluationInput            `json:"input"`
	Actor   EvaluationActor             `json:"actor"`
	State   EvaluationState             `json:"state"`
	Result  *EvaluationResponse         `json:"result"`
	Binding map[string]TaggedValue      `json:"binding"`
	Row     map[string]TaggedValue      `json:"row"`
	Clock   TaggedValue                 `json:"clock"`
}

// EvaluationInput represents request input in the context.
type EvaluationInput struct {
	RequestRef string                `json:"request_ref"`
	Method     string                `json:"method"`
	RawTarget  string                `json:"raw_target"`
	Headers    []EvaluationHeader    `json:"headers"`
	Query      []EvaluationQueryPair `json:"query"`
	Body       TaggedValue           `json:"body"`
}

// EvaluationHeader represents a name-value header.
type EvaluationHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EvaluationQueryPair represents a query parameter key and optional value.
type EvaluationQueryPair struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

// EvaluationActor represents the actor namespace.
type EvaluationActor struct {
	FixtureActorRef string          `json:"fixture_actor_ref"`
	ActualPrincipal ActualPrincipal `json:"actual_principal"`
}

// ActualPrincipal represents actual observed principal or missing.
type ActualPrincipal struct {
	Missing           *MissingValue
	Authentication    string
	SubjectLocalID    *string
	SessionLocalID    *string
	Roles             []string
	CaptureChannelRef string
	EvidenceRefs      []string
}

// MarshalJSON marshals ActualPrincipal.
func (ap ActualPrincipal) MarshalJSON() ([]byte, error) {
	if ap.Missing != nil {
		return json.Marshal(map[string]string{
			"type":        ValueTypeMissing,
			"reason_code": ap.Missing.ReasonCode,
		})
	}
	return json.Marshal(map[string]any{
		"authentication":      ap.Authentication,
		"subject_local_id":    ap.SubjectLocalID,
		"session_local_id":    ap.SessionLocalID,
		"roles":               ap.Roles,
		"capture_channel_ref": ap.CaptureChannelRef,
		"evidence_refs":       ap.EvidenceRefs,
	})
}

// UnmarshalJSON unmarshals ActualPrincipal.
func (ap *ActualPrincipal) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if rawType, ok := raw["type"]; ok {
		var typeStr string
		if err := json.Unmarshal(rawType, &typeStr); err != nil {
			return err
		}
		if typeStr == ValueTypeMissing {
			var reason string
			if err := json.Unmarshal(raw["reason_code"], &reason); err != nil {
				return err
			}
			ap.Missing = &MissingValue{Type: ValueTypeMissing, ReasonCode: reason}
			return nil
		}
	}
	var auth string
	if err := json.Unmarshal(raw["authentication"], &auth); err != nil {
		return err
	}
	ap.Authentication = auth
	if rawSubj, ok := raw["subject_local_id"]; ok && string(bytes.TrimSpace(rawSubj)) != "null" {
		var subj string
		if err := json.Unmarshal(rawSubj, &subj); err != nil {
			return err
		}
		ap.SubjectLocalID = &subj
	}
	if rawSess, ok := raw["session_local_id"]; ok && string(bytes.TrimSpace(rawSess)) != "null" {
		var sess string
		if err := json.Unmarshal(rawSess, &sess); err != nil {
			return err
		}
		ap.SessionLocalID = &sess
	}
	if rawRoles, ok := raw["roles"]; ok {
		if err := json.Unmarshal(rawRoles, &ap.Roles); err != nil {
			return err
		}
	}
	if rawChan, ok := raw["capture_channel_ref"]; ok {
		if err := json.Unmarshal(rawChan, &ap.CaptureChannelRef); err != nil {
			return err
		}
	}
	if rawEv, ok := raw["evidence_refs"]; ok {
		if err := json.Unmarshal(rawEv, &ap.EvidenceRefs); err != nil {
			return err
		}
	}
	return nil
}

// EvaluationState represents the state namespace.
type EvaluationState struct {
	DB       DBTablesOrMissing `json:"db"`
	Sessions TaggedValue       `json:"sessions"`
	Cache    TaggedValue       `json:"cache"`
	Files    TaggedValue       `json:"files"`
	Queues   TaggedValue       `json:"queues"`
}

// DBTablesOrMissing represents db tables or missing.
type DBTablesOrMissing struct {
	Missing *MissingValue
	Tables  []EvaluationDBTable
}

// MarshalJSON marshals DBTablesOrMissing.
func (db DBTablesOrMissing) MarshalJSON() ([]byte, error) {
	if db.Missing != nil {
		return json.Marshal(map[string]string{
			"type":        ValueTypeMissing,
			"reason_code": db.Missing.ReasonCode,
		})
	}
	return json.Marshal(db.Tables)
}

// UnmarshalJSON unmarshals DBTablesOrMissing.
func (db *DBTablesOrMissing) UnmarshalJSON(data []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err == nil {
		if rawType, ok := obj["type"]; ok {
			var typeStr string
			if err := json.Unmarshal(rawType, &typeStr); err == nil && typeStr == ValueTypeMissing {
				var reason string
				if err := json.Unmarshal(obj["reason_code"], &reason); err != nil {
					return err
				}
				db.Missing = &MissingValue{Type: ValueTypeMissing, ReasonCode: reason}
				return nil
			}
		}
	}
	var tables []EvaluationDBTable
	if err := json.Unmarshal(data, &tables); err != nil {
		return err
	}
	db.Tables = tables
	return nil
}

// EvaluationDBTable represents a captured database table state.
type EvaluationDBTable struct {
	ConnectionID string      `json:"connection_id"`
	Catalog      string      `json:"catalog"`
	Schema       *string     `json:"schema"`
	Table        string      `json:"table"`
	Rows         TaggedValue `json:"rows"`
}

// EvaluationResponse represents an HTTP response in the context.
type EvaluationResponse struct {
	SchemaVersion    string             `json:"schema_version"`
	Status           int                `json:"status"`
	Headers          []EvaluationHeader `json:"headers"`
	Cookies          []EvaluationCookie `json:"cookies"`
	Body             TaggedValue        `json:"body"`
	Session          TaggedValue        `json:"session"`
	Principal        TaggedValue        `json:"principal"`
	EffectRefs       []string           `json:"effect_refs"`
	ApplicationError TaggedValue        `json:"application_error"`
}

// EvaluationCookie represents a parsed cookie in the response.
type EvaluationCookie struct {
	Name     string      `json:"name"`
	Value    TaggedValue `json:"value"`
	Domain   *string     `json:"domain"`
	Path     *string     `json:"path"`
	MaxAge   *int        `json:"max_age"`
	Expires  *string     `json:"expires"`
	SameSite *string     `json:"same_site"`
	Secure   bool        `json:"secure"`
	HTTPOnly bool        `json:"http_only"`
}

// MissingValue represents a missing tagged property.
type MissingValue struct {
	Type       string `json:"type"`
	ReasonCode string `json:"reason_code"`
}

// ValidateTaggedValue validates a raw JSON tagged value against specs/behavior/evaluation-context.schema.json#/$defs/value.
func ValidateTaggedValue(raw []byte) error {
	var tv TaggedValue
	if err := json.Unmarshal(raw, &tv); err != nil {
		return err
	}
	return nil
}

// ValidateResponse validates a raw JSON response record against specs/behavior/response.schema.json.
func ValidateResponse(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return &ResponseError{Reason: "not a valid JSON object", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"status",
		"headers",
		"cookies",
		"body",
		"session",
		"principal",
		"effect_refs",
		"application_error",
	}

	for _, f := range expectedFields {
		if _, ok := record[f]; !ok {
			return &ResponseError{Reason: "missing required field " + f}
		}
	}
	if len(record) != len(expectedFields) {
		return &ResponseError{Reason: "response contains unrecognized fields"}
	}

	var schemaVer string
	if err := json.Unmarshal(record["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &ResponseError{Reason: "schema_version must be '1.0'"}
	}

	var status int
	if err := json.Unmarshal(record["status"], &status); err != nil || status < 100 || status > 599 {
		return &ResponseError{Reason: "status must be integer between 100 and 599"}
	}

	// Headers
	var headers []map[string]json.RawMessage
	if err := json.Unmarshal(record["headers"], &headers); err != nil {
		return &ResponseError{Reason: "headers must be an array", Cause: err}
	}
	for _, h := range headers {
		if len(h) != 2 {
			return &ResponseError{Reason: "header contains unrecognized fields"}
		}
		rawName, hasName := h["name"]
		rawValue, hasValue := h["value"]
		if !hasName || !hasValue {
			return &ResponseError{Reason: "header missing required name or value"}
		}
		var name, value string
		if err := json.Unmarshal(rawName, &name); err != nil || !headerNamePattern.MatchString(name) {
			return &ResponseError{Reason: "invalid header name format"}
		}
		if err := json.Unmarshal(rawValue, &value); err != nil || strings.ContainsAny(value, "\r\n") {
			return &ResponseError{Reason: "invalid header value format (CRLF forbidden)"}
		}
	}

	// Cookies
	var cookies []map[string]json.RawMessage
	if err := json.Unmarshal(record["cookies"], &cookies); err != nil {
		return &ResponseError{Reason: "cookies must be an array", Cause: err}
	}
	cookieReq := []string{"name", "value", "domain", "path", "max_age", "expires", "same_site", "secure", "http_only"}
	for _, c := range cookies {
		for _, f := range cookieReq {
			if _, ok := c[f]; !ok {
				return &ResponseError{Reason: "cookie missing required field " + f}
			}
		}
		if len(c) != len(cookieReq) {
			return &ResponseError{Reason: "cookie contains unrecognized fields"}
		}
		var cName string
		if err := json.Unmarshal(c["name"], &cName); err != nil {
			return &ResponseError{Reason: "cookie name must be string"}
		}
		if err := ValidateTaggedValue(c["value"]); err != nil {
			return &ResponseError{Reason: "cookie value must be valid tagged value", Cause: err}
		}
		var secure, httpOnly bool
		if err := json.Unmarshal(c["secure"], &secure); err != nil {
			return &ResponseError{Reason: "cookie secure must be boolean"}
		}
		if err := json.Unmarshal(c["http_only"], &httpOnly); err != nil {
			return &ResponseError{Reason: "cookie http_only must be boolean"}
		}
		if string(bytes.TrimSpace(c["same_site"])) != "null" {
			var ss string
			if err := json.Unmarshal(c["same_site"], &ss); err != nil || (ss != "Strict" && ss != "Lax" && ss != "None") {
				return &ResponseError{Reason: "cookie same_site must be Strict, Lax, None, or null"}
			}
		}
		if string(bytes.TrimSpace(c["expires"])) != "null" {
			var exp string
			if err := json.Unmarshal(c["expires"], &exp); err != nil {
				return &ResponseError{Reason: "cookie expires must be string or null"}
			}
			if _, err := time.Parse(time.RFC3339, exp); err != nil {
				return &ResponseError{Reason: "cookie expires must be RFC 3339 format"}
			}
		}
	}

	// Tagged values: body, session, principal, application_error
	if err := ValidateTaggedValue(record["body"]); err != nil {
		return &ResponseError{Reason: "response body must be valid tagged value", Cause: err}
	}
	if err := ValidateTaggedValue(record["session"]); err != nil {
		return &ResponseError{Reason: "response session must be valid tagged value", Cause: err}
	}
	if err := ValidateTaggedValue(record["principal"]); err != nil {
		return &ResponseError{Reason: "response principal must be valid tagged value", Cause: err}
	}
	if err := ValidateTaggedValue(record["application_error"]); err != nil {
		return &ResponseError{Reason: "response application_error must be valid tagged value", Cause: err}
	}

	// effect_refs
	var effectRefs []string
	if err := json.Unmarshal(record["effect_refs"], &effectRefs); err != nil {
		return &ResponseError{Reason: "effect_refs must be an array of sha256 strings"}
	}
	for _, ref := range effectRefs {
		if !digestPattern.MatchString(ref) {
			return &ResponseError{Reason: "effect_refs item must match sha256 digest pattern"}
		}
	}

	return nil
}

// ValidateEvaluationContext validates raw JSON bytes against specs/behavior/evaluation-context.schema.json.
func ValidateEvaluationContext(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return &EvaluationContextError{Reason: "not a valid JSON object", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"context_id",
		"project_id",
		"side",
		"occurrence_id",
		"projection_ref",
		"phase",
		"origin",
		"evidence_refs",
		"namespaces",
	}

	for _, f := range expectedFields {
		if _, ok := record[f]; !ok {
			return &EvaluationContextError{Reason: "missing required field " + f}
		}
	}
	if len(record) != len(expectedFields) {
		return &EvaluationContextError{Reason: "context contains unrecognized fields"}
	}

	// schema_version
	var schemaVer string
	if err := json.Unmarshal(record["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &EvaluationContextError{Reason: "schema_version must be '1.0'"}
	}

	// context_id, project_id, projection_ref
	var contextID, projectID, projectionRef string
	if err := json.Unmarshal(record["context_id"], &contextID); err != nil || !contextIDPattern.MatchString(contextID) {
		return &EvaluationContextError{Reason: "invalid context_id pattern"}
	}
	if err := json.Unmarshal(record["project_id"], &projectID); err != nil || !projectIDPattern.MatchString(projectID) {
		return &EvaluationContextError{Reason: "invalid project_id pattern"}
	}
	if err := json.Unmarshal(record["projection_ref"], &projectionRef); err != nil || !digestPattern.MatchString(projectionRef) {
		return &EvaluationContextError{Reason: "invalid projection_ref pattern"}
	}

	// side, phase, origin
	var side, phase, origin string
	if err := json.Unmarshal(record["side"], &side); err != nil {
		return &EvaluationContextError{Reason: "side must be string"}
	}
	if _, ok := allowedSides[side]; !ok {
		return &EvaluationContextError{Reason: fmt.Sprintf("unrecognized side %q", side)}
	}

	if err := json.Unmarshal(record["phase"], &phase); err != nil {
		return &EvaluationContextError{Reason: "phase must be string"}
	}
	if _, ok := allowedPhases[phase]; !ok {
		return &EvaluationContextError{Reason: fmt.Sprintf("unrecognized phase %q", phase)}
	}

	if err := json.Unmarshal(record["origin"], &origin); err != nil {
		return &EvaluationContextError{Reason: "origin must be string"}
	}
	if _, ok := allowedOrigins[origin]; !ok {
		return &EvaluationContextError{Reason: fmt.Sprintf("unrecognized origin %q", origin)}
	}

	// occurrence_id
	var occID *string
	if string(bytes.TrimSpace(record["occurrence_id"])) != "null" {
		var s string
		if err := json.Unmarshal(record["occurrence_id"], &s); err != nil || !occurrencePattern.MatchString(s) {
			return &EvaluationContextError{Reason: "occurrence_id must be null or match pattern"}
		}
		occID = &s
	}

	// evidence_refs
	var evidenceRefs []string
	if err := json.Unmarshal(record["evidence_refs"], &evidenceRefs); err != nil || len(evidenceRefs) < 1 {
		return &EvaluationContextError{Reason: "evidence_refs must be a non-empty array of sha256 digests"}
	}
	for _, ref := range evidenceRefs {
		if !digestPattern.MatchString(ref) {
			return &EvaluationContextError{Reason: "evidence_refs item must match sha256 digest pattern"}
		}
	}

	// Invariant rules from allOf:
	// 1. phase == SIMULATION -> origin == MODEL_EXPECTED and side == SHADOW
	//    phase != SIMULATION -> side in [SOURCE, TARGET]
	if phase == PhaseSimulation {
		if origin != OriginModelExpected {
			return &EvaluationContextError{Reason: "phase SIMULATION requires origin MODEL_EXPECTED"}
		}
		if side != SideShadow {
			return &EvaluationContextError{Reason: "phase SIMULATION requires side SHADOW"}
		}
	} else {
		if side != SideSource && side != SideTarget {
			return &EvaluationContextError{Reason: "non-SIMULATION phase requires side SOURCE or TARGET"}
		}
	}

	// 2. phase == FIXTURE -> origin == FIXTURE_PREPARED
	if phase == PhaseFixture {
		if origin != OriginFixturePrepared {
			return &EvaluationContextError{Reason: "phase FIXTURE requires origin FIXTURE_PREPARED"}
		}
	}

	// 3. phase in [PRE_OPERATION, POST_OPERATION, COMPLETION] -> origin == ACTUAL_OBSERVED and occurrence_id is string
	if phase == PhasePreOperation || phase == PhasePostOperation || phase == PhaseCompletion {
		if origin != OriginActualObserved {
			return &EvaluationContextError{Reason: fmt.Sprintf("phase %s requires origin ACTUAL_OBSERVED", phase)}
		}
		if occID == nil {
			return &EvaluationContextError{Reason: fmt.Sprintf("phase %s requires non-null occurrence_id", phase)}
		}
	}

	// Namespaces
	var ns map[string]json.RawMessage
	if err := json.Unmarshal(record["namespaces"], &ns); err != nil {
		return &EvaluationContextError{Reason: "namespaces must be an object", Cause: err}
	}
	nsExpected := []string{"input", "actor", "state", "result", "binding", "row", "clock"}
	for _, f := range nsExpected {
		if _, ok := ns[f]; !ok {
			return &EvaluationContextError{Reason: "namespaces missing required field " + f}
		}
	}
	if len(ns) != len(nsExpected) {
		return &EvaluationContextError{Reason: "namespaces contains unrecognized fields"}
	}

	// 4. phase == PRE_OPERATION -> namespaces.result must be null
	if phase == PhasePreOperation {
		if string(bytes.TrimSpace(ns["result"])) != "null" {
			return &EvaluationContextError{Reason: "phase PRE_OPERATION requires namespaces.result to be null"}
		}
	}

	// Validate namespaces.input
	if string(bytes.TrimSpace(ns["input"])) != "null" {
		var inputObj map[string]json.RawMessage
		if err := json.Unmarshal(ns["input"], &inputObj); err != nil {
			return &EvaluationContextError{Reason: "input must be null or an object", Cause: err}
		}
		inputReq := []string{"request_ref", "method", "raw_target", "headers", "query", "body"}
		for _, f := range inputReq {
			if _, ok := inputObj[f]; !ok {
				return &EvaluationContextError{Reason: "input missing required field " + f}
			}
		}
		if len(inputObj) != len(inputReq) {
			return &EvaluationContextError{Reason: "input contains unrecognized fields"}
		}
		var reqRef string
		if err := json.Unmarshal(inputObj["request_ref"], &reqRef); err != nil || !digestPattern.MatchString(reqRef) {
			return &EvaluationContextError{Reason: "input.request_ref must be sha256 digest"}
		}
		if err := ValidateTaggedValue(inputObj["body"]); err != nil {
			return &EvaluationContextError{Reason: "input.body must be valid tagged value", Cause: err}
		}
	}

	// Validate namespaces.actor
	var actorObj map[string]json.RawMessage
	if err := json.Unmarshal(ns["actor"], &actorObj); err != nil {
		return &EvaluationContextError{Reason: "actor must be an object", Cause: err}
	}
	actorReq := []string{"fixture_actor_ref", "actual_principal"}
	for _, f := range actorReq {
		if _, ok := actorObj[f]; !ok {
			return &EvaluationContextError{Reason: "actor missing required field " + f}
		}
	}
	if len(actorObj) != len(actorReq) {
		return &EvaluationContextError{Reason: "actor contains unrecognized fields"}
	}
	var fixActorRef string
	if err := json.Unmarshal(actorObj["fixture_actor_ref"], &fixActorRef); err != nil || !digestPattern.MatchString(fixActorRef) {
		return &EvaluationContextError{Reason: "actor.fixture_actor_ref must be sha256 digest"}
	}
	if err := validateActualPrincipal(actorObj["actual_principal"]); err != nil {
		return &EvaluationContextError{Reason: "invalid actor.actual_principal", Cause: err}
	}

	// Validate namespaces.state
	var stateObj map[string]json.RawMessage
	if err := json.Unmarshal(ns["state"], &stateObj); err != nil {
		return &EvaluationContextError{Reason: "state must be an object", Cause: err}
	}
	stateReq := []string{"db", "sessions", "cache", "files", "queues"}
	for _, f := range stateReq {
		if _, ok := stateObj[f]; !ok {
			return &EvaluationContextError{Reason: "state missing required field " + f}
		}
	}
	if len(stateObj) != len(stateReq) {
		return &EvaluationContextError{Reason: "state contains unrecognized fields"}
	}
	if err := validateStateDB(stateObj["db"]); err != nil {
		return &EvaluationContextError{Reason: "invalid state.db", Cause: err}
	}
	for _, f := range []string{"sessions", "cache", "files", "queues"} {
		if err := ValidateTaggedValue(stateObj[f]); err != nil {
			return &EvaluationContextError{Reason: fmt.Sprintf("state.%s must be valid tagged value", f), Cause: err}
		}
	}

	// Validate namespaces.result
	if string(bytes.TrimSpace(ns["result"])) != "null" {
		if err := ValidateResponse(ns["result"]); err != nil {
			return &EvaluationContextError{Reason: "invalid namespaces.result", Cause: err}
		}
	}

	// Validate namespaces.binding
	var bindingObj map[string]json.RawMessage
	if err := json.Unmarshal(ns["binding"], &bindingObj); err != nil {
		return &EvaluationContextError{Reason: "binding must be an object", Cause: err}
	}
	for k, v := range bindingObj {
		if err := ValidateTaggedValue(v); err != nil {
			return &EvaluationContextError{Reason: fmt.Sprintf("binding key %q has invalid tagged value", k), Cause: err}
		}
	}

	// Validate namespaces.row
	if string(bytes.TrimSpace(ns["row"])) != "null" {
		var rowObj map[string]json.RawMessage
		if err := json.Unmarshal(ns["row"], &rowObj); err != nil {
			return &EvaluationContextError{Reason: "row must be null or an object", Cause: err}
		}
		for k, v := range rowObj {
			if err := ValidateTaggedValue(v); err != nil {
				return &EvaluationContextError{Reason: fmt.Sprintf("row column %q has invalid tagged value", k), Cause: err}
			}
		}
	}

	// Validate namespaces.clock
	if err := ValidateTaggedValue(ns["clock"]); err != nil {
		return &EvaluationContextError{Reason: "clock must be valid tagged value", Cause: err}
	}

	return nil
}

func validateActualPrincipal(raw json.RawMessage) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return errors.New("actual_principal must be an object")
	}

	// Check if missing
	if rawType, ok := obj["type"]; ok {
		var typeStr string
		if err := json.Unmarshal(rawType, &typeStr); err != nil || typeStr != ValueTypeMissing {
			return errors.New("tagged actual_principal must have type 'missing'")
		}
		rawReason, hasReason := obj["reason_code"]
		if !hasReason {
			return errors.New("missing actual_principal requires reason_code")
		}
		var reason string
		if err := json.Unmarshal(rawReason, &reason); err != nil || reason == "" {
			return errors.New("reason_code must be non-empty string")
		}
		if len(obj) != 2 {
			return errors.New("missing actual_principal has unrecognized fields")
		}
		return nil
	}

	// Principal object
	req := []string{"authentication", "subject_local_id", "session_local_id", "roles", "capture_channel_ref", "evidence_refs"}
	for _, f := range req {
		if _, ok := obj[f]; !ok {
			return fmt.Errorf("principal missing required field %s", f)
		}
	}
	if len(obj) != len(req) {
		return errors.New("principal contains unrecognized fields")
	}

	var auth string
	if err := json.Unmarshal(obj["authentication"], &auth); err != nil || (auth != "AUTHENTICATED" && auth != "ANONYMOUS") {
		return errors.New("authentication must be AUTHENTICATED or ANONYMOUS")
	}

	var channelRef string
	if err := json.Unmarshal(obj["capture_channel_ref"], &channelRef); err != nil || !digestPattern.MatchString(channelRef) {
		return errors.New("capture_channel_ref must be sha256 digest")
	}

	var evRefs []string
	if err := json.Unmarshal(obj["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
		return errors.New("evidence_refs must be non-empty array of sha256 digests")
	}
	for _, r := range evRefs {
		if !digestPattern.MatchString(r) {
			return errors.New("evidence_refs item must match sha256 digest pattern")
		}
	}

	var roles []string
	if err := json.Unmarshal(obj["roles"], &roles); err != nil {
		return errors.New("roles must be array of strings")
	}

	if auth == "AUTHENTICATED" {
		var subj string
		if err := json.Unmarshal(obj["subject_local_id"], &subj); err != nil || !contextIDPattern.MatchString(subj) {
			return errors.New("AUTHENTICATED principal requires valid subject_local_id string")
		}
	} else {
		// ANONYMOUS: subject_local_id must be null and roles must be empty
		if string(bytes.TrimSpace(obj["subject_local_id"])) != "null" {
			return errors.New("ANONYMOUS principal requires subject_local_id to be null")
		}
		if len(roles) != 0 {
			return errors.New("ANONYMOUS principal must not have roles")
		}
	}

	return nil
}

func validateStateDB(raw json.RawMessage) error {
	// Either missing or array of tables
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		if rawType, ok := obj["type"]; ok {
			var typeStr string
			if err := json.Unmarshal(rawType, &typeStr); err != nil || typeStr != ValueTypeMissing {
				return errors.New("tagged state.db must have type 'missing'")
			}
			rawReason, hasReason := obj["reason_code"]
			if !hasReason {
				return errors.New("missing state.db requires reason_code")
			}
			var reason string
			if err := json.Unmarshal(rawReason, &reason); err != nil || reason == "" {
				return errors.New("reason_code must be non-empty string")
			}
			if len(obj) != 2 {
				return errors.New("missing state.db has unrecognized fields")
			}
			return nil
		}
	}

	var tables []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &tables); err != nil {
		return errors.New("state.db must be missing tagged value or array of table objects")
	}

	tblReq := []string{"connection_id", "catalog", "schema", "table", "rows"}
	for _, t := range tables {
		for _, f := range tblReq {
			if _, ok := t[f]; !ok {
				return fmt.Errorf("state.db table missing required field %s", f)
			}
		}
		if len(t) != len(tblReq) {
			return errors.New("state.db table contains unrecognized fields")
		}
		var connID string
		if err := json.Unmarshal(t["connection_id"], &connID); err != nil || !contextIDPattern.MatchString(connID) {
			return errors.New("connection_id must match identifier pattern")
		}
		if err := ValidateTaggedValue(t["rows"]); err != nil {
			return fmt.Errorf("state.db table rows must be valid tagged value: %w", err)
		}
	}
	return nil
}

// ParseEvaluationContext validates and parses an EvaluationContext from raw JSON.
func ParseEvaluationContext(raw []byte) (*EvaluationContext, error) {
	if err := ValidateEvaluationContext(raw); err != nil {
		return nil, err
	}
	var ctx EvaluationContext
	if err := json.Unmarshal(raw, &ctx); err != nil {
		return nil, &EvaluationContextError{Reason: "unmarshal failure", Cause: err}
	}
	return &ctx, nil
}

// CanonicalJSON serializes the EvaluationContext to rwt-cjson-1 canonical JSON bytes.
func (c *EvaluationContext) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

// ResolvePointer resolves an RFC 6901 pointer relative to the evaluation context namespaces.
// Full path syntax: "/<namespace>/<relative-pointer>".
// Transparently traverses object and array tagged value containers.
// Returns MISSING with appropriate reason code if property missing or namespace unavailable.
func (c *EvaluationContext) ResolvePointer(fullPath string) (TaggedValue, error) {
	if fullPath == "" || !strings.HasPrefix(fullPath, "/") {
		return TaggedValue{}, &EvaluationContextError{Reason: "pointer must start with '/'"}
	}

	parts := strings.Split(fullPath[1:], "/")
	for i := range parts {
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(parts[i], "~1", "/"), "~0", "~")
	}

	nsName := parts[0]
	nsRoot, err := c.getNamespaceRoot(nsName)
	if err != nil {
		return TaggedValue{}, err
	}

	if len(parts) == 1 {
		return nsRoot, nil
	}

	curr := nsRoot
	for _, token := range parts[1:] {
		if curr.Type == ValueTypeMissing {
			return curr, nil
		}
		switch curr.Type {
		case ValueTypeObject:
			objMap, ok := curr.Value.(map[string]TaggedValue)
			if !ok {
				return TaggedValue{}, &EvaluationContextError{Reason: "internal object representation mismatch"}
			}
			child, exists := objMap[token]
			if !exists {
				if nsName == "binding" {
					return NewMissing("ERR_UNBOUND_VALUE"), nil
				}
				return NewMissing("ERR_MISSING_PROPERTY"), nil
			}
			curr = child
		case ValueTypeArray:
			arr, ok := curr.Value.([]TaggedValue)
			if !ok {
				return TaggedValue{}, &EvaluationContextError{Reason: "internal array representation mismatch"}
			}
			idx, err := strconv.Atoi(token)
			if err != nil || idx < 0 || idx >= len(arr) {
				return NewMissing("ERR_INDEX_OUT_OF_BOUNDS"), nil
			}
			curr = arr[idx]
		default:
			return NewMissing("ERR_TYPE"), nil
		}
	}
	return curr, nil
}

func (c *EvaluationContext) getNamespaceRoot(nsName string) (TaggedValue, error) {
	switch nsName {
	case "clock":
		return c.Namespaces.Clock, nil
	case "binding":
		return NewObject(c.Namespaces.Binding), nil
	case "row":
		if c.Namespaces.Row == nil {
			return NewMissing("ERR_UNAVAILABLE_NAMESPACE"), nil
		}
		return NewObject(c.Namespaces.Row), nil
	case "result":
		if c.Namespaces.Result == nil {
			return NewMissing("ERR_UNAVAILABLE_NAMESPACE"), nil
		}
		res := c.Namespaces.Result
		resMap := map[string]TaggedValue{
			"schema_version":    NewString(res.SchemaVersion),
			"status":            NewStringLike(ValueTypeInteger, strconv.Itoa(res.Status)),
			"body":              res.Body,
			"session":           res.Session,
			"principal":         res.Principal,
			"application_error": res.ApplicationError,
		}
		// Headers
		hdrArr := make([]TaggedValue, len(res.Headers))
		for i, h := range res.Headers {
			hdrArr[i] = NewObject(map[string]TaggedValue{
				"name":  NewString(h.Name),
				"value": NewString(h.Value),
			})
		}
		resMap["headers"] = NewArray(hdrArr)

		// Cookies
		cookieArr := make([]TaggedValue, len(res.Cookies))
		for i, ck := range res.Cookies {
			ckMap := map[string]TaggedValue{
				"name":      NewString(ck.Name),
				"value":     ck.Value,
				"secure":    NewBoolean(ck.Secure),
				"http_only": NewBoolean(ck.HTTPOnly),
			}
			if ck.Domain != nil {
				ckMap["domain"] = NewString(*ck.Domain)
			} else {
				ckMap["domain"] = NewNull()
			}
			if ck.Path != nil {
				ckMap["path"] = NewString(*ck.Path)
			} else {
				ckMap["path"] = NewNull()
			}
			if ck.MaxAge != nil {
				ckMap["max_age"] = NewStringLike(ValueTypeInteger, strconv.Itoa(*ck.MaxAge))
			} else {
				ckMap["max_age"] = NewNull()
			}
			if ck.Expires != nil {
				ckMap["expires"] = NewString(*ck.Expires)
			} else {
				ckMap["expires"] = NewNull()
			}
			if ck.SameSite != nil {
				ckMap["same_site"] = NewString(*ck.SameSite)
			} else {
				ckMap["same_site"] = NewNull()
			}
			cookieArr[i] = NewObject(ckMap)
		}
		resMap["cookies"] = NewArray(cookieArr)

		// Effect refs
		effArr := make([]TaggedValue, len(res.EffectRefs))
		for i, ref := range res.EffectRefs {
			effArr[i] = NewStringLike(ValueTypeBytes, ref)
		}
		resMap["effect_refs"] = NewArray(effArr)

		return NewObject(resMap), nil
	case "input":
		if c.Namespaces.Input == nil {
			return NewMissing("ERR_UNAVAILABLE_NAMESPACE"), nil
		}
		inp := c.Namespaces.Input
		inpMap := map[string]TaggedValue{
			"request_ref": NewStringLike(ValueTypeBytes, inp.RequestRef),
			"method":      NewString(inp.Method),
			"raw_target":  NewString(inp.RawTarget),
			"body":        inp.Body,
		}
		hdrArr := make([]TaggedValue, len(inp.Headers))
		for i, h := range inp.Headers {
			hdrArr[i] = NewObject(map[string]TaggedValue{
				"name":  NewString(h.Name),
				"value": NewString(h.Value),
			})
		}
		inpMap["headers"] = NewArray(hdrArr)

		queryArr := make([]TaggedValue, len(inp.Query))
		for i, q := range inp.Query {
			qMap := map[string]TaggedValue{"key": NewString(q.Key)}
			if q.Value != nil {
				qMap["value"] = NewString(*q.Value)
			} else {
				qMap["value"] = NewNull()
			}
			queryArr[i] = NewObject(qMap)
		}
		inpMap["query"] = NewArray(queryArr)
		return NewObject(inpMap), nil
	case "actor":
		actor := c.Namespaces.Actor
		actorMap := map[string]TaggedValue{
			"fixture_actor_ref": NewStringLike(ValueTypeBytes, actor.FixtureActorRef),
		}
		if actor.ActualPrincipal.Missing != nil {
			actorMap["actual_principal"] = NewMissing(actor.ActualPrincipal.Missing.ReasonCode)
		} else {
			pMap := map[string]TaggedValue{
				"authentication":      NewString(actor.ActualPrincipal.Authentication),
				"capture_channel_ref": NewStringLike(ValueTypeBytes, actor.ActualPrincipal.CaptureChannelRef),
			}
			if actor.ActualPrincipal.SubjectLocalID != nil {
				pMap["subject_local_id"] = NewString(*actor.ActualPrincipal.SubjectLocalID)
			} else {
				pMap["subject_local_id"] = NewNull()
			}
			if actor.ActualPrincipal.SessionLocalID != nil {
				pMap["session_local_id"] = NewString(*actor.ActualPrincipal.SessionLocalID)
			} else {
				pMap["session_local_id"] = NewNull()
			}
			rolesArr := make([]TaggedValue, len(actor.ActualPrincipal.Roles))
			for i, r := range actor.ActualPrincipal.Roles {
				rolesArr[i] = NewString(r)
			}
			pMap["roles"] = NewArray(rolesArr)

			evArr := make([]TaggedValue, len(actor.ActualPrincipal.EvidenceRefs))
			for i, ev := range actor.ActualPrincipal.EvidenceRefs {
				evArr[i] = NewStringLike(ValueTypeBytes, ev)
			}
			pMap["evidence_refs"] = NewArray(evArr)
			actorMap["actual_principal"] = NewObject(pMap)
		}
		return NewObject(actorMap), nil
	case "state":
		st := c.Namespaces.State
		stMap := map[string]TaggedValue{
			"sessions": st.Sessions,
			"cache":    st.Cache,
			"files":    st.Files,
			"queues":   st.Queues,
		}
		if st.DB.Missing != nil {
			stMap["db"] = NewMissing(st.DB.Missing.ReasonCode)
		} else {
			tblArr := make([]TaggedValue, len(st.DB.Tables))
			for i, t := range st.DB.Tables {
				tMap := map[string]TaggedValue{
					"connection_id": NewString(t.ConnectionID),
					"catalog":       NewString(t.Catalog),
					"table":         NewString(t.Table),
					"rows":          t.Rows,
				}
				if t.Schema != nil {
					tMap["schema"] = NewString(*t.Schema)
				} else {
					tMap["schema"] = NewNull()
				}
				tblArr[i] = NewObject(tMap)
			}
			stMap["db"] = NewArray(tblArr)
		}
		return NewObject(stMap), nil
	default:
		return NewMissing("ERR_UNAVAILABLE_NAMESPACE"), nil
	}
}

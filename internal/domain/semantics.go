package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Limits specified by BEHAVIOR-MODEL.md
const (
	MaxExpressionDepth = 32
	MaxExpressionNodes = 1000
	MaxBranches        = 100
	MaxActions         = 100
	MaxArrayItems      = 100
	MaxTransitions     = 100
)

var (
	branchIDPattern      = regexp.MustCompile(`^[a-z0-9_-]+$`)
	connectionIDPattern  = regexp.MustCompile(`^[a-z0-9_-]+$`)
	transactionIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
	bindPattern          = regexp.MustCompile(`^[a-z0-9_-]+$`)
	pointerPattern       = regexp.MustCompile(`^(?:/.*)?$`)
	literalIntPattern    = regexp.MustCompile(`^(?:0|-?[1-9][0-9]*)$`)
	literalDecPattern    = regexp.MustCompile(`^-?(0|[1-9][0-9]*)\.[0-9]+$`)
	literalBytesPattern  = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

var allowedNamespaces = map[string]struct{}{
	"input":   {},
	"actor":   {},
	"state":   {},
	"result":  {},
	"clock":   {},
	"binding": {},
	"row":     {},
}

var allowedExpressionValueTypes = map[string]struct{}{
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

var allowedLiteralValueTypes = map[string]struct{}{
	"string":    {},
	"boolean":   {},
	"integer":   {},
	"decimal":   {},
	"null":      {},
	"timestamp": {},
	"bytes":     {},
}

var allowedBinaryOps = map[string]struct{}{
	"eq":           {},
	"lt":           {},
	"lte":          {},
	"in":           {},
	"add":          {},
	"sub":          {},
	"before":       {},
	"at_or_before": {},
}

var allowedNaryOps = map[string]struct{}{
	"and": {},
	"or":  {},
}

var allowedUnaryOps = map[string]struct{}{
	"not":        {},
	"exists":     {},
	"is_null":    {},
	"is_missing": {},
}

var allowedCoerceProfiles = map[string]struct{}{
	"INTEGER_TO_DECIMAL":          {},
	"CANONICAL_STRING_TO_INTEGER": {},
}

var allowedEffectClasses = map[string]struct{}{
	"PURE":           {},
	"READ_ONLY":      {},
	"LOCAL_WRITE":    {},
	"EXTERNAL_WRITE": {},
	"IRREVERSIBLE":   {},
}

var allowedCompletionModes = map[string]struct{}{
	"PROCESS_EXIT":        {},
	"HTTP_RESPONSE":       {},
	"SINK_ACK_QUIESCENCE": {},
}

var allowedMutations = map[string]struct{}{
	"INSERT": {},
	"UPDATE": {},
	"DELETE": {},
}

var allowedSessionOperations = map[string]struct{}{
	"CREATE":     {},
	"ROTATE":     {},
	"UPDATE":     {},
	"INVALIDATE": {},
}

// SemanticsError describes an invalid semantics record or execution error.
type SemanticsError struct {
	Code   string
	Reason string
	Cause  error
}

func (e *SemanticsError) Error() string {
	if e.Code != "" {
		if e.Cause != nil {
			return fmt.Sprintf("semantics error %s: %s: %v", e.Code, e.Reason, e.Cause)
		}
		return fmt.Sprintf("semantics error %s: %s", e.Code, e.Reason)
	}
	if e.Cause != nil {
		return fmt.Sprintf("semantics error: %s: %v", e.Reason, e.Cause)
	}
	return fmt.Sprintf("semantics error: %s", e.Reason)
}

func (e *SemanticsError) Unwrap() error { return e.Cause }

// BehaviorSemantics represents a behavior semantics contract.
type BehaviorSemantics struct {
	Preconditions  []*SemanticsExpression `json:"preconditions"`
	Postconditions []*SemanticsExpression `json:"postconditions"`
	Invariants     []*SemanticsExpression `json:"invariants"`
	Branches       []SemanticsBranch      `json:"branches"`
	Completion     SemanticsCompletion    `json:"completion"`
}

// SemanticsBranch represents a single branch with guard and actions.
type SemanticsBranch struct {
	BranchID string               `json:"branch_id"`
	Guard    *SemanticsExpression `json:"guard"`
	Actions  []SemanticsAction    `json:"actions"`
}

// SemanticsCompletion represents a completion boundary definition.
type SemanticsCompletion struct {
	Mode             string               `json:"mode"`
	TimeoutMs        int                  `json:"timeout_ms"`
	ExpectedReceipts int                  `json:"expected_receipts"`
	QuiescenceMs     int                  `json:"quiescence_ms"`
	Predicate        *SemanticsExpression `json:"predicate"`
}

// SemanticsExpression represents a closed tagged expression.
type SemanticsExpression struct {
	Op        string                          `json:"op"`
	ValueType string                          `json:"value_type,omitempty"`
	Value     any                             `json:"value,omitempty"`
	Namespace string                          `json:"namespace,omitempty"`
	Pointer   string                          `json:"pointer,omitempty"`
	Left      *SemanticsExpression            `json:"left,omitempty"`
	Right     *SemanticsExpression            `json:"right,omitempty"`
	Args      []*SemanticsExpression          `json:"args,omitempty"`
	Arg       *SemanticsExpression            `json:"arg,omitempty"`
	Profile   string                          `json:"profile,omitempty"`
	Items     []*SemanticsExpression          `json:"items,omitempty"`
	Fields    map[string]*SemanticsExpression `json:"fields,omitempty"`
}

// SemanticsAction represents an executable model action.
type SemanticsAction struct {
	Action             string                          `json:"action"`
	Predicate          *SemanticsExpression            `json:"predicate,omitempty"`
	ErrorCode          string                          `json:"error_code,omitempty"`
	Pointer            string                          `json:"pointer,omitempty"`
	Value              *SemanticsExpression            `json:"value,omitempty"`
	ConnectionID       string                          `json:"connection_id,omitempty"`
	TransactionID      *string                         `json:"transaction_id,omitempty"`
	Catalog            string                          `json:"catalog,omitempty"`
	Table              string                          `json:"table,omitempty"`
	Where              *SemanticsExpression            `json:"where,omitempty"`
	MaxRows            int                             `json:"max_rows,omitempty"`
	OrderBy            []string                        `json:"order_by,omitempty"`
	Bind               *string                         `json:"bind,omitempty"`
	Mutation           string                          `json:"mutation,omitempty"`
	Values             map[string]*SemanticsExpression `json:"values,omitempty"`
	EffectClass        string                          `json:"effect_class,omitempty"`
	DestinationID      string                          `json:"destination_id,omitempty"`
	Payload            *SemanticsExpression            `json:"payload,omitempty"`
	IdempotencyKey     *SemanticsExpression            `json:"idempotency_key,omitempty"`
	Status             int                             `json:"status,omitempty"`
	Body               *SemanticsExpression            `json:"body,omitempty"`
	Headers            []ActionHeader                  `json:"headers,omitempty"`
	Cookies            []ActionCookie                  `json:"cookies,omitempty"`
	RedirectLocation   *SemanticsExpression            `json:"redirect_location,omitempty"`
	Code               string                          `json:"code,omitempty"`
	Message            *SemanticsExpression            `json:"message,omitempty"`
	Operation          string                          `json:"operation,omitempty"`
	Principal          *SemanticsExpression            `json:"principal,omitempty"`
	PreviousIdentity   *SemanticsExpression            `json:"previous_identity,omitempty"`
	CurrentIdentity    *SemanticsExpression            `json:"current_identity,omitempty"`
	InvalidatePrevious bool                            `json:"invalidate_previous,omitempty"`
}

// ActionHeader represents an HTTP response header action.
type ActionHeader struct {
	Name  string               `json:"name"`
	Value *SemanticsExpression `json:"value"`
}

// ActionCookie represents an HTTP response cookie action.
type ActionCookie struct {
	Name     string               `json:"name"`
	Value    *SemanticsExpression `json:"value"`
	Domain   *SemanticsExpression `json:"domain"`
	Path     *SemanticsExpression `json:"path"`
	MaxAge   *SemanticsExpression `json:"max_age"`
	Expires  *SemanticsExpression `json:"expires"`
	SameSite *SemanticsExpression `json:"same_site"`
	Secure   bool                 `json:"secure"`
	HttpOnly bool                 `json:"http_only"`
}

// StepTransition represents a transition edge after step execution.
type StepTransition struct {
	Guard      *SemanticsExpression `json:"guard"`
	NextStepID *string              `json:"next_step_id"`
	TerminalID *string              `json:"terminal_id"`
}

// ValidateSemantics validates a behavior semantics document against specs/behavior/semantics.schema.json.
func ValidateSemantics(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return &SemanticsError{Reason: "not a valid JSON object", Cause: err}
	}

	expectedFields := []string{"preconditions", "postconditions", "invariants", "branches", "completion"}
	for _, f := range expectedFields {
		if _, ok := record[f]; !ok {
			return &SemanticsError{Reason: "missing required field " + f}
		}
	}
	if len(record) != len(expectedFields) {
		return &SemanticsError{Reason: "semantics contains unrecognized fields"}
	}

	nodeCounter := 0

	// Preconditions
	var rawPre []json.RawMessage
	if err := json.Unmarshal(record["preconditions"], &rawPre); err != nil {
		return &SemanticsError{Reason: "preconditions must be array", Cause: err}
	}
	for _, rp := range rawPre {
		if _, err := validateExpressionRaw(rp, 1, &nodeCounter); err != nil {
			return err
		}
	}

	// Postconditions
	var rawPost []json.RawMessage
	if err := json.Unmarshal(record["postconditions"], &rawPost); err != nil {
		return &SemanticsError{Reason: "postconditions must be array", Cause: err}
	}
	for _, rp := range rawPost {
		if _, err := validateExpressionRaw(rp, 1, &nodeCounter); err != nil {
			return err
		}
	}

	// Invariants
	var rawInv []json.RawMessage
	if err := json.Unmarshal(record["invariants"], &rawInv); err != nil {
		return &SemanticsError{Reason: "invariants must be array", Cause: err}
	}
	for _, rp := range rawInv {
		if _, err := validateExpressionRaw(rp, 1, &nodeCounter); err != nil {
			return err
		}
	}

	// Branches
	var rawBranches []json.RawMessage
	if err := json.Unmarshal(record["branches"], &rawBranches); err != nil {
		return &SemanticsError{Reason: "branches must be array", Cause: err}
	}
	if len(rawBranches) < 1 || len(rawBranches) > MaxBranches {
		return &SemanticsError{Code: "ERR_UNSUPPORTED_CAPABILITY", Reason: fmt.Sprintf("branches length %d out of bounds (1..%d)", len(rawBranches), MaxBranches)}
	}

	for _, rb := range rawBranches {
		var branch map[string]json.RawMessage
		if err := json.Unmarshal(rb, &branch); err != nil {
			return &SemanticsError{Reason: "branch must be object", Cause: err}
		}
		if _, ok := branch["branch_id"]; !ok {
			return &SemanticsError{Reason: "branch missing branch_id"}
		}
		if _, ok := branch["guard"]; !ok {
			return &SemanticsError{Reason: "branch missing guard"}
		}
		if _, ok := branch["actions"]; !ok {
			return &SemanticsError{Reason: "branch missing actions"}
		}
		if len(branch) != 3 {
			return &SemanticsError{Reason: "branch contains unrecognized fields"}
		}

		var bID string
		if err := json.Unmarshal(branch["branch_id"], &bID); err != nil || !branchIDPattern.MatchString(bID) {
			return &SemanticsError{Reason: "invalid branch_id pattern"}
		}

		if _, err := validateExpressionRaw(branch["guard"], 1, &nodeCounter); err != nil {
			return err
		}

		var rawActions []json.RawMessage
		if err := json.Unmarshal(branch["actions"], &rawActions); err != nil {
			return &SemanticsError{Reason: "actions must be array", Cause: err}
		}
		if len(rawActions) < 1 || len(rawActions) > MaxActions {
			return &SemanticsError{Code: "ERR_UNSUPPORTED_CAPABILITY", Reason: fmt.Sprintf("actions length %d out of bounds (1..%d)", len(rawActions), MaxActions)}
		}
		for _, ra := range rawActions {
			if _, err := validateActionRaw(ra, &nodeCounter); err != nil {
				return err
			}
		}
	}

	// Completion
	if _, err := validateCompletionRaw(record["completion"], &nodeCounter); err != nil {
		return err
	}

	if nodeCounter > MaxExpressionNodes {
		return &SemanticsError{Code: "ERR_UNSUPPORTED_CAPABILITY", Reason: fmt.Sprintf("total expression nodes %d exceeds limit %d", nodeCounter, MaxExpressionNodes)}
	}

	return nil
}

// ValidateAction validates an action definition against specs/behavior/semantics.schema.json#/$defs/action.
func ValidateAction(raw []byte) error {
	nodeCounter := 0
	_, err := validateActionRaw(raw, &nodeCounter)
	return err
}

func validateActionRaw(raw json.RawMessage, nodeCounter *int) (*SemanticsAction, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, &SemanticsError{Reason: "action must be object", Cause: err}
	}
	rawAct, ok := m["action"]
	if !ok {
		return nil, &SemanticsError{Reason: "missing required field action"}
	}
	var actName string
	if err := json.Unmarshal(rawAct, &actName); err != nil {
		return nil, &SemanticsError{Reason: "action field must be string", Cause: err}
	}

	switch actName {
	case "assert":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: "assert action unrecognized fields"}
		}
		predRaw, ok1 := m["predicate"]
		errCodeRaw, ok2 := m["error_code"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: "assert action missing predicate or error_code"}
		}
		var errCode string
		if err := json.Unmarshal(errCodeRaw, &errCode); err != nil || len(errCode) < 1 {
			return nil, &SemanticsError{Reason: "assert error_code must be non-empty string"}
		}
		pred, err := validateExpressionRaw(predRaw, 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsAction{Action: actName, Predicate: pred, ErrorCode: errCode}, nil

	case "set_state":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: "set_state action unrecognized fields"}
		}
		ptrRaw, ok1 := m["pointer"]
		valRaw, ok2 := m["value"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: "set_state missing pointer or value"}
		}
		var ptr string
		if err := json.Unmarshal(ptrRaw, &ptr); err != nil || len(ptr) < 1 {
			return nil, &SemanticsError{Reason: "set_state pointer must be non-empty string"}
		}
		val, err := validateExpressionRaw(valRaw, 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsAction{Action: actName, Pointer: ptr, Value: val}, nil

	case "begin", "commit", "rollback":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: actName + " action unrecognized fields"}
		}
		connRaw, ok1 := m["connection_id"]
		txRaw, ok2 := m["transaction_id"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: actName + " missing connection_id or transaction_id"}
		}
		var connID, txID string
		if err := json.Unmarshal(connRaw, &connID); err != nil || !connectionIDPattern.MatchString(connID) {
			return nil, &SemanticsError{Reason: "invalid connection_id pattern"}
		}
		if err := json.Unmarshal(txRaw, &txID); err != nil || !transactionIDPattern.MatchString(txID) {
			return nil, &SemanticsError{Reason: "invalid transaction_id pattern"}
		}
		return &SemanticsAction{Action: actName, ConnectionID: connID, TransactionID: &txID}, nil

	case "db_read":
		if len(m) != 8 {
			return nil, &SemanticsError{Reason: "db_read action unrecognized fields"}
		}
		reqFields := []string{"connection_id", "catalog", "table", "where", "max_rows", "order_by", "bind"}
		for _, f := range reqFields {
			if _, ok := m[f]; !ok {
				return nil, &SemanticsError{Reason: "db_read missing " + f}
			}
		}
		var connID, catalog, table, bind string
		if err := json.Unmarshal(m["connection_id"], &connID); err != nil || !connectionIDPattern.MatchString(connID) {
			return nil, &SemanticsError{Reason: "invalid connection_id pattern"}
		}
		if err := json.Unmarshal(m["catalog"], &catalog); err != nil || len(catalog) < 1 {
			return nil, &SemanticsError{Reason: "catalog must be non-empty string"}
		}
		if err := json.Unmarshal(m["table"], &table); err != nil || len(table) < 1 {
			return nil, &SemanticsError{Reason: "table must be non-empty string"}
		}
		if err := json.Unmarshal(m["bind"], &bind); err != nil || !bindPattern.MatchString(bind) {
			return nil, &SemanticsError{Reason: "invalid bind pattern"}
		}
		var maxRows int
		if err := json.Unmarshal(m["max_rows"], &maxRows); err != nil || maxRows < 0 || maxRows > 10000 {
			return nil, &SemanticsError{Reason: "max_rows must be 0..10000"}
		}
		var orderBy []string
		if err := json.Unmarshal(m["order_by"], &orderBy); err != nil {
			return nil, &SemanticsError{Reason: "order_by must be array of strings"}
		}
		seenOrder := make(map[string]struct{}, len(orderBy))
		for _, col := range orderBy {
			if len(col) < 1 {
				return nil, &SemanticsError{Reason: "order_by column must be non-empty string"}
			}
			if _, exists := seenOrder[col]; exists {
				return nil, &SemanticsError{Reason: "order_by must have unique items"}
			}
			seenOrder[col] = struct{}{}
		}
		whereExpr, err := validateExpressionRaw(m["where"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsAction{
			Action:       actName,
			ConnectionID: connID,
			Catalog:      catalog,
			Table:        table,
			Where:        whereExpr,
			MaxRows:      maxRows,
			OrderBy:      orderBy,
			Bind:         &bind,
		}, nil

	case "db_write":
		reqFields := []string{"connection_id", "catalog", "table", "mutation", "where", "values", "transaction_id"}
		for _, f := range reqFields {
			if _, ok := m[f]; !ok {
				return nil, &SemanticsError{Reason: "db_write missing " + f}
			}
		}
		hasBind := false
		if _, ok := m["bind"]; ok {
			hasBind = true
		}
		expectedLen := 8
		if hasBind {
			expectedLen = 9
		}
		if len(m) != expectedLen {
			return nil, &SemanticsError{Reason: "db_write contains unrecognized fields"}
		}

		var connID, catalog, table, mutation string
		if err := json.Unmarshal(m["connection_id"], &connID); err != nil || !connectionIDPattern.MatchString(connID) {
			return nil, &SemanticsError{Reason: "invalid connection_id pattern"}
		}
		if err := json.Unmarshal(m["catalog"], &catalog); err != nil || len(catalog) < 1 {
			return nil, &SemanticsError{Reason: "catalog must be non-empty string"}
		}
		if err := json.Unmarshal(m["table"], &table); err != nil || len(table) < 1 {
			return nil, &SemanticsError{Reason: "table must be non-empty string"}
		}
		if err := json.Unmarshal(m["mutation"], &mutation); err != nil {
			return nil, &SemanticsError{Reason: "mutation must be string"}
		}
		if _, ok := allowedMutations[mutation]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported mutation %q", mutation)}
		}

		var txID *string
		if string(m["transaction_id"]) != "null" {
			var s string
			if err := json.Unmarshal(m["transaction_id"], &s); err != nil || !transactionIDPattern.MatchString(s) {
				return nil, &SemanticsError{Reason: "invalid transaction_id pattern"}
			}
			txID = &s
		}

		var bind *string
		if hasBind {
			var b string
			if err := json.Unmarshal(m["bind"], &b); err != nil || !bindPattern.MatchString(b) {
				return nil, &SemanticsError{Reason: "invalid bind pattern"}
			}
			bind = &b
		}

		whereExpr, err := validateExpressionRaw(m["where"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}

		var rawValues map[string]json.RawMessage
		if err := json.Unmarshal(m["values"], &rawValues); err != nil {
			return nil, &SemanticsError{Reason: "values must be object", Cause: err}
		}
		values := make(map[string]*SemanticsExpression, len(rawValues))
		for k, rv := range rawValues {
			ve, err := validateExpressionRaw(rv, 1, nodeCounter)
			if err != nil {
				return nil, err
			}
			values[k] = ve
		}

		return &SemanticsAction{
			Action:        actName,
			ConnectionID:  connID,
			Catalog:       catalog,
			Table:         table,
			Mutation:      mutation,
			Where:         whereExpr,
			Values:        values,
			TransactionID: txID,
			Bind:          bind,
		}, nil

	case "emit":
		if len(m) != 5 {
			return nil, &SemanticsError{Reason: "emit action unrecognized fields"}
		}
		reqFields := []string{"effect_class", "destination_id", "payload", "idempotency_key"}
		for _, f := range reqFields {
			if _, ok := m[f]; !ok {
				return nil, &SemanticsError{Reason: "emit missing " + f}
			}
		}
		var effClass, destID string
		if err := json.Unmarshal(m["effect_class"], &effClass); err != nil {
			return nil, &SemanticsError{Reason: "effect_class must be string"}
		}
		if _, ok := allowedEffectClasses[effClass]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported effect_class %q", effClass)}
		}
		if err := json.Unmarshal(m["destination_id"], &destID); err != nil || !connectionIDPattern.MatchString(destID) {
			return nil, &SemanticsError{Reason: "invalid destination_id pattern"}
		}
		payloadExpr, err := validateExpressionRaw(m["payload"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		idempotencyKeyExpr, err := validateExpressionRaw(m["idempotency_key"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsAction{
			Action:         actName,
			EffectClass:    effClass,
			DestinationID:  destID,
			Payload:        payloadExpr,
			IdempotencyKey: idempotencyKeyExpr,
		}, nil

	case "respond":
		reqFields := []string{"status", "body"}
		for _, f := range reqFields {
			if _, ok := m[f]; !ok {
				return nil, &SemanticsError{Reason: "respond missing " + f}
			}
		}
		var status int
		if err := json.Unmarshal(m["status"], &status); err != nil || status < 100 || status > 599 {
			return nil, &SemanticsError{Reason: "status must be 100..599"}
		}
		bodyExpr, err := validateExpressionRaw(m["body"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}

		var headers []ActionHeader
		if rawHeaders, ok := m["headers"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawHeaders, &rawList); err != nil {
				return nil, &SemanticsError{Reason: "headers must be array", Cause: err}
			}
			headers = make([]ActionHeader, 0, len(rawList))
			for _, rh := range rawList {
				var hm map[string]json.RawMessage
				if err := json.Unmarshal(rh, &hm); err != nil || len(hm) != 2 {
					return nil, &SemanticsError{Reason: "header item must have exactly name and value"}
				}
				nameRaw, ok1 := hm["name"]
				valRaw, ok2 := hm["value"]
				if !ok1 || !ok2 {
					return nil, &SemanticsError{Reason: "header item missing name or value"}
				}
				var hName string
				if err := json.Unmarshal(nameRaw, &hName); err != nil || !headerNamePattern.MatchString(hName) {
					return nil, &SemanticsError{Reason: "invalid header name pattern"}
				}
				if strings.EqualFold(hName, "set-cookie") {
					return nil, &SemanticsError{Reason: "Set-Cookie cannot be placed in generic headers action; use cookies"}
				}
				hVal, err := validateExpressionRaw(valRaw, 1, nodeCounter)
				if err != nil {
					return nil, err
				}
				headers = append(headers, ActionHeader{Name: hName, Value: hVal})
			}
		}

		var cookies []ActionCookie
		if rawCookies, ok := m["cookies"]; ok {
			var rawList []json.RawMessage
			if err := json.Unmarshal(rawCookies, &rawList); err != nil {
				return nil, &SemanticsError{Reason: "cookies must be array", Cause: err}
			}
			cookies = make([]ActionCookie, 0, len(rawList))
			cookieFields := []string{"name", "value", "domain", "path", "max_age", "expires", "same_site", "secure", "http_only"}
			for _, rc := range rawList {
				var cm map[string]json.RawMessage
				if err := json.Unmarshal(rc, &cm); err != nil || len(cm) != len(cookieFields) {
					return nil, &SemanticsError{Reason: "cookie item must have exactly the 9 required fields"}
				}
				for _, cf := range cookieFields {
					if _, ok := cm[cf]; !ok {
						return nil, &SemanticsError{Reason: "cookie item missing " + cf}
					}
				}
				var cName string
				if err := json.Unmarshal(cm["name"], &cName); err != nil {
					return nil, &SemanticsError{Reason: "cookie name must be string"}
				}
				cVal, err := validateExpressionRaw(cm["value"], 1, nodeCounter)
				if err != nil {
					return nil, err
				}
				var secure, httpOnly bool
				if err := json.Unmarshal(cm["secure"], &secure); err != nil {
					return nil, &SemanticsError{Reason: "cookie secure must be boolean"}
				}
				if err := json.Unmarshal(cm["http_only"], &httpOnly); err != nil {
					return nil, &SemanticsError{Reason: "cookie http_only must be boolean"}
				}

				parseOptionalExpr := func(rawMsg json.RawMessage) (*SemanticsExpression, error) {
					if string(rawMsg) == "null" {
						return nil, nil
					}
					return validateExpressionRaw(rawMsg, 1, nodeCounter)
				}

				domainExpr, err := parseOptionalExpr(cm["domain"])
				if err != nil {
					return nil, err
				}
				pathExpr, err := parseOptionalExpr(cm["path"])
				if err != nil {
					return nil, err
				}
				maxAgeExpr, err := parseOptionalExpr(cm["max_age"])
				if err != nil {
					return nil, err
				}
				expiresExpr, err := parseOptionalExpr(cm["expires"])
				if err != nil {
					return nil, err
				}
				sameSiteExpr, err := parseOptionalExpr(cm["same_site"])
				if err != nil {
					return nil, err
				}

				cookies = append(cookies, ActionCookie{
					Name:     cName,
					Value:    cVal,
					Domain:   domainExpr,
					Path:     pathExpr,
					MaxAge:   maxAgeExpr,
					Expires:  expiresExpr,
					SameSite: sameSiteExpr,
					Secure:   secure,
					HttpOnly: httpOnly,
				})
			}
		}

		var redirectLoc *SemanticsExpression
		if rawRedir, ok := m["redirect_location"]; ok {
			if string(rawRedir) != "null" {
				rExpr, err := validateExpressionRaw(rawRedir, 1, nodeCounter)
				if err != nil {
					return nil, err
				}
				redirectLoc = rExpr
			}
		}

		return &SemanticsAction{
			Action:           actName,
			Status:           status,
			Body:             bodyExpr,
			Headers:          headers,
			Cookies:          cookies,
			RedirectLocation: redirectLoc,
		}, nil

	case "error":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: "error action unrecognized fields"}
		}
		cRaw, ok1 := m["code"]
		mRaw, ok2 := m["message"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: "error action missing code or message"}
		}
		var code string
		if err := json.Unmarshal(cRaw, &code); err != nil || len(code) < 1 {
			return nil, &SemanticsError{Reason: "code must be non-empty string"}
		}
		msgExpr, err := validateExpressionRaw(mRaw, 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsAction{Action: actName, Code: code, Message: msgExpr}, nil

	case "session":
		if len(m) != 7 {
			return nil, &SemanticsError{Reason: "session action unrecognized fields"}
		}
		reqFields := []string{"operation", "principal", "previous_identity", "current_identity", "values", "invalidate_previous"}
		for _, f := range reqFields {
			if _, ok := m[f]; !ok {
				return nil, &SemanticsError{Reason: "session missing " + f}
			}
		}
		var op string
		if err := json.Unmarshal(m["operation"], &op); err != nil {
			return nil, &SemanticsError{Reason: "operation must be string"}
		}
		if _, ok := allowedSessionOperations[op]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported session operation %q", op)}
		}
		princExpr, err := validateExpressionRaw(m["principal"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		prevIdExpr, err := validateExpressionRaw(m["previous_identity"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		currIdExpr, err := validateExpressionRaw(m["current_identity"], 1, nodeCounter)
		if err != nil {
			return nil, err
		}
		var invalPrev bool
		if err := json.Unmarshal(m["invalidate_previous"], &invalPrev); err != nil {
			return nil, &SemanticsError{Reason: "invalidate_previous must be boolean"}
		}

		var rawValues map[string]json.RawMessage
		if err := json.Unmarshal(m["values"], &rawValues); err != nil {
			return nil, &SemanticsError{Reason: "values must be object"}
		}
		values := make(map[string]*SemanticsExpression, len(rawValues))
		for k, rv := range rawValues {
			ve, err := validateExpressionRaw(rv, 1, nodeCounter)
			if err != nil {
				return nil, err
			}
			values[k] = ve
		}

		return &SemanticsAction{
			Action:             actName,
			Operation:          op,
			Principal:          princExpr,
			PreviousIdentity:   prevIdExpr,
			CurrentIdentity:    currIdExpr,
			Values:             values,
			InvalidatePrevious: invalPrev,
		}, nil

	default:
		return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported action %q", actName)}
	}
}

// ValidateCompletion validates a completion definition against specs/behavior/semantics.schema.json#/$defs/completion.
func ValidateCompletion(raw []byte) error {
	nodeCounter := 0
	_, err := validateCompletionRaw(raw, &nodeCounter)
	return err
}

func validateCompletionRaw(raw json.RawMessage, nodeCounter *int) (*SemanticsCompletion, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, &SemanticsError{Reason: "completion must be object", Cause: err}
	}

	reqFields := []string{"mode", "timeout_ms", "expected_receipts", "quiescence_ms", "predicate"}
	for _, f := range reqFields {
		if _, ok := m[f]; !ok {
			return nil, &SemanticsError{Reason: "completion missing " + f}
		}
	}
	if len(m) != len(reqFields) {
		return nil, &SemanticsError{Reason: "completion contains unrecognized fields"}
	}

	var mode string
	if err := json.Unmarshal(m["mode"], &mode); err != nil {
		return nil, &SemanticsError{Reason: "mode must be string"}
	}
	if _, ok := allowedCompletionModes[mode]; !ok {
		return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported completion mode %q", mode)}
	}

	var timeoutMs, expectedReceipts, quiescenceMs int
	if err := json.Unmarshal(m["timeout_ms"], &timeoutMs); err != nil || timeoutMs < 1 || timeoutMs > 300000 {
		return nil, &SemanticsError{Reason: "timeout_ms must be 1..300000"}
	}
	if err := json.Unmarshal(m["expected_receipts"], &expectedReceipts); err != nil || expectedReceipts < 0 {
		return nil, &SemanticsError{Reason: "expected_receipts must be >= 0"}
	}
	if err := json.Unmarshal(m["quiescence_ms"], &quiescenceMs); err != nil || quiescenceMs < 0 {
		return nil, &SemanticsError{Reason: "quiescence_ms must be >= 0"}
	}

	// Cross-field schema rules
	if mode == "SINK_ACK_QUIESCENCE" {
		if quiescenceMs < 1 {
			return nil, &SemanticsError{Reason: "SINK_ACK_QUIESCENCE requires quiescence_ms >= 1"}
		}
		if quiescenceMs > timeoutMs {
			return nil, &SemanticsError{Reason: "quiescence_ms cannot exceed timeout_ms"}
		}
	} else {
		if expectedReceipts != 0 {
			return nil, &SemanticsError{Reason: fmt.Sprintf("%s requires expected_receipts == 0", mode)}
		}
		if quiescenceMs != 0 {
			return nil, &SemanticsError{Reason: fmt.Sprintf("%s requires quiescence_ms == 0", mode)}
		}
	}

	predExpr, err := validateExpressionRaw(m["predicate"], 1, nodeCounter)
	if err != nil {
		return nil, err
	}

	return &SemanticsCompletion{
		Mode:             mode,
		TimeoutMs:        timeoutMs,
		ExpectedReceipts: expectedReceipts,
		QuiescenceMs:     quiescenceMs,
		Predicate:        predExpr,
	}, nil
}

// ValidateExpression validates an expression against specs/behavior/semantics.schema.json#/$defs/expression.
func ValidateExpression(raw []byte) error {
	nodeCounter := 0
	_, err := validateExpressionRaw(raw, 1, &nodeCounter)
	return err
}

func validateExpressionRaw(raw json.RawMessage, depth int, nodeCounter *int) (*SemanticsExpression, error) {
	if depth > MaxExpressionDepth {
		return nil, &SemanticsError{
			Code:   "ERR_UNSUPPORTED_CAPABILITY",
			Reason: fmt.Sprintf("expression depth %d exceeds maximum allowed %d", depth, MaxExpressionDepth),
		}
	}
	*nodeCounter++
	if *nodeCounter > MaxExpressionNodes {
		return nil, &SemanticsError{
			Code:   "ERR_UNSUPPORTED_CAPABILITY",
			Reason: fmt.Sprintf("expression node count %d exceeds maximum allowed %d", *nodeCounter, MaxExpressionNodes),
		}
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, &SemanticsError{Reason: "expression must be object", Cause: err}
	}

	opRaw, ok := m["op"]
	if !ok {
		return nil, &SemanticsError{Reason: "expression missing required op"}
	}
	var op string
	if err := json.Unmarshal(opRaw, &op); err != nil {
		return nil, &SemanticsError{Reason: "expression op must be string"}
	}

	switch op {
	case "literal":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: "literal expression must have exactly op, value_type, and value"}
		}
		vtRaw, ok1 := m["value_type"]
		valRaw, ok2 := m["value"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: "literal missing value_type or value"}
		}
		var vt string
		if err := json.Unmarshal(vtRaw, &vt); err != nil {
			return nil, &SemanticsError{Reason: "value_type must be string"}
		}
		if _, ok := allowedLiteralValueTypes[vt]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported literal value_type %q", vt)}
		}

		var parsedVal any
		switch vt {
		case "string":
			var s string
			if err := json.Unmarshal(valRaw, &s); err != nil {
				return nil, &SemanticsError{Reason: "literal string value must be string", Cause: err}
			}
			parsedVal = s
		case "boolean":
			var b bool
			if err := json.Unmarshal(valRaw, &b); err != nil {
				return nil, &SemanticsError{Reason: "literal boolean value must be boolean", Cause: err}
			}
			parsedVal = b
		case "integer":
			var s string
			if err := json.Unmarshal(valRaw, &s); err != nil || !literalIntPattern.MatchString(s) {
				return nil, &SemanticsError{Reason: "literal integer value must match canonical integer pattern"}
			}
			parsedVal = s
		case "decimal":
			var s string
			if err := json.Unmarshal(valRaw, &s); err != nil || !literalDecPattern.MatchString(s) {
				return nil, &SemanticsError{Reason: "literal decimal value must match decimal pattern with decimal point"}
			}
			parsedVal = s
		case "null":
			if string(valRaw) != "null" {
				return nil, &SemanticsError{Reason: "literal null value must be null"}
			}
			parsedVal = nil
		case "timestamp":
			var s string
			if err := json.Unmarshal(valRaw, &s); err != nil {
				return nil, &SemanticsError{Reason: "literal timestamp value must be string"}
			}
			if _, err := time.Parse(time.RFC3339, s); err != nil {
				return nil, &SemanticsError{Reason: "literal timestamp value must be RFC3339 date-time", Cause: err}
			}
			parsedVal = s
		case "bytes":
			var s string
			if err := json.Unmarshal(valRaw, &s); err != nil || !literalBytesPattern.MatchString(s) {
				return nil, &SemanticsError{Reason: "literal bytes value must match sha256: pattern"}
			}
			parsedVal = s
		}

		return &SemanticsExpression{
			Op:        op,
			ValueType: vt,
			Value:     parsedVal,
		}, nil

	case "ref":
		if len(m) != 4 {
			return nil, &SemanticsError{Reason: "ref expression must have exactly op, namespace, pointer, value_type"}
		}
		nsRaw, ok1 := m["namespace"]
		ptrRaw, ok2 := m["pointer"]
		vtRaw, ok3 := m["value_type"]
		if !ok1 || !ok2 || !ok3 {
			return nil, &SemanticsError{Reason: "ref missing namespace, pointer, or value_type"}
		}
		var ns, ptr, vt string
		if err := json.Unmarshal(nsRaw, &ns); err != nil {
			return nil, &SemanticsError{Reason: "namespace must be string"}
		}
		if _, ok := allowedNamespaces[ns]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported namespace %q", ns)}
		}
		if err := json.Unmarshal(ptrRaw, &ptr); err != nil || !pointerPattern.MatchString(ptr) {
			return nil, &SemanticsError{Reason: "invalid pointer pattern"}
		}
		if err := json.Unmarshal(vtRaw, &vt); err != nil {
			return nil, &SemanticsError{Reason: "value_type must be string"}
		}
		if _, ok := allowedExpressionValueTypes[vt]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported ref value_type %q", vt)}
		}

		return &SemanticsExpression{
			Op:        op,
			Namespace: ns,
			Pointer:   ptr,
			ValueType: vt,
		}, nil

	case "eq", "lt", "lte", "in", "add", "sub", "before", "at_or_before":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: op + " expression must have exactly op, left, right"}
		}
		leftRaw, ok1 := m["left"]
		rightRaw, ok2 := m["right"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: op + " missing left or right"}
		}
		left, err := validateExpressionRaw(leftRaw, depth+1, nodeCounter)
		if err != nil {
			return nil, err
		}
		right, err := validateExpressionRaw(rightRaw, depth+1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsExpression{
			Op:    op,
			Left:  left,
			Right: right,
		}, nil

	case "and", "or":
		if len(m) != 2 {
			return nil, &SemanticsError{Reason: op + " expression must have exactly op and args"}
		}
		argsRaw, ok := m["args"]
		if !ok {
			return nil, &SemanticsError{Reason: op + " missing args"}
		}
		var rawList []json.RawMessage
		if err := json.Unmarshal(argsRaw, &rawList); err != nil {
			return nil, &SemanticsError{Reason: "args must be array", Cause: err}
		}
		if len(rawList) < 2 || len(rawList) > MaxArrayItems {
			return nil, &SemanticsError{Reason: fmt.Sprintf("args length %d out of bounds (2..%d)", len(rawList), MaxArrayItems)}
		}
		args := make([]*SemanticsExpression, 0, len(rawList))
		for _, ra := range rawList {
			argExpr, err := validateExpressionRaw(ra, depth+1, nodeCounter)
			if err != nil {
				return nil, err
			}
			args = append(args, argExpr)
		}
		return &SemanticsExpression{
			Op:   op,
			Args: args,
		}, nil

	case "not", "exists", "is_null", "is_missing":
		if len(m) != 2 {
			return nil, &SemanticsError{Reason: op + " expression must have exactly op and arg"}
		}
		argRaw, ok := m["arg"]
		if !ok {
			return nil, &SemanticsError{Reason: op + " missing arg"}
		}
		arg, err := validateExpressionRaw(argRaw, depth+1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsExpression{
			Op:  op,
			Arg: arg,
		}, nil

	case "coerce":
		if len(m) != 3 {
			return nil, &SemanticsError{Reason: "coerce expression must have exactly op, profile, arg"}
		}
		profRaw, ok1 := m["profile"]
		argRaw, ok2 := m["arg"]
		if !ok1 || !ok2 {
			return nil, &SemanticsError{Reason: "coerce missing profile or arg"}
		}
		var prof string
		if err := json.Unmarshal(profRaw, &prof); err != nil {
			return nil, &SemanticsError{Reason: "profile must be string"}
		}
		if _, ok := allowedCoerceProfiles[prof]; !ok {
			return nil, &SemanticsError{Reason: fmt.Sprintf("unsupported coerce profile %q", prof)}
		}
		arg, err := validateExpressionRaw(argRaw, depth+1, nodeCounter)
		if err != nil {
			return nil, err
		}
		return &SemanticsExpression{
			Op:      op,
			Profile: prof,
			Arg:     arg,
		}, nil

	case "array":
		if len(m) != 2 {
			return nil, &SemanticsError{Reason: "array expression must have exactly op and items"}
		}
		itemsRaw, ok := m["items"]
		if !ok {
			return nil, &SemanticsError{Reason: "array missing items"}
		}
		var rawList []json.RawMessage
		if err := json.Unmarshal(itemsRaw, &rawList); err != nil {
			return nil, &SemanticsError{Reason: "items must be array", Cause: err}
		}
		if len(rawList) > MaxArrayItems {
			return nil, &SemanticsError{Reason: fmt.Sprintf("items length %d exceeds max %d", len(rawList), MaxArrayItems)}
		}
		items := make([]*SemanticsExpression, 0, len(rawList))
		for _, ri := range rawList {
			itemExpr, err := validateExpressionRaw(ri, depth+1, nodeCounter)
			if err != nil {
				return nil, err
			}
			items = append(items, itemExpr)
		}
		return &SemanticsExpression{
			Op:    op,
			Items: items,
		}, nil

	case "object":
		if len(m) != 2 {
			return nil, &SemanticsError{Reason: "object expression must have exactly op and fields"}
		}
		fieldsRaw, ok := m["fields"]
		if !ok {
			return nil, &SemanticsError{Reason: "object missing fields"}
		}
		var rawMap map[string]json.RawMessage
		if err := json.Unmarshal(fieldsRaw, &rawMap); err != nil {
			return nil, &SemanticsError{Reason: "fields must be object", Cause: err}
		}
		fields := make(map[string]*SemanticsExpression, len(rawMap))
		for k, rf := range rawMap {
			fieldExpr, err := validateExpressionRaw(rf, depth+1, nodeCounter)
			if err != nil {
				return nil, err
			}
			fields[k] = fieldExpr
		}
		return &SemanticsExpression{
			Op:     op,
			Fields: fields,
		}, nil

	default:
		return nil, &SemanticsError{
			Code:   "ERR_UNSUPPORTED_OPERATOR",
			Reason: fmt.Sprintf("unsupported expression operator %q", op),
		}
	}
}

// ParseSemantics parses and validates a BehaviorSemantics record.
func ParseSemantics(raw []byte) (*BehaviorSemantics, error) {
	if err := ValidateSemantics(raw); err != nil {
		return nil, err
	}
	var sem BehaviorSemantics
	if err := json.Unmarshal(raw, &sem); err != nil {
		return nil, &SemanticsError{Reason: "failed to unmarshal semantics", Cause: err}
	}
	return &sem, nil
}

// ValidateStepGuard enforces that a step guard in the v2 subset must be literal boolean true (X03 / B2).
func ValidateStepGuard(guard *SemanticsExpression) error {
	if guard == nil {
		return &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "missing required step guard"}
	}
	if guard.Op == "literal" && guard.ValueType == ValueTypeBoolean && guard.Value == true {
		return nil
	}
	return &SemanticsError{
		Code:   "ERR_UNSUPPORTED_CAPABILITY",
		Reason: "step guard must be literal true in v2 subset; conditional routing belongs to transitions",
	}
}

// EvaluateExpression evaluates a closed expression tree against an EvaluationContext.
// It enforces strict type checks, no implicit cross-type coercion, exact integer/decimal arithmetic,
// and short-circuit boolean semantics.
func EvaluateExpression(expr *SemanticsExpression, ctx *EvaluationContext) (TaggedValue, error) {
	if expr == nil {
		return NewMissing("ERR_INVALID_CONTRACT"), nil
	}

	switch expr.Op {
	case "literal":
		switch expr.ValueType {
		case ValueTypeString:
			return NewString(expr.Value.(string)), nil
		case ValueTypeBoolean:
			return NewBoolean(expr.Value.(bool)), nil
		case ValueTypeInteger:
			return NewInteger(expr.Value.(string))
		case ValueTypeDecimal:
			// Normalize literal decimal (e.g. trailing zeros removed, -0.0 -> 0)
			val := ProjectRawJSONNumber(expr.Value.(string))
			if val.Type != ValueTypeMissing {
				val.Type = ValueTypeDecimal
			}
			return val, nil
		case ValueTypeNull:
			return NewNull(), nil
		case ValueTypeTimestamp:
			return NewTimestamp(expr.Value.(string))
		case ValueTypeBytes:
			return NewBytes(expr.Value.(string))
		default:
			return NewMissing("ERR_TYPE"), nil
		}

	case "ref":
		if ctx == nil {
			return NewMissing("ERR_UNBOUND_VALUE"), nil
		}
		if expr.Namespace == "row" && ctx.Namespaces.Row == nil {
			return NewMissing("ERR_UNBOUND_VALUE"), nil
		}
		fullPath := "/" + expr.Namespace + expr.Pointer
		val, err := ctx.ResolvePointer(fullPath)
		if err != nil {
			return NewMissing("ERR_UNBOUND_VALUE"), nil
		}
		if val.Type == ValueTypeMissing {
			return val, nil
		}
		// Strict declared type check
		if val.Type != expr.ValueType {
			return NewMissing("ERR_TYPE"), nil
		}
		return val, nil

	case "eq":
		left, err := EvaluateExpression(expr.Left, ctx)
		if err != nil {
			return left, err
		}
		right, err := EvaluateExpression(expr.Right, ctx)
		if err != nil {
			return right, err
		}

		if left.Type == ValueTypeMissing {
			return left, nil
		}
		if right.Type == ValueTypeMissing {
			return right, nil
		}

		// Strict equal declared types: cross-type is ERR_TYPE / UNKNOWN
		if left.Type != right.Type {
			return NewMissing("ERR_TYPE"), nil
		}

		equal, _, err := CompareTaggedValues(left, right)
		if err != nil {
			return NewMissing("ERR_TYPE"), err
		}
		return NewBoolean(equal), nil

	case "lt", "lte":
		left, err := EvaluateExpression(expr.Left, ctx)
		if err != nil {
			return left, err
		}
		right, err := EvaluateExpression(expr.Right, ctx)
		if err != nil {
			return right, err
		}

		if left.Type == ValueTypeMissing {
			return left, nil
		}
		if right.Type == ValueTypeMissing {
			return right, nil
		}

		if left.Type != right.Type {
			return NewMissing("ERR_TYPE"), nil
		}

		switch left.Type {
		case ValueTypeInteger:
			lInt, err1 := strconv.ParseInt(left.Value.(string), 10, 64)
			rInt, err2 := strconv.ParseInt(right.Value.(string), 10, 64)
			if err1 != nil || err2 != nil {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
			}
			if expr.Op == "lt" {
				return NewBoolean(lInt < rInt), nil
			}
			return NewBoolean(lInt <= rInt), nil

		case ValueTypeDecimal:
			lRat := new(big.Rat)
			if _, ok := lRat.SetString(left.Value.(string)); !ok {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
			}
			rRat := new(big.Rat)
			if _, ok := rRat.SetString(right.Value.(string)); !ok {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
			}
			cmp := lRat.Cmp(rRat)
			if expr.Op == "lt" {
				return NewBoolean(cmp < 0), nil
			}
			return NewBoolean(cmp <= 0), nil

		case ValueTypeString:
			lStr := left.Value.(string)
			rStr := right.Value.(string)
			if expr.Op == "lt" {
				return NewBoolean(lStr < rStr), nil
			}
			return NewBoolean(lStr <= rStr), nil

		default:
			return NewMissing("ERR_TYPE"), nil
		}

	case "before", "at_or_before":
		left, err := EvaluateExpression(expr.Left, ctx)
		if err != nil {
			return left, err
		}
		right, err := EvaluateExpression(expr.Right, ctx)
		if err != nil {
			return right, err
		}

		if left.Type == ValueTypeMissing {
			return left, nil
		}
		if right.Type == ValueTypeMissing {
			return right, nil
		}

		if left.Type != ValueTypeTimestamp || right.Type != ValueTypeTimestamp {
			return NewMissing("ERR_TYPE"), nil
		}

		t1, err1 := time.Parse(time.RFC3339, left.Value.(string))
		t2, err2 := time.Parse(time.RFC3339, right.Value.(string))
		if err1 != nil || err2 != nil {
			return NewMissing("ERR_TYPE"), nil
		}

		if expr.Op == "before" {
			return NewBoolean(t1.Before(t2)), nil
		}
		return NewBoolean(t1.Before(t2) || t1.Equal(t2)), nil

	case "in":
		left, err := EvaluateExpression(expr.Left, ctx)
		if err != nil {
			return left, err
		}
		right, err := EvaluateExpression(expr.Right, ctx)
		if err != nil {
			return right, err
		}

		if left.Type == ValueTypeMissing {
			return left, nil
		}
		if right.Type == ValueTypeMissing {
			return right, nil
		}

		if right.Type != ValueTypeArray {
			return NewMissing("ERR_TYPE"), nil
		}

		arr, ok := right.Value.([]TaggedValue)
		if !ok {
			return NewMissing("ERR_TYPE"), nil
		}

		for _, item := range arr {
			eq, _, _ := CompareTaggedValues(left, item)
			if eq {
				return NewBoolean(true), nil
			}
		}
		return NewBoolean(false), nil

	case "add", "sub":
		left, err := EvaluateExpression(expr.Left, ctx)
		if err != nil {
			return left, err
		}
		right, err := EvaluateExpression(expr.Right, ctx)
		if err != nil {
			return right, err
		}

		if left.Type == ValueTypeMissing {
			return left, nil
		}
		if right.Type == ValueTypeMissing {
			return right, nil
		}

		if left.Type != right.Type {
			return NewMissing("ERR_TYPE"), nil
		}

		switch left.Type {
		case ValueTypeInteger:
			lInt, err1 := strconv.ParseInt(left.Value.(string), 10, 64)
			rInt, err2 := strconv.ParseInt(right.Value.(string), 10, 64)
			if err1 != nil || err2 != nil {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
			}
			var res int64
			if expr.Op == "add" {
				if (rInt > 0 && lInt > math.MaxInt64-rInt) || (rInt < 0 && lInt < math.MinInt64-rInt) {
					return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
				}
				res = lInt + rInt
			} else {
				if (rInt < 0 && lInt > math.MaxInt64+rInt) || (rInt > 0 && lInt < math.MinInt64+rInt) {
					return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
				}
				res = lInt - rInt
			}
			return NewInteger(strconv.FormatInt(res, 10))

		case ValueTypeDecimal:
			lRat := new(big.Rat)
			if _, ok := lRat.SetString(left.Value.(string)); !ok {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
			}
			rRat := new(big.Rat)
			if _, ok := rRat.SetString(right.Value.(string)); !ok {
				return NewMissing("ERR_UNSUPPORTED_CAPABILITY"), nil
			}
			resRat := new(big.Rat)
			if expr.Op == "add" {
				resRat.Add(lRat, rRat)
			} else {
				resRat.Sub(lRat, rRat)
			}
			// Exact rational result with up to 18 decimal places
			decStr := resRat.FloatString(18)
			projected := ProjectRawJSONNumber(decStr)
			return projected, nil

		default:
			return NewMissing("ERR_TYPE"), nil
		}

	case "and":
		// Left-to-right evaluation with short-circuit.
		// False ends and. Missing/wrong-type yields UNKNOWN.
		for _, arg := range expr.Args {
			v, err := EvaluateExpression(arg, ctx)
			if err != nil {
				return v, err
			}
			if v.Type != ValueTypeBoolean {
				return v, nil // UNKNOWN stops without evaluating unvisited operands
			}
			if v.Value.(bool) == false {
				return NewBoolean(false), nil // Short-circuit
			}
		}
		return NewBoolean(true), nil

	case "or":
		// Left-to-right evaluation with short-circuit.
		// True ends or. Missing/wrong-type yields UNKNOWN.
		for _, arg := range expr.Args {
			v, err := EvaluateExpression(arg, ctx)
			if err != nil {
				return v, err
			}
			if v.Type == ValueTypeBoolean && v.Value.(bool) == true {
				return NewBoolean(true), nil // Short-circuit
			}
			if v.Type != ValueTypeBoolean {
				return v, nil // UNKNOWN stops without evaluating unvisited operands
			}
		}
		return NewBoolean(false), nil

	case "not":
		v, err := EvaluateExpression(expr.Arg, ctx)
		if err != nil {
			return v, err
		}
		if v.Type == ValueTypeMissing {
			return v, nil
		}
		if v.Type != ValueTypeBoolean {
			return NewMissing("ERR_TYPE"), nil
		}
		return NewBoolean(!v.Value.(bool)), nil

	case "exists", "is_null", "is_missing":
		// Per BEHAVIOR-MODEL.md: exists/is_missing/is_null accept reference operands only
		if expr.Arg == nil || expr.Arg.Op != "ref" {
			return NewMissing("ERR_TYPE"), nil
		}
		fullPath := "/" + expr.Arg.Namespace + expr.Arg.Pointer
		val, err := ctx.ResolvePointer(fullPath)
		if err != nil {
			val = NewMissing("ERR_UNBOUND_VALUE")
		}

		switch expr.Op {
		case "exists":
			return NewBoolean(val.Type != ValueTypeMissing), nil
		case "is_missing":
			return NewBoolean(val.Type == ValueTypeMissing), nil
		case "is_null":
			return NewBoolean(val.Type == ValueTypeNull), nil
		}

	case "coerce":
		v, err := EvaluateExpression(expr.Arg, ctx)
		if err != nil {
			return v, err
		}
		if v.Type == ValueTypeMissing {
			return v, nil
		}
		switch expr.Profile {
		case "INTEGER_TO_DECIMAL":
			if v.Type != ValueTypeInteger {
				return NewMissing("ERR_TYPE"), nil
			}
			// Exact decimal representation of integer
			return NewDecimal(v.Value.(string))
		case "CANONICAL_STRING_TO_INTEGER":
			if v.Type != ValueTypeString {
				return NewMissing("ERR_TYPE"), nil
			}
			s := v.Value.(string)
			if !canonicalIntRegex.MatchString(s) {
				return NewMissing("ERR_TYPE"), nil
			}
			return NewInteger(s)
		default:
			return NewMissing("ERR_UNSUPPORTED_OPERATOR"), nil
		}

	case "array":
		items := make([]TaggedValue, len(expr.Items))
		for i, it := range expr.Items {
			val, err := EvaluateExpression(it, ctx)
			if err != nil {
				return val, err
			}
			items[i] = val
		}
		return NewArray(items), nil

	case "object":
		fields := make(map[string]TaggedValue, len(expr.Fields))
		for k, fe := range expr.Fields {
			val, err := EvaluateExpression(fe, ctx)
			if err != nil {
				return val, err
			}
			fields[k] = val
		}
		return NewObject(fields), nil

	default:
		return NewMissing("ERR_UNSUPPORTED_OPERATOR"), &SemanticsError{
			Code:   "ERR_UNSUPPORTED_OPERATOR",
			Reason: fmt.Sprintf("unsupported expression operator %q", expr.Op),
		}
	}

	return NewMissing("ERR_UNSUPPORTED_OPERATOR"), nil
}

// EvaluateTransitionGuard evaluates a transition guard against an EvaluationContext.
// Returns matched (bool), verdict ("PASS", "MISMATCH", "UNKNOWN"), and error.
func EvaluateTransitionGuard(guard *SemanticsExpression, ctx *EvaluationContext) (bool, string, error) {
	if guard == nil {
		return false, "UNKNOWN", &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "nil transition guard"}
	}
	val, err := EvaluateExpression(guard, ctx)
	if err != nil {
		return false, "UNKNOWN", err
	}
	if val.Type == ValueTypeBoolean {
		if b, ok := val.Value.(bool); ok {
			if b {
				return true, "PASS", nil
			}
			return false, "MISMATCH", nil
		}
	}
	return false, "UNKNOWN", nil
}

// SelectTransition evaluates ordered transitions after step execution.
// False transition guards advance to the next transition; an evaluated UNKNOWN stops without selecting a fallback.
func SelectTransition(transitions []StepTransition, ctx *EvaluationContext) (*StepTransition, string, error) {
	if len(transitions) == 0 {
		return nil, "UNKNOWN", &SemanticsError{Code: "ERR_INVALID_CONTRACT", Reason: "empty transition list"}
	}
	if len(transitions) > MaxTransitions {
		return nil, "UNKNOWN", &SemanticsError{Code: "ERR_UNSUPPORTED_CAPABILITY", Reason: "transitions exceed limit"}
	}

	for i := range transitions {
		matched, verdict, err := EvaluateTransitionGuard(transitions[i].Guard, ctx)
		if err != nil {
			return nil, "UNKNOWN", err
		}
		if verdict == "UNKNOWN" {
			// Evaluated UNKNOWN stops without selecting a fallback!
			return nil, "UNKNOWN", nil
		}
		if matched {
			return &transitions[i], "PASS", nil
		}
	}
	return nil, "MISMATCH", nil
}

// SelectBranch chooses the first branch whose guard evaluates to boolean true.
// An unknown earlier guard yields UNKNOWN and stops, never falling through.
func SelectBranch(branches []SemanticsBranch, ctx *EvaluationContext) (*SemanticsBranch, string, error) {
	for i := range branches {
		val, err := EvaluateExpression(branches[i].Guard, ctx)
		if err != nil {
			return nil, "UNKNOWN", err
		}
		if val.Type == ValueTypeMissing {
			return nil, "UNKNOWN", nil // Unknown earlier guard yields UNKNOWN, not fall-through
		}
		if val.Type != ValueTypeBoolean {
			return nil, "UNKNOWN", nil
		}
		if val.Value.(bool) == true {
			return &branches[i], "PASS", nil
		}
	}
	return nil, "MISMATCH", nil
}

// EvaluatePreconditions evaluates all preconditions against an EvaluationContext.
// Returns "PASS", "MISMATCH", or "UNKNOWN".
func EvaluatePreconditions(preconditions []*SemanticsExpression, ctx *EvaluationContext) (string, error) {
	for _, prec := range preconditions {
		val, err := EvaluateExpression(prec, ctx)
		if err != nil {
			return "UNKNOWN", err
		}
		if val.Type == ValueTypeMissing {
			return "UNKNOWN", nil
		}
		if val.Type != ValueTypeBoolean || val.Value.(bool) != true {
			return "MISMATCH", nil
		}
	}
	return "PASS", nil
}

// EvaluatePostconditions evaluates all postconditions against an EvaluationContext.
// Returns "PASS", "MISMATCH", or "UNKNOWN".
func EvaluatePostconditions(postconditions []*SemanticsExpression, ctx *EvaluationContext) (string, error) {
	for _, post := range postconditions {
		val, err := EvaluateExpression(post, ctx)
		if err != nil {
			return "UNKNOWN", err
		}
		if val.Type == ValueTypeMissing {
			return "UNKNOWN", nil
		}
		if val.Type != ValueTypeBoolean || val.Value.(bool) != true {
			return "MISMATCH", nil
		}
	}
	return "PASS", nil
}

// EvaluateInvariants evaluates all invariants against an EvaluationContext.
// Returns "PASS", "MISMATCH", or "UNKNOWN".
func EvaluateInvariants(invariants []*SemanticsExpression, ctx *EvaluationContext) (string, error) {
	for _, inv := range invariants {
		val, err := EvaluateExpression(inv, ctx)
		if err != nil {
			return "UNKNOWN", err
		}
		if val.Type == ValueTypeMissing {
			return "UNKNOWN", nil
		}
		if val.Type != ValueTypeBoolean || val.Value.(bool) != true {
			return "MISMATCH", nil
		}
	}
	return "PASS", nil
}

// ShadowState holds purely in-memory model simulation state.
// It has ZERO handles, credentials, or network connections to measured stores.
type ShadowState struct {
	DB                 map[string][]map[string]TaggedValue // "catalog.table" -> rows
	Sessions           map[string]ShadowSession
	Cache              map[string]TaggedValue
	StateSlots         map[string]TaggedValue
	Emissions          []ShadowEmission
	ActiveTransactions map[string]string // connection_id -> transaction_id
}

// ShadowSession holds in-memory session details during model simulation.
type ShadowSession struct {
	PrincipalTag    string
	PreviousID      string
	CurrentID       string
	Values          map[string]TaggedValue
	PreviousInvalid bool
}

// ShadowEmission records an emitted effect during model simulation.
type ShadowEmission struct {
	EffectClass    string
	DestinationID  string
	Payload        TaggedValue
	IdempotencyKey TaggedValue
}

// NewShadowState creates an isolated in-memory shadow simulation state.
func NewShadowState() *ShadowState {
	return &ShadowState{
		DB:                 make(map[string][]map[string]TaggedValue),
		Sessions:           make(map[string]ShadowSession),
		Cache:              make(map[string]TaggedValue),
		StateSlots:         make(map[string]TaggedValue),
		Emissions:          make([]ShadowEmission, 0),
		ActiveTransactions: make(map[string]string),
	}
}

// ActionResult records the outcome of executing a model action.
type ActionResult struct {
	Terminated bool
	Status     int
	ErrorCode  string
	Message    string
	Response   *EvaluationResponse
}

// ExecuteAction executes a single model action against the isolated ShadowState.
// It strictly refuses to touch measured application stores.
func (s *ShadowState) ExecuteAction(action *SemanticsAction, ctx *EvaluationContext) (*ActionResult, error) {
	if action == nil {
		return nil, errors.New("nil action")
	}

	switch action.Action {
	case "assert":
		val, err := EvaluateExpression(action.Predicate, ctx)
		if err != nil {
			return nil, err
		}
		if val.Type != ValueTypeBoolean || val.Value.(bool) != true {
			return &ActionResult{Terminated: true, ErrorCode: action.ErrorCode}, nil
		}
		return &ActionResult{}, nil

	case "set_state":
		val, err := EvaluateExpression(action.Value, ctx)
		if err != nil {
			return nil, err
		}
		s.StateSlots[action.Pointer] = val
		return &ActionResult{}, nil

	case "begin":
		if _, active := s.ActiveTransactions[action.ConnectionID]; active {
			return nil, fmt.Errorf("active transaction already exists on connection %s", action.ConnectionID)
		}
		if action.TransactionID == nil {
			return nil, errors.New("begin requires transaction_id")
		}
		s.ActiveTransactions[action.ConnectionID] = *action.TransactionID
		return &ActionResult{}, nil

	case "commit":
		active, exists := s.ActiveTransactions[action.ConnectionID]
		if !exists || action.TransactionID == nil || active != *action.TransactionID {
			return nil, fmt.Errorf("commit transaction_id mismatch on connection %s", action.ConnectionID)
		}
		delete(s.ActiveTransactions, action.ConnectionID)
		return &ActionResult{}, nil

	case "rollback":
		active, exists := s.ActiveTransactions[action.ConnectionID]
		if !exists || action.TransactionID == nil || active != *action.TransactionID {
			return nil, fmt.Errorf("rollback transaction_id mismatch on connection %s", action.ConnectionID)
		}
		delete(s.ActiveTransactions, action.ConnectionID)
		return &ActionResult{}, nil

	case "db_write":
		// Transaction check
		if action.TransactionID != nil {
			active, exists := s.ActiveTransactions[action.ConnectionID]
			if !exists || active != *action.TransactionID {
				return nil, fmt.Errorf("write transaction_id mismatch on connection %s", action.ConnectionID)
			}
		} else {
			if _, active := s.ActiveTransactions[action.ConnectionID]; active {
				return nil, fmt.Errorf("cannot autocommit write while transaction is active on connection %s", action.ConnectionID)
			}
		}

		key := action.Catalog + "." + action.Table

		switch action.Mutation {
		case "INSERT":
			// Per BEHAVIOR-MODEL.md: INSERT requires where=literal true and emits one row from typed values
			if action.Where == nil || action.Where.Op != "literal" || action.Where.ValueType != ValueTypeBoolean || action.Where.Value != true {
				return nil, errors.New("INSERT requires where=literal true")
			}
			newRow := make(map[string]TaggedValue, len(action.Values))
			for col, vexpr := range action.Values {
				v, err := EvaluateExpression(vexpr, ctx)
				if err != nil {
					return nil, err
				}
				newRow[col] = v
			}
			s.DB[key] = append(s.DB[key], newRow)

		case "UPDATE":
			// UPDATE requires non-empty values
			if len(action.Values) == 0 {
				return nil, errors.New("empty UPDATE is unsupported")
			}
			rows := s.DB[key]
			for _, r := range rows {
				// Evaluate where against candidate row
				subCtx := cloneContextWithRow(ctx, r)
				whereVal, err := EvaluateExpression(action.Where, subCtx)
				if err != nil {
					return nil, err
				}
				if whereVal.Type == ValueTypeBoolean && whereVal.Value.(bool) == true {
					// Evaluate UPDATE values against same pre-write row and apply together
					for col, vexpr := range action.Values {
						newVal, err := EvaluateExpression(vexpr, subCtx)
						if err != nil {
							return nil, err
						}
						r[col] = newVal
					}
				}
			}

		case "DELETE":
			if len(action.Values) > 0 {
				return nil, errors.New("DELETE requires empty values")
			}
			rows := s.DB[key]
			retained := make([]map[string]TaggedValue, 0, len(rows))
			for _, r := range rows {
				subCtx := cloneContextWithRow(ctx, r)
				whereVal, err := EvaluateExpression(action.Where, subCtx)
				if err != nil {
					return nil, err
				}
				if whereVal.Type != ValueTypeBoolean || whereVal.Value.(bool) == false {
					retained = append(retained, r)
				}
			}
			s.DB[key] = retained
		}
		return &ActionResult{}, nil

	case "emit":
		payloadVal, err := EvaluateExpression(action.Payload, ctx)
		if err != nil {
			return nil, err
		}
		idempKeyVal, err := EvaluateExpression(action.IdempotencyKey, ctx)
		if err != nil {
			return nil, err
		}
		s.Emissions = append(s.Emissions, ShadowEmission{
			EffectClass:    action.EffectClass,
			DestinationID:  action.DestinationID,
			Payload:        payloadVal,
			IdempotencyKey: idempKeyVal,
		})
		return &ActionResult{}, nil

	case "respond":
		bodyVal, err := EvaluateExpression(action.Body, ctx)
		if err != nil {
			return nil, err
		}
		headers := make([]EvaluationHeader, len(action.Headers))
		for i, h := range action.Headers {
			hv, err := EvaluateExpression(h.Value, ctx)
			if err != nil {
				return nil, err
			}
			headers[i] = EvaluationHeader{Name: h.Name, Value: fmt.Sprint(hv.Value)}
		}

		cookies := make([]EvaluationCookie, len(action.Cookies))
		for i, c := range action.Cookies {
			cv, err := EvaluateExpression(c.Value, ctx)
			if err != nil {
				return nil, err
			}
			cookies[i] = EvaluationCookie{
				Name:     c.Name,
				Value:    cv,
				Secure:   c.Secure,
				HTTPOnly: c.HttpOnly,
			}
		}

		return &ActionResult{
			Terminated: true,
			Status:     action.Status,
			Response: &EvaluationResponse{
				Status:     action.Status,
				Headers:    headers,
				Cookies:    cookies,
				Body:       bodyVal,
				EffectRefs: []string{},
			},
		}, nil

	case "error":
		msgVal, err := EvaluateExpression(action.Message, ctx)
		if err != nil {
			return nil, err
		}
		return &ActionResult{
			Terminated: true,
			ErrorCode:  action.Code,
			Message:    fmt.Sprint(msgVal.Value),
		}, nil

	case "session":
		princVal, err := EvaluateExpression(action.Principal, ctx)
		if err != nil {
			return nil, err
		}
		prevIdVal, err := EvaluateExpression(action.PreviousIdentity, ctx)
		if err != nil {
			return nil, err
		}
		currIdVal, err := EvaluateExpression(action.CurrentIdentity, ctx)
		if err != nil {
			return nil, err
		}

		vals := make(map[string]TaggedValue, len(action.Values))
		for k, ve := range action.Values {
			v, err := EvaluateExpression(ve, ctx)
			if err != nil {
				return nil, err
			}
			vals[k] = v
		}

		s.Sessions[fmt.Sprint(currIdVal.Value)] = ShadowSession{
			PrincipalTag:    fmt.Sprint(princVal.Value),
			PreviousID:      fmt.Sprint(prevIdVal.Value),
			CurrentID:       fmt.Sprint(currIdVal.Value),
			Values:          vals,
			PreviousInvalid: action.InvalidatePrevious,
		}
		return &ActionResult{}, nil
	}

	return nil, fmt.Errorf("unsupported action %q", action.Action)
}

func cloneContextWithRow(ctx *EvaluationContext, row map[string]TaggedValue) *EvaluationContext {
	if ctx == nil {
		return &EvaluationContext{
			Namespaces: EvaluationContextNamespaces{
				Row: row,
			},
		}
	}
	clone := *ctx
	clone.Namespaces = ctx.Namespaces
	clone.Namespaces.Row = row
	return &clone
}

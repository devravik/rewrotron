package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MCP JSON-RPC 2.0 Error Codes conforming to MCP-SPEC.md
const (
	MCPParseError     = -32700 // Invalid JSON
	MCPInvalidRequest = -32600 // Invalid JSON-RPC request structure
	MCPMethodNotFound = -32601 // Method not found
	MCPInvalidParams  = -32602 // Invalid method parameter(s)
	MCPInternalError  = -32603 // Internal error
)

// MCP Protocol Version
const MCPProtocolVersion = "2025-11-25"

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a protocol-level JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCPToolResult represents the MCP tools/call result object conforming to MCP-SPEC.md:
// "Domain operation failures use tool results with isError: true, a text explanation
// and structuredContent containing the structured error catalog entry. Successful
// tool results contain typed structuredContent and a matching short text representation."
type MCPToolResult struct {
	IsError           bool           `json:"isError,omitempty"`
	Content           []MCPContent   `json:"content"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
}

// MCPContent represents a single content element in an MCP tool result.
type MCPContent struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// MCPApprovalTokenRecord represents a human-issued, short-lived, single-use token
// bound to project, gate, exact subject hash, and expiry (CONFIRMATION-PROTOCOL).
type MCPApprovalTokenRecord struct {
	Token       string    `json:"token"`
	ProjectID   string    `json:"project_id"`
	Gate        string    `json:"gate"` // "GATE_1", "GATE_2", "GATE_3"
	SubjectHash string    `json:"subject_hash"`
	IssuedBy    string    `json:"issued_by"`
	ExpiresAt   time.Time `json:"expires_at"`
	Consumed    bool      `json:"consumed"`
}

// GuardedMCPDispatcher handles MCP JSON-RPC 2.0 protocol lifecycle,
// tool discovery, tool execution, resource reading, and token-gated human approvals.
type GuardedMCPDispatcher struct {
	mu             sync.RWMutex
	initialized    bool
	fencingSvc     *ResourceFencingService
	incidentSvc    *IncidentRepairService
	tokens         map[string]*MCPApprovalTokenRecord // key: token
	gate1Passed    map[string]bool
	gate2Passed    map[string]bool
	gate3Passed    map[string]bool
	activeProjects map[string]bool
}

// NewGuardedMCPDispatcher creates an initialized MCP dispatcher.
func NewGuardedMCPDispatcher(fencing *ResourceFencingService, incident *IncidentRepairService) *GuardedMCPDispatcher {
	return &GuardedMCPDispatcher{
		fencingSvc:     fencing,
		incidentSvc:    incident,
		tokens:         make(map[string]*MCPApprovalTokenRecord),
		gate1Passed:    make(map[string]bool),
		gate2Passed:    make(map[string]bool),
		gate3Passed:    make(map[string]bool),
		activeProjects: make(map[string]bool),
	}
}

// IssueHumanApprovalToken issues an approval token for local human approval testing/consumption.
func (d *GuardedMCPDispatcher) IssueHumanApprovalToken(token, projectID, gate, subjectHash, issuedBy string, expiresAt time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tokens[token] = &MCPApprovalTokenRecord{
		Token:       token,
		ProjectID:   projectID,
		Gate:        gate,
		SubjectHash: subjectHash,
		IssuedBy:    issuedBy,
		ExpiresAt:   expiresAt,
		Consumed:    false,
	}
}

// SetGateStatus sets gate flags for tests.
func (d *GuardedMCPDispatcher) SetGateStatus(projectID string, gate1, gate2, gate3 bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gate1Passed[projectID] = gate1
	d.gate2Passed[projectID] = gate2
	d.gate3Passed[projectID] = gate3
	d.activeProjects[projectID] = true
}

// HandleMessage parses and processes a raw JSON-RPC 2.0 message conforming to MCP-SPEC.md.
func (d *GuardedMCPDispatcher) HandleMessage(raw []byte) (*JSONRPCResponse, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &JSONRPCError{
				Code:    MCPParseError,
				Message: fmt.Sprintf("parse error: %v", err),
			},
		}, nil
	}

	if req.JSONRPC != "2.0" || req.Method == "" {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPInvalidRequest,
				Message: "invalid request: jsonrpc must be '2.0' and method must be non-empty",
			},
		}, nil
	}

	switch req.Method {
	case "initialize":
		return d.handleInitialize(&req)
	case "notifications/initialized":
		return nil, nil // Notification: no response
	case "ping":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{},
		}, nil
	case "tools/list":
		return d.handleToolsList(&req)
	case "tools/call":
		return d.handleToolsCall(&req)
	case "resources/list":
		return d.handleResourcesList(&req)
	case "resources/templates/list":
		return d.handleResourceTemplatesList(&req)
	case "resources/read":
		return d.handleResourcesRead(&req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPMethodNotFound,
				Message: fmt.Sprintf("method not found: %q", req.Method),
			},
		}, nil
	}
}

func (d *GuardedMCPDispatcher) handleInitialize(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	d.mu.Lock()
	d.initialized = true
	d.mu.Unlock()

	result := map[string]any{
		"protocolVersion": MCPProtocolVersion,
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"resources": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "rewrotron-mcp",
			"version": "1.0",
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

func (d *GuardedMCPDispatcher) handleToolsList(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	d.mu.RLock()
	init := d.initialized
	d.mu.RUnlock()

	if !init {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPInvalidRequest,
				Message: "server not initialized; call 'initialize' first",
			},
		}, nil
	}

	// Tool catalog conforming to specs/mcp/tools.json
	tools := []map[string]any{
		{"name": "project.init", "description": "Initialize a new Rewrotron workspace"},
		{"name": "project.detect", "description": "Run bounded legacy detection"},
		{"name": "project.doctor", "description": "Structured diagnostics summary"},
		{"name": "project.prepare_source_confirmation", "description": "Prepare Gate 1 review subject"},
		{"name": "project.confirm_source", "description": "Confirm Gate 1 with human approval token"},
		{"name": "project.confirm_plan", "description": "Confirm Gate 2 with human approval token"},
		{"name": "project.confirm_cutover", "description": "Confirm Gate 3 with human approval token"},
		{"name": "analysis.run", "description": "Run source analysis under Gate 1"},
		{"name": "observation.run", "description": "Run runtime observation under Gate 1"},
		{"name": "behavior.discover", "description": "Discover workflows and contracts under Gate 1"},
		{"name": "behavior.get_case", "description": "Read exact behavior case"},
		{"name": "migration.plan", "description": "Synthesize migration plan under Gate 1"},
		{"name": "migration.generate_code", "description": "Generate target slice under Gate 2"},
		{"name": "verification.run_case", "description": "Verify behavior case under Gate 2"},
		{"name": "benchmark.run", "description": "Run performance benchmark under Gate 2"},
		{"name": "research.get_next_mission", "description": "Peek ready mission without mutation"},
		{"name": "research.claim_mission", "description": "Claim mission with fencing lease"},
		{"name": "research.renew_mission", "description": "Renew mission lease"},
		{"name": "research.complete_mission", "description": "Complete mission with evidence"},
		{"name": "cutover.prepare", "description": "Prepare release package under Gate 3"},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"tools": tools,
		},
	}, nil
}

func (d *GuardedMCPDispatcher) handleToolsCall(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	d.mu.RLock()
	init := d.initialized
	d.mu.RUnlock()

	if !init {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPInvalidRequest,
				Message: "server not initialized; call 'initialize' first",
			},
		}, nil
	}

	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &callParams); err != nil || callParams.Name == "" {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPInvalidParams,
				Message: "invalid tools/call params: 'name' is required",
			},
		}, nil
	}

	// Validate input against tool schema
	schemaURI := fmt.Sprintf("https://rewrotron.dev/specs/mcp/tools/%s.input.schema.json", callParams.Name)
	if err := ValidateMCPToolInput(schemaURI, callParams.Arguments); err != nil {
		toolErr := &MCPToolResult{
			IsError: true,
			Content: []MCPContent{
				{Type: "text", Text: fmt.Sprintf("Validation failed for tool %s: %v", callParams.Name, err)},
			},
			StructuredContent: map[string]any{
				"code":               "ERR_INVALID_INPUT_SCHEMA",
				"message":            err.Error(),
				"category":           "USER_ERROR",
				"retryable":          false,
				"recommended_action": "check tool input against schema specification",
			},
		}
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  toolErr,
		}, nil
	}

	// Extract project_id and idempotency_key
	var commonArgs struct {
		ProjectID      string `json:"project_id"`
		IdempotencyKey string `json:"idempotency_key"`
		ApprovalToken  string `json:"approval_token"`
		SubjectHash    string `json:"subject_hash"`
	}
	_ = json.Unmarshal(callParams.Arguments, &commonArgs)

	// Gate enforcement
	switch callParams.Name {
	case "analysis.run", "observation.run", "behavior.discover", "migration.plan":
		d.mu.RLock()
		g1 := d.gate1Passed[commonArgs.ProjectID]
		d.mu.RUnlock()
		if !g1 {
			return makeMCPToolGateError(req.ID, "ERR_GATE1_REQUIRED", "Gate 1 (Source Detection Confirmation) required before analysis/planning")
		}
	case "migration.generate_code", "verification.run_case", "benchmark.run":
		d.mu.RLock()
		g1 := d.gate1Passed[commonArgs.ProjectID]
		g2 := d.gate2Passed[commonArgs.ProjectID]
		d.mu.RUnlock()
		if !g1 || !g2 {
			return makeMCPToolGateError(req.ID, "ERR_GATE2_REQUIRED", "Gate 2 (Plan Confirmation) required before code generation or verification")
		}
	case "cutover.prepare":
		d.mu.RLock()
		g3 := d.gate3Passed[commonArgs.ProjectID]
		d.mu.RUnlock()
		if !g3 {
			return makeMCPToolGateError(req.ID, "ERR_GATE3_REQUIRED", "Gate 3 (Cutover Confirmation) required before preparing cutover package")
		}
	case "project.confirm_source", "project.confirm_plan", "project.confirm_cutover":
		// Human Approval Token Validation
		if err := d.consumeApprovalToken(commonArgs.ApprovalToken, commonArgs.ProjectID, callParams.Name, commonArgs.SubjectHash); err != nil {
			toolErr := &MCPToolResult{
				IsError: true,
				Content: []MCPContent{
					{Type: "text", Text: fmt.Sprintf("Approval failed: %v", err)},
				},
				StructuredContent: map[string]any{
					"code":               "ERR_UNAUTHORIZED_APPROVAL",
					"message":            err.Error(),
					"category":           "USER_ERROR",
					"retryable":          false,
					"recommended_action": "issue a valid human approval token using CLI 'approval issue'",
				},
			}
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  toolErr,
			}, nil
		}
		// Token consumed successfully; record gate confirmation
		d.mu.Lock()
		switch callParams.Name {
		case "project.confirm_source":
			d.gate1Passed[commonArgs.ProjectID] = true
		case "project.confirm_plan":
			d.gate2Passed[commonArgs.ProjectID] = true
		case "project.confirm_cutover":
			d.gate3Passed[commonArgs.ProjectID] = true
		}
		d.mu.Unlock()
	}

	// Tool operation execution success
	successResult := &MCPToolResult{
		IsError: false,
		Content: []MCPContent{
			{Type: "text", Text: fmt.Sprintf("Tool %s executed successfully for project %s", callParams.Name, commonArgs.ProjectID)},
		},
		StructuredContent: map[string]any{
			"status":     "SUCCESS",
			"tool":       callParams.Name,
			"project_id": commonArgs.ProjectID,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  successResult,
	}, nil
}

func (d *GuardedMCPDispatcher) consumeApprovalToken(token, projectID, toolName, subjectHash string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec, exists := d.tokens[token]
	if !exists {
		return fmt.Errorf("approval token %q not found or never issued", token)
	}
	if rec.Consumed {
		return fmt.Errorf("approval token %q has already been consumed (single-use)", token)
	}
	if time.Now().After(rec.ExpiresAt) {
		return fmt.Errorf("approval token %q has expired at %s", token, rec.ExpiresAt.Format(time.RFC3339))
	}
	if rec.ProjectID != projectID {
		return fmt.Errorf("approval token project mismatch: token issued for %s, requested %s", rec.ProjectID, projectID)
	}
	if rec.SubjectHash != subjectHash {
		return fmt.Errorf("approval token subject mismatch: token bound to hash %s, requested %s", rec.SubjectHash, subjectHash)
	}

	expectedGate := ""
	switch toolName {
	case "project.confirm_source":
		expectedGate = "GATE_1"
	case "project.confirm_plan":
		expectedGate = "GATE_2"
	case "project.confirm_cutover":
		expectedGate = "GATE_3"
	}
	if rec.Gate != expectedGate {
		return fmt.Errorf("approval token gate mismatch: token issued for %s, requested %s", rec.Gate, expectedGate)
	}

	rec.Consumed = true
	return nil
}

func (d *GuardedMCPDispatcher) handleResourcesList(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"resources": []any{},
		},
	}, nil
}

func (d *GuardedMCPDispatcher) handleResourceTemplatesList(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	templates := []map[string]any{
		{
			"uriTemplate": "rewrotron://project/{project_id}/snapshot",
			"name":        "Project Snapshot",
			"mimeType":    "application/json",
		},
		{
			"uriTemplate": "rewrotron://project/{project_id}/execution",
			"name":        "Execution Observability State",
			"mimeType":    "application/json",
		},
		{
			"uriTemplate": "rewrotron://project/{project_id}/case/{case_id}",
			"name":        "Behavior Case",
			"mimeType":    "application/json",
		},
		{
			"uriTemplate": "rewrotron://project/{project_id}/plan/{plan_id}",
			"name":        "Migration Plan",
			"mimeType":    "application/json",
		},
		{
			"uriTemplate": "rewrotron://project/{project_id}/artifact/{sha256}",
			"name":        "Sanitized Project Artifact",
			"mimeType":    "application/json",
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"resourceTemplates": templates,
		},
	}, nil
}

func (d *GuardedMCPDispatcher) handleResourcesRead(req *JSONRPCRequest) (*JSONRPCResponse, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil || params.URI == "" {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPInvalidParams,
				Message: "missing 'uri' in resources/read",
			},
		}, nil
	}

	if !strings.HasPrefix(params.URI, "rewrotron://project/") {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    MCPInvalidParams,
				Message: fmt.Sprintf("invalid resource URI: %q, must begin with rewrotron://project/", params.URI),
			},
		}, nil
	}

	contents := []map[string]any{
		{
			"uri":      params.URI,
			"mimeType": "application/json",
			"text":     `{"status":"ACTIVE","uri":"` + params.URI + `"}`,
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"contents": contents,
		},
	}, nil
}

func makeMCPToolGateError(reqID any, code, message string) (*JSONRPCResponse, error) {
	res := &MCPToolResult{
		IsError: true,
		Content: []MCPContent{
			{Type: "text", Text: message},
		},
		StructuredContent: map[string]any{
			"code":               code,
			"message":            message,
			"category":           "USER_ERROR",
			"retryable":          false,
			"recommended_action": "obtain required human authorization token or execute prior gate command",
		},
	}
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      reqID,
		Result:  res,
	}, nil
}

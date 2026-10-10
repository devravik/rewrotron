package domain

import (
	"strings"
	"testing"
	"time"
)

func TestMCPDispatcher_ProtocolAndInvariants(t *testing.T) {
	fencingSvc := NewResourceFencingService()
	incidentSvc := NewIncidentRepairService()
	dispatcher := NewGuardedMCPDispatcher(fencingSvc, incidentSvc)

	// =========================================================================
	// Subtest 1: Protocol Lifecycle: reject tools/call before initialize
	// =========================================================================
	t.Run("RejectBeforeInitialize", func(t *testing.T) {
		reqBeforeInit := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
		resp, err := dispatcher.HandleMessage(reqBeforeInit)
		if err != nil {
			t.Fatalf("unexpected handler error: %v", err)
		}
		if resp.Error == nil || resp.Error.Code != MCPInvalidRequest {
			t.Fatalf("expected MCPInvalidRequest before initialize, got: %v", resp.Error)
		}

		// Malformed JSON returns -32700
		resp, _ = dispatcher.HandleMessage([]byte(`{not-json`))
		if resp.Error == nil || resp.Error.Code != MCPParseError {
			t.Fatalf("expected MCPParseError (-32700), got: %v", resp.Error)
		}
	})

	// =========================================================================
	// Subtest 2: Successful Initialization and Tool Listing
	// =========================================================================
	t.Run("InitializeAndToolsList", func(t *testing.T) {
		initReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)
		resp, err := dispatcher.HandleMessage(initReq)
		if err != nil || resp.Error != nil {
			t.Fatalf("initialize failed: %v", resp.Error)
		}
		initRes, ok := resp.Result.(map[string]any)
		if !ok || initRes["protocolVersion"] != MCPProtocolVersion {
			t.Fatalf("expected protocolVersion %s, got: %v", MCPProtocolVersion, initRes)
		}

		// tools/list now succeeds
		toolsReq := []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
		resp, err = dispatcher.HandleMessage(toolsReq)
		if err != nil || resp.Error != nil {
			t.Fatalf("tools/list failed: %v", resp.Error)
		}
		toolsRes, ok := resp.Result.(map[string]any)
		if !ok {
			t.Fatalf("expected tools list object, got: %v", resp.Result)
		}
		toolsList, ok := toolsRes["tools"].([]map[string]any)
		if !ok || len(toolsList) == 0 {
			t.Fatalf("expected non-empty tools list, got %v", toolsRes)
		}
	})

	// =========================================================================
	// Subtest 3: Gate Enforcement & Domain Error Representation (isError: true)
	// =========================================================================
	t.Run("GateEnforcementAndDomainErrors", func(t *testing.T) {
		projectID := "mcp_gate_proj"
		dispatcher.SetGateStatus(projectID, false, false, false)

		// analysis.run requires Gate 1; without Gate 1 returns tool error (isError: true, structuredContent)
		analysisReq := []byte(`{
			"jsonrpc":"2.0",
			"id":3,
			"method":"tools/call",
			"params":{
				"name":"analysis.run",
				"arguments":{
					"project_id":"` + projectID + `",
					"idempotency_key":"key_analysis_1"
				}
			}
		}`)
		resp, err := dispatcher.HandleMessage(analysisReq)
		if err != nil {
			t.Fatalf("HandleMessage error: %v", err)
		}
		if resp.Error != nil {
			t.Fatalf("unexpected JSON-RPC protocol error; domain failures must be tool results with isError: true: %v", resp.Error)
		}
		toolRes, ok := resp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError {
			t.Fatalf("expected isError: true for unapproved Gate 1, got %v", resp.Result)
		}
		if toolRes.StructuredContent["code"] != "ERR_GATE1_REQUIRED" {
			t.Fatalf("expected ERR_GATE1_REQUIRED in structuredContent, got: %v", toolRes.StructuredContent)
		}
	})

	// =========================================================================
	// Subtest 4: Token-Gated Human Approval Invariants
	// =========================================================================
	t.Run("TokenGatedHumanApprovals", func(t *testing.T) {
		projectID := "mcp_token_proj"
		dispatcher.SetGateStatus(projectID, false, false, false)
		subjectHash := strings.Repeat("a", 64)
		token := "tok_human_single_use_01"

		// Issue valid single-use token bound to Gate 1 and subjectHash
		dispatcher.IssueHumanApprovalToken(
			token,
			projectID,
			"GATE_1",
			subjectHash,
			"human_operator",
			time.Now().Add(10*time.Minute),
		)

		// Call with wrong token: rejected
		wrongTokenReq := []byte(`{
			"jsonrpc":"2.0",
			"id":4,
			"method":"tools/call",
			"params":{
				"name":"project.confirm_source",
				"arguments":{
					"project_id":"` + projectID + `",
					"subject_hash":"` + subjectHash + `",
					"approval_token":"wrong_token",
					"idempotency_key":"key_confirm_01"
				}
			}
		}`)
		resp, _ := dispatcher.HandleMessage(wrongTokenReq)
		toolRes, ok := resp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError || toolRes.StructuredContent["code"] != "ERR_UNAUTHORIZED_APPROVAL" {
			t.Fatalf("expected ERR_UNAUTHORIZED_APPROVAL for wrong token, got: %v", resp.Result)
		}

		// Call with valid token: consumes token and confirms Gate 1
		validTokenReq := []byte(`{
			"jsonrpc":"2.0",
			"id":5,
			"method":"tools/call",
			"params":{
				"name":"project.confirm_source",
				"arguments":{
					"project_id":"` + projectID + `",
					"subject_hash":"` + subjectHash + `",
					"approval_token":"` + token + `",
					"idempotency_key":"key_confirm_02"
				}
			}
		}`)
		resp, _ = dispatcher.HandleMessage(validTokenReq)
		toolRes, ok = resp.Result.(*MCPToolResult)
		if !ok || toolRes.IsError {
			t.Fatalf("expected successful approval with valid token, got: %v", resp.Result)
		}

		// Replay of same token: rejected as already consumed (single-use token invariant)
		replayReq := []byte(`{
			"jsonrpc":"2.0",
			"id":6,
			"method":"tools/call",
			"params":{
				"name":"project.confirm_source",
				"arguments":{
					"project_id":"` + projectID + `",
					"subject_hash":"` + subjectHash + `",
					"approval_token":"` + token + `",
					"idempotency_key":"key_confirm_03"
				}
			}
		}`)
		resp, _ = dispatcher.HandleMessage(replayReq)
		toolRes, ok = resp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError || !strings.Contains(toolRes.StructuredContent["message"].(string), "already been consumed") {
			t.Fatalf("expected rejection of replayed single-use token, got: %v", resp.Result)
		}
	})

	// =========================================================================
	// Subtest 5: Resource Templates and Reading
	// =========================================================================
	t.Run("ResourceTemplatesAndReading", func(t *testing.T) {
		templatesReq := []byte(`{"jsonrpc":"2.0","id":7,"method":"resources/templates/list"}`)
		resp, err := dispatcher.HandleMessage(templatesReq)
		if err != nil || resp.Error != nil {
			t.Fatalf("resources/templates/list failed: %v", resp.Error)
		}
		res, ok := resp.Result.(map[string]any)
		if !ok {
			t.Fatalf("expected resource templates result map")
		}
		templates, ok := res["resourceTemplates"].([]map[string]any)
		if !ok || len(templates) < 5 {
			t.Fatalf("expected at least 5 resource templates, got %d", len(templates))
		}

		// Read resource
		readReq := []byte(`{"jsonrpc":"2.0","id":8,"method":"resources/read","params":{"uri":"rewrotron://project/p1/snapshot"}}`)
		resp, err = dispatcher.HandleMessage(readReq)
		if err != nil || resp.Error != nil {
			t.Fatalf("resources/read failed: %v", resp.Error)
		}
	})
}

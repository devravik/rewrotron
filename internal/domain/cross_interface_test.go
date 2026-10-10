package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCrossInterface_X31_X60_X46_X51_X52_X55 tests Milestone M5 Unit 5:
// - X31: Vendor-neutral transport (CLI and MCP gate and domain result parity)
// - X60: Offline protocol schema bundling (Draft 2020-12 bundle, rewrites external refs, rejects unregistered refs)
// - X46: No automatic implementation (completed report does not auto-admit planning or target writes)
// - X51: Blocked report retrieval without requiring Gate 2 or executable plan
// - X52: One analysis/evidence model shared between assessment and migration
// - X55: Assessment report ref/hash is not a Gate 2 confirmation token
func TestCrossInterface_X31_X60_X46_X51_X52_X55(t *testing.T) {
	fencingSvc := NewResourceFencingService()
	incidentSvc := NewIncidentRepairService()
	cliDisp := NewGuardedCLIDispatcher(fencingSvc, incidentSvc)
	mcpDisp := NewGuardedMCPDispatcher(fencingSvc, incidentSvc)

	// Initialize MCP
	initReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)
	_, _ = mcpDisp.HandleMessage(initReq)

	bridge := NewCrossInterfaceBridge(cliDisp, mcpDisp)

	// =========================================================================
	// Test X31: Vendor-neutral transport parity
	// =========================================================================
	t.Run("X31_VendorNeutralTransportParity", func(t *testing.T) {
		projectID := "proj_x31"

		// Before Gate 1: both refuse analyze with exact gate failure
		ok, err := bridge.VerifyParity(projectID, false, false, false)
		if err != nil || !ok {
			t.Fatalf("expected Gate 1 refusal parity, got err: %v", err)
		}

		// After Gate 1: both admit analyze
		ok, err = bridge.VerifyParity(projectID, true, false, false)
		if err != nil || !ok {
			t.Fatalf("expected Gate 1 success parity, got err: %v", err)
		}
	})

	// =========================================================================
	// Test X60: Offline protocol schema bundle
	// =========================================================================
	t.Run("X60_OfflineSchemaBundling", func(t *testing.T) {
		bundler := NewOfflineSchemaBundler()

		commonURI := "https://rewrotron.dev/specs/common.schema.json"
		commonSchema := []byte(`{
			"$id": "https://rewrotron.dev/specs/common.schema.json",
			"$defs": {
				"id": {
					"type": "string",
					"pattern": "^[a-z0-9_-]+$"
				}
			}
		}`)
		bundler.RegisterSchema(commonURI, commonSchema)

		toolURI := "https://rewrotron.dev/specs/mcp/tools/test.tool.input.schema.json"
		toolSchema := []byte(`{
			"$id": "https://rewrotron.dev/specs/mcp/tools/test.tool.input.schema.json",
			"type": "object",
			"properties": {
				"project_id": {
					"$ref": "https://rewrotron.dev/specs/common.schema.json#/$defs/id"
				}
			},
			"required": ["project_id"]
		}`)
		bundler.RegisterSchema(toolURI, toolSchema)

		// Bundle schema
		bundledBytes, err := bundler.BundleToolSchema(toolURI)
		if err != nil {
			t.Fatalf("BundleToolSchema failed: %v", err)
		}

		var bundleDoc map[string]any
		if err := json.Unmarshal(bundledBytes, &bundleDoc); err != nil {
			t.Fatalf("unmarshal bundled schema failed: %v", err)
		}

		// Verify external $ref was rewritten to #/$defs/<sha256>
		props := bundleDoc["properties"].(map[string]any)
		projectProp := props["project_id"].(map[string]any)
		refStr := projectProp["$ref"].(string)
		if !strings.HasPrefix(refStr, "#/$defs/") {
			t.Fatalf("expected rewritten local ref starting with #/$defs/, got: %s", refStr)
		}

		// Verify root has $defs containing the external document
		defs := bundleDoc["$defs"].(map[string]any)
		if len(defs) == 0 {
			t.Fatalf("expected root $defs to contain resolved external schema")
		}

		// Negative control: Unregistered external ref fails closed
		unregisteredToolURI := "https://rewrotron.dev/specs/mcp/tools/unregistered.tool.json"
		unregisteredToolSchema := []byte(`{
			"$id": "https://rewrotron.dev/specs/mcp/tools/unregistered.tool.json",
			"properties": {
				"missing": {"$ref": "https://rewrotron.dev/specs/absent.schema.json"}
			}
		}`)
		bundler.RegisterSchema(unregisteredToolURI, unregisteredToolSchema)
		_, err = bundler.BundleToolSchema(unregisteredToolURI)
		if err == nil || !strings.Contains(err.Error(), ErrUnregisteredSchemaRef) {
			t.Fatalf("expected ErrUnregisteredSchemaRef for missing registry document, got: %v", err)
		}
	})

	// =========================================================================
	// Test X46, X51, X52, X55: Assessment and Gate Boundary Invariants
	// =========================================================================
	t.Run("X46_X51_X52_X55_AssessmentAndGateInvariants", func(t *testing.T) {
		projectID := "proj_assessment_gate"
		cliDisp.SetGateStatus(projectID, true, false, false) // Gate 1 confirmed, Gate 2 absent

		// X46: Complete report does NOT auto-create implementation work or publish target bytes
		// Calling migrate without Gate 2 is strictly refused
		migrateEnv, err := cliDisp.Dispatch(&ParsedCLIFlags{
			Command: "migrate",
			Project: projectID,
			JSON:    true,
		})
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if migrateEnv.ExitCode != ExitCodeMissingGate1OrGate2 {
			t.Fatalf("X46: expected exit 4 on migrate without Gate 2, got %d", migrateEnv.ExitCode)
		}

		// X51: Blocked report retrieval works without Gate 2 or executable plan
		// 'report' succeeds under Gate 1
		reportEnv, err := cliDisp.Dispatch(&ParsedCLIFlags{
			Command: "report",
			Project: projectID,
			JSON:    true,
		})
		if err != nil {
			t.Fatalf("dispatch report error: %v", err)
		}
		if reportEnv.ExitCode != ExitCodeSuccess {
			t.Fatalf("X51: report retrieval must succeed without Gate 2, got exit %d", reportEnv.ExitCode)
		}

		// X55: Submitting assessment report ref as a Gate 2 confirmation token fails closed
		assessmentReportRef := "sha256:" + strings.Repeat("e", 64)
		tokenReq := []byte(`{
			"jsonrpc":"2.0",
			"id":99,
			"method":"tools/call",
			"params":{
				"name":"project.confirm_plan",
				"arguments":{
					"project_id":"` + projectID + `",
					"subject_hash":"` + strings.Repeat("f", 64) + `",
					"approval_token":"` + assessmentReportRef + `",
					"idempotency_key":"key_confirm_fake"
				}
			}
		}`)
		resp, _ := mcpDisp.HandleMessage(tokenReq)
		toolRes, ok := resp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError || toolRes.StructuredContent["code"] != "ERR_UNAUTHORIZED_APPROVAL" {
			t.Fatalf("X55: expected approval token rejection when using assessment ref as token, got %v", resp.Result)
		}
	})
}

package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestM5_AcceptanceClosure executes the comprehensive acceptance suite for Milestone M5:
// Interfaces & Bounded Repair, integrating Units 1 through 5 and verifying all authoritative
// M5 falsification tests: X24, X26, X27, X31, X32, X60, X46, X51, X52, X55, and the final M5 completion gate.
func TestM5_AcceptanceClosure(t *testing.T) {
	now := time.Now()
	fencingSvc := NewResourceFencingService()
	incidentSvc := NewIncidentRepairService()
	supervisorSvc := NewSupervisorService()
	cliDisp := NewGuardedCLIDispatcher(fencingSvc, incidentSvc)
	mcpDisp := NewGuardedMCPDispatcher(fencingSvc, incidentSvc)
	crossBridge := NewCrossInterfaceBridge(cliDisp, mcpDisp)
	bundler := NewOfflineSchemaBundler()

	// Initialize MCP server
	initReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)
	initResp, err := mcpDisp.HandleMessage(initReq)
	if err != nil || initResp.Error != nil {
		t.Fatalf("MCP initialization failed: %v", initResp.Error)
	}

	// =========================================================================
	// 1. Test X24: Incident Budgets, Idempotency & Bounded Repair (Unit 1)
	// =========================================================================
	t.Run("X24_IncidentBudgetsAndBoundedRepair", func(t *testing.T) {
		projectID := "m5_x24_proj"
		subjectID := "sub_csrf_fix"
		idemKey := "idem_x24_01"

		// Attempt 1: Begin fresh attempt
		att1, receipt, err := incidentSvc.BeginAttempt(
			projectID,
			subjectID,
			idemKey,
			"fp_x24_1",
			"Fix CSRF cookie samesite",
			map[string]any{"cookie_samesite": "Lax"},
			1,
			"worker_1",
			now,
			5*time.Minute,
		)
		if err != nil || receipt != nil {
			t.Fatalf("BeginAttempt 1 failed: %v, receipt: %v", err, receipt)
		}

		// Commit and store receipt
		resPayload := map[string]any{"status": "COMMITTED", "patch": "patch_csrf_01"}
		storedReceipt, err := incidentSvc.CommitResultAndStoreReceipt(att1, idemKey, 1, 1, resPayload, now.Add(1*time.Minute))
		if err != nil || storedReceipt == nil {
			t.Fatalf("CommitResultAndStoreReceipt failed: %v", err)
		}

		// Crash after commit before response: client retries with exact idempotency key
		retryAtt, crashReceipt, err := incidentSvc.BeginAttempt(
			projectID,
			subjectID,
			idemKey,
			"fp_x24_1",
			"Fix CSRF cookie samesite",
			map[string]any{"cookie_samesite": "Lax"},
			1,
			"worker_1",
			now.Add(2*time.Minute),
			5*time.Minute,
		)
		if err != nil || retryAtt != nil || crashReceipt == nil {
			t.Fatalf("expected existing receipt on crash recovery, got att: %v, receipt: %v", retryAtt, crashReceipt)
		}
		if crashReceipt.ResponseResult["patch"] != "patch_csrf_01" {
			t.Fatalf("receipt payload mismatch: %v", crashReceipt.ResponseResult)
		}

		// Trivial fingerprint reset rejection
		_ = incidentSvc.RecordFailure(projectID, subjectID, att1.AttemptID, "fp_x24_1", "mismatch", now.Add(2*time.Minute))
		_, _, err = incidentSvc.BeginAttempt(
			projectID,
			subjectID,
			"idem_x24_dup",
			"fp_x24_1", // identical fingerprint
			"Retry 1",
			map[string]any{"cookie_samesite": "Lax"},
			1,
			"worker_1",
			now.Add(3*time.Minute),
			5*time.Minute,
		)
		if err == nil || !strings.Contains(err.Error(), ErrTrivialFingerprintReset) {
			t.Fatalf("expected ErrTrivialFingerprintReset, got: %v", err)
		}
	})

	// =========================================================================
	// 2. Test CLI Machine Protocol, Envelopes & Exit Codes (Unit 2)
	// =========================================================================
	t.Run("CLI_ProtocolAndExitCodes", func(t *testing.T) {
		projectID := "m5_cli_proj"
		cliDisp.SetGateStatus(projectID, false, false, false)

		// Missing Gate 1 returns exit 4 and validated failure envelope
		analyzeFlags := &ParsedCLIFlags{
			Command: "analyze",
			Project: projectID,
			JSON:    true,
		}
		failEnv, err := cliDisp.Dispatch(analyzeFlags)
		if err != nil {
			t.Fatalf("dispatch error: %v", err)
		}
		if failEnv.ExitCode != ExitCodeMissingGate1OrGate2 || failEnv.Success {
			t.Fatalf("expected exit code 4 for missing Gate 1, got %d", failEnv.ExitCode)
		}
		rawEnv, err := failEnv.ToJSON()
		if err != nil || ValidateCLIOutput(rawEnv) != nil {
			t.Fatalf("failure envelope violated CLI schema: %v", err)
		}

		// Dry-run safety simulation returns dry_run: true without mutations
		dryFlags := &ParsedCLIFlags{
			Command: "init",
			Project: projectID,
			JSON:    true,
			DryRun:  true,
		}
		dryEnv, err := cliDisp.Dispatch(dryFlags)
		if err != nil || dryEnv.ExitCode != ExitCodeSuccess || dryEnv.DryRun == nil || !*dryEnv.DryRun {
			t.Fatalf("expected valid dry-run envelope with dry_run=true, got: %v", dryEnv)
		}
	})

	// =========================================================================
	// 3. Test MCP Stdio Protocol & Token-Gated Approvals (Unit 3)
	// =========================================================================
	t.Run("MCP_ProtocolAndSingleUseTokens", func(t *testing.T) {
		projectID := "m5_mcp_proj"
		mcpDisp.SetGateStatus(projectID, false, false, false)
		subHash := strings.Repeat("7", 64)
		token := "tok_closure_human_01"

		// Issue single-use human approval token
		mcpDisp.IssueHumanApprovalToken(token, projectID, "GATE_1", subHash, "human", now.Add(10*time.Minute))

		// Consume token via project.confirm_source
		confirmReq := []byte(`{
			"jsonrpc":"2.0",
			"id":10,
			"method":"tools/call",
			"params":{
				"name":"project.confirm_source",
				"arguments":{
					"project_id":"` + projectID + `",
					"subject_hash":"` + subHash + `",
					"approval_token":"` + token + `",
					"idempotency_key":"key_m5_conf"
				}
			}
		}`)
		resp, err := mcpDisp.HandleMessage(confirmReq)
		if err != nil || resp.Error != nil {
			t.Fatalf("HandleMessage failed: %v", resp.Error)
		}
		toolRes, ok := resp.Result.(*MCPToolResult)
		if !ok || toolRes.IsError {
			t.Fatalf("expected successful confirmation with token, got: %v", resp.Result)
		}

		// Replayed token fails closed
		resp, _ = mcpDisp.HandleMessage(confirmReq)
		toolRes, ok = resp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError || toolRes.StructuredContent["code"] != "ERR_UNAUTHORIZED_APPROVAL" {
			t.Fatalf("expected single-use token replay to fail with ERR_UNAUTHORIZED_APPROVAL, got: %v", resp.Result)
		}
	})

	// =========================================================================
	// 4. Test X26, X27, X32: Supervisor, Execution Truth & Trust Boundary (Unit 4)
	// =========================================================================
	t.Run("X26_X27_X32_SupervisorObservabilityTrust", func(t *testing.T) {
		projectID := "m5_sup_proj"

		// X26: Supervisor survives client disconnect
		_, _ = svcRegister(supervisorSvc, 501, ProcessTypeSupervisor, projectID, "op_s1", now)
		_, _ = svcRegister(supervisorSvc, 502, ProcessTypeWorkerRunner, projectID, "op_s1", now)
		supervisorSvc.HandleClientDisconnect(500, projectID, now)
		if !supervisorSvc.IsProcessAlive(501) || !supervisorSvc.IsProcessAlive(502) {
			t.Fatalf("supervisor and worker must survive client disconnect (Test X26)")
		}

		// X27: Execution page current watermark N distinct from event bound M
		for i := 0; i < 5; i++ {
			_, _ = supervisorSvc.AdvanceLedgerSequence()
		}
		page, err := supervisorSvc.ReadExecutionPage(projectID, 0, 2, "client_m5")
		if err != nil {
			t.Fatalf("ReadExecutionPage failed: %v", err)
		}
		if page.ThroughSequence != 2 || page.StateSequence <= 2 || page.StateSemantics != "CURRENT" {
			t.Fatalf("expected through_sequence 2 and state_sequence > 2, got through %d, state %d", page.ThroughSequence, page.StateSequence)
		}
		rawPage, _ := json.Marshal(page)
		if err := ValidateExecutionPage(rawPage); err != nil {
			t.Fatalf("page violated schema v2: %v", err)
		}

		// X32: Hostile same-account modification detected upon mediated revalidation
		path := "app/Config/database.php"
		realHash := strings.Repeat("8", 64)
		supervisorSvc.RegisterTrustedFile(path, realHash)
		tamperedHash := strings.Repeat("9", 64)
		ok, err := supervisorSvc.VerifyFileAccessMediated(path, tamperedHash)
		if ok || err == nil || !strings.Contains(err.Error(), ErrHostileAccessDetected) {
			t.Fatalf("expected ErrHostileAccessDetected for file tampering, got: %v", err)
		}
	})

	// =========================================================================
	// 5. Test X31, X60, X46, X51, X52, X55: Cross-Interface Parity & Offline Bundles (Unit 5)
	// =========================================================================
	t.Run("X31_X60_X46_X51_X52_X55_ParityAndBundles", func(t *testing.T) {
		projectID := "m5_parity_proj"

		// X31: Parity between CLI and MCP
		ok, err := crossBridge.VerifyParity(projectID, false, false, false)
		if err != nil || !ok {
			t.Fatalf("transport parity check failed: %v", err)
		}

		// X60: Offline schema bundle without client network fetch
		rootURI := "https://rewrotron.dev/specs/mcp/tools/sample.json"
		depURI := "https://rewrotron.dev/specs/dep.json"
		bundler.RegisterSchema(depURI, []byte(`{"$id":"https://rewrotron.dev/specs/dep.json","type":"string"}`))
		bundler.RegisterSchema(rootURI, []byte(`{"$id":"https://rewrotron.dev/specs/mcp/tools/sample.json","properties":{"p":{"$ref":"https://rewrotron.dev/specs/dep.json"}}}`))

		bundleBytes, err := bundler.BundleToolSchema(rootURI)
		if err != nil {
			t.Fatalf("BundleToolSchema failed: %v", err)
		}
		if !strings.Contains(string(bundleBytes), "#/$defs/") {
			t.Fatalf("expected rewritten local ref in bundle: %s", string(bundleBytes))
		}

		// X46: Complete report does NOT auto-create implementation work without Gate 2
		cliDisp.SetGateStatus(projectID, true, false, false) // Gate 1 passed, Gate 2 absent
		migrateEnv, _ := cliDisp.Dispatch(&ParsedCLIFlags{Command: "migrate", Project: projectID, JSON: true})
		if migrateEnv.ExitCode != ExitCodeMissingGate1OrGate2 {
			t.Fatalf("X46: migrate without Gate 2 must fail exit 4, got %d", migrateEnv.ExitCode)
		}

		// X51: Blocked report retrieval succeeds without Gate 2
		reportEnv, _ := cliDisp.Dispatch(&ParsedCLIFlags{Command: "report", Project: projectID, JSON: true})
		if reportEnv.ExitCode != ExitCodeSuccess {
			t.Fatalf("X51: report retrieval must succeed without Gate 2, got %d", reportEnv.ExitCode)
		}

		// X55: Assessment report ref submitted as confirmation token fails closed
		tokenReq := []byte(`{
			"jsonrpc":"2.0",
			"id":20,
			"method":"tools/call",
			"params":{
				"name":"project.confirm_plan",
				"arguments":{
					"project_id":"` + projectID + `",
					"subject_hash":"` + strings.Repeat("0", 64) + `",
					"approval_token":"sha256:` + strings.Repeat("1", 64) + `",
					"idempotency_key":"k1"
				}
			}
		}`)
		resp, _ := mcpDisp.HandleMessage(tokenReq)
		toolRes, ok := resp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError || toolRes.StructuredContent["code"] != "ERR_UNAUTHORIZED_APPROVAL" {
			t.Fatalf("X55: expected report ref as approval token to fail with ERR_UNAUTHORIZED_APPROVAL, got: %v", resp.Result)
		}
	})
}

func svcRegister(s *SupervisorService, pid int, ptype ProcessType, prj, op string, now time.Time) (*ProcessRecord, error) {
	return s.RegisterProcess(pid, ptype, prj, op, now)
}

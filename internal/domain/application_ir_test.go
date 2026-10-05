package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func validApplicationIRBytes() []byte {
	validDM := validDataModelV2Bytes()
	validWF := []byte(`{
		"schema_version": "2.0",
		"workflow_id": "checkout",
		"profile_ref": "sha256:` + strings.Repeat("a", 64) + `",
		"initial_state_ref": "sha256:` + strings.Repeat("a", 64) + `",
		"actors": [
			{
				"actor_id": "a",
				"principal_fixture_ref": "sha256:` + strings.Repeat("a", 64) + `"
			}
		],
		"start_step_id": "login",
		"steps": [
			{
				"step_id": "login",
				"actor_id": "a",
				"operation_id": "login",
				"request_ref": "sha256:` + strings.Repeat("a", 64) + `",
				"guard": {
					"op": "literal",
					"value_type": "boolean",
					"value": true
				},
				"bindings": {},
				"transitions": [
					{
						"guard": {
							"op": "literal",
							"value_type": "boolean",
							"value": true
						},
						"next_step_id": null,
						"terminal_id": "done"
					}
				],
				"completion": {
					"mode": "HTTP_RESPONSE",
					"timeout_ms": 30000,
					"expected_receipts": 0,
					"quiescence_ms": 0,
					"predicate": {
						"op": "literal",
						"value_type": "boolean",
						"value": true
					}
				}
			}
		],
		"terminals": [
			{
				"terminal_id": "done",
				"outcome": "SUCCESS",
				"state_assertions": [
					{
						"op": "literal",
						"value_type": "boolean",
						"value": true
					}
				],
				"expected_effect_refs": []
			}
		],
		"max_executed_steps": 2,
		"evidence_refs": [
			"sha256:` + strings.Repeat("a", 64) + `"
		]
	}`)

	doc := ApplicationIRDocument{
		SchemaVersion:     "1.0",
		ProjectID:         "proj-ci3-test",
		ConfigurationHash: strings.Repeat("a", 64),
		DocumentID:        "doc-app-ir-1",
		SourceIRRef:       "sha256:" + strings.Repeat("b", 64),
		Entrypoints: []ApplicationEntrypoint{
			{
				EntrypointID: "ep-http-welcome",
				Type:         EntrypointTypeHTTP,
				Dispatch: json.RawMessage(`{
					"type": "http",
					"method": "GET",
					"path": "/welcome"
				}`),
				HandlerSymbolIDs: []string{"sym-welcome-index"},
				EvidenceIDs:      []string{"ev-app-1"},
				CoverageState:    CoverageStateResolved,
			},
			{
				EntrypointID: "ep-cli-seed",
				Type:         EntrypointTypeCLI,
				Dispatch: json.RawMessage(`{
					"type": "cli",
					"command_id": "seed-db"
				}`),
				HandlerSymbolIDs: []string{"sym-seed"},
				EvidenceIDs:      []string{"ev-app-1"},
				CoverageState:    CoverageStateResolved,
			},
			{
				EntrypointID: "ep-cron-daily",
				Type:         EntrypointTypeCron,
				Dispatch: json.RawMessage(`{
					"type": "cron",
					"command_id": "daily-report",
					"schedule": "0 0 * * *",
					"timezone": "UTC"
				}`),
				HandlerSymbolIDs: []string{"sym-daily"},
				EvidenceIDs:      []string{"ev-app-1"},
				CoverageState:    CoverageStateResolved,
			},
			{
				EntrypointID: "ep-event-email",
				Type:         EntrypointTypeEvent,
				Dispatch: json.RawMessage(`{
					"type": "event",
					"channel": "notifications.email",
					"message_schema": {}
				}`),
				HandlerSymbolIDs: []string{"sym-email"},
				EvidenceIDs:      []string{"ev-app-1"},
				CoverageState:    CoverageStateResolved,
			},
		},
		Components: []ApplicationComponent{
			{
				ComponentID: "comp-handler-1",
				Role:        ComponentRoleHandler,
				SymbolIDs:   []string{"sym-welcome-index"},
				EvidenceIDs: []string{"ev-app-1"},
			},
			{
				ComponentID: "comp-service-1",
				Role:        ComponentRoleService,
				SymbolIDs:   []string{"sym-welcome-service"},
				EvidenceIDs: []string{"ev-app-1"},
			},
		},
		Dependencies: []ApplicationDependency{
			{
				EdgeID:      "dep-1",
				FromID:      "comp-handler-1",
				ToID:        ptrStr("comp-service-1"),
				Kind:        "invokes",
				Resolved:    true,
				EvidenceIDs: []string{"ev-app-1"},
			},
		},
		DataModels: []json.RawMessage{validDM},
		Workflows:  []json.RawMessage{validWF},
		Findings: []ApplicationFinding{
			{
				FindingID:   "find-app-1",
				Code:        "DYNAMIC_DISPATCH",
				Severity:    FindingSeverityMedium,
				AffectedIDs: []string{"comp-handler-1"},
				EvidenceIDs: []string{"ev-app-1"},
				Message:     "Dynamic dispatch in handler component",
			},
		},
		Evidence: []ApplicationEvidence{
			{
				EvidenceID:  "ev-app-1",
				Kind:        "STATIC",
				ArtifactRef: "sha256:" + strings.Repeat("c", 64),
				SymbolID:    ptrStr("sym-welcome-index"),
				Description: "Component binding analysis",
			},
		},
	}

	raw, _ := json.Marshal(doc)
	return raw
}

func TestValidateApplicationIR_Positive(t *testing.T) {
	raw := validApplicationIRBytes()
	if err := ValidateApplicationIR(raw); err != nil {
		t.Fatalf("expected valid application IR, got: %v", err)
	}

	var doc ApplicationIRDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cjson, err := doc.CanonicalJSON()
	if err != nil {
		t.Fatalf("canonical JSON failed: %v", err)
	}
	if len(cjson) == 0 {
		t.Fatal("expected non-empty canonical JSON")
	}

	digest, err := doc.Digest()
	if err != nil {
		t.Fatalf("digest failed: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Fatalf("invalid digest: %s", digest)
	}
}

func TestValidateApplicationIR_Negative(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m map[string]interface{})
		wantErr string
	}{
		{
			name: "invalid schema_version",
			mutate: func(m map[string]interface{}) {
				m["schema_version"] = "2.0"
			},
			wantErr: "schema_version",
		},
		{
			name: "invalid project_id",
			mutate: func(m map[string]interface{}) {
				m["project_id"] = "BAD_ID!"
			},
			wantErr: "project_id",
		},
		{
			name: "invalid source_ir_ref",
			mutate: func(m map[string]interface{}) {
				m["source_ir_ref"] = "not-a-sha"
			},
			wantErr: "source_ir_ref",
		},
		{
			name: "unexpected root field",
			mutate: func(m map[string]interface{}) {
				m["unexpected"] = "value"
			},
			wantErr: "unexpected",
		},
		{
			name: "missing required root field",
			mutate: func(m map[string]interface{}) {
				delete(m, "entrypoints")
			},
			wantErr: "missing",
		},
		{
			name: "duplicate entrypoint_id",
			mutate: func(m map[string]interface{}) {
				eps := m["entrypoints"].([]interface{})
				eps = append(eps, eps[0])
				m["entrypoints"] = eps
			},
			wantErr: "duplicate entrypoint_id",
		},
		{
			name: "invalid entrypoint type",
			mutate: func(m map[string]interface{}) {
				eps := m["entrypoints"].([]interface{})
				ep := eps[0].(map[string]interface{})
				ep["type"] = "grpc"
			},
			wantErr: "invalid entrypoint type",
		},
		{
			name: "mismatched entrypoint type and dispatch type",
			mutate: func(m map[string]interface{}) {
				eps := m["entrypoints"].([]interface{})
				ep := eps[0].(map[string]interface{})
				ep["type"] = "cli"
			},
			wantErr: "does not match dispatch type",
		},
		{
			name: "entrypoint unknown evidence_id",
			mutate: func(m map[string]interface{}) {
				eps := m["entrypoints"].([]interface{})
				ep := eps[0].(map[string]interface{})
				ep["evidence_ids"] = []interface{}{"unknown-ev"}
			},
			wantErr: "not found in evidence",
		},
		{
			name: "invalid component role",
			mutate: func(m map[string]interface{}) {
				comps := m["components"].([]interface{})
				c := comps[0].(map[string]interface{})
				c["role"] = "wizard"
			},
			wantErr: "invalid component role",
		},
		{
			name: "duplicate component_id",
			mutate: func(m map[string]interface{}) {
				comps := m["components"].([]interface{})
				comps = append(comps, comps[0])
				m["components"] = comps
			},
			wantErr: "duplicate component_id",
		},
		{
			name: "dependency unknown evidence_id",
			mutate: func(m map[string]interface{}) {
				deps := m["dependencies"].([]interface{})
				d := deps[0].(map[string]interface{})
				d["evidence_ids"] = []interface{}{"unknown-ev"}
			},
			wantErr: "not found in evidence",
		},
		{
			name: "invalid data model in data_models",
			mutate: func(m map[string]interface{}) {
				m["data_models"] = []interface{}{
					map[string]interface{}{
						"schema_version": "2.0",
						"model_id":       "bad-model",
					},
				}
			},
			wantErr: "invalid data model in application IR",
		},
		{
			name: "duplicate data_model model_id",
			mutate: func(m map[string]interface{}) {
				dms := m["data_models"].([]interface{})
				dms = append(dms, dms[0])
				m["data_models"] = dms
			},
			wantErr: "duplicate data_model model_id",
		},
		{
			name: "invalid workflow in workflows",
			mutate: func(m map[string]interface{}) {
				m["workflows"] = []interface{}{
					map[string]interface{}{
						"schema_version": "2.0",
						"workflow_id":    "bad-wf",
					},
				}
			},
			wantErr: "invalid workflow in application IR",
		},
		{
			name: "duplicate workflow_id",
			mutate: func(m map[string]interface{}) {
				wfs := m["workflows"].([]interface{})
				wfs = append(wfs, wfs[0])
				m["workflows"] = wfs
			},
			wantErr: "duplicate workflow_id",
		},
		{
			name: "duplicate finding_id",
			mutate: func(m map[string]interface{}) {
				findings := m["findings"].([]interface{})
				findings = append(findings, findings[0])
				m["findings"] = findings
			},
			wantErr: "duplicate finding_id",
		},
		{
			name: "duplicate evidence_id",
			mutate: func(m map[string]interface{}) {
				evs := m["evidence"].([]interface{})
				evs = append(evs, evs[0])
				m["evidence"] = evs
			},
			wantErr: "duplicate evidence_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := validApplicationIRBytes()
			var m map[string]interface{}
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("unmarshal base failed: %v", err)
			}
			tc.mutate(m)
			mutated, _ := json.Marshal(m)

			err := ValidateApplicationIR(mutated)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

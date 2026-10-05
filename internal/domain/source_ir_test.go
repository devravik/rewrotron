package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func validSourceIRBytes() []byte {
	doc := SourceIRDocument{
		SchemaVersion:     "1.0",
		ProjectID:         "proj-ci3-test",
		ConfigurationHash: strings.Repeat("a", 64),
		DocumentID:        "doc-source-ir-1",
		Files: []SourceFile{
			{
				FileID:       "file-1",
				Path:         "application/controllers/Welcome.php",
				SHA256:       strings.Repeat("b", 64),
				SyntaxStatus: SyntaxStatusParsed,
			},
		},
		Symbols: []SourceSymbol{
			{
				SymbolID:  "sym-welcome-class",
				FileID:    "file-1",
				Name:      "Welcome",
				Kind:      SymbolKindClass,
				StartLine: 3,
				EndLine:   25,
			},
			{
				SymbolID:  "sym-welcome-index",
				FileID:    "file-1",
				Name:      "index",
				Kind:      SymbolKindMethod,
				StartLine: 10,
				EndLine:   20,
			},
		},
		Edges: []SourceEdge{
			{
				EdgeID:      "edge-1",
				FromID:      "sym-welcome-class",
				ToID:        ptrStr("sym-welcome-index"),
				Kind:        "declares",
				Resolved:    true,
				EvidenceIDs: []string{"ev-static-1"},
			},
		},
		Findings: []SourceFinding{
			{
				FindingID:   "find-1",
				Code:        "DYNAMIC_DISPATCH_UNKNOWN",
				Severity:    FindingSeverityMedium,
				AffectedIDs: []string{"sym-welcome-index"},
				EvidenceIDs: []string{"ev-static-1"},
				Message:     "Dynamic method dispatch detected in controller action",
			},
		},
		Evidence: []SourceEvidence{
			{
				EvidenceID:  "ev-static-1",
				Kind:        "STATIC",
				ArtifactRef: "sha256:" + strings.Repeat("c", 64),
				SymbolID:    ptrStr("sym-welcome-index"),
				Description: "AST traversal extracted declared methods and call sites",
			},
		},
	}

	raw, _ := json.Marshal(doc)
	return raw
}

func ptrStr(s string) *string {
	return &s
}

func TestValidateSourceIR_Positive(t *testing.T) {
	raw := validSourceIRBytes()
	if err := ValidateSourceIR(raw); err != nil {
		t.Fatalf("expected valid source IR, got: %v", err)
	}

	var doc SourceIRDocument
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

func TestValidateSourceIR_Negative(t *testing.T) {
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
				m["project_id"] = "INVALID PROJ ID!"
			},
			wantErr: "project_id",
		},
		{
			name: "invalid configuration_hash",
			mutate: func(m map[string]interface{}) {
				m["configuration_hash"] = "short-hash"
			},
			wantErr: "configuration_hash",
		},
		{
			name: "unexpected root field",
			mutate: func(m map[string]interface{}) {
				m["extra_field"] = "not_allowed"
			},
			wantErr: "unexpected",
		},
		{
			name: "missing required root field",
			mutate: func(m map[string]interface{}) {
				delete(m, "evidence")
			},
			wantErr: "missing",
		},
		{
			name: "duplicate file_id",
			mutate: func(m map[string]interface{}) {
				files := m["files"].([]interface{})
				files = append(files, files[0])
				m["files"] = files
			},
			wantErr: "duplicate file_id",
		},
		{
			name: "invalid syntax_status",
			mutate: func(m map[string]interface{}) {
				files := m["files"].([]interface{})
				f := files[0].(map[string]interface{})
				f["syntax_status"] = "MAGIC"
			},
			wantErr: "invalid syntax_status",
		},
		{
			name: "symbol unknown file_id",
			mutate: func(m map[string]interface{}) {
				symbols := m["symbols"].([]interface{})
				s := symbols[0].(map[string]interface{})
				s["file_id"] = "unknown-file"
			},
			wantErr: "not found in files",
		},
		{
			name: "symbol invalid line range",
			mutate: func(m map[string]interface{}) {
				symbols := m["symbols"].([]interface{})
				s := symbols[0].(map[string]interface{})
				s["start_line"] = 50
				s["end_line"] = 10
			},
			wantErr: "greater than end_line",
		},
		{
			name: "edge unknown from_id",
			mutate: func(m map[string]interface{}) {
				edges := m["edges"].([]interface{})
				e := edges[0].(map[string]interface{})
				e["from_id"] = "unknown-id"
			},
			wantErr: "not found in files or symbols",
		},
		{
			name: "edge unknown evidence_id",
			mutate: func(m map[string]interface{}) {
				edges := m["edges"].([]interface{})
				e := edges[0].(map[string]interface{})
				e["evidence_ids"] = []interface{}{"unknown-ev"}
			},
			wantErr: "not found in evidence",
		},
		{
			name: "finding unknown evidence_id",
			mutate: func(m map[string]interface{}) {
				findings := m["findings"].([]interface{})
				f := findings[0].(map[string]interface{})
				f["evidence_ids"] = []interface{}{"unknown-ev"}
			},
			wantErr: "not found in evidence",
		},
		{
			name: "finding invalid severity",
			mutate: func(m map[string]interface{}) {
				findings := m["findings"].([]interface{})
				f := findings[0].(map[string]interface{})
				f["severity"] = "CATASTROPHIC"
			},
			wantErr: "invalid finding severity",
		},
		{
			name: "evidence invalid kind",
			mutate: func(m map[string]interface{}) {
				evidence := m["evidence"].([]interface{})
				ev := evidence[0].(map[string]interface{})
				ev["kind"] = "GUESS"
			},
			wantErr: "invalid evidence kind",
		},
		{
			name: "evidence unknown symbol_id",
			mutate: func(m map[string]interface{}) {
				evidence := m["evidence"].([]interface{})
				ev := evidence[0].(map[string]interface{})
				ev["symbol_id"] = "unknown-symbol"
			},
			wantErr: "not found in symbols",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := validSourceIRBytes()
			var m map[string]interface{}
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("unmarshal base failed: %v", err)
			}
			tc.mutate(m)
			mutated, _ := json.Marshal(m)

			err := ValidateSourceIR(mutated)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

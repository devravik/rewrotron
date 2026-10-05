package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func validDataModelV2Bytes() []byte {
	dm := DataModel{
		SchemaVersion:     "2.0",
		ModelID:           "model-users",
		ConnectionID:      "conn-main",
		Catalog:           "app_db",
		Schema:            nil,
		ServerBehaviorRef: "sha256:" + strings.Repeat("a", 64),
		Table:             "users",
		Columns: []DataModelColumn{
			{
				Name:           "id",
				Type:           "bigint",
				Nullable:       false,
				ValueType:      "integer",
				TypeMappingRef: ptrStr("sha256:" + strings.Repeat("b", 64)),
			},
			{
				Name:           "email",
				Type:           "varchar(255)",
				Nullable:       false,
				ValueType:      "string",
				TypeMappingRef: ptrStr("sha256:" + strings.Repeat("b", 64)),
			},
			{
				Name:           "custom_blob",
				Type:           "blob",
				Nullable:       true,
				ValueType:      "UNKNOWN",
				TypeMappingRef: nil,
			},
		},
		PrimaryKey: []string{"id"},
		UniqueKeys: [][]string{{"email"}},
		Relationships: []DataModelRelationship{
			{
				Kind:           "OBSERVED_LOGICAL",
				FromColumns:    []string{"id"},
				ToConnectionID: "conn-audit",
				ToCatalog:      "audit_db",
				ToSchema:       nil,
				ToTable:        "user_audit",
				ToColumns:      []string{"user_id"},
				EvidenceRefs:   []string{"sha256:" + strings.Repeat("c", 64)},
			},
		},
		EvidenceIDs: []string{"ev-db-1"},
	}

	raw, _ := json.Marshal(dm)
	return raw
}

func TestValidateDataModel_ValidationCases(t *testing.T) {
	raw, err := os.ReadFile("../../specs/examples/validation-cases.json")
	if err != nil {
		t.Fatalf("failed to read validation cases: %v", err)
	}

	var cases []struct {
		Name     string          `json:"name"`
		Schema   string          `json:"schema"`
		Instance json.RawMessage `json:"instance"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal validation cases: %v", err)
	}

	targetSchema := "https://rewrotron.dev/specs/ir/data-model.schema.json"
	count := 0
	for _, tc := range cases {
		if tc.Schema != targetSchema {
			continue
		}
		count++
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateDataModel(tc.Instance)
			if tc.Valid && err != nil {
				t.Fatalf("expected valid data model %q to pass, got: %v", tc.Name, err)
			}
			if !tc.Valid && err == nil {
				t.Fatalf("expected invalid data model %q to fail, but it passed", tc.Name)
			}
		})
	}

	if count < 7 {
		t.Fatalf("expected at least 7 data-model validation cases, found %d", count)
	}
}

func TestValidateDataModel_Positive(t *testing.T) {
	raw := validDataModelV2Bytes()
	if err := ValidateDataModel(raw); err != nil {
		t.Fatalf("expected valid data model, got: %v", err)
	}

	var dm DataModel
	if err := json.Unmarshal(raw, &dm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cjson, err := dm.CanonicalJSON()
	if err != nil {
		t.Fatalf("canonical JSON failed: %v", err)
	}
	if len(cjson) == 0 {
		t.Fatal("expected non-empty canonical JSON")
	}

	digest, err := dm.Digest()
	if err != nil {
		t.Fatalf("digest failed: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != 71 {
		t.Fatalf("invalid digest: %s", digest)
	}
}

func TestValidateDataModel_V1(t *testing.T) {
	v1Doc := map[string]interface{}{
		"model_id": "legacy_table",
		"table":    "legacy_items",
		"columns": []map[string]interface{}{
			{
				"name":     "id",
				"type":     "int(11)",
				"nullable": false,
			},
		},
		"primary_key":   []string{"id"},
		"unique_keys":   [][]string{},
		"relationships": []map[string]interface{}{},
		"evidence_ids":  []string{"ev-legacy-1"},
	}

	raw, _ := json.Marshal(v1Doc)
	if err := ValidateDataModel(raw); err != nil {
		t.Fatalf("expected valid v1 data model to pass, got: %v", err)
	}

	// Negative v1: missing primary_key
	delete(v1Doc, "primary_key")
	rawBad, _ := json.Marshal(v1Doc)
	if err := ValidateDataModel(rawBad); err == nil {
		t.Fatal("expected missing primary_key to fail")
	}
}

func TestValidateDataModel_Negative(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m map[string]interface{})
		wantErr string
	}{
		{
			name: "invalid schema_version",
			mutate: func(m map[string]interface{}) {
				m["schema_version"] = "3.0"
			},
			wantErr: "unsupported schema_version",
		},
		{
			name: "missing connection_id",
			mutate: func(m map[string]interface{}) {
				delete(m, "connection_id")
			},
			wantErr: "missing required v2 field: connection_id",
		},
		{
			name: "unexpected field",
			mutate: func(m map[string]interface{}) {
				m["unexpected"] = "value"
			},
			wantErr: "unexpected field",
		},
		{
			name: "duplicate column name",
			mutate: func(m map[string]interface{}) {
				cols := m["columns"].([]interface{})
				cols = append(cols, cols[0])
				m["columns"] = cols
			},
			wantErr: "duplicate column name",
		},
		{
			name: "primary_key column not in columns",
			mutate: func(m map[string]interface{}) {
				m["primary_key"] = []interface{}{"nonexistent_col"}
			},
			wantErr: "not found in columns",
		},
		{
			name: "invalid column value_type",
			mutate: func(m map[string]interface{}) {
				cols := m["columns"].([]interface{})
				c := cols[0].(map[string]interface{})
				c["value_type"] = "FLOAT64"
			},
			wantErr: "invalid column value_type",
		},
		{
			name: "unknown value_type with non-null type_mapping_ref",
			mutate: func(m map[string]interface{}) {
				cols := m["columns"].([]interface{})
				c := cols[2].(map[string]interface{}) // custom_blob
				c["type_mapping_ref"] = "sha256:" + strings.Repeat("b", 64)
			},
			wantErr: "type_mapping_ref must be null when value_type is UNKNOWN",
		},
		{
			name: "known value_type with null type_mapping_ref",
			mutate: func(m map[string]interface{}) {
				cols := m["columns"].([]interface{})
				c := cols[0].(map[string]interface{}) // id
				c["type_mapping_ref"] = nil
			},
			wantErr: "type_mapping_ref must be sha256 reference when value_type is not UNKNOWN",
		},
		{
			name: "relationship missing to_connection_id",
			mutate: func(m map[string]interface{}) {
				rels := m["relationships"].([]interface{})
				r := rels[0].(map[string]interface{})
				delete(r, "to_connection_id")
			},
			wantErr: "missing required relationship field",
		},
		{
			name: "relationship from_columns not in columns",
			mutate: func(m map[string]interface{}) {
				rels := m["relationships"].([]interface{})
				r := rels[0].(map[string]interface{})
				r["from_columns"] = []interface{}{"ghost_col"}
			},
			wantErr: "not found in columns",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := validDataModelV2Bytes()
			var m map[string]interface{}
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("unmarshal base failed: %v", err)
			}
			tc.mutate(m)
			mutated, _ := json.Marshal(m)

			err := ValidateDataModel(mutated)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

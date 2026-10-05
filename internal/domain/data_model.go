package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

var allowedColumnValueTypes = map[string]struct{}{
	"string":    {},
	"boolean":   {},
	"integer":   {},
	"decimal":   {},
	"timestamp": {},
	"bytes":     {},
	"object":    {},
	"array":     {},
	"UNKNOWN":   {},
}

var allowedRelationshipKinds = map[string]struct{}{
	"FOREIGN_KEY":      {},
	"OBSERVED_LOGICAL": {},
	"INFERRED_LOGICAL": {},
}

// DataModelError describes an invalid Data Model record without echoing private contents.
type DataModelError struct {
	Reason string
	Cause  error
}

// Error identifies the failed data model requirement.
func (e *DataModelError) Error() string {
	return "invalid data model: " + e.Reason
}

// Unwrap preserves decoding and canonicalization failures for inspection.
func (e *DataModelError) Unwrap() error { return e.Cause }

// DataModelColumn describes a column in a v2 database model.
type DataModelColumn struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Nullable       bool    `json:"nullable"`
	Precision      *int    `json:"precision,omitempty"`
	Scale          *int    `json:"scale,omitempty"`
	ValueType      string  `json:"value_type"`
	TypeMappingRef *string `json:"type_mapping_ref"`
}

// DataModelColumnV1 describes a column in a historical v1 database model.
type DataModelColumnV1 struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Nullable  bool   `json:"nullable"`
	Precision *int   `json:"precision,omitempty"`
	Scale     *int   `json:"scale,omitempty"`
}

// DataModelRelationship describes a foreign key or observed/inferred logical relationship.
type DataModelRelationship struct {
	Kind           string   `json:"kind"`
	FromColumns    []string `json:"from_columns"`
	ToConnectionID string   `json:"to_connection_id"`
	ToCatalog      string   `json:"to_catalog"`
	ToSchema       *string  `json:"to_schema"`
	ToTable        string   `json:"to_table"`
	ToColumns      []string `json:"to_columns"`
	EvidenceRefs   []string `json:"evidence_refs"`
}

// DataModel represents a database model conforming to specs/ir/data-model.schema.json (v2 qualified).
type DataModel struct {
	SchemaVersion     string                  `json:"schema_version"`
	ModelID           string                  `json:"model_id"`
	ConnectionID      string                  `json:"connection_id"`
	Catalog           string                  `json:"catalog"`
	Schema            *string                 `json:"schema"`
	ServerBehaviorRef string                  `json:"server_behavior_ref"`
	Table             string                  `json:"table"`
	Columns           []DataModelColumn       `json:"columns"`
	PrimaryKey        []string                `json:"primary_key"`
	UniqueKeys        [][]string              `json:"unique_keys"`
	Relationships     []DataModelRelationship `json:"relationships"`
	Collation         *string                 `json:"collation,omitempty"`
	Timezone          *string                 `json:"timezone,omitempty"`
	EvidenceIDs       []string                `json:"evidence_ids"`
}

// DataModelV1 represents a historical unversioned data model conforming to specs/ir/data-model.schema.json.
type DataModelV1 struct {
	ModelID       string              `json:"model_id"`
	Table         string              `json:"table"`
	Columns       []DataModelColumnV1 `json:"columns"`
	PrimaryKey    []string            `json:"primary_key"`
	UniqueKeys    [][]string          `json:"unique_keys"`
	Relationships []json.RawMessage   `json:"relationships"`
	Collation     *string             `json:"collation,omitempty"`
	Timezone      *string             `json:"timezone,omitempty"`
	EvidenceIDs   []string            `json:"evidence_ids"`
}

// CanonicalJSON renders the DataModel into canonical JSON bytes.
func (dm *DataModel) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(dm)
	if err != nil {
		return nil, &DataModelError{Reason: "serialization failure", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the sha256 digest of the canonical DataModel.
func (dm *DataModel) Digest() (string, error) {
	cjson, err := dm.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cjson)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// ValidateDataModel validates raw bytes against specs/ir/data-model.schema.json (v2.0 or historical v1).
func ValidateDataModel(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &DataModelError{Reason: "canonical JSON", Cause: err}
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &root); err != nil || root == nil {
		return &DataModelError{Reason: "object required", Cause: err}
	}

	if verRaw, ok := root["schema_version"]; ok {
		var ver string
		if err := json.Unmarshal(verRaw, &ver); err != nil {
			return &DataModelError{Reason: "schema_version must be string", Cause: err}
		}
		if ver == "2.0" {
			return validateDataModelV2(root)
		}
		return &DataModelError{Reason: fmt.Sprintf("unsupported schema_version %q", ver)}
	}

	return validateDataModelV1(root)
}

func validateDataModelV2(root map[string]json.RawMessage) error {
	expectedFields := []string{
		"model_id",
		"table",
		"columns",
		"primary_key",
		"unique_keys",
		"relationships",
		"evidence_ids",
		"schema_version",
		"connection_id",
		"catalog",
		"schema",
		"server_behavior_ref",
	}

	for _, f := range expectedFields {
		v, ok := root[f]
		if !ok {
			return &DataModelError{Reason: "missing required v2 field: " + f}
		}
		if f != "schema" && string(v) == "null" {
			return &DataModelError{Reason: "v2 field cannot be null: " + f}
		}
	}

	for k := range root {
		switch k {
		case "model_id", "table", "columns", "primary_key", "unique_keys", "relationships",
			"evidence_ids", "schema_version", "connection_id", "catalog", "schema",
			"server_behavior_ref", "collation", "timezone":
		default:
			return &DataModelError{Reason: "unexpected field in v2 data model: " + k}
		}
	}

	var modelID string
	if err := json.Unmarshal(root["model_id"], &modelID); err != nil || !validAssessmentReportID(modelID) {
		return &DataModelError{Reason: "model_id must match pattern ^[a-z0-9_-]+$"}
	}

	var connectionID string
	if err := json.Unmarshal(root["connection_id"], &connectionID); err != nil || !validAssessmentReportID(connectionID) {
		return &DataModelError{Reason: "connection_id must match pattern ^[a-z0-9_-]+$"}
	}

	var catalog string
	if err := json.Unmarshal(root["catalog"], &catalog); err != nil || len(catalog) < 1 {
		return &DataModelError{Reason: "catalog must be a non-empty string"}
	}

	if string(root["schema"]) != "null" {
		var schemaName string
		if err := json.Unmarshal(root["schema"], &schemaName); err != nil || len(schemaName) < 1 {
			return &DataModelError{Reason: "schema must be null or a non-empty string"}
		}
	}

	if !validProfileRef(root["server_behavior_ref"], false) {
		return &DataModelError{Reason: "server_behavior_ref must match pattern ^sha256:[a-f0-9]{64}$"}
	}

	var table string
	if err := json.Unmarshal(root["table"], &table); err != nil || len(table) < 1 {
		return &DataModelError{Reason: "table must be a non-empty string"}
	}

	colNames, err := validateDataModelColumnsV2(root["columns"])
	if err != nil {
		return err
	}

	if err := validateDataModelKeys(root["primary_key"], root["unique_keys"], colNames); err != nil {
		return err
	}

	if err := validateDataModelRelationshipsV2(root["relationships"], colNames); err != nil {
		return err
	}

	if err := validateDataModelOptionalStrings(root); err != nil {
		return err
	}

	if err := validateDataModelEvidenceIDs(root["evidence_ids"]); err != nil {
		return err
	}

	return nil
}

func validateDataModelV1(root map[string]json.RawMessage) error {
	expectedFields := []string{
		"model_id",
		"table",
		"columns",
		"primary_key",
		"unique_keys",
		"relationships",
		"evidence_ids",
	}

	for _, f := range expectedFields {
		v, ok := root[f]
		if !ok || string(v) == "null" {
			return &DataModelError{Reason: "missing required v1 field: " + f}
		}
	}

	for k := range root {
		switch k {
		case "model_id", "table", "columns", "primary_key", "unique_keys", "relationships",
			"evidence_ids", "collation", "timezone":
		default:
			return &DataModelError{Reason: "unexpected field in v1 data model: " + k}
		}
	}

	var modelID string
	if err := json.Unmarshal(root["model_id"], &modelID); err != nil || !validAssessmentReportID(modelID) {
		return &DataModelError{Reason: "model_id must match pattern ^[a-z0-9_-]+$"}
	}

	var table string
	if err := json.Unmarshal(root["table"], &table); err != nil || len(table) < 1 {
		return &DataModelError{Reason: "table must be a non-empty string"}
	}

	colNames, err := validateDataModelColumnsV1(root["columns"])
	if err != nil {
		return err
	}

	if err := validateDataModelKeys(root["primary_key"], root["unique_keys"], colNames); err != nil {
		return err
	}

	var rels []json.RawMessage
	if err := json.Unmarshal(root["relationships"], &rels); err != nil {
		return &DataModelError{Reason: "relationships must be an array", Cause: err}
	}
	for _, rawRel := range rels {
		var r map[string]json.RawMessage
		if err := json.Unmarshal(rawRel, &r); err != nil || r == nil {
			return &DataModelError{Reason: "relationship entry must be an object", Cause: err}
		}
	}

	if err := validateDataModelOptionalStrings(root); err != nil {
		return err
	}

	if err := validateDataModelEvidenceIDs(root["evidence_ids"]); err != nil {
		return err
	}

	return nil
}

func validateDataModelColumnsV2(raw json.RawMessage) (map[string]struct{}, error) {
	var columns []json.RawMessage
	if err := json.Unmarshal(raw, &columns); err != nil {
		return nil, &DataModelError{Reason: "columns must be an array", Cause: err}
	}

	seenColNames := make(map[string]struct{}, len(columns))
	for _, rawCol := range columns {
		var c map[string]json.RawMessage
		if err := json.Unmarshal(rawCol, &c); err != nil || c == nil {
			return nil, &DataModelError{Reason: "column entry must be an object", Cause: err}
		}

		for _, req := range []string{"name", "type", "nullable", "value_type", "type_mapping_ref"} {
			v, ok := c[req]
			if !ok {
				return nil, &DataModelError{Reason: "missing required column field: " + req}
			}
			if req != "type_mapping_ref" && string(v) == "null" {
				return nil, &DataModelError{Reason: "column field cannot be null: " + req}
			}
		}

		for k := range c {
			switch k {
			case "name", "type", "nullable", "precision", "scale", "value_type", "type_mapping_ref":
			default:
				return nil, &DataModelError{Reason: "unexpected field in column entry: " + k}
			}
		}

		var name string
		if err := json.Unmarshal(c["name"], &name); err != nil || len(name) < 1 {
			return nil, &DataModelError{Reason: "column name must be a non-empty string"}
		}
		if _, exists := seenColNames[name]; exists {
			return nil, &DataModelError{Reason: fmt.Sprintf("duplicate column name %q", name)}
		}
		seenColNames[name] = struct{}{}

		var colType string
		if err := json.Unmarshal(c["type"], &colType); err != nil || len(colType) < 1 {
			return nil, &DataModelError{Reason: "column type must be a non-empty string"}
		}

		var nullable bool
		if err := json.Unmarshal(c["nullable"], &nullable); err != nil {
			return nil, &DataModelError{Reason: "column nullable must be a boolean", Cause: err}
		}

		if pRaw, ok := c["precision"]; ok && string(pRaw) != "null" {
			var precision int
			if err := json.Unmarshal(pRaw, &precision); err != nil || precision < 0 {
				return nil, &DataModelError{Reason: "column precision must be an integer >= 0"}
			}
		}

		if sRaw, ok := c["scale"]; ok && string(sRaw) != "null" {
			var scale int
			if err := json.Unmarshal(sRaw, &scale); err != nil || scale < 0 {
				return nil, &DataModelError{Reason: "column scale must be an integer >= 0"}
			}
		}

		var valType string
		if err := json.Unmarshal(c["value_type"], &valType); err != nil {
			return nil, &DataModelError{Reason: "column value_type must be a string", Cause: err}
		}
		if _, ok := allowedColumnValueTypes[valType]; !ok {
			return nil, &DataModelError{Reason: fmt.Sprintf("invalid column value_type %q", valType)}
		}

		if valType == "UNKNOWN" {
			if string(c["type_mapping_ref"]) != "null" {
				return nil, &DataModelError{Reason: "type_mapping_ref must be null when value_type is UNKNOWN"}
			}
		} else {
			if string(c["type_mapping_ref"]) == "null" || !validProfileRef(c["type_mapping_ref"], false) {
				return nil, &DataModelError{Reason: "type_mapping_ref must be sha256 reference when value_type is not UNKNOWN"}
			}
		}
	}

	return seenColNames, nil
}

func validateDataModelColumnsV1(raw json.RawMessage) (map[string]struct{}, error) {
	var columns []json.RawMessage
	if err := json.Unmarshal(raw, &columns); err != nil {
		return nil, &DataModelError{Reason: "columns must be an array", Cause: err}
	}

	seenColNames := make(map[string]struct{}, len(columns))
	for _, rawCol := range columns {
		var c map[string]json.RawMessage
		if err := json.Unmarshal(rawCol, &c); err != nil || c == nil {
			return nil, &DataModelError{Reason: "column entry must be an object", Cause: err}
		}

		for _, req := range []string{"name", "type", "nullable"} {
			v, ok := c[req]
			if !ok || string(v) == "null" {
				return nil, &DataModelError{Reason: "missing required column field: " + req}
			}
		}

		for k := range c {
			switch k {
			case "name", "type", "nullable", "precision", "scale":
			default:
				return nil, &DataModelError{Reason: "unexpected field in v1 column entry: " + k}
			}
		}

		var name string
		if err := json.Unmarshal(c["name"], &name); err != nil || len(name) < 1 {
			return nil, &DataModelError{Reason: "column name must be a non-empty string"}
		}
		if _, exists := seenColNames[name]; exists {
			return nil, &DataModelError{Reason: fmt.Sprintf("duplicate column name %q", name)}
		}
		seenColNames[name] = struct{}{}

		var colType string
		if err := json.Unmarshal(c["type"], &colType); err != nil || len(colType) < 1 {
			return nil, &DataModelError{Reason: "column type must be a non-empty string"}
		}

		var nullable bool
		if err := json.Unmarshal(c["nullable"], &nullable); err != nil {
			return nil, &DataModelError{Reason: "column nullable must be a boolean", Cause: err}
		}

		if pRaw, ok := c["precision"]; ok && string(pRaw) != "null" {
			var precision int
			if err := json.Unmarshal(pRaw, &precision); err != nil || precision < 0 {
				return nil, &DataModelError{Reason: "column precision must be an integer >= 0"}
			}
		}

		if sRaw, ok := c["scale"]; ok && string(sRaw) != "null" {
			var scale int
			if err := json.Unmarshal(sRaw, &scale); err != nil || scale < 0 {
				return nil, &DataModelError{Reason: "column scale must be an integer >= 0"}
			}
		}
	}

	return seenColNames, nil
}

func validateDataModelKeys(pkRaw, ukRaw json.RawMessage, colNames map[string]struct{}) error {
	var pk []string
	if err := json.Unmarshal(pkRaw, &pk); err != nil {
		return &DataModelError{Reason: "primary_key must be an array of strings", Cause: err}
	}
	for _, col := range pk {
		if len(col) < 1 {
			return &DataModelError{Reason: "primary_key column name must be non-empty"}
		}
		if _, ok := colNames[col]; !ok {
			return &DataModelError{Reason: fmt.Sprintf("primary_key column %q not found in columns", col)}
		}
	}

	var uk [][]string
	if err := json.Unmarshal(ukRaw, &uk); err != nil {
		return &DataModelError{Reason: "unique_keys must be an array of arrays of strings", Cause: err}
	}
	for _, keyGroup := range uk {
		for _, col := range keyGroup {
			if len(col) < 1 {
				return &DataModelError{Reason: "unique_keys column name must be non-empty"}
			}
			if _, ok := colNames[col]; !ok {
				return &DataModelError{Reason: fmt.Sprintf("unique_keys column %q not found in columns", col)}
			}
		}
	}

	return nil
}

func validateDataModelRelationshipsV2(raw json.RawMessage, colNames map[string]struct{}) error {
	var rels []json.RawMessage
	if err := json.Unmarshal(raw, &rels); err != nil {
		return &DataModelError{Reason: "relationships must be an array", Cause: err}
	}

	for _, rawRel := range rels {
		var r map[string]json.RawMessage
		if err := json.Unmarshal(rawRel, &r); err != nil || r == nil {
			return &DataModelError{Reason: "relationship entry must be an object", Cause: err}
		}

		for _, req := range []string{
			"kind", "from_columns", "to_connection_id", "to_catalog",
			"to_schema", "to_table", "to_columns", "evidence_refs",
		} {
			v, ok := r[req]
			if !ok {
				return &DataModelError{Reason: "missing required relationship field: " + req}
			}
			if req != "to_schema" && string(v) == "null" {
				return &DataModelError{Reason: "relationship field cannot be null: " + req}
			}
		}

		for k := range r {
			switch k {
			case "kind", "from_columns", "to_connection_id", "to_catalog",
				"to_schema", "to_table", "to_columns", "evidence_refs":
			default:
				return &DataModelError{Reason: "unexpected field in relationship entry: " + k}
			}
		}

		var kind string
		if err := json.Unmarshal(r["kind"], &kind); err != nil {
			return &DataModelError{Reason: "relationship kind must be a string", Cause: err}
		}
		if _, ok := allowedRelationshipKinds[kind]; !ok {
			return &DataModelError{Reason: fmt.Sprintf("invalid relationship kind %q", kind)}
		}

		var fromCols []string
		if err := json.Unmarshal(r["from_columns"], &fromCols); err != nil || len(fromCols) < 1 {
			return &DataModelError{Reason: "from_columns must be a non-empty array of strings"}
		}
		fromSet := make(map[string]struct{}, len(fromCols))
		for _, c := range fromCols {
			if len(c) < 1 {
				return &DataModelError{Reason: "from_columns item must be non-empty"}
			}
			if _, dup := fromSet[c]; dup {
				return &DataModelError{Reason: fmt.Sprintf("duplicate column %q in from_columns", c)}
			}
			fromSet[c] = struct{}{}
			if _, ok := colNames[c]; !ok {
				return &DataModelError{Reason: fmt.Sprintf("relationship from_columns %q not found in columns", c)}
			}
		}

		var toConnID string
		if err := json.Unmarshal(r["to_connection_id"], &toConnID); err != nil || !validAssessmentReportID(toConnID) {
			return &DataModelError{Reason: "to_connection_id must match pattern ^[a-z0-9_-]+$"}
		}

		var toCatalog string
		if err := json.Unmarshal(r["to_catalog"], &toCatalog); err != nil || len(toCatalog) < 1 {
			return &DataModelError{Reason: "to_catalog must be a non-empty string"}
		}

		if string(r["to_schema"]) != "null" {
			var toSchema string
			if err := json.Unmarshal(r["to_schema"], &toSchema); err != nil || len(toSchema) < 1 {
				return &DataModelError{Reason: "to_schema must be null or a non-empty string"}
			}
		}

		var toTable string
		if err := json.Unmarshal(r["to_table"], &toTable); err != nil || len(toTable) < 1 {
			return &DataModelError{Reason: "to_table must be a non-empty string"}
		}

		var toCols []string
		if err := json.Unmarshal(r["to_columns"], &toCols); err != nil || len(toCols) < 1 {
			return &DataModelError{Reason: "to_columns must be a non-empty array of strings"}
		}
		toSet := make(map[string]struct{}, len(toCols))
		for _, c := range toCols {
			if len(c) < 1 {
				return &DataModelError{Reason: "to_columns item must be non-empty"}
			}
			if _, dup := toSet[c]; dup {
				return &DataModelError{Reason: fmt.Sprintf("duplicate column %q in to_columns", c)}
			}
			toSet[c] = struct{}{}
		}

		var evRefs []string
		if err := json.Unmarshal(r["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
			return &DataModelError{Reason: "evidence_refs must be a non-empty array of strings"}
		}
		evSet := make(map[string]struct{}, len(evRefs))
		for _, ref := range evRefs {
			if !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
				return &DataModelError{Reason: "evidence_refs item must match pattern ^sha256:[a-f0-9]{64}$"}
			}
			if _, dup := evSet[ref]; dup {
				return &DataModelError{Reason: fmt.Sprintf("duplicate evidence_ref %q", ref)}
			}
			evSet[ref] = struct{}{}
		}
	}

	return nil
}

func validateDataModelOptionalStrings(root map[string]json.RawMessage) error {
	if cRaw, ok := root["collation"]; ok && string(cRaw) != "null" {
		var collation string
		if err := json.Unmarshal(cRaw, &collation); err != nil || len(collation) < 1 {
			return &DataModelError{Reason: "collation must be a non-empty string"}
		}
	}

	if tRaw, ok := root["timezone"]; ok && string(tRaw) != "null" {
		var tz string
		if err := json.Unmarshal(tRaw, &tz); err != nil || len(tz) < 1 {
			return &DataModelError{Reason: "timezone must be a non-empty string"}
		}
	}

	return nil
}

func validateDataModelEvidenceIDs(raw json.RawMessage) error {
	var evIDs []string
	if err := json.Unmarshal(raw, &evIDs); err != nil {
		return &DataModelError{Reason: "evidence_ids must be an array of strings", Cause: err}
	}
	for _, id := range evIDs {
		if !validAssessmentReportID(id) {
			return &DataModelError{Reason: "evidence_ids item must match pattern ^[a-z0-9_-]+$"}
		}
	}
	return nil
}

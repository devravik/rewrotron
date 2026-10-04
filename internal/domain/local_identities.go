package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

// Allowed identity kinds per specs/observation/local-identities.schema.json
const (
	IdentityKindEntity    = "ENTITY"
	IdentityKindPrincipal = "PRINCIPAL"
	IdentityKindSession   = "SESSION"
	IdentityKindToken     = "TOKEN"
	IdentityKindEffect    = "EFFECT"
)

var allowedIdentityKinds = map[string]struct{}{
	IdentityKindEntity:    {},
	IdentityKindPrincipal: {},
	IdentityKindSession:   {},
	IdentityKindToken:     {},
	IdentityKindEffect:    {},
}

var (
	localIDPattern       = regexp.MustCompile(`^[a-z0-9_-]+$`)
	resourceIDPattern    = regexp.MustCompile(`^[a-z0-9_-]+$`)
	workflowIDPattern    = regexp.MustCompile(`^[a-z0-9_-]+$`)
	observationIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

// LocalIdentitiesError describes an invalid local-identities ledger record.
type LocalIdentitiesError struct {
	Reason string
	Cause  error
}

func (e *LocalIdentitiesError) Error() string {
	if e.Cause != nil {
		return "invalid local identities ledger: " + e.Reason + ": " + e.Cause.Error()
	}
	return "invalid local identities ledger: " + e.Reason
}

func (e *LocalIdentitiesError) Unwrap() error { return e.Cause }

// LocalIdentitiesLedger represents a side-local identity ledger conforming to specs/observation/local-identities.schema.json.
// It records immutable observed identities for a single observation attempt and logical workflow occurrence.
// Source-only ledgers are published before any target or verification case exists (X62).
type LocalIdentitiesLedger struct {
	SchemaVersion        string               `json:"schema_version"`
	ProjectID            string               `json:"project_id"`
	Side                 string               `json:"side"`
	ObservationAttemptID string               `json:"observation_attempt_id"`
	LogicalWorkflowID    string               `json:"logical_workflow_id"`
	ExecutionWitnessRef  string               `json:"execution_witness_ref"`
	Origin               string               `json:"origin"`
	Entries              []LocalIdentityEntry `json:"entries"`
	CaptureEvidenceRefs  []string             `json:"capture_evidence_refs"`
}

// LocalIdentityEntry represents one observed local identity.
type LocalIdentityEntry struct {
	LocalID           string             `json:"local_id"`
	Kind              string             `json:"kind"`
	Domain            string             `json:"domain"`
	ActorID           *string            `json:"actor_id"`
	ResourceID        string             `json:"resource_id"`
	Value             LocalIdentityValue `json:"value"`
	RelationshipRefs  []string           `json:"relationship_refs"`
	CaptureChannelRef string             `json:"capture_channel_ref"`
	EvidenceRefs      []string           `json:"evidence_refs"`
}

// LocalIdentityValue represents the typed observed value (exact scalar string or canonical integer string).
// Invented missing/placeholder values are strictly rejected.
type LocalIdentityValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// MarshalJSON marshals LocalIdentityValue.
func (v LocalIdentityValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":  v.Type,
		"value": v.Value,
	})
}

// UnmarshalJSON unmarshals and validates LocalIdentityValue.
func (v *LocalIdentityValue) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return &LocalIdentitiesError{Reason: "entry value must be an object", Cause: err}
	}
	rawType, ok := raw["type"]
	if !ok {
		return &LocalIdentitiesError{Reason: "entry value missing required 'type'"}
	}
	var typeStr string
	if err := json.Unmarshal(rawType, &typeStr); err != nil {
		return &LocalIdentitiesError{Reason: "entry value 'type' must be string", Cause: err}
	}

	rawValue, ok := raw["value"]
	if !ok {
		return &LocalIdentitiesError{Reason: "entry value missing required 'value'"}
	}
	if len(raw) != 2 {
		return &LocalIdentitiesError{Reason: "entry value has unrecognized fields"}
	}

	switch typeStr {
	case ValueTypeString:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &LocalIdentitiesError{Reason: "string value must be a JSON string", Cause: err}
		}
		v.Type = ValueTypeString
		v.Value = s
	case ValueTypeInteger:
		var s string
		if err := json.Unmarshal(rawValue, &s); err != nil {
			return &LocalIdentitiesError{Reason: "integer value must be a string", Cause: err}
		}
		if !canonicalIntRegex.MatchString(s) {
			return &LocalIdentitiesError{Reason: "integer value does not match canonical integer pattern"}
		}
		v.Type = ValueTypeInteger
		v.Value = s
	default:
		return &LocalIdentitiesError{Reason: fmt.Sprintf("entry value type %q is not supported in local identities (only string and integer permitted)", typeStr)}
	}
	return nil
}

// ValidateLocalIdentitiesLedger validates raw JSON bytes against specs/observation/local-identities.schema.json.
func ValidateLocalIdentitiesLedger(raw []byte) error {
	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return &LocalIdentitiesError{Reason: "not a valid JSON object", Cause: err}
	}

	expectedFields := []string{
		"schema_version",
		"project_id",
		"side",
		"observation_attempt_id",
		"logical_workflow_id",
		"execution_witness_ref",
		"origin",
		"entries",
		"capture_evidence_refs",
	}

	for _, f := range expectedFields {
		if _, ok := record[f]; !ok {
			return &LocalIdentitiesError{Reason: "missing required field " + f}
		}
	}
	if len(record) != len(expectedFields) {
		return &LocalIdentitiesError{Reason: "ledger contains unrecognized fields (target placeholders forbidden)"}
	}

	// schema_version
	var schemaVer string
	if err := json.Unmarshal(record["schema_version"], &schemaVer); err != nil || schemaVer != "1.0" {
		return &LocalIdentitiesError{Reason: "schema_version must be '1.0'"}
	}

	// project_id, observation_attempt_id, logical_workflow_id
	var projectID, obsAttemptID, workflowID string
	if err := json.Unmarshal(record["project_id"], &projectID); err != nil || !projectIDPattern.MatchString(projectID) {
		return &LocalIdentitiesError{Reason: "invalid project_id pattern"}
	}
	if err := json.Unmarshal(record["observation_attempt_id"], &obsAttemptID); err != nil || !observationIDPattern.MatchString(obsAttemptID) {
		return &LocalIdentitiesError{Reason: "invalid observation_attempt_id pattern"}
	}
	if err := json.Unmarshal(record["logical_workflow_id"], &workflowID); err != nil || !workflowIDPattern.MatchString(workflowID) {
		return &LocalIdentitiesError{Reason: "invalid logical_workflow_id pattern"}
	}

	// side: SOURCE or TARGET
	var side string
	if err := json.Unmarshal(record["side"], &side); err != nil || (side != SideSource && side != SideTarget) {
		return &LocalIdentitiesError{Reason: "side must be SOURCE or TARGET"}
	}

	// execution_witness_ref
	var witnessRef string
	if err := json.Unmarshal(record["execution_witness_ref"], &witnessRef); err != nil || !digestPattern.MatchString(witnessRef) {
		return &LocalIdentitiesError{Reason: "execution_witness_ref must be sha256 digest"}
	}

	// origin: const ACTUAL_OBSERVED
	var origin string
	if err := json.Unmarshal(record["origin"], &origin); err != nil || origin != OriginActualObserved {
		return &LocalIdentitiesError{Reason: "origin must be ACTUAL_OBSERVED"}
	}

	// capture_evidence_refs: minItems 1
	var capEvRefs []string
	if err := json.Unmarshal(record["capture_evidence_refs"], &capEvRefs); err != nil || len(capEvRefs) < 1 {
		return &LocalIdentitiesError{Reason: "capture_evidence_refs must be a non-empty array of sha256 digests"}
	}
	for _, ref := range capEvRefs {
		if !digestPattern.MatchString(ref) {
			return &LocalIdentitiesError{Reason: "capture_evidence_refs item must match sha256 digest pattern"}
		}
	}

	// entries
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(record["entries"], &entries); err != nil {
		return &LocalIdentitiesError{Reason: "entries must be an array", Cause: err}
	}

	entryReq := []string{
		"local_id",
		"kind",
		"domain",
		"actor_id",
		"resource_id",
		"value",
		"relationship_refs",
		"capture_channel_ref",
		"evidence_refs",
	}

	for _, e := range entries {
		for _, f := range entryReq {
			if _, ok := e[f]; !ok {
				return &LocalIdentitiesError{Reason: "entry missing required field " + f}
			}
		}
		if len(e) != len(entryReq) {
			return &LocalIdentitiesError{Reason: "entry contains unrecognized fields"}
		}

		var localID, kind, domain, resID string
		if err := json.Unmarshal(e["local_id"], &localID); err != nil || !localIDPattern.MatchString(localID) {
			return &LocalIdentitiesError{Reason: "entry local_id must match identifier pattern"}
		}
		if err := json.Unmarshal(e["kind"], &kind); err != nil {
			return &LocalIdentitiesError{Reason: "entry kind must be string"}
		}
		if _, ok := allowedIdentityKinds[kind]; !ok {
			return &LocalIdentitiesError{Reason: fmt.Sprintf("unsupported entry kind %q", kind)}
		}
		if err := json.Unmarshal(e["domain"], &domain); err != nil || domain == "" {
			return &LocalIdentitiesError{Reason: "entry domain must be non-empty string"}
		}
		if err := json.Unmarshal(e["resource_id"], &resID); err != nil || !resourceIDPattern.MatchString(resID) {
			return &LocalIdentitiesError{Reason: "entry resource_id must match identifier pattern"}
		}

		if string(bytes.TrimSpace(e["actor_id"])) != "null" {
			var actorID string
			if err := json.Unmarshal(e["actor_id"], &actorID); err != nil || !localIDPattern.MatchString(actorID) {
				return &LocalIdentitiesError{Reason: "entry actor_id must be null or match identifier pattern"}
			}
		}

		var val LocalIdentityValue
		if err := json.Unmarshal(e["value"], &val); err != nil {
			return &LocalIdentitiesError{Reason: "invalid entry value", Cause: err}
		}

		var relRefs []string
		if err := json.Unmarshal(e["relationship_refs"], &relRefs); err != nil {
			return &LocalIdentitiesError{Reason: "entry relationship_refs must be an array of sha256 digests"}
		}
		for _, ref := range relRefs {
			if !digestPattern.MatchString(ref) {
				return &LocalIdentitiesError{Reason: "entry relationship_refs item must match sha256 digest pattern"}
			}
		}

		var chanRef string
		if err := json.Unmarshal(e["capture_channel_ref"], &chanRef); err != nil || !digestPattern.MatchString(chanRef) {
			return &LocalIdentitiesError{Reason: "entry capture_channel_ref must match sha256 digest pattern"}
		}

		var evRefs []string
		if err := json.Unmarshal(e["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
			return &LocalIdentitiesError{Reason: "entry evidence_refs must be non-empty array of sha256 digests"}
		}
		for _, ref := range evRefs {
			if !digestPattern.MatchString(ref) {
				return &LocalIdentitiesError{Reason: "entry evidence_refs item must match sha256 digest pattern"}
			}
		}
	}

	return nil
}

// ParseLocalIdentitiesLedger validates and parses a LocalIdentitiesLedger from raw JSON.
func ParseLocalIdentitiesLedger(raw []byte) (*LocalIdentitiesLedger, error) {
	if err := ValidateLocalIdentitiesLedger(raw); err != nil {
		return nil, err
	}
	var ledger LocalIdentitiesLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		return nil, &LocalIdentitiesError{Reason: "unmarshal failure", Cause: err}
	}
	return &ledger, nil
}

// CanonicalJSON serializes the LocalIdentitiesLedger to rwt-cjson-1 canonical JSON bytes.
func (l *LocalIdentitiesLedger) CanonicalJSON() ([]byte, error) {
	raw, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

// FindEntry returns an entry by its local_id.
func (l *LocalIdentitiesLedger) FindEntry(localID string) (*LocalIdentityEntry, bool) {
	for i := range l.Entries {
		if l.Entries[i].LocalID == localID {
			return &l.Entries[i], true
		}
	}
	return nil, false
}

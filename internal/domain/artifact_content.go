package domain

import "encoding/json"

// ArtifactContentError describes an invalid transport record without exposing content.
type ArtifactContentError struct {
	Reason string
	Cause  error
}

// Error identifies the failed record obligation.
func (e *ArtifactContentError) Error() string { return "invalid artifact content: " + e.Reason }

// Unwrap preserves JSON/canonicalization failures for structured inspection.
func (e *ArtifactContentError) Unwrap() error { return e.Cause }

// ValidateArtifactContent checks the artifact-content schema shape within the
// repository's canonical JSON integer/Unicode subset. It does not decode content,
// check its digest/size against bytes, resolve registration, establish privacy or
// authorize disclosure. Empty string fields are permitted by the schema. Callers
// must qualify raw-byte integrity and guarded read permissions separately.
func ValidateArtifactContent(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &ArtifactContentError{Reason: "canonical JSON", Cause: err}
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || len(record) != 5 {
		return &ArtifactContentError{Reason: "object fields", Cause: err}
	}
	for _, field := range []string{"artifact_ref", "media_type", "encoding", "content", "size_bytes"} {
		if _, present := record[field]; !present {
			return &ArtifactContentError{Reason: "missing required field"}
		}
	}
	if !validProfileRef(record["artifact_ref"], false) {
		return &ArtifactContentError{Reason: "artifact reference"}
	}
	for _, field := range []string{"media_type", "content"} {
		var value string
		if string(record[field]) == "null" || json.Unmarshal(record[field], &value) != nil {
			return &ArtifactContentError{Reason: "string field type"}
		}
	}
	switch profileString(record["encoding"]) {
	case "utf8", "base64":
	default:
		return &ArtifactContentError{Reason: "unsupported encoding"}
	}
	var size int64
	if string(record["size_bytes"]) == "null" || json.Unmarshal(record["size_bytes"], &size) != nil || size < 0 {
		return &ArtifactContentError{Reason: "byte size"}
	}
	return nil
}

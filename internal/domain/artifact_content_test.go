package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestArtifactContentRepositoryFixtures(t *testing.T) {
	count := 0
	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/operations/artifact-content.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			err := ValidateArtifactContent(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			reasons := map[string]string{
				"artifact-content-negative-size":    "byte size",
				"artifact-content-unknown-encoding": "unsupported encoding",
			}
			assertArtifactContentRejected(t, err, reasons[fixture.Name])
		})
	}
	if count != 4 {
		t.Fatal("artifact content fixture corpus changed; review coverage")
	}
}

func artifactContentControl() map[string]any {
	return map[string]any{
		"artifact_ref": "sha256:" + strings.Repeat("a", 64),
		"media_type":   "text/plain", "encoding": "utf8", "content": "hello\n", "size_bytes": 6,
	}
}

func validateArtifactContentControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateArtifactContent(raw)
}

func assertArtifactContentRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var invalid *ArtifactContentError
	if reason == "" || !errors.As(err, &invalid) || invalid.Reason != reason {
		t.Fatalf("wanted artifact content rejection %q, got %v", reason, err)
	}
	if err.Error() != "invalid artifact content: "+reason {
		t.Fatal("error disclosed record content")
	}
}

func TestArtifactContentRecordRejections(t *testing.T) {
	for _, tc := range []struct {
		field, reason string
		value         any
	}{
		{"artifact_ref", "artifact reference", nil},
		{"artifact_ref", "artifact reference", "sha256:" + strings.Repeat("A", 64)},
		{"artifact_ref", "artifact reference", "bad"},
		{"media_type", "string field type", nil},
		{"media_type", "string field type", 42},
		{"content", "string field type", nil},
		{"content", "string field type", []any{}},
		{"encoding", "unsupported encoding", nil},
		{"encoding", "unsupported encoding", "UTF8"},
		{"encoding", "unsupported encoding", true},
		{"size_bytes", "byte size", nil},
		{"size_bytes", "byte size", "6"},
		{"size_bytes", "byte size", -1},
		{"size_bytes", "canonical JSON", json.Number("0.5")},
		{"size_bytes", "canonical JSON", json.Number("9223372036854775808")},
		{"extra", "object fields", true},
	} {
		t.Run(tc.field+"/"+tc.reason, func(t *testing.T) {
			record := artifactContentControl()
			record[tc.field] = tc.value
			assertArtifactContentRejected(t, validateArtifactContentControl(t, record), tc.reason)
		})
	}
	for field := range artifactContentControl() {
		record := artifactContentControl()
		delete(record, field)
		assertArtifactContentRejected(t, validateArtifactContentControl(t, record), "object fields")
	}
	record := artifactContentControl()
	delete(record, "content")
	record["unexpected"] = "private"
	assertArtifactContentRejected(t, validateArtifactContentControl(t, record), "missing required field")
	for _, raw := range []string{`null`, `[]`, `"private"`} {
		assertArtifactContentRejected(t, ValidateArtifactContent([]byte(raw)), "object fields")
	}
	for _, raw := range []string{`{"content":"a","content":"b"}`, `{"content":"\ud800"}`, `{} {}`} {
		err := ValidateArtifactContent([]byte(raw))
		assertArtifactContentRejected(t, err, "canonical JSON")
		if errors.Unwrap(err) == nil {
			t.Fatal("canonicalization cause lost")
		}
	}
}

func TestArtifactContentShapeDoesNotEstablishIntegrity(t *testing.T) {
	for _, encoding := range []string{"utf8", "base64"} {
		record := artifactContentControl()
		record["encoding"] = encoding
		record["media_type"] = ""
		record["content"] = "" // Schema has no minimum length or encoding assertion.
		record["size_bytes"] = int64(9223372036854775807)
		if err := validateArtifactContentControl(t, record); err != nil {
			t.Fatal(err)
		}
	}
	record := artifactContentControl()
	record["encoding"] = "base64"
	record["content"] = "not base64!"
	if err := validateArtifactContentControl(t, record); err != nil {
		t.Fatal("shape validator invented a decoding policy", err)
	}
	// Structural acceptance cannot replace the separate byte integrity check.
	if VerifyArtifactBytes(record["artifact_ref"].(string), 6, []byte("hello\n")) == nil {
		t.Fatal("synthetic reference unexpectedly established integrity")
	}
}

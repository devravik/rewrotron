package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// SourceManifestError identifies invalid declared content without echoing paths.
type SourceManifestError struct {
	Reason string
	Cause  error
}

// Error describes the invalid manifest obligation.
func (e *SourceManifestError) Error() string { return "invalid source manifest: " + e.Reason }

// Unwrap preserves decoding/canonicalization failures for inspection.
func (e *SourceManifestError) Unwrap() error { return e.Cause }

// ValidateSourceManifest validates a source manifest against specs/project/source-manifest.schema.json
// and ANALYSIS-SPEC ordering and path invariants.
func ValidateSourceManifest(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &SourceManifestError{Reason: "canonical JSON", Cause: err}
	}
	manifest, err := manifestObject(canonical, "schema_version", "identity_algorithm", "entries", "exclusions", "completeness", "scanner_ref")
	if err != nil {
		return err
	}
	if manifestString(manifest["schema_version"]) != "1.0" ||
		manifestString(manifest["identity_algorithm"]) != "rwt-source-root-1" ||
		manifestString(manifest["completeness"]) != "COMPLETE" {
		return &SourceManifestError{Reason: "unsupported version, identity or completeness"}
	}
	scanner := manifestString(manifest["scanner_ref"])
	if !strings.HasPrefix(scanner, "sha256:") || !isManifestDigest(strings.TrimPrefix(scanner, "sha256:")) {
		return &SourceManifestError{Reason: "scanner reference"}
	}
	entries, err := manifestArray(manifest["entries"])
	if err != nil {
		return err
	}
	kinds := make(map[string]string, len(entries))
	previous := ""
	for _, rawEntry := range entries {
		path, kind, err := verifyManifestEntry(rawEntry)
		if err != nil {
			return err
		}
		if path <= previous {
			return &SourceManifestError{Reason: "entry paths must be strictly ordered"}
		}
		if slash := strings.LastIndexByte(path, '/'); slash >= 0 && kinds[path[:slash]] != "DIRECTORY" {
			return &SourceManifestError{Reason: "missing directory parent"}
		}
		kinds[path] = kind
		previous = path
	}
	return verifyManifestExclusions(manifest["exclusions"], kinds)
}

// VerifySourceManifest validates manifest 1.0 and supplied captured file bytes,
// returning distinct lowercase content-root and full-manifest SHA256 digests.
// It does not acquire files, follow symlinks, resolve scanner artifacts, prove
// coherent acquisition/containment or publish a sealed snapshot. Those require
// the source service's separate admission evidence. Callers must not mutate the
// supplied bytes during verification. Only declared FILE paths are read.
func VerifySourceManifest(raw []byte, capturedFiles map[string][]byte) (rootHash, manifestHash string, err error) {
	if err := ValidateSourceManifest(raw); err != nil {
		return "", "", err
	}
	canonical, _ := CanonicalJSON(raw)
	manifest, _ := manifestObject(canonical, "schema_version", "identity_algorithm", "entries", "exclusions", "completeness", "scanner_ref")
	entries, _ := manifestArray(manifest["entries"])
	if err := verifyCapturedFiles(entries, capturedFiles); err != nil {
		return "", "", err
	}
	preimage, err := json.Marshal(struct {
		Identity string          `json:"identity"`
		Entries  json.RawMessage `json:"entries"`
	}{"rwt-source-root-1", manifest["entries"]})
	if err != nil {
		return "", "", &SourceManifestError{Reason: "root preimage", Cause: err}
	}
	rootHash, err = CanonicalHash(preimage)
	if err != nil {
		return "", "", &SourceManifestError{Reason: "root hashing", Cause: err}
	}
	digest := sha256.Sum256(canonical)
	return rootHash, hex.EncodeToString(digest[:]), nil
}

func verifyManifestEntry(raw json.RawMessage) (string, string, error) {
	var entry map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entry); err != nil || entry == nil {
		return "", "", &SourceManifestError{Reason: "entry object", Cause: err}
	}
	kind := manifestString(entry["kind"])
	var fields []string
	switch kind {
	case "FILE":
		fields = []string{"path", "kind", "sha256", "size_bytes"}
	case "DIRECTORY":
		fields = []string{"path", "kind"}
	case "SYMLINK":
		fields = []string{"path", "kind", "target"}
	default:
		return "", "", &SourceManifestError{Reason: "unsupported entry kind"}
	}
	if _, err := manifestObject(raw, fields...); err != nil {
		return "", "", err
	}
	path := manifestString(entry["path"])
	if !validManifestPath(path) {
		return "", "", &SourceManifestError{Reason: "invalid entry path"}
	}
	if kind == "SYMLINK" {
		target := manifestString(entry["target"])
		if target == "" || strings.ContainsRune(target, 0) {
			return "", "", &SourceManifestError{Reason: "invalid literal symlink target"}
		}
	}
	if kind == "FILE" {
		digest := manifestString(entry["sha256"])
		var size int64
		if !isManifestDigest(digest) || json.Unmarshal(entry["size_bytes"], &size) != nil || size < 0 {
			return "", "", &SourceManifestError{Reason: "invalid file digest or size"}
		}
	}
	return path, kind, nil
}

func verifyManifestExclusions(raw json.RawMessage, kinds map[string]string) error {
	exclusions, err := manifestArray(raw)
	if err != nil {
		return err
	}
	previous := ""
	for _, rawExclusion := range exclusions {
		exclusion, err := manifestObject(rawExclusion, "path", "reason")
		if err != nil {
			return err
		}
		path := manifestString(exclusion["path"])
		if !validManifestPath(path) || path <= previous || manifestString(exclusion["reason"]) == "" {
			return &SourceManifestError{Reason: "invalid or unordered exclusion"}
		}
		// Every ancestor of an included path is already required above. An
		// included descendant therefore necessarily includes this path too.
		if _, included := kinds[path]; included {
			return &SourceManifestError{Reason: "exclusion overlaps included entry"}
		}
		previous = path
	}
	return nil
}

func verifyCapturedFiles(entries []json.RawMessage, files map[string][]byte) error {
	for _, raw := range entries {
		var entry struct {
			Path   string `json:"path"`
			Kind   string `json:"kind"`
			Digest string `json:"sha256"`
			Size   int64  `json:"size_bytes"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			return &SourceManifestError{Reason: "file projection", Cause: err}
		}
		if entry.Kind != "FILE" {
			continue
		}
		data, present := files[entry.Path]
		if !present || int64(len(data)) != entry.Size {
			return &SourceManifestError{Reason: "missing file bytes or size mismatch"}
		}
		actual := sha256.Sum256(data)
		if hex.EncodeToString(actual[:]) != entry.Digest {
			return &SourceManifestError{Reason: "file digest mismatch"}
		}
	}
	return nil
}

func manifestObject(raw []byte, fields ...string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || len(object) != len(fields) {
		return nil, &SourceManifestError{Reason: "object fields", Cause: err}
	}
	for _, field := range fields {
		value, exists := object[field]
		if !exists || string(value) == "null" {
			return nil, &SourceManifestError{Reason: "missing or null required field"}
		}
	}
	return object, nil
}

func manifestArray(raw json.RawMessage) ([]json.RawMessage, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return nil, &SourceManifestError{Reason: "array required", Cause: err}
	}
	return values, nil
}

func manifestString(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return value
}

func validManifestPath(path string) bool {
	if strings.ContainsAny(path, "\\\x00") {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func isManifestDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

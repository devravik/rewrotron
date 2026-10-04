package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ArtifactIntegrityError identifies an invalid byte identity without disclosing content.
type ArtifactIntegrityError struct {
	Reason string
}

// Error describes the failed integrity obligation.
func (e *ArtifactIntegrityError) Error() string { return "invalid artifact integrity: " + e.Reason }

// VerifyArtifactBytes checks a sha256 reference and declared nonnegative byte size
// against supplied raw bytes. It never canonicalizes, decodes or modifies content.
// Nil and empty buffers both represent zero bytes, not evidence of availability.
// Registration, ownership, privacy, filesystem containment and availability must
// be established separately by the artifact service. This check authorizes no read
// or execution. Callers must not mutate the buffer during verification.
func VerifyArtifactBytes(ref string, sizeBytes int64, content []byte) error {
	if !strings.HasPrefix(ref, "sha256:") || !isManifestDigest(strings.TrimPrefix(ref, "sha256:")) {
		return &ArtifactIntegrityError{Reason: "artifact reference"}
	}
	if sizeBytes < 0 {
		return &ArtifactIntegrityError{Reason: "negative byte size"}
	}
	if int64(len(content)) != sizeBytes {
		return &ArtifactIntegrityError{Reason: "byte size mismatch"}
	}
	digest := sha256.Sum256(content)
	if hex.EncodeToString(digest[:]) != strings.TrimPrefix(ref, "sha256:") {
		return &ArtifactIntegrityError{Reason: "byte digest mismatch"}
	}
	return nil
}

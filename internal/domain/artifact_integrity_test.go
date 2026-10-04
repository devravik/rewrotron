package domain

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestArtifactBytesGolden(t *testing.T) {
	// Fixed raw-byte digests also used by source-manifest identity vectors;
	// expected values are independent of the implementation's hashing path.
	cases := []struct {
		name, digest string
		content      []byte
	}{
		{"empty", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", []byte{}},
		{"nil zero bytes", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil},
		{"text newline", "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03", []byte("hello\n")},
		{"binary", "e9489f37fb3051e9efa1dc916004d7274e7b63975e3209708947267f2393a9be", []byte{0, 255, 13, 10}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := bytes.Clone(tc.content)
			if err := VerifyArtifactBytes("sha256:"+tc.digest, int64(len(tc.content)), tc.content); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, tc.content) {
				t.Fatal("verification modified supplied bytes")
			}
		})
	}
}

func TestArtifactBytesRejectsCorruption(t *testing.T) {
	const digest = "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"
	const ref = "sha256:" + digest
	cases := []struct {
		name, ref, content, reason string
		size                       int64
	}{
		{"missing prefix", digest, "hello\n", "artifact reference", 6},
		{"wrong algorithm", "sha512:" + digest, "hello\n", "artifact reference", 6},
		{"uppercase digest", "sha256:" + strings.ToUpper(digest), "hello\n", "artifact reference", 6},
		{"short digest", ref[:len(ref)-1], "hello\n", "artifact reference", 6},
		{"invalid hex", "sha256:" + strings.Repeat("g", 64), "hello\n", "artifact reference", 6},
		{"trailing whitespace", ref + " ", "hello\n", "artifact reference", 6},
		{"negative size", ref, "hello\n", "negative byte size", -1},
		{"wrong declared size", ref, "hello\n", "byte size mismatch", 7},
		{"truncated content", ref, "hello", "byte size mismatch", 6},
		{"absent nonempty content", ref, "", "byte size mismatch", 6},
		{"same length mutation", ref, "HELLO\n", "byte digest mismatch", 6},
		{"line ending mutation", ref, "hello\r", "byte digest mismatch", 6},
		{"other valid digest", "sha256:" + strings.Repeat("0", 64), "hello\n", "byte digest mismatch", 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyArtifactBytes(tc.ref, tc.size, []byte(tc.content))
			var invalid *ArtifactIntegrityError
			if !errors.As(err, &invalid) || invalid.Reason != tc.reason {
				t.Fatalf("wanted %q, got %v", tc.reason, err)
			}
			if err.Error() != "invalid artifact integrity: "+tc.reason {
				t.Fatal("error disclosed supplied reference/content")
			}
		})
	}
}

func TestArtifactBytesDoesNotCanonicalizeJSON(t *testing.T) {
	canonical := []byte(`{"a":1,"b":2}`)
	digest, err := CanonicalHash(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyArtifactBytes("sha256:"+digest, int64(len(canonical)), canonical); err != nil {
		t.Fatal(err)
	}
	// Equivalent JSON of the same length must retain a different raw identity.
	reordered := []byte(`{"b":2,"a":1}`)
	err = VerifyArtifactBytes("sha256:"+digest, int64(len(reordered)), reordered)
	var invalid *ArtifactIntegrityError
	if !errors.As(err, &invalid) || invalid.Reason != "byte digest mismatch" {
		t.Fatalf("raw reordered JSON was not rejected: %v", err)
	}
}

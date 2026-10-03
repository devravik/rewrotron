package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestCanonicalApprovalVectors(t *testing.T) {
	t.Parallel()
	var vectors []struct {
		Name            string          `json:"name"`
		Subject         json.RawMessage `json:"subject"`
		Canonical       string          `json:"canonical_utf8"`
		Hash            string          `json:"sha256"`
		RejectedRawJSON string          `json:"rejected_raw_json"`
	}
	loadVectors(t, "../../specs/examples/approval-hash-vectors.json", &vectors)
	if len(vectors) == 0 {
		t.Fatal("missing approval vectors")
	}
	for _, vector := range vectors {
		t.Run(vector.Name, func(t *testing.T) {
			t.Parallel()
			if vector.RejectedRawJSON != "" {
				assertRejected(t, []byte(vector.RejectedRawJSON))
				return
			}
			assertGolden(t, vector.Subject, vector.Canonical, vector.Hash)
		})
	}
}

func TestCanonicalSourceRootVectors(t *testing.T) {
	t.Parallel()
	var document struct {
		Vectors []struct {
			Name      string          `json:"name"`
			Admission string          `json:"expected_admission"`
			Preimage  json.RawMessage `json:"preimage"`
			Canonical string          `json:"canonical_utf8"`
			Hash      string          `json:"sha256"`
		} `json:"source_root_hash"`
	}
	loadVectors(t, "../../specs/examples/contract-blocker-vectors.json", &document)
	count := 0
	for _, vector := range document.Vectors {
		if vector.Admission != "ACCEPT" {
			continue // Snapshot admission is outside this serialization unit.
		}
		count++
		t.Run(vector.Name, func(t *testing.T) {
			t.Parallel()
			assertGolden(t, vector.Preimage, vector.Canonical, vector.Hash)
		})
	}
	if count == 0 {
		t.Fatal("missing root hash preimages")
	}
}

func TestCanonicalJSONSpelling(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, input, want string }{
		{"recursive-order", ` {"z":[2,1],"a":{"y":null,"x":true}} `, `{"a":{"x":true,"y":null},"z":[2,1]}`},
		{"no-html-or-separator-escaping", "\"<>&/\u2028\u2029\"", "\"<>&/\u2028\u2029\""},
		{"short-and-hex-escapes", `"\u0000\u0008\u0009\u000a\u000c\u000d\u001f"`, `"\u0000\b\t\n\f\r\u001f"`},
		{"surrogate-pair", `"\ud83d\ude00"`, `"😀"`},
		{"literal-replacement-character", `"�"`, `"�"`},
		{"escaped-backslash", `"\\ud800"`, `"\\ud800"`},
		{"decoded-key-order", `{"\u00e9":1,"a":2,"😀":3}`, `{"a":2,"é":1,"😀":3}`},
		{"no-unicode-normalization", `{"é":1,"e\u0301":2}`, "{\"e\u0301\":2,\"é\":1}"},
		{"signed-min", `-9223372036854775808`, `-9223372036854775808`},
		{"signed-max", `9223372036854775807`, `9223372036854775807`},
		{"zero", `-0`, `0`},
		{"scalar-false", `false`, `false`},
		{"empty-object", `{}`, `{}`},
		{"empty-array", `[]`, `[]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := CanonicalJSON([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %q; want %q", got, tc.want)
			}
		})
	}
}

func TestCanonicalJSONRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	inputs := []string{
		``, `{"a":1,"a":2}`, `{"a":1,"\u0061":2}`, `{"x":{"a":1,"a":2}}`,
		`1.0`, `1e0`, `NaN`, `Infinity`, `9223372036854775808`, `-9223372036854775809`,
		`"\ud800"`, `"\udc00"`, `"\ud800\u0061"`, `"\ud800\\udc00"`,
		`"\ud800`, `"\u12"`, `"\uXXXX"`, `"\q"`, string([]byte{'"', 0xff, '"'}),
		"\xef\xbb\xbf{}", "\"a\nb\"", `01`, `+1`, `true false`, `{} []`, `null ]`,
		`[1,]`, `{"a":}`, `{"a":1`, `[1`, `"unterminated`, `"dangling\`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			assertRejected(t, []byte(input))
		})
	}
}

func FuzzCanonicalJSON(f *testing.F) {
	for _, seed := range []string{`{}`, `{"z":2,"a":1}`, `"\ud83d\ude00"`, `{"a":1,"a":2}`, `1e0`} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 65536 {
			t.Skip("bounded test input")
		}
		original := bytes.Clone(input)
		canonical, err := CanonicalJSON(input)
		if !bytes.Equal(input, original) {
			t.Fatal("mutated caller input")
		}
		if err != nil {
			if canonical != nil {
				t.Fatal("returned partial output on failure")
			}
			return
		}
		if !json.Valid(canonical) {
			t.Fatalf("invalid canonical JSON: %q", canonical)
		}
		repeated, err := CanonicalJSON(canonical)
		if err != nil || !bytes.Equal(canonical, repeated) {
			t.Fatalf("non-idempotent: %q, %v", repeated, err)
		}
	})
}

func loadVectors(t *testing.T, path string, target any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatal(err)
	}
}

func assertGolden(t *testing.T, input []byte, canonical, digest string) {
	t.Helper()
	original := bytes.Clone(input)
	got, err := CanonicalJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != canonical {
		t.Fatalf("canonical bytes differ: %q", got)
	}
	hash, err := CanonicalHash(input)
	if err != nil || hash != digest {
		t.Fatalf("digest got %s, %v; want %s", hash, err, digest)
	}
	if !bytes.Equal(input, original) {
		t.Fatal("mutated immutable fixture bytes")
	}
}

func assertRejected(t *testing.T, input []byte) {
	t.Helper()
	result, err := CanonicalJSON(input)
	var typed *CanonicalJSONError
	if !errors.As(err, &typed) || result != nil {
		t.Fatalf("got %q, %v; want typed rejection without output", result, err)
	}
	hash, err := CanonicalHash(input)
	if !errors.As(err, &typed) || hash != "" {
		t.Fatalf("got hash %q, %v; want typed rejection without digest", hash, err)
	}
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devravik/rewrotron/internal/domain"
)

func fixturePaths(t *testing.T) (string, string) {
	t.Helper()
	root, e := filepath.Abs("../../../fixtures/ci3-reference")
	if e != nil {
		t.Fatal(e)
	}
	return root, filepath.Join(root, "evidence")
}

func TestReferenceEvidence_CurrentBoundary(t *testing.T) {
	root, out := fixturePaths(t)
	b, err := derive(root, out)
	if err != nil {
		t.Fatal(err)
	}
	if err = domain.ValidateObservationAttestation(b); err != nil {
		t.Fatal(err)
	}
	var att domain.ObservationAttestationDocument
	if err = json.Unmarshal(b, &att); err != nil {
		t.Fatal(err)
	}
	if att.DerivedStatus != domain.AttestationStatusPartial {
		t.Fatal("unqualified dimensions cannot become COMPLETE")
	}
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(from, path)
		if e != nil {
			return e
		}
		dest := filepath.Join(to, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0700)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(dest, b, 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func rewrite(t *testing.T, path string, value any) {
	t.Helper()
	b, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
}

func TestReferenceEvidence_Falsification(t *testing.T) {
	for _, name := range []string{"missing-observation", "target-side", "changed-runtime", "changed-source", "changed-lock", "wrong-principal", "cookie-security", "missing-session-capture", "missing-launch", "false-control", "missing-build"} {
		t.Run(name, func(t *testing.T) {
			original, _ := fixturePaths(t)
			temp := t.TempDir()
			root := filepath.Join(temp, "fixtures/ci3-reference")
			out := filepath.Join(root, "evidence")
			copyTree(t, original, root)
			impl, e := os.ReadFile("../ci3_reference.py")
			if e != nil {
				t.Fatal(e)
			}
			if e = os.MkdirAll(filepath.Join(temp, "tools/fixtures"), 0700); e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(temp, "tools/fixtures/ci3_reference.py"), impl, 0600); e != nil {
				t.Fatal(e)
			}
			var c map[string]any
			if _, e = load(filepath.Join(out, "corpus-0.json"), &c); e != nil {
				t.Fatal(e)
			}
			switch name {
			case "changed-runtime", "changed-lock", "changed-source":
				f := map[string]string{"changed-runtime": "runtime.lock.json", "changed-lock": "composer.lock", "changed-source": "application/controllers/Reference.php"}[name]
				path := filepath.Join(root, f)
				b, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(path, append(b, ' '), 0600); e != nil {
					t.Fatal(e)
				}
			case "missing-launch", "missing-build":
				f := "launch-0.json"
				if name == "missing-build" {
					f = "qualified-0-artifact.json"
				}
				if e = os.Remove(filepath.Join(out, f)); e != nil {
					t.Fatal(e)
				}
			case "missing-observation":
				c["observations"] = []any{}
			case "target-side":
				c["side"] = "TARGET"
			case "false-control":
				c["controls"].([]any)[0].(map[string]any)["ok"] = false
			default:
				for _, item := range c["observations"].([]any) {
					o := item.(map[string]any)
					if o["step_id"] != "alice-login" {
						continue
					}
					if name == "cookie-security" {
						r := o["response"].(map[string]any)
						for _, h := range r["headers"].([]any) {
							pair := h.([]any)
							if pair[0] == "Set-Cookie" {
								pair[1] = strings.ReplaceAll(pair[1].(string), "HttpOnly", "unsafe")
							}
						}
					} else {
						after := o["after"].(map[string]any)
						if name == "missing-session-capture" {
							after["source_files"] = map[string]any{}
						} else {
							for _, value := range after["source_files"].(map[string]any) {
								record := value.(map[string]any)
								record["content"] = strings.ReplaceAll(record["content"].(string), "\"principal\":1", "\"principal\":2")
							}
						}
					}
				}
			}
			rewrite(t, filepath.Join(out, "corpus-0.json"), c)
			// Supply a coherent changed artifact address: refusal must not depend only
			// on a corrupted hash or producer-declared control failure.
			var r map[string]any
			if _, e = load(filepath.Join(out, "qualification-report.json"), &r); e != nil {
				t.Fatal(e)
			}
			b, e := json.Marshal(c)
			if e != nil {
				t.Fatal(e)
			}
			r["runs"].([]any)[0].(map[string]any)["corpus_ref"] = digest(b)
			rewrite(t, filepath.Join(out, "qualification-report.json"), r)
			if _, e = derive(root, out); e == nil {
				t.Fatalf("accepted %s", name)
			}
		})
	}
}

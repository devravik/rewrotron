// Command attest verifies fixture evidence and derives the existing observation
// attestation. It is a development evidence reader, never customer admission.
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devravik/rewrotron/internal/domain"
)

func load(path string, value any) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	canonical, err := domain.CanonicalJSON(raw)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(canonical, value); err != nil {
		return nil, err
	}
	return canonical, nil
}
func rawHash(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }
func digest(raw []byte) string {
	h, err := domain.CanonicalHash(raw)
	if err != nil {
		panic(err)
	}
	return "sha256:" + h
}

// report is an output of actual fixture execution, not a compatibility result.
type report struct {
	ProfileRef        string `json:"profile_ref"`
	ImplementationRef string `json:"implementation_ref"`
	SourceManifestRef string `json:"source_manifest_ref"`
	RuntimeRef        string `json:"runtime_ref"`
	DependencyLockRef string `json:"dependency_lock_ref"`
	Side              string `json:"side"`
	TargetInvolved    bool   `json:"target_involved"`
	Runs              []struct {
		CorpusRef string `json:"corpus_ref"`
		LaunchRef string `json:"launch_ref"`
		BuildRef  string `json:"build_ref"`
	} `json:"runs"`
}
type observation struct {
	Side   string `json:"side"`
	StepID string `json:"step_id"`
	After  struct {
		Database string          `json:"database"`
		Files    json.RawMessage `json:"source_files"`
	} `json:"after"`
	Response *struct {
		Status  int        `json:"status"`
		Body    string     `json:"body_base64"`
		Headers [][]string `json:"headers"`
	} `json:"response"`
}
type corpus struct {
	Side         string        `json:"side"`
	CaseID       *string       `json:"case_id"`
	Observations []observation `json:"observations"`
	Controls     []struct {
		ID string `json:"id"`
		OK bool   `json:"ok"`
	} `json:"controls"`
}

func derive(root, out string) ([]byte, error) {
	var r report
	_, err := load(filepath.Join(out, "qualification-report.json"), &r)
	if err != nil {
		return nil, err
	}
	if r.Side != "SOURCE" || r.TargetInvolved || len(r.Runs) != 3 {
		return nil, fmt.Errorf("UNKNOWN: source-only repetitions missing")
	}
	var manifest struct {
		Entries []struct {
			Path string `json:"path"`
			Kind string `json:"kind"`
		} `json:"entries"`
	}
	raw, err := load(filepath.Join(out, "source-manifest.json"), &manifest)
	if err != nil {
		return nil, err
	}
	if err := domain.ValidateSourceManifest(raw); err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	for _, entry := range manifest.Entries {
		if entry.Kind == "FILE" {
			location := filepath.Join(root, filepath.FromSlash(entry.Path))
			info, e := os.Lstat(location)
			if e != nil || !info.Mode().IsRegular() {
				return nil, fmt.Errorf("UNKNOWN: missing/nonregular source file")
			}
			b, e := os.ReadFile(location)
			if e != nil {
				return nil, e
			}
			files[entry.Path] = b
		}
	}
	sourceRoot, manifestHash, err := domain.VerifySourceManifest(raw, files)
	if err != nil {
		return nil, err
	}
	if r.SourceManifestRef != "sha256:"+manifestHash {
		return nil, fmt.Errorf("UNKNOWN: manifest identity changed")
	}
	for _, pair := range [][2]string{{"runtime.lock.json", r.RuntimeRef}, {"composer.lock", r.DependencyLockRef}} {
		b, e := os.ReadFile(filepath.Join(root, pair[0]))
		if e != nil {
			return nil, e
		}
		// These two identities intentionally hash raw lock bytes, not JSON canonical bytes.
		if "sha256:"+rawHash(b) != pair[1] {
			return nil, fmt.Errorf("UNKNOWN: runtime/dependency changed")
		}
	}
	var profile any
	profileRaw, err := load(filepath.Join(out, "input-profile.json"), &profile)
	if err != nil || digest(profileRaw) != r.ProfileRef {
		return nil, fmt.Errorf("UNKNOWN: profile missing/changed")
	}
	impl, err := os.ReadFile(filepath.Join(root, "../../tools/fixtures/ci3_reference.py"))
	if err != nil || "sha256:"+rawHash(impl) != r.ImplementationRef {
		return nil, fmt.Errorf("UNKNOWN: capture implementation changed")
	}
	controls := []string{}
	required := []string{"health", "missing-session", "failed-login", "alice-login", "alice-continuity", "csrf-single-use", "cross-session-csrf", "expired-csrf", "old-session-invalidated", "logout-invalidates", "expired-session", "observer-write-denied", "read-only-root", "queued-sink-ack", "insert-and-trigger-audit", "file-content-and-mode"}
	for i, run := range r.Runs {
		if run.LaunchRef == "" || run.BuildRef == "" {
			return nil, fmt.Errorf("UNKNOWN: missing build/launch witness")
		}
		var launch map[string]json.RawMessage
		b, e := load(filepath.Join(out, fmt.Sprintf("launch-%d.json", i)), &launch)
		if e != nil || digest(b) != run.LaunchRef {
			return nil, fmt.Errorf("UNKNOWN: launch evidence missing/changed")
		}
		artifactIndex := i
		if i == 2 {
			artifactIndex = 1
		}
		var built map[string]string
		b, e = load(filepath.Join(out, fmt.Sprintf("qualified-%d-artifact.json", artifactIndex)), &built)
		if e != nil || digest(b) != run.BuildRef || built["/usr/local/bin/php"] == "" {
			return nil, fmt.Errorf("UNKNOWN: build evidence missing/changed")
		}

		var c corpus
		b, e = load(filepath.Join(out, fmt.Sprintf("corpus-%d.json", i)), &c)
		if e != nil {
			return nil, e
		}
		if digest(b) != run.CorpusRef || c.Side != "SOURCE" || c.CaseID != nil {
			return nil, fmt.Errorf("UNKNOWN: invalid source corpus closure")
		}
		seen := map[string]bool{}
		for _, control := range c.Controls {
			if !control.OK || seen[control.ID] {
				return nil, fmt.Errorf("UNKNOWN: failed/duplicated control")
			}
			seen[control.ID] = true
		}
		for _, id := range required {
			if !seen[id] {
				return nil, fmt.Errorf("UNKNOWN: missing control %s", id)
			}
		}
		if len(c.Observations) < 30 {
			return nil, fmt.Errorf("UNKNOWN: incomplete capture")
		}
		seenSteps := map[string]bool{}
		for _, o := range c.Observations {
			if seenSteps[o.StepID] {
				return nil, fmt.Errorf("UNKNOWN: duplicate observation")
			}
			seenSteps[o.StepID] = true
			if o.Side != "SOURCE" {
				return nil, fmt.Errorf("UNKNOWN: target contamination")
			}
			if o.StepID == "worker" {
				continue
			}
			if o.Response == nil {
				return nil, fmt.Errorf("UNKNOWN: missing HTTP observation")
			}
			expectedStatus := map[string]int{"missing-session": 401, "failed-login": 401, "alice-login": 303, "alice-continuity": 200, "wrong-csrf": 403, "cross-session-csrf": 403, "csrf-single-use": 403, "expired-csrf": 403, "tampered-session": 401, "old-session-invalidated": 401, "logout-invalidates": 401, "expired-session": 401, "alice-create": 202, "bob-denied-alice-item": 403}
			if expected, ok := expectedStatus[o.StepID]; ok && o.Response.Status != expected {
				return nil, fmt.Errorf("MISMATCH: %s", o.StepID)
			}
			if o.StepID == "alice-login" || o.StepID == "bob-login" {
				var sessionID string
				cookies := 0
				for _, h := range o.Response.Headers {
					if len(h) != 2 {
						return nil, fmt.Errorf("UNKNOWN: malformed header")
					}
					if strings.EqualFold(h[0], "Set-Cookie") {
						cookies++
						if strings.HasPrefix(h[1], "fixture_session=") {
							parts := strings.Split(h[1], "; ")
							if len(parts) != 5 || parts[1] != "Path=/" || parts[2] != "Max-Age=60" || parts[3] != "HttpOnly" || parts[4] != "SameSite=Lax" {
								return nil, fmt.Errorf("MISMATCH: cookie security/lifetime")
							}
							sessionID = strings.TrimPrefix(parts[0], "fixture_session=")
						}
					}
				}
				if cookies != 2 || sessionID == "" {
					return nil, fmt.Errorf("UNKNOWN: cookie capture missing")
				}
				var files map[string]struct {
					Content string `json:"content"`
				}
				if e := json.Unmarshal(o.After.Files, &files); e != nil {
					return nil, fmt.Errorf("UNKNOWN: session store capture missing")
				}
				f, ok := files["/state/sessions/"+sessionID+".json"]
				if !ok {
					return nil, fmt.Errorf("UNKNOWN: actual session missing")
				}
				var session struct {
					Principal int `json:"principal"`
				}
				if e := json.Unmarshal([]byte(f.Content), &session); e != nil {
					return nil, fmt.Errorf("UNKNOWN: session capture unsupported")
				}
				principal, name := 1, "alice"
				if o.StepID == "bob-login" {
					principal, name = 2, "bob"
				}
				if session.Principal != principal {
					return nil, fmt.Errorf("MISMATCH: actual session principal")
				}
				if !strings.Contains(o.After.Database, fmt.Sprintf("\n%d\t%s\t", principal, name)) {
					return nil, fmt.Errorf("UNKNOWN: actual DB principal missing")
				}
			}
			body, e := base64.StdEncoding.DecodeString(o.Response.Body)
			if e != nil || !json.Valid(body) {
				return nil, fmt.Errorf("UNKNOWN: missing/invalid body")
			}
		}
		for _, id := range []string{"health", "alice-login", "alice-continuity", "bob-login", "wrong-csrf", "cross-session-csrf", "csrf-single-use", "expired-csrf", "logout", "logout-invalidates", "expired-session", "alice-create", "file-write", "worker", "direct-pdo-read", "bob-denied-alice-item"} {
			if !seenSteps[id] {
				return nil, fmt.Errorf("UNKNOWN: required actual observation missing")
			}
		}
		controls = append(controls, run.CorpusRef)
	}
	if r.Runs[0].BuildRef != r.Runs[1].BuildRef {
		return nil, fmt.Errorf("UNKNOWN: nonreproducible build")
	}
	// Final-state DB snapshots and diagnostics do not establish read/lock/intermediate
	// transaction visibility. Preserve that missing evidence in the attestation.
	mechanisms := []domain.ObservationMechanism{}
	for _, d := range []domain.ObservationDimension{domain.DimensionHTTP, domain.DimensionSession, domain.DimensionFiles, domain.DimensionDB, domain.DimensionExternal, domain.DimensionWorker} {
		coverage := domain.MechanismCoverageClosed
		limitations := []string{}
		if d == domain.DimensionWorker || d == domain.DimensionExternal {
			coverage = domain.MechanismCoveragePartial
			limitations = []string{"Disposable sink application/receipt and worker exit observed; current causal-span/delivery-occurrence/completion ABI is not yet qualified"}
		}
		if d == domain.DimensionDB {
			coverage = domain.MechanismCoveragePartial
			limitations = []string{"Committed fixture state and trigger effects observed; read sets, locks, intermediate transaction visibility and server outbound routines unqualified"}
		}
		mechanisms = append(mechanisms, domain.ObservationMechanism{MechanismID: "fixture_" + strings.ToLower(string(d)), Dimension: d, BoundaryRef: r.ProfileRef, InstrumentationRef: &r.ImplementationRef, SupportedPaths: []string{"ci3-reference-b1"}, ControlTestRefs: controls, ClosureTestRefs: controls, Coverage: coverage, ResidualLimitations: limitations})
	}
	att, err := domain.DeriveObservationAttestation(&domain.DeriveAttestationRequest{AttestationID: "ci3_reference_b1", ProjectID: "ci3_reference", ProfileRef: r.ProfileRef, Mechanisms: mechanisms})
	if err != nil {
		return nil, err
	}
	b, err := att.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "Verified rwt-source-root-1=%s; attestation=%s (no compatibility PASS)\n", sourceRoot, att.DerivedStatus)
	return b, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: attest FIXTURE_ROOT EVIDENCE_DIR")
		os.Exit(2)
	}
	b, err := derive(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var r report
	if _, err = load(filepath.Join(os.Args[2], "qualification-report.json"), &r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	schemaRefs := []string{}
	for _, name := range []string{"specs/project/source-manifest.schema.json", "specs/observation/attestation.schema.json"} {
		raw, e := os.ReadFile(name)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		schemaRefs = append(schemaRefs, "sha256:"+rawHash(raw))
	}
	executable, e := os.Executable()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	native, e := os.ReadFile(executable)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	controls := []string{}
	for _, run := range r.Runs {
		controls = append(controls, run.CorpusRef)
	}
	suite, e := os.ReadFile(filepath.Join(os.Args[1], "contracts.json"))
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	q := domain.QualificationRecord{SchemaVersion: "1.0", QualificationID: "ci3_reference_b1_fixture", ProjectID: "ci3_reference", SubjectKind: "ENGINE", ImplementationRef: "sha256:" + rawHash(native), SchemaRefs: schemaRefs, CapabilityIDs: []string{"ci3-reference-b1-source-fixture-evidence"}, InputProfileRef: r.ProfileRef, ControlSuiteRef: "sha256:" + rawHash(suite), ControlResultRefs: controls, Status: "SUPPORTED", ReasonCodes: []string{"EXECUTED_BOUNDED_SOURCE_FIXTURE"}, Limitations: []string{"Fixture evidence reader only; no installed adapter, target session codec, customer admission or compatibility verdict", "Database intermediate visibility/read sets/locks remain PARTIAL"}}
	qraw, e := json.Marshal(q)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if e = domain.ValidateQualificationRecord(qraw); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if e = os.WriteFile(filepath.Join(os.Args[2], "qualification-1.0.json"), append(qraw, '\n'), 0600); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if _, err = os.Stdout.Write(append(b, '\n')); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

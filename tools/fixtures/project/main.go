// Command project preserves qualified fixture capture fragments in the existing
// context/response ABI. It does not infer behavior or establish generation admission.
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/devravik/rewrotron/internal/domain"
)

func hash(b []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(b)) }
func canonical(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	return domain.CanonicalJSON(b)
}
func fields(raw json.RawMessage, allowed ...string) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if e := json.Unmarshal(raw, &m); e != nil || m == nil {
		return nil, fmt.Errorf("invalid capture object")
	}
	for k := range m {
		found := false
		for _, a := range allowed {
			found = found || k == a
		}
		if !found {
			return nil, fmt.Errorf("unsupported capture field %s", k)
		}
	}
	return m, nil
}

type fragment struct {
	Side     string          `json:"side"`
	Step     string          `json:"step_id"`
	Request  json.RawMessage `json:"request"`
	Before   json.RawMessage `json:"before"`
	After    json.RawMessage `json:"after"`
	Response json.RawMessage `json:"response"`
	Scenario string          `json:"scenario_id"`
	Case     json.RawMessage `json:"case_id"`
}
type capture struct {
	Side         string            `json:"side"`
	Run          string            `json:"run_id"`
	Scenario     string            `json:"scenario_id"`
	Case         json.RawMessage   `json:"case_id"`
	Observations []json.RawMessage `json:"observations"`
}
type artifact struct {
	Kind         string `json:"kind"`
	Ref          string `json:"ref"`
	Schema       string `json:"schema"`
	OccurrenceID string `json:"occurrence_id,omitempty"`
	Phase        string `json:"phase,omitempty"`
}
type bundle struct {
	Artifacts map[string][]byte
	Records   []artifact
	CorpusRef string
	Contexts  []*domain.EvaluationContext
}

func (b *bundle) put(kind, schema string, v any) (string, error) {
	raw, e := canonical(v)
	if e != nil {
		return "", e
	}
	r := hash(raw)
	b.Artifacts[r] = raw
	b.Records = append(b.Records, artifact{Kind: kind, Ref: r, Schema: schema})
	return r, nil
}
func (b *bundle) blob(raw []byte) string {
	r := hash(raw)
	b.Artifacts[r] = append([]byte(nil), raw...)
	return r
}
func unavailable(reason string) domain.TaggedValue { return domain.NewMissing(reason) }
func missing(reason string) *domain.MissingValue {
	return &domain.MissingValue{Type: domain.ValueTypeMissing, ReasonCode: reason}
}

func project(raw []byte, expectedRef, implementationRef string) (*bundle, error) {
	canon, e := domain.CanonicalJSON(raw)
	if e != nil {
		return nil, e
	}
	if hash(canon) != expectedRef {
		return nil, fmt.Errorf("capture digest changed")
	}
	if _, e = fields(canon, "side", "scenario_id", "case_id", "run_id", "observations", "controls", "server_profile", "server_diagnostics", "reproduction_projection"); e != nil {
		return nil, e
	}
	var c capture
	if e = json.Unmarshal(canon, &c); e != nil {
		return nil, e
	}
	if c.Side != "SOURCE" || c.Scenario != "ci3-reference" || c.Run == "" || string(c.Case) != "null" || len(c.Observations) == 0 {
		return nil, fmt.Errorf("source-only discovery capture required")
	}
	b := &bundle{Artifacts: map[string][]byte{expectedRef: canon}, CorpusRef: expectedRef}
	seen := map[string]bool{}
	for _, rawFragment := range c.Observations {
		if _, e = fields(rawFragment, "side", "scenario_id", "step_id", "case_id", "request", "before", "response", "after"); e != nil {
			return nil, e
		}
		var f fragment
		if e = json.Unmarshal(rawFragment, &f); e != nil {
			return nil, e
		}
		if f.Side != "SOURCE" || f.Step == "" || seen[f.Step] || strings.IndexFunc(f.Step, func(r rune) bool { return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' }) >= 0 {
			return nil, fmt.Errorf("missing/duplicate/wrong-side occurrence")
		}
		seen[f.Step] = true
		if len(f.Case) > 0 && string(f.Case) != "null" {
			return nil, fmt.Errorf("source capture cannot depend on a case")
		}
		if f.Scenario != "" && f.Scenario != c.Scenario {
			return nil, fmt.Errorf("wrong workflow provenance")
		}
		if len(f.Before) == 0 || len(f.After) == 0 {
			return nil, fmt.Errorf("missing phase capture")
		}
		evidence, e := b.put("capture-fragment", "", json.RawMessage(rawFragment))
		if e != nil {
			return nil, e
		}
		// Attempt and named step establish identity; array position and timestamps do not.
		occurrence := "src_" + hash([]byte(c.Run + "/" + f.Step))[7:]
		// The source corpus did not capture an intended actor for every request. Never
		// derive fixture intention from cookie ownership or observed principal.
		intended := unavailable("ERR_UNCAPTURED_ACTOR_SELECTION")
		if len(f.Request) > 0 {
			var req struct {
				Fields map[string]json.RawMessage `json:"fields"`
			}
			if err := json.Unmarshal(f.Request, &req); err != nil {
				return nil, err
			}
			if name, ok := req.Fields["name"]; ok && strings.Contains(f.Step, "login") {
				var subject string
				if err := json.Unmarshal(name, &subject); err != nil {
					return nil, err
				}
				intended = domain.NewString(subject)
			}
		}
		actor, e := b.put("principal-fixture", "specs/behavior/principal-fixture.schema.json", map[string]any{
			"schema_version": "1.0", "project_id": "ci3_reference", "actor_id": "captured_selector_" + hash([]byte(fmt.Sprint(intended.Value)))[7:23],
			"intended_subject": intended, "credential_ref": nil, "initial_session_ref": nil, "db_identity_ref": nil, "provenance_refs": []string{evidence}})
		if e != nil {
			return nil, e
		}
		for _, phase := range []string{domain.PhasePreOperation, domain.PhasePostOperation} {
			snapshot := f.Before
			if phase == domain.PhasePostOperation {
				snapshot = f.After
			}
			state, e := projectState(b, snapshot)
			if e != nil {
				return nil, e
			}
			ctx := &domain.EvaluationContext{SchemaVersion: "1.0", ProjectID: "ci3_reference", ContextID: occurrence + "_" + strings.ToLower(phase), Side: domain.SideSource, OccurrenceID: &occurrence, ProjectionRef: implementationRef, Phase: phase, Origin: domain.OriginActualObserved, EvidenceRefs: []string{expectedRef, evidence}, Namespaces: domain.EvaluationContextNamespaces{
				Actor: domain.EvaluationActor{FixtureActorRef: actor, ActualPrincipal: domain.ActualPrincipal{Missing: missing("ERR_UNCAPTURED_AUTHENTICATION_CHANNEL")}}, State: state, Binding: map[string]domain.TaggedValue{}, Clock: unavailable("ERR_UNCAPTURED_CONTROLLED_CLOCK")}}
			// Effective request headers/raw body were not captured. Keep the immutable
			// request fragment, but do not fill this closed namespace from driver defaults.
			if phase == domain.PhasePostOperation && len(f.Response) > 0 {
				response, e := projectResponse(b, f.Response)
				if e != nil {
					return nil, e
				}
				ctx.Namespaces.Result = response
				if response != nil {
					if _, e = b.put("response", "specs/behavior/response.schema.json", response); e != nil {
						return nil, e
					}
				}
			}
			ctxRaw, e := canonical(ctx)
			if e != nil {
				return nil, e
			}
			if e = domain.ValidateEvaluationContext(ctxRaw); e != nil {
				return nil, e
			}
			contextRef, e := b.put("context", "specs/behavior/evaluation-context.schema.json", ctx)
			if e != nil {
				return nil, e
			}
			b.Records[len(b.Records)-1].OccurrenceID = occurrence
			b.Records[len(b.Records)-1].Phase = phase
			_ = contextRef
			b.Contexts = append(b.Contexts, ctx)
		}
	}
	return b, nil
}

func projectState(b *bundle, raw json.RawMessage) (domain.EvaluationState, error) {
	state := domain.EvaluationState{DB: domain.DBTablesOrMissing{Missing: missing("ERR_UNQUALIFIED_DB_TYPE_AND_VISIBILITY")}, Sessions: unavailable("ERR_MISSING_SESSION_CAPTURE"), Cache: unavailable("ERR_UNCAPTURED_CACHE"), Files: unavailable("ERR_MISSING_FILE_CAPTURE"), Queues: unavailable("ERR_ASYNC_CAUSAL_COMPLETION_PARTIAL")}
	m, e := fields(raw, "side", "database", "source_files", "sink_files")
	if e != nil {
		return state, e
	}
	if string(m["side"]) != "\"SOURCE\"" {
		return state, fmt.Errorf("snapshot source provenance mismatch")
	}
	// Preserve DB TSV and sink snapshots in the capture fragment. Their native
	// type/read/transaction/delivery semantics cannot be reconstructed from TSV.
	source, ok := m["source_files"]
	if !ok || string(source) == "null" {
		return state, nil
	}
	files := map[string]json.RawMessage{}
	if string(source) != "[]" {
		if e = json.Unmarshal(source, &files); e != nil {
			return state, fmt.Errorf("unsupported file capture shape")
		}
	}
	sessions, regular := map[string]domain.TaggedValue{}, map[string]domain.TaggedValue{}
	for path, entry := range files {
		var f struct {
			Content string `json:"content"`
			Mode    string `json:"mode"`
		}
		object, e := fields(entry, "content", "mode")
		if e != nil {
			return state, e
		}
		if value, present := object["content"]; !present || string(value) == "null" {
			return state, fmt.Errorf("missing captured file content")
		}
		if value, present := object["mode"]; !present || string(value) == "null" {
			return state, fmt.Errorf("missing captured file mode")
		}
		if e = json.Unmarshal(entry, &f); e != nil {
			return state, e
		}
		if !strings.HasPrefix(path, "/state/") {
			return state, fmt.Errorf("file outside captured volume")
		}
		contentRef := b.blob([]byte(f.Content))
		bytes, e := domain.NewBytes(contentRef)
		if e != nil {
			return state, e
		}
		regular[path] = domain.NewObject(map[string]domain.TaggedValue{"content": bytes, "mode": domain.NewString(f.Mode)})
		if strings.HasPrefix(path, "/state/sessions/") {
			value, e := domain.ParseLosslessJSON([]byte(f.Content))
			if e != nil {
				value = unavailable("ERR_UNSUPPORTED_SESSION_JSON")
			}
			sessions[path] = value
		}
	}
	state.Files = domain.NewObject(regular)
	state.Sessions = domain.NewObject(sessions)
	return state, nil
}

func projectResponse(b *bundle, raw json.RawMessage) (*domain.EvaluationResponse, error) {
	m, e := fields(raw, "status", "headers", "body_base64")
	if e != nil {
		return nil, e
	}
	if len(m) != 3 || string(m["headers"]) == "null" || string(m["body_base64"]) == "null" || string(m["status"]) == "null" {
		return nil, nil
	} // incomplete capture, never an empty response
	var r struct {
		Status  int        `json:"status"`
		Headers [][]string `json:"headers"`
		Body    string     `json:"body_base64"`
	}
	if e = json.Unmarshal(raw, &r); e != nil {
		return nil, e
	}
	body, e := base64.StdEncoding.DecodeString(r.Body)
	if e != nil {
		return nil, fmt.Errorf("invalid captured body encoding")
	}
	if len(body) > 65536 {
		return nil, fmt.Errorf("capture bound exceeded")
	}
	rawBodyRef := b.blob(body)
	tag, e := domain.NewBytes(rawBodyRef)
	if e != nil {
		return nil, e
	}
	out := &domain.EvaluationResponse{SchemaVersion: "1.0", Status: r.Status, Headers: []domain.EvaluationHeader{}, Cookies: []domain.EvaluationCookie{}, Body: tag, Session: unavailable("ERR_UNCAPTURED_RESPONSE_SESSION"), Principal: unavailable("ERR_UNCAPTURED_AUTHENTICATION_CHANNEL"), EffectRefs: []string{}, ApplicationError: unavailable("ERR_UNCAPTURED_APPLICATION_ERROR_CHANNEL")}
	for _, h := range r.Headers {
		if len(h) != 2 || !utf8.ValidString(h[0]) || !utf8.ValidString(h[1]) || strings.ContainsAny(h[1], "\r\n") {
			return nil, fmt.Errorf("invalid captured header")
		}
		if strings.EqualFold(h[0], "Set-Cookie") {
			cookie, e := projectCookie(h[1])
			if e != nil {
				return nil, nil
			}
			out.Cookies = append(out.Cookies, cookie)
		} else {
			out.Headers = append(out.Headers, domain.EvaluationHeader{Name: h[0], Value: h[1]})
		}
		if utf8.Valid(body) && strings.EqualFold(h[0], "Content-Type") && strings.EqualFold(strings.TrimSpace(strings.Split(h[1], ";")[0]), "application/json") {
			parsed, e := domain.ParseLosslessJSON(body)
			if e != nil {
				out.Body = unavailable("ERR_UNSUPPORTED_JSON_BODY")
			} else {
				out.Body = parsed
			}
		}
	}
	rawOut, e := canonical(out)
	if e != nil {
		return nil, e
	}
	if e = domain.ValidateResponse(rawOut); e != nil {
		return nil, e
	}
	return out, nil
}

func projectCookie(raw string) (domain.EvaluationCookie, error) {
	result := domain.EvaluationCookie{}
	parsed := (&http.Response{Header: http.Header{"Set-Cookie": []string{raw}}}).Cookies()
	if len(parsed) != 1 || len(parsed[0].Unparsed) > 0 {
		return result, fmt.Errorf("unsupported cookie")
	}
	result.Name = parsed[0].Name
	result.Value = domain.NewString(parsed[0].Value)
	seen := map[string]bool{}
	for _, part := range strings.Split(raw, ";")[1:] {
		pieces := strings.SplitN(strings.TrimSpace(part), "=", 2)
		key := strings.ToLower(pieces[0])
		if seen[key] {
			return result, fmt.Errorf("duplicate cookie attribute")
		}
		seen[key] = true
		value := ""
		if len(pieces) == 2 {
			value = pieces[1]
		}
		switch key {
		case "path":
			result.Path = &value
		case "domain":
			result.Domain = &value
		case "max-age":
			i, e := strconv.Atoi(value)
			if e != nil {
				return result, e
			}
			result.MaxAge = &i
		case "expires":
			date, e := http.ParseTime(value)
			if e != nil {
				return result, e
			}
			s := date.UTC().Format("2006-01-02T15:04:05Z")
			result.Expires = &s
		case "samesite":
			if value != "Strict" && value != "Lax" && value != "None" {
				return result, fmt.Errorf("unsupported SameSite")
			}
			result.SameSite = &value
		case "secure":
			if len(pieces) != 1 {
				return result, fmt.Errorf("invalid Secure flag")
			}
			result.Secure = true
		case "httponly":
			if len(pieces) != 1 {
				return result, fmt.Errorf("invalid HttpOnly flag")
			}
			result.HTTPOnly = true
		default:
			return result, fmt.Errorf("unsupported cookie attribute")
		}
	}
	return result, nil
}

func run(input, output string) error {
	raw, e := os.ReadFile(filepath.Join(input, "qualification-report.json"))
	if e != nil {
		return e
	}
	var report struct {
		Side              string `json:"side"`
		TargetInvolved    bool   `json:"target_involved"`
		SourceManifestRef string `json:"source_manifest_ref"`
		Runs              []struct {
			CorpusRef string `json:"corpus_ref"`
		} `json:"runs"`
	}
	if e = json.Unmarshal(raw, &report); e != nil {
		return e
	}
	if report.Side != "SOURCE" || report.TargetInvolved {
		return fmt.Errorf("wrong source report provenance")
	}
	if len(report.Runs) != 3 {
		return fmt.Errorf("missing source repetitions")
	}
	manifest, e := os.ReadFile(filepath.Join(input, "source-manifest.json"))
	if e != nil {
		return e
	}
	canonManifest, e := domain.CanonicalJSON(manifest)
	if e != nil {
		return e
	}
	if hash(canonManifest) != report.SourceManifestRef {
		return fmt.Errorf("source manifest identity changed")
	}
	if e = domain.ValidateSourceManifest(canonManifest); e != nil {
		return e
	}
	var source struct {
		Entries []struct {
			Path string `json:"path"`
			Kind string `json:"kind"`
		} `json:"entries"`
	}
	if e = json.Unmarshal(canonManifest, &source); e != nil {
		return e
	}
	captured := map[string][]byte{}
	for _, entry := range source.Entries {
		if entry.Kind != "FILE" {
			continue
		}
		path := filepath.Join(filepath.Dir(input), filepath.FromSlash(entry.Path))
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source must remain a regular file")
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		captured[entry.Path] = content
	}
	if _, _, e = domain.VerifySourceManifest(canonManifest, captured); e != nil {
		return e
	}
	native, e := os.Executable()
	if e != nil {
		return e
	}
	binary, e := os.ReadFile(native)
	if e != nil {
		return e
	}
	implementation := hash(binary)
	if e = os.Mkdir(output, 0700); e != nil {
		return e
	}
	for i, record := range report.Runs {
		raw, e = os.ReadFile(filepath.Join(input, fmt.Sprintf("corpus-%d.json", i)))
		if e != nil {
			return e
		}
		b, e := project(raw, record.CorpusRef, implementation)
		if e != nil {
			return e
		}
		b.Artifacts[report.SourceManifestRef] = canonManifest
		for _, c := range b.Contexts {
			c.EvidenceRefs = append(c.EvidenceRefs, report.SourceManifestRef)
			encoded, err := canonical(c)
			if err != nil {
				return err
			}
			for j := range b.Records {
				r := &b.Records[j]
				if r.Kind == "context" && r.OccurrenceID == *c.OccurrenceID && r.Phase == c.Phase {
					delete(b.Artifacts, r.Ref)
					r.Ref = hash(encoded)
					b.Artifacts[r.Ref] = encoded
				}
			}
		}
		refs := make([]string, 0, len(b.Artifacts))
		for r := range b.Artifacts {
			refs = append(refs, r)
		}
		sort.Strings(refs)
		for _, r := range refs {
			if e = os.WriteFile(filepath.Join(output, r[7:]), b.Artifacts[r], 0600); e != nil {
				return e
			}
		}
		index, e := canonical(map[string]any{"side": "SOURCE", "source_manifest_ref": report.SourceManifestRef, "source_corpus_ref": b.CorpusRef, "projector_ref": implementation, "records": b.Records, "admission": "UNKNOWN", "limitations": []string{"Original capture lacks current execution witness/role qualification and controlled clock", "Effective request bytes and independent authentication channel unavailable", "DB intermediate visibility/type mapping and async causal/completion ABI remain PARTIAL", "No typed effect/step/trace publication or generation admission inferred"}})
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(output, fmt.Sprintf("run-%d-index.json", i)), append(index, '\n'), 0600); e != nil {
			return e
		}
	}
	return nil
}
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: project EVIDENCE_DIR NEW_OUTPUT_DIR")
		os.Exit(2)
	}
	if e := run(os.Args[1], os.Args[2]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

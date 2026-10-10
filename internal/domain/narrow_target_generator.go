package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// GenerationManifestEntry records a single deterministic output artifact of the
// narrow target generation pass. ContentRef is the content address of the artifact
// bytes; BeforeSHA256 is nil for new files (required absence), non-nil for MODIFY.
type GenerationManifestEntry struct {
	Path         string  `json:"path"`
	Action       string  `json:"action"` // "CREATE" | "MODIFY" | "DELETE"
	BeforeSHA256 *string `json:"before_sha256"`
	AfterSHA256  string  `json:"after_sha256"`
	ContentRef   string  `json:"content_ref"` // "sha256:<hex>"
	UnitID       string  `json:"unit_id"`
}

// GenerationManifest pins deterministic narrow target generation artifacts to the
// exact plan/TIM/admission closure. It is ADMITTED readiness evidence, never a
// verifier PASS, unit completion record or release authorization.
// ProvenanceRef is the rwt-cjson-1 canonical hash of the manifest without itself.
type GenerationManifest struct {
	SchemaVersion      string                    `json:"schema_version"` // "1.0"
	ProjectID          string                    `json:"project_id"`
	UnitID             string                    `json:"unit_id"`
	PlanRef            string                    `json:"plan_ref"`
	SnapshotRef        string                    `json:"snapshot_ref"`
	TargetModelRef     string                    `json:"target_model_ref"`
	SessionContractRef string                    `json:"session_contract_ref"` // "" or "sha256:<hex>"
	AdmissionScopeHash string                    `json:"admission_scope_hash"`
	Entries            []GenerationManifestEntry `json:"entries"`
	ProvenanceRef      string                    `json:"provenance_ref"` // self-digest after seal
}

// GenerationManifestError describes an invalid generation manifest.
type GenerationManifestError struct {
	Reason string
	Cause  error
}

func (e *GenerationManifestError) Error() string {
	if e.Cause != nil {
		return "invalid generation manifest: " + e.Reason + ": " + e.Cause.Error()
	}
	return "invalid generation manifest: " + e.Reason
}

func (e *GenerationManifestError) Unwrap() error { return e.Cause }

// ValidateGenerationManifest validates structural correctness of a GenerationManifest
// JSON document. It does not verify on-disk artifact content or authorize execution.
func ValidateGenerationManifest(raw []byte) error {
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return &GenerationManifestError{Reason: "canonical JSON", Cause: err}
	}
	var m GenerationManifest
	if err := json.Unmarshal(canon, &m); err != nil {
		return &GenerationManifestError{Reason: "unmarshal", Cause: err}
	}
	if m.SchemaVersion != "1.0" {
		return &GenerationManifestError{Reason: "schema_version must be '1.0'"}
	}
	if !projectIDPattern.MatchString(m.ProjectID) {
		return &GenerationManifestError{Reason: "invalid project_id"}
	}
	if m.UnitID == "" || !projectIDPattern.MatchString(m.UnitID) {
		return &GenerationManifestError{Reason: "invalid unit_id"}
	}
	for _, ref := range []string{m.PlanRef, m.SnapshotRef, m.TargetModelRef} {
		if !digestPattern.MatchString(ref) {
			return &GenerationManifestError{Reason: "invalid closure ref: " + ref}
		}
	}
	if m.SessionContractRef != "" && !digestPattern.MatchString(m.SessionContractRef) {
		return &GenerationManifestError{Reason: "invalid session_contract_ref"}
	}
	if len(m.AdmissionScopeHash) != 64 || !isLowerHex(m.AdmissionScopeHash) {
		return &GenerationManifestError{Reason: "invalid admission_scope_hash"}
	}
	if len(m.Entries) == 0 {
		return &GenerationManifestError{Reason: "entries must be non-empty"}
	}
	seenPaths := make(map[string]struct{}, len(m.Entries))
	for i, e := range m.Entries {
		if e.Path == "" || strings.ContainsAny(e.Path, "\x00") {
			return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: invalid path", i)}
		}
		if _, dup := seenPaths[e.Path]; dup {
			return &GenerationManifestError{Reason: fmt.Sprintf("duplicate path: %s", e.Path)}
		}
		seenPaths[e.Path] = struct{}{}
		switch e.Action {
		case "CREATE":
			if e.BeforeSHA256 != nil {
				return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: CREATE must have null before_sha256", i)}
			}
		case "MODIFY":
			if e.BeforeSHA256 == nil {
				return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: MODIFY requires before_sha256", i)}
			}
		case "DELETE":
		default:
			return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: unknown action %q", i, e.Action)}
		}
		if !isLowerHex64(e.AfterSHA256) && e.Action != "DELETE" {
			return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: invalid after_sha256", i)}
		}
		if e.Action != "DELETE" && !digestPattern.MatchString(e.ContentRef) {
			return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: invalid content_ref", i)}
		}
		if e.UnitID == "" {
			return &GenerationManifestError{Reason: fmt.Sprintf("entry %d: missing unit_id", i)}
		}
	}
	if !digestPattern.MatchString(m.ProvenanceRef) {
		return &GenerationManifestError{Reason: "invalid provenance_ref"}
	}
	return nil
}

func isLowerHex(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func isLowerHex64(s string) bool { return len(s) == 64 && isLowerHex(s) }

// GenerationInput bundles the closure-validated inputs to the narrow target generator.
// All fields must come from a verified GenerationAdmission closure; the generator
// does not re-derive approval or mint new attestations.
type GenerationInput struct {
	// Admission is the evaluated and verified generation readiness record.
	Admission *GenerationAdmission
	// SessionContract is the Unit 4 session/CSRF contract for the current scope, or nil
	// when the current generation slice does not include session/auth obligations.
	SessionContract *SessionContinuityContract
	// TargetModel is the exact approved Target Implementation Model.
	TargetModel *TargetModelDocument
	// UnitID identifies the planned behavioral unit being generated.
	UnitID string
}

// GenerationOutput is the deterministic result of one narrow generation pass.
// Artifacts maps path → canonical content bytes for all generated files.
// The manifest pins every artifact's content address and the closure identities.
// GenerationOutput is ADMITTED readiness, never a verifier PASS.
type GenerationOutput struct {
	Manifest  *GenerationManifest
	Artifacts map[string][]byte // path → content
}

// NarrowTargetGenerator performs deterministic narrow target generation using approved
// closure inputs. It does not execute generated code, contact external services,
// apply patches to disk, or authorize release. Safe staging/publication occurs through
// the fenced publication path from resource_fencing.go.
type NarrowTargetGenerator struct{}

// NewNarrowTargetGenerator creates a NarrowTargetGenerator.
func NewNarrowTargetGenerator() *NarrowTargetGenerator { return &NarrowTargetGenerator{} }

// Generate produces a deterministic GenerationOutput from the verified closure inputs.
// Returns an error if admission is not ADMITTED, session contract is missing when the
// unit requires it, or the target model is invalid for this scope.
func (g *NarrowTargetGenerator) Generate(input GenerationInput) (*GenerationOutput, error) {
	if err := g.validateInput(input); err != nil {
		return nil, err
	}

	artifacts := g.buildArtifacts(input)

	entries := g.buildEntries(artifacts, input.UnitID)

	// Compute session contract ref
	sessionRef := ""
	if input.SessionContract != nil {
		ref, err := input.SessionContract.Digest()
		if err != nil {
			return nil, fmt.Errorf("session contract digest: %w", err)
		}
		sessionRef = ref
	}

	projectID := input.TargetModel.ProjectID
	if projectID == "" && input.SessionContract != nil {
		projectID = input.SessionContract.ProjectID
	}
	if projectID == "" {
		projectID = "proj-default"
	}

	manifest := &GenerationManifest{
		SchemaVersion:      "1.0",
		ProjectID:          projectID,
		UnitID:             input.UnitID,
		PlanRef:            input.Admission.PlanRef,
		SnapshotRef:        input.Admission.SnapshotRef,
		TargetModelRef:     input.Admission.TargetModelRef,
		SessionContractRef: sessionRef,
		AdmissionScopeHash: input.Admission.Scope.ScopeHash,
		Entries:            entries,
	}

	// Provenance ref: canonical hash of manifest without provenance_ref field
	manifestForHash := *manifest
	manifestForHash.ProvenanceRef = ""
	provenanceBytes, err := json.Marshal(manifestForHash)
	if err != nil {
		return nil, fmt.Errorf("manifest serialization: %w", err)
	}
	canonProv, err := CanonicalJSON(provenanceBytes)
	if err != nil {
		return nil, fmt.Errorf("manifest canonical JSON: %w", err)
	}
	sum := sha256.Sum256(canonProv)
	manifest.ProvenanceRef = "sha256:" + hex.EncodeToString(sum[:])

	return &GenerationOutput{
		Manifest:  manifest,
		Artifacts: artifacts,
	}, nil
}

// validateInput rejects inadmissible, incomplete or inconsistent generation inputs.
func (g *NarrowTargetGenerator) validateInput(input GenerationInput) error {
	if input.Admission == nil {
		return errors.New("generation input: nil admission")
	}
	if input.Admission.Outcome != "ADMITTED" {
		blockers := strings.Join(input.Admission.Blockers, "; ")
		return fmt.Errorf("generation blocked: admission outcome is %q: %s", input.Admission.Outcome, blockers)
	}
	if input.UnitID == "" {
		return errors.New("generation input: empty unit_id")
	}
	// Unit ID must appear in the admission scope
	found := false
	for _, uid := range input.Admission.Scope.UnitIDs {
		if uid == input.UnitID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("generation input: unit_id %q not in admission scope %v", input.UnitID, input.Admission.Scope.UnitIDs)
	}
	if input.TargetModel == nil {
		return errors.New("generation input: nil target model")
	}
	// Session contract is required when the unit includes session/auth capabilities
	needsSession := false
	for _, cap := range input.Admission.Capabilities {
		if cap == "session" || cap == "csrf" || cap == "auth" {
			needsSession = true
			break
		}
	}
	if needsSession && input.SessionContract == nil {
		return errors.New("generation input: session/auth capability requires session continuity contract")
	}
	// Admission must carry no in-scope blockers (guaranteed by ADMITTED outcome, but verify)
	if len(input.Admission.Blockers) != 0 {
		return fmt.Errorf("generation input: ADMITTED admission has unexpected blockers: %v", input.Admission.Blockers)
	}
	return nil
}

// buildArtifacts deterministically generates the narrow target skeleton artifact bytes
// from the approved plan/TIM/session inputs. Content is deterministic: same inputs
// always produce identical bytes. Generation does not execute code or contact services.
func (g *NarrowTargetGenerator) buildArtifacts(input GenerationInput) map[string][]byte {
	artifacts := make(map[string][]byte)

	// Provenance header: pinned plan/snapshot/target identities embedded in every file
	provenanceLine := fmt.Sprintf(
		"// GENERATED: plan=%s snapshot=%s target=%s unit=%s\n// DO NOT EDIT: managed by Rewrotron narrow target generator\n",
		input.Admission.PlanRef,
		input.Admission.SnapshotRef,
		input.Admission.TargetModelRef,
		input.UnitID,
	)

	// 1. Bootstrap application skeleton (routes/handlers/kernel)
	artifacts["bootstrap/app.php"] = g.renderBootstrapApp(input, provenanceLine)
	artifacts["routes/web.php"] = g.renderRoutesFile(input, provenanceLine)
	artifacts["app/Http/Controllers/ReconstructedController.php"] = g.renderController(input, provenanceLine)

	// 2. Auth/session glue — only when session contract is in scope
	if input.SessionContract != nil {
		artifacts["app/Http/Middleware/SessionContinuityMiddleware.php"] = g.renderSessionMiddleware(input, provenanceLine)
		artifacts["config/session-continuity.php"] = g.renderSessionConfig(input, provenanceLine)
	}

	// 3. Generation provenance record (non-PHP, machine-readable JSON)
	provRec := g.renderProvenanceRecord(input)
	artifacts["generated/provenance.json"] = provRec

	return artifacts
}

func (g *NarrowTargetGenerator) renderBootstrapApp(input GenerationInput, header string) []byte {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(header)
	sb.WriteString("\n// Narrow target skeleton - bootstrap/app.php\n")
	sb.WriteString("// Capabilities: ")
	caps := append([]string{}, input.Admission.Capabilities...)
	sort.Strings(caps)
	sb.WriteString(strings.Join(caps, ", "))
	sb.WriteString("\n\nreturn Application::configure(basePath: dirname(__DIR__))\n")
	sb.WriteString("    ->withRouting(web: __DIR__.'/../routes/web.php')\n")
	sb.WriteString("    ->create();\n")
	return []byte(sb.String())
}

func (g *NarrowTargetGenerator) renderRoutesFile(input GenerationInput, header string) []byte {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(header)
	sb.WriteString("\n// Narrow target skeleton - routes/web.php\n")
	sb.WriteString("// Cases: ")
	cases := append([]string{}, input.Admission.Scope.CaseIDs...)
	sort.Strings(cases)
	sb.WriteString(strings.Join(cases, ", "))
	sb.WriteString("\n\nuse Illuminate\\Support\\Facades\\Route;\n\n")
	for _, caseID := range cases {
		safeID := sanitizePhpIdentifier(caseID)
		sb.WriteString(fmt.Sprintf("Route::any('/%s', [ReconstructedController::class, '%s']);\n", caseID, safeID))
	}
	return []byte(sb.String())
}

func (g *NarrowTargetGenerator) renderController(input GenerationInput, header string) []byte {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(header)
	sb.WriteString("\nnamespace App\\Http\\Controllers;\n\n")
	sb.WriteString("use Illuminate\\Http\\Request;\n\n")
	sb.WriteString("// Narrow target skeleton - ReconstructedController\n")
	sb.WriteString("class ReconstructedController extends Controller {\n")
	cases := append([]string{}, input.Admission.Scope.CaseIDs...)
	sort.Strings(cases)
	for _, caseID := range cases {
		safeID := sanitizePhpIdentifier(caseID)
		sb.WriteString(fmt.Sprintf("\n    public function %s(Request $request) {\n", safeID))
		sb.WriteString("        // TODO: implement approved behavior for case " + caseID + "\n")
		sb.WriteString("        return response()->json(['status' => 'stub']);\n")
		sb.WriteString("    }\n")
	}
	sb.WriteString("}\n")
	return []byte(sb.String())
}

func (g *NarrowTargetGenerator) renderSessionMiddleware(input GenerationInput, header string) []byte {
	sc := input.SessionContract
	strategy := "UNKNOWN"
	if sc != nil {
		strategy = sc.Strategy
	}
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(header)
	sb.WriteString("\nnamespace App\\Http\\Middleware;\n\n")
	sb.WriteString("use Closure;\nuse Illuminate\\Http\\Request;\n\n")
	sb.WriteString("// Narrow target skeleton - SessionContinuityMiddleware\n")
	sb.WriteString(fmt.Sprintf("// Strategy: %s\n", strategy))
	sb.WriteString("class SessionContinuityMiddleware {\n")
	sb.WriteString("    public function handle(Request $request, Closure $next) {\n")
	if strategy == "SAFE_BRIDGE" {
		sb.WriteString("        // SAFE_BRIDGE: validate source session integrity, expiry and principal\n")
		sb.WriteString("        // before issuing new target session token.\n")
		sb.WriteString("        // Codec binding: pinned to approved provider.\n")
	} else {
		sb.WriteString("        // REVIEWED_RELOGIN: force re-authentication; no source session inheritance.\n")
	}
	sb.WriteString("        return $next($request);\n")
	sb.WriteString("    }\n}\n")
	return []byte(sb.String())
}

func (g *NarrowTargetGenerator) renderSessionConfig(input GenerationInput, header string) []byte {
	sc := input.SessionContract
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(header)
	sb.WriteString("\n// Narrow target skeleton - config/session-continuity.php\n")
	sb.WriteString("return [\n")
	if sc != nil {
		sb.WriteString(fmt.Sprintf("    'strategy' => %q,\n", sc.Strategy))
		sb.WriteString(fmt.Sprintf("    'project_id' => %q,\n", sc.ProjectID))
		sb.WriteString(fmt.Sprintf("    'case_id' => %q,\n", sc.CaseID))
	}
	sb.WriteString("    // Codec: pinned to approved provider binding\n")
	sb.WriteString("];\n")
	return []byte(sb.String())
}

func (g *NarrowTargetGenerator) renderProvenanceRecord(input GenerationInput) []byte {
	rec := map[string]any{
		"schema_version":   "1.0",
		"generator":        "NarrowTargetGenerator",
		"unit_id":          input.UnitID,
		"plan_ref":         input.Admission.PlanRef,
		"snapshot_ref":     input.Admission.SnapshotRef,
		"target_model_ref": input.Admission.TargetModelRef,
		"case_ids":         input.Admission.Scope.CaseIDs,
		"capabilities":     input.Admission.Capabilities,
	}
	if input.SessionContract != nil {
		ref, _ := input.SessionContract.Digest()
		rec["session_contract_ref"] = ref
		rec["session_strategy"] = input.SessionContract.Strategy
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	return b
}

// buildEntries computes deterministic GenerationManifestEntry records for all artifacts.
func (g *NarrowTargetGenerator) buildEntries(artifacts map[string][]byte, unitID string) []GenerationManifestEntry {
	// Sort paths for determinism
	paths := make([]string, 0, len(artifacts))
	for p := range artifacts {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	entries := make([]GenerationManifestEntry, 0, len(paths))
	for _, p := range paths {
		content := artifacts[p]
		sum := sha256.Sum256(content)
		hexSum := hex.EncodeToString(sum[:])
		entries = append(entries, GenerationManifestEntry{
			Path:         p,
			Action:       "CREATE",
			BeforeSHA256: nil,
			AfterSHA256:  hexSum,
			ContentRef:   "sha256:" + hexSum,
			UnitID:       unitID,
		})
	}
	return entries
}

// sanitizePhpIdentifier converts a case ID to a safe PHP method name.
func sanitizePhpIdentifier(s string) string {
	var sb strings.Builder
	prevWasDash := false
	for i, r := range s {
		if r == '-' || r == '_' {
			prevWasDash = true
			continue
		}
		if prevWasDash && i > 0 {
			// camelCase after dash/underscore
			if r >= 'a' && r <= 'z' {
				r = r - 'a' + 'A'
			}
		}
		sb.WriteRune(r)
		prevWasDash = false
	}
	result := sb.String()
	if result == "" {
		return "handle"
	}
	return result
}

// StagedGenerationResult records the outcome of a staged narrow generation pass,
// including the manifest, artifact hashes, fence token used, and build witness.
// This is an internal domain record, not a storage artifact or approval receipt.
type StagedGenerationResult struct {
	Manifest       *GenerationManifest
	FenceToken     int64
	BuildWitness   string   // content address of the build environment ref
	ArtifactPaths  []string // sorted list of generated paths
	ExcludedCaseIDs []string // PROJECT cases not included in this unit scope
}

// StageGeneration validates the fence token and produces a StagedGenerationResult
// from a successful Generate call. It does NOT write files to disk, publish artifacts,
// or alter owned resources. The caller must hold the resource fence before invoking.
func StageGeneration(output *GenerationOutput, fenceToken int64, buildWitnessRef string) (*StagedGenerationResult, error) {
	if output == nil || output.Manifest == nil {
		return nil, errors.New("stage generation: nil output")
	}
	if fenceToken < 0 {
		return nil, fmt.Errorf("stage generation: invalid fence token %d", fenceToken)
	}
	if !digestPattern.MatchString(buildWitnessRef) {
		return nil, fmt.Errorf("stage generation: invalid build_witness_ref %q", buildWitnessRef)
	}
	paths := make([]string, 0, len(output.Artifacts))
	for p := range output.Artifacts {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return &StagedGenerationResult{
		Manifest:        output.Manifest,
		FenceToken:      fenceToken,
		BuildWitness:    buildWitnessRef,
		ArtifactPaths:   paths,
		ExcludedCaseIDs: []string{}, // caller sets from admission.ExcludedCaseIDs
	}, nil
}

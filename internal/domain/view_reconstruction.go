package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// ViewKind identifies the category of view or file artifact.
type ViewKind string

const (
	ViewKindBladeTemplate ViewKind = "BLADE_TEMPLATE"
	ViewKindStaticAsset   ViewKind = "STATIC_ASSET"
	ViewKindLayout        ViewKind = "LAYOUT"
	ViewKindPartial       ViewKind = "PARTIAL"
)

// EscapingMode declares the output escaping policy for template interpolation.
type EscapingMode string

const (
	EscapingModeAuto EscapingMode = "AUTO_ESCAPE"  // {{ $var }} - escaped by default
	EscapingModeRaw  EscapingMode = "RAW_APPROVED" // {!! $var !!} - explicitly approved raw output
	EscapingModeNone EscapingMode = "NONE"         // static assets / binary files
)

// ViewVariable declares an interpolation variable within a view template.
type ViewVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	RawApproved bool   `json:"raw_approved"` // true allows unescaped interpolation {!! ... !!}
}

// ViewTemplateObligation defines an approved view template or static resource
// obligation within the M3 plan.
type ViewTemplateObligation struct {
	ViewID       string         `json:"view_id"`
	SourcePath   string         `json:"source_path"`
	TargetPath   string         `json:"target_path"`
	Kind         ViewKind       `json:"kind"`
	EscapingMode EscapingMode   `json:"escaping_mode"`
	RawTemplate  string         `json:"raw_template"` // template content
	Variables    []ViewVariable `json:"variables"`
}

// ViewReconstructionError describes a failure during view or file reconstruction.
type ViewReconstructionError struct {
	Code   string
	Reason string
	Cause  error
}

func (e *ViewReconstructionError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s", e.Code, e.Reason)
	}
	return e.Reason
}

func (e *ViewReconstructionError) Unwrap() error { return e.Cause }

// ViewManifestEntry records an individual reconstructed view/file artifact.
type ViewManifestEntry struct {
	ViewID        string `json:"view_id"`
	Path          string `json:"path"`
	Kind          string `json:"kind"`
	EscapingMode  string `json:"escaping_mode"`
	ContentSHA256 string `json:"content_sha256"`
	ContentRef    string `json:"content_ref"`
}

// ViewManifest records the complete set of reconstructed view/file artifacts
// pinned to the plan and admission closure.
type ViewManifest struct {
	SchemaVersion string              `json:"schema_version"` // "1.0"
	ProjectID     string              `json:"project_id"`
	UnitID        string              `json:"unit_id"`
	PlanRef       string              `json:"plan_ref"`
	SnapshotRef   string              `json:"snapshot_ref"`
	Entries       []ViewManifestEntry `json:"entries"`
	ProvenanceRef string              `json:"provenance_ref"`
}

var viewManifestIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// ValidateViewManifest validates the structural correctness of a ViewManifest JSON.
func ValidateViewManifest(raw []byte) error {
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "canonical JSON failure", Cause: err}
	}
	var m ViewManifest
	if err := json.Unmarshal(canon, &m); err != nil {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "unmarshal failure", Cause: err}
	}
	if m.SchemaVersion != "1.0" {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "schema_version must be '1.0'"}
	}
	if !viewManifestIDPattern.MatchString(m.ProjectID) {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid project_id format"}
	}
	if !viewManifestIDPattern.MatchString(m.UnitID) {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid unit_id format"}
	}
	for _, ref := range []string{m.PlanRef, m.SnapshotRef} {
		if !digestPattern.MatchString(ref) {
			return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid digest ref: " + ref}
		}
	}
	if len(m.Entries) == 0 {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "entries must be non-empty"}
	}
	seenPaths := make(map[string]struct{}, len(m.Entries))
	for i, e := range m.Entries {
		if e.Path == "" {
			return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("entry %d: empty path", i)}
		}
		if _, dup := seenPaths[e.Path]; dup {
			return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("duplicate entry path: %s", e.Path)}
		}
		seenPaths[e.Path] = struct{}{}
		if !isLowerHex64(e.ContentSHA256) {
			return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("entry %d: invalid content_sha256", i)}
		}
		if !digestPattern.MatchString(e.ContentRef) {
			return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("entry %d: invalid content_ref", i)}
		}
	}
	if !digestPattern.MatchString(m.ProvenanceRef) {
		return &ViewReconstructionError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid provenance_ref"}
	}
	return nil
}

// ViewReconstructionInput bundles the closure-validated inputs for Unit 6 view reconstruction.
type ViewReconstructionInput struct {
	Admission    *GenerationAdmission
	TargetModel  *TargetModelDocument
	UnitID       string
	Obligations  []ViewTemplateObligation
	FencingToken int64
}

// ViewReconstructionOutput carries the deterministic results of Unit 6 reconstruction.
type ViewReconstructionOutput struct {
	Manifest  *ViewManifest
	Patch     *PatchRecord
	Artifacts map[string][]byte
}

// ViewFileReconstructor implements deterministic, safe view and file reconstruction
// for M4 Unit 6.
type ViewFileReconstructor struct{}

// NewViewFileReconstructor creates a ViewFileReconstructor.
func NewViewFileReconstructor() *ViewFileReconstructor {
	return &ViewFileReconstructor{}
}

// Reconstruct validates the input obligations and produces deterministic rendered
// view artifacts, a safe patch manifest conforming to patch.schema.json, and a ViewManifest.
func (r *ViewFileReconstructor) Reconstruct(input ViewReconstructionInput) (*ViewReconstructionOutput, error) {
	if err := r.validateInput(input); err != nil {
		return nil, err
	}

	// Validate path boundaries against Target Implementation Model
	allowedPathSet := make(map[string]bool)
	if comp, ok := input.TargetModel.ComponentByOwner(input.UnitID); ok {
		for _, p := range comp.Paths {
			allowedPathSet[filepath.Clean(p)] = true
		}
	}

	artifacts := make(map[string][]byte)
	entries := make([]ViewManifestEntry, 0, len(input.Obligations))
	patchFiles := make([]PatchFileRecord, 0, len(input.Obligations))

	// Sort obligations by TargetPath for determinism
	sortedObs := append([]ViewTemplateObligation{}, input.Obligations...)
	sort.Slice(sortedObs, func(i, j int) bool {
		return sortedObs[i].TargetPath < sortedObs[j].TargetPath
	})

	for _, ob := range sortedObs {
		if err := r.validateObligation(ob, allowedPathSet); err != nil {
			return nil, err
		}

		rendered, err := r.renderTemplate(ob)
		if err != nil {
			return nil, err
		}

		cleanPath := filepath.Clean(ob.TargetPath)
		artifacts[cleanPath] = rendered

		sum := sha256.Sum256(rendered)
		hexSum := hex.EncodeToString(sum[:])
		contentRef := "sha256:" + hexSum

		entries = append(entries, ViewManifestEntry{
			ViewID:        ob.ViewID,
			Path:          cleanPath,
			Kind:          string(ob.Kind),
			EscapingMode:  string(ob.EscapingMode),
			ContentSHA256: hexSum,
			ContentRef:    contentRef,
		})

		patchFiles = append(patchFiles, PatchFileRecord{
			Path:         cleanPath,
			Action:       PatchActionCreate,
			BeforeSHA256: nil,
			AfterSHA256:  &hexSum,
			ContentRef:   &contentRef,
		})
	}

	// Create PatchRecord conforming to patch.schema.json
	projID := input.TargetModel.ProjectID
	if projID == "" {
		projID = "proj-default"
	}
	patch := &PatchRecord{
		SchemaVersion: "1.0",
		ProjectID:     projID,
		UnitID:        input.UnitID,
		FencingToken:  input.FencingToken,
		Files:         patchFiles,
	}

	// Validate patch with canonical validator
	patchRaw, err := patch.CanonicalJSON()
	if err != nil {
		return nil, fmt.Errorf("patch canonicalization: %w", err)
	}
	if err := ValidatePatch(patchRaw); err != nil {
		return nil, fmt.Errorf("patch validation: %w", err)
	}

	// Build ViewManifest
	manifest := &ViewManifest{
		SchemaVersion: "1.0",
		ProjectID:     projID,
		UnitID:        input.UnitID,
		PlanRef:       input.Admission.PlanRef,
		SnapshotRef:   input.Admission.SnapshotRef,
		Entries:       entries,
	}

	// Compute self-provenance hash
	manCopy := *manifest
	manCopy.ProvenanceRef = ""
	manRaw, err := json.Marshal(manCopy)
	if err != nil {
		return nil, fmt.Errorf("manifest marshal: %w", err)
	}
	canonMan, err := CanonicalJSON(manRaw)
	if err != nil {
		return nil, fmt.Errorf("manifest canonicalization: %w", err)
	}
	manSum := sha256.Sum256(canonMan)
	manifest.ProvenanceRef = "sha256:" + hex.EncodeToString(manSum[:])

	return &ViewReconstructionOutput{
		Manifest:  manifest,
		Patch:     patch,
		Artifacts: artifacts,
	}, nil
}

func (r *ViewFileReconstructor) validateInput(input ViewReconstructionInput) error {
	if input.Admission == nil {
		return &ViewReconstructionError{Code: "ERR_ADMISSION", Reason: "nil generation admission"}
	}
	if input.Admission.Outcome != "ADMITTED" {
		return &ViewReconstructionError{
			Code:   "ERR_ADMISSION",
			Reason: fmt.Sprintf("admission outcome is %q, must be ADMITTED", input.Admission.Outcome),
		}
	}
	if input.UnitID == "" {
		return &ViewReconstructionError{Code: "ERR_INVALID_UNIT", Reason: "empty unit_id"}
	}
	unitFound := false
	for _, u := range input.Admission.Scope.UnitIDs {
		if u == input.UnitID {
			unitFound = true
			break
		}
	}
	if !unitFound {
		return &ViewReconstructionError{
			Code:   "ERR_SCOPE_MISMATCH",
			Reason: fmt.Sprintf("unit_id %q not in admission scope %v", input.UnitID, input.Admission.Scope.UnitIDs),
		}
	}
	if input.TargetModel == nil {
		return &ViewReconstructionError{Code: "ERR_TARGET_MODEL", Reason: "nil target model"}
	}
	if input.FencingToken < 0 {
		return &ViewReconstructionError{Code: "ERR_FENCE", Reason: fmt.Sprintf("invalid fencing token %d", input.FencingToken)}
	}
	if len(input.Obligations) == 0 {
		return &ViewReconstructionError{Code: "ERR_NO_OBLIGATIONS", Reason: "obligations list must be non-empty"}
	}
	return nil
}

func (r *ViewFileReconstructor) validateObligation(ob ViewTemplateObligation, allowedPaths map[string]bool) error {
	if ob.ViewID == "" {
		return &ViewReconstructionError{Code: "ERR_INVALID_VIEW", Reason: "empty view_id"}
	}
	targetPath := ob.TargetPath
	if targetPath == "" {
		return &ViewReconstructionError{Code: "ERR_INVALID_PATH", Reason: "empty target_path"}
	}

	// Security: detect path traversal, null bytes, absolute paths
	if strings.Contains(targetPath, "\x00") {
		return &ViewReconstructionError{Code: "ERR_TARGET_CONFLICT", Reason: "path contains null byte"}
	}
	if filepath.IsAbs(targetPath) || strings.HasPrefix(targetPath, "/") || strings.HasPrefix(targetPath, "\\") {
		return &ViewReconstructionError{Code: "ERR_TARGET_CONFLICT", Reason: "absolute path rejected"}
	}
	cleaned := filepath.Clean(targetPath)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(targetPath, "../") || strings.Contains(targetPath, "..\\") {
		return &ViewReconstructionError{Code: "ERR_TARGET_CONFLICT", Reason: "path traversal (..) rejected"}
	}

	// Enforce target directory conventions: resources/views/, public/, or resources/assets/
	if !strings.HasPrefix(cleaned, "resources/views/") &&
		!strings.HasPrefix(cleaned, "public/") &&
		!strings.HasPrefix(cleaned, "resources/assets/") {
		return &ViewReconstructionError{
			Code:   "ERR_TARGET_CONFLICT",
			Reason: fmt.Sprintf("path %q outside authorized view/asset directory spaces", cleaned),
		}
	}

	// If the Target Implementation Model specifies component paths, verify inclusion
	if len(allowedPaths) > 0 && !allowedPaths[cleaned] {
		return &ViewReconstructionError{
			Code:   "ERR_UNAPPROVED_VIEW",
			Reason: fmt.Sprintf("path %q not in approved TIM component paths for unit", cleaned),
		}
	}

	// Ensure template content is valid UTF-8
	if !utf8.ValidString(ob.RawTemplate) {
		return &ViewReconstructionError{
			Code:   "ERR_INVALID_ENCODING",
			Reason: fmt.Sprintf("template %q contains invalid UTF-8 bytes", ob.ViewID),
		}
	}

	return nil
}

// renderTemplate deterministically transforms an approved view obligation into
// safe Blade template or static asset content, enforcing escaping rules.
func (r *ViewFileReconstructor) renderTemplate(ob ViewTemplateObligation) ([]byte, error) {
	if ob.Kind == ViewKindStaticAsset {
		return []byte(ob.RawTemplate), nil
	}

	var sb strings.Builder
	// Standard provenance header
	sb.WriteString(fmt.Sprintf("{{-- GENERATED: view_id=%s kind=%s --}}\n", ob.ViewID, ob.Kind))

	// Template interpolation transformation:
	// Variable substitution:
	// Auto-escape (default): {{ $var }}
	// Raw approved: {!! $var !!}
	content := ob.RawTemplate

	// If variables are specified with RawApproved=false, verify no unescaped {!! !!} interpolation
	// exists in the raw template unless explicitly authorized.
	rawVarSet := make(map[string]bool)
	for _, v := range ob.Variables {
		if v.RawApproved {
			rawVarSet[v.Name] = true
		}
	}

	// Scan for unescaped Blade tags {!! ... !!}
	rawBladePattern := regexp.MustCompile(`\{!!\s*\$([a-zA-Z0-9_]+)\s*!!\}`)
	matches := rawBladePattern.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			varName := m[1]
			if !rawVarSet[varName] && ob.EscapingMode != EscapingModeRaw {
				return nil, &ViewReconstructionError{
					Code:   "ERR_UNAPPROVED_ESCAPING",
					Reason: fmt.Sprintf("unapproved raw interpolation {!! $%s !!} in view %s; requires RAW_APPROVED", varName, ob.ViewID),
				}
			}
		}
	}

	sb.WriteString(content)
	return []byte(sb.String()), nil
}

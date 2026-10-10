package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

// EffectDimension identifies the typed operational domain of an effect.
type EffectDimension string

const (
	EffectDimensionDBRead      EffectDimension = "DB_READ"
	EffectDimensionDBWrite     EffectDimension = "DB_WRITE"
	EffectDimensionTransaction EffectDimension = "TRANSACTION"
	EffectDimensionSession     EffectDimension = "SESSION"
	EffectDimensionCache       EffectDimension = "CACHE"
	EffectDimensionFile        EffectDimension = "FILE"
	EffectDimensionQueue       EffectDimension = "QUEUE"
	EffectDimensionExternal    EffectDimension = "EXTERNAL"
	EffectDimensionError       EffectDimension = "ERROR"
)

// EffectClass categorizes the mutation and reversibility impact of an effect.
type EffectClass string

const (
	EffectClassPure          EffectClass = "PURE"
	EffectClassReadOnly      EffectClass = "READ_ONLY"
	EffectClassLocalWrite    EffectClass = "LOCAL_WRITE"
	EffectClassExternalWrite EffectClass = "EXTERNAL_WRITE"
	EffectClassIrreversible  EffectClass = "IRREVERSIBLE"
)

var allowedEffectDimensions = map[EffectDimension]struct{}{
	EffectDimensionDBRead:      {},
	EffectDimensionDBWrite:     {},
	EffectDimensionTransaction: {},
	EffectDimensionSession:     {},
	EffectDimensionCache:       {},
	EffectDimensionFile:        {},
	EffectDimensionQueue:       {},
	EffectDimensionExternal:    {},
	EffectDimensionError:       {},
}

var effectIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// BusinessEffectRecord models an observed or planned application effect
// preserving provenance, causality, destination, and cardinality.
type BusinessEffectRecord struct {
	EffectID              string          `json:"effect_id"`
	LogicalEffectID       string          `json:"logical_effect_id"`
	Dimension             EffectDimension `json:"dimension"`
	EffectClass           EffectClass     `json:"effect_class"`
	BusinessKey           string          `json:"business_key"`
	DestinationID         string          `json:"destination_id"`
	BusinessApplicationID string          `json:"business_application_id"`
	SequenceIndex         int64           `json:"sequence_index"`
	CausalParentID        *string         `json:"causal_parent_id,omitempty"`
	PayloadRef            string          `json:"payload_ref"`
	Applied               bool            `json:"applied"`
}

// EffectSafetyError describes a violation of effect safety or causal invariants.
type EffectSafetyError struct {
	Code   string
	Reason string
	Cause  error
}

func (e *EffectSafetyError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s", e.Code, e.Reason)
	}
	return e.Reason
}

func (e *EffectSafetyError) Unwrap() error { return e.Cause }

// EffectPublicationManifest pins reconstructed effect obligations to the M3 plan.
type EffectPublicationManifest struct {
	SchemaVersion string                 `json:"schema_version"` // "1.0"
	ProjectID     string                 `json:"project_id"`
	UnitID        string                 `json:"unit_id"`
	PlanRef       string                 `json:"plan_ref"`
	SnapshotRef   string                 `json:"snapshot_ref"`
	Effects       []BusinessEffectRecord `json:"effects"`
	ProvenanceRef string                 `json:"provenance_ref"`
}

// ValidateEffectPublicationManifest validates an effect publication manifest.
func ValidateEffectPublicationManifest(raw []byte) error {
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "canonical JSON failure", Cause: err}
	}
	var m EffectPublicationManifest
	if err := json.Unmarshal(canon, &m); err != nil {
		return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "unmarshal failure", Cause: err}
	}
	if m.SchemaVersion != "1.0" {
		return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "schema_version must be '1.0'"}
	}
	if !effectIDPattern.MatchString(m.ProjectID) {
		return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid project_id"}
	}
	if !effectIDPattern.MatchString(m.UnitID) {
		return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid unit_id"}
	}
	for _, ref := range []string{m.PlanRef, m.SnapshotRef} {
		if !digestPattern.MatchString(ref) {
			return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid digest ref: " + ref}
		}
	}
	seenIDs := make(map[string]struct{}, len(m.Effects))
	for i, eff := range m.Effects {
		if !effectIDPattern.MatchString(eff.EffectID) {
			return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("effect %d: invalid effect_id", i)}
		}
		if _, dup := seenIDs[eff.EffectID]; dup {
			return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("duplicate effect_id: %s", eff.EffectID)}
		}
		seenIDs[eff.EffectID] = struct{}{}
		if _, ok := allowedEffectDimensions[eff.Dimension]; !ok {
			return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("effect %d: invalid dimension %q", i, eff.Dimension)}
		}
		if _, ok := allowedEffectClasses[string(eff.EffectClass)]; !ok {
			return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("effect %d: invalid effect_class %q", i, eff.EffectClass)}
		}
		if !digestPattern.MatchString(eff.PayloadRef) {
			return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: fmt.Sprintf("effect %d: invalid payload_ref", i)}
		}
	}
	if !digestPattern.MatchString(m.ProvenanceRef) {
		return &EffectSafetyError{Code: "ERR_INVALID_MANIFEST", Reason: "invalid provenance_ref"}
	}
	return nil
}

// EffectSafetyPolicy governs environment safety constraints during effect replay and verification.
type EffectSafetyPolicy struct {
	IsProductionEnvironment bool
	AllowIrreversible       bool
}

// CheckEffectSafety validates that the effect is permissible under the active environment policy.
func (p *EffectSafetyPolicy) CheckEffectSafety(eff BusinessEffectRecord) error {
	if p.IsProductionEnvironment {
		if eff.EffectClass != EffectClassReadOnly && eff.EffectClass != EffectClassPure {
			return &EffectSafetyError{
				Code:   "ERR_PRODUCTION_EFFECT_REFUSED",
				Reason: fmt.Sprintf("effect %s has class %s which is prohibited in production", eff.EffectID, eff.EffectClass),
			}
		}
	}
	if eff.EffectClass == EffectClassIrreversible && !p.AllowIrreversible {
		return &EffectSafetyError{
			Code:   "ERR_IRREVERSIBLE_EFFECT_REFUSED",
			Reason: fmt.Sprintf("irreversible effect %s prohibited without explicit isolated sandbox authorization", eff.EffectID),
		}
	}
	return nil
}

// VerifyEffectCausalityAndOrdering checks that effects strictly follow monotonic
// sequence ordering and that all causal parents precede their dependents.
func VerifyEffectCausalityAndOrdering(effects []BusinessEffectRecord) error {
	seenIndices := make(map[int64]string)
	seenIDs := make(map[string]int64)

	var lastIndex int64 = -1
	for _, eff := range effects {
		if eff.SequenceIndex <= lastIndex {
			return &EffectSafetyError{
				Code:   "ERR_ORDERING_VIOLATION",
				Reason: fmt.Sprintf("effect %s has sequence_index %d <= preceding index %d (must be strictly increasing)", eff.EffectID, eff.SequenceIndex, lastIndex),
			}
		}
		if existing, dup := seenIndices[eff.SequenceIndex]; dup {
			return &EffectSafetyError{
				Code:   "ERR_ORDERING_VIOLATION",
				Reason: fmt.Sprintf("duplicate sequence_index %d used by %s and %s", eff.SequenceIndex, existing, eff.EffectID),
			}
		}
		seenIndices[eff.SequenceIndex] = eff.EffectID
		seenIDs[eff.EffectID] = eff.SequenceIndex
		lastIndex = eff.SequenceIndex

		// Verify causal parent exists and precedes this effect
		if eff.CausalParentID != nil {
			parentIndex, found := seenIDs[*eff.CausalParentID]
			if !found {
				return &EffectSafetyError{
					Code:   "ERR_CAUSAL_INVERSION",
					Reason: fmt.Sprintf("effect %s references causal parent %s which has not appeared yet (causal inversion)", eff.EffectID, *eff.CausalParentID),
				}
			}
			if parentIndex >= eff.SequenceIndex {
				return &EffectSafetyError{
					Code:   "ERR_CAUSAL_INVERSION",
					Reason: fmt.Sprintf("effect %s sequence_index %d <= causal parent %s sequence_index %d", eff.EffectID, eff.SequenceIndex, *eff.CausalParentID, parentIndex),
				}
			}
		}
	}
	return nil
}

// EffectCardinalityResult captures the comparison outcome for business effects.
type EffectCardinalityResult struct {
	Verdict     string   // "PASS" | "MISMATCH" | "UNKNOWN"
	Reason      string
	MismatchIDs []string
}

// VerifyEffectCardinality compares source and target effects, enforcing:
// - Each source business effect must be matched by a target business effect.
// - Retries with the same BusinessApplicationID are treated as one application.
// - Duplicate applied business effects yield MISMATCH.
// - Missing business effects yield MISMATCH.
// - Misbound destination or wrong business key yields MISMATCH.
func VerifyEffectCardinality(sourceEffects, targetEffects []BusinessEffectRecord) *EffectCardinalityResult {
	res := &EffectCardinalityResult{
		Verdict:     "PASS",
		MismatchIDs: []string{},
	}

	// Group source applied effects by BusinessApplicationID
	srcAppMap := make(map[string][]BusinessEffectRecord)
	for _, se := range sourceEffects {
		if se.Applied {
			srcAppMap[se.BusinessApplicationID] = append(srcAppMap[se.BusinessApplicationID], se)
		}
	}

	// Group target applied effects by BusinessApplicationID
	tgtAppMap := make(map[string][]BusinessEffectRecord)
	for _, te := range targetEffects {
		if te.Applied {
			tgtAppMap[te.BusinessApplicationID] = append(tgtAppMap[te.BusinessApplicationID], te)
		}
	}

	// 1. Check for missing business effects in target
	for appID, srcList := range srcAppMap {
		tgtList, found := tgtAppMap[appID]
		if !found || len(tgtList) == 0 {
			res.Verdict = "MISMATCH"
			res.MismatchIDs = append(res.MismatchIDs, "mismatch-missing-effect")
			res.Reason = fmt.Sprintf("missing business effect in target for application ID %q", appID)
			return res
		}

		// 2. Check for duplicate applied effects
		if len(tgtList) > len(srcList) {
			res.Verdict = "MISMATCH"
			res.MismatchIDs = append(res.MismatchIDs, "mismatch-duplicate-effect")
			res.Reason = fmt.Sprintf("duplicate business effect: cardinality violated for application ID %q (source %d vs target %d)", appID, len(srcList), len(tgtList))
			return res
		}

		// 3. Check destination and business key fidelity
		srcRepresentative := srcList[0]
		tgtRepresentative := tgtList[0]
		if srcRepresentative.DestinationID != tgtRepresentative.DestinationID {
			res.Verdict = "MISMATCH"
			res.MismatchIDs = append(res.MismatchIDs, "mismatch-misbound-effect")
			res.Reason = fmt.Sprintf("misbound destination for application ID %q: source %q vs target %q", appID, srcRepresentative.DestinationID, tgtRepresentative.DestinationID)
			return res
		}
		if srcRepresentative.BusinessKey != tgtRepresentative.BusinessKey {
			res.Verdict = "MISMATCH"
			res.MismatchIDs = append(res.MismatchIDs, "mismatch-misbound-effect")
			res.Reason = fmt.Sprintf("business key divergence for application ID %q: source %q vs target %q", appID, srcRepresentative.BusinessKey, tgtRepresentative.BusinessKey)
			return res
		}
	}

	// 4. Check for unprompted extra target effects (not in source)
	for appID := range tgtAppMap {
		if _, found := srcAppMap[appID]; !found {
			res.Verdict = "MISMATCH"
			res.MismatchIDs = append(res.MismatchIDs, "mismatch-unexpected-effect")
			res.Reason = fmt.Sprintf("unexpected target business effect for application ID %q not present in source", appID)
			return res
		}
	}

	return res
}

// EffectReconstructionInput bundles closure inputs for Unit 7 effect reconstruction.
type EffectReconstructionInput struct {
	Admission    *GenerationAdmission
	TargetModel  *TargetModelDocument
	UnitID       string
	Effects      []BusinessEffectRecord
	SafetyPolicy EffectSafetyPolicy
}

// EffectReconstructionOutput carries the reconstructed effect artifacts and publication manifest.
type EffectReconstructionOutput struct {
	Manifest *EffectPublicationManifest
}

// EffectReconstructor coordinates safe, deterministic effect reconstruction for M4 Unit 7.
type EffectReconstructor struct{}

// NewEffectReconstructor creates an EffectReconstructor.
func NewEffectReconstructor() *EffectReconstructor {
	return &EffectReconstructor{}
}

// Reconstruct validates the input effects, enforces safety policy and causal ordering,
// and produces a deterministic EffectPublicationManifest.
func (r *EffectReconstructor) Reconstruct(input EffectReconstructionInput) (*EffectReconstructionOutput, error) {
	if input.Admission == nil {
		return nil, &EffectSafetyError{Code: "ERR_ADMISSION", Reason: "nil generation admission"}
	}
	if input.Admission.Outcome != "ADMITTED" {
		return nil, &EffectSafetyError{
			Code:   "ERR_ADMISSION",
			Reason: fmt.Sprintf("admission outcome is %q, must be ADMITTED", input.Admission.Outcome),
		}
	}
	if input.UnitID == "" {
		return nil, &EffectSafetyError{Code: "ERR_INVALID_UNIT", Reason: "empty unit_id"}
	}
	if input.TargetModel == nil {
		return nil, &EffectSafetyError{Code: "ERR_TARGET_MODEL", Reason: "nil target model"}
	}
	if len(input.Effects) == 0 {
		return nil, &EffectSafetyError{Code: "ERR_NO_EFFECTS", Reason: "effects list must be non-empty"}
	}

	// 1. Enforce effect safety policy for every effect
	for _, eff := range input.Effects {
		if err := input.SafetyPolicy.CheckEffectSafety(eff); err != nil {
			return nil, err
		}
	}

	// 2. Sort effects by sequence_index for deterministic causality verification
	sortedEffects := append([]BusinessEffectRecord{}, input.Effects...)
	sort.Slice(sortedEffects, func(i, j int) bool {
		return sortedEffects[i].SequenceIndex < sortedEffects[j].SequenceIndex
	})

	// 3. Verify monotonic sequence ordering and causal consistency
	if err := VerifyEffectCausalityAndOrdering(sortedEffects); err != nil {
		return nil, err
	}

	// 4. Build EffectPublicationManifest
	projID := input.TargetModel.ProjectID
	if projID == "" {
		projID = "proj-default"
	}
	manifest := &EffectPublicationManifest{
		SchemaVersion: "1.0",
		ProjectID:     projID,
		UnitID:        input.UnitID,
		PlanRef:       input.Admission.PlanRef,
		SnapshotRef:   input.Admission.SnapshotRef,
		Effects:       sortedEffects,
	}

	// Self-provenance digest
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
	sum := sha256.Sum256(canonMan)
	manifest.ProvenanceRef = "sha256:" + hex.EncodeToString(sum[:])

	// Validate with canonical validator
	manBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	if err := ValidateEffectPublicationManifest(manBytes); err != nil {
		return nil, fmt.Errorf("effect manifest validation: %w", err)
	}

	return &EffectReconstructionOutput{
		Manifest: manifest,
	}, nil
}

package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// Primitive capabilities that do not require external qualified provider bindings.
// Per TARGET-ARCHITECTURE.md and BEHAVIOR-MODEL.md, core primitives are built-in.
var primitiveCapabilities = map[string]struct{}{
	"http":       {},
	"routing":    {},
	"json":       {},
	"filesystem": {},
	"cli":        {},
	"logging":    {},
	"clock":      {},
	"math":       {},
	"standard":   {},
}

// TargetModelError describes an invalid target implementation model specification record.
type TargetModelError struct {
	Reason string
	Cause  error
}

// Error identifies the failed target model obligation.
func (e *TargetModelError) Error() string {
	return "invalid target model: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *TargetModelError) Unwrap() error { return e.Cause }

// LoweringBinding binds an exact capability requirement to an independently qualified
// target lowering provider, per TARGET-ARCHITECTURE.md v2.1.
type LoweringBinding struct {
	CapabilityID        string   `json:"capability_id"`
	ProviderManifestRef string   `json:"provider_manifest_ref"`
	QualificationRef    string   `json:"qualification_ref"`
	ObligationRefs      []string `json:"obligation_refs"`
	InputShapeRef       string   `json:"input_shape_ref"`
	OutputShapeRef      string   `json:"output_shape_ref"`
}

// TargetModelComponent defines an owned architectural component in the target implementation,
// including its unique owner unit, consumers, safe paths, contract mappings, and lowering bindings.
type TargetModelComponent struct {
	ComponentID          string            `json:"component_id"`
	OwnerUnitID          string            `json:"owner_unit_id"`
	ConsumerUnitIDs      []string          `json:"consumer_unit_ids"`
	Paths                []string          `json:"paths"`
	ContractRefs         []string          `json:"contract_refs"`
	LoweringChoice       string            `json:"lowering_choice"`
	RequiredCapabilities []string          `json:"required_capabilities"`
	ConstraintRefs       []string          `json:"constraint_refs"`
	LoweringBindings     []LoweringBinding `json:"lowering_bindings,omitempty"`
}

// CapabilityGap records an unresolved or unsupported capability requirement,
// indicating whether it blocks plan feasibility.
type CapabilityGap struct {
	Code         string   `json:"code"`
	ComponentIDs []string `json:"component_ids"`
	EvidenceRefs []string `json:"evidence_refs"`
	Blocking     bool     `json:"blocking"`
}

// TargetModelDocument represents the complete Target Implementation Model (TIM) document
// conforming to specs/migration/target-model.schema.json (version 2.0 or 2.1).
type TargetModelDocument struct {
	SchemaVersion          string                 `json:"schema_version"`
	ModelID                string                 `json:"model_id"`
	Revision               int                    `json:"revision"`
	ProjectID              string                 `json:"project_id"`
	BehaviorIRRef          string                 `json:"behavior_ir_ref"`
	ResolvedEnvironmentRef string                 `json:"resolved_environment_ref"`
	Components             []TargetModelComponent `json:"components"`
	CapabilityGaps         []CapabilityGap        `json:"capability_gaps"`
}

// ComponentByOwner retrieves the component owned by the given owner unit ID.
func (d *TargetModelDocument) ComponentByOwner(ownerUnitID string) (*TargetModelComponent, bool) {
	for i := range d.Components {
		if d.Components[i].OwnerUnitID == ownerUnitID {
			return &d.Components[i], true
		}
	}
	return nil, false
}

// ComponentByID retrieves the component with the given component ID.
func (d *TargetModelDocument) ComponentByID(componentID string) (*TargetModelComponent, bool) {
	for i := range d.Components {
		if d.Components[i].ComponentID == componentID {
			return &d.Components[i], true
		}
	}
	return nil, false
}

// HasBlockingGaps reports whether the model contains any blocking capability gaps.
func (d *TargetModelDocument) HasBlockingGaps() bool {
	for _, gap := range d.CapabilityGaps {
		if gap.Blocking {
			return true
		}
	}
	return false
}

// BlockingGaps returns all capability gaps marked as blocking.
func (d *TargetModelDocument) BlockingGaps() []CapabilityGap {
	var blocking []CapabilityGap
	for _, gap := range d.CapabilityGaps {
		if gap.Blocking {
			blocking = append(blocking, gap)
		}
	}
	return blocking
}

// ConsumersForUnit returns all consumer unit IDs that depend on the component owned by ownerUnitID.
func (d *TargetModelDocument) ConsumersForUnit(ownerUnitID string) []string {
	comp, ok := d.ComponentByOwner(ownerUnitID)
	if !ok {
		return nil
	}
	res := make([]string, len(comp.ConsumerUnitIDs))
	copy(res, comp.ConsumerUnitIDs)
	return res
}

// Validate checks all schema and domain invariant obligations for the model document.
func (d *TargetModelDocument) Validate() error {
	raw, err := d.MarshalJSON()
	if err != nil {
		return &TargetModelError{Reason: "json marshal", Cause: err}
	}
	return ValidateTargetModel(raw)
}

// CanonicalJSON produces canonical rwt-cjson-1 bytes of the model.
func (d *TargetModelDocument) CanonicalJSON() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	raw, err := d.MarshalJSON()
	if err != nil {
		return nil, &TargetModelError{Reason: "json marshal", Cause: err}
	}
	return CanonicalJSON(raw)
}

// Digest computes the canonical sha256 digest of the model document with "sha256:" prefix.
func (d *TargetModelDocument) Digest() (string, error) {
	canonicalBytes, err := d.CanonicalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(canonicalBytes)
	return "sha256:" + hex.EncodeToString(h[:]), nil
}

// MarshalJSON marshals the document respecting schema 2.0 vs 2.1 differences
// (2.1 requires lowering_bindings in components, 2.0 omits it).
func (d *TargetModelDocument) MarshalJSON() ([]byte, error) {
	type componentV20JSON struct {
		ComponentID          string   `json:"component_id"`
		OwnerUnitID          string   `json:"owner_unit_id"`
		ConsumerUnitIDs      []string `json:"consumer_unit_ids"`
		Paths                []string `json:"paths"`
		ContractRefs         []string `json:"contract_refs"`
		LoweringChoice       string   `json:"lowering_choice"`
		RequiredCapabilities []string `json:"required_capabilities"`
		ConstraintRefs       []string `json:"constraint_refs"`
	}

	type componentV21JSON struct {
		ComponentID          string            `json:"component_id"`
		OwnerUnitID          string            `json:"owner_unit_id"`
		ConsumerUnitIDs      []string          `json:"consumer_unit_ids"`
		Paths                []string          `json:"paths"`
		ContractRefs         []string          `json:"contract_refs"`
		LoweringChoice       string            `json:"lowering_choice"`
		RequiredCapabilities []string          `json:"required_capabilities"`
		ConstraintRefs       []string          `json:"constraint_refs"`
		LoweringBindings     []LoweringBinding `json:"lowering_bindings"`
	}

	type targetModelJSON struct {
		SchemaVersion          string          `json:"schema_version"`
		ModelID                string          `json:"model_id"`
		Revision               int             `json:"revision"`
		ProjectID              string          `json:"project_id"`
		BehaviorIRRef          string          `json:"behavior_ir_ref"`
		ResolvedEnvironmentRef string          `json:"resolved_environment_ref"`
		Components             interface{}     `json:"components"`
		CapabilityGaps         []CapabilityGap `json:"capability_gaps"`
	}

	gaps := d.CapabilityGaps
	if gaps == nil {
		gaps = []CapabilityGap{}
	} else {
		cleanedGaps := make([]CapabilityGap, len(gaps))
		for i, g := range gaps {
			cg := g
			if cg.ComponentIDs == nil {
				cg.ComponentIDs = []string{}
			}
			if cg.EvidenceRefs == nil {
				cg.EvidenceRefs = []string{}
			}
			cleanedGaps[i] = cg
		}
		gaps = cleanedGaps
	}

	if d.SchemaVersion == "2.0" {
		v20Comps := make([]componentV20JSON, len(d.Components))
		for i, c := range d.Components {
			consumers := c.ConsumerUnitIDs
			if consumers == nil {
				consumers = []string{}
			}
			paths := c.Paths
			if paths == nil {
				paths = []string{}
			}
			contracts := c.ContractRefs
			if contracts == nil {
				contracts = []string{}
			}
			caps := c.RequiredCapabilities
			if caps == nil {
				caps = []string{}
			}
			constraints := c.ConstraintRefs
			if constraints == nil {
				constraints = []string{}
			}
			v20Comps[i] = componentV20JSON{
				ComponentID:          c.ComponentID,
				OwnerUnitID:          c.OwnerUnitID,
				ConsumerUnitIDs:      consumers,
				Paths:                paths,
				ContractRefs:         contracts,
				LoweringChoice:       c.LoweringChoice,
				RequiredCapabilities: caps,
				ConstraintRefs:       constraints,
			}
		}
		return json.Marshal(targetModelJSON{
			SchemaVersion:          d.SchemaVersion,
			ModelID:                d.ModelID,
			Revision:               d.Revision,
			ProjectID:              d.ProjectID,
			BehaviorIRRef:          d.BehaviorIRRef,
			ResolvedEnvironmentRef: d.ResolvedEnvironmentRef,
			Components:             v20Comps,
			CapabilityGaps:         gaps,
		})
	}

	// Default to 2.1
	v21Comps := make([]componentV21JSON, len(d.Components))
	for i, c := range d.Components {
		consumers := c.ConsumerUnitIDs
		if consumers == nil {
			consumers = []string{}
		}
		paths := c.Paths
		if paths == nil {
			paths = []string{}
		}
		contracts := c.ContractRefs
		if contracts == nil {
			contracts = []string{}
		}
		caps := c.RequiredCapabilities
		if caps == nil {
			caps = []string{}
		}
		constraints := c.ConstraintRefs
		if constraints == nil {
			constraints = []string{}
		}
		bindings := c.LoweringBindings
		if bindings == nil {
			bindings = []LoweringBinding{}
		} else {
			cleanedBindings := make([]LoweringBinding, len(bindings))
			for j, b := range bindings {
				cb := b
				if cb.ObligationRefs == nil {
					cb.ObligationRefs = []string{}
				}
				cleanedBindings[j] = cb
			}
			bindings = cleanedBindings
		}
		v21Comps[i] = componentV21JSON{
			ComponentID:          c.ComponentID,
			OwnerUnitID:          c.OwnerUnitID,
			ConsumerUnitIDs:      consumers,
			Paths:                paths,
			ContractRefs:         contracts,
			LoweringChoice:       c.LoweringChoice,
			RequiredCapabilities: caps,
			ConstraintRefs:       constraints,
			LoweringBindings:     bindings,
		}
	}

	return json.Marshal(targetModelJSON{
		SchemaVersion:          d.SchemaVersion,
		ModelID:                d.ModelID,
		Revision:               d.Revision,
		ProjectID:              d.ProjectID,
		BehaviorIRRef:          d.BehaviorIRRef,
		ResolvedEnvironmentRef: d.ResolvedEnvironmentRef,
		Components:             v21Comps,
		CapabilityGaps:         gaps,
	})
}

// ValidateTargetModel validates raw JSON against specs/migration/target-model.schema.json
// and domain semantic invariants (unique owner_unit_id, non-overlapping paths, lowering bindings).
func ValidateTargetModel(raw []byte) error {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return &TargetModelError{Reason: "canonical JSON", Cause: err}
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &record); err != nil || record == nil {
		return &TargetModelError{Reason: "object fields", Cause: err}
	}

	versionRaw, present := record["schema_version"]
	if !present {
		return &TargetModelError{Reason: "missing schema_version"}
	}
	version := profileString(versionRaw)
	if version != "2.0" && version != "2.1" {
		return &TargetModelError{Reason: "unsupported schema_version: " + version}
	}

	requiredFields := []string{
		"schema_version",
		"model_id",
		"revision",
		"project_id",
		"behavior_ir_ref",
		"resolved_environment_ref",
		"components",
		"capability_gaps",
	}

	for _, f := range requiredFields {
		if _, ok := record[f]; !ok {
			return &TargetModelError{Reason: "missing required field: " + f}
		}
	}
	if len(record) != len(requiredFields) {
		return &TargetModelError{Reason: "unexpected field in target model root"}
	}

	modelID := profileString(record["model_id"])
	if !validAssessmentReportID(modelID) {
		return &TargetModelError{Reason: "invalid model_id format"}
	}

	var revision int
	if err := json.Unmarshal(record["revision"], &revision); err != nil || revision < 1 {
		return &TargetModelError{Reason: "revision must be integer >= 1"}
	}

	projectID := profileString(record["project_id"])
	if !validAssessmentReportID(projectID) {
		return &TargetModelError{Reason: "invalid project_id format"}
	}

	if !validProfileRef(record["behavior_ir_ref"], false) {
		return &TargetModelError{Reason: "invalid behavior_ir_ref format"}
	}

	if !validProfileRef(record["resolved_environment_ref"], false) {
		return &TargetModelError{Reason: "invalid resolved_environment_ref format"}
	}

	// Validate components
	var rawComponents []json.RawMessage
	if err := json.Unmarshal(record["components"], &rawComponents); err != nil || len(rawComponents) < 1 {
		return &TargetModelError{Reason: "components must be non-empty array"}
	}

	componentIDs := make(map[string]struct{}, len(rawComponents))
	ownerUnitIDs := make(map[string]struct{}, len(rawComponents))
	ownedPaths := make(map[string]string) // path -> componentID

	type parsedComp struct {
		id                   string
		owner                string
		requiredCapabilities []string
		loweringBindings     []string
	}
	parsedList := make([]parsedComp, 0, len(rawComponents))

	for i, rawComp := range rawComponents {
		var compRecord map[string]json.RawMessage
		if err := json.Unmarshal(rawComp, &compRecord); err != nil || compRecord == nil {
			return &TargetModelError{Reason: fmt.Sprintf("component[%d] must be an object", i)}
		}

		compReq := []string{
			"component_id",
			"owner_unit_id",
			"consumer_unit_ids",
			"paths",
			"contract_refs",
			"lowering_choice",
			"required_capabilities",
			"constraint_refs",
		}
		if version == "2.1" {
			compReq = append(compReq, "lowering_bindings")
		}

		for _, req := range compReq {
			if _, ok := compRecord[req]; !ok {
				return &TargetModelError{Reason: fmt.Sprintf("component[%d] missing required field %q", i, req)}
			}
		}
		if len(compRecord) != len(compReq) {
			return &TargetModelError{Reason: fmt.Sprintf("component[%d] contains unexpected fields", i)}
		}

		compID := profileString(compRecord["component_id"])
		if !validAssessmentReportID(compID) {
			return &TargetModelError{Reason: fmt.Sprintf("component[%d] invalid component_id format", i)}
		}
		if _, exists := componentIDs[compID]; exists {
			return &TargetModelError{Reason: fmt.Sprintf("duplicate component_id %q", compID)}
		}
		componentIDs[compID] = struct{}{}

		ownerID := profileString(compRecord["owner_unit_id"])
		if !validAssessmentReportID(ownerID) {
			return &TargetModelError{Reason: fmt.Sprintf("component[%d] invalid owner_unit_id format", i)}
		}
		// Invariant: Component owner_unit_id is unique across components (TARGET-ARCHITECTURE.md:9)
		if _, exists := ownerUnitIDs[ownerID]; exists {
			return &TargetModelError{Reason: fmt.Sprintf("duplicate owner_unit_id %q: each component must have a unique owner_unit_id", ownerID)}
		}
		ownerUnitIDs[ownerID] = struct{}{}

		// Validate consumer_unit_ids
		var consumers []string
		if err := json.Unmarshal(compRecord["consumer_unit_ids"], &consumers); err != nil {
			return &TargetModelError{Reason: fmt.Sprintf("component[%s] consumer_unit_ids must be array of strings", compID)}
		}
		seenConsumers := make(map[string]struct{}, len(consumers))
		for _, c := range consumers {
			if !validAssessmentReportID(c) {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] invalid consumer_unit_id format %q", compID, c)}
			}
			if _, exists := seenConsumers[c]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] duplicate consumer_unit_id %q", compID, c)}
			}
			seenConsumers[c] = struct{}{}
			if c == ownerID {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] consumer_unit_ids cannot contain its own owner_unit_id %q", compID, ownerID)}
			}
		}

		// Validate paths
		var paths []string
		if err := json.Unmarshal(compRecord["paths"], &paths); err != nil {
			return &TargetModelError{Reason: fmt.Sprintf("component[%s] paths must be array of strings", compID)}
		}
		seenCompPaths := make(map[string]struct{}, len(paths))
		for _, p := range paths {
			if len(p) == 0 {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] path cannot be empty", compID)}
			}
			normPath := normalizePath(p)
			if _, exists := seenCompPaths[normPath]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] duplicate path %q", compID, p)}
			}
			seenCompPaths[normPath] = struct{}{}

			// Check path overlap across components
			for otherPath, otherOwner := range ownedPaths {
				if normPath == otherPath {
					return &TargetModelError{Reason: fmt.Sprintf("path conflict: path %q claimed by both component %q and %q", p, compID, otherOwner)}
				}
				// Check directory prefix conflict: e.g. "app/Http" vs "app/Http/Kernel.php"
				if strings.HasPrefix(normPath, otherPath+"/") {
					return &TargetModelError{Reason: fmt.Sprintf("ambiguous path overlap: component %q path %q is inside component %q path %q", compID, p, otherOwner, otherPath)}
				}
				if strings.HasPrefix(otherPath, normPath+"/") {
					return &TargetModelError{Reason: fmt.Sprintf("ambiguous path overlap: component %q path %q contains component %q path %q", compID, p, otherOwner, otherPath)}
				}
			}
			ownedPaths[normPath] = compID
		}

		// Validate contract_refs (minItems: 1)
		var contracts []string
		if err := json.Unmarshal(compRecord["contract_refs"], &contracts); err != nil || len(contracts) < 1 {
			return &TargetModelError{Reason: fmt.Sprintf("component[%s] contract_refs must be non-empty array", compID)}
		}
		seenContracts := make(map[string]struct{}, len(contracts))
		for _, cr := range contracts {
			if !isSHA256Ref(cr) {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] invalid contract_ref %q", compID, cr)}
			}
			if _, exists := seenContracts[cr]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] duplicate contract_ref %q", compID, cr)}
			}
			seenContracts[cr] = struct{}{}
		}

		// Validate lowering_choice
		var choice string
		if err := json.Unmarshal(compRecord["lowering_choice"], &choice); err != nil || len(choice) < 1 {
			return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_choice must be non-empty string", compID)}
		}

		// Validate required_capabilities
		var caps []string
		if err := json.Unmarshal(compRecord["required_capabilities"], &caps); err != nil {
			return &TargetModelError{Reason: fmt.Sprintf("component[%s] required_capabilities must be array of strings", compID)}
		}
		seenCaps := make(map[string]struct{}, len(caps))
		for _, c := range caps {
			if len(c) < 1 {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] required_capabilities entry cannot be empty", compID)}
			}
			if _, exists := seenCaps[c]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] duplicate required_capability %q", compID, c)}
			}
			seenCaps[c] = struct{}{}
		}

		// Validate constraint_refs
		var constraints []string
		if err := json.Unmarshal(compRecord["constraint_refs"], &constraints); err != nil {
			return &TargetModelError{Reason: fmt.Sprintf("component[%s] constraint_refs must be array of strings", compID)}
		}
		seenConstraints := make(map[string]struct{}, len(constraints))
		for _, c := range constraints {
			if !isSHA256Ref(c) {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] invalid constraint_ref %q", compID, c)}
			}
			if _, exists := seenConstraints[c]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] duplicate constraint_ref %q", compID, c)}
			}
			seenConstraints[c] = struct{}{}
		}

		var boundCapIDs []string
		// Validate lowering_bindings for 2.1
		if version == "2.1" {
			var bindings []json.RawMessage
			if err := json.Unmarshal(compRecord["lowering_bindings"], &bindings); err != nil {
				return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_bindings must be array", compID)}
			}
			seenBoundCaps := make(map[string]struct{}, len(bindings))
			for bIdx, rawBinding := range bindings {
				var bRecord map[string]json.RawMessage
				if err := json.Unmarshal(rawBinding, &bRecord); err != nil || bRecord == nil {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] must be an object", compID, bIdx)}
				}
				bReq := []string{
					"capability_id",
					"provider_manifest_ref",
					"qualification_ref",
					"obligation_refs",
					"input_shape_ref",
					"output_shape_ref",
				}
				for _, r := range bReq {
					if _, ok := bRecord[r]; !ok {
						return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] missing required field %q", compID, bIdx, r)}
					}
				}
				if len(bRecord) != len(bReq) {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] contains unexpected fields", compID, bIdx)}
				}

				capID := profileString(bRecord["capability_id"])
				if !validAssessmentReportID(capID) {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] invalid capability_id %q", compID, bIdx, capID)}
				}
				if _, exists := seenBoundCaps[capID]; exists {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] duplicate lowering_binding capability_id %q", compID, capID)}
				}
				seenBoundCaps[capID] = struct{}{}
				boundCapIDs = append(boundCapIDs, capID)

				if !validProfileRef(bRecord["provider_manifest_ref"], false) {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] invalid provider_manifest_ref", compID, bIdx)}
				}
				if !validProfileRef(bRecord["qualification_ref"], false) {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] invalid qualification_ref", compID, bIdx)}
				}
				if !validProfileRef(bRecord["input_shape_ref"], false) {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] invalid input_shape_ref", compID, bIdx)}
				}
				if !validProfileRef(bRecord["output_shape_ref"], false) {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] invalid output_shape_ref", compID, bIdx)}
				}

				var obligations []string
				if err := json.Unmarshal(bRecord["obligation_refs"], &obligations); err != nil || len(obligations) < 1 {
					return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] obligation_refs must be non-empty array", compID, bIdx)}
				}
				for _, obl := range obligations {
					if !isSHA256Ref(obl) {
						return &TargetModelError{Reason: fmt.Sprintf("component[%s] lowering_binding[%d] invalid obligation_ref %q", compID, bIdx, obl)}
					}
				}
			}
		}

		parsedList = append(parsedList, parsedComp{
			id:                   compID,
			owner:                ownerID,
			requiredCapabilities: caps,
			loweringBindings:     boundCapIDs,
		})
	}

	// Validate capability_gaps
	var rawGaps []json.RawMessage
	if err := json.Unmarshal(record["capability_gaps"], &rawGaps); err != nil {
		return &TargetModelError{Reason: "capability_gaps must be an array"}
	}

	gapsByComponent := make(map[string]map[string]struct{}) // compID -> set of gap codes

	for i, rawGap := range rawGaps {
		var gapRecord map[string]json.RawMessage
		if err := json.Unmarshal(rawGap, &gapRecord); err != nil || gapRecord == nil {
			return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] must be an object", i)}
		}
		gapReq := []string{"code", "component_ids", "evidence_refs", "blocking"}
		for _, req := range gapReq {
			if _, ok := gapRecord[req]; !ok {
				return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] missing required field %q", i, req)}
			}
		}
		if len(gapRecord) != len(gapReq) {
			return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] contains unexpected fields", i)}
		}

		var code string
		if err := json.Unmarshal(gapRecord["code"], &code); err != nil || len(code) < 1 {
			return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] code must be non-empty string", i)}
		}

		var cIDs []string
		if err := json.Unmarshal(gapRecord["component_ids"], &cIDs); err != nil || len(cIDs) < 1 {
			return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] component_ids must be non-empty array", i)}
		}
		seenGapComps := make(map[string]struct{}, len(cIDs))
		for _, cid := range cIDs {
			if !validAssessmentReportID(cid) {
				return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] invalid component_id %q", i, cid)}
			}
			if _, exists := seenGapComps[cid]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] duplicate component_id %q", i, cid)}
			}
			seenGapComps[cid] = struct{}{}

			if _, exists := componentIDs[cid]; !exists {
				return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] references unknown component_id %q", i, cid)}
			}

			if gapsByComponent[cid] == nil {
				gapsByComponent[cid] = make(map[string]struct{})
			}
			gapsByComponent[cid][code] = struct{}{}
		}

		var evRefs []string
		if err := json.Unmarshal(gapRecord["evidence_refs"], &evRefs); err != nil || len(evRefs) < 1 {
			return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] evidence_refs must be non-empty array", i)}
		}
		seenEv := make(map[string]struct{}, len(evRefs))
		for _, ev := range evRefs {
			if !isSHA256Ref(ev) {
				return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] invalid evidence_ref %q", i, ev)}
			}
			if _, exists := seenEv[ev]; exists {
				return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] duplicate evidence_ref %q", i, ev)}
			}
			seenEv[ev] = struct{}{}
		}

		var blocking bool
		if err := json.Unmarshal(gapRecord["blocking"], &blocking); err != nil {
			return &TargetModelError{Reason: fmt.Sprintf("capability_gap[%d] blocking must be boolean", i)}
		}
	}

	// Verify non-primitive required capabilities in 2.1:
	// "Primitive core capabilities need no provider binding; each nonprimitive requirement
	// (password verification, session codec, template rendering) must resolve exactly one qualified
	// binding or remain a blocking gap." (TARGET-ARCHITECTURE.md:23)
	if version == "2.1" {
		for _, comp := range parsedList {
			boundMap := make(map[string]struct{}, len(comp.loweringBindings))
			for _, b := range comp.loweringBindings {
				boundMap[b] = struct{}{}
			}

			for _, capID := range comp.requiredCapabilities {
				if _, isPrimitive := primitiveCapabilities[capID]; isPrimitive {
					continue
				}
				// Non-primitive: must either be bound or covered by capability_gaps
				if _, isBound := boundMap[capID]; isBound {
					continue
				}
				compGaps := gapsByComponent[comp.id]
				hasGap := false
				for gapCode := range compGaps {
					if strings.EqualFold(gapCode, capID) || strings.Contains(strings.ToLower(gapCode), strings.ToLower(capID)) {
						hasGap = true
						break
					}
				}
				if !hasGap {
					return &TargetModelError{
						Reason: fmt.Sprintf("component %q has non-primitive capability %q without a qualified lowering binding or declared capability gap (TARGET-ARCHITECTURE.md:23)", comp.id, capID),
					}
				}
			}
		}
	}

	return nil
}

// ParseTargetModel parses and validates a raw JSON target implementation model.
func ParseTargetModel(raw []byte) (*TargetModelDocument, error) {
	if err := ValidateTargetModel(raw); err != nil {
		return nil, err
	}
	var doc TargetModelDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, &TargetModelError{Reason: "json unmarshal", Cause: err}
	}
	return &doc, nil
}

func normalizePath(p string) string {
	cleaned := path.Clean(strings.ReplaceAll(p, "\\", "/"))
	return strings.Trim(cleaned, "/")
}

func isSHA256Ref(value string) bool {
	return strings.HasPrefix(value, "sha256:") && isManifestDigest(strings.TrimPrefix(value, "sha256:"))
}

// TargetModelBuilder constructs valid TargetModelDocuments incrementally.
type TargetModelBuilder struct {
	schemaVersion          string
	modelID                string
	revision               int
	projectID              string
	behaviorIRRef          string
	resolvedEnvironmentRef string
	components             []TargetModelComponent
	capabilityGaps         []CapabilityGap
}

// NewTargetModelBuilder creates a builder with default schema_version "2.1" and revision 1.
func NewTargetModelBuilder(modelID, projectID, behaviorIRRef, resolvedEnvRef string) *TargetModelBuilder {
	return &TargetModelBuilder{
		schemaVersion:          "2.1",
		modelID:                modelID,
		revision:               1,
		projectID:              projectID,
		behaviorIRRef:          behaviorIRRef,
		resolvedEnvironmentRef: resolvedEnvRef,
		components:             make([]TargetModelComponent, 0),
		capabilityGaps:         make([]CapabilityGap, 0),
	}
}

// WithSchemaVersion sets the schema version ("2.0" or "2.1").
func (b *TargetModelBuilder) WithSchemaVersion(version string) *TargetModelBuilder {
	b.schemaVersion = version
	return b
}

// WithRevision sets the model revision (>= 1).
func (b *TargetModelBuilder) WithRevision(revision int) *TargetModelBuilder {
	b.revision = revision
	return b
}

// AddComponent adds an architectural component with initialized non-nil slice fields.
func (b *TargetModelBuilder) AddComponent(comp TargetModelComponent) *TargetModelBuilder {
	if comp.ConsumerUnitIDs == nil {
		comp.ConsumerUnitIDs = []string{}
	}
	if comp.Paths == nil {
		comp.Paths = []string{}
	}
	if comp.ContractRefs == nil {
		comp.ContractRefs = []string{}
	}
	if comp.RequiredCapabilities == nil {
		comp.RequiredCapabilities = []string{}
	}
	if comp.ConstraintRefs == nil {
		comp.ConstraintRefs = []string{}
	}
	if comp.LoweringBindings == nil {
		comp.LoweringBindings = []LoweringBinding{}
	}
	b.components = append(b.components, comp)
	return b
}

// AddCapabilityGap adds a capability gap with initialized non-nil slice fields.
func (b *TargetModelBuilder) AddCapabilityGap(gap CapabilityGap) *TargetModelBuilder {
	if gap.ComponentIDs == nil {
		gap.ComponentIDs = []string{}
	}
	if gap.EvidenceRefs == nil {
		gap.EvidenceRefs = []string{}
	}
	b.capabilityGaps = append(b.capabilityGaps, gap)
	return b
}

// Build validates all invariants and returns the compiled TargetModelDocument.
func (b *TargetModelBuilder) Build() (*TargetModelDocument, error) {
	doc := &TargetModelDocument{
		SchemaVersion:          b.schemaVersion,
		ModelID:                b.modelID,
		Revision:               b.revision,
		ProjectID:              b.projectID,
		BehaviorIRRef:          b.behaviorIRRef,
		ResolvedEnvironmentRef: b.resolvedEnvironmentRef,
		Components:             b.components,
		CapabilityGaps:         b.capabilityGaps,
	}
	if err := doc.Validate(); err != nil {
		return nil, err
	}
	return doc, nil
}

// CreateSuccessor creates a successor TargetModelBuilder from an existing document,
// incrementing the revision. This models the requirement from TARGET-ARCHITECTURE.md:13:
// "Changing ownership/lowering/resolution creates a successor model and Gate 2."
func (d *TargetModelDocument) CreateSuccessor(newRevision int) *TargetModelBuilder {
	b := &TargetModelBuilder{
		schemaVersion:          d.SchemaVersion,
		modelID:                d.ModelID,
		revision:               newRevision,
		projectID:              d.ProjectID,
		behaviorIRRef:          d.BehaviorIRRef,
		resolvedEnvironmentRef: d.ResolvedEnvironmentRef,
		components:             make([]TargetModelComponent, len(d.Components)),
		capabilityGaps:         make([]CapabilityGap, len(d.CapabilityGaps)),
	}
	copy(b.components, d.Components)
	copy(b.capabilityGaps, d.CapabilityGaps)
	return b
}

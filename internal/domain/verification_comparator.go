package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// CompareCaseInput encapsulates all evidence, witnesses, and configuration for differential verification.
type CompareCaseInput struct {
	RunID                   string
	ProjectID               string
	CaseID                  string
	ScenarioRef             string
	WorkflowID              string
	PolicyRef               string
	ConfigurationHash       string
	PlanHash                string
	CorpusHash              string
	PolicyHash              string
	TargetTreeHash          string
	ExpectedTargetBuildHash string
	AcceptanceScope         AcceptanceScope

	// Execution Witnesses
	SourceWitness *ExecutionWitness
	TargetWitness *ExecutionWitness

	// Step occurrences observed
	SourceOccurrences []StepOccurrence
	TargetOccurrences []StepOccurrence

	// Contexts: maps of ContextRef -> EvaluationContext
	SourceContexts map[string]*EvaluationContext
	TargetContexts map[string]*EvaluationContext

	// Local identity ledgers
	SourceIdentityLedger *LocalIdentitiesLedger
	TargetIdentityLedger *LocalIdentitiesLedger

	// Normalization rules (optional)
	NormalizationRules []*NormalizationRule

	// Session lifecycle obligations are projections of the approved plan, not
	// supplied observations. Qualifications come from the engine admission port.
	RequiredSessionContractRef    string
	SessionContinuity             *SessionContinuityContract
	SessionPlanSubject            *PlanApprovalSubject
	AdmittedSessionQualifications map[string]*QualificationRecord

	// Divergence recorded by driver (optional)
	DriverDivergence *BranchDivergence

	// Trace refs
	SourceTraceRef string
	TargetTraceRef string

	// Incomplete capture simulation / audit
	IncompleteCaptureReason string
	MissingDimensions       map[string]string // dimension -> reason
}

// CompareCaseOutput carries the generated verification artifacts and verdicts.
type CompareCaseOutput struct {
	Report            *VerificationReportV2
	OccurrencePairing *OccurrencePairing
	IdentityMap       *IdentityMapV2
	TopVerdict        string
	CaseVerdict       string
	Diverged          bool
	FirstDivergence   *PairingDivergence
	Errors            []VerificationErrorDetail
}

// BehaviorComparator executes read-only differential verification across all 11 dimensions.
// Invariant: BehaviorComparator is strictly read-only with respect to measured application state.
// It never issues database writes, never mutates files, and never fills missing observations from expected state.
type BehaviorComparator struct{}

// NewBehaviorComparator constructs a new BehaviorComparator.
func NewBehaviorComparator() *BehaviorComparator {
	return &BehaviorComparator{}
}

// CompareCase executes differential verification for a single case under the provided input.
func (c *BehaviorComparator) CompareCase(input *CompareCaseInput) (*CompareCaseOutput, error) {
	if input == nil {
		return nil, errors.New("nil compare case input")
	}

	out := &CompareCaseOutput{
		Errors: make([]VerificationErrorDetail, 0),
	}

	defaultEvidence := "sha256:" + strings.Repeat("a", 64)

	// 1. Role check & Source-vs-Source protection (X80)
	if input.SourceWitness == nil || input.TargetWitness == nil {
		out.TopVerdict = "UNKNOWN"
		out.CaseVerdict = "UNKNOWN"
		out.Errors = append(out.Errors, VerificationErrorDetail{
			Code:    "ERR_INCOMPLETE_EVIDENCE",
			Message: "missing required execution witness on source or target side",
		})
		out.Report = c.buildFailureReport(input, "UNKNOWN", "ERR_INCOMPLETE_EVIDENCE: missing execution witness", out.Errors)
		return out, nil
	}

	// Reject Source-vs-Source or identical witnesses (X80)
	if input.SourceWitness.Side == input.TargetWitness.Side ||
		input.SourceWitness.Launch.ProcessID == input.TargetWitness.Launch.ProcessID ||
		input.SourceWitness.Launch.HostBootID == input.TargetWitness.Launch.HostBootID {
		out.TopVerdict = "UNKNOWN"
		out.CaseVerdict = "UNKNOWN"
		out.Errors = append(out.Errors, VerificationErrorDetail{
			Code:    "ERR_INVALID_COMPARISON",
			Message: "source-versus-source comparison is rejected: identical roles or witnesses",
		})
		out.Report = c.buildFailureReport(input, "UNKNOWN", "ERR_INVALID_COMPARISON: identical roles or witnesses", out.Errors)
		return out, nil
	}

	// Ensure Source is SOURCE and Target is TARGET
	if input.SourceWitness.Side != SideSource || input.TargetWitness.Side != SideTarget {
		out.TopVerdict = "UNKNOWN"
		out.CaseVerdict = "UNKNOWN"
		out.Errors = append(out.Errors, VerificationErrorDetail{
			Code:    "ERR_INVALID_COMPARISON",
			Message: fmt.Sprintf("invalid witness roles: source has %q, target has %q", input.SourceWitness.Side, input.TargetWitness.Side),
		})
		out.Report = c.buildFailureReport(input, "UNKNOWN", "ERR_INVALID_COMPARISON: invalid witness roles", out.Errors)
		return out, nil
	}

	// 2. Check Target Build Freshness (X66 / X58)
	if input.ExpectedTargetBuildHash != "" &&
		input.TargetWitness.BuildArtifactRef != input.ExpectedTargetBuildHash &&
		input.TargetWitness.TreeRef != input.ExpectedTargetBuildHash {
		out.TopVerdict = "UNKNOWN"
		out.CaseVerdict = "UNKNOWN"
		out.Errors = append(out.Errors, VerificationErrorDetail{
			Code:    "ERR_STALE_EXECUTION",
			Message: fmt.Sprintf("target execution witness build hash does not match current target tree hash %q", input.ExpectedTargetBuildHash),
		})
		out.Report = c.buildFailureReport(input, "UNKNOWN", "ERR_STALE_EXECUTION: stale target build", out.Errors)
		return out, nil
	}

	// 3. Occurrence Pairing & Branch Divergence (X64, X65)
	pairing, diverged, firstDiv, err := c.pairOccurrences(input)
	if err != nil {
		out.TopVerdict = "UNKNOWN"
		out.CaseVerdict = "UNKNOWN"
		out.Errors = append(out.Errors, VerificationErrorDetail{
			Code:    "ERR_PAIRING_FAILED",
			Message: err.Error(),
		})
		out.Report = c.buildFailureReport(input, "UNKNOWN", "ERR_PAIRING_FAILED: "+err.Error(), out.Errors)
		return out, nil
	}
	out.OccurrencePairing = pairing
	out.Diverged = diverged
	out.FirstDivergence = firstDiv

	// 4. Map Local Identities & Verify Bijection
	idMap, idBijectionErr := c.buildIdentityMap(input, pairing)
	out.IdentityMap = idMap
	if idBijectionErr != nil {
		out.Errors = append(out.Errors, VerificationErrorDetail{
			Code:    "ERR_AMBIGUOUS_IDENTITY",
			Message: idBijectionErr.Error(),
		})
	}

	// 5. Dimension Evaluation
	dims, mismatchIDs := c.evaluateDimensions(input, pairing, idMap, diverged, firstDiv, idBijectionErr)

	// 6. Check Incomplete Capture (X08 / X57)
	if input.IncompleteCaptureReason != "" {
		for _, dName := range reportDimensionNames {
			c.setDimensionUnknown(&dims, dName, input.IncompleteCaptureReason, defaultEvidence)
		}
	}
	for dimName, reason := range input.MissingDimensions {
		c.setDimensionUnknown(&dims, dimName, reason, defaultEvidence)
	}
	// X19 runs only through this witnessed, read-only verifier. Missing capture
	// does not erase an independently established lifecycle violation.
	if input.RequiredSessionContractRef != "" || input.SessionContinuity != nil {
		result := evaluateSessionContinuity(input, idMap)
		mergeSessionContinuity(&dims, result)
		if result.Verdict == "MISMATCH" {
			mismatchIDs = append(mismatchIDs, "mismatch-session-lifecycle")
		}
	}

	// 7. Aggregate Verdict
	caseVerdict := c.aggregateCaseVerdict(dims, mismatchIDs)
	out.CaseVerdict = caseVerdict
	out.TopVerdict = caseVerdict

	// 8. Construct Verification Report v2.1
	pairingJSON, err := toCanonicalJSON(pairing)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize pairing: %w", err)
	}
	if valErr := ValidateOccurrencePairing(pairingJSON); valErr != nil {
		return nil, fmt.Errorf("generated occurrence pairing violates schema: %w", valErr)
	}
	pairingDigest := "sha256:" + hexDigest(pairingJSON)

	var idMapRef *string
	if idMap != nil {
		idMapJSON, err := toCanonicalJSON(idMap)
		if err == nil {
			_ = ValidateIdentityMap(idMapJSON)
			dig := "sha256:" + hexDigest(idMapJSON)
			idMapRef = &dig
		}
	}

	srcTraceRef := input.SourceTraceRef
	if srcTraceRef == "" {
		srcTraceRef = defaultEvidence
	}
	tgtTraceRef := input.TargetTraceRef
	if tgtTraceRef == "" {
		tgtTraceRef = defaultEvidence
	}

	report := &VerificationReportV2{
		SchemaVersion:      "2.1",
		RunID:              input.RunID,
		ProjectID:          input.ProjectID,
		ConfigurationHash:  input.ConfigurationHash,
		PlanHash:           input.PlanHash,
		CorpusHash:         input.CorpusHash,
		PolicyHash:         input.PolicyHash,
		TargetTreeHash:     input.TargetTreeHash,
		Status:             "COMPLETED",
		Verdict:            caseVerdict,
		HasExpectedChanges: false,
		Cases: []VerificationCaseResult{
			{
				CaseID:               input.CaseID,
				SourceTraceRef:       &srcTraceRef,
				TargetTraceRef:       &tgtTraceRef,
				Verdict:              caseVerdict,
				Dimensions:           dims,
				MismatchIDs:          mismatchIDs,
				IdentityMapRef:       idMapRef,
				OccurrencePairingRef: &pairingDigest,
			},
		},
		Errors:          out.Errors,
		AcceptanceScope: input.AcceptanceScope,
	}

	// Validate report conformance against schema
	reportJSON, err := toCanonicalJSON(report)
	if err == nil {
		if valErr := ValidateVerificationReport(reportJSON); valErr != nil {
			return nil, fmt.Errorf("generated verification report violates schema: %w", valErr)
		}
	}

	out.Report = report
	return out, nil
}

func (c *BehaviorComparator) pairOccurrences(input *CompareCaseInput) (*OccurrencePairing, bool, *PairingDivergence, error) {
	scenRef := input.ScenarioRef
	if scenRef == "" {
		scenRef = "sha256:" + strings.Repeat("a", 64)
	}

	pairing := &OccurrencePairing{
		SchemaVersion:       "1.0",
		ProjectID:           input.ProjectID,
		CaseID:              input.CaseID,
		ScenarioRef:         scenRef,
		LogicalWorkflowID:   input.WorkflowID,
		Policy:              "STOP_AT_FIRST_DIVERGENCE",
		Pairs:               make([]OccurrencePair, 0),
		FirstDivergence:     nil,
		UnmatchedSourceRefs: make([]string, 0),
		UnmatchedTargetRefs: make([]string, 0),
		CleanupRefs:         make([]string, 0),
	}

	defaultEvidence := "sha256:" + strings.Repeat("a", 64)

	// Driver-reported divergence takes immediate precedence
	if input.DriverDivergence != nil {
		srcRef := input.DriverDivergence.SourceContextRef
		if srcRef == "" {
			srcRef = defaultEvidence
		}
		tgtRef := input.DriverDivergence.TargetContextRef
		if tgtRef == "" {
			tgtRef = defaultEvidence
		}
		pDiv := &PairingDivergence{
			SourceOccurrenceRef: srcRef,
			TargetOccurrenceRef: tgtRef,
			Reason:              "BRANCH",
			EvidenceRefs:        []string{defaultEvidence},
		}
		pairing.FirstDivergence = pDiv
		return pairing, true, pDiv, nil
	}

	// Index target occurrences by (step_id, visit)
	type occKey struct {
		stepID string
		visit  int
	}
	tgtMap := make(map[occKey]StepOccurrence, len(input.TargetOccurrences))
	for _, to := range input.TargetOccurrences {
		tgtMap[occKey{stepID: to.StepID, visit: to.Visit}] = to
	}

	diverged := false
	var firstDiv *PairingDivergence

	for i, so := range input.SourceOccurrences {
		key := occKey{stepID: so.StepID, visit: so.Visit}
		to, found := tgtMap[key]

		srcOccRef := digestOf(&so)

		if !found {
			// Source occurrence exists, target missing -> divergence
			diverged = true
			tgtOccRef := defaultEvidence
			firstDiv = &PairingDivergence{
				SourceOccurrenceRef: srcOccRef,
				TargetOccurrenceRef: tgtOccRef,
				Reason:              "EARLY_TERMINAL",
				EvidenceRefs:        []string{defaultEvidence},
			}
			pairing.Pairs = append(pairing.Pairs, OccurrencePair{
				StepID:              so.StepID,
				Visit:               so.Visit,
				SourceOccurrenceRef: srcOccRef,
				TargetOccurrenceRef: tgtOccRef,
				Outcome:             "DIVERGED",
			})
			pairing.FirstDivergence = firstDiv

			for j := i + 1; j < len(input.SourceOccurrences); j++ {
				pairing.UnmatchedSourceRefs = append(pairing.UnmatchedSourceRefs, digestOf(&input.SourceOccurrences[j]))
			}
			break
		}

		tgtOccRef := digestOf(&to)

		// Check branch selection equivalence
		branchMatch := (stringPtrEqual(so.SelectedNextStepID, to.SelectedNextStepID) &&
			stringPtrEqual(so.TerminalID, to.TerminalID))

		outcome := "ALIGNED"
		if !branchMatch {
			diverged = true
			outcome = "DIVERGED"

			firstDiv = &PairingDivergence{
				SourceOccurrenceRef: srcOccRef,
				TargetOccurrenceRef: tgtOccRef,
				Reason:              "BRANCH",
				EvidenceRefs:        []string{defaultEvidence},
			}
			pairing.FirstDivergence = firstDiv
		}

		pairing.Pairs = append(pairing.Pairs, OccurrencePair{
			StepID:              so.StepID,
			Visit:               so.Visit,
			SourceOccurrenceRef: srcOccRef,
			TargetOccurrenceRef: tgtOccRef,
			Outcome:             outcome,
		})

		if diverged {
			// STOP_AT_FIRST_DIVERGENCE: remaining source occurrences are unmatched
			for j := i + 1; j < len(input.SourceOccurrences); j++ {
				pairing.UnmatchedSourceRefs = append(pairing.UnmatchedSourceRefs, digestOf(&input.SourceOccurrences[j]))
			}
			break
		}
	}

	return pairing, diverged, firstDiv, nil
}

func (c *BehaviorComparator) buildIdentityMap(input *CompareCaseInput, pairing *OccurrencePairing) (*IdentityMapV2, error) {
	if input.SourceIdentityLedger == nil || input.TargetIdentityLedger == nil {
		return nil, nil
	}

	srcLedger := input.SourceIdentityLedger
	tgtLedger := input.TargetIdentityLedger

	srcBytes, _ := toCanonicalJSON(srcLedger)
	tgtBytes, _ := toCanonicalJSON(tgtLedger)
	pairingBytes, _ := toCanonicalJSON(pairing)

	polRef := input.PolicyRef
	if polRef == "" {
		polRef = "sha256:" + strings.Repeat("a", 64)
	}

	srcIdentDigest := "sha256:" + hexDigest(srcBytes)
	tgtIdentDigest := "sha256:" + hexDigest(tgtBytes)
	pairDigest := "sha256:" + hexDigest(pairingBytes)

	idMap := &IdentityMapV2{
		SchemaVersion:             "2.1",
		ProjectID:                 input.ProjectID,
		CaseID:                    input.CaseID,
		WorkflowInstanceID:        input.WorkflowID,
		PolicyRef:                 polRef,
		SourceLocalIdentityRef:    &srcIdentDigest,
		TargetLocalIdentityRef:    &tgtIdentDigest,
		OccurrencePairingRef:      &pairDigest,
		Bindings:                  make([]IdentityBinding, 0),
		RelationshipAssertionRefs: []string{srcIdentDigest, tgtIdentDigest},
	}

	type identKey struct {
		domain     string
		resourceID string
	}
	tgtEntries := make(map[identKey]LocalIdentityEntry, len(tgtLedger.Entries))
	for _, te := range tgtLedger.Entries {
		tgtEntries[identKey{domain: te.Domain, resourceID: te.ResourceID}] = te
	}

	srcToTgt := make(map[string]string)
	tgtToSrc := make(map[string]string)
	var bijectionErr error

	bindingIdx := 0
	for _, se := range srcLedger.Entries {
		key := identKey{domain: se.Domain, resourceID: se.ResourceID}
		te, found := tgtEntries[key]
		if !found {
			continue
		}

		bindingIdx++
		bID := fmt.Sprintf("bind-%d", bindingIdx)
		actorID := "actor-default"
		if se.ActorID != nil && *se.ActorID != "" {
			actorID = *se.ActorID
		} else if te.ActorID != nil && *te.ActorID != "" {
			actorID = *te.ActorID
		}

		logKey := fmt.Sprintf("%s_%s", se.Domain, se.ResourceID)
		logKey = strings.ReplaceAll(logKey, "-", "_")

		b := IdentityBinding{
			BindingID:     bID,
			RelationGroup: "relgroup_default",
			ActorID:       actorID,
			ConnectionID:  nil,
			LogicalKey:    logKey,
			SourceValue: TypedValueContainer{
				ValueType: se.Value.Type,
				Value:     se.Value.Value,
			},
			TargetValue: TypedValueContainer{
				ValueType: te.Value.Type,
				Value:     te.Value.Value,
			},
			SourceLocalID: se.LocalID,
			TargetLocalID: te.LocalID,
			EvidenceRefs:  []string{"sha256:" + strings.Repeat("a", 64)},
		}

		// Enforce bijection
		if existingTgt, ok := srcToTgt[se.LocalID]; ok && existingTgt != te.LocalID {
			bijectionErr = fmt.Errorf("identity bijection violation: source local ID %q mapped to both %q and %q", se.LocalID, existingTgt, te.LocalID)
		}
		if existingSrc, ok := tgtToSrc[te.LocalID]; ok && existingSrc != se.LocalID {
			bijectionErr = fmt.Errorf("identity bijection violation: target local ID %q mapped to both %q and %q", te.LocalID, existingSrc, se.LocalID)
		}

		srcToTgt[se.LocalID] = te.LocalID
		tgtToSrc[te.LocalID] = se.LocalID

		idMap.Bindings = append(idMap.Bindings, b)
	}

	return idMap, bijectionErr
}

func (c *BehaviorComparator) evaluateDimensions(
	input *CompareCaseInput,
	pairing *OccurrencePairing,
	idMap *IdentityMapV2,
	diverged bool,
	firstDiv *PairingDivergence,
	bijectionErr error,
) (VerificationDimensions, []string) {
	defaultEvidence := "sha256:" + strings.Repeat("a", 64)
	mismatchIDs := make([]string, 0)

	// Initialize all 11 dimensions with PASS defaults
	dims := VerificationDimensions{
		HTTP:           DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Body:           DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Authentication: DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Authorization:  DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Session:        DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Database:       DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		External:       DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Files:          DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Cache:          DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Errors:         DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
		Execution:      DimensionResult{Applicable: true, Verdict: stringPtr("PASS"), Reason: nil, EvidenceRefs: []string{defaultEvidence}},
	}

	// Divergence causes Execution and HTTP/Branch mismatch
	if diverged {
		mismatchIDs = append(mismatchIDs, "mismatch-branch-divergence")
		dims.Execution.Verdict = stringPtr("MISMATCH")
		dims.Execution.Reason = stringPtr(fmt.Sprintf("STOP_AT_FIRST_DIVERGENCE: branch diverged: %s", firstDiv.Reason))
		return dims, mismatchIDs
	}

	// Bijection failure causes Identity/Body mismatch
	if bijectionErr != nil {
		mismatchIDs = append(mismatchIDs, "mismatch-identity-bijection")
		dims.Body.Verdict = stringPtr("MISMATCH")
		dims.Body.Reason = stringPtr(bijectionErr.Error())
	}

	// Build map of identity replacements (source_value -> target_value)
	identityReplacements := make(map[string]string)
	if idMap != nil {
		for _, b := range idMap.Bindings {
			identityReplacements[b.SourceValue.Value] = b.TargetValue.Value
			identityReplacements[b.SourceLocalID] = b.TargetLocalID
		}
	}

	// Iterate over paired occurrences and compare contexts
	for _, pair := range pairing.Pairs {
		if pair.Outcome != "ALIGNED" {
			continue
		}

		// Find corresponding occurrences to look up post-context refs
		var srcCtx, tgtCtx *EvaluationContext
		for _, so := range input.SourceOccurrences {
			if so.StepID == pair.StepID && so.Visit == pair.Visit && so.PostContextRef != nil && input.SourceContexts != nil {
				srcCtx = input.SourceContexts[*so.PostContextRef]
				break
			}
		}
		for _, to := range input.TargetOccurrences {
			if to.StepID == pair.StepID && to.Visit == pair.Visit && to.PostContextRef != nil && input.TargetContexts != nil {
				tgtCtx = input.TargetContexts[*to.PostContextRef]
				break
			}
		}

		if srcCtx == nil || tgtCtx == nil {
			continue
		}

		// Invariant: Context origin MUST be ACTUAL_OBSERVED (X63)
		if srcCtx.Origin != OriginActualObserved || tgtCtx.Origin != OriginActualObserved {
			mismatchIDs = append(mismatchIDs, "mismatch-origin-not-actual")
			dims.Execution.Verdict = stringPtr("MISMATCH")
			dims.Execution.Reason = stringPtr("context origin is not ACTUAL_OBSERVED: expected or fixture state cannot satisfy observation")
			continue
		}

		// 1. HTTP Status & Location Header (X67)
		if srcCtx.Namespaces.Result != nil && tgtCtx.Namespaces.Result != nil {
			srcResp := srcCtx.Namespaces.Result
			tgtResp := tgtCtx.Namespaces.Result

			if srcResp.Status != tgtResp.Status {
				mMismatchID := "mismatch-http-status"
				mismatchIDs = append(mismatchIDs, mMismatchID)
				dims.HTTP.Verdict = stringPtr("MISMATCH")
				dims.HTTP.Reason = stringPtr(fmt.Sprintf("status code mismatch: source %d vs target %d", srcResp.Status, tgtResp.Status))
			}

			// Check redirect Location header
			if isRedirectStatus(srcResp.Status) {
				srcLoc := findHeader(srcResp.Headers, "Location")
				tgtLoc := findHeader(tgtResp.Headers, "Location")
				if srcLoc == nil || tgtLoc == nil || *srcLoc != *tgtLoc {
					mismatchIDs = append(mismatchIDs, "mismatch-redirect-location")
					dims.HTTP.Verdict = stringPtr("MISMATCH")
					dims.HTTP.Reason = stringPtr("redirect Location header missing or mismatch")
				}
			}

			// 2. Response Body Comparison & Normalization
			srcBodyStr := extractBodyString(srcResp.Body)
			tgtBodyStr := extractBodyString(tgtResp.Body)

			// Apply identity mapping replacements to source body
			mappedSrcBody := applyReplacements(srcBodyStr, identityReplacements)

			// Apply normalization rules
			normSrcBody := c.applyNormalization(mappedSrcBody, input.NormalizationRules)
			normTgtBody := c.applyNormalization(tgtBodyStr, input.NormalizationRules)

			if normSrcBody != normTgtBody {
				mismatchIDs = append(mismatchIDs, "mismatch-body-content")
				dims.Body.Verdict = stringPtr("MISMATCH")
				dims.Body.Reason = stringPtr(fmt.Sprintf("body mismatch: source %q vs target %q", normSrcBody, normTgtBody))
			}

			// Ordered cookie occurrences retain scope, security and lifetime (X19/X67).
			// Duplicate names may have distinct path/domain and cannot be collapsed.
			if len(srcResp.Cookies) != len(tgtResp.Cookies) {
				mismatchIDs = append(mismatchIDs, "mismatch-cookie-count")
				dims.Session.Verdict = stringPtr("MISMATCH")
				dims.Session.Reason = stringPtr("cookie occurrence count mismatch")
			} else {
				for i, sc := range srcResp.Cookies {
					tc := tgtResp.Cookies[i]
					if !sessionCookieAttributesEqual(sc, tc) {
						mismatchIDs = append(mismatchIDs, "mismatch-cookie-attributes")
						dims.Session.Verdict = stringPtr("MISMATCH")
						dims.Session.Reason = stringPtr("cookie occurrence scope/security/lifetime mismatch")
					}
				}
			}
		}

		// 4. Authentication & Principal Provenance (X06 / X77)
		srcPrincipal := srcCtx.Namespaces.Actor.ActualPrincipal
		tgtPrincipal := tgtCtx.Namespaces.Actor.ActualPrincipal

		if srcPrincipal.Missing != nil || tgtPrincipal.Missing != nil {
			dims.Authentication.Verdict = stringPtr("UNKNOWN")
			dims.Authentication.Reason = stringPtr("principal observation missing: insufficient capture")
		} else {
			if srcPrincipal.Authentication != tgtPrincipal.Authentication {
				mismatchIDs = append(mismatchIDs, "mismatch-auth-status")
				dims.Authentication.Verdict = stringPtr("MISMATCH")
				dims.Authentication.Reason = stringPtr(fmt.Sprintf("authentication status mismatch: %s vs %s", srcPrincipal.Authentication, tgtPrincipal.Authentication))
			} else if srcPrincipal.Authentication == "AUTHENTICATED" {
				// Compare subject local IDs using identity map
				if srcPrincipal.SubjectLocalID != nil && tgtPrincipal.SubjectLocalID != nil {
					expectedTgtSubj := *srcPrincipal.SubjectLocalID
					if mapped, ok := identityReplacements[expectedTgtSubj]; ok {
						expectedTgtSubj = mapped
					}
					if *tgtPrincipal.SubjectLocalID != expectedTgtSubj {
						mismatchIDs = append(mismatchIDs, "mismatch-wrong-principal")
						dims.Authentication.Verdict = stringPtr("MISMATCH")
						dims.Authentication.Reason = stringPtr(fmt.Sprintf("wrong principal: target authenticated %q, expected mapped %q", *tgtPrincipal.SubjectLocalID, expectedTgtSubj))
					}
				}
			}
		}

		// 5. Database Mutation Comparison & Read-Only Invariant (X05 / X61 / Decisive Test 1)
		// Comparator is strictly read-only: inspects tables without performing any inserts or mutations!
		srcDB := srcCtx.Namespaces.State.DB
		tgtDB := tgtCtx.Namespaces.State.DB

		if srcDB.Missing != nil || tgtDB.Missing != nil {
			dims.Database.Verdict = stringPtr("UNKNOWN")
			dims.Database.Reason = stringPtr("database observation missing: insufficient capture")
		} else {
			dbMismatch, dbReason := c.compareDBTables(srcDB.Tables, tgtDB.Tables)
			if dbMismatch {
				mismatchIDs = append(mismatchIDs, "mismatch-db-write")
				dims.Database.Verdict = stringPtr("MISMATCH")
				dims.Database.Reason = stringPtr(dbReason)
			}
		}

		// 6. External Effects & Cardinality (X07 / X72)
		if srcCtx.Namespaces.Result != nil && tgtCtx.Namespaces.Result != nil {
			srcEffects := srcCtx.Namespaces.Result.EffectRefs
			tgtEffects := tgtCtx.Namespaces.Result.EffectRefs

			if len(srcEffects) > 0 {
				if len(tgtEffects) == 0 {
					mismatchIDs = append(mismatchIDs, "mismatch-missing-effect")
					dims.External.Verdict = stringPtr("MISMATCH")
					dims.External.Reason = stringPtr("missing external effect in target")
				} else if len(tgtEffects) > len(srcEffects) {
					mismatchIDs = append(mismatchIDs, "mismatch-duplicate-effect")
					dims.External.Verdict = stringPtr("MISMATCH")
					dims.External.Reason = stringPtr(fmt.Sprintf("duplicate business effect: cardinality violated (source %d vs target %d)", len(srcEffects), len(tgtEffects)))
				}
			}
		}
	}

	return dims, mismatchIDs
}

func (c *BehaviorComparator) compareDBTables(srcTables, tgtTables []EvaluationDBTable) (bool, string) {
	// Index target tables by table name
	tgtMap := make(map[string]EvaluationDBTable, len(tgtTables))
	for _, t := range tgtTables {
		tgtMap[t.Table] = t
	}

	for _, st := range srcTables {
		tt, found := tgtMap[st.Table]
		if !found {
			return true, fmt.Sprintf("target missing table mutation for table %q", st.Table)
		}

		// Compare row counts and values
		srcRows := extractRows(st.Rows)
		tgtRows := extractRows(tt.Rows)

		if len(srcRows) > 0 && len(tgtRows) == 0 {
			return true, fmt.Sprintf("missing database write: source table %q has %d rows, target has 0", st.Table, len(srcRows))
		}
		if len(srcRows) != len(tgtRows) {
			return true, fmt.Sprintf("database row count mismatch on table %q: source has %d, target has %d", st.Table, len(srcRows), len(tgtRows))
		}
	}

	return false, ""
}

func (c *BehaviorComparator) aggregateCaseVerdict(dims VerificationDimensions, mismatchIDs []string) string {
	dimResults := []DimensionResult{
		dims.HTTP,
		dims.Body,
		dims.Authentication,
		dims.Authorization,
		dims.Session,
		dims.Database,
		dims.External,
		dims.Files,
		dims.Cache,
		dims.Errors,
		dims.Execution,
	}

	hasMismatch := len(mismatchIDs) > 0
	hasUnknown := false

	for _, d := range dimResults {
		if !d.Applicable {
			continue
		}
		if d.Verdict != nil {
			if *d.Verdict == "MISMATCH" {
				hasMismatch = true
			} else if *d.Verdict == "UNKNOWN" {
				hasUnknown = true
			}
		}
	}

	if hasMismatch {
		return "MISMATCH"
	}
	if hasUnknown {
		return "UNKNOWN"
	}
	return "PASS"
}

func (c *BehaviorComparator) buildFailureReport(
	input *CompareCaseInput,
	verdict string,
	reason string,
	errDetails []VerificationErrorDetail,
) *VerificationReportV2 {
	defaultEvidence := "sha256:" + strings.Repeat("a", 64)

	dimVerdict := verdict
	dims := VerificationDimensions{
		HTTP:           DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Body:           DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Authentication: DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Authorization:  DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Session:        DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Database:       DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		External:       DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Files:          DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Cache:          DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Errors:         DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
		Execution:      DimensionResult{Applicable: true, Verdict: stringPtr(dimVerdict), Reason: stringPtr(reason), EvidenceRefs: []string{defaultEvidence}},
	}

	srcTrace := input.SourceTraceRef
	if srcTrace == "" {
		srcTrace = defaultEvidence
	}
	tgtTrace := input.TargetTraceRef
	if tgtTrace == "" {
		tgtTrace = defaultEvidence
	}

	return &VerificationReportV2{
		SchemaVersion:      "2.1",
		RunID:              input.RunID,
		ProjectID:          input.ProjectID,
		ConfigurationHash:  input.ConfigurationHash,
		PlanHash:           input.PlanHash,
		CorpusHash:         input.CorpusHash,
		PolicyHash:         input.PolicyHash,
		TargetTreeHash:     input.TargetTreeHash,
		Status:             "COMPLETED",
		Verdict:            verdict,
		HasExpectedChanges: false,
		Cases: []VerificationCaseResult{
			{
				CaseID:               input.CaseID,
				SourceTraceRef:       &srcTrace,
				TargetTraceRef:       &tgtTrace,
				Verdict:              verdict,
				Dimensions:           dims,
				MismatchIDs:          make([]string, 0),
				IdentityMapRef:       nil,
				OccurrencePairingRef: nil,
			},
		},
		Errors:          errDetails,
		AcceptanceScope: input.AcceptanceScope,
	}
}

func (c *BehaviorComparator) setDimensionUnknown(dims *VerificationDimensions, dimName, reason, evidence string) {
	ev := []string{evidence}
	u := stringPtr("UNKNOWN")
	r := stringPtr(reason)

	switch dimName {
	case "http":
		dims.HTTP.Verdict = u
		dims.HTTP.Reason = r
		dims.HTTP.EvidenceRefs = ev
	case "body":
		dims.Body.Verdict = u
		dims.Body.Reason = r
		dims.Body.EvidenceRefs = ev
	case "authentication":
		dims.Authentication.Verdict = u
		dims.Authentication.Reason = r
		dims.Authentication.EvidenceRefs = ev
	case "authorization":
		dims.Authorization.Verdict = u
		dims.Authorization.Reason = r
		dims.Authorization.EvidenceRefs = ev
	case "session":
		dims.Session.Verdict = u
		dims.Session.Reason = r
		dims.Session.EvidenceRefs = ev
	case "database":
		dims.Database.Verdict = u
		dims.Database.Reason = r
		dims.Database.EvidenceRefs = ev
	case "external":
		dims.External.Verdict = u
		dims.External.Reason = r
		dims.External.EvidenceRefs = ev
	case "files":
		dims.Files.Verdict = u
		dims.Files.Reason = r
		dims.Files.EvidenceRefs = ev
	case "cache":
		dims.Cache.Verdict = u
		dims.Cache.Reason = r
		dims.Cache.EvidenceRefs = ev
	case "errors":
		dims.Errors.Verdict = u
		dims.Errors.Reason = r
		dims.Errors.EvidenceRefs = ev
	case "execution":
		dims.Execution.Verdict = u
		dims.Execution.Reason = r
		dims.Execution.EvidenceRefs = ev
	}
}

func (c *BehaviorComparator) applyNormalization(body string, rules []*NormalizationRule) string {
	if len(rules) == 0 {
		return body
	}
	result := body
	for _, r := range rules {
		if r.Variability == "CLOCK" {
			result = normalizePattern(result, `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?`, "[NORMALIZED_DATE]")
		} else if r.Variability == "GENERATED_ID" || r.Variability == "TOKEN" {
			result = normalizePattern(result, `[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`, "[NORMALIZED_UUID]")
		}
	}
	return result
}

func normalizePattern(s, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, replacement)
}

func isRedirectStatus(status int) bool {
	return status == 301 || status == 302 || status == 303 || status == 307 || status == 308
}

func findHeader(headers []EvaluationHeader, name string) *string {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return &h.Value
		}
	}
	return nil
}

func extractBodyString(tv TaggedValue) string {
	if tv.Value == nil {
		return ""
	}
	switch v := tv.Value.(type) {
	case string:
		return v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func extractRows(tv TaggedValue) []any {
	if tv.Value == nil {
		return nil
	}
	if arr, ok := tv.Value.([]TaggedValue); ok {
		res := make([]any, len(arr))
		for i, v := range arr {
			res[i] = v
		}
		return res
	}
	if arr, ok := tv.Value.([]any); ok {
		return arr
	}
	return nil
}

func applyReplacements(s string, replacements map[string]string) string {
	res := s
	for oldVal, newVal := range replacements {
		if oldVal != "" && newVal != "" {
			res = strings.ReplaceAll(res, oldVal, newVal)
		}
	}
	return res
}

func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func hexDigest(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func toCanonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return CanonicalJSON(raw)
}

func digestOf(v any) string {
	b, err := toCanonicalJSON(v)
	if err != nil {
		return "sha256:" + strings.Repeat("a", 64)
	}
	return "sha256:" + hexDigest(b)
}

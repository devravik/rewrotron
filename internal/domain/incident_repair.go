package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// Incident and bounded repair error codes.
const (
	ErrIncidentBudgetExceeded    = "ERR_INCIDENT_BUDGET_EXCEEDED"
	ErrTrivialFingerprintReset   = "ERR_TRIVIAL_FINGERPRINT_RESET"
	ErrStaleProjectPublication   = "ERR_STALE_PROJECT_PUBLICATION"
	ErrStaleIncidentPublication  = "ERR_STALE_INCIDENT_PUBLICATION"
	ErrStaleLeasePublication     = "ERR_STALE_LEASE_PUBLICATION"
	ErrIdempotencyReceiptFound   = "ERR_IDEMPOTENCY_RECEIPT_FOUND"
	ErrInvalidIdempotencyKey     = "ERR_INVALID_IDEMPOTENCY_KEY"
	ErrIncidentNotFound          = "ERR_INCIDENT_NOT_FOUND"
	ErrInvalidIncidentState      = "ERR_INVALID_INCIDENT_STATE"
)

// MaxAttemptsPerIncident is the authoritative maximum retry budget per incident subject (X24).
const MaxAttemptsPerIncident = 3

var (
	idempotencyKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)
	incidentSubjectRegex = regexp.MustCompile(`^[a-zA-Z0-9_:.-]{1,256}$`)
)

// IncidentError describes an incident repair or budget violation.
type IncidentError struct {
	Code   string
	Reason string
	Cause  error
}

// Error formats the incident repair failure.
func (e *IncidentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Reason, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

// Unwrap returns the underlying error cause.
func (e *IncidentError) Unwrap() error { return e.Cause }

// IncidentSubjectRecord represents the authoritative state and budget for an incident subject.
type IncidentSubjectRecord struct {
	SubjectID             string          `json:"subject_id"`               // e.g., "auth_session_cookie_missing"
	ProjectID             string          `json:"project_id"`
	CurrentVersion        int64           `json:"current_version"`          // Project/incident state generation
	AttemptCount          int             `json:"attempt_count"`           // Number of attempts consumed (max 3)
	KnownFingerprints     map[string]bool `json:"known_fingerprints"`      // Tracked failed attempt fingerprints
	LastKnownContextHash  string          `json:"last_known_context_hash"`  // Canonical context hash
	IsBlocked             bool            `json:"is_blocked"`              // True if budget exhausted (3 attempts)
	ReviewBlockerReason   string          `json:"review_blocker_reason"`   // Reason code requiring review
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// IncidentAttemptRecord represents a specific repair attempt.
type IncidentAttemptRecord struct {
	AttemptID       string         `json:"attempt_id"`
	SubjectID       string         `json:"subject_id"`
	ProjectID       string         `json:"project_id"`
	ProjectVersion  int64          `json:"project_version"`
	FencingToken    int64          `json:"fencing_token"`
	AttemptNumber   int            `json:"attempt_number"` // 1, 2, or 3
	Fingerprint     string         `json:"fingerprint"`
	Context         map[string]any `json:"context"`
	Hypothesis      string         `json:"hypothesis"`
	LeaseOwner      string         `json:"lease_owner"`
	LeaseExpiresAt  time.Time      `json:"lease_expires_at"`
}

// IdempotencyReceiptRecord represents the durable committed result of an execution/attempt.
type IdempotencyReceiptRecord struct {
	IdempotencyKey string         `json:"idempotency_key"`
	SubjectID      string         `json:"subject_id"`
	ProjectID      string         `json:"project_id"`
	AttemptNumber  int            `json:"attempt_number"`
	FencingToken   int64          `json:"fencing_token"`
	CommittedAt    time.Time      `json:"committed_at"`
	ResponseResult map[string]any `json:"response_result"`
}

// IncidentRepairService manages incident tracking, bounded repair budgets,
// idempotency receipts, trivial context/patch rejection, and stale publication rejection.
type IncidentRepairService struct {
	mu           sync.RWMutex
	incidents    map[string]*IncidentSubjectRecord    // key: project_id + ":" + subject_id
	receipts     map[string]*IdempotencyReceiptRecord // key: project_id + ":" + idempotency_key
	projectVers  map[string]int64                     // key: project_id -> current project version
}

// NewIncidentRepairService creates an initialized IncidentRepairService.
func NewIncidentRepairService() *IncidentRepairService {
	return &IncidentRepairService{
		incidents:   make(map[string]*IncidentSubjectRecord),
		receipts:    make(map[string]*IdempotencyReceiptRecord),
		projectVers: make(map[string]int64),
	}
}

// SetProjectVersion sets or advances the authoritative version of a project.
func (s *IncidentRepairService) SetProjectVersion(projectID string, version int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projectVers[projectID] = version
}

// GetProjectVersion returns the current project version.
func (s *IncidentRepairService) GetProjectVersion(projectID string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projectVers[projectID]
}

// RegisterIncident initializes or retrieves an incident subject record.
func (s *IncidentRepairService) RegisterIncident(projectID, subjectID string, now time.Time) (*IncidentSubjectRecord, error) {
	if !incidentSubjectRegex.MatchString(subjectID) {
		return nil, &IncidentError{Code: "ERR_INVALID_ARGUMENT", Reason: fmt.Sprintf("invalid subject_id %q", subjectID)}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	key := projectID + ":" + subjectID
	inc, exists := s.incidents[key]
	if !exists {
		currentVer := s.projectVers[projectID]
		if currentVer == 0 {
			currentVer = 1
			s.projectVers[projectID] = 1
		}
		inc = &IncidentSubjectRecord{
			SubjectID:         subjectID,
			ProjectID:         projectID,
			CurrentVersion:    currentVer,
			AttemptCount:      0,
			KnownFingerprints: make(map[string]bool),
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		s.incidents[key] = inc
	}
	cloned := *inc
	return &cloned, nil
}

// ComputeNormalizedContextHash computes a deterministic SHA256 digest of context fields,
// excluding trivial non-semantic keys (such as comments, whitespace annotations, or timestamps).
func ComputeNormalizedContextHash(context map[string]any) string {
	if context == nil {
		return sha256Hex([]byte("{}"))
	}
	// Strip known trivial noise keys
	cleaned := make(map[string]any)
	for k, v := range context {
		switch k {
		case "comment", "notes", "timestamp", "description", "_comment", "nonce":
			// Trivial non-semantic context fields do not change the substantive hypothesis
			continue
		default:
			cleaned[k] = v
		}
	}
	raw, err := json.Marshal(cleaned)
	if err != nil {
		return sha256Hex([]byte(fmt.Sprintf("%v", cleaned)))
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return sha256Hex(raw)
	}
	return sha256Hex(canon)
}

// CheckIdempotencyReceipt checks if an idempotency key was already committed (Test X24).
// Invariant: "Crash after commit before response returns existing result".
func (s *IncidentRepairService) CheckIdempotencyReceipt(projectID, idempotencyKey string) (*IdempotencyReceiptRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := projectID + ":" + idempotencyKey
	receipt, exists := s.receipts[key]
	if !exists {
		return nil, false
	}
	cloned := *receipt
	return &cloned, true
}

// BeginAttempt verifies that a repair attempt is authorized under the three-attempt budget (X24):
// 1. Checks if idempotency key has existing receipt; if so returns it immediately.
// 2. Enforces max 3 attempts per incident subject.
// 3. Rejects trivial context / patch edits (same substantive context hash or known failed fingerprint without new hypothesis).
// 4. Verifies project version currency.
func (s *IncidentRepairService) BeginAttempt(
	projectID string,
	subjectID string,
	idempotencyKey string,
	attemptFingerprint string,
	hypothesis string,
	context map[string]any,
	fencingToken int64,
	owner string,
	now time.Time,
	leaseDuration time.Duration,
) (*IncidentAttemptRecord, *IdempotencyReceiptRecord, error) {
	if !idempotencyKeyRegex.MatchString(idempotencyKey) {
		return nil, nil, &IncidentError{
			Code:   ErrInvalidIdempotencyKey,
			Reason: fmt.Sprintf("invalid idempotency key format: %q", idempotencyKey),
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check idempotency receipt (Crash after commit before response returns existing result)
	receiptKey := projectID + ":" + idempotencyKey
	if receipt, exists := s.receipts[receiptKey]; exists {
		clonedReceipt := *receipt
		return nil, &clonedReceipt, nil
	}

	// 2. Fetch incident record
	incidentKey := projectID + ":" + subjectID
	inc, exists := s.incidents[incidentKey]
	if !exists {
		currentVer := s.projectVers[projectID]
		if currentVer == 0 {
			currentVer = 1
			s.projectVers[projectID] = 1
		}
		inc = &IncidentSubjectRecord{
			SubjectID:         subjectID,
			ProjectID:         projectID,
			CurrentVersion:    currentVer,
			AttemptCount:      0,
			KnownFingerprints: make(map[string]bool),
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		s.incidents[incidentKey] = inc
	}

	// 3. Enforce 3-attempt budget per incident subject (Test X24)
	if inc.AttemptCount >= MaxAttemptsPerIncident || inc.IsBlocked {
		inc.IsBlocked = true
		inc.ReviewBlockerReason = "INCIDENT_BUDGET_EXHAUSTED"
		return nil, nil, &IncidentError{
			Code:   ErrIncidentBudgetExceeded,
			Reason: fmt.Sprintf("incident subject %q in project %q has exhausted its 3-attempt repair budget (consumed %d attempts); further attempts require explicit reviewed incident/cause change (Test X24)", subjectID, projectID, inc.AttemptCount),
		}
	}

	// 4. Enforce trivial context edit rejection (Test X24):
	// "trivial context edit cannot reset three-attempt budget"
	// "trivial patch/context edits cannot reset the budget. Known failed fingerprints require a new hypothesis/context or a concrete review blocker."
	contextHash := ComputeNormalizedContextHash(context)
	if inc.KnownFingerprints[attemptFingerprint] {
		return nil, nil, &IncidentError{
			Code:   ErrTrivialFingerprintReset,
			Reason: fmt.Sprintf("attempt fingerprint %q is an already-known failed fingerprint for incident subject %q; trivial edits cannot reset budget (Test X24)", attemptFingerprint, subjectID),
		}
	}

	if inc.AttemptCount > 0 && inc.LastKnownContextHash != "" && inc.LastKnownContextHash == contextHash && hypothesis == "" {
		return nil, nil, &IncidentError{
			Code:   ErrTrivialFingerprintReset,
			Reason: fmt.Sprintf("context edit for incident %q is substantively identical to prior failed attempt without a new hypothesis; trivial edits cannot reset or consume budget (Test X24)", subjectID),
		}
	}

	// Consume one attempt from the budget
	inc.AttemptCount++
	inc.LastKnownContextHash = contextHash
	inc.KnownFingerprints[attemptFingerprint] = true
	inc.UpdatedAt = now

	attempt := &IncidentAttemptRecord{
		AttemptID:      fmt.Sprintf("att_%s_%d", subjectID, inc.AttemptCount),
		SubjectID:      subjectID,
		ProjectID:      projectID,
		ProjectVersion: inc.CurrentVersion,
		FencingToken:   fencingToken,
		AttemptNumber:  inc.AttemptCount,
		Fingerprint:    attemptFingerprint,
		Context:        context,
		Hypothesis:     hypothesis,
		LeaseOwner:     owner,
		LeaseExpiresAt: now.Add(leaseDuration),
	}

	return attempt, nil, nil
}

// RecordFailure records a failed repair attempt outcome, preserving the consumed budget.
func (s *IncidentRepairService) RecordFailure(
	projectID string,
	subjectID string,
	attemptID string,
	fingerprint string,
	reason string,
	now time.Time,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	incidentKey := projectID + ":" + subjectID
	inc, exists := s.incidents[incidentKey]
	if !exists {
		return &IncidentError{Code: ErrIncidentNotFound, Reason: "incident not found"}
	}

	inc.KnownFingerprints[fingerprint] = true
	inc.UpdatedAt = now
	if inc.AttemptCount >= MaxAttemptsPerIncident {
		inc.IsBlocked = true
		inc.ReviewBlockerReason = "INCIDENT_BUDGET_EXHAUSTED"
	}

	return nil
}

// CommitResultAndStoreReceipt commits the successful repair outcome and durably stores the idempotency receipt (Test X24):
// 1. Verifies that the project version hasn't drifted (stale project publication rejection).
// 2. Verifies that the incident version hasn't drifted (stale incident publication rejection).
// 3. Verifies that the lease hasn't expired (stale lease publication rejection).
// 4. Stores the idempotency receipt with the response payload.
func (s *IncidentRepairService) CommitResultAndStoreReceipt(
	attempt *IncidentAttemptRecord,
	idempotencyKey string,
	expectedProjectVersion int64,
	currentFencingToken int64,
	responseResult map[string]any,
	now time.Time,
) (*IdempotencyReceiptRecord, error) {
	if attempt == nil {
		return nil, &IncidentError{Code: "ERR_INVALID_ARGUMENT", Reason: "attempt record is nil"}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Stale project publication check (Test X24)
	actualProjectVer := s.projectVers[attempt.ProjectID]
	if actualProjectVer > expectedProjectVersion {
		return nil, &IncidentError{
			Code:   ErrStaleProjectPublication,
			Reason: fmt.Sprintf("stale project publication: project %q is at version %d, publication expected version %d (Test X24)", attempt.ProjectID, actualProjectVer, expectedProjectVersion),
		}
	}

	// 2. Stale incident publication check (Test X24)
	incidentKey := attempt.ProjectID + ":" + attempt.SubjectID
	inc, exists := s.incidents[incidentKey]
	if !exists {
		return nil, &IncidentError{Code: ErrIncidentNotFound, Reason: "incident not found"}
	}
	if inc.CurrentVersion > attempt.ProjectVersion {
		return nil, &IncidentError{
			Code:   ErrStaleIncidentPublication,
			Reason: fmt.Sprintf("stale incident publication: incident %q is at version %d, publication attempt was at version %d (Test X24)", attempt.SubjectID, inc.CurrentVersion, attempt.ProjectVersion),
		}
	}

	// 3. Stale lease publication check (Test X24)
	if now.After(attempt.LeaseExpiresAt) {
		return nil, &IncidentError{
			Code:   ErrStaleLeasePublication,
			Reason: fmt.Sprintf("stale lease publication: lease for attempt %q expired at %s, current time %s (Test X24)", attempt.AttemptID, attempt.LeaseExpiresAt.Format(time.RFC3339), now.Format(time.RFC3339)),
		}
	}

	// Fencing token check
	if attempt.FencingToken < currentFencingToken {
		return nil, &IncidentError{
			Code:   ErrStaleFenceToken,
			Reason: fmt.Sprintf("stale fence token: attempt has token %d, current project token is %d (Test X24)", attempt.FencingToken, currentFencingToken),
		}
	}

	// 4. Durably store idempotency receipt
	receiptKey := attempt.ProjectID + ":" + idempotencyKey
	receipt := &IdempotencyReceiptRecord{
		IdempotencyKey: idempotencyKey,
		SubjectID:      attempt.SubjectID,
		ProjectID:      attempt.ProjectID,
		AttemptNumber:  attempt.AttemptNumber,
		FencingToken:   attempt.FencingToken,
		CommittedAt:    now,
		ResponseResult: responseResult,
	}
	s.receipts[receiptKey] = receipt

	inc.UpdatedAt = now
	cloned := *receipt
	return &cloned, nil
}

// ResetBudgetWithReviewedCause allows an explicit, reviewed incident cause change to reset the budget.
func (s *IncidentRepairService) ResetBudgetWithReviewedCause(
	projectID string,
	subjectID string,
	newCauseReviewComment string,
	now time.Time,
) error {
	if newCauseReviewComment == "" {
		return &IncidentError{
			Code:   "ERR_INVALID_REVIEW",
			Reason: "cannot reset budget without an explicit reviewed cause/hypothesis comment",
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	incidentKey := projectID + ":" + subjectID
	inc, exists := s.incidents[incidentKey]
	if !exists {
		return &IncidentError{Code: ErrIncidentNotFound, Reason: "incident not found"}
	}

	inc.AttemptCount = 0
	inc.IsBlocked = false
	inc.ReviewBlockerReason = ""
	inc.KnownFingerprints = make(map[string]bool)
	inc.LastKnownContextHash = ""
	inc.CurrentVersion++
	inc.UpdatedAt = now

	return nil
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

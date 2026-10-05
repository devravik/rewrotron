package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ResourceFencingError describes a fencing or resource ownership error.
type ResourceFencingError struct {
	Code   string
	Reason string
	Cause  error
}

// Error formats the resource fencing failure.
func (e *ResourceFencingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Reason, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

// Unwrap returns the underlying cause.
func (e *ResourceFencingError) Unwrap() error { return e.Cause }

// Common error codes for resource ownership and fencing.
const (
	ErrDuplicateActiveWork  = "ERR_DUPLICATE_ACTIVE_WORK"
	ErrResourceLocked       = "ERR_RESOURCE_LOCKED"
	ErrStaleFenceToken      = "ERR_STALE_FENCE_TOKEN"
	ErrContentHashMismatch  = "ERR_CONTENT_HASH_MISMATCH"
	ErrMissionNotFound      = "ERR_MISSION_NOT_FOUND"
	ErrInvalidMissionState  = "ERR_INVALID_MISSION_STATE"
	ErrLeaseExpired         = "ERR_LEASE_EXPIRED"
	ErrUnauthorizedOwner    = "ERR_UNAUTHORIZED_OWNER"
)

// MissionStatus defines the state of a mission.
type MissionStatus string

const (
	MissionStatusQueued    MissionStatus = "QUEUED"
	MissionStatusLeased    MissionStatus = "LEASED"
	MissionStatusRunning   MissionStatus = "RUNNING"
	MissionStatusCompleted MissionStatus = "COMPLETED"
	MissionStatusBlocked   MissionStatus = "BLOCKED"
	MissionStatusFailed    MissionStatus = "FAILED"
	MissionStatusCancelled MissionStatus = "CANCELLED"
)

// IsActive returns true if the mission status represents active work.
func (s MissionStatus) IsActive() bool {
	switch s {
	case MissionStatusQueued, MissionStatusLeased, MissionStatusRunning, MissionStatusBlocked:
		return true
	default:
		return false
	}
}

// MissionRecord represents an scheduled work item.
type MissionRecord struct {
	MissionID      string         `json:"mission_id"`
	ProjectID      string         `json:"project_id"`
	Type           string         `json:"type"`
	Priority       int            `json:"priority"`
	Status         MissionStatus  `json:"status"`
	UnitID         *string        `json:"unit_id"`
	CaseID         *string        `json:"case_id"`
	Description    string         `json:"description"`
	Context        map[string]any `json:"context"`
	Fingerprint    string         `json:"fingerprint"`
	AttemptCount   int            `json:"attempt_count"`
	FencingToken   int64          `json:"fencing_token"`
	LeaseOwner     string         `json:"lease_owner"`
	LeaseExpiresAt *time.Time     `json:"lease_expires_at"`
}

// ResourceClaimRecord represents a durable lease over a project resource/file.
type ResourceClaimRecord struct {
	ProjectID    string    `json:"project_id"`
	ResourceKey  string    `json:"resource_key"`
	MissionID    string    `json:"mission_id"`
	LeaseOwner   string    `json:"lease_owner"`
	FencingToken int64     `json:"fencing_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ResourceFencingService provides in-memory thread-safe mission scheduling,
// lease fencing, resource ownership locks, and publication verification.
type ResourceFencingService struct {
	mu             sync.RWMutex
	missions       map[string]*MissionRecord       // key: project_id + ":" + mission_id
	resourceClaims map[string]*ResourceClaimRecord // key: project_id + ":" + resource_key
	// activeFingerprints maps project_id + ":" + type + ":" + fingerprint -> mission_id
	activeFingerprints map[string]string
}

// NewResourceFencingService creates an initialized ResourceFencingService.
func NewResourceFencingService() *ResourceFencingService {
	return &ResourceFencingService{
		missions:           make(map[string]*MissionRecord),
		resourceClaims:     make(map[string]*ResourceClaimRecord),
		activeFingerprints: make(map[string]string),
	}
}

// SubmitMission enqueues a mission, enforcing active work uniqueness (Test X22).
// Invariant: Two agents cannot submit duplicate active work with the same project, type, and fingerprint.
func (s *ResourceFencingService) SubmitMission(mission *MissionRecord) (*MissionRecord, error) {
	if mission == nil {
		return nil, &ResourceFencingError{Code: "ERR_INVALID_ARGUMENT", Reason: "mission is nil"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	activeKey := mission.ProjectID + ":" + mission.Type + ":" + mission.Fingerprint
	if existingID, exists := s.activeFingerprints[activeKey]; exists {
		// Check if the existing mission is still active
		existingMission := s.missions[mission.ProjectID+":"+existingID]
		if existingMission != nil && existingMission.Status.IsActive() {
			return nil, &ResourceFencingError{
				Code:   ErrDuplicateActiveWork,
				Reason: fmt.Sprintf("mission with project %s, type %s, fingerprint %s already active as mission %s (Test X22)", mission.ProjectID, mission.Type, mission.Fingerprint, existingID),
			}
		}
	}

	cloned := *mission
	if cloned.Status == "" {
		cloned.Status = MissionStatusQueued
	}
	clonedKey := cloned.ProjectID + ":" + cloned.MissionID
	s.missions[clonedKey] = &cloned

	if cloned.Status.IsActive() {
		s.activeFingerprints[activeKey] = cloned.MissionID
	}

	return &cloned, nil
}

// ClaimMission claims a queued or expired mission, incrementing the fencing token and attempt count.
func (s *ResourceFencingService) ClaimMission(
	projectID string,
	missionID string,
	owner string,
	now time.Time,
	leaseDuration time.Duration,
) (*MissionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := projectID + ":" + missionID
	mission, exists := s.missions[key]
	if !exists {
		return nil, &ResourceFencingError{
			Code:   ErrMissionNotFound,
			Reason: fmt.Sprintf("mission %s not found in project %s", missionID, projectID),
		}
	}

	// Can claim if QUEUED or if previous lease expired
	if mission.Status == MissionStatusLeased || mission.Status == MissionStatusRunning {
		if mission.LeaseExpiresAt != nil && now.Before(*mission.LeaseExpiresAt) {
			if mission.LeaseOwner != owner {
				return nil, &ResourceFencingError{
					Code:   ErrResourceLocked,
					Reason: fmt.Sprintf("mission %s is already claimed by %s until %s (Test X22)", missionID, mission.LeaseOwner, mission.LeaseExpiresAt.Format(time.RFC3339)),
				}
			}
		}
	} else if mission.Status != MissionStatusQueued {
		return nil, &ResourceFencingError{
			Code:   ErrInvalidMissionState,
			Reason: fmt.Sprintf("mission %s is in non-claimable state %s", missionID, mission.Status),
		}
	}

	// Increment fencing token and attempt count
	mission.FencingToken++
	mission.AttemptCount++
	mission.Status = MissionStatusLeased
	mission.LeaseOwner = owner
	expires := now.Add(leaseDuration)
	mission.LeaseExpiresAt = &expires

	activeKey := mission.ProjectID + ":" + mission.Type + ":" + mission.Fingerprint
	s.activeFingerprints[activeKey] = mission.MissionID

	ret := *mission
	return &ret, nil
}

// AcquireResourceClaim locks a project resource/file under a mission lease (Test X22).
// Invariant: Two agents editing a shared file/resource cannot both hold active claims simultaneously.
func (s *ResourceFencingService) AcquireResourceClaim(
	projectID string,
	resourceKey string,
	missionID string,
	owner string,
	fencingToken int64,
	now time.Time,
	leaseDuration time.Duration,
) (*ResourceClaimRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	claimKey := projectID + ":" + resourceKey
	existing, exists := s.resourceClaims[claimKey]
	if exists {
		// If held by another owner and not expired, refuse lock
		if existing.LeaseOwner != owner && now.Before(existing.ExpiresAt) {
			return nil, &ResourceFencingError{
				Code:   ErrResourceLocked,
				Reason: fmt.Sprintf("resource %s in project %s is locked by owner %s until %s (Test X22)", resourceKey, projectID, existing.LeaseOwner, existing.ExpiresAt.Format(time.RFC3339)),
			}
		}
	}

	claim := &ResourceClaimRecord{
		ProjectID:    projectID,
		ResourceKey:  resourceKey,
		MissionID:    missionID,
		LeaseOwner:   owner,
		FencingToken: fencingToken,
		ExpiresAt:    now.Add(leaseDuration),
	}
	s.resourceClaims[claimKey] = claim

	ret := *claim
	return &ret, nil
}

// RenewMissionLease extends a mission lease and its associated resource claims.
func (s *ResourceFencingService) RenewMissionLease(
	projectID string,
	missionID string,
	owner string,
	fencingToken int64,
	now time.Time,
	leaseDuration time.Duration,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := projectID + ":" + missionID
	mission, exists := s.missions[key]
	if !exists {
		return &ResourceFencingError{Code: ErrMissionNotFound, Reason: "mission not found"}
	}
	if mission.LeaseOwner != owner {
		return &ResourceFencingError{
			Code:   ErrUnauthorizedOwner,
			Reason: fmt.Sprintf("cannot renew: owner mismatch (expected %s, got %s)", mission.LeaseOwner, owner),
		}
	}
	if mission.FencingToken != fencingToken {
		return &ResourceFencingError{
			Code:   ErrStaleFenceToken,
			Reason: fmt.Sprintf("cannot renew: stale fencing token (expected %d, got %d)", mission.FencingToken, fencingToken),
		}
	}
	if mission.LeaseExpiresAt != nil && now.After(*mission.LeaseExpiresAt) {
		return &ResourceFencingError{
			Code:   ErrLeaseExpired,
			Reason: "cannot renew: lease has already expired",
		}
	}

	newExpiry := now.Add(leaseDuration)
	mission.LeaseExpiresAt = &newExpiry

	// Extend any matching resource claims
	for _, claim := range s.resourceClaims {
		if claim.ProjectID == projectID && claim.MissionID == missionID && claim.LeaseOwner == owner {
			claim.ExpiresAt = newExpiry
		}
	}

	return nil
}

// ValidatePublication verifies that a patch can be safely published into the target environment.
// Enforces Test X23:
// 1. Fencing token check: old owner paused after fence check, expires, resumes; publication rejected
//    if fencing token is superseded or lease is expired.
// 2. Before-hash check: current target files must match before_sha256 to prevent lost edits under serial baseline.
func (s *ResourceFencingService) ValidatePublication(
	patch *PatchRecord,
	currentTargetFiles map[string][]byte,
	now time.Time,
) error {
	if patch == nil {
		return &ResourceFencingError{Code: "ERR_INVALID_ARGUMENT", Reason: "patch is nil"}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Verify fencing token and lease active status
	var activeMission *MissionRecord
	for _, m := range s.missions {
		if m.ProjectID == patch.ProjectID && m.FencingToken == patch.FencingToken {
			activeMission = m
			break
		}
	}

	if activeMission != nil {
		// Check if lease expired
		if activeMission.LeaseExpiresAt != nil && now.After(*activeMission.LeaseExpiresAt) {
			return &ResourceFencingError{
				Code:   ErrStaleFenceToken,
				Reason: fmt.Sprintf("fenced publication failed: mission %s lease expired at %s, current time %s; expired process cannot publish (Test X23)", activeMission.MissionID, activeMission.LeaseExpiresAt.Format(time.RFC3339), now.Format(time.RFC3339)),
			}
		}
	}

	// Check if any resource claims for this patch's files are superseded by a higher fencing token
	for _, file := range patch.Files {
		claimKey := patch.ProjectID + ":" + file.Path
		if claim, exists := s.resourceClaims[claimKey]; exists {
			if claim.FencingToken > patch.FencingToken {
				return &ResourceFencingError{
					Code:   ErrStaleFenceToken,
					Reason: fmt.Sprintf("fenced publication failed: resource %s has been transferred/superseded by fencing token %d, current patch has stale token %d (Test X23)", file.Path, claim.FencingToken, patch.FencingToken),
				}
			}
		}
	}

	// 2. Verify before-hashes against current target files (serial baseline lost-edit prevention)
	for _, file := range patch.Files {
		currentBytes, fileExists := currentTargetFiles[file.Path]

		switch file.Action {
		case PatchActionCreate:
			if fileExists {
				return &ResourceFencingError{
					Code:   ErrContentHashMismatch,
					Reason: fmt.Sprintf("fenced publication failed: file %s already exists for CREATE action (Test X23)", file.Path),
				}
			}
		case PatchActionModify, PatchActionDelete:
			if !fileExists {
				return &ResourceFencingError{
					Code:   ErrContentHashMismatch,
					Reason: fmt.Sprintf("fenced publication failed: file %s does not exist for %s action (Test X23)", file.Path, file.Action),
				}
			}
			h := sha256.Sum256(currentBytes)
			currentHex := hex.EncodeToString(h[:])
			if file.BeforeSHA256 == nil || *file.BeforeSHA256 != currentHex {
				expected := "<nil>"
				if file.BeforeSHA256 != nil {
					expected = *file.BeforeSHA256
				}
				return &ResourceFencingError{
					Code:   ErrContentHashMismatch,
					Reason: fmt.Sprintf("fenced publication failed: file %s current hash %s does not match patch before_sha256 %s; lost edit prevented (Test X23)", file.Path, currentHex, expected),
				}
			}
		}
	}

	return nil
}

// ApplyPatch atomically validates fencing and content hashes, then applies patch changes to the target files map.
func (s *ResourceFencingService) ApplyPatch(
	patch *PatchRecord,
	currentTargetFiles map[string][]byte,
	newFileContents map[string][]byte,
	now time.Time,
) (map[string][]byte, error) {
	if err := s.ValidatePublication(patch, currentTargetFiles, now); err != nil {
		return nil, err
	}

	// Create updated copy of target files
	updated := make(map[string][]byte, len(currentTargetFiles)+len(patch.Files))
	for k, v := range currentTargetFiles {
		updated[k] = v
	}

	for _, file := range patch.Files {
		switch file.Action {
		case PatchActionCreate, PatchActionModify:
			content, exists := newFileContents[file.Path]
			if !exists {
				return nil, &ResourceFencingError{
					Code:   "ERR_MISSING_CONTENT",
					Reason: fmt.Sprintf("new content for file %s not provided in payload", file.Path),
				}
			}
			// Verify content matches after_sha256
			h := sha256.Sum256(content)
			contentHex := hex.EncodeToString(h[:])
			if file.AfterSHA256 == nil || *file.AfterSHA256 != contentHex {
				return nil, &ResourceFencingError{
					Code:   ErrContentHashMismatch,
					Reason: fmt.Sprintf("file %s new content hash %s does not match after_sha256 %v", file.Path, contentHex, file.AfterSHA256),
				}
			}
			updated[file.Path] = content
		case PatchActionDelete:
			delete(updated, file.Path)
		}
	}

	return updated, nil
}

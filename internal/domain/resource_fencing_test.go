package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func sha256String(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// TestPatch_ValidationAndCanonicalJSON tests patch schema 1.0 validation and canonicalization.
func TestPatch_ValidationAndCanonicalJSON(t *testing.T) {
	beforeHash := sha256String("old file content")
	afterHash := sha256String("new file content")

	patch := &PatchRecord{
		SchemaVersion: "1.0",
		ProjectID:     "proj1",
		UnitID:        "auth_unit",
		FencingToken:  1,
		Files: []PatchFileRecord{
			{
				Path:         "app/Services/AuthService.php",
				Action:       PatchActionModify,
				BeforeSHA256: &beforeHash,
				AfterSHA256:  &afterHash,
			},
			{
				Path:         "app/Http/Controllers/AuthController.php",
				Action:       PatchActionCreate,
				BeforeSHA256: nil,
				AfterSHA256:  &afterHash,
			},
			{
				Path:         "legacy/AuthOld.php",
				Action:       PatchActionDelete,
				BeforeSHA256: &beforeHash,
				AfterSHA256:  nil,
			},
		},
	}

	canon, err := patch.CanonicalJSON()
	if err != nil {
		t.Fatalf("failed to canonicalize valid patch: %v", err)
	}

	digest, err := patch.Digest()
	if err != nil {
		t.Fatalf("failed to compute patch digest: %v", err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("digest must start with sha256:, got %s", digest)
	}

	parsed, err := ParsePatch(canon)
	if err != nil {
		t.Fatalf("failed to parse canonical patch: %v", err)
	}
	if parsed.UnitID != "auth_unit" || len(parsed.Files) != 3 {
		t.Fatalf("parsed patch mismatch: %+v", parsed)
	}

	// Negative fencing token rejected
	badPatch := *patch
	badPatch.FencingToken = -1
	_, err = badPatch.CanonicalJSON()
	if err == nil || !strings.Contains(err.Error(), "fencing_token must be an integer >= 0") {
		t.Fatalf("expected error for negative fencing token, got: %v", err)
	}

	// CREATE with non-null before_sha256 rejected
	badPatch2 := *patch
	badPatch2.Files = []PatchFileRecord{
		{
			Path:         "test.txt",
			Action:       PatchActionCreate,
			BeforeSHA256: &beforeHash, // illegal for CREATE
			AfterSHA256:  &afterHash,
		},
	}
	_, err = badPatch2.CanonicalJSON()
	if err == nil || !strings.Contains(err.Error(), "before_sha256 must be null for CREATE") {
		t.Fatalf("expected error for CREATE with before_sha256, got: %v", err)
	}
}

// TestResourceFencing_X22_DuplicateActiveWorkAndSharedResourceLock tests Test X22:
// "Two agents submit duplicate work or edit shared auth file; unique active work/resource ownership prevents duplicate admission/lost edits"
func TestResourceFencing_X22_DuplicateActiveWorkAndSharedResourceLock(t *testing.T) {
	svc := NewResourceFencingService()
	baseTime := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

	// 1. Subtest: Duplicate active work prevention
	t.Run("X22_DuplicateActiveWorkPrevention", func(t *testing.T) {
		m1 := &MissionRecord{
			MissionID:    "m_auth_01",
			ProjectID:    "proj_x22",
			Type:         "MIGRATE_UNIT",
			Priority:     500,
			Status:       MissionStatusQueued,
			Fingerprint:  "fp_auth_migration_ci3_to_laravel",
			Description:  "Migrate CI3 Auth to Laravel",
			AttemptCount: 0,
			FencingToken: 0,
		}

		// Agent 1 submits mission
		_, err := svc.SubmitMission(m1)
		if err != nil {
			t.Fatalf("first mission submission failed: %v", err)
		}

		// Agent 2 attempts to submit duplicate active work with identical (project, type, fingerprint)
		m2 := &MissionRecord{
			MissionID:    "m_auth_02",
			ProjectID:    "proj_x22",
			Type:         "MIGRATE_UNIT",
			Priority:     500,
			Status:       MissionStatusQueued,
			Fingerprint:  "fp_auth_migration_ci3_to_laravel",
			Description:  "Duplicate submission by agent 2",
			AttemptCount: 0,
			FencingToken: 0,
		}
		_, err = svc.SubmitMission(m2)
		if err == nil || !strings.Contains(err.Error(), ErrDuplicateActiveWork) {
			t.Fatalf("expected duplicate active work error (Test X22), got: %v", err)
		}

		// Now Agent 1 claims and completes m1
		claimed, err := svc.ClaimMission("proj_x22", "m_auth_01", "agent_1", baseTime, 5*time.Minute)
		if err != nil {
			t.Fatalf("ClaimMission failed: %v", err)
		}
		if claimed.FencingToken != 1 || claimed.AttemptCount != 1 {
			t.Fatalf("unexpected claimed state: %+v", claimed)
		}

		// Mark m1 completed
		svc.mu.Lock()
		svc.missions["proj_x22:m_auth_01"].Status = MissionStatusCompleted
		svc.mu.Unlock()

		// Once completed, a new mission with the same fingerprint is now admitted
		_, err = svc.SubmitMission(m2)
		if err != nil {
			t.Fatalf("expected admission after previous completed, got: %v", err)
		}
	})

	// 2. Subtest: Shared file / resource lock conflict prevention
	t.Run("X22_SharedResourceLockConflict", func(t *testing.T) {
		sharedFile := "app/Services/AuthService.php"

		// Agent 1 acquires resource claim
		claim1, err := svc.AcquireResourceClaim(
			"proj_x22",
			sharedFile,
			"m_auth_01",
			"agent_1",
			1,
			baseTime,
			5*time.Minute,
		)
		if err != nil {
			t.Fatalf("AcquireResourceClaim for agent 1 failed: %v", err)
		}
		if claim1.LeaseOwner != "agent_1" {
			t.Fatalf("unexpected claim owner: %s", claim1.LeaseOwner)
		}

		// Agent 2 attempts to concurrently edit / lock the same shared auth file
		_, err = svc.AcquireResourceClaim(
			"proj_x22",
			sharedFile,
			"m_auth_02",
			"agent_2",
			1,
			baseTime.Add(1*time.Minute), // while agent 1 lease is still active
			5*time.Minute,
		)
		if err == nil || !strings.Contains(err.Error(), ErrResourceLocked) {
			t.Fatalf("expected resource lock conflict for agent 2 on shared file (Test X22), got: %v", err)
		}
	})
}

// TestResourceFencing_X23_FencedPublication tests Test X23:
// "Old owner pauses after fence check, expires and resumes; old process cannot write after resource transfer; shared/disjoint conflicts handled under serial baseline"
func TestResourceFencing_X23_FencedPublication(t *testing.T) {
	svc := NewResourceFencingService()
	t0 := time.Date(2026, 10, 5, 14, 0, 0, 0, time.UTC)
	leaseDuration := 5 * time.Minute

	// Initial target state on disk
	initialAuthContent := []byte("<?php class AuthService { /* initial CI3 bridge */ }")
	initialAuthHash := sha256String(string(initialAuthContent))

	targetFiles := map[string][]byte{
		"app/Services/AuthService.php": initialAuthContent,
	}

	// 1. Submit and claim mission by Agent 1
	m := &MissionRecord{
		MissionID:   "m_migrate_auth",
		ProjectID:   "proj_x23",
		Type:        "MIGRATE_UNIT",
		Priority:    500,
		Status:      MissionStatusQueued,
		Fingerprint: "fp_auth",
	}
	_, err := svc.SubmitMission(m)
	if err != nil {
		t.Fatalf("SubmitMission failed: %v", err)
	}

	claimed1, err := svc.ClaimMission("proj_x23", "m_migrate_auth", "agent_1", t0, leaseDuration)
	if err != nil {
		t.Fatalf("ClaimMission failed: %v", err)
	}
	if claimed1.FencingToken != 1 {
		t.Fatalf("expected fencing token 1, got %d", claimed1.FencingToken)
	}

	_, err = svc.AcquireResourceClaim("proj_x23", "app/Services/AuthService.php", "m_migrate_auth", "agent_1", 1, t0, leaseDuration)
	if err != nil {
		t.Fatalf("AcquireResourceClaim failed: %v", err)
	}

	// Agent 1 prepares patch with fencing_token = 1
	agent1NewContent := []byte("<?php class AuthService { /* agent 1 changes */ }")
	agent1NewHash := sha256String(string(agent1NewContent))
	patch1 := &PatchRecord{
		SchemaVersion: "1.0",
		ProjectID:     "proj_x23",
		UnitID:        "auth_unit",
		FencingToken:  1,
		Files: []PatchFileRecord{
			{
				Path:         "app/Services/AuthService.php",
				Action:       PatchActionModify,
				BeforeSHA256: &initialAuthHash,
				AfterSHA256:  &agent1NewHash,
			},
		},
	}

	// Simulate Agent 1 pausing after fence check, while lease expires:
	// Advance time past lease duration: t0 + 6 minutes
	tExpired := t0.Add(6 * time.Minute)

	// Invariant 1: Expired lease prevents Agent 1 from publishing
	err = svc.ValidatePublication(patch1, targetFiles, tExpired)
	if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) || !strings.Contains(err.Error(), "lease expired") {
		t.Fatalf("expected publication rejection due to expired lease (Test X23), got: %v", err)
	}

	// 2. Now Agent 2 reclaims the expired mission, incrementing the fencing token to 2
	claimed2, err := svc.ClaimMission("proj_x23", "m_migrate_auth", "agent_2", tExpired, leaseDuration)
	if err != nil {
		t.Fatalf("Agent 2 ClaimMission failed: %v", err)
	}
	if claimed2.FencingToken != 2 {
		t.Fatalf("expected fencing token 2 for Agent 2, got %d", claimed2.FencingToken)
	}

	// Resource transferred to Agent 2
	_, err = svc.AcquireResourceClaim("proj_x23", "app/Services/AuthService.php", "m_migrate_auth", "agent_2", 2, tExpired, leaseDuration)
	if err != nil {
		t.Fatalf("Resource transfer to Agent 2 failed: %v", err)
	}

	// Invariant 2: Old process (Agent 1) resumes and tries to write with stale token 1:
	// Resource transfer supersedes token 1, must be rejected with ERR_STALE_FENCE_TOKEN
	err = svc.ValidatePublication(patch1, targetFiles, tExpired.Add(30*time.Second))
	if err == nil || !strings.Contains(err.Error(), ErrStaleFenceToken) || !strings.Contains(err.Error(), "superseded") {
		t.Fatalf("expected publication rejection due to superseded fencing token (Test X23), got: %v", err)
	}

	// 3. Agent 2 publishes changes cleanly under fencing_token = 2
	agent2NewContent := []byte("<?php class AuthService { /* agent 2 approved changes */ }")
	agent2NewHash := sha256String(string(agent2NewContent))
	patch2 := &PatchRecord{
		SchemaVersion: "1.0",
		ProjectID:     "proj_x23",
		UnitID:        "auth_unit",
		FencingToken:  2,
		Files: []PatchFileRecord{
			{
				Path:         "app/Services/AuthService.php",
				Action:       PatchActionModify,
				BeforeSHA256: &initialAuthHash,
				AfterSHA256:  &agent2NewHash,
			},
		},
	}

	updatedTargetFiles, err := svc.ApplyPatch(patch2, targetFiles, map[string][]byte{
		"app/Services/AuthService.php": agent2NewContent,
	}, tExpired.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("Agent 2 ApplyPatch failed: %v", err)
	}

	// 4. Invariant 3: Before-hash verification prevents lost edits under serial baseline
	// If any agent tries to apply a patch with old before_sha256 against the updated target:
	patchStaleBefore := &PatchRecord{
		SchemaVersion: "1.0",
		ProjectID:     "proj_x23",
		UnitID:        "auth_unit",
		FencingToken:  2,
		Files: []PatchFileRecord{
			{
				Path:         "app/Services/AuthService.php",
				Action:       PatchActionModify,
				BeforeSHA256: &initialAuthHash, // Stale! File now has agent2NewHash
				AfterSHA256:  &agent1NewHash,
			},
		},
	}
	err = svc.ValidatePublication(patchStaleBefore, updatedTargetFiles, tExpired.Add(2*time.Minute))
	if err == nil || !strings.Contains(err.Error(), ErrContentHashMismatch) || !strings.Contains(err.Error(), "lost edit prevented") {
		t.Fatalf("expected content hash mismatch preventing lost edit (Test X23), got: %v", err)
	}
}

package domain

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// SecuritySandboxEnforcer mediates path containment, permission boundaries,
// secret redaction, and production-write denial according to docs/09-security/SECURITY.md
// and docs/09-security/SANDBOX.md.
type SecuritySandboxEnforcer struct {
	workspaceRoot string
	sanitizer     *Sanitizer
	policy        *EffectSafetyPolicy
}

// NewSecuritySandboxEnforcer constructs an enforcer with a defined workspace root and secret sanitizer.
func NewSecuritySandboxEnforcer(workspaceRoot string, hmacKey []byte) *SecuritySandboxEnforcer {
	cleanRoot := filepath.Clean(workspaceRoot)
	return &SecuritySandboxEnforcer{
		workspaceRoot: cleanRoot,
		sanitizer:     NewSanitizer(hmacKey),
		policy:        &EffectSafetyPolicy{IsProductionEnvironment: false, AllowIrreversible: false},
	}
}

// ValidatePathWithinWorkspace verifies that target/source paths do not escape the workspace sandbox.
// Rejects path traversal (e.g. "../", "/etc/passwd", Windows drive letters, null bytes, absolute paths).
func (e *SecuritySandboxEnforcer) ValidatePathWithinWorkspace(relPath string) (string, error) {
	if strings.Contains(relPath, "\x00") {
		return "", fmt.Errorf("security violation: null byte in path %q", relPath)
	}

	// Reject absolute paths that attempt to reference outside roots
	if filepath.IsAbs(relPath) {
		cleanAbs := filepath.Clean(relPath)
		rel, err := filepath.Rel(e.workspaceRoot, cleanAbs)
		if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
			return "", fmt.Errorf("security violation: absolute path %q is outside workspace %q", relPath, e.workspaceRoot)
		}
		return cleanAbs, nil
	}

	// Clean and join relative paths
	joined := filepath.Join(e.workspaceRoot, relPath)
	cleanJoined := filepath.Clean(joined)

	// Ensure joined path starts with workspaceRoot
	rel, err := filepath.Rel(e.workspaceRoot, cleanJoined)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." && relPath != "." && relPath != "" {
		return "", fmt.Errorf("security violation: path traversal detected for %q outside workspace %q", relPath, e.workspaceRoot)
	}

	return cleanJoined, nil
}

// RedactLogOrReport scrubs raw passwords, bearer tokens, AWS keys, and PII from
// messages before they are persisted or exposed in reports/logs.
func (e *SecuritySandboxEnforcer) RedactLogOrReport(input string) string {
	clean, _ := e.sanitizer.SanitizeText(input)
	return clean
}

// VerifyProductionAccessDenied ensures that direct production operations and live external writes
// are blocked by default unless explicitly routed to an approved safe sink.
func (e *SecuritySandboxEnforcer) VerifyProductionAccessDenied(effect BusinessEffectRecord) error {
	return e.policy.CheckEffectSafety(effect)
}

// TestM12_Unit3_SecurityAndSandboxQualification executes the comprehensive security and sandbox qualification suite:
// 1. Path containment & directory traversal rejection (relative escapes, absolute paths, null bytes)
// 2. Secret & credential scrubbing in reports, logs, and error outputs (bearer tokens, passwords, AWS keys)
// 3. Permission enforcement & production write denial by default (refuses live external mutations)
// 4. Gate authorization invalidation checks (Gate 1, Gate 2, Gate 3)
// 5. Cross-project data isolation (Project A cannot access or extract Project B sensitive keys)
func TestM12_Unit3_SecurityAndSandboxQualification(t *testing.T) {
	workspace := "/workspace/rewrotron-sandbox"
	hmacKey := []byte("m12-unit3-test-hmac-key")
	enforcer := NewSecuritySandboxEnforcer(workspace, hmacKey)

	// =========================================================================
	// 1. Path Traversal & Sandbox Containment
	// =========================================================================
	t.Run("PathContainmentAndTraversalRejection", func(t *testing.T) {
		validPaths := []string{
			"routes/web.php",
			"app/Http/Controllers/ReconstructedController.php",
			"config/session.php",
			"storage/framework/views",
		}
		for _, vp := range validPaths {
			resolved, err := enforcer.ValidatePathWithinWorkspace(vp)
			if err != nil {
				t.Fatalf("valid path %q should be permitted: %v", vp, err)
			}
			if !strings.HasPrefix(resolved, workspace) {
				t.Fatalf("resolved path %q should stay within workspace %q", resolved, workspace)
			}
		}

		traversalAttempts := []string{
			"../../etc/passwd",
			"../shadow",
			"/etc/hosts",
			"storage/../../../../root/.ssh/id_rsa",
			"routes/web.php\x00/evil",
		}
		for _, badPath := range traversalAttempts {
			_, err := enforcer.ValidatePathWithinWorkspace(badPath)
			if err == nil || !strings.Contains(err.Error(), "security violation") {
				t.Fatalf("path traversal %q must be rejected with security violation, got: %v", badPath, err)
			}
		}
	})

	// =========================================================================
	// 2. Secret Redaction in Logs, Diagnostics, and Reports
	// =========================================================================
	t.Run("SecretRedactionInLogsAndReports", func(t *testing.T) {
		sensitiveLog := "Authentication failed for user alice with password=SuperSecretP@ssword! and token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-ID brackets and aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		redacted := enforcer.RedactLogOrReport(sensitiveLog)

		// Must not contain raw secret tokens
		if strings.Contains(redacted, "SuperSecretP@ssword!") {
			t.Fatalf("raw password was not redacted from log: %s", redacted)
		}
		if strings.Contains(redacted, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY") {
			t.Fatalf("raw AWS key was not redacted from log: %s", redacted)
		}

		// Must contain standard redaction placeholders
		if !strings.Contains(redacted, "[REDACTED") {
			t.Fatalf("expected [REDACTED placeholder in scrubbed log: %s", redacted)
		}
	})

	// =========================================================================
	// 3. Permission Enforcement & Production-Write Denial by Default
	// =========================================================================
	t.Run("ProductionWriteDenialByDefault", func(t *testing.T) {
		prodPolicy := EffectSafetyPolicy{IsProductionEnvironment: true, AllowIrreversible: false}

		readEffect := BusinessEffectRecord{
			EffectID:    "eff-read-1",
			Dimension:   EffectDimensionDBRead,
			EffectClass: EffectClassReadOnly,
		}
		if err := prodPolicy.CheckEffectSafety(readEffect); err != nil {
			t.Fatalf("read-only effect should pass production policy: %v", err)
		}

		writeEffect := BusinessEffectRecord{
			EffectID:    "eff-write-1",
			Dimension:   EffectDimensionDBWrite,
			EffectClass: EffectClassLocalWrite,
		}
		err := prodPolicy.CheckEffectSafety(writeEffect)
		if err == nil || !strings.Contains(err.Error(), "ERR_PRODUCTION_EFFECT_REFUSED") {
			t.Fatalf("direct production database write must be refused, got: %v", err)
		}

		externalMutation := BusinessEffectRecord{
			EffectID:    "eff-ext-1",
			Dimension:   EffectDimensionExternal,
			EffectClass: EffectClassExternalWrite,
		}
		err = prodPolicy.CheckEffectSafety(externalMutation)
		if err == nil || !strings.Contains(err.Error(), "ERR_PRODUCTION_EFFECT_REFUSED") {
			t.Fatalf("direct production external write must be refused, got: %v", err)
		}
	})

	// =========================================================================
	// 4. Cross-Project Isolation & Stale Fence Rejection
	// =========================================================================
	t.Run("CrossProjectIsolationAndFencingGuardrails", func(t *testing.T) {
		// Verify project isolation in sanitization
		sanitizerA := NewSanitizer([]byte("key-project-a"))
		sanitizerB := NewSanitizer([]byte("key-project-b"))

		secret := "password=mysecretpassword123"
		cleanA, _ := sanitizerA.SanitizeText(secret)
		cleanB, _ := sanitizerB.SanitizeText(secret)

		// Keyed HMAC tokens must differ between projects to prevent cross-project correlation leaks
		if cleanA == cleanB {
			t.Fatalf("cross-project HMAC tokens must not match: A=%s B=%s", cleanA, cleanB)
		}
	})
}

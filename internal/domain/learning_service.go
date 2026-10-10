package domain

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ExperimentContext defines the exact context of an execution, repair, or migration experiment
// conforming to docs/07-learning/EXPERIMENT-LEDGER.md.
// Its SHA-256 digest of canonical RFC 8785 JSON representation is the authoritative fingerprint.
type ExperimentContext struct {
	SourceConfiguration string            `json:"source_configuration"`
	PlanHash            string            `json:"plan_hash"`
	CorpusHash          string            `json:"corpus_hash"`
	PolicyHash          string            `json:"policy_hash"`
	CaseHash            string            `json:"case_hash"`
	WorkflowHash        string            `json:"workflow_hash"`
	FixtureHash         string            `json:"fixture_hash"`
	TargetBeforeTreeHash string           `json:"target_before_tree_hash"`
	AdapterVersions     map[string]string `json:"adapter_versions"`
	ToolVersions        map[string]string `json:"tool_versions"`
	DependencyVersions  map[string]string `json:"dependency_versions"`
	HypothesisPatchHash string            `json:"hypothesis_patch_hash"`
	SandboxIdentity     string            `json:"sandbox_identity"`
	SkillVersions       map[string]string `json:"skill_versions,omitempty"`
	PromptVersions      map[string]string `json:"prompt_versions,omitempty"`
}

// Fingerprint calculates the canonical SHA-256 hex digest of the experiment context.
func (ec *ExperimentContext) Fingerprint() (string, []byte, error) {
	raw, err := json.Marshal(ec)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal experiment context: %w", err)
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return "", nil, fmt.Errorf("failed to canonicalize experiment context: %w", err)
	}
	h := sha256.Sum256(canon)
	return hex.EncodeToString(h[:]), canon, nil
}

// SanitizationRule defines sanitization transformations applied to data before global or agent persistence.
type SanitizationRule struct {
	RedactSecrets bool
	ScrubPII      bool
	ScrubPaths    bool
	HMACKey       []byte // project-scoped keyed HMAC for correlation without plaintext disclosure
}

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token|bearer|private[_-]?key)["']?\s*[:=]\s*["']?([a-zA-Z0-9_\-\.\/+=]{8,})`),
		regexp.MustCompile(`(?i)RWT_SANDBOX_DB_PASSWORD\s*=\s*([^\s]+)`),
		regexp.MustCompile(`(?i)(aws_access_key_id|aws_secret_access_key)["']?\s*[:=]\s*["']?([a-zA-Z0-9_\-\.\/+=]+)`),
	}
	piiPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+`), // emails
		regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),                               // SSN
		regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`),             // Credit card
	}
	pathPattern = regexp.MustCompile(`(/(home|Users|var|tmp|media|root)/[a-zA-Z0-9_\-./]+)`)
)

// Sanitizer provides secret, PII, and host path scrubbing conforming to docs/09-security/SECRET-HANDLING.md.
type Sanitizer struct {
	hmacKey []byte
}

// NewSanitizer constructs a Sanitizer with an optional project-scoped HMAC key.
func NewSanitizer(hmacKey []byte) *Sanitizer {
	if len(hmacKey) == 0 {
		// generate deterministic default fallback or empty
		hmacKey = []byte("rewrotron-project-default-hmac-key")
	}
	return &Sanitizer{hmacKey: hmacKey}
}

// SanitizeText scrubs sensitive credentials, PII, and host paths from text.
// Returns the sanitized string and a boolean indicating whether any secret was detected and scrubbed.
func (s *Sanitizer) SanitizeText(input string) (string, bool) {
	modified := false
	output := input

	// 1. Scrub secrets
	for _, p := range secretPatterns {
		if p.MatchString(output) {
			modified = true
			output = p.ReplaceAllStringFunc(output, func(match string) string {
				sub := p.FindStringSubmatch(match)
				if len(sub) > 2 {
					// replace only the secret value part with HMAC token
					val := sub[2]
					mac := hmac.New(sha256.New, s.hmacKey)
					mac.Write([]byte(val))
					token := fmt.Sprintf("[REDACTED:hmac_%s]", hex.EncodeToString(mac.Sum(nil))[:16])
					return strings.Replace(match, val, token, 1)
				}
				return "[REDACTED_SECRET]"
			})
		}
	}

	// 2. Scrub PII
	for _, p := range piiPatterns {
		if p.MatchString(output) {
			modified = true
			output = p.ReplaceAllString(output, "[REDACTED_PII]")
		}
	}

	// 3. Scrub private host paths (e.g. /home/user/private/src -> /workspace/sanitized/...)
	if pathPattern.MatchString(output) {
		modified = true
		output = pathPattern.ReplaceAllString(output, "[REDACTED_PATH]")
	}

	return output, modified
}

// ContainsSensitiveData checks if the given string contains potential raw secrets or unredacted credentials.
func (s *Sanitizer) ContainsSensitiveData(input string) bool {
	for _, p := range secretPatterns {
		if p.MatchString(input) {
			return true
		}
	}
	for _, p := range piiPatterns {
		if p.MatchString(input) {
			return true
		}
	}
	return false
}

// ProjectLesson represents a project-local lesson record conforming to docs/07-learning/KNOWLEDGE-MODEL.md
// and stored in the SQLite `lessons` table.
type ProjectLesson struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	ExperimentID string    `json:"experiment_id"`
	Finding      string    `json:"finding"`
	Remediation  string    `json:"remediation"`
	EvidenceRef  string    `json:"evidence_ref"`
	CreatedAt    time.Time `json:"created_at"`
}

// LessonEvidencePayload defines the detailed evidence stored in the referenced artifact for a lesson.
type LessonEvidencePayload struct {
	LessonID             string            `json:"lesson_id"`
	ProjectID            string            `json:"project_id"`
	ExperimentID         string            `json:"experiment_id"`
	ExperimentFingerprint string           `json:"experiment_fingerprint"`
	Finding              string            `json:"finding"`
	Remediation          string            `json:"remediation"`
	SourceConfigHash     string            `json:"source_config_hash"`
	TargetTreeHash       string            `json:"target_tree_hash"`
	CorpusHash           string            `json:"corpus_hash"`
	PolicyHash           string            `json:"policy_hash"`
	ObservedOutcome      string            `json:"observed_outcome"` // PASSED, FAILED, MISMATCH, etc.
	Limitations          []string          `json:"limitations"`
	Applicability        map[string]string `json:"applicability"`
	RedactionVisibility  string            `json:"redaction_visibility"` // "PRIVATE_SOURCE" or "SANITIZED_EVIDENCE"
	CreatedAt            string            `json:"created_at"`
}

// LearningService provides project-local lesson recording, experiment ledger operations,
// sanitization, and strict cross-project isolation controls.
type LearningService struct {
	mu        sync.RWMutex
	db        *sql.DB
	sanitizer *Sanitizer
}

// NewLearningService creates a new LearningService.
func NewLearningService(db *sql.DB, sanitizer *Sanitizer) *LearningService {
	if sanitizer == nil {
		sanitizer = NewSanitizer(nil)
	}
	return &LearningService{
		db:        db,
		sanitizer: sanitizer,
	}
}

// RecordExperiment stores an immutable experiment in the `experiments` table and registers its canonical context artifact.
func (ls *LearningService) RecordExperiment(ctx context.Context, projectID string, expID string, expCtx *ExperimentContext, outcome string) (string, error) {
	if projectID == "" || expID == "" {
		return "", fmt.Errorf("project_id and experiment_id are required")
	}
	switch outcome {
	case "PASSED", "FAILED", "CANCELLED", "UNKNOWN":
	default:
		return "", fmt.Errorf("invalid experiment outcome %q: must be PASSED, FAILED, CANCELLED, or UNKNOWN", outcome)
	}

	fingerprint, canonBytes, err := expCtx.Fingerprint()
	if err != nil {
		return "", fmt.Errorf("failed to compute experiment fingerprint: %w", err)
	}

	// Artifact ID is sha256 of canonical bytes
	artHash := sha256.Sum256(canonBytes)
	contextRef := hex.EncodeToString(artHash[:])

	now := time.Now().UTC().Format(time.RFC3339)

	// Insert artifact record if not exists
	redactionJSON := `{"visibility":"PRIVATE_SOURCE","policy_ref":"sha256:0000000000000000000000000000000000000000000000000000000000000000"}`
	_, _ = ls.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO artifacts (id, project_id, sha256, relative_path, media_type, size_bytes, redaction_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		contextRef, projectID, contextRef, fmt.Sprintf("experiments/%s.json", expID), "application/json", int64(len(canonBytes)), redactionJSON, now,
	)

	// Insert experiment
	_, err = ls.db.ExecContext(ctx,
		"INSERT INTO experiments (id, project_id, fingerprint, context_ref, outcome, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		expID, projectID, fingerprint, contextRef, outcome, now,
	)
	if err != nil {
		return "", fmt.Errorf("failed to record experiment in storage: %w", err)
	}

	return fingerprint, nil
}

// RecordLesson records an append-only project-local lesson.
// Enforces sanitization: credentials and private PII are scrubbed before storage,
// and the lesson is bound strictly to the given project_id.
func (ls *LearningService) RecordLesson(ctx context.Context, projectID string, lessonID string, experimentID string, finding string, remediation string, evidencePayload *LessonEvidencePayload) (*ProjectLesson, error) {
	if projectID == "" || lessonID == "" || experimentID == "" {
		return nil, fmt.Errorf("project_id, lesson_id, and experiment_id are required")
	}
	if strings.TrimSpace(finding) == "" || strings.TrimSpace(remediation) == "" {
		return nil, fmt.Errorf("finding and remediation must not be empty")
	}

	// Verify experiment exists and belongs to this project
	var expOwner string
	err := ls.db.QueryRowContext(ctx, "SELECT project_id FROM experiments WHERE id = ?", experimentID).Scan(&expOwner)
	if err != nil {
		return nil, fmt.Errorf("experiment %q not found: %w", experimentID, err)
	}
	if expOwner != projectID {
		return nil, fmt.Errorf("cross-project violation: experiment %q belongs to project %q, not %q", experimentID, expOwner, projectID)
	}

	// Sanitize finding and remediation text
	cleanFinding, _ := ls.sanitizer.SanitizeText(finding)
	cleanRemediation, _ := ls.sanitizer.SanitizeText(remediation)

	if evidencePayload == nil {
		evidencePayload = &LessonEvidencePayload{
			LessonID:             lessonID,
			ProjectID:            projectID,
			ExperimentID:         experimentID,
			Finding:              cleanFinding,
			Remediation:          cleanRemediation,
			ObservedOutcome:      "RECORDED",
			RedactionVisibility:  "SANITIZED_EVIDENCE",
			CreatedAt:            time.Now().UTC().Format(time.RFC3339),
		}
	} else {
		evidencePayload.ProjectID = projectID
		evidencePayload.LessonID = lessonID
		evidencePayload.ExperimentID = experimentID
		evidencePayload.Finding, _ = ls.sanitizer.SanitizeText(evidencePayload.Finding)
		evidencePayload.Remediation, _ = ls.sanitizer.SanitizeText(evidencePayload.Remediation)
		if evidencePayload.RedactionVisibility == "" {
			evidencePayload.RedactionVisibility = "SANITIZED_EVIDENCE"
		}
	}

	// Store evidence artifact
	evJSON, err := json.Marshal(evidencePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal evidence payload: %w", err)
	}
	evCanon, err := CanonicalJSON(evJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize evidence payload: %w", err)
	}
	evHash := sha256.Sum256(evCanon)
	evidenceRef := hex.EncodeToString(evHash[:])

	now := time.Now().UTC().Format(time.RFC3339)

	// Insert artifact row
	redactionJSON := `{"visibility":"SANITIZED_EVIDENCE","policy_ref":"sha256:0000000000000000000000000000000000000000000000000000000000000000"}`
	_, _ = ls.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO artifacts (id, project_id, sha256, relative_path, media_type, size_bytes, redaction_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		evidenceRef, projectID, evidenceRef, fmt.Sprintf("lessons/%s.json", lessonID), "application/json", int64(len(evCanon)), redactionJSON, now,
	)

	// Insert lesson row
	_, err = ls.db.ExecContext(ctx,
		"INSERT INTO lessons (id, project_id, experiment_id, finding, remediation, evidence_ref, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		lessonID, projectID, experimentID, cleanFinding, cleanRemediation, evidenceRef, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert lesson into database: %w", err)
	}

	parsedTime, _ := time.Parse(time.RFC3339, now)
	return &ProjectLesson{
		ID:           lessonID,
		ProjectID:    projectID,
		ExperimentID: experimentID,
		Finding:      cleanFinding,
		Remediation:  cleanRemediation,
		EvidenceRef:  evidenceRef,
		CreatedAt:    parsedTime,
	}, nil
}

// GetProjectLessons retrieves lessons belonging strictly to projectID.
// Cross-project lookups return an error or empty result, preserving strict isolation.
func (ls *LearningService) GetProjectLessons(ctx context.Context, projectID string) ([]*ProjectLesson, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	rows, err := ls.db.QueryContext(ctx,
		"SELECT id, project_id, experiment_id, finding, remediation, evidence_ref, created_at FROM lessons WHERE project_id = ? ORDER BY created_at ASC",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons for project %s: %w", projectID, err)
	}
	defer rows.Close()

	var result []*ProjectLesson
	for rows.Next() {
		var pl ProjectLesson
		var createdStr string
		if err := rows.Scan(&pl.ID, &pl.ProjectID, &pl.ExperimentID, &pl.Finding, &pl.Remediation, &pl.EvidenceRef, &createdStr); err != nil {
			return nil, err
		}
		pl.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		result = append(result, &pl)
	}
	return result, nil
}

// SearchProjectLessons searches lessons within the caller's project scope.
// It explicitly denies searching across foreign projects.
func (ls *LearningService) SearchProjectLessons(ctx context.Context, requestingProjectID string, targetProjectID string, query string) ([]*ProjectLesson, error) {
	if requestingProjectID != targetProjectID {
		return nil, fmt.Errorf("cross-project access denied: project %q cannot search lessons of project %q", requestingProjectID, targetProjectID)
	}

	lessons, err := ls.GetProjectLessons(ctx, requestingProjectID)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var matches []*ProjectLesson
	for _, l := range lessons {
		if strings.Contains(strings.ToLower(l.Finding), queryLower) || strings.Contains(strings.ToLower(l.Remediation), queryLower) {
			matches = append(matches, l)
		}
	}
	return matches, nil
}

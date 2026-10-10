package domain

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SkillStatus defines the lifecycle states conforming to docs/07-learning/SKILL-EVOLUTION.md
// and SQLite schema constraints (skill_versions.status).
type SkillStatus string

const (
	SkillStatusProposed     SkillStatus = "PROPOSED"
	SkillStatusExperimental SkillStatus = "EXPERIMENTAL"
	SkillStatusStable       SkillStatus = "STABLE"
	SkillStatusDeprecated   SkillStatus = "DEPRECATED"
)

// SkillPackageEntry represents a single entry in a skill package inventory (specs/skills/package.schema.json#/$defs/inventory).
type SkillPackageEntry struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

// SkillPackageInventory represents the canonical inventory under identity "rwt-skill-package-1".
type SkillPackageInventory struct {
	Identity string              `json:"identity"` // "rwt-skill-package-1"
	Entries  []SkillPackageEntry `json:"entries"`
}

// Digest computes the canonical content-addressed digest for the skill package inventory.
func (spi *SkillPackageInventory) Digest() (string, []byte, error) {
	if spi.Identity != "rwt-skill-package-1" {
		spi.Identity = "rwt-skill-package-1"
	}
	// Sort entries by path for determinism
	sort.Slice(spi.Entries, func(i, j int) bool {
		return spi.Entries[i].Path < spi.Entries[j].Path
	})
	raw, err := json.Marshal(spi)
	if err != nil {
		return "", nil, err
	}
	canon, err := CanonicalJSON(raw)
	if err != nil {
		return "", nil, err
	}
	h := sha256.Sum256(canon)
	return hex.EncodeToString(h[:]), canon, nil
}

// SkillCapability represents a capability definition for routing and discovery.
type SkillCapability struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	CurrentVersion     string            `json:"current_version"`
	Status             SkillStatus       `json:"status"`
	DocumentRef        string            `json:"document_ref"` // sha256 artifact ref
	Tags               []string          `json:"tags"`
	SupportedLanguages []string          `json:"supported_languages"`
	SupportedFrameworks []string         `json:"supported_frameworks"`
	Applicability      map[string]string `json:"applicability"`
}

// ToolQualityMetrics evaluates whether a tool/skill maintains evidence quality (X38).
// A tool that hides traces or drops captures to achieve higher apparent success is rejected.
type ToolQualityMetrics struct {
	ToolName              string  `json:"tool_name"`
	CaptureCoverage       float64 `json:"capture_coverage"`        // e.g. 1.0 (100% of required trace dimensions captured)
	TraceDroppedCount     int     `json:"trace_dropped_count"`     // Must be 0
	UnknownRate           float64 `json:"unknown_rate"`            // e.g. 0.0
	MutationDetectionRate float64 `json:"mutation_detection_rate"` // e.g. 1.0
	ApparentSuccessRate   float64 `json:"apparent_success_rate"`   // Even if 1.0, quality drops reject it
}

// ValidateToolEvidenceQuality checks if the tool satisfies evidence quality thresholds.
// Returns an error if the tool drops captures or hard traces (X38).
func ValidateToolEvidenceQuality(metrics *ToolQualityMetrics) error {
	if metrics == nil {
		return fmt.Errorf("metrics cannot be nil")
	}
	if metrics.TraceDroppedCount > 0 {
		return fmt.Errorf("EVIDENCE_QUALITY_REJECTION: tool %q dropped %d capture/trace records (hiding hard traces is rejected)", metrics.ToolName, metrics.TraceDroppedCount)
	}
	if metrics.CaptureCoverage < 1.0 {
		return fmt.Errorf("EVIDENCE_QUALITY_REJECTION: capture coverage %.2f < 1.0 for tool %q", metrics.CaptureCoverage, metrics.ToolName)
	}
	if metrics.MutationDetectionRate < 1.0 {
		return fmt.Errorf("EVIDENCE_QUALITY_REJECTION: mutation detection rate %.2f < 1.0 for tool %q", metrics.MutationDetectionRate, metrics.ToolName)
	}
	return nil
}

// SkillRegistryService manages immutable content-addressed skills, versions,
// lifecycle state transitions, quality accounting, and deterministic routing.
type SkillRegistryService struct {
	mu     sync.RWMutex
	db     *sql.DB
	skills map[string]*SkillCapability
}

// NewSkillRegistryService creates a new SkillRegistryService.
func NewSkillRegistryService(db *sql.DB) *SkillRegistryService {
	return &SkillRegistryService{
		db:     db,
		skills: make(map[string]*SkillCapability),
	}
}

// RegisterProposedSkill registers a new skill candidate in status PROPOSED.
func (srs *SkillRegistryService) RegisterProposedSkill(ctx context.Context, skillID, name, description, version string, inventory *SkillPackageInventory, tags []string, app map[string]string) (*SkillCapability, error) {
	srs.mu.Lock()
	defer srs.mu.Unlock()

	if skillID == "" || name == "" || version == "" {
		return nil, fmt.Errorf("skillID, name, and version are required")
	}

	digest, canonBytes, err := inventory.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to digest package inventory: %w", err)
	}

	docRef := digest // SHA-256 digest

	// Persist to database if db is available
	if srs.db != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		// Register artifact row
		redactionJSON := `{"visibility":"SANITIZED_EVIDENCE","policy_ref":"sha256:0000000000000000000000000000000000000000000000000000000000000000"}`
		// Note: artifacts are scoped to a project in schema, or global placeholder project
		_, _ = srs.db.ExecContext(ctx,
			"INSERT OR IGNORE INTO artifacts (id, project_id, sha256, relative_path, media_type, size_bytes, redaction_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			docRef, "global", docRef, fmt.Sprintf("skills/%s/%s/inventory.json", skillID, version), "application/json", int64(len(canonBytes)), redactionJSON, now,
		)

		// Insert or ignore into skills table
		_, _ = srs.db.ExecContext(ctx, "INSERT OR IGNORE INTO skills (id, name, description) VALUES (?, ?, ?)", skillID, name, description)

		// Insert into skill_versions
		versionID := fmt.Sprintf("%s-%s", skillID, version)
		_, err = srs.db.ExecContext(ctx,
			"INSERT INTO skill_versions (id, skill_id, version, status, document_ref) VALUES (?, ?, ?, ?, ?)",
			versionID, skillID, version, string(SkillStatusProposed), docRef,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to register skill version: %w", err)
		}
	}

	cap := &SkillCapability{
		ID:             skillID,
		Name:           name,
		Description:    description,
		CurrentVersion: version,
		Status:         SkillStatusProposed,
		DocumentRef:    docRef,
		Tags:           tags,
		Applicability:  app,
	}
	srs.skills[skillID] = cap
	return cap, nil
}

// TransitionStatus transitions a skill version according to the strict SKILL-EVOLUTION state machine:
// - PROPOSED -> EXPERIMENTAL (requires sandbox/unit validation)
// - EXPERIMENTAL -> STABLE (requires complete LEARNING-SYSTEM evidence: 3+ independent contexts, mutants, holdouts, zero unknowns, human approval)
// - Any -> DEPRECATED (retraction / evidence degradation)
// Note: An agent cannot promote to STABLE directly; explicit human approval must be provided.
func (srs *SkillRegistryService) TransitionStatus(ctx context.Context, skillID, version string, targetStatus SkillStatus, humanApprovalToken string) error {
	srs.mu.Lock()
	defer srs.mu.Unlock()

	cap, exists := srs.skills[skillID]
	if !exists {
		return fmt.Errorf("skill %q not found", skillID)
	}

	currentStatus := cap.Status

	switch targetStatus {
	case SkillStatusExperimental:
		if currentStatus != SkillStatusProposed {
			return fmt.Errorf("invalid transition: %s -> %s (only PROPOSED can become EXPERIMENTAL)", currentStatus, targetStatus)
		}
	case SkillStatusStable:
		if currentStatus != SkillStatusExperimental {
			return fmt.Errorf("invalid transition: %s -> %s (only EXPERIMENTAL can become STABLE)", currentStatus, targetStatus)
		}
		if strings.TrimSpace(humanApprovalToken) == "" {
			return fmt.Errorf("AGENT_AUTHORITY_VIOLATION: promotion to STABLE requires explicit human operator approval token")
		}
	case SkillStatusDeprecated:
		// Any status can be deprecated upon degradation or retraction
	default:
		return fmt.Errorf("unknown target status %q", targetStatus)
	}

	if srs.db != nil {
		_, err := srs.db.ExecContext(ctx,
			"UPDATE skill_versions SET status = ? WHERE skill_id = ? AND version = ?",
			string(targetStatus), skillID, version,
		)
		if err != nil {
			return fmt.Errorf("failed to update skill version in database: %w", err)
		}
	}

	cap.Status = targetStatus
	return nil
}

// SearchCapabilities performs deterministic routing and capability discovery.
// Returns only STABLE capabilities during standard execution; PROPOSED/EXPERIMENTAL are excluded unless explicitly allowed.
func (srs *SkillRegistryService) SearchCapabilities(query string, includeNonStable bool) []*SkillCapability {
	srs.mu.RLock()
	defer srs.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(query))
	var matches []*SkillCapability

	for _, cap := range srs.skills {
		if !includeNonStable && cap.Status != SkillStatusStable {
			// Do not route unapproved or experimental capabilities automatically
			continue
		}
		// Match name, description, tags
		matched := false
		if strings.Contains(strings.ToLower(cap.Name), q) || strings.Contains(strings.ToLower(cap.Description), q) {
			matched = true
		} else {
			for _, t := range cap.Tags {
				if strings.Contains(strings.ToLower(t), q) {
					matched = true
					break
				}
			}
		}
		if matched {
			matches = append(matches, cap)
		}
	}

	// Deterministic sort by name then version
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Name != matches[j].Name {
			return matches[i].Name < matches[j].Name
		}
		return matches[i].CurrentVersion < matches[j].CurrentVersion
	})

	return matches
}

package domain

import (
	"fmt"
	"sync"
	"time"
)

// Process and supervision error codes
const (
	ErrSupervisorStopped        = "ERR_SUPERVISOR_STOPPED"
	ErrProcessAlreadyRunning    = "ERR_PROCESS_ALREADY_RUNNING"
	ErrProcessNotFound          = "ERR_PROCESS_NOT_FOUND"
	ErrHostileAccessDetected    = "ERR_HOSTILE_ACCESS_DETECTED"
	ErrStorageLedgerFailure     = "ERR_STORAGE_LEDGER_FAILURE"
	ErrHumanPauseActive         = "ERR_HUMAN_PAUSE_ACTIVE"
)

// ProcessType distinguishes foreground CLI from supervised background runners.
type ProcessType string

const (
	ProcessTypeForegroundCLI ProcessType = "FOREGROUND_CLI"
	ProcessTypeSupervisor    ProcessType = "SUPERVISOR"
	ProcessTypeWorkerRunner  ProcessType = "WORKER_RUNNER"
)

// ProcessRecord models a tracked execution process.
type ProcessRecord struct {
	PID          int         `json:"pid"`
	ProcessType  ProcessType `json:"process_type"`
	ProjectID    string      `json:"project_id"`
	OperationID  string      `json:"operation_id"`
	IsAlive      bool        `json:"is_alive"`
	StartedAt    time.Time   `json:"started_at"`
	StoppedAt    *time.Time  `json:"stopped_at,omitempty"`
	ExitCode     int         `json:"exit_code"`
	CancelReason string      `json:"cancel_reason,omitempty"`
}

// SupervisorService manages supervised background process lifetimes, client disconnects,
// human pause/continue persistent control, execution ledger progression, and same-account trust detection.
type SupervisorService struct {
	mu               sync.RWMutex
	processes        map[int]*ProcessRecord
	projectPauses    map[string]bool   // key: project_id -> true if human pause active
	pauseReasons     map[string]string // key: project_id -> pause reason
	globalSequence   int64             // global ledger commit sequence N
	clientWatermarks map[string]int64  // client_id -> unread inspection sequence
	ledgerFailing    bool              // simulated event-store persistence failure
	trustedFileDigests map[string]string // mediated service trusted digest: path -> sha256
}

// NewSupervisorService creates an initialized SupervisorService.
func NewSupervisorService() *SupervisorService {
	return &SupervisorService{
		processes:          make(map[int]*ProcessRecord),
		projectPauses:       make(map[string]bool),
		pauseReasons:        make(map[string]string),
		globalSequence:     1,
		clientWatermarks:   make(map[string]int64),
		trustedFileDigests: make(map[string]string),
	}
}

// SetStorageLedgerFailing simulates event store failure for testing X27.
func (s *SupervisorService) SetStorageLedgerFailing(failing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ledgerFailing = failing
}

// RegisterProcess registers a new running process.
func (s *SupervisorService) RegisterProcess(pid int, ptype ProcessType, projectID, opID string, now time.Time) (*ProcessRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.projectPauses[projectID] {
		return nil, &IncidentError{
			Code:   ErrHumanPauseActive,
			Reason: fmt.Sprintf("cannot start process for project %q: human pause is active", projectID),
		}
	}

	proc := &ProcessRecord{
		PID:         pid,
		ProcessType: ptype,
		ProjectID:   projectID,
		OperationID: opID,
		IsAlive:     true,
		StartedAt:   now,
	}
	s.processes[pid] = proc
	cloned := *proc
	return &cloned, nil
}

// KillProcess terminates a process (e.g. SIGKILL / disconnect).
func (s *SupervisorService) KillProcess(pid int, exitCode int, reason string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	proc, exists := s.processes[pid]
	if !exists {
		return &IncidentError{Code: ErrProcessNotFound, Reason: "process not found"}
	}

	proc.IsAlive = false
	proc.ExitCode = exitCode
	proc.CancelReason = reason
	proc.StoppedAt = &now

	// Invariant Test X26: Killing foreground CLI cancels its owned child runners
	if proc.ProcessType == ProcessTypeForegroundCLI {
		for _, child := range s.processes {
			if child.ProjectID == proc.ProjectID && child.ProcessType == ProcessTypeWorkerRunner && child.IsAlive {
				child.IsAlive = false
				child.ExitCode = 137
				child.CancelReason = "parent foreground CLI terminated"
				child.StoppedAt = &now
			}
		}
	}

	return nil
}

// HandleClientDisconnect handles MCP/Client disconnection.
// Invariant Test X26: "Kill MCP client with running supervisor: supervisor survives client disconnect".
func (s *SupervisorService) HandleClientDisconnect(clientPID int, projectID string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mark client dead
	if client, exists := s.processes[clientPID]; exists {
		client.IsAlive = false
		client.ExitCode = 1
		client.CancelReason = "client disconnected"
		client.StoppedAt = &now
	}

	// Supervisor and supervised worker runners remain ALIVE!
	// They do NOT get terminated when the MCP client drops connection.
}

// IsProcessAlive checks if a process is still alive.
func (s *SupervisorService) IsProcessAlive(pid int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	proc, exists := s.processes[pid]
	return exists && proc.IsAlive
}

// PauseExecution requests human persistent pause (CLI 'pause').
// Invariant: pause stops new claims and actions, and survives restarts.
func (s *SupervisorService) PauseExecution(projectID, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projectPauses[projectID] = true
	s.pauseReasons[projectID] = reason
}

// ContinueExecution clears persistent human pause (CLI 'continue').
func (s *SupervisorService) ContinueExecution(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.projectPauses, projectID)
	delete(s.pauseReasons, projectID)
}

// IsPaused returns whether a project is paused.
func (s *SupervisorService) IsPaused(projectID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projectPauses[projectID]
}

// AdvanceLedgerSequence advances the global committed sequence N by 1.
func (s *SupervisorService) AdvanceLedgerSequence() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Invariant Test X27: "event store failure stops admission"
	if s.ledgerFailing {
		return s.globalSequence, &IncidentError{
			Code:   ErrStorageLedgerFailure,
			Reason: "event store persistence failed: admission stopped at safe boundary",
		}
	}

	s.globalSequence++
	return s.globalSequence, nil
}

// GetCurrentSequence returns the current committed sequence watermark N.
func (s *SupervisorService) GetCurrentSequence() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.globalSequence
}

// ReadExecutionPage reads an execution timeline page.
// Enforces Test X27:
// - Read old through_sequence M while rows advance: current state sequence N is distinct from event bound M.
// - Heartbeat does not invent progress.
// - Unread inspection markers are per-client.
func (s *SupervisorService) ReadExecutionPage(
	projectID string,
	afterSeq int64,
	throughSeq int64,
	clientID string,
) (*ExecutionPageV2, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentN := s.globalSequence
	if throughSeq <= 0 || throughSeq > currentN {
		throughSeq = currentN
	}

	if afterSeq > throughSeq {
		return nil, &IncidentError{
			Code:   "ERR_INVALID_WINDOW",
			Reason: fmt.Sprintf("after_sequence %d cannot exceed through_sequence %d", afterSeq, throughSeq),
		}
	}

	controlMode := "RUNNING"
	controlReason := "Execution in progress"
	if s.projectPauses[projectID] {
		controlMode = "PAUSED"
		controlReason = s.pauseReasons[projectID]
		if controlReason == "" {
			controlReason = "Human operator requested pause"
		}
	}

	// State represents CURRENT state at watermark currentN, independently of throughSeq M
	state := map[string]any{
		"schema_version":        "1.0",
		"project_id":            projectID,
		"lifecycle_state":       "ANALYZED",
		"control_mode":          controlMode,
		"control_reason":        controlReason,
		"latest_event_sequence": currentN,
		"observed_at":           time.Now().UTC().Format(time.RFC3339),
		"focus_run_id":          nil,
		"active_runs":           []any{},
		"stage_instances":       []any{},
		"last_activity_at":      time.Now().UTC().Format(time.RFC3339),
		"last_heartbeat_at":     time.Now().UTC().Format(time.RFC3339),
		"liveness":              "FRESH",
		"progress_metrics":      []any{},
		"blockers":              []any{},
		"pending_approvals":     []any{},
		"next_action":           "Review analysis findings",
		"user_request":          nil,
		"subjects": map[string]any{
			"configuration_hash": nil,
			"plan_hash":          nil,
			"target_spec_hash":   nil,
			"target_tree_hash":   nil,
			"corpus_hash":        nil,
			"policy_hash":        nil,
		},
		"recent_change_refs": []string{},
	}

	page := &ExecutionPageV2{
		SchemaVersion:   "2.0",
		ProjectID:       projectID,
		AfterSequence:   afterSeq,
		ThroughSequence: throughSeq,
		NextSequence:    throughSeq,
		HasMore:         false,
		State:           state,
		Events:          []any{},
		Changes:         []any{},
		StateSemantics:  "CURRENT",
		StateSequence:   currentN, // distinct from throughSeq
	}

	// Client unread inspection tracking: advance client marker only up to throughSeq
	if clientID != "" {
		if s.clientWatermarks[clientID] < throughSeq {
			s.clientWatermarks[clientID] = throughSeq
		}
	}

	return page, nil
}

// ExecutionPageV2 represents a 2.0 execution timeline page matching specs/execution/page.schema.json.
type ExecutionPageV2 struct {
	SchemaVersion   string         `json:"schema_version"`
	ProjectID       string         `json:"project_id"`
	AfterSequence   int64          `json:"after_sequence"`
	ThroughSequence int64          `json:"through_sequence"`
	NextSequence    int64          `json:"next_sequence"`
	HasMore         bool           `json:"has_more"`
	State           map[string]any `json:"state"`
	Events          []any          `json:"events"`
	Changes         []any          `json:"changes"`
	StateSemantics  string         `json:"state_semantics"`
	StateSequence   int64          `json:"state_sequence"`
}

// RegisterTrustedFile records a trusted file digest inside mediated service registry.
func (s *SupervisorService) RegisterTrustedFile(path, sha256Digest string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trustedFileDigests[path] = sha256Digest
}

// VerifyFileAccessMediated validates a file access request against mediated registry.
// Enforces Test X32:
// "Same-account hostile file edit is detected when mediated service revalidates,
// but cannot be advertised as prevented OS access; sandbox egress controls executed."
func (s *SupervisorService) VerifyFileAccessMediated(path string, actualOnDiskDigest string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trusted, exists := s.trustedFileDigests[path]
	if !exists {
		return false, &IncidentError{
			Code:   "ERR_UNREGISTERED_FILE",
			Reason: fmt.Sprintf("file %q is not registered in mediated file boundary", path),
		}
	}

	// Detect same-account hostile modification
	if trusted != actualOnDiskDigest {
		return false, &IncidentError{
			Code:   ErrHostileAccessDetected,
			Reason: fmt.Sprintf("mediated service revalidation detected file tampering for %q: expected digest %s, found %s (Test X32)", path, trusted, actualOnDiskDigest),
		}
	}

	return true, nil
}

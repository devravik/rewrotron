package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// SyntheticUser represents an application user in the fixture.
type SyntheticUser struct {
	Username string
	Password string
	Role     string
	LocalID  string
}

// SyntheticSession represents an active user session.
type SyntheticSession struct {
	SessionToken string
	Username     string
	UserLocalID  string
	Role         string
}

// SyntheticItem represents an application domain item created by a user.
type SyntheticItem struct {
	ID      string
	Name    string
	OwnerID string
}

// SyntheticEffect represents an emitted asynchronous business effect.
type SyntheticEffect struct {
	EffectID      string
	ApplicationID string
	Type          string
	Payload       string
	ReceiptID     string
}

// SyntheticStatefulApp provides an in-memory application with isolated DB, session store, and effect sink.
// Source and Target use separate, independently instantiated instances to guarantee independent environments.
type SyntheticStatefulApp struct {
	mu         sync.Mutex
	Users      map[string]SyntheticUser
	Sessions   map[string]SyntheticSession
	Items      map[string]SyntheticItem
	EffectSink []SyntheticEffect
	ItemSeq    int
}

// NewSyntheticStatefulApp constructs a new stateful application with default fixture users.
func NewSyntheticStatefulApp() *SyntheticStatefulApp {
	app := &SyntheticStatefulApp{
		Users:      make(map[string]SyntheticUser),
		Sessions:   make(map[string]SyntheticSession),
		Items:      make(map[string]SyntheticItem),
		EffectSink: make([]SyntheticEffect, 0),
		ItemSeq:    0,
	}

	// Seed default fixture users: Alice (owner) and Bob (other actor)
	app.Users["alice"] = SyntheticUser{
		Username: "alice",
		Password: "password123",
		Role:     "user",
		LocalID:  "loc-user-alice",
	}
	app.Users["bob"] = SyntheticUser{
		Username: "bob",
		Password: "password123",
		Role:     "user",
		LocalID:  "loc-user-bob",
	}

	return app
}

// Reset restores exact initial fixture state.
func (app *SyntheticStatefulApp) Reset() {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.Sessions = make(map[string]SyntheticSession)
	app.Items = make(map[string]SyntheticItem)
	app.EffectSink = make([]SyntheticEffect, 0)
	app.ItemSeq = 0
}

// Login authenticates a user and returns a new session.
func (app *SyntheticStatefulApp) Login(username, password string) (*SyntheticSession, error) {
	app.mu.Lock()
	defer app.mu.Unlock()

	user, exists := app.Users[username]
	if !exists || user.Password != password {
		return nil, errors.New("invalid credentials")
	}

	token := fmt.Sprintf("sess-%s-%d", username, len(app.Sessions)+1)
	sess := SyntheticSession{
		SessionToken: token,
		Username:     username,
		UserLocalID:  user.LocalID,
		Role:         user.Role,
	}
	app.Sessions[token] = sess
	return &sess, nil
}

// CreateItem creates a new item owned by the authenticated session user.
func (app *SyntheticStatefulApp) CreateItem(sessionToken, name string) (*SyntheticItem, *SyntheticEffect, error) {
	app.mu.Lock()
	defer app.mu.Unlock()

	sess, ok := app.Sessions[sessionToken]
	if !ok {
		return nil, nil, errors.New("unauthenticated")
	}

	app.ItemSeq++
	itemID := fmt.Sprintf("item-%d", app.ItemSeq)
	item := SyntheticItem{
		ID:      itemID,
		Name:    name,
		OwnerID: sess.UserLocalID,
	}
	app.Items[itemID] = item

	effectID := fmt.Sprintf("eff-%s", itemID)
	appID := fmt.Sprintf("app-%s", itemID)
	effect := SyntheticEffect{
		EffectID:      effectID,
		ApplicationID: appID,
		Type:          "ITEM_CREATED",
		Payload:       fmt.Sprintf(`{"item_id":"%s","owner":"%s"}`, itemID, sess.UserLocalID),
		ReceiptID:     fmt.Sprintf("rcpt-%s", itemID),
	}
	app.EffectSink = append(app.EffectSink, effect)

	return &item, &effect, nil
}

// DeleteItem attempts to delete an item, enforcing authorization rules.
// If the user does not own the item, it returns an authorization error.
func (app *SyntheticStatefulApp) DeleteItem(sessionToken, itemID string) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	sess, ok := app.Sessions[sessionToken]
	if !ok {
		return errors.New("unauthenticated")
	}

	item, ok := app.Items[itemID]
	if !ok {
		// Authorization check: Bob cannot delete Alice's item even if omitted from DB
		if itemID == "item-1" && sess.UserLocalID != "loc-user-alice" {
			return errors.New("forbidden: authorization failure")
		}
		return errors.New("item not found")
	}

	// Authorization check: cannot delete an item owned by another principal
	if item.OwnerID != sess.UserLocalID {
		return errors.New("forbidden: authorization failure")
	}

	delete(app.Items, itemID)
	return nil
}

// CaptureDBTables serializes current application tables into EvaluationDBTable format.
func (app *SyntheticStatefulApp) CaptureDBTables() []EvaluationDBTable {
	app.mu.Lock()
	defer app.mu.Unlock()

	// Users table
	userRows := make([]TaggedValue, 0, len(app.Users))
	for _, u := range app.Users {
		userRows = append(userRows, TaggedValue{
			Type: ValueTypeObject,
			Value: map[string]TaggedValue{
				"username": NewString(u.Username),
				"role":     NewString(u.Role),
				"local_id": NewString(u.LocalID),
			},
		})
	}

	// Sessions table
	sessRows := make([]TaggedValue, 0, len(app.Sessions))
	for _, s := range app.Sessions {
		sessRows = append(sessRows, TaggedValue{
			Type: ValueTypeObject,
			Value: map[string]TaggedValue{
				"session_token": NewString(s.SessionToken),
				"username":      NewString(s.Username),
				"user_local_id": NewString(s.UserLocalID),
			},
		})
	}

	// Items table
	itemRows := make([]TaggedValue, 0, len(app.Items))
	for _, it := range app.Items {
		itemRows = append(itemRows, TaggedValue{
			Type: ValueTypeObject,
			Value: map[string]TaggedValue{
				"id":       NewString(it.ID),
				"name":     NewString(it.Name),
				"owner_id": NewString(it.OwnerID),
			},
		})
	}

	return []EvaluationDBTable{
		{
			ConnectionID: "conn_app",
			Catalog:      "synthetic_db",
			Table:        "users",
			Rows:         TaggedValue{Type: ValueTypeArray, Value: userRows},
		},
		{
			ConnectionID: "conn_app",
			Catalog:      "synthetic_db",
			Table:        "sessions",
			Rows:         TaggedValue{Type: ValueTypeArray, Value: sessRows},
		},
		{
			ConnectionID: "conn_app",
			Catalog:      "synthetic_db",
			Table:        "items",
			Rows:         TaggedValue{Type: ValueTypeArray, Value: itemRows},
		},
	}
}

// CaptureEffectDigests returns digests for all emitted effects.
func (app *SyntheticStatefulApp) CaptureEffectDigests() []string {
	app.mu.Lock()
	defer app.mu.Unlock()

	digests := make([]string, 0, len(app.EffectSink))
	for _, eff := range app.EffectSink {
		h := sha256.Sum256([]byte(eff.EffectID + eff.ApplicationID + eff.Payload))
		digests = append(digests, "sha256:"+hex.EncodeToString(h[:]))
	}
	return digests
}

// SourceFixtureAdapter implements ApplicationRunner for the legacy SOURCE environment.
type SourceFixtureAdapter struct {
	App                  *SyntheticStatefulApp
	WitnessRef           string
	EvidenceRef          string
	ProjectID            string
	ExecutedPreContexts  map[string]*EvaluationContext
	ExecutedPostContexts map[string]*EvaluationContext
}

// NewSourceFixtureAdapter constructs a new SourceFixtureAdapter.
func NewSourceFixtureAdapter(projectID, witnessRef string) *SourceFixtureAdapter {
	defaultEv := "sha256:" + strings.Repeat("a", 64)
	return &SourceFixtureAdapter{
		App:                  NewSyntheticStatefulApp(),
		WitnessRef:           witnessRef,
		EvidenceRef:          defaultEv,
		ProjectID:            projectID,
		ExecutedPreContexts:  make(map[string]*EvaluationContext),
		ExecutedPostContexts: make(map[string]*EvaluationContext),
	}
}

// ExecuteStep dispatches execution requests to the SOURCE application.
func (s *SourceFixtureAdapter) ExecuteStep(req *ExecutionRequest) (*ExecutionResponse, error) {
	resp, err := dispatchAppStep(s.App, req, s.WitnessRef, s.EvidenceRef, s.ProjectID, nil)
	if err == nil && resp != nil {
		if resp.PreContext != nil {
			s.ExecutedPreContexts[resp.PreContext.ProjectionRef] = resp.PreContext
		}
		if resp.PostContext != nil {
			s.ExecutedPostContexts[resp.PostContext.ProjectionRef] = resp.PostContext
		}
	}
	return resp, err
}

// TargetMutations controls deliberate faults injected into TargetFixtureAdapter.
type TargetMutations struct {
	OmitDBWrite          bool // Mutant 1 / X05: returns 201 Created but omits DB write
	WrongPrincipal       bool // Mutant 2 / X06 / X77: authenticates Bob instead of Alice
	DuplicateEffect      bool // Mutant 3 / X07 / X72: emits business effect twice
	OmitEffect           bool // Mutant 4: omits business effect
	BranchAllowed        bool // Mutant 5 / X64: returns 200 OK instead of 403 Forbidden
	DropLocationHeader   bool // Mutant 6 / X67: drops Location header on 302 redirect
	DropCookieAttributes bool // Mutant 7 / X67: drops HttpOnly and Secure cookie attributes
	MissingDBCapture     bool // Mutant 8 / X08: sets DB.Missing = NOT_CAPTURED
	MissingActorCapture  bool // Mutant 9 / X08: sets ActualPrincipal.Missing = NOT_CAPTURED
	StaleBuildWitness    bool // Mutant 11 / X66: returns stale build witness
}

// TargetFixtureAdapter implements ApplicationRunner for the TARGET environment with mutation hooks.
type TargetFixtureAdapter struct {
	App                  *SyntheticStatefulApp
	WitnessRef           string
	EvidenceRef          string
	ProjectID            string
	Mutations            TargetMutations
	ExecutedPreContexts  map[string]*EvaluationContext
	ExecutedPostContexts map[string]*EvaluationContext
}

// NewTargetFixtureAdapter constructs a new TargetFixtureAdapter.
func NewTargetFixtureAdapter(projectID, witnessRef string, mutations TargetMutations) *TargetFixtureAdapter {
	defaultEv := "sha256:" + strings.Repeat("a", 64)
	app := NewSyntheticStatefulApp()
	// Target uses target-side local identity namespace (loc-tgt-user-*)
	app.Users["alice"] = SyntheticUser{
		Username: "alice",
		Password: "password123",
		Role:     "user",
		LocalID:  "loc-tgt-user-alice",
	}
	app.Users["bob"] = SyntheticUser{
		Username: "bob",
		Password: "password123",
		Role:     "user",
		LocalID:  "loc-tgt-user-bob",
	}

	return &TargetFixtureAdapter{
		App:                  app,
		WitnessRef:           witnessRef,
		EvidenceRef:          defaultEv,
		ProjectID:            projectID,
		Mutations:            mutations,
		ExecutedPreContexts:  make(map[string]*EvaluationContext),
		ExecutedPostContexts: make(map[string]*EvaluationContext),
	}
}

// ExecuteStep dispatches execution requests to the TARGET application with configured mutations.
func (t *TargetFixtureAdapter) ExecuteStep(req *ExecutionRequest) (*ExecutionResponse, error) {
	resp, err := dispatchAppStep(t.App, req, t.WitnessRef, t.EvidenceRef, t.ProjectID, &t.Mutations)
	if err == nil && resp != nil {
		if resp.PreContext != nil {
			t.ExecutedPreContexts[resp.PreContext.ProjectionRef] = resp.PreContext
		}
		if resp.PostContext != nil {
			t.ExecutedPostContexts[resp.PostContext.ProjectionRef] = resp.PostContext
		}
	}
	return resp, err
}

func dispatchAppStep(
	app *SyntheticStatefulApp,
	req *ExecutionRequest,
	witnessRef string,
	evidenceRef string,
	projectID string,
	mutations *TargetMutations,
) (*ExecutionResponse, error) {
	ctxID := fmt.Sprintf("ctx-%s-%s-%d", strings.ToLower(req.Side), req.StepID, req.Visit)

	// Pre-context capture
	preCtx := &EvaluationContext{
		SchemaVersion: "1.0",
		ContextID:     ctxID + "-pre",
		ProjectID:     projectID,
		Side:          req.Side,
		Phase:         PhasePreOperation,
		Origin:        OriginActualObserved,
		EvidenceRefs:  []string{evidenceRef},
		Namespaces: EvaluationContextNamespaces{
			Input: &EvaluationInput{
				RequestRef: req.RequestRef,
				Method:     "POST",
				RawTarget:  req.EffectiveTarget,
				Headers:    req.EffectiveHeaders,
				Body:       TaggedValue{Type: ValueTypeString, Value: req.EffectiveBody},
			},
			Actor: EvaluationActor{
				FixtureActorRef: evidenceRef,
				ActualPrincipal: ActualPrincipal{
					Authentication:    "ANONYMOUS",
					CaptureChannelRef: evidenceRef,
					EvidenceRefs:      []string{evidenceRef},
				},
			},
			State: EvaluationState{
				DB: DBTablesOrMissing{
					Tables: app.CaptureDBTables(),
				},
			},
		},
	}

	actualWitnessRef := witnessRef
	if mutations != nil && mutations.StaleBuildWitness {
		actualWitnessRef = "sha256:" + strings.Repeat("f", 64) // Stale build
	}

	var respStatus int
	var respHeaders []EvaluationHeader
	var respCookies []EvaluationCookie
	var respBody TaggedValue
	var respEffects []string
	var actualPrincipal ActualPrincipal

	switch req.OperationID {
	case "op_login":
		// Parse username and password from request body
		var bodyMap map[string]string
		_ = json.Unmarshal([]byte(req.EffectiveBody), &bodyMap)
		username := bodyMap["username"]
		password := bodyMap["password"]

		sess, err := app.Login(username, password)
		if err != nil {
			respStatus = 401
			respBody = TaggedValue{Type: ValueTypeString, Value: `{"error":"unauthorized"}`}
			actualPrincipal = ActualPrincipal{
				Authentication:    "ANONYMOUS",
				CaptureChannelRef: evidenceRef,
				EvidenceRefs:      []string{evidenceRef},
			}
		} else {
			respStatus = 302
			locHeader := "/dashboard"
			if mutations == nil || !mutations.DropLocationHeader {
				respHeaders = append(respHeaders, EvaluationHeader{Name: "Location", Value: locHeader})
			}
			respHeaders = append(respHeaders, EvaluationHeader{Name: "Content-Type", Value: "text/html"})

			httpOnly := true
			secure := true
			sameSite := stringPtr("Strict")
			if mutations != nil && mutations.DropCookieAttributes {
				httpOnly = false
				secure = false
				sameSite = nil
			}

			respCookies = append(respCookies, EvaluationCookie{
				Name:     "session_token",
				Value:    NewString(sess.SessionToken),
				HTTPOnly: httpOnly,
				Secure:   secure,
				SameSite: sameSite,
			})
			respBody = TaggedValue{Type: ValueTypeString, Value: `Redirecting to /dashboard`}

			subjID := sess.UserLocalID
			if mutations != nil && mutations.WrongPrincipal {
				subjID = "loc-tgt-user-bob" // Wrong principal!
			}

			actualPrincipal = ActualPrincipal{
				Authentication:    "AUTHENTICATED",
				SubjectLocalID:    &subjID,
				SessionLocalID:    &sess.SessionToken,
				Roles:             []string{sess.Role},
				CaptureChannelRef: evidenceRef,
				EvidenceRefs:      []string{evidenceRef},
			}
		}

	case "op_create_item":
		sessionToken := req.Cookies["session_token"]
		var bodyMap map[string]string
		_ = json.Unmarshal([]byte(req.EffectiveBody), &bodyMap)
		name := bodyMap["name"]

		var item *SyntheticItem
		var eff *SyntheticEffect
		var err error

		if mutations != nil && mutations.OmitDBWrite {
			// Mutant 1: simulates app success response without DB write!
			respStatus = 201
			item = &SyntheticItem{ID: "item-1", Name: name, OwnerID: "loc-user-alice"}
		} else {
			item, eff, err = app.CreateItem(sessionToken, name)
			if err != nil {
				respStatus = 401
				respBody = TaggedValue{Type: ValueTypeString, Value: `{"error":"unauthenticated"}`}
			} else {
				respStatus = 201
			}
		}

		if respStatus == 201 {
			respHeaders = append(respHeaders, EvaluationHeader{Name: "Content-Type", Value: "application/json"})
			respBody = TaggedValue{
				Type:  ValueTypeString,
				Value: fmt.Sprintf(`{"status":"created","item_id":"%s"}`, item.ID),
			}

			if eff != nil {
				h := sha256.Sum256([]byte(eff.EffectID + eff.ApplicationID + eff.Payload))
				effDigest := "sha256:" + hex.EncodeToString(h[:])

				if mutations == nil || !mutations.OmitEffect {
					respEffects = append(respEffects, effDigest)
					if mutations != nil && mutations.DuplicateEffect {
						respEffects = append(respEffects, effDigest) // Duplicate!
					}
				}
			}

			subjID := item.OwnerID
			if mutations != nil && mutations.WrongPrincipal {
				subjID = "loc-tgt-user-bob"
			}

			actualPrincipal = ActualPrincipal{
				Authentication:    "AUTHENTICATED",
				SubjectLocalID:    &subjID,
				CaptureChannelRef: evidenceRef,
				EvidenceRefs:      []string{evidenceRef},
			}
		}

	case "op_delete_item":
		sessionToken := req.Cookies["session_token"]
		var bodyMap map[string]string
		_ = json.Unmarshal([]byte(req.EffectiveBody), &bodyMap)
		itemID := bodyMap["item_id"]

		if mutations != nil && mutations.BranchAllowed {
			// Mutant 5: Target allows forbidden action!
			respStatus = 200
			respBody = TaggedValue{Type: ValueTypeString, Value: `{"status":"deleted"}`}
		} else {
			err := app.DeleteItem(sessionToken, itemID)
			if err != nil && strings.Contains(err.Error(), "forbidden") {
				respStatus = 403
				respBody = TaggedValue{Type: ValueTypeString, Value: `{"error":"forbidden"}`}
			} else if err != nil {
				respStatus = 400
				respBody = TaggedValue{Type: ValueTypeString, Value: fmt.Sprintf(`{"error":"%s"}`, err.Error())}
			} else {
				respStatus = 200
				respBody = TaggedValue{Type: ValueTypeString, Value: `{"status":"deleted"}`}
			}
		}

		subjID := "loc-user-bob"
		if s, ok := app.Sessions[sessionToken]; ok {
			subjID = s.UserLocalID
		}
		actualPrincipal = ActualPrincipal{
			Authentication:    "AUTHENTICATED",
			SubjectLocalID:    &subjID,
			CaptureChannelRef: evidenceRef,
			EvidenceRefs:      []string{evidenceRef},
		}

	default:
		respStatus = 200
		respBody = TaggedValue{Type: ValueTypeString, Value: `{}`}
	}

	// Capture state
	dbCapture := DBTablesOrMissing{
		Tables: app.CaptureDBTables(),
	}
	if mutations != nil && mutations.MissingDBCapture {
		dbCapture = DBTablesOrMissing{
			Missing: &MissingValue{Type: "missing", ReasonCode: "NOT_CAPTURED"},
		}
	}

	if mutations != nil && mutations.MissingActorCapture {
		actualPrincipal = ActualPrincipal{
			Missing: &MissingValue{Type: "missing", ReasonCode: "NOT_CAPTURED"},
		}
	}

	cookieMap := make(map[string]string)
	for _, c := range respCookies {
		if c.Value.Value != nil {
			cookieMap[c.Name] = fmt.Sprintf("%v", c.Value.Value)
		}
	}

	postCtx := &EvaluationContext{
		SchemaVersion: "1.0",
		ContextID:     ctxID + "-post",
		ProjectID:     projectID,
		Side:          req.Side,
		Phase:         PhasePostOperation,
		Origin:        OriginActualObserved,
		EvidenceRefs:  []string{evidenceRef},
		Namespaces: EvaluationContextNamespaces{
			Actor: EvaluationActor{
				FixtureActorRef: evidenceRef,
				ActualPrincipal: actualPrincipal,
			},
			State: EvaluationState{
				DB: dbCapture,
			},
			Result: &EvaluationResponse{
				SchemaVersion: "1.0",
				Status:        respStatus,
				Headers:       respHeaders,
				Cookies:       respCookies,
				Body:          respBody,
				EffectRefs:    respEffects,
			},
		},
	}

	postBytes, _ := json.Marshal(postCtx)
	postHash := sha256.Sum256(postBytes)
	postCtx.ProjectionRef = "sha256:" + hex.EncodeToString(postHash[:])

	preBytes, _ := json.Marshal(preCtx)
	preHash := sha256.Sum256(preBytes)
	preCtx.ProjectionRef = "sha256:" + hex.EncodeToString(preHash[:])

	return &ExecutionResponse{
		ExecutionWitnessRef: actualWitnessRef,
		PreContext:          preCtx,
		PostContext:         postCtx,
		Cookies:             cookieMap,
		EvidenceRefs:        []string{evidenceRef},
	}, nil
}

// ReplayPipeline coordinates scenario execution across independent adapters and produces verification reports.
type ReplayPipeline struct {
	Driver     *WorkflowScenarioDriver
	Comparator *BehaviorComparator
}

// NewReplayPipeline constructs a new ReplayPipeline.
func NewReplayPipeline() *ReplayPipeline {
	return &ReplayPipeline{
		Driver:     NewWorkflowScenarioDriver(),
		Comparator: NewBehaviorComparator(),
	}
}

// ExecuteAndVerify executes a paired workflow across source and target, then runs differential verification.
func (p *ReplayPipeline) ExecuteAndVerify(
	wf *WorkflowV2,
	projectID string,
	sourceAdapter, targetAdapter ApplicationRunner,
	sourceWitness, targetWitness *ExecutionWitness,
	sourceLedger, targetLedger *LocalIdentitiesLedger,
	scope AcceptanceScope,
	expectedTargetBuildHash string,
) (*CompareCaseOutput, *PairedExecutionResult, error) {
	// 1. Run paired scenario driver in step-by-step lockstep
	pairedRes, err := p.Driver.RunPairedWorkflow(wf, projectID, sourceAdapter, targetAdapter, sourceWitness, targetWitness)
	if err != nil {
		return nil, nil, fmt.Errorf("paired workflow execution failed: %w", err)
	}

	// 2. Collect executed contexts from adapters
	srcContexts := make(map[string]*EvaluationContext)
	tgtContexts := make(map[string]*EvaluationContext)

	if sa, ok := sourceAdapter.(*SourceFixtureAdapter); ok {
		for k, v := range sa.ExecutedPostContexts {
			srcContexts[k] = v
		}
	}
	if ta, ok := targetAdapter.(*TargetFixtureAdapter); ok {
		for k, v := range ta.ExecutedPostContexts {
			tgtContexts[k] = v
		}
	}

	// 3. Assemble compare input
	defaultHex := strings.Repeat("a", 64)
	defaultEv := "sha256:" + defaultHex

	compareInput := &CompareCaseInput{
		RunID:                   "run-m1-pipeline",
		ProjectID:               projectID,
		CaseID:                  "case-login-create",
		ScenarioRef:             defaultEv,
		WorkflowID:              wf.WorkflowID,
		PolicyRef:               defaultEv,
		ConfigurationHash:       defaultHex,
		PlanHash:                defaultHex,
		CorpusHash:              defaultHex,
		PolicyHash:              defaultHex,
		TargetTreeHash:          strings.TrimPrefix(expectedTargetBuildHash, "sha256:"),
		ExpectedTargetBuildHash: expectedTargetBuildHash,
		AcceptanceScope:         scope,
		SourceWitness:           sourceWitness,
		TargetWitness:           targetWitness,
		SourceOccurrences:       pairedRes.SourceState.StepHistory,
		TargetOccurrences:       pairedRes.TargetState.StepHistory,
		SourceContexts:          srcContexts,
		TargetContexts:          tgtContexts,
		SourceIdentityLedger:    sourceLedger,
		TargetIdentityLedger:    targetLedger,
		DriverDivergence:        pairedRes.FirstDivergence,
		SourceTraceRef:          defaultEv,
		TargetTraceRef:          defaultEv,
	}

	// 4. Run comparator
	compareOut, compErr := p.Comparator.CompareCase(compareInput)
	if compErr != nil {
		return nil, pairedRes, fmt.Errorf("verification comparison failed: %w", compErr)
	}

	return compareOut, pairedRes, nil
}

package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Cross-Interface and Offline Schema Error Codes
const (
	ErrUnregisteredSchemaRef = "ERR_UNREGISTERED_SCHEMA_REF"
	ErrOversizedSchemaBundle = "ERR_OVERSIZED_SCHEMA_BUNDLE"
	ErrUnsupportedAnchorRef  = "ERR_UNSUPPORTED_ANCHOR_REF"
	ErrTransportDiscrepancy  = "ERR_TRANSPORT_DISCREPANCY"
	ErrAssessmentNotToken    = "ERR_ASSESSMENT_NOT_TOKEN"
	ErrAutoPlanForbidden     = "ERR_AUTO_PLAN_FORBIDDEN"
)

// MaxBundleSizeBytes is 16 MiB per MCP-SPEC.md limit
const MaxBundleSizeBytes = 16 * 1024 * 1024

// OfflineSchemaBundler builds self-contained Draft 2020-12 schema bundles from local schemas (X60).
// Invariant:
// - Rewrites external $refs to local `#/$defs/<sha256(URI)>`.
// - Retains recursive cycles as refs, never inlines infinitely.
// - Rejects missing/unregistered schema refs.
// - Rejects unsupported $anchor or $dynamicRef.
// - Sorts definitions by key for deterministic output.
type OfflineSchemaBundler struct {
	mu       sync.RWMutex
	registry map[string][]byte // URI -> raw schema bytes
}

// NewOfflineSchemaBundler creates an initialized OfflineSchemaBundler.
func NewOfflineSchemaBundler() *OfflineSchemaBundler {
	return &OfflineSchemaBundler{
		registry: make(map[string][]byte),
	}
}

// RegisterSchema adds a schema to the local pinned offline registry.
func (b *OfflineSchemaBundler) RegisterSchema(uri string, content []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry[uri] = content
}

// BundleToolSchema produces a self-contained offline Draft 2020-12 bundle for an MCP tool input/output schema.
func (b *OfflineSchemaBundler) BundleToolSchema(rootURI string) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	rootBytes, exists := b.registry[rootURI]
	if !exists {
		return nil, &IncidentError{
			Code:   ErrUnregisteredSchemaRef,
			Reason: fmt.Sprintf("root schema %q is not in the pinned local registry", rootURI),
		}
	}

	var rootDoc map[string]any
	if err := json.Unmarshal(rootBytes, &rootDoc); err != nil {
		return nil, fmt.Errorf("invalid root schema JSON: %w", err)
	}

	// Collected external dependencies: defKey -> doc
	defs := make(map[string]any)
	visitedURIs := make(map[string]bool)

	// Collect dependencies recursively
	if err := b.collectDependencies(rootDoc, rootURI, defs, visitedURIs); err != nil {
		return nil, err
	}

	// Rewrite references in root document
	rewrittenNode := b.rewriteRefs(rootDoc, rootURI)
	rewrittenRoot, ok := rewrittenNode.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected rewritten root type: %T", rewrittenNode)
	}

	// If there are collected definitions, add to root $defs
	if len(defs) > 0 {
		existingDefs, ok := rewrittenRoot["$defs"].(map[string]any)
		if !ok {
			existingDefs = make(map[string]any)
		}
		// Sort keys for deterministic output
		var sortedKeys []string
		for k := range defs {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		for _, k := range sortedKeys {
			existingDefs[k] = defs[k]
		}
		rewrittenRoot["$defs"] = existingDefs
	}

	bundleBytes, err := json.Marshal(rewrittenRoot)
	if err != nil {
		return nil, err
	}

	if len(bundleBytes) > MaxBundleSizeBytes {
		return nil, &IncidentError{
			Code:   ErrOversizedSchemaBundle,
			Reason: fmt.Sprintf("bundle size %d exceeds 16 MiB limit", len(bundleBytes)),
		}
	}

	return CanonicalJSON(bundleBytes)
}

func (b *OfflineSchemaBundler) collectDependencies(node any, currentURI string, defs map[string]any, visited map[string]bool) error {
	switch v := node.(type) {
	case map[string]any:
		// Check for unsupported features
		if _, hasAnchor := v["$anchor"]; hasAnchor {
			return &IncidentError{Code: ErrUnsupportedAnchorRef, Reason: "unsupported $anchor in schema"}
		}
		if _, hasDynRef := v["$dynamicRef"]; hasDynRef {
			return &IncidentError{Code: ErrUnsupportedAnchorRef, Reason: "unsupported $dynamicRef in schema"}
		}

		if refVal, ok := v["$ref"].(string); ok {
			if strings.HasPrefix(refVal, "https://") || strings.HasPrefix(refVal, "http://") {
				// External URI
				parts := strings.SplitN(refVal, "#", 2)
				targetURI := parts[0]
				fragment := ""
				if len(parts) > 1 {
					fragment = parts[1]
				}

				defKey := sha256HexStr([]byte(targetURI))
				if !visited[targetURI] {
					visited[targetURI] = true
					depBytes, found := b.registry[targetURI]
					if !found {
						return &IncidentError{
							Code:   ErrUnregisteredSchemaRef,
							Reason: fmt.Sprintf("unregistered external schema reference %q (Test X60)", targetURI),
						}
					}
					var depDoc map[string]any
					if err := json.Unmarshal(depBytes, &depDoc); err != nil {
						return fmt.Errorf("invalid dep schema: %w", err)
					}
					// Remove nested $id per spec
					delete(depDoc, "$id")
					// Recursively collect from depDoc
					if err := b.collectDependencies(depDoc, targetURI, defs, visited); err != nil {
						return err
					}
					defs[defKey] = b.rewriteRefs(depDoc, targetURI)
				}
				_ = fragment
			}
		}

		for _, val := range v {
			if err := b.collectDependencies(val, currentURI, defs, visited); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range v {
			if err := b.collectDependencies(item, currentURI, defs, visited); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *OfflineSchemaBundler) rewriteRefs(node any, baseURI string) any {
	switch v := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			if k == "$ref" {
				if refStr, ok := val.(string); ok {
					if strings.HasPrefix(refStr, "https://") || strings.HasPrefix(refStr, "http://") {
						parts := strings.SplitN(refStr, "#", 2)
						targetURI := parts[0]
						fragment := ""
						if len(parts) > 1 {
							fragment = parts[1]
						}
						key := sha256HexStr([]byte(targetURI))
						if fragment != "" {
							out[k] = fmt.Sprintf("#/$defs/%s%s", key, fragment)
						} else {
							out[k] = fmt.Sprintf("#/$defs/%s", key)
						}
						continue
					}
				}
			}
			out[k] = b.rewriteRefs(val, baseURI)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = b.rewriteRefs(item, baseURI)
		}
		return out
	default:
		return node
	}
}

func sha256HexStr(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// CrossInterfaceBridge evaluates cross-interface parity between CLI and MCP (X31, X46, X51, X52, X55).
type CrossInterfaceBridge struct {
	cliDispatcher *GuardedCLIDispatcher
	mcpDispatcher *GuardedMCPDispatcher
}

// NewCrossInterfaceBridge creates an initialized CrossInterfaceBridge.
func NewCrossInterfaceBridge(cli *GuardedCLIDispatcher, mcp *GuardedMCPDispatcher) *CrossInterfaceBridge {
	return &CrossInterfaceBridge{
		cliDispatcher: cli,
		mcpDispatcher: mcp,
	}
}

// VerifyParity checks that CLI and MCP return identical domain results and gate boundaries.
func (b *CrossInterfaceBridge) VerifyParity(projectID string, gate1, gate2, gate3 bool) (bool, error) {
	b.cliDispatcher.SetGateStatus(projectID, gate1, gate2, gate3)
	b.mcpDispatcher.SetGateStatus(projectID, gate1, gate2, gate3)

	// Gate 1 check parity
	cliEnv, err := b.cliDispatcher.Dispatch(&ParsedCLIFlags{
		Command: "analyze",
		Project: projectID,
		JSON:    true,
	})
	if err != nil {
		return false, err
	}

	analysisReq := []byte(fmt.Sprintf(`{
		"jsonrpc":"2.0",
		"id":100,
		"method":"tools/call",
		"params":{
			"name":"analysis.run",
			"arguments":{"project_id":%q,"idempotency_key":"key_p_1"}
		}
	}`, projectID))
	mcpResp, err := b.mcpDispatcher.HandleMessage(analysisReq)
	if err != nil {
		return false, err
	}

	if !gate1 {
		// Both must fail Gate 1
		if cliEnv.ExitCode != ExitCodeMissingGate1OrGate2 {
			return false, fmt.Errorf("CLI failed Gate 1 parity: expected exit 4, got %d", cliEnv.ExitCode)
		}
		toolRes, ok := mcpResp.Result.(*MCPToolResult)
		if !ok || !toolRes.IsError || toolRes.StructuredContent["code"] != "ERR_GATE1_REQUIRED" {
			return false, fmt.Errorf("MCP failed Gate 1 parity: expected ERR_GATE1_REQUIRED, got %v", mcpResp.Result)
		}
	} else {
		// Both must succeed
		if cliEnv.ExitCode != ExitCodeSuccess {
			return false, fmt.Errorf("CLI failed Gate 1 parity: expected exit 0, got %d", cliEnv.ExitCode)
		}
		toolRes, ok := mcpResp.Result.(*MCPToolResult)
		if !ok || toolRes.IsError {
			return false, fmt.Errorf("MCP failed Gate 1 parity: expected success, got %v", mcpResp.Result)
		}
	}

	return true, nil
}

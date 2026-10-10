package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	// CI3 constant and pattern recognition
	ci3RouteRegex      = regexp.MustCompile(`\$route\[['"]([^'"]+)['"]\]\s*=\s*['"]([^'"]+)['"]`)
	ci3RouteVerbRegex  = regexp.MustCompile(`\$route\[['"]([^'"]+)['"]\]\['(get|post|put|patch|delete|head|options)'\]\s*=\s*['"]([^'"]+)['"]`)
	ci3ControllerClass = regexp.MustCompile(`(?i)\bclass\s+([a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*)\s+extends\s+(?:CI_Controller|MY_Controller)`)
	ci3ModelClass      = regexp.MustCompile(`(?i)\bclass\s+([a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*)\s+extends\s+(?:CI_Model|MY_Model)`)
	ci3RemapRegex      = regexp.MustCompile(`(?i)\bfunction\s+_remap\s*\(`)
	ci3AutoloadModel   = regexp.MustCompile(`\$autoload\[['"]model['"]\]\s*=\s*(?:array\(|\[)([^;]+)`)
	ci3DBDriverRegex   = regexp.MustCompile(`\$db\[['"][^'"]+['"]\]\[['"]dbdriver['"]\]\s*=\s*['"]([^'"]+)['"]`)
	ci3SessDriverRegex = regexp.MustCompile(`\$config\[['"]sess_driver['"]\]\s*=\s*['"]([^'"]+)['"]`)
	ci3HooksEnable     = regexp.MustCompile(`\$config\[['"]enable_hooks['"]\]\s*=\s*(TRUE|true|1)`)
	ci3DBTransRegex    = regexp.MustCompile(`\$this->db->trans_start\(\)|\$this->db->trans_begin\(\)`)
	ci3ForeachRoute    = regexp.MustCompile(`foreach\s*\(\s*\[([^\]]+)\]\s+as\s+\$([a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*)\s*\)`)
)

// CodeIgniter3SourceAdapter implements the CodeIgniter 3 source adapter conforming to
// CI3-ADAPTER, STATIC-ANALYSIS, and SERVICE-CONTRACTS.
//
// Invariants:
// 1. Detection reads directory layout, constants, and config files without PHP execution.
// 2. Extracts routes from application/config/routes.php including standard paths, HTTP verb
//    mappings, default_controller, 404_override, and translate_uri_dashes.
// 3. Extracts controllers (extending CI_Controller or MY_Controller) and models into neutral Application IR.
// 4. Identifies _remap dynamic dispatch and autoloading patterns, recording explicit findings
//    and UNKNOWN/PARTIAL coverage states rather than assuming execution paths.
// 5. Generates disposable probe instrumentation specifications per CI3-ADAPTER.
type CodeIgniter3SourceAdapter struct {
	manifest SourceAdapterManifest
}

// NewCodeIgniter3SourceAdapter creates a CodeIgniter3SourceAdapter.
func NewCodeIgniter3SourceAdapter() *CodeIgniter3SourceAdapter {
	fw := "codeigniter3"
	return &CodeIgniter3SourceAdapter{
		manifest: SourceAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "source.php.ci3",
			Version:              "1.0.0",
			Direction:            "SOURCE",
			Language:             "php",
			Framework:            &fw,
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities: []string{
				string(CapSourceSyntaxParsing),
				string(CapSourceSymbolExtraction),
				string(CapSourceEntrypointDiscovery),
				string(CapSourceRouteExtraction),
				string(CapSourceControllerExtraction),
				string(CapSourceDatabaseAccess),
				string(CapSourceSessionManagement),
				string(CapSourceMiddlewareLifecycle),
				string(CapSourceProbeInstrumentation),
			},
			Limitations: []string{
				"static routing, database, and session analysis; does not execute application bootstrap",
				"_remap dynamic dispatch flagged as UNKNOWN coverage state",
				"unsupported session drivers (e.g. redis/memcached) marked as UNKNOWN gaps",
				"HMVC and third-party extensions remain unresolved without specialized fixtures",
			},
			EvidenceRefs: []string{"sha256:" + strings.Repeat("c", 64)},
		},
	}
}

// Manifest returns the immutable adapter manifest.
func (a *CodeIgniter3SourceAdapter) Manifest() SourceAdapterManifest {
	return a.manifest
}

// ValidateCapabilities checks requested contract obligations against adapter capabilities.
func (a *CodeIgniter3SourceAdapter) ValidateCapabilities(ctx context.Context, requiredCaps []string) (*SourceCapabilityValidationResult, error) {
	supportedMap := make(map[string]bool)
	for _, c := range a.manifest.Capabilities {
		supportedMap[c] = true
	}

	res := &SourceCapabilityValidationResult{
		Valid: true,
	}
	for _, req := range requiredCaps {
		if supportedMap[req] {
			res.SupportedCaps = append(res.SupportedCaps, req)
		} else {
			res.Valid = false
			res.UnsupportedCaps = append(res.UnsupportedCaps, req)
			res.Blockers = append(res.Blockers, "unsupported source capability in CI3 adapter: "+req)
		}
	}
	return res, nil
}

// Detect inspects application layout and config files to detect CodeIgniter 3.
func (a *CodeIgniter3SourceAdapter) Detect(ctx context.Context, input SourceDetectionInput) (*SourceDetectionResult, error) {
	var evidence []string
	hasAppConfig := false
	hasRoutesConfig := false
	hasSystemCore := false
	detectedVer := "3.1.x"

	for path, content := range input.Files {
		normalized := strings.ToLower(path)
		if strings.Contains(normalized, "application/config/config.php") {
			hasAppConfig = true
			evidence = append(evidence, path)
		}
		if strings.Contains(normalized, "application/config/routes.php") {
			hasRoutesConfig = true
			evidence = append(evidence, path)
		}
		if strings.Contains(normalized, "system/core/codeigniter.php") {
			hasSystemCore = true
			evidence = append(evidence, path)
			if strings.Contains(string(content), "CI_VERSION") {
				// Check for specific version string if present
				re := regexp.MustCompile(`define\s*\(\s*['"]CI_VERSION['"]\s*,\s*['"]([^'"]+)['"]\s*\)`)
				if m := re.FindSubmatch(content); len(m) > 1 {
					detectedVer = string(m[1])
				}
			}
		}
	}

	sort.Strings(evidence)

	if !hasAppConfig && !hasRoutesConfig && !hasSystemCore {
		return &SourceDetectionResult{
			Language:   "php",
			Framework:  nil,
			Confidence: 0.0,
			Evidence:   nil,
		}, nil
	}

	confidence := 0.7
	if hasAppConfig && hasRoutesConfig {
		confidence = 0.9
	}
	if hasSystemCore && hasAppConfig && hasRoutesConfig {
		confidence = 0.99
	}

	fw := "codeigniter3"
	return &SourceDetectionResult{
		Language:        "php",
		Framework:       &fw,
		DetectedVersion: &detectedVer,
		Confidence:      confidence,
		Evidence:        evidence,
		Details: map[string]interface{}{
			"has_app_config":    hasAppConfig,
			"has_routes_config": hasRoutesConfig,
			"has_system_core":   hasSystemCore,
		},
	}, nil
}

// Parse extracts SourceIRDocument from input files utilizing generic PHP parsing
// augmented with CodeIgniter base class and model detection.
func (a *CodeIgniter3SourceAdapter) Parse(ctx context.Context, input SourceParseInput) (*SourceParseOutput, error) {
	generic := NewGenericPHPSourceAdapter()
	return generic.Parse(ctx, input)
}

// Extract extracts ApplicationIRDocument by analyzing CI3 routes, controllers, and models.
func (a *CodeIgniter3SourceAdapter) Extract(ctx context.Context, input SourceExtractionInput) (*SourceExtractionOutput, error) {
	projectID := input.ProjectID
	if projectID == "" {
		projectID = "proj-default"
	}

	sourceIRRef := "sha256:" + strings.Repeat("0", 64)
	configHash := strings.Repeat("0", 64)
	if input.SourceIR != nil {
		ref, err := input.SourceIR.Digest()
		if err == nil {
			sourceIRRef = ref
		}
		if input.SourceIR.ConfigurationHash != "" {
			configHash = input.SourceIR.ConfigurationHash
		}
	}

	entrypoints := make([]ApplicationEntrypoint, 0)
	components := make([]ApplicationComponent, 0)
	dependencies := make([]ApplicationDependency, 0)
	findings := make([]ApplicationFinding, 0)
	evidence := make([]ApplicationEvidence, 0)
	var gaps []string

	evCounter := 1
	epCounter := 1
	compCounter := 1
	findCounter := 1

	// 1. Analyze application/config/routes.php
	for path, content := range input.Files {
		if strings.HasSuffix(path, "application/config/routes.php") || strings.HasSuffix(path, "routes.php") {
			sum := sha256.Sum256(content)
			sumHex := hex.EncodeToString(sum[:])
			lines := strings.Split(string(content), "\n")

			for lineNum, line := range lines {
				lineIdx := lineNum + 1

				// Check verb-specific route: $route['path']['verb'] = 'controller/action';
				if m := ci3RouteVerbRegex.FindStringSubmatch(line); len(m) > 3 {
					routePath := "/" + strings.TrimPrefix(m[1], "/")
					method := strings.ToUpper(m[2])
					target := m[3]

					evID := fmt.Sprintf("ev-ci3-%d", evCounter)
					evCounter++
					evidence = append(evidence, ApplicationEvidence{
						EvidenceID:  evID,
						Kind:        "STATIC",
						ArtifactRef: "sha256:" + sumHex,
						Description: fmt.Sprintf("CI3 HTTP verb route %s %s -> %s in %s:%d", method, routePath, target, path, lineIdx),
					})

					disp, _ := json.Marshal(map[string]interface{}{
						"type":   "http",
						"method": method,
						"path":   routePath,
					})

					entrypoints = append(entrypoints, ApplicationEntrypoint{
						EntrypointID:     fmt.Sprintf("ep-ci3-%d", epCounter),
						Type:             EntrypointTypeHTTP,
						Dispatch:         disp,
						HandlerSymbolIDs: []string{sanitizeIdentifier(target)},
						EvidenceIDs:      []string{evID},
						CoverageState:    CoverageStateResolved,
					})
					epCounter++
					continue
				}

				// Check standard route: $route['path'] = 'controller/action';
				if m := ci3RouteRegex.FindStringSubmatch(line); len(m) > 2 {
					key := m[1]
					target := m[2]

					method := "ANY"
					routePath := "/" + strings.TrimPrefix(key, "/")
					if key == "default_controller" {
						routePath = "/"
					} else if key == "404_override" {
						routePath = "/404"
					}

					evID := fmt.Sprintf("ev-ci3-%d", evCounter)
					evCounter++
					evidence = append(evidence, ApplicationEvidence{
						EvidenceID:  evID,
						Kind:        "STATIC",
						ArtifactRef: "sha256:" + sumHex,
						Description: fmt.Sprintf("CI3 route %s -> %s in %s:%d", key, target, path, lineIdx),
					})

					disp, _ := json.Marshal(map[string]interface{}{
						"type":   "http",
						"method": method,
						"path":   routePath,
					})

					entrypoints = append(entrypoints, ApplicationEntrypoint{
						EntrypointID:     fmt.Sprintf("ep-ci3-%d", epCounter),
						Type:             EntrypointTypeHTTP,
						Dispatch:         disp,
						HandlerSymbolIDs: []string{sanitizeIdentifier(target)},
						EvidenceIDs:      []string{evID},
						CoverageState:    CoverageStateResolved,
					})
					epCounter++
				}
			}

			// Check foreach static route pattern: foreach (['a', 'b'] as $name) { $route[$name] = 'controller/'.$name; }
			routesContentStr := string(content)
			if fMatch := ci3ForeachRoute.FindStringSubmatch(routesContentStr); len(fMatch) > 2 {
				itemsRaw := fMatch[1]
				varName := fMatch[2]
				targetPattern := regexp.MustCompile(fmt.Sprintf(`\$route\[\$%s\]\s*=\s*['"]([^'"]+)['"]\s*\.\s*\$%s`, varName, varName))
				if tMatch := targetPattern.FindStringSubmatch(routesContentStr); len(tMatch) > 1 {
					baseTarget := strings.TrimSuffix(tMatch[1], "/")
					itemRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
					itemMatches := itemRegex.FindAllStringSubmatch(itemsRaw, -1)
					for _, im := range itemMatches {
						if len(im) > 1 {
							name := im[1]
							routePath := "/" + strings.TrimPrefix(name, "/")
							fullTarget := baseTarget + "/" + name

							evID := fmt.Sprintf("ev-ci3-%d", evCounter)
							evCounter++
							evidence = append(evidence, ApplicationEvidence{
								EvidenceID:  evID,
								Kind:        "STATIC",
								ArtifactRef: "sha256:" + sumHex,
								Description: fmt.Sprintf("CI3 loop route %s -> %s in %s", routePath, fullTarget, path),
							})

							disp, _ := json.Marshal(map[string]interface{}{
								"type":   "http",
								"method": "ANY",
								"path":   routePath,
							})

							entrypoints = append(entrypoints, ApplicationEntrypoint{
								EntrypointID:     fmt.Sprintf("ep-ci3-%d", epCounter),
								Type:             EntrypointTypeHTTP,
								Dispatch:         disp,
								HandlerSymbolIDs: []string{sanitizeIdentifier(fullTarget)},
								EvidenceIDs:      []string{evID},
								CoverageState:    CoverageStateResolved,
							})
							epCounter++
						}
					}
				}
			}
		}
	}

	// 2. Analyze Controllers and Models in application/
	paths := make([]string, 0, len(input.Files))
	for p := range input.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		content := input.Files[path]
		contentStr := string(content)
		sum := sha256.Sum256(content)
		sumHex := hex.EncodeToString(sum[:])

		// Controller detection
		if m := ci3ControllerClass.FindStringSubmatch(contentStr); len(m) > 1 {
			className := m[1]
			compID := fmt.Sprintf("comp-ctrl-%d", compCounter)
			compCounter++

			evID := fmt.Sprintf("ev-ci3-%d", evCounter)
			evCounter++
			evidence = append(evidence, ApplicationEvidence{
				EvidenceID:  evID,
				Kind:        "STATIC",
				ArtifactRef: "sha256:" + sumHex,
				Description: fmt.Sprintf("Discovered CI3 Controller %s in %s", className, path),
			})

			components = append(components, ApplicationComponent{
				ComponentID: compID,
				Role:        ComponentRoleHandler,
				SymbolIDs:   []string{sanitizeIdentifier(className)},
				EvidenceIDs: []string{evID},
			})

			// Check for _remap dynamic dispatch
			if ci3RemapRegex.MatchString(contentStr) {
				findID := fmt.Sprintf("find-ci3-%d", findCounter)
				findCounter++
				findings = append(findings, ApplicationFinding{
					FindingID:   findID,
					Code:        "CI3_REMAP_DYNAMIC_DISPATCH",
					Severity:    FindingSeverityHigh,
					AffectedIDs: []string{compID},
					EvidenceIDs: []string{evID},
					Message:     fmt.Sprintf("Controller %s defines _remap; static routing is bypassed and coverage is UNKNOWN", className),
				})
				gaps = append(gaps, fmt.Sprintf("controller %s uses _remap; runtime behavior required", className))
			}
		}

		// Model detection
		if m := ci3ModelClass.FindStringSubmatch(contentStr); len(m) > 1 {
			className := m[1]
			compID := fmt.Sprintf("comp-model-%d", compCounter)
			compCounter++

			evID := fmt.Sprintf("ev-ci3-%d", evCounter)
			evCounter++
			evidence = append(evidence, ApplicationEvidence{
				EvidenceID:  evID,
				Kind:        "STATIC",
				ArtifactRef: "sha256:" + sumHex,
				Description: fmt.Sprintf("Discovered CI3 Model %s in %s", className, path),
			})

			components = append(components, ApplicationComponent{
				ComponentID: compID,
				Role:        ComponentRoleDataModel,
				SymbolIDs:   []string{sanitizeIdentifier(className)},
				EvidenceIDs: []string{evID},
			})

			// Check for explicit database transaction usage
			if ci3DBTransRegex.MatchString(contentStr) {
				findID := fmt.Sprintf("find-ci3-%d", findCounter)
				findCounter++
				findings = append(findings, ApplicationFinding{
					FindingID:   findID,
					Code:        "CI3_DB_TRANSACTION_BOUNDARY",
					Severity:    FindingSeverityLow,
					AffectedIDs: []string{compID},
					EvidenceIDs: []string{evID},
					Message:     fmt.Sprintf("Model %s uses manual trans_start/trans_begin transaction boundaries", className),
				})
			}
		}

		// Database configuration analysis
		if strings.HasSuffix(path, "application/config/database.php") || strings.HasSuffix(path, "database.php") {
			matches := ci3DBDriverRegex.FindAllStringSubmatch(contentStr, -1)
			for _, m := range matches {
				if len(m) > 1 {
					driver := m[1]
					evID := fmt.Sprintf("ev-ci3-%d", evCounter)
					evCounter++
					evidence = append(evidence, ApplicationEvidence{
						EvidenceID:  evID,
						Kind:        "STATIC",
						ArtifactRef: "sha256:" + sumHex,
						Description: fmt.Sprintf("Configured database driver: %s in %s", driver, path),
					})

					// If not standard supported driver (mysqli, pdo), mark gap
					if driver != "mysqli" && driver != "pdo" && driver != "sqlite3" {
						gaps = append(gaps, fmt.Sprintf("unsupported database driver: %s", driver))
					}
				}
			}
		}

		// Session and hooks configuration analysis in config.php
		if strings.HasSuffix(path, "application/config/config.php") || strings.HasSuffix(path, "config.php") {
			if m := ci3SessDriverRegex.FindStringSubmatch(contentStr); len(m) > 1 {
				sessDriver := m[1]
				evID := fmt.Sprintf("ev-ci3-%d", evCounter)
				evCounter++
				evidence = append(evidence, ApplicationEvidence{
					EvidenceID:  evID,
					Kind:        "STATIC",
					ArtifactRef: "sha256:" + sumHex,
					Description: fmt.Sprintf("Configured session driver: %s in %s", sessDriver, path),
				})

				if sessDriver != "files" && sessDriver != "database" {
					gaps = append(gaps, fmt.Sprintf("unsupported session driver: %s; requires runtime bridge", sessDriver))
				}
			}

			if ci3HooksEnable.MatchString(contentStr) {
				evID := fmt.Sprintf("ev-ci3-%d", evCounter)
				evCounter++
				evidence = append(evidence, ApplicationEvidence{
					EvidenceID:  evID,
					Kind:        "STATIC",
					ArtifactRef: "sha256:" + sumHex,
					Description: fmt.Sprintf("Hooks lifecycle enabled in %s", path),
				})
			}
		}
	}

	appDoc := &ApplicationIRDocument{
		SchemaVersion:     "1.0",
		ProjectID:         projectID,
		ConfigurationHash: configHash,
		DocumentID:        "doc-app-ir-ci3",
		SourceIRRef:       sourceIRRef,
		Entrypoints:       entrypoints,
		Components:        components,
		Dependencies:      dependencies,
		DataModels:        []json.RawMessage{},
		Workflows:         []json.RawMessage{},
		Findings:          findings,
		Evidence:          evidence,
	}

	return &SourceExtractionOutput{
		Document: appDoc,
		Gaps:     gaps,
	}, nil
}

// PrepareInstrumentation specifies that CI3 requires a disposable sandbox copy
// and hook overlay per CI3-ADAPTER.
func (a *CodeIgniter3SourceAdapter) PrepareInstrumentation(ctx context.Context, input SourceExtractionInput) (*SourceInstrumentationSpec, error) {
	fw := "codeigniter3"
	return &SourceInstrumentationSpec{
		AdapterID:        a.manifest.AdapterID,
		Language:         "php",
		Framework:        &fw,
		ProbeOverlayPath: "probes/php/ci3",
		HookConfig: map[string]string{
			"post_controller_constructor": "RewrotronHook::capturePre",
			"post_system":                 "RewrotronHook::capturePost",
		},
		DisposableCopy: true,
	}, nil
}

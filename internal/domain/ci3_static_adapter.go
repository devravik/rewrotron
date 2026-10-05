package domain

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// CI3StaticAdapterError describes an error during static parsing and entrypoint extraction.
type CI3StaticAdapterError struct {
	Reason string
	Cause  error
}

// Error identifies the failed extraction or parsing requirement.
func (e *CI3StaticAdapterError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("ci3 static adapter error: %s: %v", e.Reason, e.Cause)
	}
	return "ci3 static adapter error: " + e.Reason
}

// Unwrap preserves the underlying decoding or parsing error.
func (e *CI3StaticAdapterError) Unwrap() error { return e.Cause }

// CI3AdapterInput provides the file content and environment context to the CI3 static adapter.
type CI3AdapterInput struct {
	ProjectID           string            `json:"project_id"`
	ConfigurationHash   string            `json:"configuration_hash"`
	DocumentID          string            `json:"document_id"`
	Files               map[string][]byte `json:"files"`
	KnownVendorHashes   map[string]string `json:"known_vendor_hashes,omitempty"`
	DeclaredEnvironment string            `json:"declared_environment,omitempty"`
	OperationalRoots    []string          `json:"operational_roots,omitempty"`
}

// CI3AdapterResult contains the generated Source IR and Application IR documents, plus detection facts.
type CI3AdapterResult struct {
	SourceIR                *SourceIRDocument      `json:"source_ir"`
	ApplicationIR           *ApplicationIRDocument `json:"application_ir"`
	ModifiedVendorFiles     []string               `json:"modified_vendor_files"`
	DynamicDispatchDetected bool                   `json:"dynamic_dispatch_detected"`
	RemapControllers        []string               `json:"remap_controllers"`
}

var (
	classRegex    = regexp.MustCompile(`(?i)\bclass\s+([a-zA-Z0-9_]+)(?:\s+extends\s+([a-zA-Z0-9_]+))?(?:\s+implements\s+([a-zA-Z0-9_,\s]+))?`)
	methodRegex   = regexp.MustCompile(`(?i)(?:(public|protected|private)\s+)?function\s+([a-zA-Z0-9_]+)\s*\(([^)]*)\)`)
	propRegex     = regexp.MustCompile(`(?i)(?:(public|protected|private|var)\s+)\$([a-zA-Z0-9_]+)`)
	constRegex    = regexp.MustCompile(`(?i)\bconst\s+([a-zA-Z0-9_]+)\s*=`)
	defineRegex   = regexp.MustCompile(`(?i)\bdefine\s*\(\s*['"]([a-zA-Z0-9_]+)['"]\s*,\s*['"]?([^'")]+)['"]?\)`)
	loadModelRe   = regexp.MustCompile(`\$this->load->model\s*\(\s*['"]([a-zA-Z0-9_/]+)['"]`)
	loadViewRe    = regexp.MustCompile(`\$this->load->view\s*\(\s*['"]([a-zA-Z0-9_/]+)['"]`)
	dynDispatchRe = regexp.MustCompile(`(?:\$this->\$([a-zA-Z0-9_]+)|\$([a-zA-Z0-9_]+)->\$([a-zA-Z0-9_]+)|\bcall_user_func(?:_array)?\s*\(|\$\$[a-zA-Z0-9_]+)`)
	dynIncludeRe  = regexp.MustCompile(`\b(?:include|require|include_once|require_once)\s*\(?\s*\$`)
	reflectionRe  = regexp.MustCompile(`\bnew\s+Reflection(?:Class|Method|Function|Property)\s*\(`)
	evalRe        = regexp.MustCompile(`\beval\s*\(`)
	envDefineRe   = regexp.MustCompile(`(?i)define\s*\(\s*['"]ENVIRONMENT['"]\s*,\s*['"]([a-zA-Z0-9_]+)['"]`)
	isCliRe       = regexp.MustCompile(`\$this->input->is_cli_request\s*\(`)
	routeExactRe  = regexp.MustCompile(`\$route\[['"]([^'"]+)['"]\]\s*=\s*['"]([^'"]+)['"]`)
	routeVerbRe   = regexp.MustCompile(`\$route\[['"]([^'"]+)['"]\]\[['"](get|post|put|delete|patch|head|options)['"]\]\s*=\s*['"]([^'"]+)['"]`)
)

// ExtractCI3 analyzes a CodeIgniter 3 source repository, extracting AST symbols, route configurations,
// dynamic dispatch flags, modified vendor files, and emitting validated Source IR and Application IR documents.
func ExtractCI3(input *CI3AdapterInput) (*CI3AdapterResult, error) {
	if input == nil {
		return nil, &CI3StaticAdapterError{Reason: "input must not be nil"}
	}
	if !validAssessmentReportID(input.ProjectID) {
		return nil, &CI3StaticAdapterError{Reason: "project_id must match pattern ^[a-z0-9_-]+$"}
	}
	if !isManifestDigest(input.ConfigurationHash) {
		return nil, &CI3StaticAdapterError{Reason: "configuration_hash must match pattern ^[a-f0-9]{64}$"}
	}
	if !validAssessmentReportID(input.DocumentID) {
		return nil, &CI3StaticAdapterError{Reason: "document_id must match pattern ^[a-z0-9_-]+$"}
	}

	result := &CI3AdapterResult{
		ModifiedVendorFiles: []string{},
		RemapControllers:    []string{},
	}

	// Canonical static evidence reference
	primaryEvidenceID := "ev_ci3_static_analysis"
	primaryEvidenceRef := "sha256:" + strings.Repeat("a", 64)

	files := make([]SourceFile, 0)
	symbols := make([]SourceSymbol, 0)
	edges := make([]SourceEdge, 0)
	findings := make([]SourceFinding, 0)
	evidence := make([]SourceEvidence, 0)

	evidence = append(evidence, SourceEvidence{
		EvidenceID:  primaryEvidenceID,
		Kind:        "STATIC",
		ArtifactRef: primaryEvidenceRef,
		Description: "CodeIgniter 3 static parsing, route discovery, and AST extraction",
	})

	// Sort file paths for determinism
	var sortedPaths []string
	for p := range input.Files {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	// Maps for cross-referencing
	pathToID := make(map[string]string)
	controllerMethods := make(map[string][]string) // controllerName -> list of method names
	controllerHasRemap := make(map[string]bool)
	controllerSymbols := make(map[string]string)   // controllerName -> symbolID
	methodSymbols := make(map[string]string)       // controllerName.methodName -> symbolID
	fileSymbols := make(map[string][]string)       // fileID -> list of symbolIDs

	edgeCounter := 0
	findingCounter := 0

	// 1. Process files and detect vendor modifications
	for _, path := range sortedPaths {
		content := input.Files[path]
		fileID := sanitizeID("file_" + path)
		pathToID[path] = fileID

		shaBytes := sha256.Sum256(content)
		fileSHA := hex.EncodeToString(shaBytes[:])

		syntaxStatus := SyntaxStatusParsed
		if !utf8.Valid(content) {
			syntaxStatus = SyntaxStatusFailed
		} else if bytes.Contains(content, []byte("PHP Parse error")) || bytes.Contains(content, []byte("Parse error:")) {
			syntaxStatus = SyntaxStatusFailed
		}

		files = append(files, SourceFile{
			FileID:       fileID,
			Path:         path,
			SHA256:       fileSHA,
			SyntaxStatus: syntaxStatus,
		})

		// Test X10: Vendor / Framework patch detection
		if strings.HasPrefix(path, "system/") || strings.HasPrefix(path, "vendor/") {
			if expectedSHA, ok := input.KnownVendorHashes[path]; ok {
				if fileSHA != expectedSHA {
					result.ModifiedVendorFiles = append(result.ModifiedVendorFiles, path)
					findingCounter++
					findings = append(findings, SourceFinding{
						FindingID:   fmt.Sprintf("find_mod_vendor_%d", findingCounter),
						Code:        "MODIFIED_VENDOR",
						Severity:    FindingSeverityHigh,
						AffectedIDs: []string{fileID},
						EvidenceIDs: []string{primaryEvidenceID},
						Message:     fmt.Sprintf("Vendor framework file %q hash %s differs from expected canonical hash %s", path, fileSHA, expectedSHA),
					})
				}
			}
		}

		// Check environment definition in index.php
		if path == "index.php" || strings.HasSuffix(path, "/index.php") {
			if match := envDefineRe.FindSubmatch(content); len(match) > 1 {
				sourceEnv := string(match[1])
				if input.DeclaredEnvironment != "" && !strings.EqualFold(sourceEnv, input.DeclaredEnvironment) {
					findingCounter++
					findings = append(findings, SourceFinding{
						FindingID:   fmt.Sprintf("find_env_mismatch_%d", findingCounter),
						Code:        "ENVIRONMENT_MISMATCH",
						Severity:    FindingSeverityHigh,
						AffectedIDs: []string{fileID},
						EvidenceIDs: []string{primaryEvidenceID},
						Message:     fmt.Sprintf("Declared environment %q contradicts source constant ENVIRONMENT=%q", input.DeclaredEnvironment, sourceEnv),
					})
				}
			}
		}

		if syntaxStatus == SyntaxStatusFailed {
			findingCounter++
			findings = append(findings, SourceFinding{
				FindingID:   fmt.Sprintf("find_syntax_err_%d", findingCounter),
				Code:        "SYNTAX_PARSE_FAILED",
				Severity:    FindingSeverityCritical,
				AffectedIDs: []string{fileID},
				EvidenceIDs: []string{primaryEvidenceID},
				Message:     fmt.Sprintf("File %q contains invalid syntax or non-UTF-8 content", path),
			})
			continue
		}

		// 2. Parse PHP declarations (classes, methods, functions, properties)
		if strings.HasSuffix(path, ".php") {
			scanner := bufio.NewScanner(bytes.NewReader(content))
			lineNum := 0
			var currentClass string
			var currentClassSymID string

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()

				// Detect class
				if match := classRegex.FindStringSubmatch(line); len(match) > 1 {
					className := match[1]
					currentClass = className
					symID := sanitizeID(fmt.Sprintf("sym_class_%s", className))
					currentClassSymID = symID
					controllerSymbols[className] = symID

					symbols = append(symbols, SourceSymbol{
						SymbolID:  symID,
						FileID:    fileID,
						Name:      className,
						Kind:      SymbolKindClass,
						StartLine: lineNum,
						EndLine:   lineNum + 50, // bounded estimate
					})
					fileSymbols[fileID] = append(fileSymbols[fileID], symID)
				}

				// Detect methods / functions
				if match := methodRegex.FindStringSubmatch(line); len(match) > 2 {
					methodName := match[2]
					kind := SymbolKindFunction
					if currentClass != "" {
						kind = SymbolKindMethod
						controllerMethods[currentClass] = append(controllerMethods[currentClass], methodName)
						if methodName == "_remap" {
							controllerHasRemap[currentClass] = true
							result.RemapControllers = append(result.RemapControllers, currentClass)
							findingCounter++
							findings = append(findings, SourceFinding{
								FindingID:   fmt.Sprintf("find_remap_%d", findingCounter),
								Code:        "DYNAMIC_DISPATCH_REMAP",
								Severity:    FindingSeverityMedium,
								AffectedIDs: []string{currentClassSymID},
								EvidenceIDs: []string{primaryEvidenceID},
								Message:     fmt.Sprintf("Controller %q defines custom _remap method; static dispatch is non-standard", currentClass),
							})
						}
					}

					symID := sanitizeID(fmt.Sprintf("sym_fn_%s_%s_%d", currentClass, methodName, lineNum))
					if currentClass != "" {
						methodSymbols[currentClass+"."+methodName] = symID
					}

					symbols = append(symbols, SourceSymbol{
						SymbolID:  symID,
						FileID:    fileID,
						Name:      methodName,
						Kind:      kind,
						StartLine: lineNum,
						EndLine:   lineNum + 10,
					})
					fileSymbols[fileID] = append(fileSymbols[fileID], symID)

					if currentClassSymID != "" {
						edgeCounter++
						edges = append(edges, SourceEdge{
							EdgeID:      fmt.Sprintf("edge_decl_%d", edgeCounter),
							FromID:      currentClassSymID,
							ToID:        &symID,
							Kind:        "declares",
							Resolved:    true,
							EvidenceIDs: []string{primaryEvidenceID},
						})
					}
				}

				// Detect properties
				if match := propRegex.FindStringSubmatch(line); len(match) > 2 && currentClassSymID != "" {
					propName := match[2]
					symID := sanitizeID(fmt.Sprintf("sym_prop_%s_%s_%d", currentClass, propName, lineNum))
					symbols = append(symbols, SourceSymbol{
						SymbolID:  symID,
						FileID:    fileID,
						Name:      propName,
						Kind:      SymbolKindProperty,
						StartLine: lineNum,
						EndLine:   lineNum,
					})
					fileSymbols[fileID] = append(fileSymbols[fileID], symID)
				}

				// Detect constants
				if match := constRegex.FindStringSubmatch(line); len(match) > 1 {
					constName := match[1]
					symID := sanitizeID(fmt.Sprintf("sym_const_%s_%d", constName, lineNum))
					symbols = append(symbols, SourceSymbol{
						SymbolID:  symID,
						FileID:    fileID,
						Name:      constName,
						Kind:      SymbolKindConstant,
						StartLine: lineNum,
						EndLine:   lineNum,
					})
					fileSymbols[fileID] = append(fileSymbols[fileID], symID)
				}

				// Test X09: Detect Dynamic Dispatch and Includes
				if dynDispatchRe.MatchString(line) {
					result.DynamicDispatchDetected = true
					edgeCounter++
					from := currentClassSymID
					if from == "" {
						from = fileID
					}
					expr := strings.TrimSpace(line)
					edges = append(edges, SourceEdge{
						EdgeID:      fmt.Sprintf("edge_dyn_%d", edgeCounter),
						FromID:      from,
						ToID:        nil,
						Kind:        "dynamic_dispatch",
						Resolved:    false,
						EvidenceIDs: []string{primaryEvidenceID},
						Expression:  &expr,
					})

					findingCounter++
					findings = append(findings, SourceFinding{
						FindingID:   fmt.Sprintf("find_dyn_%d", findingCounter),
						Code:        "DYNAMIC_DISPATCH",
						Severity:    FindingSeverityMedium,
						AffectedIDs: []string{from},
						EvidenceIDs: []string{primaryEvidenceID},
						Message:     fmt.Sprintf("Dynamic dispatch expression detected at line %d: %s", lineNum, expr),
					})
				}

				if dynIncludeRe.MatchString(line) {
					result.DynamicDispatchDetected = true
					edgeCounter++
					from := currentClassSymID
					if from == "" {
						from = fileID
					}
					expr := strings.TrimSpace(line)
					edges = append(edges, SourceEdge{
						EdgeID:      fmt.Sprintf("edge_inc_%d", edgeCounter),
						FromID:      from,
						ToID:        nil,
						Kind:        "dynamic_include",
						Resolved:    false,
						EvidenceIDs: []string{primaryEvidenceID},
						Expression:  &expr,
					})

					findingCounter++
					findings = append(findings, SourceFinding{
						FindingID:   fmt.Sprintf("find_inc_%d", findingCounter),
						Code:        "DYNAMIC_INCLUDE",
						Severity:    FindingSeverityHigh,
						AffectedIDs: []string{from},
						EvidenceIDs: []string{primaryEvidenceID},
						Message:     fmt.Sprintf("Dynamic include/require detected at line %d: %s", lineNum, expr),
					})
				}

				if reflectionRe.MatchString(line) {
					result.DynamicDispatchDetected = true
					findingCounter++
					from := currentClassSymID
					if from == "" {
						from = fileID
					}
					findings = append(findings, SourceFinding{
						FindingID:   fmt.Sprintf("find_reflect_%d", findingCounter),
						Code:        "REFLECTION_USAGE",
						Severity:    FindingSeverityMedium,
						AffectedIDs: []string{from},
						EvidenceIDs: []string{primaryEvidenceID},
						Message:     fmt.Sprintf("Reflection class/method instantiation detected at line %d", lineNum),
					})
				}

				if evalRe.MatchString(line) {
					result.DynamicDispatchDetected = true
					findingCounter++
					from := currentClassSymID
					if from == "" {
						from = fileID
					}
					findings = append(findings, SourceFinding{
						FindingID:   fmt.Sprintf("find_eval_%d", findingCounter),
						Code:        "UNSUPPORTED_EVAL",
						Severity:    FindingSeverityCritical,
						AffectedIDs: []string{from},
						EvidenceIDs: []string{primaryEvidenceID},
						Message:     fmt.Sprintf("Dynamic eval execution detected at line %d", lineNum),
					})
				}
			}
		}
	}

	// 3. Extract Routes and Entrypoints from routes.php
	entrypoints := make([]ApplicationEntrypoint, 0)
	epCounter := 0

	routesFileContent, hasRoutes := input.Files["application/config/routes.php"]
	if hasRoutes {
		scanner := bufio.NewScanner(bytes.NewReader(routesFileContent))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
				continue
			}

			// Verb-specific routes: $route['path']['get'] = 'controller/method'
			if match := routeVerbRe.FindStringSubmatch(line); len(match) > 3 {
				routePath := match[1]
				verb := strings.ToUpper(match[2])
				target := match[3]
				epCounter++
				epID := fmt.Sprintf("ep_http_%s_%s_%d", strings.ToLower(verb), sanitizeID(routePath), epCounter)

				dispatchJSON, _ := json.Marshal(map[string]any{
					"type":   "http",
					"method": verb,
					"path":   "/" + strings.TrimPrefix(routePath, "/"),
				})

				handlerSyms := resolveHandlerSymbols(target, controllerSymbols, methodSymbols)
				covState := CoverageStateResolved
				parts := strings.Split(target, "/")
				if len(parts) > 0 && controllerHasRemap[parts[0]] {
					covState = CoverageStatePartial
				}

				entrypoints = append(entrypoints, ApplicationEntrypoint{
					EntrypointID:     epID,
					Type:             EntrypointTypeHTTP,
					Dispatch:         dispatchJSON,
					HandlerSymbolIDs: handlerSyms,
					EvidenceIDs:      []string{primaryEvidenceID},
					CoverageState:    covState,
				})
			} else if match := routeExactRe.FindStringSubmatch(line); len(match) > 2 {
				// Exact or pattern routes: $route['path'] = 'controller/method'
				routePath := match[1]
				target := match[2]

				// Skip special CI3 control keys for HTTP endpoints
				if routePath == "default_controller" || routePath == "404_override" || routePath == "translate_uri_dashes" {
					continue
				}

				epCounter++
				epID := fmt.Sprintf("ep_http_any_%s_%d", sanitizeID(routePath), epCounter)

				dispatchJSON, _ := json.Marshal(map[string]any{
					"type":   "http",
					"method": "ANY",
					"path":   "/" + strings.TrimPrefix(routePath, "/"),
				})

				handlerSyms := resolveHandlerSymbols(target, controllerSymbols, methodSymbols)
				covState := CoverageStateResolved
				parts := strings.Split(target, "/")
				if len(parts) > 0 && controllerHasRemap[parts[0]] {
					covState = CoverageStatePartial
				}

				entrypoints = append(entrypoints, ApplicationEntrypoint{
					EntrypointID:     epID,
					Type:             EntrypointTypeHTTP,
					Dispatch:         dispatchJSON,
					HandlerSymbolIDs: handlerSyms,
					EvidenceIDs:      []string{primaryEvidenceID},
					CoverageState:    covState,
				})
			}
		}
	}

	// 4. Discover Controller Convention Routes
	for ctrlName, methods := range controllerMethods {
		// Controllers in application/controllers/
		ctrlSymID, hasSym := controllerSymbols[ctrlName]
		if !hasSym {
			continue
		}

		for _, mName := range methods {
			if strings.HasPrefix(mName, "_") || mName == "__construct" || mName == "get_instance" {
				continue
			}

			path := fmt.Sprintf("/%s/%s", strings.ToLower(ctrlName), strings.ToLower(mName))
			if mName == "index" {
				path = fmt.Sprintf("/%s", strings.ToLower(ctrlName))
			}

			epCounter++
			epID := fmt.Sprintf("ep_http_conv_%s_%s_%d", strings.ToLower(ctrlName), strings.ToLower(mName), epCounter)

			dispatchJSON, _ := json.Marshal(map[string]any{
				"type":   "http",
				"method": "ANY",
				"path":   path,
			})

			handlerSym := methodSymbols[ctrlName+"."+mName]
			var handlers []string
			if handlerSym != "" {
				handlers = append(handlers, handlerSym)
			} else {
				handlers = append(handlers, ctrlSymID)
			}

			covState := CoverageStateResolved
			if controllerHasRemap[ctrlName] {
				covState = CoverageStatePartial
			}

			entrypoints = append(entrypoints, ApplicationEntrypoint{
				EntrypointID:     epID,
				Type:             EntrypointTypeHTTP,
				Dispatch:         dispatchJSON,
				HandlerSymbolIDs: handlers,
				EvidenceIDs:      []string{primaryEvidenceID},
				CoverageState:    covState,
			})
		}
	}

	// 5. Discover CLI and Cron Operational Roots
	for _, opRoot := range input.OperationalRoots {
		epCounter++
		if strings.HasPrefix(opRoot, "cron:") {
			cmd := strings.TrimPrefix(opRoot, "cron:")
			epID := fmt.Sprintf("ep_cron_%s_%d", sanitizeID(cmd), epCounter)
			disp, _ := json.Marshal(map[string]any{
				"type":       "cron",
				"command_id": sanitizeID(cmd),
				"schedule":   "0 0 * * *",
				"timezone":   "UTC",
			})
			entrypoints = append(entrypoints, ApplicationEntrypoint{
				EntrypointID:     epID,
				Type:             EntrypointTypeCron,
				Dispatch:         disp,
				HandlerSymbolIDs: []string{},
				EvidenceIDs:      []string{primaryEvidenceID},
				CoverageState:    CoverageStateResolved,
			})
		} else {
			cmd := opRoot
			epID := fmt.Sprintf("ep_cli_%s_%d", sanitizeID(cmd), epCounter)
			disp, _ := json.Marshal(map[string]any{
				"type":       "cli",
				"command_id": sanitizeID(cmd),
			})
			entrypoints = append(entrypoints, ApplicationEntrypoint{
				EntrypointID:     epID,
				Type:             EntrypointTypeCLI,
				Dispatch:         disp,
				HandlerSymbolIDs: []string{},
				EvidenceIDs:      []string{primaryEvidenceID},
				CoverageState:    CoverageStateResolved,
			})
		}
	}

	// If no entrypoints were discovered at all, supply at least one fallback root
	if len(entrypoints) == 0 {
		epCounter++
		epID := fmt.Sprintf("ep_http_fallback_%d", epCounter)
		disp, _ := json.Marshal(map[string]any{
			"type":   "http",
			"method": "GET",
			"path":   "/",
		})
		entrypoints = append(entrypoints, ApplicationEntrypoint{
			EntrypointID:     epID,
			Type:             EntrypointTypeHTTP,
			Dispatch:         disp,
			HandlerSymbolIDs: []string{},
			EvidenceIDs:      []string{primaryEvidenceID},
			CoverageState:    CoverageStateUnknown,
		})
	}

	// 6. Build Components
	var components []ApplicationComponent
	compCounter := 0
	for ctrlName, symID := range controllerSymbols {
		compCounter++
		compID := fmt.Sprintf("comp_ctrl_%s_%d", sanitizeID(ctrlName), compCounter)
		var syms []string
		syms = append(syms, symID)
		for _, mName := range controllerMethods[ctrlName] {
			if mSym, ok := methodSymbols[ctrlName+"."+mName]; ok {
				syms = append(syms, mSym)
			}
		}

		components = append(components, ApplicationComponent{
			ComponentID: compID,
			Role:        ComponentRoleHandler,
			SymbolIDs:   syms,
			EvidenceIDs: []string{primaryEvidenceID},
		})
	}

	if len(components) == 0 {
		compCounter++
		components = append(components, ApplicationComponent{
			ComponentID: fmt.Sprintf("comp_root_%d", compCounter),
			Role:        ComponentRoleHandler,
			SymbolIDs:   []string{},
			EvidenceIDs: []string{primaryEvidenceID},
		})
	}

	// 7. Build Dependencies
	dependencies := make([]ApplicationDependency, 0)
	depCounter := 0
	for _, edge := range edges {
		depCounter++
		dependencies = append(dependencies, ApplicationDependency{
			EdgeID:      fmt.Sprintf("dep_%d", depCounter),
			FromID:      edge.FromID,
			ToID:        edge.ToID,
			Kind:        edge.Kind,
			Resolved:    edge.Resolved,
			EvidenceIDs: edge.EvidenceIDs,
			Expression:  edge.Expression,
		})
	}

	// 8. Construct Source IR Document
	sourceIR := &SourceIRDocument{
		SchemaVersion:     "1.0",
		ProjectID:         input.ProjectID,
		ConfigurationHash: input.ConfigurationHash,
		DocumentID:        input.DocumentID,
		Files:             files,
		Symbols:           symbols,
		Edges:             edges,
		Findings:          findings,
		Evidence:          evidence,
	}

	sourceIRDigest, err := sourceIR.Digest()
	if err != nil {
		return nil, &CI3StaticAdapterError{Reason: "failed to compute Source IR digest", Cause: err}
	}

	// Map findings and evidence for Application IR
	appFindings := make([]ApplicationFinding, 0)
	for _, f := range findings {
		appFindings = append(appFindings, ApplicationFinding{
			FindingID:   f.FindingID,
			Code:        f.Code,
			Severity:    f.Severity,
			AffectedIDs: f.AffectedIDs,
			EvidenceIDs: f.EvidenceIDs,
			Message:     f.Message,
		})
	}

	appEvidence := make([]ApplicationEvidence, 0)
	for _, ev := range evidence {
		appEvidence = append(appEvidence, ApplicationEvidence{
			EvidenceID:  ev.EvidenceID,
			Kind:        ev.Kind,
			ArtifactRef: ev.ArtifactRef,
			Description: ev.Description,
		})
	}

	// 9. Construct Application IR Document
	applicationIR := &ApplicationIRDocument{
		SchemaVersion:     "1.0",
		ProjectID:         input.ProjectID,
		ConfigurationHash: input.ConfigurationHash,
		DocumentID:        sanitizeID(input.DocumentID + "_app"),
		SourceIRRef:       sourceIRDigest,
		Entrypoints:       entrypoints,
		Components:        components,
		Dependencies:      dependencies,
		DataModels:        []json.RawMessage{},
		Workflows:         []json.RawMessage{},
		Findings:          appFindings,
		Evidence:          appEvidence,
	}

	// Validate both generated documents against canonical schemas
	sCjson, err := sourceIR.CanonicalJSON()
	if err != nil {
		return nil, &CI3StaticAdapterError{Reason: "Source IR canonical JSON failed", Cause: err}
	}
	if err := ValidateSourceIR(sCjson); err != nil {
		return nil, &CI3StaticAdapterError{Reason: "generated Source IR failed schema validation", Cause: err}
	}

	aCjson, err := applicationIR.CanonicalJSON()
	if err != nil {
		return nil, &CI3StaticAdapterError{Reason: "Application IR canonical JSON failed", Cause: err}
	}
	if err := ValidateApplicationIR(aCjson); err != nil {
		return nil, &CI3StaticAdapterError{Reason: "generated Application IR failed schema validation", Cause: err}
	}

	result.SourceIR = sourceIR
	result.ApplicationIR = applicationIR

	return result, nil
}

func sanitizeID(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	res := b.String()
	for strings.Contains(res, "__") {
		res = strings.ReplaceAll(res, "__", "_")
	}
	res = strings.Trim(res, "_")
	if res == "" {
		return "node"
	}
	return res
}

func resolveHandlerSymbols(target string, ctrlSyms, methSyms map[string]string) []string {
	parts := strings.Split(target, "/")
	if len(parts) >= 2 {
		key := parts[0] + "." + parts[1]
		if sym, ok := methSyms[key]; ok {
			return []string{sym}
		}
	}
	if len(parts) >= 1 {
		if sym, ok := ctrlSyms[parts[0]]; ok {
			return []string{sym}
		}
	}
	return []string{}
}

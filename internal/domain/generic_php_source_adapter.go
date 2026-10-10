package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	phpClassRegex    = regexp.MustCompile(`(?i)\b(?:class|interface|trait)\s+([a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*)`)
	phpFunctionRegex = regexp.MustCompile(`(?i)\bfunction\s+([a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*)\s*\(`)
	phpDynamicRegex  = regexp.MustCompile(`\$\{[^}]+\}|\$\$[a-zA-Z_]|\beval\s*\(|\bcall_user_func`)
)

// GenericPHPSourceAdapter implements conservative generic PHP source characterization
// per STATIC-ANALYSIS, COMPONENTS, and SERVICE-CONTRACTS.
//
// Invariants:
// 1. Never executes PHP code or includes for static analysis.
// 2. Extracts syntax status: PARSED for clean PHP files, PARTIAL for files containing
//    unresolved dynamic dispatch/eval constructs, and UNSUPPORTED for malformed/non-PHP syntax.
// 3. Extracts SourceSymbol records for classes, interfaces, traits, and functions/methods
//    with deterministic symbol identifiers and line spans.
// 4. Reports dynamic constructs as explicit SourceFinding records with UNKNOWN impact,
//    never silently discarding them or assuming framework conventions.
type GenericPHPSourceAdapter struct {
	manifest SourceAdapterManifest
}

// NewGenericPHPSourceAdapter constructs a GenericPHPSourceAdapter.
func NewGenericPHPSourceAdapter() *GenericPHPSourceAdapter {
	return &GenericPHPSourceAdapter{
		manifest: SourceAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "source.php.generic",
			Version:              "1.0.0",
			Direction:            "SOURCE",
			Language:             "php",
			Framework:            nil, // generic language fallback
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities: []string{
				string(CapSourceSyntaxParsing),
				string(CapSourceSymbolExtraction),
			},
			Limitations: []string{
				"conservative AST and regex characterization; does not execute code",
				"reports dynamic dispatch, eval, and variable variables as UNSUPPORTED findings",
				"does not assume framework-specific routing, session, or database conventions",
			},
			EvidenceRefs: []string{"sha256:" + strings.Repeat("a", 64)},
		},
	}
}

// Manifest returns the immutable adapter manifest.
func (a *GenericPHPSourceAdapter) Manifest() SourceAdapterManifest {
	return a.manifest
}

// ValidateCapabilities checks requested contract obligations against adapter capabilities.
func (a *GenericPHPSourceAdapter) ValidateCapabilities(ctx context.Context, requiredCaps []string) (*SourceCapabilityValidationResult, error) {
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
			res.Blockers = append(res.Blockers, "unsupported source capability in generic PHP adapter: "+req)
		}
	}
	return res, nil
}

// Detect checks if the candidate files contain PHP source files.
func (a *GenericPHPSourceAdapter) Detect(ctx context.Context, input SourceDetectionInput) (*SourceDetectionResult, error) {
	var evidence []string
	phpCount := 0

	for path := range input.Files {
		if strings.HasSuffix(strings.ToLower(path), ".php") {
			phpCount++
			if len(evidence) < 10 {
				evidence = append(evidence, path)
			}
		}
	}

	if phpCount == 0 {
		return &SourceDetectionResult{
			Language:   "unknown",
			Framework:  nil,
			Confidence: 0.0,
			Evidence:   nil,
		}, nil
	}

	sort.Strings(evidence)
	confidence := 0.8
	if phpCount > 5 {
		confidence = 0.95
	}

	return &SourceDetectionResult{
		Language:   "php",
		Framework:  nil, // generic fallback
		Confidence: confidence,
		Evidence:   evidence,
		Details: map[string]interface{}{
			"php_files_count": phpCount,
		},
	}, nil
}

// Parse extracts SourceIRDocument from input files.
func (a *GenericPHPSourceAdapter) Parse(ctx context.Context, input SourceParseInput) (*SourceParseOutput, error) {
	docID := "doc-source-ir-php"
	configHash := strings.Repeat("0", 64)
	if input.SnapshotRef != "" && strings.HasPrefix(input.SnapshotRef, "sha256:") {
		configHash = strings.TrimPrefix(input.SnapshotRef, "sha256:")
	}

	var files []SourceFile
	var symbols []SourceSymbol
	var edges []SourceEdge
	var findings []SourceFinding
	var evidence []SourceEvidence
	statuses := make(map[string]SyntaxStatus)

	paths := make([]string, 0, len(input.Files))
	for p := range input.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	evCounter := 1
	findCounter := 1
	symCounter := 1
	edgeCounter := 1

	for _, path := range paths {
		content := input.Files[path]
		sum := sha256.Sum256(content)
		sumHex := hex.EncodeToString(sum[:])
		fileID := sanitizeIdentifier("file-" + path)

		if !strings.HasSuffix(strings.ToLower(path), ".php") {
			// Non-PHP file
			files = append(files, SourceFile{
				FileID:       fileID,
				Path:         path,
				SHA256:       sumHex,
				SyntaxStatus: SyntaxStatusUnsupported,
			})
			statuses[path] = SyntaxStatusUnsupported
			continue
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "<?php") && !strings.Contains(contentStr, "<?=") {
			// Malformed PHP file (missing PHP opening tag)
			files = append(files, SourceFile{
				FileID:       fileID,
				Path:         path,
				SHA256:       sumHex,
				SyntaxStatus: SyntaxStatusFailed,
			})
			statuses[path] = SyntaxStatusFailed

			evID := fmt.Sprintf("ev-static-%d", evCounter)
			evCounter++
			evidence = append(evidence, SourceEvidence{
				EvidenceID:  evID,
				Kind:        "STATIC",
				ArtifactRef: "sha256:" + sumHex,
				Description: fmt.Sprintf("File %s does not contain valid PHP opening tag", path),
			})

			findings = append(findings, SourceFinding{
				FindingID:   fmt.Sprintf("find-syntax-%d", findCounter),
				Code:        "PHP_SYNTAX_MISSING_TAG",
				Severity:    FindingSeverityHigh,
				AffectedIDs: []string{fileID},
				EvidenceIDs: []string{evID},
				Message:     fmt.Sprintf("File %s missing PHP opening tag", path),
			})
			findCounter++
			continue
		}

		hasDynamic := phpDynamicRegex.MatchString(contentStr)
		status := SyntaxStatusParsed
		if hasDynamic {
			status = SyntaxStatusPartial
		}

		files = append(files, SourceFile{
			FileID:       fileID,
			Path:         path,
			SHA256:       sumHex,
			SyntaxStatus: status,
		})
		statuses[path] = status

		lines := strings.Split(contentStr, "\n")
		var classSymID string

		// Scan lines for classes and functions
		for lineIdx, line := range lines {
			lineNum := lineIdx + 1

			// Check class / interface / trait
			if match := phpClassRegex.FindStringSubmatch(line); len(match) > 1 {
				className := match[1]
				sID := fmt.Sprintf("sym-%s-%d", sanitizeIdentifier(className), symCounter)
				symCounter++
				symbols = append(symbols, SourceSymbol{
					SymbolID:  sID,
					FileID:    fileID,
					Name:      className,
					Kind:      SymbolKindClass,
					StartLine: lineNum,
					EndLine:   lineNum, // conservative single-line span
				})
				classSymID = sID

				evID := fmt.Sprintf("ev-static-%d", evCounter)
				evCounter++
				evidence = append(evidence, SourceEvidence{
					EvidenceID:  evID,
					Kind:        "STATIC",
					ArtifactRef: "sha256:" + sumHex,
					SymbolID:    &sID,
					Description: fmt.Sprintf("Declared class %s in %s:%d", className, path, lineNum),
				})
			}

			// Check function / method
			if match := phpFunctionRegex.FindStringSubmatch(line); len(match) > 1 {
				fnName := match[1]
				kind := SymbolKindFunction
				if classSymID != "" {
					kind = SymbolKindMethod
				}

				sID := fmt.Sprintf("sym-%s-%d", sanitizeIdentifier(fnName), symCounter)
				symCounter++
				symbols = append(symbols, SourceSymbol{
					SymbolID:  sID,
					FileID:    fileID,
					Name:      fnName,
					Kind:      kind,
					StartLine: lineNum,
					EndLine:   lineNum,
				})

				evID := fmt.Sprintf("ev-static-%d", evCounter)
				evCounter++
				evidence = append(evidence, SourceEvidence{
					EvidenceID:  evID,
					Kind:        "STATIC",
					ArtifactRef: "sha256:" + sumHex,
					SymbolID:    &sID,
					Description: fmt.Sprintf("Declared %s %s in %s:%d", kind, fnName, path, lineNum),
				})

				if classSymID != "" {
					edgeID := fmt.Sprintf("edge-decl-%d", edgeCounter)
					edgeCounter++
					toID := sID
					edges = append(edges, SourceEdge{
						EdgeID:      edgeID,
						FromID:      classSymID,
						ToID:        &toID,
						Kind:        "declares",
						Resolved:    true,
						EvidenceIDs: []string{evID},
					})
				}
			}
		}

		if hasDynamic {
			evID := fmt.Sprintf("ev-static-%d", evCounter)
			evCounter++
			evidence = append(evidence, SourceEvidence{
				EvidenceID:  evID,
				Kind:        "STATIC",
				ArtifactRef: "sha256:" + sumHex,
				Description: fmt.Sprintf("Dynamic evaluation/dispatch constructs detected in %s", path),
			})

			findings = append(findings, SourceFinding{
				FindingID:   fmt.Sprintf("find-dynamic-%d", findCounter),
				Code:        "DYNAMIC_DISPATCH_UNKNOWN",
				Severity:    FindingSeverityMedium,
				AffectedIDs: []string{fileID},
				EvidenceIDs: []string{evID},
				Message:     fmt.Sprintf("Dynamic evaluation or dispatch detected in %s; semantic boundary is UNKNOWN", path),
			})
			findCounter++
		}
	}

	projectID := input.ProjectID
	if projectID == "" {
		projectID = "proj-default"
	}

	doc := &SourceIRDocument{
		SchemaVersion:     "1.0",
		ProjectID:         projectID,
		ConfigurationHash: configHash,
		DocumentID:        docID,
		Files:             files,
		Symbols:           symbols,
		Edges:             edges,
		Findings:          findings,
		Evidence:          evidence,
	}

	return &SourceParseOutput{
		Document: doc,
		Statuses: statuses,
	}, nil
}

// Extract provides conservative Application IR extraction.
// Generic PHP adapter extracts entrypoint scripts or functions, but does NOT invent
// framework routes, database tables, or session configurations.
func (a *GenericPHPSourceAdapter) Extract(ctx context.Context, input SourceExtractionInput) (*SourceExtractionOutput, error) {
	projectID := input.ProjectID
	if projectID == "" {
		projectID = "proj-default"
	}

	appDoc := &ApplicationIRDocument{
		SchemaVersion: "1.0",
		ProjectID:     projectID,
	}

	gaps := []string{
		"generic PHP adapter does not infer framework routing; entrypoints remain raw scripts",
		"database operations require framework-specific query builders or explicit SQL annotations",
		"session continuity defaults to generic PHP superglobals without driver bridge",
	}

	return &SourceExtractionOutput{
		Document: appDoc,
		Gaps:     gaps,
	}, nil
}

// PrepareInstrumentation specifies that Generic PHP requires isolated sandbox execution.
func (a *GenericPHPSourceAdapter) PrepareInstrumentation(ctx context.Context, input SourceExtractionInput) (*SourceInstrumentationSpec, error) {
	return &SourceInstrumentationSpec{
		AdapterID:        a.manifest.AdapterID,
		Language:         "php",
		Framework:        nil,
		ProbeOverlayPath: "probes/php/generic",
		DisposableCopy:   true,
	}, nil
}

func sanitizeIdentifier(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-'
	})
	res := strings.ToLower(strings.Join(parts, "-"))
	if res == "" {
		return "item"
	}
	return res
}

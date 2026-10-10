package domain

import (
	"context"
	"strings"
	"testing"
)

func TestGenericPHPSourceAdapter_Detect(t *testing.T) {
	ctx := context.Background()
	adapter := NewGenericPHPSourceAdapter()

	// 1. Valid PHP files detected
	input := SourceDetectionInput{
		ProjectID:   "proj-detect-test",
		SnapshotRef: "sha256:" + strings.Repeat("1", 64),
		Files: map[string][]byte{
			"index.php":      []byte("<?php echo 'hello';"),
			"config.php":     []byte("<?php return [];"),
			"assets/app.css": []byte("body { color: red; }"),
		},
	}

	res, err := adapter.Detect(ctx, input)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if res.Language != "php" || res.Framework != nil {
		t.Errorf("expected language=php framework=nil, got language=%s framework=%v", res.Language, res.Framework)
	}
	if res.Confidence < 0.7 {
		t.Errorf("expected high confidence, got %f", res.Confidence)
	}

	// 2. Non-PHP project returns unknown
	inputNoPHP := SourceDetectionInput{
		ProjectID: "proj-no-php",
		Files: map[string][]byte{
			"main.go": []byte("package main"),
			"app.py":  []byte("print('hello')"),
		},
	}
	resNoPHP, err := adapter.Detect(ctx, inputNoPHP)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if resNoPHP.Language != "unknown" || resNoPHP.Confidence != 0.0 {
		t.Errorf("expected unknown and 0.0 confidence, got %+v", resNoPHP)
	}
}

func TestGenericPHPSourceAdapter_Parse_ConservativeStatuses(t *testing.T) {
	ctx := context.Background()
	adapter := NewGenericPHPSourceAdapter()

	files := map[string][]byte{
		// Clean file -> PARSED
		"src/User.php": []byte(`<?php
class User {
    function getName() {
        return "Alice";
    }
}
`),
		// Dynamic dispatch file -> PARTIAL + Finding
		"src/Dynamic.php": []byte(`<?php
class DynamicRunner {
    function run($action) {
        eval("return 42;");
    }
}
`),
		// Malformed PHP file (missing tag) -> FAILED + Finding
		"src/Malformed.php": []byte(`
class Broken {
    function fail() {}
}
`),
		// Non-PHP file -> UNSUPPORTED
		"src/readme.txt": []byte("This is plain text"),
	}

	input := SourceParseInput{
		ProjectID:   "proj-parse-test",
		SnapshotRef: "sha256:" + strings.Repeat("2", 64),
		Files:       files,
	}

	out, err := adapter.Parse(ctx, input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check status categorization
	if out.Statuses["src/User.php"] != SyntaxStatusParsed {
		t.Errorf("expected PARSED for User.php, got %s", out.Statuses["src/User.php"])
	}
	if out.Statuses["src/Dynamic.php"] != SyntaxStatusPartial {
		t.Errorf("expected PARTIAL for Dynamic.php, got %s", out.Statuses["src/Dynamic.php"])
	}
	if out.Statuses["src/Malformed.php"] != SyntaxStatusFailed {
		t.Errorf("expected FAILED for Malformed.php, got %s", out.Statuses["src/Malformed.php"])
	}
	if out.Statuses["src/readme.txt"] != SyntaxStatusUnsupported {
		t.Errorf("expected UNSUPPORTED for readme.txt, got %s", out.Statuses["src/readme.txt"])
	}

	// Check findings: Dynamic file must produce DYNAMIC_DISPATCH_UNKNOWN finding
	foundDynamicFinding := false
	for _, f := range out.Document.Findings {
		if f.Code == "DYNAMIC_DISPATCH_UNKNOWN" {
			foundDynamicFinding = true
			break
		}
	}
	if !foundDynamicFinding {
		t.Errorf("expected DYNAMIC_DISPATCH_UNKNOWN finding for eval in Dynamic.php")
	}

	// Validate the generated Source IR document against canonical Source IR schema rules
	cjson, err := out.Document.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	if err := ValidateSourceIR(cjson); err != nil {
		t.Fatalf("generated SourceIRDocument failed ValidateSourceIR: %v", err)
	}
}

func TestGenericPHPSourceAdapter_Extract_ExplicitGaps(t *testing.T) {
	ctx := context.Background()
	adapter := NewGenericPHPSourceAdapter()

	input := SourceExtractionInput{
		ProjectID:   "proj-extract-test",
		SnapshotRef: "sha256:" + strings.Repeat("3", 64),
	}

	out, err := adapter.Extract(ctx, input)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(out.Gaps) == 0 {
		t.Fatalf("expected explicit gaps reporting from Generic PHP adapter, got 0 gaps")
	}
}

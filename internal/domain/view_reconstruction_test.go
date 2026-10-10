package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func makeTestViewAdmission(t *testing.T, unitID string) *GenerationAdmission {
	t.Helper()
	plan, checks := admissionTestPlan(t)
	adm, err := EvaluateGenerationAdmission(plan, "sync-auth", checks)
	if err != nil {
		t.Fatal(err)
	}
	adm.Scope.UnitIDs = []string{unitID}
	adm.Scope.CaseIDs = []string{"login-view", "dashboard-view"}
	adm.Capabilities = []string{"http", "filesystem"}
	return adm
}

func makeTestViewTargetModel(t *testing.T, unitID string, allowedPaths []string) *TargetModelDocument {
	t.Helper()
	builder := NewTargetModelBuilder("model-views", "proj-test", validTestRef("a"), validTestRef("b"))
	builder.WithRevision(1)
	builder.AddComponent(TargetModelComponent{
		ComponentID:          "comp-views-" + unitID,
		OwnerUnitID:          unitID,
		ConsumerUnitIDs:      []string{},
		Paths:                allowedPaths,
		ContractRefs:         []string{validTestRef("c")},
		LoweringChoice:       "blade_templates",
		RequiredCapabilities: []string{"http", "filesystem"},
		ConstraintRefs:       []string{validTestRef("d")},
	})
	doc, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestViewFileReconstructor_DeterministicReconstruction(t *testing.T) {
	unitID := "unit-views"
	allowedPaths := []string{
		"resources/views/auth/login.blade.php",
		"public/css/app.css",
	}

	adm := makeTestViewAdmission(t, unitID)
	tm := makeTestViewTargetModel(t, unitID, allowedPaths)

	obligations := []ViewTemplateObligation{
		{
			ViewID:       "view-login",
			SourcePath:   "application/views/auth/login.php",
			TargetPath:   "resources/views/auth/login.blade.php",
			Kind:         ViewKindBladeTemplate,
			EscapingMode: EscapingModeAuto,
			RawTemplate:  "<form action='/login'><h1>Login</h1><p>{{ $title }}</p></form>",
			Variables: []ViewVariable{
				{Name: "title", Type: "string", RawApproved: false},
			},
		},
		{
			ViewID:       "asset-css",
			SourcePath:   "assets/css/app.css",
			TargetPath:   "public/css/app.css",
			Kind:         ViewKindStaticAsset,
			EscapingMode: EscapingModeNone,
			RawTemplate:  "body { font-family: sans-serif; }",
		},
	}

	reconstructor := NewViewFileReconstructor()
	input := ViewReconstructionInput{
		Admission:    adm,
		TargetModel:  tm,
		UnitID:       unitID,
		Obligations:  obligations,
		FencingToken: 10,
	}

	out1, err := reconstructor.Reconstruct(input)
	if err != nil {
		t.Fatalf("first reconstruction failed: %v", err)
	}

	out2, err := reconstructor.Reconstruct(input)
	if err != nil {
		t.Fatalf("second reconstruction failed: %v", err)
	}

	// Deterministic artifacts
	if len(out1.Artifacts) != len(out2.Artifacts) {
		t.Fatalf("artifact counts differ: %d vs %d", len(out1.Artifacts), len(out2.Artifacts))
	}
	for path, bytes1 := range out1.Artifacts {
		bytes2, ok := out2.Artifacts[path]
		if !ok || string(bytes1) != string(bytes2) {
			t.Fatalf("artifact %s nondeterministic", path)
		}
	}

	// Deterministic manifest
	manRaw1, err := json.Marshal(out1.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	manRaw2, err := json.Marshal(out2.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	if string(manRaw1) != string(manRaw2) {
		t.Fatalf("manifest nondeterministic: %s vs %s", string(manRaw1), string(manRaw2))
	}

	// Validate manifest with canonical validator
	if err := ValidateViewManifest(manRaw1); err != nil {
		t.Fatalf("view manifest validation failed: %v", err)
	}

	// Deterministic patch
	patchRaw1, err := out1.Patch.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	patchRaw2, err := out2.Patch.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(patchRaw1) != string(patchRaw2) {
		t.Fatalf("patch nondeterministic")
	}

	// Validate patch with canonical validator
	if err := ValidatePatch(patchRaw1); err != nil {
		t.Fatalf("patch validation failed: %v", err)
	}
}

func TestViewFileReconstructor_PathTraversal_Rejected(t *testing.T) {
	unitID := "unit-views"
	adm := makeTestViewAdmission(t, unitID)
	tm := makeTestViewTargetModel(t, unitID, nil) // no TIM path restriction to isolate traversal checks
	reconstructor := NewViewFileReconstructor()

	traversalCases := []struct {
		name       string
		targetPath string
	}{
		{"dot-dot traversal", "resources/views/../../etc/passwd"},
		{"embedded dot-dot", "resources/views/sub/../../../etc/shadow"},
		{"null byte", "resources/views/login.blade.php\x00evil"},
		{"absolute path", "/resources/views/login.blade.php"},
		{"outside allowed space", "app/Http/Controllers/Evil.php"},
		{"storage space", "storage/framework/views/evil.php"},
	}

	for _, tc := range traversalCases {
		t.Run(tc.name, func(t *testing.T) {
			input := ViewReconstructionInput{
				Admission:   adm,
				TargetModel: tm,
				UnitID:      unitID,
				Obligations: []ViewTemplateObligation{
					{
						ViewID:       "view-test",
						SourcePath:   "test.php",
						TargetPath:   tc.targetPath,
						Kind:         ViewKindBladeTemplate,
						EscapingMode: EscapingModeAuto,
						RawTemplate:  "<p>hello</p>",
					},
				},
				FencingToken: 1,
			}

			_, err := reconstructor.Reconstruct(input)
			if err == nil {
				t.Fatalf("expected error for path traversal vector %q, but succeeded", tc.targetPath)
			}
			var viewErr *ViewReconstructionError
			if !strings.Contains(err.Error(), "ERR_TARGET_CONFLICT") {
				t.Fatalf("expected ERR_TARGET_CONFLICT for %q, got: %v", tc.targetPath, err)
			}
			_ = viewErr
		})
	}
}

func TestViewFileReconstructor_UnapprovedTIMPath_Rejected(t *testing.T) {
	unitID := "unit-views"
	adm := makeTestViewAdmission(t, unitID)
	// TIM only approves login.blade.php
	tm := makeTestViewTargetModel(t, unitID, []string{"resources/views/auth/login.blade.php"})

	reconstructor := NewViewFileReconstructor()
	input := ViewReconstructionInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
		Obligations: []ViewTemplateObligation{
			{
				ViewID:       "view-unapproved",
				SourcePath:   "unapproved.php",
				TargetPath:   "resources/views/unapproved/secret.blade.php", // not in TIM!
				Kind:         ViewKindBladeTemplate,
				EscapingMode: EscapingModeAuto,
				RawTemplate:  "<p>secret</p>",
			},
		},
		FencingToken: 1,
	}

	_, err := reconstructor.Reconstruct(input)
	if err == nil {
		t.Fatal("expected error for path not in TIM component paths")
	}
	if !strings.Contains(err.Error(), "ERR_UNAPPROVED_VIEW") {
		t.Fatalf("expected ERR_UNAPPROVED_VIEW, got: %v", err)
	}
}

func TestViewFileReconstructor_InvalidEncoding_Rejected(t *testing.T) {
	unitID := "unit-views"
	adm := makeTestViewAdmission(t, unitID)
	tm := makeTestViewTargetModel(t, unitID, nil)

	reconstructor := NewViewFileReconstructor()
	// Invalid UTF-8 sequence \xff\xfe
	invalidUTF8 := string([]byte{0xff, 0xfe, 0x61, 0x62})

	input := ViewReconstructionInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
		Obligations: []ViewTemplateObligation{
			{
				ViewID:       "view-invalid-utf8",
				SourcePath:   "bad.php",
				TargetPath:   "resources/views/bad.blade.php",
				Kind:         ViewKindBladeTemplate,
				EscapingMode: EscapingModeAuto,
				RawTemplate:  invalidUTF8,
			},
		},
		FencingToken: 1,
	}

	_, err := reconstructor.Reconstruct(input)
	if err == nil {
		t.Fatal("expected error for invalid UTF-8 template")
	}
	if !strings.Contains(err.Error(), "ERR_INVALID_ENCODING") {
		t.Fatalf("expected ERR_INVALID_ENCODING, got: %v", err)
	}
}

func TestViewFileReconstructor_UnapprovedRawEscaping_Rejected(t *testing.T) {
	unitID := "unit-views"
	adm := makeTestViewAdmission(t, unitID)
	tm := makeTestViewTargetModel(t, unitID, nil)
	reconstructor := NewViewFileReconstructor()

	t.Run("UnapprovedRawInterpolation_Fails", func(t *testing.T) {
		input := ViewReconstructionInput{
			Admission:   adm,
			TargetModel: tm,
			UnitID:      unitID,
			Obligations: []ViewTemplateObligation{
				{
					ViewID:       "view-unapproved-raw",
					SourcePath:   "raw.php",
					TargetPath:   "resources/views/raw.blade.php",
					Kind:         ViewKindBladeTemplate,
					EscapingMode: EscapingModeAuto,
					RawTemplate:  "<div>{!! $untrusted_input !!}</div>",
					Variables: []ViewVariable{
						{Name: "untrusted_input", Type: "string", RawApproved: false},
					},
				},
			},
			FencingToken: 1,
		}

		_, err := reconstructor.Reconstruct(input)
		if err == nil {
			t.Fatal("expected error for unapproved raw interpolation")
		}
		if !strings.Contains(err.Error(), "ERR_UNAPPROVED_ESCAPING") {
			t.Fatalf("expected ERR_UNAPPROVED_ESCAPING, got: %v", err)
		}
	})

	t.Run("ApprovedRawInterpolation_Succeeds", func(t *testing.T) {
		input := ViewReconstructionInput{
			Admission:   adm,
			TargetModel: tm,
			UnitID:      unitID,
			Obligations: []ViewTemplateObligation{
				{
					ViewID:       "view-approved-raw",
					SourcePath:   "raw.php",
					TargetPath:   "resources/views/raw.blade.php",
					Kind:         ViewKindBladeTemplate,
					EscapingMode: EscapingModeAuto,
					RawTemplate:  "<div>{!! $safe_html !!}</div>",
					Variables: []ViewVariable{
						{Name: "safe_html", Type: "string", RawApproved: true},
					},
				},
			},
			FencingToken: 1,
		}

		out, err := reconstructor.Reconstruct(input)
		if err != nil {
			t.Fatalf("approved raw interpolation should succeed, got: %v", err)
		}
		if !strings.Contains(string(out.Artifacts["resources/views/raw.blade.php"]), "{!! $safe_html !!}") {
			t.Fatal("expected rendered output to contain raw interpolation")
		}
	})
}

func TestViewFileReconstructor_ValidateManifest_Vectors(t *testing.T) {
	unitID := "unit-views"
	adm := makeTestViewAdmission(t, unitID)
	tm := makeTestViewTargetModel(t, unitID, nil)

	reconstructor := NewViewFileReconstructor()
	out, err := reconstructor.Reconstruct(ViewReconstructionInput{
		Admission:   adm,
		TargetModel: tm,
		UnitID:      unitID,
		Obligations: []ViewTemplateObligation{
			{
				ViewID:       "view-test",
				SourcePath:   "test.php",
				TargetPath:   "resources/views/test.blade.php",
				Kind:         ViewKindBladeTemplate,
				EscapingMode: EscapingModeAuto,
				RawTemplate:  "<p>test</p>",
			},
		},
		FencingToken: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	baseRaw, err := json.Marshal(out.Manifest)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ValidManifest_Passes", func(t *testing.T) {
		if err := ValidateViewManifest(baseRaw); err != nil {
			t.Fatalf("expected valid manifest to pass: %v", err)
		}
	})

	t.Run("InvalidSchemaVersion_Fails", func(t *testing.T) {
		var m ViewManifest
		json.Unmarshal(baseRaw, &m)
		m.SchemaVersion = "2.0"
		raw, _ := json.Marshal(m)
		if err := ValidateViewManifest(raw); err == nil {
			t.Fatal("expected error for invalid schema_version")
		}
	})

	t.Run("DuplicateEntryPaths_Fails", func(t *testing.T) {
		var m ViewManifest
		json.Unmarshal(baseRaw, &m)
		m.Entries = append(m.Entries, m.Entries[0])
		raw, _ := json.Marshal(m)
		if err := ValidateViewManifest(raw); err == nil {
			t.Fatal("expected error for duplicate entry paths")
		}
	})
}

func TestViewFileReconstructor_AdmissionGating(t *testing.T) {
	unitID := "unit-views"
	adm := makeTestViewAdmission(t, unitID)
	tm := makeTestViewTargetModel(t, unitID, nil)
	reconstructor := NewViewFileReconstructor()

	ob := []ViewTemplateObligation{
		{
			ViewID:       "view-test",
			SourcePath:   "test.php",
			TargetPath:   "resources/views/test.blade.php",
			Kind:         ViewKindBladeTemplate,
			EscapingMode: EscapingModeAuto,
			RawTemplate:  "<p>test</p>",
		},
	}

	t.Run("NilAdmission_Fails", func(t *testing.T) {
		_, err := reconstructor.Reconstruct(ViewReconstructionInput{
			Admission:    nil,
			TargetModel:  tm,
			UnitID:       unitID,
			Obligations:  ob,
			FencingToken: 1,
		})
		if err == nil || !strings.Contains(err.Error(), "ERR_ADMISSION") {
			t.Fatalf("expected ERR_ADMISSION for nil admission, got: %v", err)
		}
	})

	t.Run("BlockedAdmission_Fails", func(t *testing.T) {
		admBlocked := makeTestViewAdmission(t, unitID)
		admBlocked.Outcome = "BLOCKED"
		_, err := reconstructor.Reconstruct(ViewReconstructionInput{
			Admission:    admBlocked,
			TargetModel:  tm,
			UnitID:       unitID,
			Obligations:  ob,
			FencingToken: 1,
		})
		if err == nil || !strings.Contains(err.Error(), "ERR_ADMISSION") {
			t.Fatalf("expected ERR_ADMISSION for BLOCKED admission, got: %v", err)
		}
	})

	t.Run("UnitNotInScope_Fails", func(t *testing.T) {
		_, err := reconstructor.Reconstruct(ViewReconstructionInput{
			Admission:    adm,
			TargetModel:  tm,
			UnitID:       "other-unadmitted-unit",
			Obligations:  ob,
			FencingToken: 1,
		})
		if err == nil || !strings.Contains(err.Error(), "ERR_SCOPE_MISMATCH") {
			t.Fatalf("expected ERR_SCOPE_MISMATCH, got: %v", err)
		}
	})
}

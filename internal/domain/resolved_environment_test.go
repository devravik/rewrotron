package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestResolvedEnvironmentRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasonMap := map[string]string{
		"v2-reject-resolution-13.wild":          "runtime version",
		"v2-reject-resolution-latest":           "runtime version",
		"v2-reject-resolution-^8.3":             "runtime version",
		"v2-reject-resolution-8.3":              "runtime version",
		"review2-opaque-release-needs-provider": "missing runtime field",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/resolved-environment.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateResolvedEnvironment(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid resolved environment for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, ok := reasonMap[fixture.Name]
			if !ok {
				t.Fatalf("missing expected reason mapping for invalid fixture: %s", fixture.Name)
			}
			assertResolvedEnvironmentRejected(t, err, expectedReason)
		})
	}
	if count != 7 {
		t.Fatalf("expected 7 resolved environment fixtures, found %d", count)
	}
}

func resolvedEnvironmentControl(t *testing.T, name string) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == name {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatalf("missing %s fixture", name)
	return nil
}

func assertResolvedEnvironmentRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var resErr *ResolvedEnvironmentError
	if !errors.As(err, &resErr) || resErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid resolved environment: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestResolvedEnvironmentValidRecords(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"v2-exact-resolution", "review2-exact-nonnumeric-release"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := resolvedEnvironmentControl(t, name)
			raw, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := CanonicalJSON(raw)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateResolvedEnvironment(canonical); err != nil {
				t.Fatalf("expected valid %s, got: %v", name, err)
			}
		})
	}
}

func TestResolvedEnvironmentWithServicesAndFramework(t *testing.T) {
	t.Parallel()
	ctrl := resolvedEnvironmentControl(t, "v2-exact-resolution")
	ctrl["framework"] = map[string]any{
		"name":             "laravel",
		"version":          "11.0.0",
		"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	ctrl["package_manager"] = map[string]any{
		"name":             "composer",
		"version":          "2.7.0",
		"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	ctrl["services"] = []any{
		map[string]any{
			"service_id": "db",
			"kind":       "DATABASE",
			"component": map[string]any{
				"name":             "mysql",
				"version":          "8.0.35",
				"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			"profile_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		map[string]any{
			"service_id": "redis",
			"kind":       "CACHE",
			"component": map[string]any{
				"name":             "valkey",
				"version":          "7.2.0",
				"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			"profile_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateResolvedEnvironment(canonical); err != nil {
		t.Fatalf("expected valid complex resolved environment, got: %v", err)
	}
}

func TestResolvedEnvironmentMutationsV20(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		reason  string
		rawJSON string
	}{
		{
			name:    "invalid-json",
			rawJSON: `{"schema_version": "2.0",`,
			reason:  "canonical JSON",
		},
		{
			name:    "non-canonical-float",
			rawJSON: `{"schema_version":"2.0","val":12.34}`,
			reason:  "canonical JSON",
		},
		{
			name: "extra-root-field",
			mutate: func(m map[string]any) {
				m["extra"] = "forbidden"
			},
			reason: "object fields",
		},
		{
			name: "missing-schema-version",
			mutate: func(m map[string]any) {
				delete(m, "schema_version")
			},
			reason: "missing required field",
		},
		{
			name: "unsupported-schema-version-10",
			mutate: func(m map[string]any) {
				m["schema_version"] = "1.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "unsupported-schema-version-30",
			mutate: func(m map[string]any) {
				m["schema_version"] = "3.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "missing-resolution-id",
			mutate: func(m map[string]any) {
				delete(m, "resolution_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-resolution-id",
			mutate: func(m map[string]any) {
				m["resolution_id"] = "RES_UPPER"
			},
			reason: "resolution identity",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "PROJ_UPPER"
			},
			reason: "project identity",
		},
		{
			name: "missing-intent-ref",
			mutate: func(m map[string]any) {
				delete(m, "intent_ref")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-intent-ref",
			mutate: func(m map[string]any) {
				m["intent_ref"] = "bad-hash"
			},
			reason: "intent_ref",
		},
		{
			name: "empty-language",
			mutate: func(m map[string]any) {
				m["language"] = "  "
			},
			reason: "language",
		},
		{
			name: "empty-os",
			mutate: func(m map[string]any) {
				m["os"] = ""
			},
			reason: "os",
		},
		{
			name: "empty-architecture",
			mutate: func(m map[string]any) {
				m["architecture"] = " "
			},
			reason: "architecture",
		},
		{
			name: "runtime-not-object",
			mutate: func(m map[string]any) {
				m["runtime"] = "php"
			},
			reason: "runtime object",
		},
		{
			name: "runtime-missing-field",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				delete(rt, "distribution_ref")
			},
			reason: "missing runtime field",
		},
		{
			name: "runtime-extra-field",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				rt["extra"] = "forbidden"
			},
			reason: "runtime fields",
		},
		{
			name: "runtime-empty-name",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				rt["name"] = ""
			},
			reason: "runtime name",
		},
		{
			name: "runtime-invalid-semver",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				rt["version"] = "v8.3.1"
			},
			reason: "runtime version",
		},
		{
			name: "runtime-invalid-distribution-ref",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				rt["distribution_ref"] = "sha256:bad"
			},
			reason: "runtime distribution_ref",
		},
		{
			name: "framework-invalid-object",
			mutate: func(m map[string]any) {
				m["framework"] = "laravel"
			},
			reason: "framework object",
		},
		{
			name: "framework-invalid-semver",
			mutate: func(m map[string]any) {
				m["framework"] = map[string]any{
					"name":             "laravel",
					"version":          "latest",
					"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				}
			},
			reason: "framework version",
		},
		{
			name: "package-manager-invalid-distribution-ref",
			mutate: func(m map[string]any) {
				m["package_manager"] = map[string]any{
					"name":             "composer",
					"version":          "2.7.0",
					"distribution_ref": "not-a-hash",
				}
			},
			reason: "package manager distribution_ref",
		},
		{
			name: "services-not-array",
			mutate: func(m map[string]any) {
				m["services"] = "none"
			},
			reason: "services array",
		},
		{
			name: "service-not-object",
			mutate: func(m map[string]any) {
				m["services"] = []any{"mysql"}
			},
			reason: "service object",
		},
		{
			name: "service-missing-field",
			mutate: func(m map[string]any) {
				m["services"] = []any{
					map[string]any{
						"service_id": "db",
						"kind":       "DATABASE",
					},
				}
			},
			reason: "missing service field",
		},
		{
			name: "service-extra-field",
			mutate: func(m map[string]any) {
				m["services"] = []any{
					map[string]any{
						"service_id": "db",
						"kind":       "DATABASE",
						"component": map[string]any{
							"name":             "mysql",
							"version":          "8.0.35",
							"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						},
						"profile_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"extra":       "forbidden",
					},
				}
			},
			reason: "service fields",
		},
		{
			name: "service-invalid-id",
			mutate: func(m map[string]any) {
				m["services"] = []any{
					map[string]any{
						"service_id": "DB_UPPER",
						"kind":       "DATABASE",
						"component": map[string]any{
							"name":             "mysql",
							"version":          "8.0.35",
							"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						},
						"profile_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				}
			},
			reason: "service identity",
		},
		{
			name: "service-duplicate-id",
			mutate: func(m map[string]any) {
				svc := map[string]any{
					"service_id": "db",
					"kind":       "DATABASE",
					"component": map[string]any{
						"name":             "mysql",
						"version":          "8.0.35",
						"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
					"profile_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				}
				m["services"] = []any{svc, svc}
			},
			reason: "duplicate service identity",
		},
		{
			name: "service-invalid-kind",
			mutate: func(m map[string]any) {
				m["services"] = []any{
					map[string]any{
						"service_id": "db",
						"kind":       "UNKNOWN_KIND",
						"component": map[string]any{
							"name":             "mysql",
							"version":          "8.0.35",
							"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						},
						"profile_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				}
			},
			reason: "invalid service kind",
		},
		{
			name: "service-invalid-profile-ref",
			mutate: func(m map[string]any) {
				m["services"] = []any{
					map[string]any{
						"service_id": "db",
						"kind":       "DATABASE",
						"component": map[string]any{
							"name":             "mysql",
							"version":          "8.0.35",
							"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						},
						"profile_ref": "bad-ref",
					},
				}
			},
			reason: "service profile_ref",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var raw []byte
			if tc.rawJSON != "" {
				raw = []byte(tc.rawJSON)
			} else {
				ctrl := resolvedEnvironmentControl(t, "v2-exact-resolution")
				tc.mutate(ctrl)
				data, err := json.Marshal(ctrl)
				if err != nil {
					t.Fatal(err)
				}
				canon, err := CanonicalJSON(data)
				if err != nil {
					t.Fatal(err)
				}
				raw = canon
			}

			err := ValidateResolvedEnvironment(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertResolvedEnvironmentRejected(t, err, tc.reason)
		})
	}
}

func TestResolvedEnvironmentMutationsV21(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(map[string]any)
		reason string
	}{
		{
			name: "v21-missing-toolchain-components",
			mutate: func(m map[string]any) {
				delete(m, "toolchain_components")
			},
			reason: "missing required field",
		},
		{
			name: "v21-with-valid-toolchain-component",
			mutate: func(m map[string]any) {
				m["toolchain_components"] = []any{
					map[string]any{
						"provider_id":        "composer_provider",
						"name":               "composer",
						"version":            "2.7.2",
						"distribution_ref":   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"interpretation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				}
			},
			reason: "", // valid case
		},
		{
			name: "v21-toolchain-duplicate",
			mutate: func(m map[string]any) {
				tool := map[string]any{
					"provider_id":        "composer_provider",
					"name":               "composer",
					"version":            "2.7.2",
					"distribution_ref":   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"interpretation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				}
				m["toolchain_components"] = []any{tool, tool}
			},
			reason: "duplicate toolchain component",
		},
		{
			name: "v21-toolchain-missing-interpretation-ref",
			mutate: func(m map[string]any) {
				m["toolchain_components"] = []any{
					map[string]any{
						"provider_id":      "composer_provider",
						"name":             "composer",
						"version":          "2.7.2",
						"distribution_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				}
			},
			reason: "missing toolchain component field",
		},
		{
			name: "v21-toolchain-empty-name",
			mutate: func(m map[string]any) {
				m["toolchain_components"] = []any{
					map[string]any{
						"provider_id":        "composer_provider",
						"name":               "",
						"version":            "2.7.2",
						"distribution_ref":   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"interpretation_ref": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				}
			},
			reason: "toolchain component name",
		},
		{
			name: "v21-toolchain-invalid-interpretation-ref",
			mutate: func(m map[string]any) {
				m["toolchain_components"] = []any{
					map[string]any{
						"provider_id":        "composer_provider",
						"name":               "composer",
						"version":            "2.7.2",
						"distribution_ref":   "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"interpretation_ref": "bad-ref",
					},
				}
			},
			reason: "toolchain component interpretation_ref",
		},
		{
			name: "v21-runtime-missing-interpretation-ref",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				delete(rt, "interpretation_ref")
			},
			reason: "missing runtime field",
		},
		{
			name: "v21-runtime-invalid-interpretation-ref",
			mutate: func(m map[string]any) {
				rt := m["runtime"].(map[string]any)
				rt["interpretation_ref"] = "sha256:short"
			},
			reason: "runtime interpretation_ref",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := resolvedEnvironmentControl(t, "review2-exact-nonnumeric-release")
			tc.mutate(ctrl)
			data, err := json.Marshal(ctrl)
			if err != nil {
				t.Fatal(err)
			}
			canon, err := CanonicalJSON(data)
			if err != nil {
				t.Fatal(err)
			}

			err = ValidateResolvedEnvironment(canon)
			if tc.reason == "" {
				if err != nil {
					t.Fatalf("expected valid for case %s, got: %v", tc.name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertResolvedEnvironmentRejected(t, err, tc.reason)
		})
	}
}

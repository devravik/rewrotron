package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestWorkScopeRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasonMap := map[string]string{
		"work-scope-cannot-escape-snapshot":   "invalid source path",
		"work-scope-cannot-address-host-root": "invalid source path",
		"work-scope-needs-a-boundary":         "included_source_paths array",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/work-scope.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateWorkScope(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid work scope for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, ok := reasonMap[fixture.Name]
			if !ok {
				t.Fatalf("missing expected reason mapping for invalid fixture: %s", fixture.Name)
			}
			assertWorkScopeRejected(t, err, expectedReason)
		})
	}
	if count != 4 {
		t.Fatalf("expected 4 work scope fixtures, found %d", count)
	}
}

func workScopeControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "shared-work-scope-needs-no-executable-corpus" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing shared-work-scope-needs-no-executable-corpus fixture")
	return nil
}

func assertWorkScopeRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var scopeErr *WorkScopeError
	if !errors.As(err, &scopeErr) || scopeErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid work scope: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestWorkScopeValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := workScopeControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkScope(canonical); err != nil {
		t.Fatalf("expected valid control work scope, got: %v", err)
	}
}

func TestWorkScopeWithOperationalProfileAndEntrypoints(t *testing.T) {
	t.Parallel()
	ctrl := workScopeControl(t)
	ctrl["operational_profile_ref"] = "sha256:" + strings.Repeat("a", 64)
	ctrl["included_entrypoint_ids"] = []any{"index", "api"}
	ctrl["included_source_paths"] = []any{"app", "config/routes.php", "public/index.php"}
	ctrl["critical_workflow_refs"] = []any{"sha256:" + strings.Repeat("b", 64)}
	ctrl["constraint_refs"] = []any{"sha256:" + strings.Repeat("c", 64)}
	ctrl["business_requirement_refs"] = []any{"sha256:" + strings.Repeat("d", 64)}
	ctrl["excluded_scope_refs"] = []any{"sha256:" + strings.Repeat("e", 64)}

	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkScope(canonical); err != nil {
		t.Fatalf("expected valid populated work scope, got: %v", err)
	}
}

func TestWorkScopeMutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(map[string]any)
		reason  string
		rawJSON string
	}{
		{
			name:    "invalid-json",
			rawJSON: `{"schema_version": "1.0",`,
			reason:  "canonical JSON",
		},
		{
			name:    "non-canonical-float",
			rawJSON: `{"schema_version":"1.0","val":12.34}`,
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
			name: "unsupported-schema-version",
			mutate: func(m map[string]any) {
				m["schema_version"] = "2.0"
			},
			reason: "unsupported schema version",
		},
		{
			name: "missing-scope-id",
			mutate: func(m map[string]any) {
				delete(m, "scope_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-scope-id",
			mutate: func(m map[string]any) {
				m["scope_id"] = "SCOPE_UPPER"
			},
			reason: "scope identity",
		},
		{
			name: "missing-project-id",
			mutate: func(m map[string]any) {
				delete(m, "project_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-project-id",
			mutate: func(m map[string]any) {
				m["project_id"] = "PROJ_UPPER"
			},
			reason: "project identity",
		},
		{
			name: "missing-source-snapshot-ref",
			mutate: func(m map[string]any) {
				delete(m, "source_snapshot_ref")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-source-snapshot-ref",
			mutate: func(m map[string]any) {
				m["source_snapshot_ref"] = "invalid-hash"
			},
			reason: "source_snapshot_ref",
		},
		{
			name: "entrypoint-ids-not-array",
			mutate: func(m map[string]any) {
				m["included_entrypoint_ids"] = "not-array"
			},
			reason: "included_entrypoint_ids array",
		},
		{
			name: "entrypoint-ids-invalid-id",
			mutate: func(m map[string]any) {
				m["included_entrypoint_ids"] = []any{"ENTRY_UPPER"}
			},
			reason: "invalid included_entrypoint_ids identity",
		},
		{
			name: "entrypoint-ids-duplicate",
			mutate: func(m map[string]any) {
				m["included_entrypoint_ids"] = []any{"entry-1", "entry-1"}
			},
			reason: "duplicate included_entrypoint_ids identity",
		},
		{
			name: "source-paths-not-array",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = "."
			},
			reason: "included_source_paths array",
		},
		{
			name: "source-paths-empty",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = []any{}
			},
			reason: "included_source_paths array",
		},
		{
			name: "source-paths-slash-prefix",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = []any{"/var/www"}
			},
			reason: "invalid source path",
		},
		{
			name: "source-paths-dot-dot",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = []any{"app/../secret"}
			},
			reason: "invalid source path",
		},
		{
			name: "source-paths-backslash",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = []any{"app\\secret"}
			},
			reason: "invalid source path",
		},
		{
			name: "source-paths-null-byte",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = []any{"app\x00file"}
			},
			reason: "invalid source path",
		},
		{
			name: "source-paths-duplicate",
			mutate: func(m map[string]any) {
				m["included_source_paths"] = []any{".", "."}
			},
			reason: "duplicate source path",
		},
		{
			name: "critical-workflows-not-array",
			mutate: func(m map[string]any) {
				m["critical_workflow_refs"] = 123
			},
			reason: "critical_workflow_refs array",
		},
		{
			name: "critical-workflows-invalid-ref",
			mutate: func(m map[string]any) {
				m["critical_workflow_refs"] = []any{"bad-ref"}
			},
			reason: "invalid critical_workflow_refs reference",
		},
		{
			name: "critical-workflows-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:" + strings.Repeat("a", 64)
				m["critical_workflow_refs"] = []any{ref, ref}
			},
			reason: "duplicate critical_workflow_refs reference",
		},
		{
			name: "constraints-not-array",
			mutate: func(m map[string]any) {
				m["constraint_refs"] = "not-array"
			},
			reason: "constraint_refs array",
		},
		{
			name: "constraints-invalid-ref",
			mutate: func(m map[string]any) {
				m["constraint_refs"] = []any{"bad-ref"}
			},
			reason: "invalid constraint_refs reference",
		},
		{
			name: "business-requirements-not-array",
			mutate: func(m map[string]any) {
				m["business_requirement_refs"] = false
			},
			reason: "business_requirement_refs array",
		},
		{
			name: "business-requirements-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:" + strings.Repeat("b", 64)
				m["business_requirement_refs"] = []any{ref, ref}
			},
			reason: "duplicate business_requirement_refs reference",
		},
		{
			name: "operational-profile-invalid-ref",
			mutate: func(m map[string]any) {
				m["operational_profile_ref"] = "bad-ref"
			},
			reason: "operational_profile_ref",
		},
		{
			name: "excluded-scope-refs-not-array",
			mutate: func(m map[string]any) {
				m["excluded_scope_refs"] = "not-array"
			},
			reason: "excluded_scope_refs array",
		},
		{
			name: "limitations-not-array",
			mutate: func(m map[string]any) {
				m["limitations"] = 123
			},
			reason: "limitations array",
		},
		{
			name: "limitations-empty-string",
			mutate: func(m map[string]any) {
				m["limitations"] = []any{""}
			},
			reason: "invalid limitation entry",
		},
		{
			name: "limitations-duplicate",
			mutate: func(m map[string]any) {
				m["limitations"] = []any{"limit-1", "limit-1"}
			},
			reason: "duplicate limitation entry",
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
				ctrl := workScopeControl(t)
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

			err := ValidateWorkScope(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertWorkScopeRejected(t, err, tc.reason)
		})
	}
}

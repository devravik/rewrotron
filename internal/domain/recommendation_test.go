package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestRecommendationRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0
	reasonMap := map[string]string{
		"reject-untyped-recommendation":          "invalid type",
		"reject-recommendation-without-evidence": "evidence_refs array",
	}

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/project/recommendation.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRecommendation(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid recommendation for %s, got: %v", fixture.Name, err)
				}
				return
			}
			expectedReason, ok := reasonMap[fixture.Name]
			if !ok {
				t.Fatalf("missing expected reason mapping for invalid fixture: %s", fixture.Name)
			}
			assertRecommendationRejected(t, err, expectedReason)
		})
	}
	if count != 3 {
		t.Fatalf("expected 3 recommendation fixtures, found %d", count)
	}
}

func recommendationControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "pending-runtime-recommendation" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing pending-runtime-recommendation fixture")
	return nil
}

func assertRecommendationRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var recErr *RecommendationError
	if !errors.As(err, &recErr) || recErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid recommendation: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func TestRecommendationValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := recommendationControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRecommendation(canonical); err != nil {
		t.Fatalf("expected valid control recommendation, got: %v", err)
	}
}

func TestRecommendationMutations(t *testing.T) {
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
			name: "missing-recommendation-id",
			mutate: func(m map[string]any) {
				delete(m, "recommendation_id")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-recommendation-id",
			mutate: func(m map[string]any) {
				m["recommendation_id"] = "REC_UPPER"
			},
			reason: "recommendation identity",
		},
		{
			name: "missing-revision",
			mutate: func(m map[string]any) {
				delete(m, "revision")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-revision-zero",
			mutate: func(m map[string]any) {
				m["revision"] = 0
			},
			reason: "revision",
		},
		{
			name: "invalid-revision-negative",
			mutate: func(m map[string]any) {
				m["revision"] = -1
			},
			reason: "revision",
		},
		{
			name: "missing-target-spec-hash",
			mutate: func(m map[string]any) {
				delete(m, "target_spec_hash")
			},
			reason: "missing required field",
		},
		{
			name: "invalid-target-spec-hash",
			mutate: func(m map[string]any) {
				m["target_spec_hash"] = "short"
			},
			reason: "target_spec_hash",
		},
		{
			name: "empty-category",
			mutate: func(m map[string]any) {
				m["category"] = "  "
			},
			reason: "category",
		},
		{
			name: "invalid-type",
			mutate: func(m map[string]any) {
				m["type"] = "CRITICAL"
			},
			reason: "invalid type",
		},
		{
			name: "invalid-current-type",
			mutate: func(m map[string]any) {
				m["current"] = 123
			},
			reason: "current",
		},
		{
			name: "empty-proposed-action",
			mutate: func(m map[string]any) {
				m["proposed_action"] = ""
			},
			reason: "proposed_action",
		},
		{
			name: "reasons-not-array",
			mutate: func(m map[string]any) {
				m["reasons"] = "reason"
			},
			reason: "reasons array",
		},
		{
			name: "reasons-empty",
			mutate: func(m map[string]any) {
				m["reasons"] = []any{}
			},
			reason: "reasons array",
		},
		{
			name: "reasons-empty-string",
			mutate: func(m map[string]any) {
				m["reasons"] = []any{""}
			},
			reason: "invalid reasons entry",
		},
		{
			name: "reasons-duplicate",
			mutate: func(m map[string]any) {
				m["reasons"] = []any{"r1", "r1"}
			},
			reason: "duplicate reasons entry",
		},
		{
			name: "invalid-risk",
			mutate: func(m map[string]any) {
				m["risk"] = "EXTREME"
			},
			reason: "invalid risk",
		},
		{
			name: "affected-ids-not-array",
			mutate: func(m map[string]any) {
				m["affected_ids"] = "id"
			},
			reason: "affected_ids array",
		},
		{
			name: "affected-ids-invalid-id",
			mutate: func(m map[string]any) {
				m["affected_ids"] = []any{"ID_UPPER"}
			},
			reason: "invalid affected_ids identity",
		},
		{
			name: "affected-ids-duplicate",
			mutate: func(m map[string]any) {
				m["affected_ids"] = []any{"id-1", "id-1"}
			},
			reason: "duplicate affected_ids identity",
		},
		{
			name: "source-files-not-array",
			mutate: func(m map[string]any) {
				m["source_files"] = 123
			},
			reason: "source_files array",
		},
		{
			name: "source-files-duplicate",
			mutate: func(m map[string]any) {
				m["source_files"] = []any{"f1", "f1"}
			},
			reason: "duplicate source_files entry",
		},
		{
			name: "evidence-refs-not-array",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = false
			},
			reason: "evidence_refs array",
		},
		{
			name: "evidence-refs-empty",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = []any{}
			},
			reason: "evidence_refs array",
		},
		{
			name: "evidence-refs-invalid-ref",
			mutate: func(m map[string]any) {
				m["evidence_refs"] = []any{"bad-ref"}
			},
			reason: "invalid evidence_refs reference",
		},
		{
			name: "evidence-refs-duplicate",
			mutate: func(m map[string]any) {
				ref := "sha256:" + strings.Repeat("a", 64)
				m["evidence_refs"] = []any{ref, ref}
			},
			reason: "duplicate evidence_refs reference",
		},
		{
			name: "required-actions-not-array",
			mutate: func(m map[string]any) {
				m["required_actions"] = 42
			},
			reason: "required_actions array",
		},
		{
			name: "required-actions-duplicate",
			mutate: func(m map[string]any) {
				m["required_actions"] = []any{"act1", "act1"}
			},
			reason: "duplicate required_actions entry",
		},
		{
			name: "alternatives-not-array",
			mutate: func(m map[string]any) {
				m["alternatives"] = "none"
			},
			reason: "alternatives array",
		},
		{
			name: "alternatives-duplicate",
			mutate: func(m map[string]any) {
				m["alternatives"] = []any{"alt1", "alt1"}
			},
			reason: "duplicate alternatives entry",
		},
		{
			name: "invalid-decision",
			mutate: func(m map[string]any) {
				m["decision"] = "UNKNOWN"
			},
			reason: "invalid decision",
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
				ctrl := recommendationControl(t)
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

			err := ValidateRecommendation(raw)
			if err == nil {
				t.Fatalf("expected error for case %s, got nil", tc.name)
			}
			assertRecommendationRejected(t, err, tc.reason)
		})
	}
}

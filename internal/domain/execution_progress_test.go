package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestExecutionProgressRepositoryFixtures(t *testing.T) {
	t.Parallel()
	count := 0

	for _, fixture := range admissionFixtures(t) {
		if fixture.Schema != "https://rewrotron.dev/specs/execution/progress.schema.json" {
			continue
		}
		count++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			err := ValidateExecutionProgress(fixture.Instance)
			if fixture.Valid {
				if err != nil {
					t.Fatalf("expected valid execution progress for %s, got: %v", fixture.Name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected invalid execution progress for %s, got nil error", fixture.Name)
			}
		})
	}
	if count != 6 {
		t.Fatalf("expected 6 execution progress fixtures, found %d", count)
	}
}

func executionProgressControl(t *testing.T) map[string]any {
	t.Helper()
	for _, fixture := range admissionFixtures(t) {
		if fixture.Name == "measured-current-case-progress" {
			var record map[string]any
			if err := json.Unmarshal(fixture.Instance, &record); err != nil {
				t.Fatal(err)
			}
			return record
		}
	}
	t.Fatal("missing measured-current-case-progress fixture")
	return nil
}

func assertExecutionProgressRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var progErr *ExecutionProgressError
	if !errors.As(err, &progErr) || progErr.Reason != reason {
		t.Fatalf("wanted reason %q, got: %v", reason, err)
	}
	if err.Error() != "invalid execution progress: "+reason {
		t.Fatalf("error formatted unexpectedly: %v", err)
	}
}

func validateExecutionProgressControl(t *testing.T, record map[string]any) error {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return ValidateExecutionProgress(raw)
}

func TestExecutionProgressValidRecord(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	if err := validateExecutionProgressControl(t, ctrl); err != nil {
		t.Fatalf("control record should be valid: %v", err)
	}
}

func TestExecutionProgressMissingRequiredFields(t *testing.T) {
	t.Parallel()
	required := []string{
		"metric_id",
		"definition",
		"numerator",
		"denominator",
		"unit",
		"percent",
		"scope_hash",
		"basis_ref",
		"accounting_revision",
		"evidence_sequence",
		"freshness",
	}

	for _, field := range required {
		ctrl := executionProgressControl(t)
		delete(ctrl, field)
		assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "missing required field")
	}
}

func TestExecutionProgressExtraFields(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["unexpected"] = "field"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "object fields")
}

func TestExecutionProgressInvalidMetricID(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["metric_id"] = "Invalid_Metric"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "metric identity")

	ctrl = executionProgressControl(t)
	ctrl["metric_id"] = ""
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "metric identity")

	ctrl = executionProgressControl(t)
	ctrl["metric_id"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "metric identity")
}

func TestExecutionProgressInvalidDefinition(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["definition"] = "   "
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "definition")

	ctrl = executionProgressControl(t)
	ctrl["definition"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "definition")
}

func TestExecutionProgressInvalidNumerator(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["numerator"] = -1
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "numerator")

	ctrl = executionProgressControl(t)
	ctrl["numerator"] = "not a number"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "numerator")
}

func TestExecutionProgressInvalidDenominator(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["denominator"] = -5
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "denominator")

	ctrl = executionProgressControl(t)
	ctrl["denominator"] = "invalid"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "denominator")
}

func TestExecutionProgressInvalidUnit(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["unit"] = ""
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "unit")

	ctrl = executionProgressControl(t)
	ctrl["unit"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "unit")
}

func TestExecutionProgressInvalidPercent(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["percent"] = -0.5
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "percent")

	ctrl = executionProgressControl(t)
	ctrl["percent"] = 100.1
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "percent")

	ctrl = executionProgressControl(t)
	ctrl["percent"] = "100"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "percent")
}

func TestExecutionProgressInvalidScopeHash(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["scope_hash"] = "short"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "scope_hash")

	ctrl = executionProgressControl(t)
	ctrl["scope_hash"] = "000000000000000000000000000000000000000000000000000000000000000Z"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "scope_hash")
}

func TestExecutionProgressInvalidBasisRef(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["basis_ref"] = "invalid-ref"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "basis_ref")
}

func TestExecutionProgressInvalidAccountingRevision(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["accounting_revision"] = 0
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "accounting_revision")

	ctrl = executionProgressControl(t)
	ctrl["accounting_revision"] = -2
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "accounting_revision")

	ctrl = executionProgressControl(t)
	ctrl["accounting_revision"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "accounting_revision")
}

func TestExecutionProgressInvalidEvidenceSequence(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["evidence_sequence"] = -1
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "evidence_sequence")

	ctrl = executionProgressControl(t)
	ctrl["evidence_sequence"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "evidence_sequence")
}

func TestExecutionProgressInvalidFreshness(t *testing.T) {
	t.Parallel()
	ctrl := executionProgressControl(t)
	ctrl["freshness"] = "INVALID"
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "freshness")

	ctrl = executionProgressControl(t)
	ctrl["freshness"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "freshness")
}

func TestExecutionProgressAccountingInvariants(t *testing.T) {
	t.Parallel()

	// Zero denominator requires zero numerator
	ctrl := executionProgressControl(t)
	ctrl["denominator"] = 0
	ctrl["numerator"] = 5
	ctrl["percent"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "zero denominator requires zero numerator")

	// Zero denominator requires null percent
	ctrl = executionProgressControl(t)
	ctrl["denominator"] = 0
	ctrl["numerator"] = 0
	ctrl["percent"] = 0
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "zero denominator requires null percent")

	// Null denominator requires null percent
	ctrl = executionProgressControl(t)
	ctrl["denominator"] = nil
	ctrl["numerator"] = 10
	ctrl["percent"] = 50
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "null denominator requires null percent")

	// Null numerator requires null percent
	ctrl = executionProgressControl(t)
	ctrl["denominator"] = 10
	ctrl["numerator"] = nil
	ctrl["percent"] = 50
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "null numerator requires null percent")

	// Numeric percent requires denominator >= 1
	ctrl = executionProgressControl(t)
	ctrl["percent"] = 50
	ctrl["denominator"] = 0
	ctrl["numerator"] = 0
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "zero denominator requires null percent")

	// Numeric percent requires scope_hash
	ctrl = executionProgressControl(t)
	ctrl["scope_hash"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "numeric percent requires scope_hash")

	// Numeric percent requires basis_ref
	ctrl = executionProgressControl(t)
	ctrl["basis_ref"] = nil
	assertExecutionProgressRejected(t, validateExecutionProgressControl(t, ctrl), "numeric percent requires basis_ref")
}

func TestExecutionProgressJSONParsing(t *testing.T) {
	t.Parallel()

	// Invalid UTF-8
	err := ValidateExecutionProgress([]byte{0xff, 0xfe})
	assertExecutionProgressRejected(t, err, "invalid UTF-8")

	// Duplicate keys
	dupJSON := []byte(`{
		"metric_id": "test",
		"metric_id": "test2"
	}`)
	assertExecutionProgressRejected(t, ValidateExecutionProgress(dupJSON), "duplicate object key")

	// Trailing tokens
	ctrl := executionProgressControl(t)
	raw, err := json.Marshal(ctrl)
	if err != nil {
		t.Fatal(err)
	}
	trailing := append(raw, []byte(` extra`)...)
	assertExecutionProgressRejected(t, ValidateExecutionProgress(trailing), "trailing input")

	// Not an object
	assertExecutionProgressRejected(t, ValidateExecutionProgress([]byte(`[1, 2, 3]`)), "object fields")
}

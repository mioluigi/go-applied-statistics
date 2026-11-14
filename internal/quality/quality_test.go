package quality

import (
	"strings"
	"testing"
)

// TestRepositoryFollowsTheRules is the gate: the whole repository must have no violations.
func TestRepositoryFollowsTheRules(t *testing.T) {
	violations, err := Check("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, violation := range violations {
		t.Error(violation)
	}
}

// TestCheckFindsEachKindOfViolation proves the gate can fail: a sample file breaks each rule once.
func TestCheckFindsEachKindOfViolation(t *testing.T) {
	violations, err := CheckIncludingTestdata("testdata/bad")
	if err != nil {
		t.Fatal(err)
	}
	wanted := []string{
		"exported function Undocumented has no doc comment",
		"exported UndocumentedType has no doc comment",
		"exported UndocumentedConstant has no doc comment",
		"function tooLong is 63 lines long (limit 60)",
	}
	joined := strings.Join(violations, "\n")
	for _, want := range wanted {
		if !strings.Contains(joined, want) {
			t.Errorf("missing violation %q in:\n%s", want, joined)
		}
	}
	if len(violations) != len(wanted) {
		t.Errorf("got %d violations, want %d:\n%s", len(violations), len(wanted), joined)
	}
}

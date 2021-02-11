package apache

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	testcases := []struct {
		input  string
		parsed *ParsedId
	}{
		{"Scala", nil},
		{"apache/hadoop", &ParsedId{
			ProjectName: "hadoop",
			Component:   "",
		}},
		{"apache/spark/SparkR", &ParsedId{
			ProjectName: "spark",
			Component:   "SparkR",
		}},
	}
	for _, tc := range testcases {
		parsed, err := parse(tc.input)
		if tc.parsed == nil {
			// expect err
			if err == nil {
				t.Errorf("Expected parse failure on %q, got: %v", tc.input, parsed)
			}
			continue
		}

		if diffstr := cmp.Diff(&parsed, tc.parsed); diffstr != "" {
			t.Errorf("Unexpected diff parsing %q: %s", tc.input, diffstr)
		}
	}
}

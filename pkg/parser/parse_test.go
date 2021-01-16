package parser

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tcs := []struct {
		input  string
		expect queryIntermediate
	}{
		{"consul", queryIntermediate{
			SoftwareId:  "consul",
			VerRangeStr: "",
			Prerelease:  false,
		}},
		{"go@123", queryIntermediate{
			SoftwareId:  "go",
			VerRangeStr: ">=123.0.0 <124.0.0 ",
			Prerelease:  false,
		}},
		{"go@123.456", queryIntermediate{
			SoftwareId:  "go",
			VerRangeStr: ">=123.456.0 <123.457.0 ",
			Prerelease:  false,
		}},
		{"go@123.456.789", queryIntermediate{
			SoftwareId:  "go",
			VerRangeStr: "=123.456.789 ",
			Prerelease:  false,
		}},
		{"go>1.15.0", queryIntermediate{
			SoftwareId:  "go",
			VerRangeStr: ">1.15.0 ",
			Prerelease:  false,
		}},
		{"github.com/foo/bar@v4:prerelease", queryIntermediate{
			SoftwareId:  "github.com/foo/bar",
			VerRangeStr: ">=4.0.0 <5.0.0 ",
			Prerelease:  true,
		}},
	}

	for _, tc := range tcs {
		actual, err := parseInternal(tc.input)
		if err != nil {
			t.Errorf("Failed to parse %q. Error: %v", tc.input, err)
			continue
		}

		if diffstr := cmp.Diff(actual, &tc.expect); diffstr != "" {
			t.Errorf("Parsing %q lead to unexpected result: %s", tc.input, diffstr)
		}
		if actual.VerRangeStr != "" {
			if _, err := semver.ParseRange(actual.VerRangeStr); err != nil {
				t.Errorf("Failed to ParseRange(%q). Error: %v", actual.VerRangeStr, err)
			}
		}
	}
}

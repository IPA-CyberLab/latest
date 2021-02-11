package parser

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
)

func TestParseComponentAndVersion(t *testing.T) {
	tcs := []struct {
		input     string
		component string
		verstr    string
	}{
		{"apache-wicket-8.11.0", "apache-wicket", "8.11.0"},
		{"go1.15", "go", "1.15"},
		{"SparkR_3.0.1", "SparkR", "3.0.1"},
	}

	for _, tc := range tcs {
		component, ver, err := ParseComponentAndVersion(tc.input)
		if err != nil {
			t.Errorf("Parse %q failed: %v", tc.input, err)
			continue
		}

		if component != tc.component {
			t.Errorf("%q: Expected component %q, got %q", tc.input, tc.component, component)
		}

		verExpected, err := semver.ParseTolerant(tc.verstr)
		if err != nil {
			t.Errorf("Failed to parse test case ver %q: %v", tc.verstr, err)
			continue
		}

		if !verExpected.Equals(ver) {
			t.Errorf("%q: Expected ver %v, got %v", tc.input, verExpected, ver)
		}
	}
}

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

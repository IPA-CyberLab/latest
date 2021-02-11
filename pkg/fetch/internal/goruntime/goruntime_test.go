package goruntime_test

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/goruntime"
	"github.com/IPA-CyberLab/latest/pkg/releases"
)

func init() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
}

func TestMatch(t *testing.T) {
	testcases := []struct {
		input       string
		expectMatch bool
	}{
		{"go", true},
		{"Go", true},
		{"golang", true},
		{"Golang", true},
		{"Scala", false},
	}

	for _, tc := range testcases {
		actual := goruntime.Match(tc.input)
		if actual != tc.expectMatch {
			t.Errorf("Match(%q) expected %t actual %t", tc.input, tc.expectMatch, actual)
		}
	}
}

func TestParse(t *testing.T) {
	testcases := []struct {
		jsonstr  string
		expected releases.Releases
	}{
		{
			`[{
				"version": "go1.15.6",
				"stable": true,
				"files": [
				 {
					"filename": "go1.15.6.src.tar.gz",
					"os": "",
					"arch": "",
					"version": "go1.15.6",
					"sha256": "890bba73c5e2b19ffb1180e385ea225059eb008eb91b694875dd86ea48675817",
					"size": 23019337,
					"kind": "source"
				 },
				 {
					"filename": "go1.15.6.darwin-amd64.tar.gz",
					"os": "darwin",
					"arch": "amd64",
					"version": "go1.15.6",
					"sha256": "940a73b45993a3bae5792cf324140dded34af97c548af4864d22fd6d49f3bd9f",
					"size": 122234016,
					"kind": "archive"
				 },
				 {
					"filename": "go1.15.6.windows-amd64.msi",
					"os": "windows",
					"arch": "amd64",
					"version": "go1.15.6",
					"sha256": "bedc8243116297d14a8ba15fcb280e7419dcf344a957263e2c815d74d463397e",
					"size": 120832000,
					"kind": "installer"
				 }
				]
			 }]`,
			releases.Releases{
				{"go1.15.6", semver.MustParse("1.15.6"), false, []string{
					"https://dl.google.com/go/go1.15.6.src.tar.gz",
					"https://dl.google.com/go/go1.15.6.darwin-amd64.tar.gz",
					"https://dl.google.com/go/go1.15.6.windows-amd64.msi",
				}},
			},
		},
		{
			// what if version does not end with .0?
			`[{
				"version": "go1.15",
				"stable": true,
				"files": [{
					"filename": "go1.15.src.tar.gz",
					"os": "",
					"arch": "",
					"version": "go1.15",
					"sha256": "69438f7ed4f532154ffaf878f3dfd83747e7a00b70b3556eddabf7aaee28ac3a",
					"size": 23002901,
					"kind": "source"
				 }]
			 }]`,
			releases.Releases{
				{"go1.15", semver.MustParse("1.15.0"), false, []string{
					"https://dl.google.com/go/go1.15.src.tar.gz",
				}},
			},
		},
	}

	for _, tc := range testcases {
		rs, err := goruntime.Parse([]byte(tc.jsonstr))
		if err != nil {
			t.Errorf("Failed to parse test case: %q", tc.jsonstr)
			continue
		}

		if diffstr := cmp.Diff(rs, tc.expected); diffstr != "" {
			t.Errorf("Unexpected diff: %s", diffstr)
		}
	}
}

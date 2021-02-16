package releases

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	"go.uber.org/zap"
)

type Release struct {
	OriginalName string         `json:"original_name"`
	Version      semver.Version `json:"version"`
	Prerelease   bool           `json:"prerelease"`
	AssetURLs    []string       `json:"asset_urls"`
}

type Releases []Release

var NotFoundErr = errors.New("No matching Release found.")
var AssetNotFoundErr = errors.New("No matching asset found.")

func (rs Releases) SelectAll(vrange semver.Range) Releases {
	selected := make(Releases, 0)

	for _, r := range rs {
		if vrange != nil && !vrange(r.Version) {
			continue
		}

		selected = append(selected, r)
	}

	return selected
}

func (rs Releases) Select(vrange semver.Range) (Release, error) {
	for _, r := range rs {
		if vrange != nil && !vrange(r.Version) {
			continue
		}

		return r, nil
	}
	return Release{}, NotFoundErr
}

func (rs Releases) RemovePrerelease() Releases {
	ret := make(Releases, 0, len(rs))

	for _, r := range rs {
		if r.Prerelease {
			continue
		}

		ret = append(ret, r)
	}
	return ret
}

func filterIfMatches(ss []string, needle string) []string {
	zap.S().Debugf("asset filter %q", needle)

	var invert bool
	if needle[0] == '!' {
		invert = true
		needle = needle[1:]
	} else {
		invert = false
	}

	filtered := make([]string, 0, len(ss))
	for _, s := range ss {
		match := strings.Contains(strings.ToLower(s), needle)
		if match == !invert {
			// zap.S().Debugf("%s %t match %s\n", needle, invert, strings.ToLower(s))
			filtered = append(filtered, s)
		}
	}

	if len(filtered) > 0 {
		return filtered
	} else {
		return ss
	}
}

var archAlias = map[string]string{
	"amd64": "x86_64",
}

func (r Release) PickAsset() (string, error) {
	if len(r.AssetURLs) == 0 {
		return "", AssetNotFoundErr
	}

	filters := []string{
		"!.txt",
		"!.sha256sum",
		"!.asc",
		runtime.GOOS,
		runtime.GOARCH,
	}
	if alias, ok := archAlias[runtime.GOARCH]; ok {
		filters = append(filters, alias)
	}

	filtered := r.AssetURLs
	for _, f := range filters {
		filtered = filterIfMatches(filtered, f)
	}
	if len(filtered) != 1 {
		return "", fmt.Errorf("Too many matches: %v", filtered)
	}
	return filtered[0], nil
}

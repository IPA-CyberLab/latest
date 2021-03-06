package releases

import (
	"errors"
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

func filter(ss []string, needle string) []string {
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
	return filtered
}

func (r *Release) FilterAssets(needle string) {
	r.AssetURLs = filter(r.AssetURLs, needle)
}

func filterIfMatches(ss []string, needle string) []string {
	filtered := filter(ss, needle)

	if len(filtered) > 0 {
		return filtered
	} else {
		return ss
	}
}

var archAlias = map[string]string{
	"amd64": "x86_64",
}

func (r *Release) PickAsset() {
	filters := []string{
		"!.txt",
		"!.sha256sum",
		"!.asc",
		"!.log",
		runtime.GOOS,
		runtime.GOARCH,
	}
	if alias, ok := archAlias[runtime.GOARCH]; ok {
		filters = append(filters, alias)
	}

	for _, f := range filters {
		r.AssetURLs = filterIfMatches(r.AssetURLs, f)
	}
}

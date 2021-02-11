package hashicorp

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"go.uber.org/zap"

	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/httpcli"
	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/releases"
)

const HandlerName = "hashicorp"
const endpoint = "https://releases.hashicorp.com"

func getHtml(ctx context.Context, productName string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/", endpoint, productName)
	return httpcli.Get(ctx, url)
}

var Products = map[string]struct{}{
	"consul": {},
}

func Match(softwareId string) bool {
	_, ok := Products[softwareId]
	return ok
}

var reHref = regexp.MustCompile(`href=\"/\w+/([\w\+-\.]+)/\"`)
var reEnt = regexp.MustCompile(`\+ent`)
var assetSuffixes = []string{
	"_SHA256SUMS",
	"_SHA256SUMS.sig",
	"_darwin_amd64.zip",
	"_freebsd_386.zip",
	"_freebsd_amd64.zip",
	"_linux_386.zip",
	"_linux_amd64.zip",
	"_linux_arm64.zip",
	"_linux_armelv5.zip",
	"_linux_armhfv6.zip",
	"_solaris_amd64.zip",
	"_windows_386.zip",
	"_windows_amd64.zip",
}

func Parse(softwareId string, htmlbs []byte) (releases.Releases, error) {
	l := zap.S()

	ms := reHref.FindAllSubmatch(htmlbs, -1)
	rs := make(releases.Releases, 0, len(ms))
	for _, m := range ms {
		versionStr := string(m[1])

		if reEnt.MatchString(versionStr) {
			// FIXME[P3]: support +ent
			continue
		}

		ver, err := parser.ParseVersion(versionStr)
		if err != nil {
			l.Debugf("Failed to parse version: %s", versionStr)
			continue
		}

		assetURLs := make([]string, 0, len(assetSuffixes))
		for _, suffix := range assetSuffixes {
			u := fmt.Sprintf("https://releases.hashicorp.com/%[1]s/%[2]s/%[1]s_%[2]s%[3]s", softwareId, versionStr, suffix)
			assetURLs = append(assetURLs, u)
		}

		r := releases.Release{
			OriginalName: versionStr,
			Version:      ver,
			AssetURLs:    assetURLs,
		}
		rs = append(rs, r)
	}

	return rs, nil
}

func Fetch(ctx context.Context, softwareId string) (releases.Releases, error) {
	if !Match(softwareId) {
		return nil, ferrors.ErrSoftwareIdParseFailed{
			Input:       softwareId,
			HandlerName: HandlerName,
			Err:         nil,
		}
	}

	bs, err := getHtml(ctx, softwareId)
	if err != nil {
		return nil, err
	}

	rs, err := Parse(softwareId, bs)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].Version.GT(rs[j].Version)
	})

	return rs, nil
}

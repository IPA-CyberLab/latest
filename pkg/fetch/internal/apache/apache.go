package apache

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/httpcli"
	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/releases"
	"go.uber.org/zap"
)

const HandlerName = "apache"

var reId = regexp.MustCompile(`^apache/(\w+)(/(\w+))?$`)

type ParsedId struct {
	ProjectName string
	Component   string
}

func parse(softwareId string) (ParsedId, error) {
	ms := reId.FindStringSubmatch(softwareId)
	if len(ms) == 0 {
		return ParsedId{}, ferrors.ErrSoftwareIdParseFailed{
			Input:       softwareId,
			HandlerName: HandlerName,
			Err:         nil,
		}
	}

	return ParsedId{ProjectName: ms[1], Component: ms[3]}, nil
}

func Match(softwareId string) bool {
	_, err := parse(softwareId)
	return err != nil
}

const endpoint = "https://projects.apache.org/json/foundation/releases.json"

func Fetch(ctx context.Context, softwareId string) (releases.Releases, error) {
	l := zap.S()

	parsed, err := parse(softwareId)
	if err != nil {
		return nil, err
	}

	bs, err := httpcli.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var db map[string]map[string]string
	if err := json.Unmarshal(bs, &db); err != nil {
		return nil, err
	}

	projReleases, ok := db[parsed.ProjectName]
	if !ok {
		return nil, fmt.Errorf("Failed to find an Apache project named %q", parsed.ProjectName)
	}

	rs := make(releases.Releases, 0)

	for releaseName, dateStr := range projReleases {
		_ = dateStr // FIXME

		component, ver, err := parser.ParseComponentAndVersion(releaseName)
		if err != nil {
			l.Debugf("Failed to parse release version of %q: %v", releaseName, err)
			continue
		}

		if parsed.Component != "" && component != parsed.Component {
			continue
		}

		r := releases.Release{
			OriginalName: releaseName,
			Version:      ver,
			Prerelease:   false,
			AssetURLs:    nil, // FIXME
		}
		rs = append(rs, r)
	}

	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].Version.GT(rs[j].Version)
	})

	return rs, nil
}

package goruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"

	"go.uber.org/zap"

	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/httpcli"
	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/releases"
)

const endpoint = "https://golang.org/dl/?mode=json&include=all"

func getJson(ctx context.Context) ([]byte, error) {
	hc := httpcli.HttpClient

	req, err := http.NewRequestWithContext((ctx), "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct http.Request: %w", err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to issue request to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body of %s: %w", endpoint, err)
	}

	return bs, nil
}

var reGoSoftwareId = regexp.MustCompile(`[Gg]o(lang)?`)

func Match(softwareId string) bool {
	return reGoSoftwareId.MatchString(softwareId)
}

func Parse(jsonbs []byte) (releases.Releases, error) {
	l := zap.S()

	type File struct {
		Filename string `json:"filename"`
		Os       string `json:"os"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
		Sha256   string `json:"sha256"`
		Size     int    `json:"size"`
		Kind     string `json:"kind"`
	}
	type RawRelease struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
		Files   []File `json:"files"`
	}

	var rawrs []RawRelease
	if err := json.Unmarshal(jsonbs, &rawrs); err != nil {
		return nil, fmt.Errorf("Failed to parse response: %w", err)
	}

	rs := make(releases.Releases, 0, len(rawrs))
	for _, rawr := range rawrs {
		// if rawr.Stable {
		// 	continue
		// }

		r := releases.Release{
			Prerelease: !rawr.Stable,
			AssetURLs:  make([]string, 0, len(rawr.Files)),
		}

		if ver, err := parser.ParseVersion(rawr.Version); err == nil {
			r.OriginalName = rawr.Version
			r.Version = ver
		} else {
			l.Warnf("Failed to parse Go version %q", rawr.Version)
			continue
		}

		for _, a := range rawr.Files {
			url := fmt.Sprintf("https://dl.google.com/go/%s", a.Filename)
			r.AssetURLs = append(r.AssetURLs, url)
		}

		rs = append(rs, r)
	}
	return rs, nil
}

func Fetch(ctx context.Context, softwareId string) (releases.Releases, error) {
	if !Match(softwareId) {
		return nil, ferrors.ErrSoftwareIdParseFailed{
			Input:       softwareId,
			HandlerName: "goruntime",
			Err:         nil,
		}
	}

	bs, err := getJson(ctx)
	if err != nil {
		return nil, err
	}

	rs, err := Parse(bs)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].Version.GT(rs[j].Version)
	})

	return rs, nil
}

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/httpcli"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/scrapeutil"
	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/releases"
)

var apiResultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "latest",
	Subsystem: "github",
	Name:      "queries_total",
	Help:      "Total number of the GitHub API queried by its result status code.",
}, []string{"status"})
var apiSecondsHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: "latest",
	Subsystem: "github",
	Name:      "duration_seconds",

	Help: "Seconds took to process GitHub API call.",
})

func githubHttpGet(ctx context.Context, url string) ([]byte, error) {
	start := time.Now()
	defer func() {
		apiSecondsHistogram.Observe(time.Since(start).Seconds())
	}()

	l := zap.S()
	l.Debugf("github API call: %v", url)

	hc := httpcli.HttpClient
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct http.Request: %w", err)
	}

	req.Header = http.Header{}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to issue request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	apiResultTotal.WithLabelValues(strconv.Itoa(resp.StatusCode)).Inc()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Github API returned status %s", resp.Status)
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body of %s: %w", url, err)
	}

	return bs, nil
}

var reGithub = regexp.MustCompile(`^github.com/([A-z0-9]+-?[A-z0-9]*)/([A-z0-9\-_]+)$`)

func Fetch(ctx context.Context, softwareId string) (releases.Releases, error) {
	l := zap.S()

	ms := reGithub.FindStringSubmatch(softwareId)
	if len(ms) == 0 {
		return nil, ferrors.ErrSoftwareIdParseFailed{
			Input:       softwareId,
			HandlerName: "github",
			Err:         nil,
		}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", ms[1], ms[2])

	bs, err := githubHttpGet(ctx, url)
	if err != nil {
		return nil, err
	}

	type Assets struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	}

	type RawRelease struct {
		Name       string   `json:"name"`
		TagName    string   `json:"tag_name"`
		Draft      bool     `json:"draft"`
		Prerelease bool     `json:"prerelease"`
		Assets     []Assets `json:"assets"`
		Body       string   `json:"body"`
	}

	var rawrs []RawRelease
	if err := json.Unmarshal(bs, &rawrs); err != nil {
		return nil, fmt.Errorf("Failed to parse response: %w", err)
	}

	rs := make(releases.Releases, 0, len(rawrs))
	for _, rawr := range rawrs {
		if rawr.Draft {
			continue
		}

		r := releases.Release{
			Prerelease: rawr.Prerelease,
			AssetURLs:  make([]string, 0, len(rawr.Assets)),
		}

		if ver, err := parser.ParseVersion(rawr.TagName); err == nil {
			r.OriginalName = rawr.TagName
			r.Version = ver
		} else if ver, err := parser.ParseVersion(rawr.Name); err == nil {
			r.OriginalName = rawr.Name
			r.Version = ver
		} else {
			l.Warnf("Failed to parse version from release name %q tagname %q", rawr.TagName, rawr.Name)
			continue
		}
		// l.Debugf("Parse version from release name %q tagname %q -> %v", rawr.TagName, rawr.Name, r.Version)

		r.AssetURLs = scrapeutil.ScrapeLinks(rawr.Body)
		for _, a := range rawr.Assets {
			r.AssetURLs = append(r.AssetURLs, a.BrowserDownloadURL)
		}

		rs = append(rs, r)
	}
	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].Version.GT(rs[j].Version)
	})

	return rs, nil
}

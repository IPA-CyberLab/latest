package maven

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"

	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/httpcli"
	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/releases"
)

// FIXME: make configurable
const MavenRoot = "https://repo1.maven.org/maven2"

func Fetch(ctx context.Context, softwareId string) (releases.Releases, error) {
	l := zap.S()

	// m2:[groupId]:[artifactId]:[classifier]:[extension]
	ss := strings.Split(softwareId, ":")
	if ss[0] != "m2" {
		return nil, ferrors.ErrSoftwareIdParseFailed{
			Input:       softwareId,
			HandlerName: "maven",
			Err:         errors.New("bad prefix"),
		}
	}
	if len(ss) < 3 || len(ss) > 5 {
		return nil, ferrors.ErrSoftwareIdParseFailed{
			Input:       softwareId,
			HandlerName: "maven",
			Err:         errors.New("unexpected number of :s"),
		}
	}
	groupId, artifactId := ss[1], ss[2]
	classifier, extension := "", "jar"
	if len(ss) > 3 {
		classifier = ss[3]
	}
	if len(ss) > 4 {
		extension = ss[4]
	}

	projectRoot := fmt.Sprintf("%s/%s/%s", MavenRoot, strings.ReplaceAll(groupId, ".", "/"), artifactId)
	metadataUrl := fmt.Sprintf("%s/maven-metadata.xml", projectRoot)
	l.Debugf("url: %s", metadataUrl)

	metadataXml, err := httpcli.Get(ctx, metadataUrl)
	if err != nil {
		return nil, fmt.Errorf("Failed to get maven-metadata.xml: %w", err)
	}

	// http://maven.apache.org/ref/3.3.9/maven-repository-metadata/repository-metadata.html

	type Metadata struct {
		Latest      string   `xml:"versioning>latest"`
		Release     string   `xml:"versioning>release"`
		Versions    []string `xml:"versioning>versions>version"`
		LastUpdated string   `xml:"versioning>lastUpdated"`
	}

	var metadata Metadata
	if err := xml.Unmarshal(metadataXml, &metadata); err != nil {
		return nil, fmt.Errorf("Failed to parse maven-metadata.xml: %w", err)
	}

	rs := make(releases.Releases, 0, len(metadata.Versions))
	l.Debugf("metadata: %+v", metadata)
	l.Debugf("classifier %q extension %q", classifier, extension)

	/*
		verLatest, err := parser.ParseVersion(metadata.Latest)
		if err != nil {
			l.Warnf("Failed to parse latest version %q: %v", metadata.Latest, err)
		}
	*/

	verRelease, err := parser.ParseVersion(metadata.Release)
	if err != nil {
		l.Warnf("Failed to parse release version %q: %v", metadata.Release, err)
	}

	for _, versionStr := range metadata.Versions {
		version, err := parser.ParseVersion(versionStr)
		if err != nil {
			l.Warnf("Failed to parse version %q: %v", versionStr, err)
		}

		var filename string
		if classifier == "" {
			filename = fmt.Sprintf("%s-%s.%s", artifactId, versionStr, extension)
		} else {
			filename = fmt.Sprintf("%s-%s-%s.%s", artifactId, versionStr, classifier, extension)
		}
		assetURL := fmt.Sprintf("%s/%s/%s", projectRoot, versionStr, filename)

		r := releases.Release{
			OriginalName: versionStr,
			Version:      version,
			Prerelease:   version.GT(verRelease),
			AssetURLs:    []string{assetURL},
		}
		rs = append(rs, r)
	}

	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].Version.GT(rs[j].Version)
	})

	return rs, nil
}

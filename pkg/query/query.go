package query

import (
	"context"

	"github.com/blang/semver/v4"

	"github.com/IPA-CyberLab/latest/pkg/releases"
)

type Query struct {
	SoftwareId string
	VerRange   semver.Range
	Prerelease bool
}

type Fetcher interface {
	Fetch(ctx context.Context, softwareId string) (rs releases.Releases, err error)
}

func (q *Query) Execute(ctx context.Context, fetcher Fetcher) (releases.Releases, error) {
	rs, err := fetcher.Fetch(ctx, q.SoftwareId)
	if err != nil {
		return nil, err
	}

	rs = rs.SelectAll(q.VerRange)

	if !q.Prerelease {
		rs = rs.RemovePrerelease()
	}

	return rs, nil
}

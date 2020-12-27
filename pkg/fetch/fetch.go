package fetch

import (
	"context"
	"errors"

	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/github"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/goruntime"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/hashicorp"
	"github.com/IPA-CyberLab/latest/pkg/releases"
	"go.uber.org/zap"
)

var fetchImpls = []func(ctx context.Context, softwareId string) (releases.Releases, error){
	hashicorp.Fetch,
	goruntime.Fetch,
	github.Fetch,
}

type Fetcher interface {
	Fetch(ctx context.Context, softwareId string) (rs releases.Releases, err error)
}

type Direct struct{}

func (Direct) Fetch(ctx context.Context, softwareId string) (rs releases.Releases, err error) {
	for _, fetchImpl := range fetchImpls {
		rs, err = fetchImpl(ctx, softwareId)
		if err == nil {
			return
		} else {
			var parseErr ferrors.ErrSoftwareIdParseFailed
			if !errors.As(err, &parseErr) {
				return
			}
			zap.S().Debugf("%s", err)
		}
	}

	return
}

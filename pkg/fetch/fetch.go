package fetch

import (
	"context"
	"errors"
	"time"

	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/apache"
	ferrors "github.com/IPA-CyberLab/latest/pkg/fetch/internal/errors"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/github"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/goruntime"
	"github.com/IPA-CyberLab/latest/pkg/fetch/internal/hashicorp"
	"github.com/IPA-CyberLab/latest/pkg/releases"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var fetchImpls = []func(ctx context.Context, softwareId string) (releases.Releases, error){
	hashicorp.Fetch,
	goruntime.Fetch,
	apache.Fetch,
	github.Fetch,
}

var directSecondsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "latest",
	Subsystem: "direct_fetcher",
	Name:      "success_duration_seconds",

	Help: "Seconds took to fetch releases by softwareId. Recorded only on fetch success.",
}, []string{"software"})

type Direct struct{}

func (Direct) Fetch(ctx context.Context, softwareId string) (rs releases.Releases, err error) {
	start := time.Now()
	defer func() {
		if err == nil {
			directSecondsHistogram.WithLabelValues(softwareId).Observe(time.Since(start).Seconds())
		}
	}()

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

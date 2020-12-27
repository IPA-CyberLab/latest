package exporter

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/IPA-CyberLab/latest/pkg/fetch"
)

type Handler struct {
	Fetcher fetch.Fetcher
}

func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	l := zap.S()

	reg := prometheus.NewRegistry()
	releaseVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Namespace: "latest", Name: "release", Help: "Information about a software release."},
		[]string{"software", "version", "semver", "prerelease"})
	reg.MustRegister(releaseVec)

	vals := req.URL.Query()
	qs := vals["q"]

	for _, q := range qs {
		// FIXME[P1]: q -> split to softwareId and range query
		softwareId := q
		rs, err := h.Fetcher.Fetch(req.Context(), softwareId)
		if err != nil {
			l.Infof("%s err: %v\n", q, err)
			continue
		}
		if len(rs) == 0 {
			l.Infof("%s : No releases found.", q)
			continue
		}
		r := rs[0]

		prereleaseInt := "0"
		if r.Prerelease {
			prereleaseInt = "1"
		}

		releaseVec.WithLabelValues(softwareId, r.OriginalName, r.Version.String(), prereleaseInt).Set(1)
	}

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, req)
}

package exporter

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/query"
)

type Handler struct {
	Fetcher query.Fetcher
}

func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	l := zap.S()

	reg := prometheus.NewRegistry()
	releaseVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Namespace: "latest", Name: "release", Help: "Information about a software release."},
		[]string{"query", "software", "version", "semver", "prerelease"})
	reg.MustRegister(releaseVec)

	vals := req.URL.Query()

	for _, qval := range vals["q"] {
		q, err := parser.Parse(qval)
		if err != nil {
			l.Infof("Failed to parse %q err: %v\n", q, err)
			continue
		}

		rs, err := q.Execute(req.Context(), h.Fetcher)
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

		releaseVec.WithLabelValues(qval, q.SoftwareId, r.OriginalName, r.Version.String(), prereleaseInt).Set(1)
	}

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, req)
}

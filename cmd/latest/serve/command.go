package serve

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/IPA-CyberLab/latest/pkg/exporter"
	"github.com/IPA-CyberLab/latest/pkg/fetch"
)

var indexhtml = []byte(`<!DOCTYPE html>
<head>
<style>
.code { font-family: monospace; background: #eee; }
</style>
<body>
<h1><span class="code">latest serve</span> serving queries in <a href="https://prometheus.io/">prometheus</a> format.</h1>
<a href="/probe?q=github.com/IPA-CyberLab/latest">example</a><br>
<a href="/metrics">exporter metrics</a><br>
`)

var Command = &cli.Command{
	Name:  "serve",
	Usage: "Launch a http server that serves queries in prometheus format",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen-addr",
			Usage: "server listen `HOST:ADDR`",
			Value: ":16480",
		},
	},
	Action: func(c *cli.Context) error {
		mux := http.NewServeMux()

		fetcher := fetch.NewCachedFetcher(fetch.Direct{})
		mux.Handle("/probe", exporter.Handler{Fetcher: fetcher})

		prometheus.MustRegister(prometheus.NewBuildInfoCollector())
		mux.Handle("/metrics", promhttp.Handler())

		mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("content-type", "text/plain")
			if _, err := w.Write([]byte("ok\n")); err != nil {
				zap.S().Warnf("Failed to write ok: %v", err)
			}
		})

		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("content-type", "text/html; charset=UTF-8")
			if _, err := w.Write(indexhtml); err != nil {
				zap.S().Warnf("Failed to write index.html: %v", err)

			}
		})

		listenAddr := c.String("listen-addr")
		zap.S().Infof("About to start listening on %s", listenAddr)
		if err := http.ListenAndServe(listenAddr, mux); err != nil {
			return err
		}

		return nil
	},
}

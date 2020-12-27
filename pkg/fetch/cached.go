package fetch

import (
	"context"
	"time"

	"github.com/IPA-CyberLab/latest/pkg/releases"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type cachedFetcher struct {
	backend Fetcher
	reqC    chan request
}

var _ = Fetcher(&cachedFetcher{})

func NewCachedFetcher(backend Fetcher) Fetcher {
	c := &cachedFetcher{
		backend: backend,
		reqC:    make(chan request),
	}
	go c.cacheMain()

	return c
}

var NowImpl func() time.Time = time.Now
var EntryLifetime = 30 * time.Minute

type response struct {
	rs  []releases.Release
	err error
}

type request struct {
	softwareId string
	resC       chan<- response
}

func (c *cachedFetcher) Fetch(ctx context.Context, softwareId string) (rs releases.Releases, err error) {
	resC := make(chan response)
	c.reqC <- request{softwareId: softwareId, resC: resC}
	resp := <-resC
	return resp.rs, resp.err
}

var hitrateVec = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "latest",
	Subsystem: "cached_fetcher",
	Name:      "fetches_total",
	Help:      "Total number of fetches through cached_fetcher by its softwareId and cache hit.",
}, []string{"software", "cache_hit"})

func (c cachedFetcher) cacheMain() {
	l := zap.S()

	type entry struct {
		waitC chan struct{}

		fetchedTime time.Time
		rs          []releases.Release
		err         error
	}
	entries := make(map[string]*entry)

	for req := range c.reqC {
		now := NowImpl()

		e := entries[req.softwareId]
		if e != nil && now.Sub(e.fetchedTime) > EntryLifetime {
			e = nil
		}

		if e != nil {
			<-e.waitC

			l.Debugf("cache hit: returning entry %q from cache.", req.softwareId)
			hitrateVec.WithLabelValues(req.softwareId, "hit").Inc()
			req.resC <- response{rs: e.rs, err: e.err}
			close(req.resC)
			continue
		}
		e = &entry{waitC: make(chan struct{})}
		entries[req.softwareId] = e

		l.Debugf("cache miss: %q new fetch.", req.softwareId)
		hitrateVec.WithLabelValues(req.softwareId, "miss").Inc()
		go func() {
			e.rs, e.err = c.backend.Fetch(context.Background(), req.softwareId)
			e.fetchedTime = NowImpl()
			close(e.waitC)
		}()

		<-e.waitC
		req.resC <- response{rs: e.rs, err: e.err}
		close(req.resC)
	}
}

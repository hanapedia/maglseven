package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/hanapedia/maglseven/pkg/maglev"
)

type HandoffProxy struct {
	destPort     string
	router       *maglev.VersionedRouter
	replicaCount int
	maxJumps     int
}

func NewHandoffProxy(destPort string, router *maglev.VersionedRouter, replicaCount, maxJumps int) *HandoffProxy {
	return &HandoffProxy{
		destPort:     destPort,
		router:       router,
		replicaCount: replicaCount,
		maxJumps:     maxJumps,
	}
}

func (p *HandoffProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-Room-ID")
	clientGenHeader := r.Header.Get(maglev.HeaderGeneration)

	result := p.router.Route(key, clientGenHeader, p.replicaCount, p.maxJumps)

	target, err := url.Parse("http://" + strings.TrimSpace(result.Backend.ID) + ":" + p.destPort)
	if err != nil {
		http.Error(w, "Invalid backend URL", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the Director to inject headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req) // copies method, URL, etc. from incoming request

		// Copy our custom headers to outbound request
		req.Header.Set(maglev.HeaderGeneration, strconv.FormatUint(result.Generation, 10))
		req.Header.Set(maglev.HeaderReplicationPeers, joinPeers(result.Peers))

		if result.RequiresRecovery && result.PrevPrimary != nil {
			req.Header.Set(maglev.HeaderPreviousPrimary, result.PrevPrimary.ID)
			req.Header.Set(maglev.HeaderPreviousPeers, joinPeers(result.PrevPeers))
		}
	}

	proxy.ServeHTTP(w, r)
}

func joinPeers(peers []maglev.Backend) string {
	ids := make([]string, len(peers))
	for i, b := range peers {
		ids[i] = b.ID
	}
	return strings.Join(ids, ",")
}

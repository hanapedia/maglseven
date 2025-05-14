package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/hanapedia/magseven/pkg/dispatcher"
)

type Proxy struct {
	dispatcher *dispatcher.Dispatcher
}

func NewProxy(d *dispatcher.Dispatcher) *Proxy {
	return &Proxy{dispatcher: d}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backendHost := p.dispatcher.Route(r)
	target, err := url.Parse("http://" + strings.TrimSpace(backendHost))
	if err != nil {
		http.Error(w, "Invalid backend URL", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}

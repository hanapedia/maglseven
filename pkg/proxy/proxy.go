package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/hanapedia/maglseven/pkg/dispatcher"
)

type Proxy struct {
	destPort string
	dispatcher *dispatcher.Dispatcher
}

func NewProxy(dp string, d *dispatcher.Dispatcher) *Proxy {
	return &Proxy{destPort: dp, dispatcher: d}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backendHost := p.dispatcher.Route(r)
	target, err := url.Parse("http://" + strings.TrimSpace(backendHost) + ":" + p.destPort)
	if err != nil {
		http.Error(w, "Invalid backend URL", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}

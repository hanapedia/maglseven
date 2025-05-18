package watcher

import (
	"context"
	"net"
	"time"

	"github.com/hanapedia/maglseven/pkg/maglev"
	"github.com/hanapedia/maglseven/pkg/util"
)

type DNSWatcher struct {
	FQDN     string        // e.g., "myapp.default.svc.cluster.local"
	Interval time.Duration // how often to resolve
}

func (w *DNSWatcher) Watch(ctx context.Context, updates chan<- []maglev.Backend) error {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	var lastHash string

	resolveAndSend := func() error {
		ips, err := net.LookupIP(w.FQDN)
		if err != nil {
			return err
		}

		backends := make([]maglev.Backend, 0, len(ips))
		for _, ip := range ips {
			backends = append(backends, maglev.Backend{
				ID: ip.String(),
			})
		}

		newHash := util.HashBackends(backends)
		if newHash == lastHash {
			return nil // no change → skip update
		}
		lastHash = newHash

		select {
		case updates <- backends:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	// Initial resolve
	if err := resolveAndSend(); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			_ = resolveAndSend()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

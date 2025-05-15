package watcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"sort"
	"time"

	"github.com/hanapedia/magseven/pkg/maglev"
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

		newHash := hashBackends(backends)
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

func hashBackends(backends []maglev.Backend) string {
	ids := make([]string, 0, len(backends))
	for _, b := range backends {
		ids = append(ids, b.ID)
	}
	sort.Strings(ids) // ensure order-insensitive hash

	joined := ""
	for _, id := range ids {
		joined += id + "\n"
	}

	sum := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(sum[:])
}

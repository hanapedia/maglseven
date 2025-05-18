package watcher

import (
	"context"

	"github.com/hanapedia/maglseven/pkg/maglev"
)

type StaticWatcher struct {
	Backends []maglev.Backend
}

func (w *StaticWatcher) Watch(ctx context.Context, updates chan<- []maglev.Backend) error {
	// Just send once and close the channel
	select {
	case updates <- w.Backends:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

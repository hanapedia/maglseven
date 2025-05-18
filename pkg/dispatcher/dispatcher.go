package dispatcher

import (
	"log/slog"
	"net/http"

	"github.com/hanapedia/maglseven/pkg/maglev"
	"github.com/hanapedia/maglseven/pkg/util"
)

type Dispatcher struct {
	table  *maglev.Table
	keyFn  func(*http.Request) string
	logger util.Logger
}

// NewDispatcher returns a dispatcher with a given Maglev table and key extraction function
func NewDispatcher(table *maglev.Table, keyFn func(*http.Request) string) *Dispatcher {
	return &Dispatcher{
		table: table,
		keyFn: keyFn,
		logger: slog.Default(),
	}
}

// Route selects a backend URL for an incoming HTTP request
func (d *Dispatcher) Route(r *http.Request) string {
	key := d.keyFn(r)
	id := d.table.Lookup(key).ID
	d.logger.Debug("Routing request", "key", key, "ID", id) // Debug log for "key -> ID" mapping
	return id
}

func (d *Dispatcher) UpdateTable(t *maglev.Table) {
	oldTable := d.table
	d.table = t
	d.logger.Info("Updated Maglev table", "oldTable", oldTable.String(), "newTable", t.String()) // Info log for table update
}

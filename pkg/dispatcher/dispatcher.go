package dispatcher

import (
	"net/http"

	"github.com/hanapedia/magseven/pkg/maglev"
)

type Dispatcher struct {
	table  *maglev.Table
	keyFn  func(*http.Request) string
}

// NewDispatcher returns a dispatcher with a given Maglev table and key extraction function
func NewDispatcher(table *maglev.Table, keyFn func(*http.Request) string) *Dispatcher {
	return &Dispatcher{
		table: table,
		keyFn: keyFn,
	}
}

// Route selects a backend URL for an incoming HTTP request
func (d *Dispatcher) Route(r *http.Request) string {
	key := d.keyFn(r)
	return d.table.Lookup(key).ID
}

func (d *Dispatcher) UpdateTable(t *maglev.Table) {
	d.table = t
}

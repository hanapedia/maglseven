package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hanapedia/magseven/pkg/dispatcher"
	"github.com/hanapedia/magseven/pkg/maglev"
	"github.com/hanapedia/magseven/pkg/proxy"
	"github.com/hanapedia/magseven/pkg/watcher"
)

func getenv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func main() {
	fqdn := getenv("BACKEND_FQDN", "hello-headless.default.svc.cluster.local")
	intervalStr := getenv("RESOLVE_INTERVAL", "10s")
	listenPort := getenv("LISTEN_PORT", "8080")
	destPort := getenv("DEST_PORT", "8080")
	headerName := getenv("ROUTE_HEADER", "X-Room-ID")

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatalf("Invalid RESOLVE_INTERVAL: %v", err)
	}

	log.Printf("Starting Maglev Proxy: resolving %s every %s on :%s using header %s",
		fqdn, interval, listenPort, headerName)

	watcher := &watcher.DNSWatcher{
		FQDN:     fqdn,
		Interval: interval,
	}

	updates := make(chan []maglev.Backend, 1)
	ctx := context.Background()

	go func() {
		if err := watcher.Watch(ctx, updates); err != nil {
			log.Fatalf("Watcher failed: %v", err)
		}
	}()

	var dispatcherInstance *dispatcher.Dispatcher

	// Blocking wait for first backend list
	initial := <-updates
	table, _ := maglev.Build(initial, maglev.DefaultTableSize)
	dispatcherInstance = dispatcher.NewDispatcher(table, func(r *http.Request) string {
		return r.Header.Get(headerName)
	})

	// Watch for updates
	go func() {
		for backends := range updates {
			newTable, _ := maglev.Build(backends, maglev.DefaultTableSize)
			dispatcherInstance.UpdateTable(newTable)
			log.Printf("Updated Maglev table with %d backends", len(backends))
		}
	}()

	handler := proxy.NewProxy(destPort, dispatcherInstance)

	log.Fatal(http.ListenAndServe(":"+listenPort, handler))
}

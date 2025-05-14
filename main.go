package main

import (
	"context"
	"log"
	"net/http"

	"github.com/hanapedia/magseven/pkg/watcher"
	"github.com/hanapedia/magseven/pkg/dispatcher"
	"github.com/hanapedia/magseven/pkg/maglev"
	"github.com/hanapedia/magseven/pkg/proxy"
)

func main() {
	backends := []maglev.Backend{
		{ID: "localhost:9001"},
		{ID: "localhost:9002"},
		{ID: "localhost:9003"},
	}

	watcher := &watcher.StaticWatcher{Backends: backends}
	updates := make(chan []maglev.Backend, 1)

	go func() {
		if err := watcher.Watch(context.Background(), updates); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for initial update
	backendList := <-updates
	table, _ := maglev.Build(backendList, maglev.DefaultTableSize)
	keyFn := func(r *http.Request) string {
		return r.Header.Get("X-Room-ID")
	}

	dispatch := dispatcher.NewDispatcher(table, keyFn)
	proxyHandler := proxy.NewProxy(dispatch)

	log.Println("Starting Maglev Proxy on :8080")
	log.Fatal(http.ListenAndServe(":8080", proxyHandler))
}

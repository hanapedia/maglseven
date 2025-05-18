package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hanapedia/maglseven/pkg/maglev"
	"github.com/hanapedia/maglseven/pkg/proxy"
	"github.com/hanapedia/maglseven/pkg/util"
	"github.com/hanapedia/maglseven/pkg/watcher"
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
	maxHistoryStr := getenv("MAX_HISTORY", "5")
	replicaCountStr := getenv("REPLICA_COUNT", "3")
	maxJumpsStr := getenv("MAX_JUMPS", "5")
	failureCIDRStr := getenv("FAILURE_CIDR", "32")

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatalf("Invalid RESOLVE_INTERVAL: %v", err)
	}

	maxHistory, err := strconv.Atoi(maxHistoryStr)
	if err != nil || maxHistory <= 0 {
		log.Fatalf("Invalid MAX_HISTORY value: %v", err)
	}

	replicaCount, err := strconv.Atoi(replicaCountStr)
	if err != nil || replicaCount <= 0 {
		log.Fatalf("Invalid REPLICA_COUNT value: %v", err)
	}

	maxJumps, err := strconv.Atoi(maxJumpsStr)
	if err != nil || maxJumps <= 0 {
		log.Fatalf("Invalid MAX_JUMPS value: %v", err)
	}

	failureCIDR, err := util.ParseCIDRMaskFromString(failureCIDRStr)
	if err != nil {
		log.Fatalf("Invalid FAILURE_CIDR value: %v", err)
	}

	log.Printf("Starting Handoff Proxy: resolving %s every %s on :%s using header %s",
		fqdn, interval, listenPort, headerName)

	w := &watcher.DNSWatcher{
		FQDN:        fqdn,
		Interval:    interval,
		FailureCIDR: &failureCIDR,
	}

	updates := make(chan []maglev.Backend, 1)
	ctx := context.Background()

	go func() {
		if err := w.Watch(ctx, updates); err != nil {
			log.Fatalf("Watcher failed: %v", err)
		}
	}()

	// First maglev table
	initial := <-updates
	table, _ := maglev.Build(initial, maglev.DefaultTableSize)
	// Initialize VersionedRouter
	versionedRouter := maglev.NewVersionedRouter(maxHistory)
	versionedRouter.AddGeneration(1, table)

	// Track current generation count
	var genCounter uint64 = 1

	// Watch updates and rotate generations
	go func() {
		for backends := range updates {
			genCounter++
			newTable, _ := maglev.Build(backends, maglev.DefaultTableSize)
			versionedRouter.AddGeneration(genCounter, newTable)
			log.Printf("Updated maglev table to generation %d with %d backends", genCounter, len(backends))
		}
	}()

	// Use new versioned proxy
	handler := proxy.NewHandoffProxy(destPort, versionedRouter, replicaCount, maxJumps)

	log.Fatal(http.ListenAndServe(":"+listenPort, handler))
}

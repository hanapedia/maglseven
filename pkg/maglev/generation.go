package maglev

import (
	"strconv"
	"sync"
)

const (
	HeaderGeneration       = "X-Maglev-Generation"
	HeaderReplicationPeers = "X-Maglev-Replication-Peers"
	HeaderPreviousPeers    = "X-Maglev-Previous-Peers"
	HeaderPreviousPrimary  = "X-Maglev-Previous-Primary"
)

type VersionedRouter struct {
	mu         sync.RWMutex
	tables     map[uint64]*Table // generation → table
	order      []uint64          // oldest first
	maxHistory int
	currentGen uint64
}

func NewVersionedRouter(maxHistory int) *VersionedRouter {
	return &VersionedRouter{
		tables:     make(map[uint64]*Table),
		maxHistory: maxHistory,
	}
}

func (vr *VersionedRouter) AddGeneration(gen uint64, table *Table) {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	if _, exists := vr.tables[gen]; exists {
		return
	}

	vr.tables[gen] = table
	vr.order = append(vr.order, gen)
	vr.currentGen = gen

	if len(vr.order) > vr.maxHistory {
		old := vr.order[0]
		delete(vr.tables, old)
		vr.order = vr.order[1:]
	}
}

type RouteResult struct {
	Backend          Backend   // current primary
	Peers            []Backend // current generation peers
	PrevPeers        []Backend // previous generation peers
	PrevPrimary      *Backend  // if backend changed
	Generation       uint64    // current generation
	ClientGeneration *uint64   // parsed client generation (nil if absent/invalid)
	RequiresRecovery bool      // client is stale and backend changed
}

func (vr *VersionedRouter) Route(key string, clientGenHeader string, replicationCount, maxJumps int) RouteResult {
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	currentGen := vr.currentGen
	currentTable := vr.tables[currentGen]

	// Compute current peers and primary
	newPeers := currentTable.LookupNWithDomainIsolation(key, replicationCount, maxJumps)
	var newPrimary Backend
	if len(newPeers) > 0 {
		newPrimary = newPeers[0]
		newPeers = newPeers[1:]
	}

	// Default result if no header
	if clientGenHeader == "" {
		return RouteResult{
			Backend:          newPrimary,
			Peers:            newPeers,
			Generation:       currentGen,
			ClientGeneration: nil,
		}
	}

	// Parse client generation
	clientGen, err := strconv.ParseUint(clientGenHeader, 10, 64)
	if err != nil || vr.tables[clientGen] == nil {
		// Fallback: treat as fresh
		return RouteResult{
			Backend:          newPrimary,
			Peers:            newPeers,
			Generation:       currentGen,
			ClientGeneration: nil,
		}
	}

	// Compare backends across generations
	clientTable := vr.tables[clientGen]
	oldPeers := clientTable.LookupNWithDomainIsolation(key, replicationCount, maxJumps)
	var oldPrimary Backend
	if len(oldPeers) > 0 {
		oldPrimary = oldPeers[0]
		oldPeers = oldPeers[1:]
	}

	// Check whether backend changed
	if oldPrimary.ID != newPrimary.ID {
		return RouteResult{
			Backend:          newPrimary,
			Peers:            newPeers,
			PrevPeers:        oldPeers,
			PrevPrimary:      &oldPrimary,
			Generation:       currentGen,
			ClientGeneration: &clientGen,
			RequiresRecovery: true,
		}
	}

	// Backend didn't change, just re-sync peers
	return RouteResult{
		Backend:          newPrimary,
		Peers:            newPeers,
		Generation:       currentGen,
		ClientGeneration: &clientGen,
	}
}

package util

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/hanapedia/maglseven/pkg/maglev"
)

func HashBackends(backends []maglev.Backend) string {
	ids := make([]string, 0, len(backends))
	for _, b := range backends {
		ids = append(ids, b.ID)
	}
	sort.Strings(ids)

	joined := ""
	for _, id := range ids {
		joined += id + "\n"
	}

	sum := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(sum[:])
}

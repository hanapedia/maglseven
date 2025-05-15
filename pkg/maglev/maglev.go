package maglev

import (
	"errors"
	"hash/fnv"
)

type Backend struct {
	ID string // Unique identifier (e.g., Pod IP or name)
}

type Table struct {
	slots    []int       // slot index -> backend index
	backends []Backend
	m        int         // table size
}

// Prime number recommended, e.g., 65537
const DefaultTableSize = 65537

// Build creates a Maglev lookup table for the given backends
func Build(backends []Backend, tableSize int) (*Table, error) {
	numBackends := len(backends)
	if numBackends == 0 || tableSize <= 0 {
		return nil, errors.New("invalid inputs")
	}

	offsets := make([]int, numBackends)
	skips := make([]int, numBackends)
	table := make([]int, tableSize)
	for i := range table {
		table[i] = -1
	}

	// Compute offset and skip for each backend
	for i, b := range backends {
		h := hash32(b.ID)
		offsets[i] = int(h % uint32(tableSize))

		// skips must be coprime to tableSize 
		// ([0, tableSize-1]) is coprime since tableSize is prime
		skips[i] = int((hash32(b.ID + "#") % uint32(tableSize-1)) + 1)
	}

	// Fill the table using Maglev permutation
	posForBackends := make([]int, numBackends)
	for filled := 0; filled < tableSize; {
		for i := range numBackends {
			c := (offsets[i] + posForBackends[i]*skips[i]) % tableSize
			for table[c] != -1 {
				posForBackends[i]++
				c = (offsets[i] + posForBackends[i]*skips[i]) % tableSize
			}
			table[c] = i
			posForBackends[i]++
			filled++
			if filled >= tableSize {
				break
			}
		}
	}

	return &Table{
		slots:    table,
		backends: backends,
		m:        tableSize,
	}, nil
}

// Lookup returns the backend for a given key
func (t *Table) Lookup(key string) Backend {
	idx := int(hash32(key) % uint32(t.m))
	backendIndex := t.slots[idx]
	return t.backends[backendIndex]
}

// LookupN returns up to `n` unique backends for the given key,
// scanning forward from the hash index in the lookup table.
func (t *Table) LookupN(key string, n int) []Backend {
    if n <= 0 {
        return nil
    }

    seen := make(map[int]struct{})
    result := make([]Backend, 0, n)
    start := int(hash32(key) % uint32(t.m))

    // Linear scan through slots starting at hashed offset
    for i := 0; len(result) < n && i < t.m; i++ {
        slot := (start + i) % t.m
        backendIndex := t.slots[slot]

        if _, ok := seen[backendIndex]; !ok {
            seen[backendIndex] = struct{}{}
            result = append(result, t.backends[backendIndex])
        }
    }

    return result
}

// Simple FNV-1a hash
// Can be swapped with other stronger hashes such as xxHash or MurmurHash3
func hash32(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

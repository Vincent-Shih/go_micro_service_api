package helper_test

import (
	"go_micro_service_api/pkg/helper"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SafeMap struct {
	mu    sync.Mutex
	store map[int64]bool
}

func (m *SafeMap) Set(key int64, value bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = value
}

func (m *SafeMap) Get(key int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store[key]
}

func TestNextID(t *testing.T) {
	sf := helper.NewSnowflake(1)

	// Test generating a single ID
	id := sf.NextID()
	assert.Greater(t, id, int64(0))
}

func TestMultipleIDs(t *testing.T) {
	sf := helper.NewSnowflake(1)

	// Test generating multiple IDs
	testCases := []struct {
		name     string
		expected int
	}{
		{"low", 1_000},
		{"medium", 1_000_000},
		// {"high", 10_000_000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set := make(map[int64]bool)

			for i := 0; i < tc.expected; i++ {
				id := sf.NextID()
				set[id] = true
			}

			assert.Equal(t, tc.expected, len(set))
		})
	}
}

func TestMutipleNodeIDs(t *testing.T) {
	const nodeNumber = 10

	sfs := make([]*helper.Snowflake, nodeNumber)
	for i := 0; i < nodeNumber; i++ {
		sfs[i] = helper.NewSnowflake(helper.MachineID(i))
	}

	// Test generating multiple IDs
	testCases := []struct {
		name     string
		burden   int
		expected int
	}{
		{"low", 1_000, 1_000 * nodeNumber},
		{"medium", 1_000_000, 1_000_000 * nodeNumber},
		// {"high", 10_000_000, 10_000_000 * nodeNumber},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			safeMap := SafeMap{
				mu:    sync.Mutex{},
				store: make(map[int64]bool),
			}

			var wg sync.WaitGroup

			for _, sf := range sfs {
				go func(wg *sync.WaitGroup, safeMap *SafeMap, sf *helper.Snowflake) {
					defer wg.Done()
					for i := 0; i < tc.burden; i++ {
						id := sf.NextID()
						safeMap.Set(id, true)
					}
				}(&wg, &safeMap, sf)
			}

			wg.Add(10)
			wg.Wait()

			assert.Equal(t, tc.expected, len(safeMap.store))
		})
	}
}

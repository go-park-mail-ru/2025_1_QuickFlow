package thread_safe_map

import (
	"sync"
	"testing"
)

func TestThreadSafeMap(t *testing.T) {
	type testCase struct {
		name      string
		operation func(m *ThreadSafeMap[string, int]) bool
		expect    bool
	}

	tests := []testCase{
		{
			name: "Set and Get",
			operation: func(m *ThreadSafeMap[string, int]) bool {
				m.Set("key1", 42)
				val, ok := m.Get("key1")
				return ok && val == 42
			},
			expect: true,
		},
		{
			name: "Get non-existent key",
			operation: func(m *ThreadSafeMap[string, int]) bool {
				_, ok := m.Get("unknown")
				return !ok
			},
			expect: true,
		},
		{
			name: "Delete key",
			operation: func(m *ThreadSafeMap[string, int]) bool {
				m.Set("key2", 99)
				m.Delete("key2")
				_, ok := m.Get("key2")
				return !ok
			},
			expect: true,
		},
		{
			name: "HasKey",
			operation: func(m *ThreadSafeMap[string, int]) bool {
				m.Set("key3", 88)
				return m.HasKey("key3")
			},
			expect: true,
		},
		{
			name: "GetKeys",
			operation: func(m *ThreadSafeMap[string, int]) bool {
				m.Set("keyA", 1)
				m.Set("keyB", 2)
				keys := m.GetKeys()
				return len(keys) == 2
			},
			expect: true,
		},
		{
			name: "GetValues",
			operation: func(m *ThreadSafeMap[string, int]) bool {
				m.Set("one", 1)
				m.Set("two", 2)
				values := m.GetValues()
				return len(values) == 2
			},
			expect: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewThreadSafeMap[string, int]()
			if result := tc.operation(m); result != tc.expect {
				t.Errorf("Test failed: expected %v, got %v", tc.expect, result)
			}
		})
	}
}

func TestThreadSafeMap_Concurrency(t *testing.T) {
	m := NewThreadSafeMap[int, int]()
	var wg sync.WaitGroup

	numGoroutines := 100
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			m.Set(i, i*10)
		}(i)

		go func(i int) {
			defer wg.Done()
			m.Get(i)
		}(i)
	}

	wg.Wait()
}

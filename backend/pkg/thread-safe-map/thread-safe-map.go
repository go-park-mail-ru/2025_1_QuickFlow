package thread_safe_map

import "sync"

// ThreadSafeMap is a thread-safe map.
type ThreadSafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewThreadSafeMap creates a new thread-safe map.
func NewThreadSafeMap[K comparable, V any]() *ThreadSafeMap[K, V] {
	return &ThreadSafeMap[K, V]{
		m: make(map[K]V),
	}
}

// Set sets a value for a key.
func (m *ThreadSafeMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	m.m[key] = value
	m.mu.Unlock()
}

// Get gets a value for a key.
func (m *ThreadSafeMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	value, ok := m.m[key]
	m.mu.RUnlock()
	return value, ok
}

// Delete deletes a key.
func (m *ThreadSafeMap[K, V]) Delete(key K) {
	m.mu.Lock()
	delete(m.m, key)
	m.mu.Unlock()
}

// HasKey checks if a key exists.
func (m *ThreadSafeMap[K, V]) HasKey(key K) bool {
	m.mu.RLock()
	_, ok := m.m[key]
	m.mu.RUnlock()
	return ok
}

// GetKeys returns all keys.
func (m *ThreadSafeMap[K, V]) GetKeys() []K {
	m.mu.RLock()
	keys := make([]K, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	m.mu.RUnlock()
	return keys
}

// GetValues returns all values.
func (m *ThreadSafeMap[K, V]) GetValues() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.m))
	for _, v := range m.m {
		values = append(values, v)
	}
	return values
}

// GetMapCopy returns a copy of the map.
func (m *ThreadSafeMap[K, V]) GetMapCopy() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copyMap := make(map[K]V, len(m.m))
	for k, v := range m.m {
		copyMap[k] = v
	}
	return copyMap
}

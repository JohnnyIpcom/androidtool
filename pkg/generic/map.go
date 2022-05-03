package generic

import "sync"

// KV is a key/value pair.
type KV[K comparable, V any] struct {
	Key   K
	Value V
}

// NewKV returns a new KV.
func NewKV[K comparable, V any](k K, v V) KV[K, V] {
	return KV[K, V]{
		Key:   k,
		Value: v,
	}
}

// Map is a thread-safe map container.
type Map[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

// NewMap returns a new Map.
func NewMap[K comparable, V any](items ...KV[K, V]) *Map[K, V] {
	m := &Map[K, V]{
		m: make(map[K]V),
	}

	m.StoreAllKV(items)
	return m
}

// NewMapFromMap returns a new Map containing a copy of the provided map.
func NewMapFromMap[K comparable, V any](m map[K]V) *Map[K, V] {
	return &Map[K, V]{
		m: m,
	}
}

// Store adds the key to the map.
func (m *Map[K, V]) Store(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[k] = v
}

// StoreAll adds all the keys in the provided slice to the map.
func (m *Map[K, V]) StoreAll(items []KV[K, V]) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range items {
		m.m[item.Key] = item.Value
	}
}

// StoreAllKV adds all the keys in the provided slice to the map.
func (m *Map[K, V]) StoreAllKV(items []KV[K, V]) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range items {
		m.m[item.Key] = item.Value
	}
}

// Load returns the value for the key in the map.
func (m *Map[K, V]) Load(k K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[k]
	return v, ok
}

// LoadOrStore returns the value for the key in the map, or stores the default
// value if the key is not in the map.
func (m *Map[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	actual, loaded = m.m[k]
	if !loaded {
		m.m[k] = v
		actual = v
	}
	return actual, loaded
}

// StoreAllMap stores all key/value pairs from the provided map into the map.
func (m *Map[K, V]) StoreAllMap(items map[K]V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range items {
		m.m[k] = v
	}
}

// Has returns true if the map contains the key.
func (m *Map[K, V]) Has(k K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.m[k]
	return ok
}

// Delete removes the key from the map.
func (m *Map[K, V]) Delete(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, k)
}

// Len returns the number of key/value pairs in the map.
func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}

// Clear removes all keys from the map.
func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m = make(map[K]V)
}

// Keys returns a slice of all keys in the map.
func (m *Map[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of all values in the map.
func (m *Map[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.m))
	for _, v := range m.m {
		values = append(values, v)
	}
	return values
}

// Each calls the provided function for each key/value pair in the map.
func (m *Map[K, V]) Each(f func(K, V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.m {
		if !f(k, v) {
			break
		}
	}
}

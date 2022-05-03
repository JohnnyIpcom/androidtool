package generic

import "sync"

// Set is a thread-safe set container.
type Set[K comparable] struct {
	m  map[K]struct{}
	mu sync.RWMutex
}

// New returns a new Set.
func NewSet[K comparable](items ...K) *Set[K] {
	s := &Set[K]{
		m: make(map[K]struct{}),
	}

	s.StoreAll(items)
	return s
}

// Store adds the key to the set.
func (s *Set[K]) Store(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[k] = struct{}{}
}

// StoreAll adds all the keys in the provided slice to the set.
func (s *Set[K]) StoreAll(keys []K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range keys {
		s.m[k] = struct{}{}
	}
}

// Has returns true if the set contains the key.
func (s *Set[K]) Has(k K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[k]
	return ok
}

// Delete removes the key from the set.
func (s *Set[K]) Delete(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, k)
}

// Len returns the number of keys in the set.
func (s *Set[K]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

// Clear removes all keys from the set.
func (s *Set[K]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[K]struct{})
}

// Keys returns a slice of the keys in the set.
func (s *Set[K]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, len(s.m))
	i := 0
	for k := range s.m {
		keys[i] = k
		i++
	}
	return keys
}

// Each calls f for each key in the set.
func (s *Set[K]) Each(f func(K) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.m {
		if !f(k) {
			break
		}
	}
}

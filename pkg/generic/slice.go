package generic

import "sync"

// Slice is a thread-safe implementation of a slice.
type Slice[K any] struct {
	mu    sync.RWMutex
	items []K
}

// NewSlice returns a new Slice.
func NewSlice[K any](items ...K) *Slice[K] {
	s := &Slice[K]{}

	s.StoreAll(items)
	return s
}

// Store adds a new item to the slice.
func (s *Slice[K]) Store(item K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// StoreAll adds multiple items to the slice.
func (s *Slice[K]) StoreAll(items []K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, items...)
}

// Load returns the item at the given index.
func (s *Slice[K]) Load(index int) K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items[index]
}

// Delete removes the item at the given index.
func (s *Slice[K]) Delete(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items[:index], s.items[index+1:]...)
}

// Len returns the number of items in the slice.
func (s *Slice[K]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// Clear removes all items from the slice.
func (s *Slice[K]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = []K{}
}

// Values returns all items in the slice.
func (s *Slice[K]) Values() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items
}

// Each calls the provided function for each item in the slice.
func (s *Slice[K]) Each(f func(int, K) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for index, item := range s.items {
		if !f(index, item) {
			break
		}
	}
}

package generic

import (
	"math/rand"
	"sync"
	"testing"
)

// checkSlice that the set contains the expected keys
func checkSlice[K any](t *testing.T, s *Slice[K], get func(int, K) bool) {
	s.Each(func(index int, k K) bool {
		if !get(index, k) {
			t.Fail()
		}
		return true
	})
}

func TestSliceAppending(t *testing.T) {
	slice := NewSlice[uint32]()

	const nops = 200000
	var wg sync.WaitGroup
	wg.Add(2 * nops)
	for i := 0; i < nops; i++ {
		value := rand.Uint32() + 1
		go func(value uint32) {
			defer wg.Done()
			slice.Store(value)
		}(value)

		go func() {
			defer wg.Done()
			checkSlice(t, slice, func(id int, v uint32) bool {
				return v != 0
			})
		}()
	}
	wg.Wait()
}

func TestSliceAppending2(t *testing.T) {
	slice := make([]uint32, 0)

	fail := false

	const nops = 200000
	var wg sync.WaitGroup
	wg.Add(2 * nops)
	for i := 0; i < nops; i++ {
		value := rand.Uint32() + 1
		go func(value uint32) {
			defer wg.Done()
			slice = append(slice, value)
		}(value)

		go func() {
			defer wg.Done()
			if len(slice) > 0 && slice[len(slice)-1] == 0 {
				fail = true
			}
		}()
	}
	wg.Wait()

	if !fail {
		t.Fail()
	}
}

package generic

import (
	"math/rand"
	"sync"
	"testing"
)

// checkSet that the set contains the expected keys
func checkSet[K comparable](t *testing.T, s *Set[K], get func(K) bool) {
	s.Each(func(k K) bool {
		if !get(k) {
			t.Fatalf("Expected %v to be in the set", k)
		}
		return true
	})
}

func TestSetCrossCheck(t *testing.T) {
	stdm := sync.Map{}
	set := NewSet[int]()

	const nops = 1000
	var wg sync.WaitGroup
	wg.Add(nops)
	for i := 0; i < nops; i++ {
		op := rand.Intn(2)
		switch op {
		case 0:
			key := rand.Int()
			go func(key int) {
				defer wg.Done()
				stdm.Store(key, true)
				set.Store(key)
			}(key)
		case 1:
			var del int
			stdm.Range(func(key, value interface{}) bool {
				del = key.(int)
				return false
			})
			go func(key int) {
				defer wg.Done()
				stdm.Delete(key)
				set.Delete(del)
			}(del)
		}
	}

	wg.Wait()
	checkSet(t, set, func(k int) bool {
		_, ok := stdm.Load(k)
		return ok
	})
}

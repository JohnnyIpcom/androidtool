package generic

import (
	"math/rand"
	"sync"
	"testing"
)

// checkMap that the map contains the expected keys and values
func checkMap(t *testing.T, m *Map[int, int], get func(int) bool) {
	m.Each(func(k, v int) bool {
		if !get(k) {
			t.Fatalf("Expected %v to be in the map", k)
		}
		return true
	})
}

func TestMapCrossCheck(t *testing.T) {
	stdm := sync.Map{}
	m := NewMap[int, int]()

	const nops = 1000
	var wg sync.WaitGroup
	wg.Add(nops)
	for i := 0; i < nops; i++ {
		op := rand.Intn(2)
		switch op {
		case 0:
			key, value := rand.Int(), rand.Int()
			go func(key int) {
				defer wg.Done()
				stdm.Store(key, value)
				m.Store(key, value)
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
				m.Delete(del)
			}(del)
		}
	}

	wg.Wait()
	checkMap(t, m, func(k int) bool {
		_, ok := stdm.Load(k)
		return ok
	})
}

package concurrentrand

import (
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
)

const testSeed = 42

func TestSameOrder(t *testing.T) {
	// If used from a single goroutine, a concurrentrand source returns
	// numbers in the same order as the default rand.Source.
	csrc, src := NewSource(testSeed), rand.NewSource(testSeed)
	for i := 0; i < 2*BufferSize; i++ {
		got, want := csrc.Int63(), src.Int63()
		if got != want {
			t.Errorf("iteration %d, got %d, want %d", i, got, want)
		}
	}
}

type int64Slice []int64

func (s int64Slice) Len() int           { return len(s) }
func (s int64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s int64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func TestMultipleGoroutines(t *testing.T) {
	// Using a concurrentrand source from multiple goroutines returns the same
	// numbers as the default rand.Source, but not necessarily in the same
	// order (depending on the order in which the goroutintes are scheduled).
	got, want := make([]int64, 2*BufferSize), make([]int64, 2*BufferSize)
	csrc, src := NewSource(testSeed), rand.NewSource(testSeed)
	var wg sync.WaitGroup
	for i := 0; i < 2*BufferSize; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i] = csrc.Int63()
		}(i)
		want[i] = src.Int63()
	}
	wg.Wait()
	sort.Sort(int64Slice(got))
	sort.Sort(int64Slice(want))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %v\nwant %v", got, want)
	}
}

func benchmarkSource(b *testing.B, r func() int64) {
	// Benchmark the source used by calling r().  Set a high amount of
	// parallelism, since that's when concurrentrand is actually useful.
	b.SetParallelism(500)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		for pb.Next() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r()
			}()
		}
		wg.Wait()
	})
}

func BenchmarkDefaultSource(b *testing.B) {
	// BenchmarkDefaultSource   1000000              2494 ns/op
	rand.Seed(testSeed)
	benchmarkSource(b, rand.Int63)
}

func BenchmarkConcurrentSource(b *testing.B) {
	// BenchmarkConcurrentSource        1000000              1069 ns/op
	r := rand.New(NewSource(testSeed))
	benchmarkSource(b, r.Int63)
}

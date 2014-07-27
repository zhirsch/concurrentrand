// Package concurrentrand contains types and functions for generating random
// numbers from multiple concurrent goroutines.  This is most useful when
// random numbers are needed in a lot of concurrently running goroutines; for
// a small number of goroutines, the default rand.Source is probably a better
// choice.
package concurrentrand

import (
	"math/rand"
)

var BufferSize = 500

// A source supports generating random numbers from multiple goroutines,
// without locking.
type source struct {
	ch <-chan int64
}

// NewSource returns a rand.Source that is safe for use from multiple
// goroutines, without locking. Changing the seed of this source after it has
// been created is not supported. The actual source of the random numbers is
// the same as from the standard library's rand.NewSource.
func NewSource(seed int64) rand.Source {
	ch := make(chan int64, BufferSize)
	go func(src rand.Source) {
		for {
			ch <- src.Int63()
		}
	}(rand.NewSource(seed))
	return &source{ch}
}

// Int63 returns a new 63-bit random number.
func (cs *source) Int63() int64 {
	return <-cs.ch
}

// Seed is not implemented.
func (cs *source) Seed(seed int64) {
	panic("changing the seed is not supported")
}

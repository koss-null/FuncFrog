package stream

import (
	"sync"

	"github.com/koss-null/lambda-go/internal/tools"
)

type (
	stream[T any] struct {
		wrks chan struct{}
		dt   []T
		dtMu sync.Mutex
		bm   tools.Bitmask
		bmMu sync.Mutex
		eq   func(a, b T) bool
		fns  []func([]T, int)
	}

	StreamI[T any] interface {
		// Skip removes first n elements from underlying slice
		Skip(n uint) StreamI[T]
		// Trim removes last n elements from underlying slice
		Trim(n uint) StreamI[T]

		Map(func(T) T) StreamI[T]
		Filter(func(T) bool) StreamI[T]
		// Sorted sotrs the underlying array
		Sorted(less func(a, b T) bool) StreamI[T]
		// Go splits execution into n goroutines
		Go(n uint) StreamI[T]
		// CmpWith sets compare function
		CmpWith(eq func(a, b T) bool) StreamI[T]

		// Functions which does not return the stream
		// Slice returns resulting slice
		Slice() []T
		// Any returns true if the underlying slice is not empty and at least one element is true
		Any() bool
		// None is not Any()
		None() bool
		// Count returns the length of the result array
		Count() int
		// Contains returns true if element is in the array
		Contains(item T) bool
	}
)

// Stream creates
func Stream[T any](data []T) StreamI[T] {
	dtCopy := make([]T, 0, len(data))
	copy(dtCopy, data)
	workers := make(chan struct{}, 1)
	workers <- struct{}{}
	return &stream[T]{wrks: workers, dt: dtCopy}
}

func (st *stream[T]) Skip(n uint) StreamI[T] {
	return st
}

func (st *stream[T]) Trim(n uint) StreamI[T] {
	return st
}

func (st *stream[T]) Map(fn func(T) T) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, _ int) {
		<-st.wrks

		for i := range dt {
			dt[i] = fn(dt[i])
		}

		st.wrks <- struct{}{}
	})
	return st
}

func (st *stream[T]) Filter(fn func(T) bool) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		<-st.wrks

		res := make([]bool, len(dt))
		for i := range dt {
			res[i] = fn(dt[i])
		}

		st.bmMu.Lock()
		for i := range res {
			st.bm.Put(uint(offset+i), res[i])
		}
		st.bmMu.Unlock()

		st.wrks <- struct{}{}
	})
	return st
}

func (st *stream[T]) Sorted(less func(a, b T) bool) StreamI[T] {
	return st
}

func (st *stream[T]) Go(n uint) StreamI[T] {
	return st
}

func (st *stream[T]) CmpWith(eq func(a, b T) bool) StreamI[T] {
	st.eq = eq
	return st
}

func (st *stream[T]) Slice() []T {
	return st.dt
}

func (st *stream[T]) Any() bool {
	return len(st.dt) == 1
}

func (st *stream[T]) None() bool {
	return len(st.dt) == 0
}

func (st *stream[T]) Count() int {
	return len(st.dt)
}

func (st *stream[T]) Contains(item T) bool {
	if st.eq == nil {
		return false
	}
	for i := range st.dt {
		if st.eq(st.dt[i], item) {
			return true
		}
	}
	return false
}

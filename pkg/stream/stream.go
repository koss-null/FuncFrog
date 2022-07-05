package stream

import (
	"sort"
	"sync"

	"github.com/koss-null/lambda-go/internal/tools"
)

const (
	u0 = uint(0)
)

type (
	stream[T any] struct {
		wrks    chan struct{}
		wrksLen uint
		dt      []T
		dtMu    sync.Mutex
		bm      *tools.Bitmask
		bmMu    sync.Mutex
		eq      func(a, b T) bool
		fns     []func([]T, int)
	}

	// StreamI describes all functions available for the stream API
	StreamI[T any] interface {
		// Trimming the sequence
		// Skip removes first n elements from underlying slice
		Skip(n uint) StreamI[T]
		// Trim removes last n elements from underlying slice
		Trim(n uint) StreamI[T]

		// Actions on sequence
		// Map executes function on each element of an underlying slice and
		// changes the element to the result of the function
		Map(func(T) T) StreamI[T]
		// Filteer
		Filter(func(T) bool) StreamI[T]
		// Sorted sotrs the underlying array
		Sorted(less func(a, b T) bool) StreamI[T]
		// Split splits initial slice into multiple
		Split(func(a T) []T) StreamI[T]

		// Config functions
		// Go splits execution into n goroutines, if multiple Go functions are along
		// the stream pipeline, it applies as soon as it's met in the sequence
		Go(n uint) StreamI[T]
		// Cmp sets compare function
		Cmp(eq func(a, b T) bool) StreamI[T]

		// Final functions (which does not return the stream)
		// Slice returns resulting slice
		Slice() []T
		// Any returns true if the underlying slice is not empty and at least one element is true
		Any() bool
		// None is not Any()
		None() bool
		// Count returns the length of the result array
		Count() int
		// Sum sums up the items in the array
		Sum() int
		// Contains returns true if element is in the array
		Contains(item T) bool
	}
)

// Stream creates new instance of stream
func Stream[T any](data []T) StreamI[T] {
	dtCopy := make([]T, len(data))
	copy(dtCopy, data)
	workers := make(chan struct{}, 1)
	workers <- struct{}{}
	bm := tools.Bitmask{}
	bm.PutLine(0, uint(len(data)), true)
	return &stream[T]{wrks: workers, wrksLen: 1, dt: dtCopy, bm: &bm}
}

// S is a shortened Stream
func S[T any](data []T) StreamI[T] {
	return Stream(data)
}

func (st *stream[T]) Skip(n uint) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}

		<-st.wrks

		st.bmMu.Lock()
		_ = st.bm.CaSBorder(0, true, n)
		st.bmMu.Unlock()

		st.wrks <- struct{}{}
	})
	return st
}

func (st *stream[T]) Trim(n uint) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}

		<-st.wrks

		st.bmMu.Lock()
		_ = st.bm.CaSBorderBw(n)
		st.bmMu.Unlock()

		st.wrks <- struct{}{}
	})
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

		bm := st.bm.Copy(uint(offset), uint(offset+len(dt)))
		res := make([]bool, len(dt))
		for i := range dt {
			if bm.Get(uint(offset + i)) {
				res[i] = fn(dt[i])
			}
		}

		st.bmMu.Lock()
		for i := range res {
			// Filter does not add already removed items
			if !res[i] && bm.Get(uint(offset+i)) {
				st.bm.Put(uint(offset+i), false)
			}
		}
		st.bmMu.Unlock()

		st.wrks <- struct{}{}
	})
	return st
}

// Sorted adds two functions: first one sorts everything
// FIXME: Sorted does not apply the bitmask
func (st *stream[T]) Sorted(less func(a, b T) bool) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		<-st.wrks

		sort.Slice(dt, func(i, j int) bool { return less(dt[j], dt[j]) })

		st.wrks <- struct{}{}
	})
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}

		<-st.wrks

		sort.Slice(st.dt, func(i, j int) bool {
			return less(dt[j], dt[j])
		})

		st.wrks <- struct{}{}
	})

	return st
}

func (st *stream[T]) Go(n uint) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		for i := u0; i < st.wrksLen; i++ {
			<-st.wrks
		}
		close(st.wrks)

		wrks := make(chan struct{}, n)
		for i := u0; i < n; i++ {
			wrks <- struct{}{}
		}
		st.wrksLen = n
	})
	return st
}

func (st *stream[T]) Cmp(eq func(a, b T) bool) StreamI[T] {
	st.eq = eq
	return st
}

func (st *stream[T]) Split(sp func(T) []T) StreamI[T] {
	// TODO: implement
	return st
}

// run start executing st.fns
func (st *stream[T]) run() {
	for _, fn := range st.fns {
		fn(st.dt, 0)
	}
}

func (st *stream[T]) Slice() []T {
	var data []T
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}
		for i := u0; i < st.wrksLen; i++ {
			<-st.wrks
		}

		st.dtMu.Lock()
		st.bmMu.Lock()
		defer st.dtMu.Unlock()
		defer st.bmMu.Unlock()

		data = make([]T, 0, len(st.dt))
		for _, dt := range st.dt {
			_, exist := st.bm.Next()
			if exist {
				data = append(data, dt)
			}
		}
	})

	st.run()
	return data
}

func (st *stream[T]) Any() bool {
	st.dtMu.Lock()
	defer st.dtMu.Unlock()
	return len(st.dt) == 1
}

func (st *stream[T]) None() bool {
	st.dtMu.Lock()
	defer st.dtMu.Unlock()
	return len(st.dt) == 0
}

func (st *stream[T]) Count() int {
	return len(st.dt)
}

func (st *stream[T]) Sum() int {
	// TODO: implement
	return 0
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

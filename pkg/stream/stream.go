package stream

import (
	"sort"
	"sync"

	"github.com/koss-null/lambda/internal/tools"
)

type (
	stream[T any] struct {
		onceRun sync.Once

		ops        []Operation[T]
		curThreads uint
		dt         []T
		bm         *tools.Bitmask[T]
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
		// Reduce takes the result of the prev operation and applies it to the next item
		Reduce(func(T, T) T) StreamI[T]
		// Filteer
		Filter(func(T) bool) StreamI[T]
		// Sort sotrs the underlying array
		Sort(less func(a, b T) bool) StreamI[T]
		// Split splits initial slice into multiple
		Split(func(a T) []T) StreamI[T]

		// Config functions
		// Go splits execution into n goroutines, if multiple Go functions are along
		// the stream pipeline, it applies as soon as it's met in the sequence
		// if n is 0 (which is by default), n is chousen automatically
		Go(n uint) StreamI[T]

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
		Sum(func(int64, T) int64) int64
		// Contains returns true if element is in the array
		Contains(item T, eq func(T, T) bool) bool
	}
)

// Stream creates new instance of stream
func Stream[T any](data []T) StreamI[T] {
	dtCopy := make([]T, len(data))
	copy(dtCopy, data)
	bm := tools.Bitmask[T]{}
	bm.PutLine(0, uint(len(data)), true)
	return &stream[T]{dt: dtCopy, bm: &bm}
}

// S is a shortened Stream
func S[T any](data []T) StreamI[T] {
	return Stream(data)
}

func (st *stream[T]) Skip(n uint) StreamI[T] {
	return st
}

func (st *stream[T]) Trim(n uint) StreamI[T] {
	return st
}

func mapFn[T any](fn func(T) T) func(dt []T, bm *tools.Bitmask[T]) {
	return func(dt []T, bm *tools.Bitmask[T]) {
		for i := range dt {
			if bm.Get(uint(i)) {
				dt[i] = fn(dt[i])
			}
		}
	}
}

func (st *stream[T]) Map(fn func(T) T) StreamI[T] {
	st.ops = append(st.ops, Operation[T]{
		GrtCnt: st.curThreads,
		op:     OpTypeMap,
		fn:     mapFn(fn),
	})
	return st
}

func reduceFn[T any](fn func(T, T) T) func(dt []T, bm *tools.Bitmask[T]) {
	return func(dt []T, bm *tools.Bitmask[T]) {
		firstIdx := -1
		for i := range dt {
			if bm.Get(uint(i)) {
				if firstIdx == -1 {
					firstIdx = i
					continue
				}
				dt[firstIdx] = fn(dt[firstIdx], dt[i])
				bm.Put(uint(i), false)
			}
		}
	}
}

func (st *stream[T]) Reduce(fn func(T, T) T) StreamI[T] {
	st.ops = append(st.ops, Operation[T]{
		GrtCnt: st.curThreads,
		op:     OpTypeReduce,
		fn:     reduceFn(fn),
	})
	return st
}

func filterFn[T any](fn func(T) bool) func(dt []T, bm *tools.Bitmask[T]) {
	return func(dt []T, bm *tools.Bitmask[T]) {
		for i := range dt {
			if bm.Get(uint(i)) {
				if !fn(dt[i]) {
					bm.Put(uint(i), false)
				}
			}
		}
	}
}

func (st *stream[T]) Filter(fn func(T) bool) StreamI[T] {
	st.ops = append(st.ops, Operation[T]{
		GrtCnt: st.curThreads,
		op:     OpTypeFilter,
		fn:     filterFn(fn),
	})
	return st
}

func sortFn[T any](less func(a, b T) bool) func(dt []T, bm *tools.Bitmask[T]) {
	return func(dt []T, bm *tools.Bitmask[T]) {
		sort.Slice(dt,
			func(i, j int) bool {
				return less(dt[i], dt[j])
			},
		)
	}
}

// Sort adds two functions: first one sorts everything
func (st *stream[T]) Sort(less func(a, b T) bool) StreamI[T] {
	// FIXME: sort does not work nike this
	st.ops = append(st.ops, Operation[T]{
		GrtCnt: st.curThreads,
		op:     OpTypeSort,
		fn:     sortFn(less),
	})
	return st
}

func (st *stream[T]) Go(n uint) StreamI[T] {
	st.ops = append(st.ops, Operation[T]{
		GrtCnt: 1,
		op:     OpTypeGo,
		fn: func(_ []T, _ *tools.Bitmask[T]) {
			st.curThreads = n
		},
	})
	return st
}

func splitFn[T any](fn func(T) []T) func(dt []T, bm *tools.Bitmask[T]) {
	// WATAFCK
	return nil
}

func (st *stream[T]) Split(fn func(T) []T) StreamI[T] {
	st.ops = append(st.ops, Operation[T]{
		GrtCnt: st.curThreads,
		op:     OpTypeGo,
		fn:     splitFn(fn),
	})
	return st
}

func (st *stream[T]) Slice() []T {
}

func (st *stream[T]) Any() bool {
}

func (st *stream[T]) None() bool {
}

func (st *stream[T]) Count() int {
}

func (st *stream[T]) Sum(sum func(int64, T) int64) int64 {
}

// Contains returns if the underlying array contains an item
func (st *stream[T]) Contains(item T, eq func(a, b T) bool) bool {
}

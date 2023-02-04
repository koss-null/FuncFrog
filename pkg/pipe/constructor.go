package pipe

import (
	"golang.org/x/exp/constraints"

	"github.com/koss-null/lambda/internal/primitive/pointer"
)

// Slice creates a Pipe from a slice
func Slice[T any](dt []T) *Pipe[T] {
	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			dtCp := make([]T, len(dt))
			copy(dtCp, dt)
			return func(i int) (*T, bool) {
				if i >= len(dtCp) {
					return nil, true
				}
				return &dtCp[i], false
			}
		},
		len:      pointer.To(len(dt)),
		valLim:   pointer.To(0),
		parallel: pointer.To(defaultParallelWrks),
		prlSet:   pointer.To(false),
	}
}

// Func creates a lazy sequence d[i] = fn(i).
// fn is a function, that returns an object(T) and does it exist(bool).
// Initiating the pipe from a func you have to set either the output value
// amount using Get(n int) or the amount of generated values Gen(n int), or set
// the limit predicate Until(func(x T) bool).
func Func[T any](fn func(i int) (T, bool)) *PipeNL[T] {
	return &PipeNL[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				obj, exist := fn(i)
				return &obj, !exist
			}
		},
		len:      pointer.To(-1),
		valLim:   pointer.To(0),
		parallel: pointer.To(defaultParallelWrks),
		prlSet:   pointer.To(false),
	}
}

// Fu creates a lazy sequence d[i] = fn(i).
// fn is a shortened version of Func where the second argument is true by default
// Initiating the pipe from a func you have to set either the output value
// amount using Get(n int) or the amount of generated values Gen(n int), or set
// the limit predicate Until(func(x T) bool).
func Fn[T any](fn func(i int) T) *PipeNL[T] {
	return Func(func(i int) (T, bool) {
		obj := fn(i)
		return obj, true
	})
}

// Cycle creates new pipe that cycles through the elements of the provided slice.
// Initiating the pipe from a func you have to set either the output value
// amount using Get(n int) or the amount of generated values Gen(n int), or set
// the limit predicate Until(func(x T) bool).
func Cycle[T any](a []T) *PipeNL[T] {
	return Fn(func(i int) T {
		return a[i%len(a)]
	})
}

// Range creates a slice [start, finish) with a provided step.
// Pipe initialized with Range can be considered as the one madi with Slice(range() []T).
func Range[T constraints.Integer | constraints.Float](start, finish, step T) *Pipe[T] {
	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				val := start + T(i)*step
				return &val, val < finish
			}
		},
		len:      pointer.To(int((finish - start) / step)),
		valLim:   pointer.To(0),
		parallel: pointer.To(defaultParallelWrks),
		prlSet:   pointer.To(false),
	}
}

package internalpipe

import (
	"math"

	"golang.org/x/exp/constraints"
)

const (
	defaultParallelWrks = 1
	notSet              = -1
)

func Slice[T any](dt []T) Pipe[T] {
	dtCp := make([]T, len(dt))
	copy(dtCp, dt)

	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			if i >= len(dtCp) {
				return nil, true
			}
			return &dtCp[i], false
		},
		Len:           len(dtCp),
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Func[T any](fn func(i int) (T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			obj, exist := fn(i)
			return &obj, !exist
		},
		Len:           notSet,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func FuncP[T any](fn func(i int) (*T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			obj, exist := fn(i)
			return obj, !exist
		},
		Len:           notSet,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Cycle[T any](a []T) Pipe[T] {
	if len(a) == 0 {
		return Func(func(_ int) (x T, exist bool) {
			return
		})
	}

	return Func(func(i int) (T, bool) {
		return a[i%len(a)], true
	})
}

func Range[T constraints.Integer | constraints.Float](start, finish, step T) Pipe[T] {
	if (step == 0) ||
		(step > 0 && start >= finish) ||
		(step < 0 && finish >= start) {
		return Pipe[T]{
			Fn:            nil,
			Len:           0,
			ValLim:        notSet,
			GoroutinesCnt: defaultParallelWrks,
		}
	}

	pred := func(x T) bool {
		return x >= finish
	}
	if step < 0 {
		pred = func(x T) bool {
			return x <= finish
		}
	}
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			val := start + T(i)*step
			return &val, pred(val)
		},
		Len:           ceil(float64(finish-start) / float64(step)),
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Repeat[T any](x T, n int) Pipe[T] {
	if n <= 0 {
		return Pipe[T]{
			Fn:            nil,
			Len:           0,
			ValLim:        notSet,
			GoroutinesCnt: defaultParallelWrks,
		}
	}

	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			cp := x
			return &cp, i >= n
		},
		Len:           n,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func ceil[T constraints.Integer | constraints.Float](a T) int {
	return int(math.Ceil(float64(a)))
}

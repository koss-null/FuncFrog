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
	return Pipe[T]{
		Fn: func(i int, dst *T) bool {
			if i >= len(dt) {
				return true
			}
			*dst = dt[i]
			return false
		},
		Len:           len(dt),
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Func[T any](fn func(i int) (T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int, dst *T) bool {
			obj, exist := fn(i)
			if !exist {
				return true
			}
			*dst = obj
			return false
		},
		Len:           notSet,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func FuncP[T any](fn func(i int) (*T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int, dst *T) bool {
			obj, exist := fn(i)
			if !exist {
				return true
			}
			*dst = *obj
			return false
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
		Fn: func(i int, dst *T) bool {
			val := start + T(i)*step
			*dst = val
			return pred(val)
		},
		Len:           rangeLen(start, finish, step, pred),
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
		Fn: func(i int, dst *T) bool {
			*dst = x
			return i >= n
		},
		Len:           n,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

// rangeLen returns the amount of values Range will produce. It starts from the
// float estimate and corrects it against the very predicate the pipe uses, so
// Len always agrees with Fn and float rounding cannot drop the last element.
func rangeLen[T constraints.Integer | constraints.Float](start, finish, step T, pred func(T) bool) int {
	n := int(math.Ceil(float64(finish-start)/float64(step))) + 1
	if n < 0 {
		n = 0
	}
	for n > 0 && pred(start+T(n-1)*step) {
		n--
	}
	return n
}

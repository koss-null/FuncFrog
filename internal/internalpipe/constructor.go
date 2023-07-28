package internalpipe

import "golang.org/x/exp/constraints"

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
		Fn:            fn,
		Len:           notSet,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Range[T constraints.Integer | constraints.Float](start, finish, step T) Pipe[T] {
	if start >= finish {
		return Pipe[T]{
			Fn: func(int) (*T, bool) {
				return nil, true
			},
			Len:           0,
			ValLim:        notSet,
			GoroutinesCnt: defaultParallelWrks,
		}
	}
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			val := start + T(i)*step
			return &val, val >= finish
		},
		Len:           max(int((finish-start)/step), 1),
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}
}

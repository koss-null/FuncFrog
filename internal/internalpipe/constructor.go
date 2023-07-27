package internalpipe

import "golang.org/x/exp/constraints"

const defaultParallelWrks = 1

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
		ValLim:        -1,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Func[T any](fn func(i int) (T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			obj, exist := fn(i)
			return &obj, !exist
		},
		Len:           -1,
		ValLim:        -1,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func FuncP[T any](fn func(i int) (*T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn:            fn,
		Len:           -1,
		ValLim:        -1,
		GoroutinesCnt: defaultParallelWrks,
	}
}

func Range[T constraints.Integer | constraints.Float](start, finish, step T) Pipe[T] {
	if start+step >= finish {
		return Pipe[T]{
			Fn: func(int) (*T, bool) {
				return nil, true
			},
			Len:           0,
			ValLim:        -1,
			GoroutinesCnt: defaultParallelWrks,
		}
	}
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			val := start + T(i)*step
			return &val, val < finish
		},
		Len:           int((finish - start) / step),
		ValLim:        -1,
		GoroutinesCnt: defaultParallelWrks,
	}
}

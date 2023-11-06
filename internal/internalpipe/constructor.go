package internalpipe

import (
	"unsafe"

	"golang.org/x/exp/constraints"
)

const (
	defaultParallelWrks = 1
	notSet              = -1
)

func Slice[T any](dt []T) Pipe[T] {
	dtCp := make([]T, len(dt))
	copy(dtCp, dt)
	p := Pipe[T]{
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

	ptr := uintptr(unsafe.Pointer(&p))
	p.prevP = ptr
	return p
}

func Func[T any](fn func(i int) (T, bool)) Pipe[T] {
	p := Pipe[T]{
		Fn: func(i int) (*T, bool) {
			obj, exist := fn(i)
			return &obj, !exist
		},
		Len:           notSet,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}

	ptr := uintptr(unsafe.Pointer(&p))
	p.prevP = ptr
	return p
}

func FuncP[T any](fn func(i int) (*T, bool)) Pipe[T] {
	p := Pipe[T]{
		Fn:            fn,
		Len:           notSet,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}

	ptr := uintptr(unsafe.Pointer(&p))
	p.prevP = ptr
	return p
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
	if start >= finish {
		p := Pipe[T]{
			Fn: func(int) (*T, bool) {
				return nil, true
			},
			Len:           0,
			ValLim:        notSet,
			GoroutinesCnt: defaultParallelWrks,
		}

		ptr := uintptr(unsafe.Pointer(&p))
		p.prevP = ptr
		return p
	}

	p := Pipe[T]{
		Fn: func(i int) (*T, bool) {
			val := start + T(i)*step
			return &val, val >= finish
		},
		Len:           max(int((finish-start)/step), 1),
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}

	ptr := uintptr(unsafe.Pointer(&p))
	p.prevP = ptr
	return p
}

func Repeat[T any](x T, n int) Pipe[T] {
	if n <= 0 {
		p := Pipe[T]{
			Fn: func(int) (*T, bool) {
				return nil, true
			},
			Len:           0,
			ValLim:        notSet,
			GoroutinesCnt: defaultParallelWrks,
		}

		ptr := uintptr(unsafe.Pointer(&p))
		p.prevP = ptr
		return p
	}

	p := Pipe[T]{
		Fn: func(i int) (*T, bool) {
			cp := x
			return &cp, i >= n
		},
		Len:           n,
		ValLim:        notSet,
		GoroutinesCnt: defaultParallelWrks,
	}

	ptr := uintptr(unsafe.Pointer(&p))
	p.prevP = ptr
	return p
}

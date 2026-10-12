package internalpipe

import "math"

type AccumFn[T any] func(*T, *T) T

// Reduce applies the result of a function to each element one-by-one: f(p[n], f(p[n-1], f(p[n-2, ...]))).
func (p Pipe[T]) Reduce(fn AccumFn[T]) *T {
	if p.GoroutinesCnt > 1 {
		// Fold over the parallel evaluation result.
		return reduceSlice(p.Do(), fn)
	}

	var (
		res T
		set bool
	)
	fold := func(obj *T) {
		if set {
			res = fn(&res, obj)
		} else {
			res = *obj
			set = true
		}
	}

	var v T
	if p.limitSet() {
		for i, got := 0, 0; got < p.ValLim; i++ {
			if !p.Fn(i, &v) {
				fold(&v)
				got++
			}
			if i == math.MaxInt {
				panic(panicLimitExceededMsg)
			}
		}
	} else {
		limit := p.limit()
		for i := 0; i < limit; i++ {
			if !p.Fn(i, &v) {
				fold(&v)
			}
		}
	}

	if !set {
		return nil
	}
	return &res
}

func reduceSlice[T any](data []T, fn AccumFn[T]) *T {
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &data[0]
	default:
		res := data[0]
		for _, val := range data[1:] {
			res = fn(&res, &val)
		}
		return &res
	}
}

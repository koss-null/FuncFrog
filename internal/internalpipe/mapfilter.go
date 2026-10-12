package internalpipe

// MapFilter applies given function to each element of the underlying slice,
// if the second returning value of fn is false, the element is skipped (may be useful for error handling).
// returns the slice where each element is n[i] = f(p[i]) if it is not skipped.
func (p Pipe[T]) MapFilter(fn func(T) (T, bool)) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int, dst *T) bool {
			if p.Fn(i, dst) {
				return true
			}
			res, take := fn(*dst)
			*dst = res
			return !take
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		y: p.y,
	}
}

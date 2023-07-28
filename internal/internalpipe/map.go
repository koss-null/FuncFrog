package internalpipe

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i]).
func (p Pipe[T]) Map(fn func(T) T) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			if obj, skipped := p.Fn(i); !skipped {
				res := fn(*obj)
				return &res, false
			}
			return nil, true
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,
	}
}

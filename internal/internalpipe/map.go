package internalpipe

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i]).
func (p Pipe[T]) Map(fn func(T) T) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int, dst *T) bool {
			if p.Fn(i, dst) {
				return true
			}
			*dst = fn(*dst)
			return false
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		y: p.y,
	}
}

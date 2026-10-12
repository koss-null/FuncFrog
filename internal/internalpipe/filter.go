package internalpipe

// Filter leaves only items with true predicate fn.
func (p Pipe[T]) Filter(fn func(*T) bool) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int, dst *T) bool {
			if p.Fn(i, dst) {
				return true
			}
			return !fn(dst)
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		y: p.y,
	}
}

package internalpipe

// Filter leaves only items with true predicate fn.
func (p Pipe[T]) Filter(fn func(*T) bool) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			if obj, skipped := p.Fn(i); !skipped {
				if !fn(obj) {
					return nil, true
				}
				return obj, false
			}
			return nil, true
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		y: p.y,
	}
}

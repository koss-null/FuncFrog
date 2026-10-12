package internalpipe

func (p Pipe[T]) Erase() Pipe[any] {
	return Pipe[any]{
		Fn: func(i int, dst *any) bool {
			var obj T
			if p.Fn(i, &obj) {
				return true
			}
			// Keep the historical behaviour: the erased value stays a *T,
			// so Collect can assert it back to *DstT.
			*dst = any(&obj)
			return false
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		y: p.y,
	}
}

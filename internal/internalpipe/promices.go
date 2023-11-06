package internalpipe

func (p Pipe[T]) Promices() []func() (T, bool) {
	proms := make([]func() (T, bool), p.limit())
	var empty T
	for i := 0; i < p.limit(); i++ {
		cpi := i
		proms[i] = func() (T, bool) {
			obj, skipped := p.Fn(cpi)
			if skipped {
				return empty, false
			}
			return *obj, true
		}
	}
	return proms
}

package pipe

// PrefixPipes are more powerfull and more ugly
// You can use all operations the same way but the result type of any operation can be arbitary

func Map[SrcT, DstT any](p *Pipe[SrcT], fn func(x SrcT) DstT) *Pipe[DstT] {
	return &Pipe[DstT]{
		fn: func() func(i int) (*DstT, bool) {
			return func(i int) (*DstT, bool) {
				if obj, skipped := p.fn()(i); !skipped {
					dst := fn(*obj)
					return &dst, false
				}
				return nil, true
			}
		},
		len:      p.len,
		valLim:   p.valLim,
		skip:     p.skip,
		parallel: p.parallel,
	}
}

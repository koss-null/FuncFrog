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
		parallel: p.parallel,
	}
}

func Reduce[SrcT any, DstT any](p *Pipe[SrcT], fn func(DstT, SrcT) DstT, initVal DstT) DstT {
	data := p.Do()
	switch len(data) {
	case 0:
		return initVal
	case 1:
		return fn(initVal, data[0])
	default:
		res := fn(initVal, data[0])
		for i := range data[1:] {
			res = fn(res, data[i])
		}
		return res
	}
}

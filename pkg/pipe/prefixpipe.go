package pipe

// Map applies function on a Pipe of type SrcT and returns a Pipe of type DstT.
func Map[SrcT, DstT any](p Piper[SrcT], fn func(x SrcT) DstT) Piper[DstT] {
	derivedPipe := p.(*Pipe[SrcT])
	return &Pipe[DstT]{
		fn: func() func(i int) (*DstT, bool) {
			return func(i int) (*DstT, bool) {
				if obj, skipped := derivedPipe.fn()(i); !skipped {
					dst := fn(*obj)
					return &dst, false
				}
				return nil, true
			}
		},
		len:      derivedPipe.len,
		valLim:   derivedPipe.valLim,
		parallel: derivedPipe.parallel,
	}
}

// Reduce applies reduce operation on Pipe of type SrcT an returns result of type DstT.
// initVal is an optional parameter to initialize a value that should be used on the first step of reduce.
func Reduce[SrcT any, DstT any](p Piper[SrcT], fn func(DstT, SrcT) DstT, initVal ...DstT) DstT {
	var init DstT
	if len(initVal) > 0 {
		init = initVal[0]
	}
	data := p.Do()
	switch len(data) {
	case 0:
		return init
	case 1:
		return fn(init, data[0])
	default:
		res := fn(init, data[0])
		for i := range data[1:] {
			res = fn(res, data[i])
		}
		return res
	}
}

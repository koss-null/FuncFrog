package pipe

import "github.com/koss-null/lambda/internal/internalpipe"

type entrailsPipe[T any] interface {
	Piper[T]
	Entrails() *internalpipe.Pipe[T]
}

type anyPipe[T any] interface {
	Pipe[T] | PipeNL[T]
}

// Map applies function on a Pipe of type SrcT and returns a Pipe of type DstT.
func Map[SrcT, DstT any](
	p Piper[SrcT],
	fn func(x SrcT) DstT,
) Piper[DstT] {
	pp := p.(entrailsPipe[SrcT]).Entrails()
	return &Pipe[DstT]{internalpipe.Pipe[DstT]{
		Fn: func() func(i int) (*DstT, bool) {
			return func(i int) (*DstT, bool) {
				if obj, skipped := pp.Fn()(i); !skipped {
					dst := fn(*obj)
					return &dst, false
				}
				return nil, true
			}
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
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

package pipe

import "github.com/koss-null/funcfrog/internal/internalpipe"

type entrails[T any] interface {
	Entrails() *internalpipe.Pipe[T]
}

// Map applies function on a Piper of type SrcT and returns a Pipe of type DstT.
func Map[SrcT any, DstT any](
	p Piper[SrcT],
	fn func(x SrcT) DstT,
) Piper[DstT] {
	pp := any(p).(entrails[SrcT]).Entrails()
	return &Pipe[DstT]{internalpipe.Pipe[DstT]{
		Fn: func(i int) (*DstT, bool) {
			if obj, skipped := pp.Fn(i); !skipped {
				dst := fn(*obj)
				return &dst, false
			}
			return nil, true
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
}

// MapNL applies function on a PiperNoLen of type SrcT and returns a Pipe of type DstT.
func MapNL[SrcT, DstT any](
	p PiperNoLen[SrcT],
	fn func(x SrcT) DstT,
) PiperNoLen[DstT] {
	pp := any(p).(entrails[SrcT]).Entrails()
	return &PipeNL[DstT]{internalpipe.Pipe[DstT]{
		Fn: func(i int) (*DstT, bool) {
			if obj, skipped := pp.Fn(i); !skipped {
				dst := fn(*obj)
				return &dst, false
			}
			return nil, true
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
}

// MapFilter applies function on a Piper of type SrcT and returns a Pipe of type DstT.
// fn returns a value of DstT type and true if this value is not skipped.
func MapFilter[SrcT, DstT any](
	p Piper[SrcT],
	fn func(x SrcT) (DstT, bool),
) Piper[DstT] {
	pp := any(p).(entrails[SrcT]).Entrails()
	return &Pipe[DstT]{internalpipe.Pipe[DstT]{
		Fn: func(i int) (*DstT, bool) {
			if obj, skipped := pp.Fn(i); !skipped {
				dst, exist := fn(*obj)
				return &dst, !exist
			}
			return nil, true
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
}

// MapFilterNL applies function on a PiperNoLen of type SrcT and returns a Pipe of type DstT.
// fn returns a value of DstT type and true if this value is not skipped.
func MapFilterNL[SrcT, DstT any](
	p PiperNoLen[SrcT],
	fn func(x SrcT) (DstT, bool),
) PiperNoLen[DstT] {
	pp := any(p).(entrails[SrcT]).Entrails()
	return &PipeNL[DstT]{internalpipe.Pipe[DstT]{
		Fn: func(i int) (*DstT, bool) {
			if obj, skipped := pp.Fn(i); !skipped {
				dst, exist := fn(*obj)
				return &dst, !exist
			}
			return nil, true
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
}

// Reduce applies reduce operation on Pipe of type SrcT and returns result of type DstT.
// initVal is an optional parameter to initialize a value that should be used on the first step of reduce.
func Reduce[SrcT, DstT any](p Piper[SrcT], fn func(*DstT, *SrcT) DstT, initVal ...DstT) DstT {
	var init DstT
	if len(initVal) > 0 {
		init = initVal[0]
	}
	data := p.Do()
	switch len(data) {
	case 0:
		return init
	case 1:
		return fn(&init, &data[0])
	default:
		res := fn(&init, &data[0])
		for i := range data[1:] {
			res = fn(&res, &data[i+1])
		}
		return res
	}
}

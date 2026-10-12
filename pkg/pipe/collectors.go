package pipe

import "github.com/koss-null/funcfrog/internal/internalpipe"

// Collect translates Piper with erased type (achieved by calling an Erase method of any type).
func Collect[DstT any](p Piper[any]) Piper[DstT] {
	pp := any(p).(entrails[any]).Entrails()
	return &Pipe[DstT]{internalpipe.Pipe[DstT]{
		Fn: func(i int, dst *DstT) bool {
			var obj any
			if pp.Fn(i, &obj) {
				return true
			}
			if p, ok := obj.(*DstT); ok {
				*dst = *p
				return false
			}
			return true
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
}

// CollectNL translates PiperNL with erased type (achieved by calling an Erase method of any type).
func CollectNL[DstT any](p PiperNoLen[any]) PiperNoLen[DstT] {
	pp := any(p).(entrails[any]).Entrails()
	return &PipeNL[DstT]{internalpipe.Pipe[DstT]{
		Fn: func(i int, dst *DstT) bool {
			var obj any
			if pp.Fn(i, &obj) {
				return true
			}
			if p, ok := obj.(*DstT); ok {
				*dst = *p
				return false
			}
			return true
		},
		Len:           pp.Len,
		ValLim:        pp.ValLim,
		GoroutinesCnt: pp.GoroutinesCnt,
	}}
}

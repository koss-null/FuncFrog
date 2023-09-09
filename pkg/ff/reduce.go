package ff

import "github.com/koss-null/funcfrog/pkg/pipe"

func Reduce[SrcT any, DstT any](a []SrcT, fn func(*DstT, *SrcT) DstT, initVal ...DstT) DstT {
	return pipe.Reduce(pipe.Slice(a), fn)
}

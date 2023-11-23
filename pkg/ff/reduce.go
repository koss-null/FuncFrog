package ff

import "github.com/koss-null/funcfrog/pkg/pipe"

// Reduce is a short way to create DstT value from a slice of SrcT applying Reduce function fn.
func Reduce[SrcT any, DstT any](a []SrcT, fn func(*DstT, *SrcT) DstT, initVal ...DstT) DstT {
	return pipe.Reduce(pipe.Slice(a), fn)
}

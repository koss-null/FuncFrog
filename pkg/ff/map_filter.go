package ff

import "github.com/koss-null/funcfrog/pkg/pipe"

// MapFilter is a short way to create a Pipe of DstT from a slice of SrcT applying MapFilter function fn.
func MapFilter[SrcT, DstT any](a []SrcT, fn func(SrcT) (DstT, bool)) pipe.Piper[DstT] {
	return pipe.MapFilter(pipe.Slice(a), fn)
}

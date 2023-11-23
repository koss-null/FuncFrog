package ff

import "github.com/koss-null/funcfrog/pkg/pipe"

// Map is a short way to create a Pipe of DstT from a slice of SrcT applying Map function fn.
func Map[SrcT, DstT any](a []SrcT, fn func(SrcT) DstT) pipe.Piper[DstT] {
	return pipe.Map(pipe.Slice(a), fn)
}

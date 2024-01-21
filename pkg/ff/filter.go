package ff

import "github.com/koss-null/funcfrog/pkg/pipe"

// Filter is a short way to create a Pipe from a slice of SrcT applying Filter function fn.
func Filter[SrcT any](a []SrcT, fn func(*SrcT) bool) pipe.Piper[SrcT] {
	return pipe.Slice(a).Filter(fn)
}

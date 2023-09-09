package ff

import "github.com/koss-null/funcfrog/pkg/pipe"

func Map[SrcT, DstT any](a []SrcT, fn func(SrcT) DstT) pipe.Piper[DstT] {
	return pipe.Map(pipe.Slice(a), fn)
}

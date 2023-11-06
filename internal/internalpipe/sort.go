package internalpipe

import (
	"sync"

	"github.com/koss-null/funcfrog/internal/algo/parallel/qsort"
)

// Sort sorts the underlying slice on a current step of a pipeline.
func (p Pipe[T]) Sort(less func(*T, *T) bool) Pipe[T] {
	var once sync.Once
	var sorted []T

	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			once.Do(func() {
				data := p.Do()
				if len(data) == 0 {
					return
				}
				sorted = qsort.Sort(data, less, p.GoroutinesCnt)
			})
			if i >= len(sorted) {
				return nil, true
			}
			return &sorted[i], false
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		y: p.y,
	}
}

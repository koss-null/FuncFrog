package stream

import (
	"github.com/koss-null/lambda-go/internal/tools"
	"github.com/koss-null/lambda/pkg/funcmodel"
)

const (
	minSplitLen = 1024
)

type Operation[T any] struct {
	// sync if true, all op runs in single goroutine
	sync bool
	fn   func(dt []T, bm *tools.Bitmask)
	// finished executes after the operation was finished
	finished funcmodel.Empty
}

// Do runs operation on n goroutines
func (op *Operation) Do(dt []T, bm *tools.Bitmask) {
	op.fn(dt, bm)
}

func (op *Operation) Sync() bool {
	return op.sync
}

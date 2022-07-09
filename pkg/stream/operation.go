package stream

import (
	"github.com/google/uuid"

	"github.com/koss-null/lambda/internal/tools"
	"github.com/koss-null/lambda/pkg/funcmodel"
)

const (
	minSplitLen = 1024
)

type Operation[T any] struct {
	id uuid.UUID
	// sync if true, all op runs in single goroutine
	sync bool
	fn   func(dt []T, bm *tools.Bitmask[T])
	// finished executes after the operation was finished
	finished funcmodel.Empty
}

// Do runs operation on n goroutines
func (op *Operation[T]) Do(dt []T, bm *tools.Bitmask[T]) {
	op.fn(dt, bm)
}

func (op *Operation[T]) Sync() bool {
	return op.sync
}

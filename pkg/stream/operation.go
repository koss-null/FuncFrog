package stream

import (
	"github.com/koss-null/lambda/internal/tools"
)

const (
	minSplitLen = 1024
)

type opType string

const (
	OpTypeMap      opType = "map"
	OpTypeReduce          = "reduce"
	OpTypeFilter          = "filter"
	OpTypeSplit           = "split"
	OpTypeSort            = "sort"
	OpTypeSkip            = "skip"
	OpTypeTrim            = "trim"
	OpTypeSum             = "sum"
	OpTypeContains        = "contains"
	OpTypeGo              = "go"
)

var singleThreadOps = map[opType]struct{}{
	OpTypeReduce: {},
	OpTypeSkip:   {},
	OpTypeTrim:   {},
	OpTypeGo:     {},
}

type Operation[T any] struct {
	GrtCnt uint

	op opType
	fn func(dt []T, bm *tools.Bitmask[T])
}

// Do runs operation on n goroutines
func (op *Operation[T]) Do(dt []T, bm *tools.Bitmask[T]) {
	op.fn(dt, bm)
}

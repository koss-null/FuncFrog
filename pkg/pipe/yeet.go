package pipe

import (
	"github.com/koss-null/funcfrog/internal/internalpipe"
)

const (
	initErrsAmount     = 10
	initHandlersAmount = 5
)

// NewYeti creates a brand new Yeti - an object for error handling.
func NewYeti() internalpipe.YeetSnag {
	return internalpipe.NewYeti()
}

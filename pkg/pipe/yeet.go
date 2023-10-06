package pipe

import (
	"sync"

	"github.com/koss-null/funcfrog/internal/internalpipe"
)

const (
	initErrsAmount     = 10
	initHandlersAmount = 5
)

type Yeti interface {
	// yeet an error
	Yeet(err error)
	// snag and handle the error
	Snag(handler func(err error))
}

// FIXME
func Yeet() Yeti {
	return &internalpipe.Yeti{
		Errs:     make(map[*any][]error, initErrsAmount),
		Handlers: make(map[*any][]internalpipe.ErrHandler, initHandlersAmount),
		Mx:       &sync.Mutex{},
	}
}

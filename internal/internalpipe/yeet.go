package internalpipe

import (
	"sync"
)

type ErrHandler func(error)

type Yeti struct {
	eMx      *sync.Mutex
	errs     []error
	hMx      *sync.Mutex
	handlers []ErrHandler
	yMx      *sync.Mutex
	yetis    []yeti
}

func NewYeti() *Yeti {
	const yetiExpectedErrors = 6
	return &Yeti{
		errs:     make([]error, 0, yetiExpectedErrors),
		handlers: make([]ErrHandler, 0, yetiExpectedErrors),
		yetis:    make([]yeti, 0),
		eMx:      &sync.Mutex{},
		hMx:      &sync.Mutex{},
		yMx:      &sync.Mutex{},
	}
}

func (y *Yeti) Yeet(err error) {
	y.eMx.Lock()
	y.errs = append(y.errs, err)
	y.eMx.Unlock()
}

func (y *Yeti) Snag(handler ErrHandler) {
	y.hMx.Lock()
	y.handlers = append(y.handlers, handler)
	y.hMx.Unlock()
}

func (y *Yeti) Handle() {
	y.yMx.Lock()
	prevYs := y.yetis
	y.yMx.Unlock()
	for _, prevYetti := range prevYs {
		prevYetti.Handle()
	}

	y.hMx.Lock()
	y.eMx.Lock()
	defer y.hMx.Unlock()
	defer y.eMx.Unlock()

	for _, err := range y.errs {
		for _, handle := range y.handlers {
			handle(err)
		}
	}
}

func (y *Yeti) AddYeti(yt yeti) {
	y.yMx.Lock()
	y.yetis = append(y.yetis, yt)
	y.yMx.Unlock()
}

type yeti interface {
	Yeet(err error)
	Snag(h ErrHandler)
	// TODO: Handle should be called after each Pipe function eval
	Handle()
	AddYeti(y yeti)
}

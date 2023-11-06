package internalpipe

import (
	"sync"
)

type ErrHandler func(error)

type Yeti struct {
	EMx      *sync.Mutex
	errs     []error
	HMx      *sync.Mutex
	handlers []ErrHandler
	YMx      *sync.Mutex
	yetties  []yeti
}

func NewYeti() *Yeti {
	return &Yeti{
		errs:     make([]error, 0),
		handlers: make([]ErrHandler, 0),
		yetties:  make([]yeti, 0),
		EMx:      &sync.Mutex{},
		HMx:      &sync.Mutex{},
		YMx:      &sync.Mutex{},
	}
}

func (y *Yeti) Yeet(err error) {
	y.EMx.Lock()
	y.errs = append(y.errs, err)
	y.EMx.Unlock()
}

func (y *Yeti) Snag(handler ErrHandler) {
	y.HMx.Lock()
	y.handlers = append(y.handlers, handler)
	y.HMx.Unlock()
}

func (y *Yeti) Handle() {
	if y == nil {
		return
	}

	y.YMx.Lock()
	prevYs := y.yetties
	y.YMx.Unlock()
	for _, prevYetti := range prevYs {
		prevYetti.Handle()
	}

	y.HMx.Lock()
	y.EMx.Lock()
	defer y.HMx.Unlock()
	defer y.EMx.Unlock()

	for _, err := range y.errs {
		for _, handle := range y.handlers {
			handle(err)
		}
	}
}

func (y *Yeti) AddYeti(yt yeti) {
	y.YMx.Lock()
	y.yetties = append(y.yetties, yt)
	y.YMx.Unlock()
}

type yeti interface {
	Yeet(err error)
	Snag(h ErrHandler)
	// TODO: Handle should be called after each Pipe function eval
	Handle()
	AddYeti(y yeti)
}

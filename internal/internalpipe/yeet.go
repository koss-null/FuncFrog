package internalpipe

import (
	"sync"
	"unsafe"
)

type ErrHandler func(err error)

type Yeti struct {
	Errs       []error
	Handlers   []ErrHandler
	Pipe2Hdlrs map[unsafe.Pointer][]ErrHandler
	EMx        *sync.Mutex
	HMx        *sync.Mutex
}

func NewYeti() *Yeti {
	return &Yeti{
		Errs:       make([]error, 0),
		Handlers:   make([]ErrHandler, 0),
		Pipe2Hdlrs: make(map[unsafe.Pointer][]ErrHandler),
		EMx:        &sync.Mutex{},
		HMx:        &sync.Mutex{},
	}
}

func (y *Yeti) Yeet(err error) {
	y.EMx.Lock()
	y.Errs = append(y.Errs, err)
	y.EMx.Unlock()
}

func (y *Yeti) Snag(handler ErrHandler) {
	y.Handlers = append(y.Handlers, handler)
}

func (y *Yeti) SnagPipe(p unsafe.Pointer, h ErrHandler) {
	y.HMx.Lock()
	if hdlrs, ok := y.Pipe2Hdlrs[p]; ok {
		y.Pipe2Hdlrs[p] = append(hdlrs, h)
	} else {
		// FIXME: use constant, remove else
		y.Pipe2Hdlrs[p] = append(make([]ErrHandler, 0, 10), h)
	}
	y.HMx.Unlock()
}

func (y *Yeti) Handle(p unsafe.Pointer) {
	// TODO: impl
}

type yeti interface {
	Yeet(err error)
	SnagPipe(p unsafe.Pointer, h ErrHandler)
	// TODO: Handle should be called after each Pipe function eval
	Handle(p unsafe.Pointer)
}

package internalpipe

import "unsafe"

func (p Pipe[T]) Erase() Pipe[any] {
	return Pipe[any]{
		Fn: func(i int) (*any, bool) {
			if obj, skipped := p.Fn(i); !skipped {
				anyObj := any(obj)
				return &anyObj, false
			}
			return nil, true
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,

		prevP: uintptr(unsafe.Pointer(&p)),
		y:     p.y,
	}
}

package internalpipe

// Sang ads error handler to a current Pipe step.
func (p Pipe[T]) Snag(h ErrHandler) Pipe[T] {
	if p.y == nil {
		return p
	}

	p.y.Snag(h)
	return p
}

type YeetSnag interface {
	// yeet an error
	Yeet(err error)
	// snag and handle the error
	Snag(ErrHandler)
}

// Yeti adds Yeti error handler to the pipe.
// If some other handlers were set before, they are handled by the Snag
func (p Pipe[T]) Yeti(y YeetSnag) Pipe[T] {
	yet := y.(yeti)
	if p.y != nil {
		yet.AddYeti(p.y)
	}
	p.y = yet
	return p
}

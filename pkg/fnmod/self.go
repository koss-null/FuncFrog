package fnmod

type SelfFn[T any] func() T

// Self wraps any object into a function returning this object
func Self[T any](x T) SelfFn[T] {
	return func() T {
		return x
	}
}

func (fn SelfFn[T]) Wrap[T any](wrapper func(SelfFn[T]) SelfFn[T]) SelfFn[T] {
	return wrapper(fn)
}

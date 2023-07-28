package internalpipe

type AccumFn[T any] func(*T, *T) T

// Reduce applies the result of a function to each element one-by-one: f(p[n], f(p[n-1], f(p[n-2, ...]))).
func (p Pipe[T]) Reduce(fn AccumFn[T]) *T {
	data := p.Do()
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &data[0]
	default:
		res := data[0]
		for _, val := range data[1:] {
			res = fn(&res, &val)
		}
		return &res
	}
}

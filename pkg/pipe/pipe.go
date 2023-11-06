package pipe

import (
	"github.com/koss-null/funcfrog/internal/internalpipe"
)

// Pipe implements the pipe on any slice.
// Pipe may be initialized with `Slice`, `Func`, `Cycle` or `Range`.
type Pipe[T any] struct {
	internalpipe.Pipe[T]
}

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i]).
func (p *Pipe[T]) Map(fn func(T) T) Piper[T] {
	return &Pipe[T]{p.Pipe.Map(fn)}
}

// Filter leaves only items with true predicate fn.
func (p *Pipe[T]) Filter(fn Predicate[T]) Piper[T] {
	return &Pipe[T]{p.Pipe.Filter(fn)}
}

// MapFilter applies given function to each element of the underlying slice,
// if the second returning value of fn is false, the element is skipped (may be useful for error handling).
// returns the slice where each element is n[i] = f(p[i]) if it is not skipped.
func (p *Pipe[T]) MapFilter(fn func(T) (T, bool)) Piper[T] {
	return &Pipe[T]{p.Pipe.MapFilter(fn)}
}

// Sort sorts the underlying slice on a current step of a pipeline.
func (p *Pipe[T]) Sort(less Comparator[T]) Piper[T] {
	return &Pipe[T]{p.Pipe.Sort(less)}
}

// Reduce applies the result of a function to each element one-by-one: f(p[n], f(p[n-1], f(p[n-2, ...]))).
// It is recommended to use reducers from the default reducer if possible to decrease memory allocations.
func (p *Pipe[T]) Reduce(fn Accum[T]) *T {
	return p.Pipe.Reduce(internalpipe.AccumFn[T](fn))
}

// Sum returns the sum of all elements. It is similar to Reduce but is able to work in parallel.
func (p *Pipe[T]) Sum(plus Accum[T]) T {
	return p.Pipe.Sum(internalpipe.AccumFn[T](plus))
}

// First returns the first element of the pipe.
func (p *Pipe[T]) First() *T {
	return p.Pipe.First()
}

// Any returns a pointer to a random element in the pipe or nil if none left.
func (p *Pipe[T]) Any() *T {
	return p.Pipe.Any()
}

// Parallel set n - the amount of goroutines to run on.
// Only the first Parallel() in a pipe chain is applied.
func (p *Pipe[T]) Parallel(n uint16) Piper[T] {
	return &Pipe[T]{p.Pipe.Parallel(n)}
}

// Do evaluates all the pipeline and returns the result slice.
func (p *Pipe[T]) Do() []T {
	return p.Pipe.Do()
}

// Count evaluates all the pipeline and returns the amount of items.
func (p *Pipe[T]) Count() int {
	return p.Pipe.Count()
}

// Promices returns an array of Promice values - functions to be evaluated to get the value on i'th place.
// Promice returns two values: the evaluated value and if it is not skipped.
func (p *Pipe[T]) Promices() []func() (T, bool) {
	return p.Pipe.Promices()
}

// Erase wraps all pipe values to interface{} type, so you are able to use pipe methods with type convertions.
// You can use collectors from collectiors.go file of this package to collect results into a particular type.
func (p *Pipe[T]) Erase() Piper[any] {
	return &Pipe[any]{p.Pipe.Erase()}
}

// Snag links an error handler to the previous Pipe method.
func (p *Pipe[T]) Snag(h func(error)) Piper[T] {
	return &Pipe[T]{p.Pipe.Snag(internalpipe.ErrHandler(h))}
}

// Yeti links a yeti error handler to the Pipe.
func (p *Pipe[T]) Yeti(y internalpipe.YeetSnag) Piper[T] {
	return &Pipe[T]{p.Pipe.Yeti(y)}
}

// Entrails is an out-of-Piper interface method to provide Map[T1 -> T2].
func (p *Pipe[T]) Entrails() *internalpipe.Pipe[T] {
	return &p.Pipe
}

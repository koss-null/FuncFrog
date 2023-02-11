package pipe

import (
	"github.com/koss-null/lambda/internal/internalpipe"
)

// Pipe implements the pipe on any slice.
// Pipe may be initialized with `Slice`, `Func`, `Cycle` or `Range`.
type Pipe[T any] struct {
	internal internalpipe.Pipe[T]
}

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i]).
func (p *Pipe[T]) Map(fn func(T) T) Piper[T] {
	return &Pipe[T]{p.internal.Map(fn)}
}

// Filter leaves only items with true predicate fn.
func (p *Pipe[T]) Filter(fn Predicate[T]) Piper[T] {
	return &Pipe[T]{p.internal.Filter(fn)}
}

// Sort sorts the underlying slice on a current step of a pipeline.
func (p *Pipe[T]) Sort(less Comparator[T]) Piper[T] {
	return &Pipe[T]{p.internal.Sort(less)}
}

// Reduce applies the result of a function to each element one-by-one: f(p[n], f(p[n-1], f(p[n-2, ...]))).
// It is recommended to use reducers from the default reducer if possible to decrease memory allocations.
func (p *Pipe[T]) Reduce(fn Accum[T]) *T {
	return p.internal.Reduce(internalpipe.AccumFn[T](fn))
}

// Sum returns the sum of all elements. It is similar to Reduce but is able to work in parallel.
func (p *Pipe[T]) Sum(plus Accum[T]) T {
	return p.internal.Sum(internalpipe.AccumFn[T](plus))
}

// First returns the first element of the pipe.
func (p *Pipe[T]) First() *T {
	return p.internal.First()
}

// Any returns a pointer to a random element in the pipe or nil if none left.
func (p *Pipe[T]) Any() *T {
	return p.internal.Any()
}

// Parallel set n - the amount of goroutines to run on.
// Only the first Parallel() in a pipe chain is applied.
func (p *Pipe[T]) Parallel(n uint16) Piper[T] {
	return &Pipe[T]{p.internal.Parallel(n)}
}

// Do evaluates all the pipeline and returns the result slice.
func (p *Pipe[T]) Do() []T {
	return p.internal.Do()
}

// Count evaluates all the pipeline and returns the amount of items.
func (p *Pipe[T]) Count() int {
	return p.internal.Count()
}

// Entrails is an out of Piper interface method to provide Map[T1 -> T2].
func (p *Pipe[T]) Entrails() *internalpipe.Pipe[T] {
	return &p.internal
}

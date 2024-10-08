package pipe

import (
	"github.com/koss-null/funcfrog/internal/internalpipe"
)

// PipeNL implements the pipe on any slice.
// PipeNL may be initialized with `Slice`, `Func`, `Cycle` or `Range`.
type PipeNL[T any] struct {
	internalpipe.Pipe[T]
}

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i]).
func (p *PipeNL[T]) Map(fn func(T) T) PiperNoLen[T] {
	return &PipeNL[T]{p.Pipe.Map(fn)}
}

// Filter leaves only items with true predicate fn.
func (p *PipeNL[T]) Filter(fn Predicate[T]) PiperNoLen[T] {
	return &PipeNL[T]{p.Pipe.Filter(fn)}
}

// MapFilter applies given function to each element of the underlying slice,
// if the second returning value of fn is false, the element is skipped (may be useful for error handling).
// returns the slice where each element is n[i] = f(p[i]) if it is not skipped.
func (p *PipeNL[T]) MapFilter(fn func(T) (T, bool)) PiperNoLen[T] {
	return &PipeNL[T]{p.Pipe.MapFilter(fn)}
}

// First returns the first element of the pipe.
func (p *PipeNL[T]) First() *T {
	return p.Pipe.First()
}

// Any returns a pointer to a random element in the pipe or nil if none left.
func (p *PipeNL[T]) Any() *T {
	return p.Pipe.Any()
}

// Take is used to set the amount of values expected to be in result slice.
// It's applied only the first Gen() or Take() function in the pipe.
func (p *PipeNL[T]) Take(n int) Piper[T] {
	return &Pipe[T]{p.Pipe.Take(n)}
}

// Gen set the amount of values to generate as initial array.
// It's applied only the first Gen() or Take() function in the pipe.
func (p *PipeNL[T]) Gen(n int) Piper[T] {
	return &Pipe[T]{p.Pipe.Gen(n)}
}

// Parallel set n - the amount of goroutines to run on.
// Only the first Parallel() in a pipe chain is applied.
func (p *PipeNL[T]) Parallel(n uint16) PiperNoLen[T] {
	return &PipeNL[T]{p.Pipe.Parallel(n)}
}

// Erase wraps all pipe values to interface{} type, so you are able to use pipe methods with type convertions.
// You can use collectors from collectiors.go file of this package to collect results into a particular type.
func (p *PipeNL[T]) Erase() PiperNoLen[any] {
	return &PipeNL[any]{p.Pipe.Erase()}
}

// Snag links an error handler to the previous Pipe method.
func (p *PipeNL[T]) Snag(h func(error)) PiperNoLen[T] {
	return &PipeNL[T]{p.Pipe.Snag(internalpipe.ErrHandler(h))}
}

// Yeti links a yeti error handler to the Pipe.
func (p *PipeNL[T]) Yeti(y internalpipe.YeetSnag) PiperNoLen[T] {
	return &PipeNL[T]{p.Pipe.Yeti(y)}
}

// Entrails is an out of Piper interface method to provide Map[T1 -> T2].
func (p *PipeNL[T]) Entrails() *internalpipe.Pipe[T] {
	return &p.Pipe
}

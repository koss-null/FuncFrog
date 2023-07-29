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

// Entrails is an out of Piper interface method to provide Map[T1 -> T2].
func (p *PipeNL[T]) Entrails() *internalpipe.Pipe[T] {
	return &p.Pipe
}

package internalpipe

import (
	"math"
	"unsafe"

	"golang.org/x/exp/constraints"
)

const (
	panicLimitExceededMsg = "the limit have been exceeded, but the result is not calculated"
)

type GeneratorFn[T any] func(int) (*T, bool)

type Pipe[T any] struct {
	Fn            GeneratorFn[T]
	Len           int
	ValLim        int
	GoroutinesCnt int
	y             yeti
}

// Parallel set n - the amount of goroutines to run on.
// Only the first Parallel() in a pipe chain is applied.
func (p Pipe[T]) Parallel(n uint16) Pipe[T] {
	if p.GoroutinesCnt != defaultParallelWrks || n < 1 {
		return p
	}
	p.GoroutinesCnt = int(n)
	return p
}

// Take is used to set the amount of values expected to be in result slice.
// It's applied only the first Gen() or Take() function in the pipe.
func (p Pipe[T]) Take(n int) Pipe[T] {
	if p.limitSet() || p.lenSet() || n < 0 {
		return p
	}
	p.ValLim = n
	return p
}

// Gen set the amount of values to generate as initial array.
// It's applied only the first Gen() or Take() function in the pipe.
func (p Pipe[T]) Gen(n int) Pipe[T] {
	if p.limitSet() || p.lenSet() || n < 0 {
		return p
	}
	p.Len = n
	return p
}

// Count evaluates all the pipeline and returns the amount of items.
func (p Pipe[T]) Count() int {
	if p.limitSet() {
		return p.ValLim
	}
	_, cnt := p.do(false)
	return cnt
}

// Sang ads error handler to a current Pipe step.
func (p Pipe[T]) Snag(h ErrHandler) Pipe[T] {
	// FIXME: this pointer should be taken from p as the pointer to the previous Pipe step
	p.y.SnagPipe(unsafe.Pointer(&p), h)
	return p
}

// limit returns the upper border limit as the pipe evaluation limit.
func (p *Pipe[T]) limit() int {
	switch {
	case p.lenSet():
		return p.Len
	case p.limitSet():
		return p.ValLim
	default:
		return math.MaxInt - 1
	}
}

func (p *Pipe[T]) lenSet() bool {
	return p.Len != notSet
}

func (p *Pipe[T]) limitSet() bool {
	return p.ValLim != notSet
}

func min[T constraints.Ordered](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func max[T constraints.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func divUp(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}

func genTickets(n int) chan struct{} {
	tickets := make(chan struct{}, n)
	n = max(n, 1)
	for i := 0; i < n; i++ {
		tickets <- struct{}{}
	}
	return tickets
}

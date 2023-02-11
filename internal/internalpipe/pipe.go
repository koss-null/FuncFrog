package internalpipe

import (
	"math"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/constraints"

	"github.com/koss-null/lambda/internal/algo/parallel/qsort"
	"github.com/koss-null/lambda/internal/primitive/pointer"
)

const (
	firstCheckInterval = 345
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
}

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i]).
func (p Pipe[T]) Map(fn func(T) T) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			if obj, skipped := p.Fn(i); !skipped {
				res := fn(*obj)
				return &res, false
			}
			return nil, true
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,
	}
}

// Filter leaves only items with true predicate fn.
func (p Pipe[T]) Filter(fn func(T) bool) Pipe[T] {
	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			if obj, skipped := p.Fn(i); !skipped {
				if !fn(*obj) {
					return nil, true
				}
				return obj, false
			}
			return nil, true
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,
	}
}

// Sort sorts the underlying slice on a current step of a pipeline.
func (p Pipe[T]) Sort(less func(T, T) bool) Pipe[T] {
	var once sync.Once
	var sorted []T

	return Pipe[T]{
		Fn: func(i int) (*T, bool) {
			if sorted == nil {
				once.Do(func() {
					data := p.Do()
					if len(data) == 0 {
						return
					}
					sorted = qsort.Sort(data, less, p.GoroutinesCnt)
				})
			}
			if i >= len(sorted) {
				return nil, true
			}
			return &sorted[i], false
		},
		Len:           p.Len,
		ValLim:        p.ValLim,
		GoroutinesCnt: p.GoroutinesCnt,
	}
}

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
			res = pointer.From(fn(&res, &val))
		}
		return &res
	}
}

// Sum returns the sum of all elements. It is similar to Reduce but is able to work in parallel.
func (p Pipe[T]) Sum(plus AccumFn[T]) *T {
	return Sum(p.GoroutinesCnt, p.limit(), plus, p.Fn)
}

// First returns the first element of the pipe.
func (p Pipe[T]) First() *T {
	return First(p.limit(), p.GoroutinesCnt, p.Fn)
}

// Any returns a pointer to a random element in the pipe or nil if none left.
func (p Pipe[T]) Any() *T {
	return Any(p.lenSet(), p.limit(), p.GoroutinesCnt, p.Fn)
}

// Parallel set n - the amount of goroutines to run on.
// Only the first Parallel() in a pipe chain is applied.
func (p Pipe[T]) Parallel(n uint16) Pipe[T] {
	if n < 1 {
		return p
	}
	p.GoroutinesCnt = int(n)
	return p
}

// Take is used to set the amount of values expected to be in result slice.
// It's applied only the first Gen() or Take() function in the pipe.
func (p Pipe[T]) Take(n int) Pipe[T] {
	if n < 0 {
		return p
	}
	p.ValLim = n
	return p
}

// Gen set the amount of values to generate as initial array.
// It's applied only the first Gen() or Take() function in the pipe.
func (p Pipe[T]) Gen(n int) Pipe[T] {
	if n < 0 {
		return p
	}
	p.Len = n
	return p
}

// Do evaluates all the pipeline and returns the result slice.
func (p Pipe[T]) Do() []T {
	res, _ := p.do(true)
	return res
}

// Count evaluates all the pipeline and returns the amount of items.
func (p Pipe[T]) Count() int {
	if p.limitSet() {
		return p.ValLim
	}
	_, cnt := p.do(false)
	return cnt
}

// doToLimit executor for Take
func (p *Pipe[T]) doToLimit() []T {
	if p.ValLim == 0 {
		return []T{}
	}

	res := make([]T, 0, p.ValLim)
	for i := 0; len(res) < p.ValLim; i++ {
		obj, skipped := p.Fn(i)
		if !skipped {
			res = append(res, *obj)
		}

		if i == math.MaxInt {
			panic(panicLimitExceededMsg)
		}
	}
	return res
}

type ev[T any] struct {
	obj     *T
	skipped bool
}

// do is the main result evaluation pipeline
func (p *Pipe[T]) do(needResult bool) ([]T, int) {
	if p.limitSet() {
		res := p.doToLimit()
		return res, len(res)
	}

	var (
		eval    []ev[T]
		limit   = p.limit()
		step    = max(divUp(limit, p.GoroutinesCnt), 1)
		wg      sync.WaitGroup
		skipCnt atomic.Int64
	)
	if needResult && limit > 0 {
		eval = make([]ev[T], limit)
	}
	tickets := genTickets(p.GoroutinesCnt)
	for i := 0; i > -1 && i < limit; i += step {
		<-tickets
		wg.Add(1)
		go func(lf, rg int) {
			if rg < 0 {
				rg = limit
			}
			rg = min(rg, limit)
			var sCnt int64
			for j := lf; j < rg; j++ {
				obj, skipped := p.Fn(j)
				if skipped {
					sCnt++
				}
				if needResult {
					eval[j] = ev[T]{obj, skipped}
				}
			}
			skipCnt.Add(sCnt)
			tickets <- struct{}{}
			wg.Done()
		}(i, i+step)
	}
	wg.Wait()

	res := make([]T, 0, limit-int(skipCnt.Load()))
	for i := range eval {
		if !eval[i].skipped {
			res = append(res, *eval[i].obj)
		}
	}
	return res, limit - int(skipCnt.Load())
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
	return p.Len != -1
}

func (p *Pipe[T]) limitSet() bool {
	return p.ValLim != -1
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

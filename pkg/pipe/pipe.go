package pipe

import (
	"math"
	"sync"

	"github.com/koss-null/lambda/internal/algo/parallel/mergesort"
	"github.com/koss-null/lambda/internal/bitmap"

	"go.uber.org/atomic"
	"golang.org/x/exp/constraints"
)

const (
	defaultParallelWrks = 4
	maxJobsInTask       = 384
)

type ev[T any] struct {
	obj     *T
	skipped bool
}

// Pipe implements the pipe on any slice.
// Pipe should be initialized with New() or NewFn()
type Pipe[T any] struct {
	fn       func() func(int) (*T, bool)
	len      *int
	valLim   *int64
	skip     func(i int)
	parallel int
}

// Slice creates a Pipe from a slice
func Slice[T any](dt []T) *Pipe[T] {
	dtCp := make([]T, len(dt))
	copy(dtCp, dt)
	length := len(dt)
	varLim := int64(0)

	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				if i >= len(dtCp) {
					return nil, true
				}
				return &dtCp[i], false
			}
		},
		len:      &length,
		valLim:   &varLim,
		skip:     bitmap.NewNaive(len(dtCp)).SetTrue,
		parallel: defaultParallelWrks,
	}
}

// Func creates a lazy sequence d[i] = fn(i)
// fn is a function, that returns an object(T) and does it exist(bool)
// Initiating the pipe from a func you have to set either the
// output value amount using Get(n int) or
// the amount of generated values Gen(n int)
func Func[T any](fn func(i int) (T, bool)) *Pipe[T] {
	length := -1
	zero := int64(0)
	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				// FIXME: this code looks ugly
				obj, exist := fn(i)
				return &obj, !exist
			}
		},
		len:      &length,
		valLim:   &zero,
		skip:     bitmap.NewNaive(1024).SetTrue,
		parallel: defaultParallelWrks,
	}
}

// Map applies given function to each element of the underlying slice
// returns the slice where each element is n[i] = f(p[i])
func (p *Pipe[T]) Map(fn func(T) T) *Pipe[T] {
	return &Pipe[T]{
		fn: func() func(i int) (*T, bool) {
			return func(i int) (*T, bool) {
				if obj, skipped := p.fn()(i); !skipped {
					*obj = fn(*obj)
					return obj, false
				}
				return nil, true
			}
		},
		len:      p.len,
		valLim:   p.valLim,
		skip:     p.skip,
		parallel: p.parallel,
	}
}

// Filter leaves only items of an underlying slice where f(p[i]) is true
func (p *Pipe[T]) Filter(fn func(T) bool) *Pipe[T] {
	return &Pipe[T]{
		fn: func() func(i int) (*T, bool) {
			return func(i int) (*T, bool) {
				if obj, skipped := p.fn()(i); !skipped {
					if !fn(*obj) {
						p.skip(i)
						return nil, true
					}
					return obj, false
				}
				return nil, true
			}
		},
		len:      p.len,
		valLim:   p.valLim,
		skip:     p.skip,
		parallel: p.parallel,
	}
}

// Sort sorts the underlying slice on a current step of a pipeline
func (p *Pipe[T]) Sort(less func(T, T) bool) *Pipe[T] {
	var once sync.Once
	var sorted []T

	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				if sorted == nil {
					once.Do(func() {
						data := p.Do()
						if len(data) == 0 {
							return
						}
						sorted = mergesort.Sort(data, less, p.parallel)
					})
				}
				if i >= len(sorted) {
					return nil, true
				}
				return &sorted[i], false
			}
		},
		len:      p.len,
		valLim:   p.valLim,
		skip:     p.skip,
		parallel: p.parallel,
	}
}

// Reduce applies the result of a function to each element one-by-one: f(p[n], f(p[n-1], f(p[n-2, ...])))
func (p *Pipe[T]) Reduce(fn func(T, T) T) *T {
	data := p.Do()
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &data[0]
	default:
		res := data[0]
		for i := range data[1:] {
			res = fn(res, data[i+1])
		}
		return &res
	}
}

// Take is used to set the amount of values expected to be in result slice.
// It's applied only the first Gen() or Take() function in the pipe
func (p *Pipe[T]) Take(n int) *Pipe[T] {
	if n < 0 || *p.valLim != 0 || *p.len != -1 {
		return p
	}
	valLim := int64(n)
	p.valLim = &valLim
	return p
}

// Gen set the amount of values to generate as initial array.
// It's applied only the first Gen() or Take() function in the pipe
func (p *Pipe[T]) Gen(n int) *Pipe[T] {
	if n < 0 || *p.len != -1 || *p.valLim != 0 {
		return p
	}
	p.len = &n
	return p
}

// Parallel set n - the amount of goroutines to run on. The value by defalut is 4
// Only the first Parallel() call is not ignored
func (p *Pipe[T]) Parallel(n uint16) *Pipe[T] {
	if n < 1 {
		return p
	}

	p.parallel = int(n)
	return p
}

// Do evaluates all the pipeline and returns the result slice
func (p *Pipe[T]) Do() []T {
	res, _ := p.do(true)
	return res
}

// Count evaluates all the pipeline and returns the amount of left items
func (p *Pipe[T]) Count() int {
	if *p.valLim != 0 {
		return int(*p.valLim)
	}
	_, cnt := p.do(false)
	return cnt
}

// Sum returns the sum of all elements
func (p *Pipe[T]) Sum(sum func(T, T) T) *T {
	data := p.Do()
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &data[0]
	default:
		totalLen := int64(*p.len)
		if totalLen == -1 {
			totalLen = *p.valLim
		}
		if totalLen == 0 {
			return nil
		}
		var (
			step        = ceil(totalLen, p.parallel)
			totalLength = ceil(totalLen, step)
			totalRes    = make([]*T, totalLength)

			stepCnt int64
			wg      sync.WaitGroup
		)
		for lf := int64(0); lf < totalLen; lf += step {
			var rs int64
			for i := lf; i < min(lf+step, totalLen); i++ {
				rs += i
			}
			// totalRes = append(totalRes, zero)
			wg.Add(1)
			go func(data []T, stepCnt int64) {
				for i := 1; i < len(data); i++ {
					data[0] = sum(data[0], data[i])
				}
				totalRes[stepCnt] = &data[0]
				wg.Done()
			}(data[lf:min(lf+step, totalLen)], stepCnt)
			stepCnt++
		}
		wg.Wait()

		res := *totalRes[0]
		// no NPE since switch checks above
		for i := 1; i < len(totalRes); i++ {
			res = sum(res, *(totalRes[i]))
		}
		return &res
	}
}

// TODO: technically, First can work with Func pipe without Take() or Gen()
// First returns the first element of the pipe
func (p *Pipe[T]) First() *T {
	if *p.len == -1 && *p.valLim == 0 {
		return nil
	}

	// pipe init was done with a Func function
	if *p.len == -1 {
		// FIXME: may work in parallel
		pfn := p.fn()
		limit := math.MaxInt
		if *p.valLim < int64(limit) {
			limit = int(*p.valLim)
		}
		for i := 0; i < limit; i++ {
			obj, skipped := pfn(i)
			if !skipped {
				return obj
			}

			if i == math.MaxInt {
				return nil
			}
		}
		return nil
	}

	var (
		step = int(max(ceil(*p.len, p.parallel), 1))
		res  = make([]*T, ceil(*p.len, step))
		// dirty hack to be able to check zero step element was found fast
		res0    = make(chan *T, 1)
		pfn     = p.fn()
		wg      sync.WaitGroup
		stepCnt int
	)
	for i := 0; i < int(*p.len); i += step {
		wg.Add(1)
		go func(lf, rg, stepCnt int) {
			defer wg.Done()
			if rg > int(*p.len) {
				rg = int(*p.len)
			}
			for i := lf; i < rg; i++ {
				obj, skipped := pfn(i)
				if !skipped {
					if stepCnt == 0 {
						res0 <- obj
					}
					res[stepCnt] = obj
					return
				}
				if stepCnt != 0 && i%345 == 0 {
					// check if there is any result from the left
					if res[stepCnt-1] != nil {
						return
					}
				}
			}
		}(i, i+step, stepCnt)
		stepCnt++
	}

	result := make(chan *T)
	go func() {
		wg.Wait()
		for i := range res {
			if res[i] != nil {
				result <- res[i]
				return
			}
		}
		result <- nil
	}()

	select {
	case r := <-result:
		return r
	case r := <-res0:
		return r
	}
}

// doToLimit internal executor for Take
func (p *Pipe[T]) doToLimit() []T {
	pfn := p.fn()
	res := make([]T, 0, *p.valLim)
	for i := 0; int64(len(res)) < *p.valLim; i++ {
		obj, skipped := pfn(i)
		if !skipped {
			res = append(res, *obj)
		}

		if i == math.MaxInt {
			// TODO: should we panic here?
			break
		}
	}
	return res
}

// do is the main result evaluation pipeline
func (p *Pipe[T]) do(needResult bool) ([]T, int) {
	if *p.len == -1 && *p.valLim == 0 {
		return []T{}, 0
	}

	if *p.valLim != 0 {
		res := p.doToLimit()
		return res, len(res)
	}

	var (
		skipCnt atomic.Int64
		res     []T
		evals   []ev[T]
	)
	if needResult {
		res = make([]T, 0, *p.len)
		evals = make([]ev[T], *p.len)
	}

	step := int(math.Ceil(float64(*p.len) / float64(p.parallel)))
	pfn := p.fn()
	var wg sync.WaitGroup
	for i := 0; i < int(*p.len); i += step {
		wg.Add(1)
		go func(lf, rg int) {
			if rg > int(*p.len) {
				rg = int(*p.len)
			}
			for i := lf; i < rg; i++ {
				obj, skipped := pfn(i)
				if skipped {
					skipCnt.Add(1)
				}
				if needResult {
					evals[i] = ev[T]{obj, skipped}
				}
			}
			wg.Done()
		}(i, i+step)
	}
	wg.Wait()

	if needResult {
		for _, ev := range evals {
			if !ev.skipped {
				res = append(res, *ev.obj)
			}
		}
	}

	return res, *p.len - int(skipCnt.Load())
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

func ceil[T1, T2 int | int64](a T1, b T2) int64 {
	return int64(math.Ceil(float64(a) / float64(b)))
}

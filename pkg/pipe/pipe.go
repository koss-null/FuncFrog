package pipe

import (
	"math"
	"sync"

	"github.com/koss-null/lambda/internal/algo/parallel/mergesort"
	"github.com/koss-null/lambda/internal/primitive/pointer"

	"go.uber.org/atomic"
	"golang.org/x/exp/constraints"
)

const (
	defaultParallelWrks = 4
	firstCheckInterval  = 345
)

// Pipe implements the pipe on any slice.
// Pipe should be initialized with New() or NewFn()
type Pipe[T any] struct {
	fn       func() func(int) (*T, bool)
	len      *int
	valLim   *int
	parallel int
	prlSet   *bool
}

// Slice creates a Pipe from a slice
func Slice[T any](dt []T) *Pipe[T] {
	dtCp := make([]T, len(dt))
	copy(dtCp, dt)

	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				if i >= len(dtCp) {
					return nil, true
				}
				return &dtCp[i], false
			}
		},
		len:      pointer.To(len(dtCp)),
		valLim:   pointer.To(0),
		parallel: defaultParallelWrks,
		prlSet:   pointer.To(false),
	}
}

// Func creates a lazy sequence d[i] = fn(i)
// fn is a function, that returns an object(T) and does it exist(bool)
// Initiating the pipe from a func you have to set either the
// output value amount using Get(n int) or
// the amount of generated values Gen(n int)
func Func[T any](fn func(i int) (T, bool)) *Pipe[T] {
	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				obj, exist := fn(i)
				return &obj, !exist
			}
		},
		len:      pointer.To(-1),
		valLim:   pointer.To(0),
		parallel: defaultParallelWrks,
		prlSet:   pointer.To(false),
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
		parallel: p.parallel,
		prlSet:   p.prlSet,
	}
}

// Filter leaves only items of an underlying slice where f(p[i]) is true
func (p *Pipe[T]) Filter(fn func(T) bool) *Pipe[T] {
	return &Pipe[T]{
		fn: func() func(i int) (*T, bool) {
			return func(i int) (*T, bool) {
				if obj, skipped := p.fn()(i); !skipped {
					if !fn(*obj) {
						return nil, true
					}
					return obj, false
				}
				return nil, true
			}
		},
		len:      p.len,
		valLim:   p.valLim,
		parallel: p.parallel,
		prlSet:   p.prlSet,
	}
}

// Sort sorts the underlying slice on a current step of a pipeline
func (p *Pipe[T]) Sort(less func(T, T) bool) *Pipe[T] {
	var (
		once   sync.Once
		sorted []T
	)

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
		parallel: p.parallel,
		prlSet:   p.prlSet,
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
	p.valLim = &n
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
	*p.prlSet = true
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
		return *p.valLim
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
		if *p.valLim == 0 && *p.len == -1 {
			return nil
		}

		totalLen := *p.len
		if totalLen == -1 {
			totalLen = *p.valLim
		}

		var (
			step        = divUp(totalLen, p.parallel)
			totalResLen = divUp(totalLen, step)
			totalRes    = make([]*T, totalResLen)

			stepCnt int64
			wg      sync.WaitGroup
		)
		wg.Add(totalResLen)

		for lf := 0; lf < totalLen; lf += step {
			rs := 0
			for i := lf; i < min(lf+step, totalLen); i++ {
				rs += i
			}
			// totalRes = append(totalRes, zero)
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

// First returns the first element of the pipe
func (p *Pipe[T]) First() *T {
	if *p.len == -1 && *p.valLim == 0 {
		return nil
	}

	// pipe init was done with a Func function
	if *p.len == -1 {
		limit := math.MaxInt
		if *p.valLim < limit {
			limit = *p.valLim
		}

		pfn := p.fn()
		// FIXME: may work in parallel
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
		step = max(divUp(*p.len, p.parallel), 1)
		res  = make([]*T, divUp(*p.len, step))
		// it's a hack to be able to check zero step element was found fast
		res0 = make(chan *T, 1)
		pfn  = p.fn()

		wg      sync.WaitGroup
		stepCnt int
	)
	for i := 0; i < *p.len; i += step {
		wg.Add(1)
		go func(lf, rg, stepCnt int) {
			defer wg.Done()
			if rg > *p.len {
				rg = *p.len
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
				if stepCnt != 0 && i%firstCheckInterval == 0 {
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

func (p *Pipe[T]) Any() *T {
	limit := *p.len
	// pipe init was done with a Func function
	if *p.len == -1 {
		limit = math.MaxInt - 1
		if *p.valLim != 0 && *p.valLim < limit {
			limit = *p.valLim
		}
	}

	var (
		step    = max(divUp(limit, p.parallel), 1)
		tickets = make(chan struct{}, p.parallel)
		res     = make(chan *T, 1)
		pfn     = p.fn()

		wg   sync.WaitGroup
		done bool
	)
	// if p.len is not set, we need tickets to control the amount of goroutines
	for i := 0; i < max(p.parallel, 1); i++ {
		tickets <- struct{}{}
	}
	if !*p.prlSet && *p.len == -1 {
		step = 1 << 15
	}
	go func() {
		for i := 0; i < limit; i += step {
			wg.Add(1)
			<-tickets
			go func(lf, rg int) {
				defer func() {
					wg.Done()
					tickets <- struct{}{}
				}()
				if rg > limit {
					rg = limit
				}

				for i := lf; i < rg && !done; i++ {
					obj, skipped := pfn(i)
					if !skipped {
						if done {
							continue
						} // this check is just for fun here
						res <- obj // this one may jam
						done = true
						return
					}
				}
			}(i, i+step)
		}

		go func() {
			wg.Wait()
			res <- nil
			done = true
		}()
	}()

	return <-res
}

// doToLimit internal executor for Take
func (p *Pipe[T]) doToLimit() []T {
	pfn := p.fn()
	res := make([]T, 0, *p.valLim)
	for i := 0; len(res) < *p.valLim; i++ {
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

	var (
		step = divUp(*p.len, p.parallel)
		pfn  = p.fn()
		wg   sync.WaitGroup
	)
	for i := 0; i < *p.len; i += step {
		wg.Add(1)
		go func(lf, rg int) {
			if rg > *p.len {
				rg = *p.len
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

type ev[T any] struct {
	obj     *T
	skipped bool
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

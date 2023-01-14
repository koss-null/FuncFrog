package pipe

import (
	"math"
	"sync"

	"github.com/koss-null/lambda/internal/algo/parallel/qsort"
	"github.com/koss-null/lambda/internal/primitive/pointer"

	"go.uber.org/atomic"
	"golang.org/x/exp/constraints"
)

const (
	defaultParallelWrks = 1
	firstCheckInterval  = 345
)

const (
	panicLimitExceededMsg = "the limit is exceeded, but the result is not calculated"
)

type ev[T any] struct {
	obj     *T
	skipped bool
}

// Pipe implements the pipe on any slice.
// Pipe may be initialized with `Slice`, `Func`, `Cycle` or `Range`.
type Pipe[T any] struct {
	fn       func() func(int) (*T, bool)
	len      *int
	valLim   *int
	parallel int
	prlSet   *bool
}

// Slice creates a Pipe from a slice
func Slice[T any](dt []T) Pipe[T] {
	dtCp := make([]T, len(dt))
	copy(dtCp, dt)

	return Pipe[T]{
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

// Func creates a lazy sequence d[i] = fn(i).
// fn is a function, that returns an object(T) and does it exist(bool).
// Initiating the pipe from a func you have to set either the output value
// amount using Get(n int) or the amount of generated values Gen(n int).
func Func[T any](fn func(i int) (T, bool)) Pipe[T] {
	return Pipe[T]{
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
// returns the slice where each element is n[i] = f(p[i]).
func (p Pipe[T]) Map(fn func(T) T) Pipe[T] {
	return Pipe[T]{
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

// Filter leaves only items with true predicate fn.
func (p Pipe[T]) Filter(fn func(T) bool) Pipe[T] {
	return Pipe[T]{
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

// Sort sorts the underlying slice on a current step of a pipeline.
func (p Pipe[T]) Sort(less func(T, T) bool) Pipe[T] {
	var (
		once   sync.Once
		sorted []T
	)

	return Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				if sorted == nil {
					once.Do(func() {
						data := p.Do()
						if len(data) == 0 {
							return
						}
						sorted = qsort.Sort(data, less, p.parallel)
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

// Reduce applies the result of a function to each element one-by-one: f(p[n], f(p[n-1], f(p[n-2, ...]))).
func (p Pipe[T]) Reduce(fn func(T, T) T) *T {
	data := p.Do()
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &data[0]
	default:
		res := data[0]
		for _, val := range data[1:] {
			res = fn(res, val)
		}
		return &res
	}
}

// Sum returns the sum of all elements. It is similar to Reduce but is able to work in parallel.
func (p Pipe[T]) Sum(sum func(T, T) T) *T {
	data := p.Do()
	switch len(data) {
	case 0:
		return nil
	case 1:
		return &data[0]
	default:
		if !p.lenIsFinite() {
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

// First returns the first element of the pipe.
func (p Pipe[T]) First() *T {
	// FIXME: to be removed when the ussue with too big resStorage will be solved
	if !p.lenIsFinite() {
		return nil
	}
	var (
		limit   = p.limit()
		step    = max(divUp(limit, p.parallel), 1)
		tickets = genTickets(p.parallel)
		pfn     = p.fn()

		// this allocation may take a lot of memory on MaxInt length
		resStorage = make([]*T, divUp(limit, step))
		// it's a hack to be able to check zero step element was found fast
		res0 = make(chan *T, 1)
		// this res is calculated if there was no res found on the first step
		resNot0 = make(chan *T)

		wg      sync.WaitGroup
		stepCnt int
	)

	go func() {
		// i >= 0 is for an int owerflow case
		for i := 0; i >= 0 && i < limit; i += step {
			wg.Add(1)
			<-tickets
			// an owerflow is possible here since we rounded strep up
			next := i + step
			if next < 0 {
				next = math.MaxInt - 1
			}
			go func(lf, rg, stepCnt int) {
				defer func() {
					wg.Done()
					tickets <- struct{}{}
				}()

				rg = max(rg, limit)
				for i := lf; i < rg; i++ {
					obj, skipped := pfn(i)
					if !skipped {
						if stepCnt == 0 {
							res0 <- obj
						}
						resStorage[stepCnt] = obj
						return
					}

					if stepCnt != 0 && i%firstCheckInterval == 0 {
						// check if there is any result from the left
						if resStorage[stepCnt-1] != nil {
							return
						}
					}
				}
			}(i, next, stepCnt)
			stepCnt++
		}

		go func() {
			wg.Wait()
			for i := range resStorage {
				if resStorage[i] != nil {
					resNot0 <- resStorage[i]
					return
				}
			}
			resNot0 <- nil
		}()
	}()

	select {
	case r := <-res0:
		return r
	case r := <-resNot0:
		return r
	}
}

// Any returns a pointer to a random element in the pipe or nil if none left.
func (p Pipe[T]) Any() *T {
	var (
		res = make(chan *T, 1)
		// if p.len is not set, we need tickets to control the amount of goroutines
		tickets = genTickets(p.parallel)
		limit   = p.limit()
		step    = max(divUp(limit, p.parallel), 1)
		pfn     = p.fn()

		wg    sync.WaitGroup
		resMx sync.Mutex
		done  bool
	)
	if !*p.prlSet && !p.lenSet() {
		step = 1 << 15
	}

	setObj := func(obj *T) {
		resMx.Lock()
		defer resMx.Unlock()

		if done {
			return
		}
		res <- obj
		done = true
	}

	go func() {
		// i >= 0 is for an int owerflow case
		for i := 0; i >= 0 && i < limit; i += step {
			wg.Add(1)
			<-tickets
			go func(lf, rg int) {
				defer func() {
					wg.Done()
					tickets <- struct{}{}
				}()

				// accounting int owerflow case with max(rg, 0)
				rg = min(max(rg, 0), limit)
				for i := lf; i < rg && !done; i++ {
					obj, skipped := pfn(i)
					if !skipped {
						setObj(obj)
						return
					}
				}
			}(i, i+step)
		}

		go func() {
			wg.Wait()
			setObj(nil)
		}()
	}()

	return <-res
}

// Take is used to set the amount of values expected to be in result slice.
// It's applied only the first Gen() or Take() function in the pipe.
func (p Pipe[T]) Take(n int) Pipe[T] {
	if n < 0 || p.lenIsFinite() {
		return p
	}
	p.valLim = &n
	return p
}

// Gen set the amount of values to generate as initial array.
// It's applied only the first Gen() or Take() function in the pipe.
func (p Pipe[T]) Gen(n int) Pipe[T] {
	if n < 0 || p.lenIsFinite() {
		return p
	}
	p.len = &n
	return p
}

// Parallel set n - the amount of goroutines to run on.
// Only the first Parallel() in a pipe chain is applied.
func (p Pipe[T]) Parallel(n uint16) Pipe[T] {
	if n < 1 {
		return p
	}

	p.parallel = int(n)
	*p.prlSet = true
	return p
}

// Do evaluates all the pipeline and returns the result slice.
func (p Pipe[T]) Do() []T {
	res, _ := p.do(true)
	return res
}

// Count evaluates all the pipeline and returns the amount of items.
func (p Pipe[T]) Count() int {
	if *p.valLim != 0 {
		return *p.valLim
	}
	_, cnt := p.do(false)
	return cnt
}

// doToLimit internal executor for Take
func (p Pipe[T]) doToLimit() []T {
	pfn := p.fn()
	res := make([]T, 0, *p.valLim)
	for i := 0; len(res) < *p.valLim; i++ {
		obj, skipped := pfn(i)
		if !skipped {
			res = append(res, *obj)
		}

		if i == math.MaxInt {
			panic(panicLimitExceededMsg)
		}
	}
	return res
}

// do is the main result evaluation pipeline
func (p Pipe[T]) do(needResult bool) ([]T, int) {
	if !p.lenIsFinite() {
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

// limit returns the upper border limit as the pipe evaluation limit.
func (p Pipe[T]) limit() int {
	switch {
	case p.lenSet():
		return *p.len
	case p.limitSet():
		return *p.valLim
	default:
		return math.MaxInt - 1
	}
}

func (p Pipe[T]) lenSet() bool {
	return *p.len != -1
}

func (p Pipe[T]) limitSet() bool {
	return *p.valLim != 0
}

func (p Pipe[T]) lenIsFinite() bool {
	return p.lenSet() || p.limitSet()
}

// FIXME: move helpers to internal.
// helper functions.

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

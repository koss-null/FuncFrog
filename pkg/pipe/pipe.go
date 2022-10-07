package pipe

import (
	"math"
	"sort"
	"sync"

	"github.com/koss-null/lambda/internal/bitmap"
	"go.uber.org/atomic"
)

const (
	defaultParallelWrks      = 4
	maxParallelWrks          = 256
	maxJobsInTask            = 384
	singleThreadSortTreshold = 5000
)

// Pipe implements the pipe on any slice.
// Pipe should be initialized with New() or NewFn()
type Pipe[T any] struct {
	fn       func() func(int) (*T, bool)
	len      *int64
	valLim   *int64
	skip     func(i int)
	parallel int
}

// Slice creates a Pipe from a slice
func Slice[T any](dt []T) *Pipe[T] {
	dtCp := make([]T, len(dt))
	copy(dtCp, dt)
	length := int64(len(dt))
	zero := int64(0)
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
		valLim:   &zero,
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
	// FIXME: do we need fncache here?
	// fnCache := fnintcache.New[T]()
	length := int64(-1)
	zero := int64(0)
	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				// if res, found := fnCache.Get(i); found {
				// return res, false
				// }
				obj, exist := fn(i)
				// fnCache.Set(i, obj)
				return &obj, !exist
			}
		},
		len:      &length,
		valLim:   &zero,
		skip:     bitmap.NewNaive(1024).SetTrue,
		parallel: defaultParallelWrks,
	}
}

// Map applies fn to each element of the underlying slice
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

// Filter leaves only items of an underlying slice where fn(d[i]) is true
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

func merge[T any](a []T, lf, rg, lf1, rg1 int, less func(T, T) bool) {
	st := lf
	res := make([]T, 0, rg1-lf)
	for lf < rg && lf1 < rg1 {
		if less(a[lf], a[lf1]) {
			res = append(res, a[lf])
			lf++
			continue
		}
		res = append(res, a[lf1])
		lf1++
	}
	for lf < rg {
		res = append(res, a[lf])
		lf++
	}
	for lf1 < rg1 {
		res = append(res, a[lf1])
		lf1++
	}
	copy(a[st:], res)
}

func mergeSplits[T any](
	data []T,
	splits []struct{ lf, rg int },
	threads int,
	less func(T, T) bool,
) {
	jobTicket := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		jobTicket <- struct{}{}
	}
	newSplits := []struct{ lf, rg int }{}
	var wg sync.WaitGroup
	for i := 0; i < len(splits); i += 2 {
		if i+1 >= len(splits) {
			newSplits = append(newSplits, splits[i])
			break
		}

		<-jobTicket
		wg.Add(1)
		go func(i int) {
			merge(
				data,
				splits[i].lf, splits[i].rg,
				splits[i+1].lf, splits[i+1].rg,
				less,
			)
			jobTicket <- struct{}{}
			wg.Done()
		}(i)
		newSplits = append(
			newSplits,
			struct{ lf, rg int }{
				splits[i].lf,
				splits[i+1].rg,
			},
		)
	}
	wg.Wait()

	if len(newSplits) == 1 {
		return
	}
	mergeSplits(data, newSplits, threads, less)
}

func sortParallel[T any](data []T, less func(T, T) bool, threads int) []T {
	cmp := func(i, j int) bool {
		return less(data[i], data[j])
	}
	if len(data) < singleThreadSortTreshold {
		sort.Slice(data, cmp)
		return data
	}
	min := func(a, b int) int {
		if a > b {
			return b
		}
		return a
	}
	splits := []struct{ lf, rg int }{}
	step := (len(data) / threads) + 1
	lf, rg := 0, min(step, len(data))
	var wg sync.WaitGroup
	for lf < len(data) {
		wg.Add(1)
		go func(lf, rg int) {
			sort.Slice(data[lf:rg], cmp)
			wg.Done()
		}(lf, rg)
		splits = append(splits, struct {
			lf int
			rg int
		}{lf: lf, rg: rg})

		lf += step
		rg = min(rg+step, len(data))
	}
	wg.Wait()

	mergeSplits(data, splits, threads, less)
	return data
}

// Sort sorts the underlying slice
func (p *Pipe[T]) Sort(less func(T, T) bool) *Pipe[T] {
	var once sync.Once
	var sorted []T

	return &Pipe[T]{
		fn: func() func(int) (*T, bool) {
			return func(i int) (*T, bool) {
				once.Do(func() {
					data := p.Do()
					if len(data) == 0 {
						return
					}
					sorted = sortParallel(data, less, p.parallel)
				})
				if i >= len(sorted) {
					return nil, false
				}
				return &sorted[i], true
			}
		},
		len:      p.len,
		valLim:   p.valLim,
		skip:     p.skip,
		parallel: p.parallel,
	}
}

// Reduce applies the result of a function to each element one-by-one
func (p *Pipe[T]) Reduce(fn func(T, T) T) *T {
	data := p.Do()
	if len(data) == 0 {
		return nil
	}
	res := data[0]
	for i := range data[1:] {
		res = fn(res, data[i+1])
	}
	return &res
}

// Take set the amount of values expected to be in result slice
// Applied only the first Gen() or Take() function in the pipe
func (p *Pipe[T]) Take(n int) *Pipe[T] {
	if n < 0 || *p.valLim != 0 || *p.len != -1 {
		return p
	}
	valLim := int64(n)
	p.valLim = &valLim
	return p
}

// Gen set the amount of values to generate as initial array
// Applied only the first Gen() or Take() function in the pipe
func (p *Pipe[T]) Gen(n int) *Pipe[T] {
	if n < 0 || *p.len != -1 || *p.valLim != 0 {
		return p
	}
	length := int64(n)
	p.len = &length
	return p
}

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

type ev[T any] struct {
	obj     *T
	skipped bool
}

type task struct {
	jobs [maxJobsInTask]func()
	done func()
}

func worker(tasks <-chan task) func() {
	finish := false
	go func() {
		for !finish {
			if task, ok := <-tasks; ok {
				for _, job := range task.jobs {
					if job != nil {
						job()
					}
				}
				task.done()
			}
		}
	}()
	return func() { finish = false }
}

func intCircle(n int) func() int {
	var i int
	return func() int {
		if i == n {
			i = 0
		}
		i++
		return i - 1
	}
}

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

	return res, int(*p.len - skipCnt.Load())
}

// Parallel set n - the amount of goroutines to run on. The value by defalut is 4
// Only the first Parallel() call is not ignored
func (p *Pipe[T]) Parallel(n int) *Pipe[T] {
	if n < 1 {
		return p
	}
	if n > maxParallelWrks {
		n = maxParallelWrks
	}
	p.parallel = n
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

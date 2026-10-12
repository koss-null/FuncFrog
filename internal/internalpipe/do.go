package internalpipe

import (
	"math"
	"sync"
	"sync/atomic"
)

type ev[T any] struct {
	skipped bool
	obj     *T
}

// Do evaluates all the pipeline and returns the result slice.
func (p Pipe[T]) Do() []T {
	if p.limitSet() {
		res := p.doToLimit()
		return res
	}
	res, _ := p.do(true)
	return res
}

// doToLimit executor for Take
func (p *Pipe[T]) doToLimit() []T {
	if p.ValLim == 0 {
		return []T{}
	}

	if p.y != nil {
		defer p.y.Handle()
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

// do runs the result evaluation.
func (p *Pipe[T]) do(needResult bool) ([]T, int) {
	if p.y != nil {
		defer p.y.Handle()
	}

	limit := p.limit()

	// A single worker keeps the natural order, so the result can be produced
	// in one pass with no intermediate buffer.
	if p.GoroutinesCnt == 1 {
		if !needResult {
			sCnt := 0
			for i := 0; i < limit; i++ {
				if _, skipped := p.Fn(i); skipped {
					sCnt++
				}
			}
			return nil, limit - sCnt
		}

		res := make([]T, 0, limit)
		for i := 0; i < limit; i++ {
			if obj, skipped := p.Fn(i); !skipped {
				res = append(res, *obj)
			}
		}
		return res, len(res)
	}

	var (
		eval    []ev[T]
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
					eval[j] = ev[T]{
						obj:     obj,
						skipped: skipped,
					}
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

package internalpipe

import (
	"math"
	"sync"
	"sync/atomic"
)

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
	fn := p.Fn

	// A single worker keeps the natural order, so the result can be produced
	// in one pass with no intermediate buffer.
	if p.GoroutinesCnt == 1 {
		if !needResult {
			sCnt := 0
			for i := 0; i < limit; i++ {
				if _, skipped := fn(i); skipped {
					sCnt++
				}
			}
			return nil, limit - sCnt
		}

		res := make([]T, 0, limit)
		for i := 0; i < limit; i++ {
			if obj, skipped := fn(i); !skipped {
				res = append(res, *obj)
			}
		}
		return res, len(res)
	}

	// Parallel path. Every worker writes its results straight into the shared
	// result buffer at the element's own index and marks dropped positions in a
	// zeroed bitmap. This avoids keeping one alive *T per element (a Map boxes
	// every value), so the per-element allocations become collectable instead of
	// surviving until the gather. When nothing is skipped — the common Map case —
	// the buffer already is the answer and the serial gather pass disappears
	// entirely; otherwise the result is compacted in place.
	var (
		res     []T
		skipped []bool
		step    = max(divUp(limit, p.GoroutinesCnt), 1)
		wg      sync.WaitGroup
		skipCnt atomic.Int64
	)
	if needResult {
		res = make([]T, limit)
		skipped = make([]bool, limit)
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
				obj, sk := fn(j)
				if sk {
					sCnt++
					if needResult {
						skipped[j] = true
					}
					continue
				}
				if needResult {
					res[j] = *obj
				}
			}
			skipCnt.Add(sCnt)
			tickets <- struct{}{}
			wg.Done()
		}(i, i+step)
	}
	wg.Wait()

	skips := int(skipCnt.Load())
	if !needResult {
		return nil, limit - skips
	}
	if skips == 0 {
		return res, limit
	}

	w := 0
	for i := range res {
		if !skipped[i] {
			res[w] = res[i]
			w++
		}
	}
	return res[:w], w
}

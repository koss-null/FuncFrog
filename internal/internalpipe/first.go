package internalpipe

import (
	"context"
	"math"
	"sync"
)

func (p Pipe[T]) First() *T {
	limit := p.limit()
	if p.GoroutinesCnt == 1 {
		return firstSingleThread(limit, p.Fn)
	}
	return first(limit, p.GoroutinesCnt, p.Fn)
}

func firstSingleThread[T any](limit int, fn GeneratorFn[T]) *T {
	var obj T
	for i := 0; i < limit; i++ {
		if !fn(i, &obj) {
			return &obj
		}
	}
	return nil
}

type firstResult[T any] struct {
	val            *T
	step           int
	totalSteps     int
	zeroStepBorder int
	mx             *sync.Mutex
	ctx            context.Context
	cancel         func()
	done           map[int]struct{}
	sent           bool

	resForSure chan *T
}

func newFirstResult[T any](totalSteps int) *firstResult[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &firstResult[T]{
		step:       math.MaxInt,
		totalSteps: totalSteps,
		mx:         &sync.Mutex{},
		ctx:        ctx,
		cancel:     cancel,
		done:       make(map[int]struct{}, totalSteps),
		resForSure: make(chan *T),
	}
}

func (r *firstResult[T]) setVal(val *T, step int) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if step == r.zeroStepBorder {
		r.send(val)
		return
	}
	if step < r.step {
		r.val = val
		r.step = step
	}
}

// send delivers the result exactly once. It must be called with r.mx held.
func (r *firstResult[T]) send(val *T) {
	if r.sent {
		return
	}
	r.sent = true
	r.resForSure <- val
	r.cancel()
}

func (r *firstResult[T]) stepDone(step int) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.done[step] = struct{}{}

	// need to move r.zeroStepBorder up
	if step == r.zeroStepBorder {
		ok := true
		for ok {
			r.zeroStepBorder++
			if r.zeroStepBorder == r.step {
				r.send(r.val)
				return
			}

			_, ok = r.done[r.zeroStepBorder]
		}
	}

	if r.zeroStepBorder >= r.totalSteps {
		r.send(nil)
	}
}

func first[T any](limit, grtCnt int, fn func(i int, dst *T) bool) (f *T) {
	if limit == 0 {
		return
	}
	step := max(divUp(limit, grtCnt), 1)
	tickets := genTickets(grtCnt)

	totalSteps := limit / step
	if limit%step != 0 {
		totalSteps++
	}
	res := newFirstResult[T](totalSteps)

	stepCnt := 0
	for i := 0; i >= 0 && i < limit; i += step {
		<-tickets
		go func(lf, rg, stepCnt int) {
			defer func() {
				tickets <- struct{}{}
			}()

			if rg < 0 {
				rg = limit
			}
			rg = min(rg, limit)

			done := res.ctx.Done()
			var val T
			for j := lf; j < rg; j++ {
				// FIXME: this code is ugly but saves about 30% of time on locks
				if j%2 != 0 {
					if !fn(j, &val) {
						res.setVal(&val, stepCnt)
						return
					}
				} else {
					select {
					case <-done:
						return
					default:
						if !fn(j, &val) {
							res.setVal(&val, stepCnt)
							return
						}
					}
				}
			}
			res.stepDone(stepCnt)
		}(i, i+step, stepCnt)
		stepCnt++
	}

	return <-res.resForSure
}

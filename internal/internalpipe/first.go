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

func firstSingleThread[T any](limit int, fn func(i int) (*T, bool)) *T {
	var obj *T
	var skipped bool
	for i := 0; i < limit; i++ {
		obj, skipped = fn(i)
		if !skipped {
			return obj
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
		r.resForSure <- val
		r.cancel()
		return
	}
	if step < r.step {
		r.val = val
		r.step = step
	}
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
				r.resForSure <- r.val
				r.cancel()
				return
			}

			_, ok = r.done[r.zeroStepBorder]
		}
	}

	if r.zeroStepBorder >= r.totalSteps {
		r.resForSure <- nil
		r.cancel()
	}
}

func first[T any](limit, grtCnt int, fn func(i int) (*T, bool)) (f *T) {
	if limit == 0 {
		return
	}
	step := max(divUp(limit, grtCnt), 1)
	tickets := genTickets(grtCnt)

	res := newFirstResult[T](grtCnt)

	stepCnt := 0
	for i := 0; i >= 0 && i < limit; i += step {
		<-tickets
		go func(lf, rg, stepCnt int) {
			defer func() {
				tickets <- struct{}{}
			}()

			done := res.ctx.Done()
			for j := lf; j < rg; j++ {
				// FIXME: this code is ugly but saves about 30% of time on locks
				if j%2 != 0 {
					val, skipped := fn(j)
					if !skipped {
						res.setVal(val, stepCnt)
						return
					}
				} else {
					select {
					case <-done:
						return
					default:
						val, skipped := fn(j)
						if !skipped {
							res.setVal(val, stepCnt)
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

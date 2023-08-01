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

	resForSure chan *T
}

func (r *firstResult[T]) setVal(val *T, step int) {
	r.mx.Lock()
	if step == r.zeroStepBorder {
		r.resForSure <- val
		r.cancel()
	}
	if step < r.step {
		r.val = val
		r.step = step
	}
	r.mx.Unlock()
}

func (r *firstResult[T]) stepDone(step int) {
	r.mx.Lock()
	if step == r.zeroStepBorder {
		r.zeroStepBorder++
		switch r.zeroStepBorder {
		case r.step:
			r.resForSure <- r.val
			r.cancel()
		case r.totalSteps:
			r.resForSure <- nil
			r.cancel()
		}
	}
	r.mx.Unlock()
}

func newFirstResult[T any](totalSteps int) *firstResult[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &firstResult[T]{
		step:       math.MaxInt,
		totalSteps: totalSteps,
		mx:         &sync.Mutex{},
		resForSure: make(chan *T),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func first[T any](limit, grtCnt int, fn func(i int) (*T, bool)) *T {
	step := max(divUp(limit, grtCnt), 1)
	tickets := genTickets(grtCnt)

	res := newFirstResult[T](divUp(limit, step))

	stepCnt := 0
	for i := 0; i >= 0 && i < limit; i += step {
		<-tickets
		go func(lf, rg, stepCnt int) {
			defer func() {
				res.stepDone(stepCnt)
				tickets <- struct{}{}
			}()

			done := res.ctx.Done()
			for j := lf; j < rg; j++ {
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
		}(i, i+step, stepCnt)
		stepCnt++
	}

	return <-res.resForSure
}

// First returns the first element of the pipe.
// func (p Pipe[T]) First() *T {
// 	limit := p.limit()
// 	if p.GoroutinesCnt == 1 {
// 		return firstSingleThread(limit, p.Fn)
// 	}

// 	var (
// 		step    = max(divUp(limit, p.GoroutinesCnt), 1)
// 		tickets = genTickets(p.GoroutinesCnt)

// 		resContainer = struct {
// 			val *T
// 			pos int
// 		}{nil, math.MaxInt}
// 		res       = make(chan *T)
// 		resContMx sync.Mutex

// 		wg sync.WaitGroup

// 		stepCnt    int
// 		zeroStep   int
// 		zeroStepMx sync.Mutex
// 	)

// 	updStorage := func(val *T, pos int) {
// 		resContMx.Lock()
// 		if pos < resContainer.pos {
// 			resContainer.pos = pos
// 			resContainer.val = val
// 		}
// 		resContMx.Unlock()
// 	}
// 	resContainerPos := func() int {
// 		resContMx.Lock()
// 		defer resContMx.Unlock()
// 		return resContainer.pos
// 	}

// 	// this wg.Add is to make wg.Wait() wait if for loops that have not start yet
// 	wg.Add(1)
// 	go func() {
// 		var dn bool
// 		var dmx sync.Mutex
// 		isDone := func() bool {
// 			dmx.Lock()
// 			defer dmx.Unlock()
// 			return dn
// 		}
// 		done := func() {
// 			dmx.Lock()
// 			dn = true
// 			dmx.Unlock()
// 		}

// 		// i >= 0 is for an int owerflow case
// 		for i := 0; i >= 0 && i < limit && !isDone(); i += step {
// 			wg.Add(1)
// 			<-tickets
// 			go func(lf, rg, stepCnt int) {
// 				defer func() {
// 					tickets <- struct{}{}
// 					wg.Done()
// 				}()

// 				rg = max(rg, limit)
// 				for j := lf; j < rg; j++ {
// 					obj, skipped := p.Fn(j)
// 					if !skipped {
// 						if stepCnt == zeroStep {
// 							res <- obj
// 							done()
// 							return
// 						}
// 						updStorage(obj, stepCnt)
// 						return
// 					}

// 					// FIXME: mutex is taken in for loop here
// 					if resContainerPos() < stepCnt {
// 						done()
// 						return
// 					}
// 				}

// 				zeroStepMx.Lock()
// 				if stepCnt == zeroStep {
// 					zeroStep++
// 					resContMx.Lock()
// 					if resContainer.pos == zeroStep {
// 						res <- resContainer.val
// 						done()
// 						resContMx.Unlock()
// 						zeroStepMx.Unlock()
// 						return
// 					}
// 					resContMx.Unlock()
// 				}
// 				zeroStepMx.Unlock()
// 			}(i, i+step, stepCnt)
// 			stepCnt++
// 		}
// 		wg.Done()
// 	}()

// 	go func() {
// 		wg.Wait()
// 		resContMx.Lock()
// 		defer resContMx.Unlock()
// 		res <- resContainer.val
// 	}()

// 	return <-res
// }

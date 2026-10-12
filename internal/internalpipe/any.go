package internalpipe

import (
	"runtime"
	"sync"
)

// hugeLenStep is the base chunk size used when the pipe length is unknown.
const hugeLenStep = 1 << 15

// anyResCheckInterval controls how often a worker re-reads the shared result
// flag while scanning a chunk, so it does not take a lock on every element.
const anyResCheckInterval = 64

func anySingleThread[T any](limit int, fn GeneratorFn[T]) *T {
	var obj *T
	var skipped bool

	for i := 0; i < limit; i++ {
		if obj, skipped = fn(i); !skipped {
			return obj
		}
	}
	return nil
}

// Any returns a pointer to a random element in the pipe or nil if none left.
func (p Pipe[T]) Any() *T {
	limit := p.limit()
	if p.GoroutinesCnt == 1 {
		return anySingleThread(limit, p.Fn)
	}

	bounded := p.lenSet() || p.limitSet()
	step := hugeLenStep * max(1, runtime.GOMAXPROCS(0))
	if bounded {
		step = max(divUp(limit, p.GoroutinesCnt), 1)
	}

	var (
		resSet bool
		resCh  = make(chan *T, 1)
		done   = make(chan struct{})
		mx     sync.Mutex

		tickets = genTickets(p.GoroutinesCnt)
		wg      sync.WaitGroup
	)
	defer close(resCh)
	setObj := func(obj *T) {
		mx.Lock()
		if !resSet {
			resSet = true
			close(done)
			resCh <- obj
		}
		mx.Unlock()
	}
	getResSet := func() bool {
		mx.Lock()
		rs := resSet
		mx.Unlock()
		return rs
	}

	go func() {
		// i >= 0 is for an int owerflow case
	loop:
		for i := 0; i >= 0 && (!bounded || i < limit); i += step {
			select {
			case <-done:
				break loop
			default:
			}

			wg.Add(1)
			<-tickets

			go func(lf, rg int) {
				defer func() {
					tickets <- struct{}{}
					wg.Done()
				}()
				// int owerflow case
				rg = max(rg, 0)
				if bounded {
					rg = min(rg, limit)
				}

				if getResSet() {
					return
				}
				for j := lf; j < rg; j++ {
					if j != lf && (j-lf)%anyResCheckInterval == 0 && getResSet() {
						return
					}
					if obj, skipped := p.Fn(j); !skipped {
						setObj(obj)
						return
					}
				}
			}(i, i+step)
		}

		go func() {
			wg.Wait()
			setObj(nil)
			defer close(tickets)
		}()
	}()

	return <-resCh
}

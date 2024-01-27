package internalpipe

import "sync"

const hugeLenStep = 1 << 15

func anySingleThread[T any](limit int, fn GeneratorFn[T]) *T {
	var obj *T
	var skipped bool

	for i := 0; i > -1 && (i < limit); i++ {
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

	lenSet := p.lenSet()
	step := hugeLenStep
	if lenSet {
		step = max(divUp(limit, p.GoroutinesCnt), 1)
	}

	var (
		resSet bool
		resCh  = make(chan *T, 1)
		mx     sync.Mutex

		tickets = genTickets(p.GoroutinesCnt)
		wg      sync.WaitGroup
	)
	defer close(resCh)
	setObj := func(obj *T) {
		mx.Lock()
		if !resSet {
			resSet = true
			resCh <- obj
		}
		mx.Unlock()
	}

	go func() {
		// i >= 0 is for an int owerflow case
		for i := 0; i >= 0 && (!lenSet || i < limit); i += step {
			wg.Add(1)
			<-tickets

			go func(lf, rg int) {
				defer func() {
					tickets <- struct{}{}
					wg.Done()
				}()
				// int owerflow case
				rg = max(rg, 0)
				if lenSet {
					rg = min(rg, limit)
				}

				for j := lf; j < rg; j++ {
					// mx.Lock()
					rs := resSet
					// mx.Unlock()
					if !rs {
						obj, skipped := p.Fn(j)
						if !skipped {
							setObj(obj)
							return
						}
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

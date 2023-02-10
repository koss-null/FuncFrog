package internalpipe

import "sync"

const infiniteLenStep = 1 << 15

func anySingleThread[T any](lenSet bool, limit int, fn func(i int) (*T, bool)) *T {
	var obj *T
	var skipped bool
	if lenSet {
		for i := 0; i < limit; i++ {
			if obj, skipped = fn(i); !skipped {
				return obj
			}
		}
		return nil
	}
	i := 0
	for i >= 0 { // stopps on int overflow
		if obj, skipped = fn(i); !skipped {
			return obj
		}
		i++
	}
	return nil
}

func Any[T any](lenSet bool, limit int, grtCnt int, fn func(i int) (*T, bool)) *T {
	if grtCnt == 1 {
		return anySingleThread(lenSet, limit, fn)
	}

	var (
		res = make(chan *T, 1)
		// if p.len is not set, we need tickets to control the amount of goroutines
		tickets = genTickets(grtCnt)
		step    = max(divUp(limit, grtCnt), 1)

		wg    sync.WaitGroup
		resMx sync.Mutex
		done  bool
	)
	if !lenSet {
		step = infiniteLenStep
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
				for j := lf; j < rg && !done; j++ {
					obj, skipped := fn(j)
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

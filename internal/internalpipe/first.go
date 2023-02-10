package internalpipe

import (
	"math"
	"sync"
)

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

func First[T any](limit int, grtCnt int, fn func(i int) (*T, bool)) *T {
	if grtCnt == 1 {
		return firstSingleThread(limit, fn)
	}

	var (
		step    = max(divUp(limit, grtCnt), 1)
		tickets = genTickets(grtCnt)

		resStorage = struct {
			val *T
			pos int
		}{nil, math.MaxInt}
		resStorageMx sync.Mutex
		res          = make(chan *T, 1)

		wg sync.WaitGroup

		stepCnt  int
		zeroStep int
	)

	updStorage := func(val *T, pos int) {
		resStorageMx.Lock()
		if pos < resStorage.pos {
			resStorage.pos = pos
			resStorage.val = val
		}
		resStorageMx.Unlock()
	}

	// this wait is to make wg.Wait() wait if for loop have not start yet
	wg.Add(1)
	go func() {
		var done bool
		// i >= 0 is for an int owerflow case
		for i := 0; i >= 0 && i < limit && !done; i += step {
			wg.Add(1)
			<-tickets
			go func(lf, rg, stepCnt int) {
				defer func() {
					tickets <- struct{}{}
					wg.Done()
				}()

				rg = max(rg, limit)
				for j := lf; j < rg; j++ {
					obj, skipped := fn(j)
					if !skipped {
						if stepCnt == zeroStep {
							res <- obj
							done = true
							return
						}
						updStorage(obj, stepCnt)
						return
					}

					if resStorage.pos < stepCnt {
						done = true
						return
					}
				}
				// no lock needed since it's changed only in one goroutine
				if stepCnt == zeroStep {
					zeroStep++
					if resStorage.pos == zeroStep {
						res <- resStorage.val
						done = true
						return
					}
				}
			}(i, i+step, stepCnt)
			stepCnt++
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		resStorageMx.Lock()
		defer resStorageMx.Unlock()
		res <- resStorage.val
	}()

	return <-res
}

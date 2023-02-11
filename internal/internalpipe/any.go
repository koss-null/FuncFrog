package internalpipe

import "sync"

const infiniteLenStep = 1 << 15

func anySingleThread[T any](lenSet bool, limit int, fn GeneratorFn[T]) *T {
	var obj *T
	var skipped bool
	for i := 0; (!lenSet && i >= 0) || (i < limit); i++ {
		if obj, skipped = fn(i); !skipped {
			return obj
		}
	}
	return nil
}

func Any[T any](lenSet bool, limit int, grtCnt int, fn GeneratorFn[T]) *T {
	if grtCnt == 1 {
		return anySingleThread(lenSet, limit, fn)
	}

	step := infiniteLenStep
	if lenSet {
		step = max(divUp(limit, grtCnt), 1)
	}
	var (
		res = make(chan *T, 1)
		// if p.len is not set, we need tickets to control the amount of goroutines
		tickets = genTickets(grtCnt)

		wg   sync.WaitGroup
		done bool
	)
	if !lenSet {
		step = infiniteLenStep
	}

	setObj := func(obj *T) {
		if done {
			return
		}
		done = true
		res <- obj
	}

	go func() {
		// i >= 0 is for an int owerflow case
		for i := 0; i >= 0 && (!lenSet || i < limit); i += step {
			wg.Add(1)
			<-tickets
			go func(lf, rg int) {
				defer func() {
					wg.Done()
					tickets <- struct{}{}
				}()

				// accounting int owerflow case with max(rg, 0)
				rg = max(rg, 0)
				if lenSet {
					rg = min(rg, limit)
				}
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

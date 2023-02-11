package internalpipe

import "sync"

func sumSingleThread[T any](length int, plus AccumFn[T], fn GeneratorFn[T]) T {
	var res T
	var obj *T
	var skipped bool
	i := 0
	for ; i < length; i++ {
		obj, skipped = fn(i)
		if !skipped {
			res = *obj
			i++
			break
		}
	}

	for ; i < length; i++ {
		obj, skipped = fn(i)
		if !skipped {
			res = plus(&res, obj)
		}
	}
	return res
}

func Sum[T any](gortCnt int, length int, plus AccumFn[T], fn GeneratorFn[T]) T {
	if gortCnt == 1 {
		return sumSingleThread(length, plus, fn)
	}

	var (
		step = divUp(length, gortCnt)

		res   T
		resMx sync.Mutex
		wg    sync.WaitGroup
	)

	sum := func(x *T) {
		resMx.Lock()
		res = plus(&res, x)
		resMx.Unlock()
	}

	tickets := genTickets(gortCnt)
	for lf := 0; lf < length; lf += step {
		wg.Add(1)
		<-tickets
		go func(lf, rg int) {
			var inRes T
			var obj *T
			var skipped bool
			for i := lf; i < rg; i++ {
				if obj, skipped = fn(i); !skipped {
					inRes = plus(&inRes, obj)
				}
			}
			sum(&inRes)
			wg.Done()
			tickets <- struct{}{}
		}(lf, min(lf+step, length))
	}
	wg.Wait()

	return res
}

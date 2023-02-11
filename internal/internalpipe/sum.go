package internalpipe

import "sync"

func sumSingleThread[T any](length int, plus AccumFn[T], fn GeneratorFn[T]) *T {
	var zero T
	res := &zero

	var obj *T
	var skipped bool
	i := 0
	for ; i < length; i++ {
		obj, skipped = fn(i)
		if !skipped {
			res = obj
			i++
			break
		}
	}

	for ; i < length; i++ {
		obj, skipped = fn(i)
		if !skipped {
			res = plus(res, obj)
		}
	}
	return res
}

func Sum[T any](gortCnt int, length int, plus AccumFn[T], fn GeneratorFn[T]) *T {
	if gortCnt == 1 {
		return sumSingleThread(length, plus, fn)
	}

	var (
		step = divUp(length, gortCnt)

		zero  T
		resMx sync.Mutex
		wg    sync.WaitGroup
	)

	res := &zero
	sum := func(x *T) {
		resMx.Lock()
		res = plus(res, x)
		resMx.Unlock()
	}

	tickets := genTickets(gortCnt)
	for lf := 0; lf < length; lf += step {
		wg.Add(1)
		<-tickets
		var localRes T
		go func(lf, rg int, res *T) {
			var obj *T
			var skipped bool
			for i := lf; i < rg; i++ {
				if obj, skipped = fn(i); !skipped {
					res = plus(res, obj)
				}
			}
			sum(res)
			wg.Done()
			tickets <- struct{}{}
		}(lf, min(lf+step, length), &localRes)
	}
	wg.Wait()

	return res
}

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

// Sum returns the sum of all elements. It is similar to Reduce but is able to work in parallel.
func (p Pipe[T]) Sum(plus AccumFn[T]) T {
	length := p.limit()
	if p.GoroutinesCnt == 1 {
		return sumSingleThread(length, plus, p.Fn)
	}

	step := divUp(length, p.GoroutinesCnt)
	if step < 1 {
		step = 1
	}

	type partial struct {
		val T
		ok  bool
	}
	parts := make([]partial, divUp(length, step))

	var wg sync.WaitGroup
	tickets := genTickets(p.GoroutinesCnt)
	idx := 0
	for lf := 0; lf < length; lf += step {
		wg.Add(1)
		<-tickets
		i := idx
		idx++
		go func(i, lf, rg int) {
			var res T
			var ok bool
			var obj *T
			var skipped bool
			for j := lf; j < rg; j++ {
				if obj, skipped = p.Fn(j); !skipped {
					if ok {
						res = plus(&res, obj)
					} else {
						res = *obj
						ok = true
					}
				}
			}
			parts[i] = partial{val: res, ok: ok}
			tickets <- struct{}{}
			wg.Done()
		}(i, lf, min(lf+step, length))
	}
	wg.Wait()

	var res T
	var set bool
	for i := range parts {
		if !parts[i].ok {
			continue
		}
		if set {
			res = plus(&res, &parts[i].val)
		} else {
			res = parts[i].val
			set = true
		}
	}
	return res
}

package internalpipe

import "sync"

func sumSingleThread[T any](plus func(T, T) T, data []T) *T {
	res := data[0]
	for i := range data {
		res = plus(res, data[i])
	}
	return &res
}

func Sum[T any](plus func(T, T) T, data []T, gortCnt int) *T {
	switch len(data) {
	case 0:
		return nil
	case 1:
		d := data[0]
		return &d
	}

	if gortCnt == 1 {
		return sumSingleThread(plus, data)
	}

	var (
		lim         = len(data)
		step        = divUp(lim, gortCnt)
		totalPieces = divUp(lim, step)
		totalRes    = make([]*T, totalPieces)

		stepCnt int
		wg      sync.WaitGroup
	)

	wg.Add(totalPieces)
	tickets := genTickets(gortCnt)
	for lf := 0; lf < lim; lf += step {
		<-tickets
		go func(data []T, stepCnt int) {
			for i := 1; i < len(data); i++ {
				data[0] = plus(data[0], data[i])
			}
			totalRes[stepCnt] = &data[0]
			tickets <- struct{}{}
			wg.Done()
		}(data[lf:min(lf+step, lim)], stepCnt)
		stepCnt++
	}
	wg.Wait()

	res := *totalRes[0]
	// no NPE since switch checks above
	for i := 1; i < len(totalRes); i++ {
		res = plus(res, *(totalRes[i]))
	}
	return &res
}

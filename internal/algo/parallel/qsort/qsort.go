package qsort

import (
	"sort"
	"sync"
)

const (
	singleThreadSortTreshold = 5000
)

func Sort[T any](data []T, less func(*T, *T) bool, threads int) []T {
	if len(data) < 2 {
		return data
	}
	if threads < 1 {
		threads = 1
	}
	if threads == 1 {
		sort.Slice(data, func(i, j int) bool {
			return less(&data[i], &data[j])
		})
		return data
	}

	var wg sync.WaitGroup
	wg.Add(1)
	qsort(data, 0, len(data)-1, less, genTickets(threads), &wg)
	wg.Wait()
	return data
}

func qsort[T any](
	data []T,
	lf, rg int,
	less func(*T, *T) bool,
	tickets chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if lf >= rg {
		return
	}

	<-tickets
	defer func() { tickets <- struct{}{} }()

	if rg-lf < singleThreadSortTreshold {
		sort.Slice(data[lf:rg+1], func(i, j int) bool {
			return less(&data[lf+i], &data[lf+j])
		})

		return
	}

	q := partition(data, lf, rg, less)
	wg.Add(2)
	go qsort(data, lf, q, less, tickets, wg)
	go qsort(data, q+1, rg, less, tickets, wg)
}

func partition[T any](data []T, lf, rg int, less func(*T, *T) bool) int {
	// small arrays are not sorted with this method
	midIdx := (rg-lf)/2 + lf
	med := median([3]T{data[lf], data[midIdx], data[rg]}, less)

	for lf <= rg {
		for less(&data[lf], &med) {
			lf++
		}
		for less(&med, &data[rg]) {
			rg--
		}
		if lf >= rg {
			break
		}
		data[lf], data[rg] = data[rg], data[lf]
		lf++
		rg--
	}

	return rg
}

// median returns the median value of 3 elements.
func median[T any](elems [3]T, less func(*T, *T) bool) T {
	if less(&elems[1], &elems[0]) && less(&elems[0], &elems[2]) {
		return elems[0]
	}
	if less(&elems[2], &elems[0]) && less(&elems[0], &elems[1]) {
		return elems[0]
	}

	if less(&elems[0], &elems[1]) && less(&elems[1], &elems[2]) {
		return elems[1]
	}
	if less(&elems[2], &elems[1]) && less(&elems[1], &elems[0]) {
		return elems[1]
	}

	if less(&elems[0], &elems[2]) && less(&elems[2], &elems[1]) {
		return elems[2]
	}
	if less(&elems[1], &elems[2]) && less(&elems[2], &elems[0]) {
		return elems[2]
	}

	// two elements are equal
	if !less(&elems[0], &elems[1]) && !less(&elems[1], &elems[0]) {
		return elems[0]
	}
	if !less(&elems[1], &elems[2]) && !less(&elems[2], &elems[1]) {
		return elems[1]
	}
	return elems[2]
}

func genTickets(n int) chan struct{} {
	tickets := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		tickets <- struct{}{}
	}
	return tickets
}

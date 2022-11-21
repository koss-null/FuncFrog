package qsort

import "sort"

const (
	singleThreadSortTreshold = 5000
)

func Sort[T any](data []T, less func(T, T) bool, threads int) []T {
	if len(data) < singleThreadSortTreshold {
		sort.Slice(data, func(i, j int) bool {
			return less(data[i], data[j])
		})
		return data
	}

	if threads < 1 {
		threads = 1
	}
	qsort(data, less, genTickets(threads))
	return data
}

func qsort[T any](data []T, less func(T, T) bool, tickets chan struct{}) {
	<-tickets
	q := partition(data, less)
	tickets <- struct{}{}
	// FIXME: this one looks ugly
	if q > 1 {
		if q < singleThreadSortTreshold {
			sort.Slice(data[:q], func(i, j int) bool {
				return less(data[i], data[j])
			})
		} else {
			go qsort(data[:q], less, tickets)
		}
	}
	if q < len(data)-1 {
		if len(data)-q-1 < singleThreadSortTreshold {
			sort.Slice(data[q+1:], func(i, j int) bool {
				return less(data[i], data[j])
			})
		} else {
			go qsort(data[q+1:], less, tickets)
		}
	}
}

func partition[T any](data []T, less func(T, T) bool) int {
	midIdx := len(data) / 2
	mid := data[midIdx]
	lf, rg := 0, midIdx-1
	for lf <= rg {
		for less(data[lf], mid) {
			lf++
		}
		for less(mid, data[rg]) {
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

func genTickets(n int) chan struct{} {
	tickets := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		tickets <- struct{}{}
	}
	return tickets
}

package mergesort

import (
	"math"
	"sort"
	"sync"

	"golang.org/x/exp/constraints"
)

const (
	singleThreadSortTreshold = 5000
)

type border struct{ lf, rg int }

// Sort is an inner implementation of a parallel merge sort where sort.Slice()
// is used to sort sliced array parts
func Sort[T any](data []T, less func(T, T) bool, threads int) []T {
	if len(data) < singleThreadSortTreshold {
		sort.Slice(data, func(i, j int) bool {
			return less(data[i], data[j])
		})
		return data
	}

	step := max(int(math.Ceil(float64(len(data))/float64(threads))), 1)
	splits := make([]border, 0, len(data)/step+1)
	lf, rg := 0, min(step, len(data))
	var wg sync.WaitGroup
	for lf < len(data) {
		wg.Add(1)
		go func(lf, rg int) {
			d := data[lf:rg]
			cmp := func(i, j int) bool {
				return less(d[i], d[j])
			}
			sort.Slice(d, cmp)
			wg.Done()
		}(lf, rg)
		splits = append(splits, border{lf: lf, rg: rg})

		lf = rg
		rg = min(rg+step, len(data))
	}
	wg.Wait()

	mergeSplits(data, splits, threads, less)
	return data
}

// mergeSplits is an inner functnon to merge all sorted splits of a slice into a one sorted slice
func mergeSplits[T any](
	data []T,
	splits []border,
	threads int,
	less func(T, T) bool,
) {
	jobTicket := make(chan struct{}, threads)
	for i := 0; i < threads; i++ {
		jobTicket <- struct{}{}
	}

	var wg sync.WaitGroup
	// FIXME: old splits array should be reused probably since it causes huge mem alloc now
	// FIXME: n^2 mem alloc here
	newSplits := make([]border, 0, len(splits)/2+1)
	for i := 0; i < len(splits); i += 2 {
		// this is the last iteration
		if i+1 >= len(splits) {
			newSplits = append(newSplits, splits[i])
			break
		}

		// this one controls amount of simultanious merge processes running
		<-jobTicket
		wg.Add(1)
		go func(i int) {
			merge(
				data,
				splits[i].lf, splits[i].rg,
				splits[i+1].lf, splits[i+1].rg,
				less,
			)
			jobTicket <- struct{}{}
			wg.Done()
		}(i)
		newSplits = append(
			newSplits,
			struct{ lf, rg int }{
				splits[i].lf,
				splits[i+1].rg,
			},
		)
	}
	wg.Wait()

	if len(newSplits) == 1 {
		return
	}
	mergeSplits(data, newSplits, threads, less)
}

// merge is an inner function to merge two sorted slices into one sorted slice
func merge[T any](a []T, lf, rg, lf1, rg1 int, less func(T, T) bool) {
	st := lf
	res := make([]T, 0, rg1-lf)
	for lf < rg && lf1 < rg1 {
		if less(a[lf], a[lf1]) {
			res = append(res, a[lf])
			lf++
			continue
		}
		res = append(res, a[lf1])
		lf1++
	}
	// only one of the for[s] below is running
	for lf < rg {
		res = append(res, a[lf])
		lf++
	}
	for lf1 < rg1 {
		res = append(res, a[lf1])
		lf1++
	}
	copy(a[st:], res)
}

func min[T constraints.Ordered](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func max[T constraints.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}

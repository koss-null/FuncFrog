package qsort

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func rnd(n int) []int {
	res := make([]int, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		res[i] = r.Intn(n)
	}
	return res
}

func Test_partition(t *testing.T) {
	a := rnd(6000)
	q := partition(a, 0, len(a)-1, func(a, b *int) bool { return *a < *b })
	for i := 0; i <= q; i++ {
		for j := q + 1; j < len(a); j++ {
			require.LessOrEqual(t, a[i], a[j])
		}
	}
}

func Test_Sort(t *testing.T) {
	a := rnd(6000)
	Sort(a, func(a, b *int) bool { return *a < *b }, 3)
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_Sort_1thread(t *testing.T) {
	a := rnd(6000)
	Sort(a, func(a, b *int) bool { return *a < *b }, 1)
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_Sort_0thread(t *testing.T) {
	a := rnd(6000)
	Sort(a, func(a, b *int) bool { return *a < *b }, 0)
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_Sort_One(t *testing.T) {
	a := rnd(1)
	zeroIdxVal := a[0]
	Sort(a, func(a, b *int) bool { return *a < *b }, 3)
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
	require.Equal(t, len(a), 1)
	require.Equal(t, a[0], zeroIdxVal)
}

func Test_qsort(t *testing.T) {
	a := rnd(6000)
	tickets := genTickets(3)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b *int) bool { return *a < *b }, tickets, &wg)
	wg.Wait()
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_sort_one(t *testing.T) {
	a := rnd(1)
	require.Equal(t, len(a), 1)
	tickets := genTickets(3)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b *int) bool { return *a < *b }, tickets, &wg)
	wg.Wait()
	require.Equal(t, len(a), 1)
}

func Test_sort_big(t *testing.T) {
	a := rnd(100_000)
	tickets := genTickets(6)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b *int) bool { return *a < *b }, tickets, &wg)
	wg.Wait()
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_sort_huge(t *testing.T) {
	a := rnd(100_000_00)
	tickets := genTickets(12)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b *int) bool { return *a < *b }, tickets, &wg)
	wg.Wait()
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_sort_same(t *testing.T) {
	a := make([]int, 10000)
	for i := 0; i < 5000; i++ {
		a[i] = (i + 10) * (i - 1)
	} // others 5000 are 0
	tickets := genTickets(6)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b *int) bool { return *a < *b }, tickets, &wg)
	wg.Wait()
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

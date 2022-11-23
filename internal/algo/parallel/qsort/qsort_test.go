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
	q := partition(a, 0, len(a)-1, func(a, b int) bool { return a < b })
	for i := 0; i <= q; i++ {
		for j := q + 1; j < len(a); j++ {
			require.LessOrEqual(t, a[i], a[j])
		}
	}
}

func Test_sort(t *testing.T) {
	a := rnd(6000)
	tickets := genTickets(3)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b int) bool { return a < b }, tickets, &wg)
	wg.Wait()
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

func Test_sort_big(t *testing.T) {
	a := rnd(100_000)
	tickets := genTickets(6)
	var wg sync.WaitGroup
	wg.Add(1)
	qsort(a, 0, len(a)-1, func(a, b int) bool { return a < b }, tickets, &wg)
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
	qsort(a, 0, len(a)-1, func(a, b int) bool { return a < b }, tickets, &wg)
	wg.Wait()
	for i := range a {
		if i != 0 {
			require.GreaterOrEqual(t, a[i], a[i-1])
		}
	}
}

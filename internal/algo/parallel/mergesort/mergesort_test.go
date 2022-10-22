package mergesort

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func _less(a, b int) bool { return a < b }

func Test_merge(t *testing.T) {
	a := make([]int, 0, 200)
	lf, rg := 0, 0
	for i := 0; i < 200; i += 2 {
		rg++
		a = append(a, i)
	}
	lf1, rg1 := rg, 200
	for i := 1; i < 200; i += 2 {
		a = append(a, i)
	}

	merge(a, lf, rg, lf1, rg1, _less)
	prev := -1
	for _, item := range a {
		require.GreaterOrEqual(t, item, prev)
		prev = item
	}
}

func Test_mergeSplits(t *testing.T) {
	a := make([]int, 0, 200)
	lf, rg := 0, 0
	for i := 0; i < 200; i += 2 {
		rg++
		a = append(a, i)
	}
	lf1, rg1 := rg, 200
	for i := 1; i < 200; i += 2 {
		a = append(a, i)
	}

	mergeSplits(a, []border{{lf, rg}, {lf1, rg1}}, 3, _less)
	prev := -1
	for _, item := range a {
		require.GreaterOrEqual(t, item, prev)
		prev = item
	}
}

func Test_mergeSplits2(t *testing.T) {
	a := make([]int, 0, 300)
	lf, rg := 0, 100
	for i := 0; i < 300; i += 3 {
		a = append(a, i)
	}
	lf1, rg1 := 100, 200
	for i := 1; i < 300; i += 3 {
		a = append(a, i)
	}
	lf2, rg2 := 200, 300
	for i := 2; i < 300; i += 3 {
		a = append(a, i)
	}

	mergeSplits(a, []border{{lf, rg}, {lf1, rg1}, {lf2, rg2}}, 2, _less)
	prev := -1
	for _, item := range a {
		require.GreaterOrEqual(t, item, prev)
		prev = item
	}
}

func Test_Sort(t *testing.T) {
	a := make([]int, 0, 10_000)
	for i := 0; i < 10_000; i++ {
		a = append(a, 10_000-i)
	}
	res := Sort(a, _less, 12)
	t.Log(res)

	prev := -1
	for _, item := range res {
		require.GreaterOrEqual(t, item, prev)
		prev = item
	}
}

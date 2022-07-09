package sorts_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/koss-null/lambda-go/internal/sorts"

	"github.com/stretchr/testify/require"
)

const timeCircles = 50

func genArr(n int) []int {
	rand.Seed(int64(time.Now().UnixNano()))

	a := make([]int, n)
	for i := range a {
		a[i] = rand.Int()
	}
	return a
}

func Test_BubbleSort(t *testing.T) {
	a := genArr(1000)
	for n := 100; n < 10000; n *= 2 {
		arr := make([]int, n)

		var avgDur time.Duration
		for i := 0; i < timeCircles; i++ {
			copy(arr, a)
			start := time.Now()

			sorts.BubbleSort(arr, func(a, b int) bool { return a < b })

			avgDur += time.Now().Sub(start)
		}
		avgDur /= timeCircles
		fmt.Println("sample", n, "average time is", avgDur.Nanoseconds(), "(ns)")
	}

	fmt.Println("std sort: ------------")

	for n := 100; n < 10000; n *= 2 {
		arr := make([]int, n)

		var avgDur time.Duration
		for i := 0; i < timeCircles; i++ {
			copy(arr, a)
			start := time.Now()

			sort.Slice(arr, func(a, b int) bool { return arr[a] < arr[b] })

			avgDur += time.Now().Sub(start)
		}
		avgDur /= timeCircles
		fmt.Println("sample", n, "average time is", avgDur.Nanoseconds(), "(ns)")
	}

	require.True(t, false)
}

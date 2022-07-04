package stream_test

import (
	"testing"

	"github.com/koss-null/lambda-go/pkg/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genArray[T any](n uint, fn func(i uint) T) []T {
	a := make([]T, n)
	for i := uint(0); i < n; i++ {
		a[i] = fn(i)
	}
	return a
}

// var ar = genArray(1_000_000, func(i uint) int { return int(i) })

var ar = genArray(300, func(i uint) int { return int(i) })

func Test_Slice(t *testing.T) {
	arFromSlice := stream.S(ar).Slice()
	assert.Equal(t, len(ar), len(arFromSlice))
	for i := 0; i < len(ar); i++ {
		require.Equal(t, ar[i], arFromSlice[i])
	}
}

func Test_Skip(t *testing.T) {
	changed := stream.S(ar).Skip(200).Slice()
	assert.Equal(t, len(ar)-200, len(changed))
	arTr := ar[200:]
	for i := 0; i < len(changed); i++ {
		require.Equal(t, arTr[i], changed[i])
	}
}

func Test_Map(t *testing.T) {
	changed := stream.S(ar).
		Map(func(i int) int { return 2 * i }).
		Slice()
	for i := 0; i < len(changed); i++ {
		require.Equal(t, ar[i]*2, changed[i])
	}
}

func Test_Filter(t *testing.T) {
	changed := stream.S(ar).
		Filter(func(i int) bool { return i%2 == 0 }).
		Slice()
	for i := 0; i < len(changed); i += 2 {
		require.Equal(t, ar[i], changed[i/2])
	}
}

// func Test_Fun(t *testing.T) {
// 	cnt := uint(0)
// 	changed := stream.S(make([]uint, 1000)).
// 		Map(func(x uint) uint {
// 			defer func() { cnt++ }()
// 			return cnt * (cnt + 1)
// 		}). // generating a sequence this way
// 		Filter(func(x uint) bool { return x%2 == 0 }).
// 		Filter(func(x uint) bool { return x%3 == 0 }).
// 		Filter(func(x uint) bool { return x > 50 && x < 200 }).
// 		Slice()
// 	fmt.Println(changed)
// 	require.Equal(t, true, true)
// }

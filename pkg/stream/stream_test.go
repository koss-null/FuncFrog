package stream_test

import (
	"fmt"
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

var ar = genArray(1_000_000, func(i uint) int { return int(i) })

// var ar = genArray(300, func(i uint) int { return int(i) })

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
	fmt.Println(changed)
	for i := 0; i < len(changed); i++ {
		require.Equal(t, changed[i], ar[i]*2)
	}
	require.True(t, false)
}

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

func Test_Slice(t *testing.T) {
	ar := genArray(1_000_000, func(i uint) int { return int(i) })
	arFromSlice := stream.S(ar).Slice()
	assert.Equal(t, len(ar), len(arFromSlice))
	for i := 0; i < len(ar); i++ {
		require.Equal(t, ar[i], arFromSlice[i])
	}
}

func Test_Skip(t *testing.T) {
	ar := genArray(10, func(i uint) int { return int(i) })
	changed := stream.S(ar).Skip(2).Slice()
	assert.Equal(t, len(ar)-2, len(changed))
	arTr := ar[2:]
	for i := 0; i < len(changed); i++ {
		require.Equal(t, arTr[i], changed[i])
	}
}

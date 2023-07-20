package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func genSlice(n int) []int {
	res := make([]int, n)
	for i := 0; i < n; i++ {
		res[i] = i
	}
	return res
}

func TestFirst(t *testing.T) {
	s := genSlice(100_000_000)
	t.Run("single thread", func(t *testing.T) {
		f := First(len(s), 1, func(i int) (*int, bool) {
			return &s[i], i <= 90_000_000
		})
		require.Equal(t, *f, 90_000_001)
	})
	t.Run("5 threads", func(t *testing.T) {
		f := First(len(s), 5, func(i int) (*int, bool) {
			return &s[i], i <= 90_000_000
		})
		require.Equal(t, *f, 90_000_001)
	})
	t.Run("10 threads", func(t *testing.T) {
		f := First(len(s), 10, func(i int) (*int, bool) {
			return &s[i], i <= 90_000_000
		})
		require.Equal(t, *f, 90_000_001)
	})
	t.Run("1000 threads", func(t *testing.T) {
		f := First(len(s), 1000, func(i int) (*int, bool) {
			return &s[i], i <= 90_000_000
		})
		require.Equal(t, *f, 90_000_001)
	})
	t.Run("not found", func(t *testing.T) {
		f := First(len(s), 10, func(i int) (*int, bool) {
			return nil, true
		})
		require.Nil(t, f)
	})
}

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
	s := genSlice(1_000_000)
	t.Parallel()

	t.Run("single thread", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &s[i], i <= 900_000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		require.Equal(t, 900_001, *p.First())
	})
	t.Run("single thread 2", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &s[i], true
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		require.Nil(t, p.First())
	})
	t.Run("5 threads", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &s[i], i <= 900_000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 5,
		}
		require.Equal(t, 900_001, *p.First())
	})
	t.Run("1000 threads", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &s[i], i <= 900_000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 1000,
		}
		require.Equal(t, 900_001, *p.First())
	})
	t.Run("not found", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return nil, true
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 10,
		}
		require.Nil(t, p.First())
	})
	t.Run("not found len 0", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return nil, true
			},
			Len:           0,
			ValLim:        -1,
			GoroutinesCnt: 10,
		}
		require.Nil(t, p.First())
	})
}

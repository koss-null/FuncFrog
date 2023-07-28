package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Reduce(t *testing.T) {
	t.Parallel()

	t.Run("single thread lim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Equal(t, 4999950000, *res)
	})

	t.Run("seven thread lim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 7,
		}
		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Equal(t, 4999950000, *res)
	})

	t.Run("single thread ValLim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        100_000,
			GoroutinesCnt: 1,
		}
		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Equal(t, 4999950000, *res)
	})

	t.Run("seven thread ValLim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        100_000,
			GoroutinesCnt: 7,
		}

		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Equal(t, 4999950000, *res)
	})

	t.Run("single thread single value", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				i++
				return &i, false
			},
			Len:           -1,
			ValLim:        1,
			GoroutinesCnt: 1,
		}

		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Equal(t, 1, *res)
	})

	t.Run("single thread zero value", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        0,
			GoroutinesCnt: 1,
		}

		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Nil(t, res)
	})

	t.Run("seven thread single value", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				i++
				return &i, false
			},
			Len:           -1,
			ValLim:        1,
			GoroutinesCnt: 7,
		}

		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Equal(t, 1, *res)
	})

	t.Run("seven thread zero value", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        0,
			GoroutinesCnt: 7,
		}

		res := p.Reduce(func(x, y *int) int { return *x + *y })
		require.Nil(t, res)
	})
}

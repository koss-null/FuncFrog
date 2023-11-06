package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MapFilter(t *testing.T) {
	t.Parallel()

	exp := make([]int, 0, 100_000)
	for i := 0; i < 100_000; i++ {
		if i%2 == 0 {
			exp = append(exp, i+1)
		}
	}

	t.Run("single thread lim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		res := p.MapFilter(func(x int) (int, bool) { return x + 1, x%2 == 0 }).
			Do()

		require.Equal(t, len(exp), len(res))
		for i, r := range res {
			require.Equal(t, exp[i], r)
		}
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

		res := p.MapFilter(func(x int) (int, bool) { return x + 1, x%2 == 0 }).
			Parallel(7).Do()

		for i, r := range res {
			require.Equal(t, exp[i], r)
		}
	})

	t.Run("single thread ValLim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        len(exp),
			GoroutinesCnt: 1,
		}
		res := p.MapFilter(func(x int) (int, bool) { return x + 1, x%2 == 0 }).
			Do()

		require.Equal(t, len(exp), len(res))
		for i, r := range res {
			require.Equal(t, exp[i], r)
		}
	})

	t.Run("seven thread ValLim set", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        len(exp),
			GoroutinesCnt: 7,
		}
		res := p.MapFilter(func(x int) (int, bool) { return x + 1, x%2 == 0 }).
			Do()

		require.Equal(t, len(exp), len(res))
		for i, r := range res {
			require.Equal(t, exp[i], r)
		}
	})

	t.Run("seven thread ValLim multiple calls", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           -1,
			ValLim:        len(exp),
			GoroutinesCnt: 7,
		}
		res := p.MapFilter(func(x int) (int, bool) { return x + 1, x%2 == 0 }).
			MapFilter(func(x int) (int, bool) { return x, true }).
			Do()

		require.Equal(t, len(exp), len(res))
		for i, r := range res {
			require.Equal(t, exp[i], r)
		}
	})

	t.Run("single thread lim set empty", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		res := p.MapFilter(func(x int) (int, bool) { return x + 1, false }).Do()

		require.Equal(t, 0, len(res))
	})

	t.Run("seven thread lim set empty", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 7,
		}

		res := p.MapFilter(func(x int) (int, bool) { return x + 1, false }).
			Parallel(7).Do()

		require.Equal(t, 0, len(res))
	})
}

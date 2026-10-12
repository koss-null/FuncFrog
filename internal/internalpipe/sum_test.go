package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Sum(t *testing.T) {
	initA100k()
	t.Parallel()

	t.Run("Single thread happy 100k sum", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = a100k[i]
				return false
			},
			Len:           len(a100k),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x + *y
		})

		require.NotNil(t, s)
		require.Equal(t, 4999950000.0, s)
	})

	t.Run("Quadro thread happy 100k sum", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = a100k[i]
				return false
			},
			Len:           len(a100k),
			ValLim:        -1,
			GoroutinesCnt: 4,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x + *y
		})

		require.NotNil(t, s)
		require.Equal(t, 4999950000.0, s)
	})

	t.Run("Single thread empty", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = 1.
				return false
			},
			Len:           0,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x + *y
		})

		require.NotNil(t, s)
		require.Equal(t, 0.0, s)
	})

	t.Run("Quadro thread empty", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = 1.
				return false
			},
			Len:           0,
			ValLim:        -1,
			GoroutinesCnt: 4,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x + *y
		})

		require.NotNil(t, s)
		require.Equal(t, 0.0, s)
	})

	t.Run("Single thread 1 elem", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = 100500.
				return i != 0
			},
			Len:           1,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x + *y
		})

		require.NotNil(t, s)
		require.Equal(t, 100500., s)
	})

	t.Run("Quadro thread 1 elem", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = 100500.
				return i != 0
			},
			Len:           1,
			ValLim:        -1,
			GoroutinesCnt: 4,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x + *y
		})

		require.NotNil(t, s)
		require.Equal(t, 100500., s)
	})

	t.Run("Single thread multiply", func(t *testing.T) {
		vals := []float64{1, 2, 3, 4}
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = vals[i]
				return false
			},
			Len:           len(vals),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x * *y
		})

		require.Equal(t, 24.0, s)
	})

	t.Run("Quadro thread multiply", func(t *testing.T) {
		vals := []float64{1, 2, 3, 4}
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = vals[i]
				return false
			},
			Len:           len(vals),
			ValLim:        -1,
			GoroutinesCnt: 4,
		}
		s := p.Sum(func(x, y *float64) float64 {
			return *x * *y
		})

		require.Equal(t, 24.0, s)
	})

	t.Run("Single thread min", func(t *testing.T) {
		vals := []float64{4, 3, 7, 1, 9}
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = vals[i]
				return false
			},
			Len:           len(vals),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		s := p.Sum(func(x, y *float64) float64 {
			if *x < *y {
				return *x
			}
			return *y
		})

		require.Equal(t, 1.0, s)
	})

	t.Run("Quadro thread min", func(t *testing.T) {
		vals := []float64{4, 3, 7, 1, 9}
		p := Pipe[float64]{
			Fn: func(i int, dst *float64) bool {
				*dst = vals[i]
				return false
			},
			Len:           len(vals),
			ValLim:        -1,
			GoroutinesCnt: 4,
		}
		s := p.Sum(func(x, y *float64) float64 {
			if *x < *y {
				return *x
			}
			return *y
		})

		require.Equal(t, 1.0, s)
	})
}

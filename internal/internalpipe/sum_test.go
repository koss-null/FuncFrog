package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/internal/primitive/pointer"
)

func Test_Sum(t *testing.T) {
	initA100k()
	t.Parallel()

	t.Run("Single thread happy 100k sum", func(t *testing.T) {
		p := Pipe[float64]{
			Fn: func(i int) (*float64, bool) {
				return &a100k[i], false
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
			Fn: func(i int) (*float64, bool) {
				return &a100k[i], false
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
			Fn: func(i int) (*float64, bool) {
				return pointer.To(1.), false
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
			Fn: func(i int) (*float64, bool) {
				return pointer.To(1.), false
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
			Fn: func(i int) (*float64, bool) {
				return pointer.To(100500.), i != 0
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
			Fn: func(i int) (*float64, bool) {
				return pointer.To(100500.), i != 0
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
}

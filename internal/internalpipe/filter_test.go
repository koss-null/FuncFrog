package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/funcfrog/internal/primitive/pointer"
)

func Test_Filter(t *testing.T) {
	t.Parallel()

	t.Run("single thread even numbers filter", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}

		res := p.Filter(func(x *int) bool {
			return pointer.From(x)%2 == 0
		}).Do()
		j := 0
		for i := 0; i < 100_000; i += 2 {
			require.Equal(t, i, res[j])
			j++
		}
	})
	t.Run("seven thread even numbers filter", func(t *testing.T) {
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return &i, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 7,
		}

		res := p.Filter(func(x *int) bool {
			return pointer.From(x)%2 == 0
		}).Do()
		j := 0
		for i := 0; i < 100_000; i += 2 {
			require.Equal(t, i, res[j])
			j++
		}
	})
	t.Run("single thread even numbers empty res filter", func(t *testing.T) {
		pts := pointer.To(7)
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return pts, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 1,
		}

		res := p.Filter(func(x *int) bool {
			return pointer.From(x)%2 == 0
		}).Do()
		require.Equal(t, 0, len(res))
	})
	t.Run("seven thread even numbers empty res filter", func(t *testing.T) {
		pts := pointer.To(7)
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return pts, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 7,
		}

		res := p.Filter(func(x *int) bool {
			return pointer.From(x)%2 == 0
		}).Do()
		require.Equal(t, 0, len(res))
	})
	t.Run("seven thread even numbers empty res double filter", func(t *testing.T) {
		pts := pointer.To(7)
		p := Pipe[int]{
			Fn: func(i int) (*int, bool) {
				return pts, false
			},
			Len:           100_000,
			ValLim:        -1,
			GoroutinesCnt: 7,
		}

		res := p.Filter(func(x *int) bool {
			return pointer.From(x)%2 == 0
		}).Filter(func(x *int) bool {
			return pointer.From(x)%2 == 0
		}).Do()
		require.Equal(t, 0, len(res))
	})
}

package internalpipe

import (
	"testing"
	"time"

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
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				*dst = s[i]
				return i <= 900_000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		require.Equal(t, 900_001, *p.First())
	})

	t.Run("single thread 2", func(t *testing.T) {
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				*dst = s[i]
				return true
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 1,
		}
		require.Nil(t, p.First())
	})

	t.Run("5 threads", func(t *testing.T) {
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				*dst = s[i]
				return i <= 900_000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 5,
		}
		require.NotNil(t, p.First())
		require.Equal(t, 900_001, *p.First())
	})

	t.Run("2 threads first half", func(t *testing.T) {
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				*dst = s[i]
				return s[i] < 1000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 2,
		}
		require.NotNil(t, p.First())
		require.Equal(t, 1000, *p.First())
	})

	t.Run("1000 threads", func(t *testing.T) {
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				*dst = s[i]
				return i <= 900_000
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 1000,
		}
		require.NotNil(t, p.First())
		require.Equal(t, 900_001, *p.First())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				return true
			},
			Len:           len(s),
			ValLim:        -1,
			GoroutinesCnt: 10,
		}
		require.Nil(t, p.First())
	})

	t.Run("not found len 0", func(t *testing.T) {
		t.Parallel()

		p := Pipe[int]{
			Fn: func(i int, dst *int) bool {
				return true
			},
			Len:           0,
			ValLim:        0,
			GoroutinesCnt: 10,
		}
		require.Nil(t, p.First())
	})
}

// firstWithin fails instead of hanging if First never delivers a result.
func firstWithin(t *testing.T, p Pipe[int]) *int {
	t.Helper()

	res := make(chan *int, 1)
	go func() {
		res <- p.First()
	}()

	select {
	case v := <-res:
		return v
	case <-time.After(5 * time.Second):
		t.Fatalf("First() deadlocked: len=%d goroutines=%d", p.Len, p.GoroutinesCnt)
		return nil
	}
}

func TestFirst_UnevenChunks(t *testing.T) {
	t.Parallel()

	t.Run("empty not divisible", func(t *testing.T) {
		t.Parallel()

		for _, c := range []struct{ length, goroutines int }{{10, 7}, {5, 4}, {7, 3}, {13, 5}} {
			c := c
			p := Pipe[int]{
				Fn: func(i int, dst *int) bool {
					return true
				},
				Len:           c.length,
				ValLim:        -1,
				GoroutinesCnt: c.goroutines,
			}
			require.Nil(t, firstWithin(t, p))
		}
	})

	t.Run("finds first in later chunk", func(t *testing.T) {
		t.Parallel()

		s := genSlice(1_000_000)
		for _, target := range []int{0, 1, 6, 9, 500_000, 999_999} {
			target := target
			p := Pipe[int]{
				Fn: func(i int, dst *int) bool {
					*dst = s[i]
					return i != target
				},
				Len:           len(s),
				ValLim:        -1,
				GoroutinesCnt: 7,
			}
			require.Equal(t, target, *firstWithin(t, p))
		}
	})

	t.Run("matches single thread", func(t *testing.T) {
		t.Parallel()

		s := genSlice(1_000_000)
		const border = 100_000
		fn := func(i int, dst *int) bool {
			*dst = s[i]
			return s[i] < border
		}
		want := *Pipe[int]{Fn: fn, Len: len(s), ValLim: -1, GoroutinesCnt: 1}.First()
		got := firstWithin(t, Pipe[int]{Fn: fn, Len: len(s), ValLim: -1, GoroutinesCnt: 7})
		require.Equal(t, want, *got)
	})
}

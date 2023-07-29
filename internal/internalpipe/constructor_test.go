package internalpipe

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Slice(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p := Slice(a)
	require.Equal(t, p.Len, 10)
	require.Equal(t, p.ValLim, notSet)
	require.Equal(t, p.GoroutinesCnt, defaultParallelWrks)
	r, _ := p.Fn(5)
	require.Equal(t, *r, a[5])
	_, skipped := p.Fn(15)
	require.True(t, skipped)
}

func Test_FuncP(t *testing.T) {
	p := FuncP(func(i int) (*int, bool) {
		return &i, true
	})
	require.Equal(t, p.Len, notSet)
	require.Equal(t, p.ValLim, notSet)
	require.Equal(t, p.GoroutinesCnt, defaultParallelWrks)
}

func Not[T any](fn func(x T) bool) func(T) bool {
	return func(x T) bool {
		return !fn(x)
	}
}

func Test_Cycle(t *testing.T) {
	t.Parallel()

	isVowel := func(s *string) bool {
		st := struct{}{}
		_, ok := map[string]struct{}{"a": st, "e": st, "i": st, "o": st, "u": st, "y": st}[*s]
		return ok
	}

	t.Run("happy", func(t *testing.T) {
		p := Cycle([]string{"a", "b", "c", "d"})
		require.Equal(
			t,
			[]string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			p.Filter(isVowel).Map(strings.ToUpper).Take(10).Do(),
		)
	})

	t.Run("empty res", func(t *testing.T) {
		p := Cycle([]string{"a", "b", "c", "d"})
		require.Equal(
			t,
			[]string{},
			p.Filter(isVowel).Filter(Not(isVowel)).Gen(10).Do(),
		)
	})

	t.Run("empty cycle", func(t *testing.T) {
		p := Cycle([]string{})
		require.Equal(
			t,
			[]string{},
			p.Filter(isVowel).
				Map(strings.TrimSpace).
				Gen(10).
				Do(),
		)
	})
}

func Test_Range(t *testing.T) {
	t.Parallel()

	t.Run("happy", func(t *testing.T) {
		p := Range(0, 10, 1)
		require.Equal(t, p.Len, 10)
		require.Equal(t, p.ValLim, notSet)
		require.Equal(t, p.GoroutinesCnt, defaultParallelWrks)
		res := p.Do()
		for i := 0; i < 10; i++ {
			require.Equal(t, i, res[i])
		}
	})
	t.Run("single step owerflow", func(t *testing.T) {
		p := Range(1, 10, 100)
		require.Equal(t, 1, p.Len)
		require.Equal(t, notSet, p.ValLim)
		require.Equal(t, defaultParallelWrks, p.GoroutinesCnt)
		res := p.Do()
		require.Equal(t, 1, res[0])
	})
	t.Run("finish less than start", func(t *testing.T) {
		p := Range(100, 10, 100)
		require.Equal(t, 0, p.Len)
		require.Equal(t, notSet, p.ValLim)
		require.Equal(t, defaultParallelWrks, p.GoroutinesCnt)
		res := p.Do()
		require.Equal(t, 0, len(res))
	})
}

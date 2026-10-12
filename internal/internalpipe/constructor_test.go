package internalpipe

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Slice(t *testing.T) {
	t.Parallel()

	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p := Slice(a)
	require.Equal(t, p.Len, 10)
	require.Equal(t, p.ValLim, notSet)
	require.Equal(t, p.GoroutinesCnt, defaultParallelWrks)
	var r int
	require.True(t, p.Fn(15, &r))
	require.False(t, p.Fn(5, &r))
	require.Equal(t, a[5], r)
}

func Test_FuncP(t *testing.T) {
	t.Parallel()

	p := FuncP(func(i int) (*int, bool) {
		return &i, true
	})
	require.Equal(t, p.Len, notSet)
	require.Equal(t, p.ValLim, notSet)
	require.Equal(t, p.GoroutinesCnt, defaultParallelWrks)

	res := p.Gen(5).Do()
	require.Equal(t, []int{0, 1, 2, 3, 4}, res)
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
		t.Parallel()
		p := Cycle([]string{"a", "b", "c", "d"})
		require.Equal(
			t,
			[]string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
			p.Filter(isVowel).Map(strings.ToUpper).Take(10).Do(),
		)
	})

	t.Run("empty res", func(t *testing.T) {
		t.Parallel()
		p := Cycle([]string{"a", "b", "c", "d"})
		require.Equal(
			t,
			[]string{},
			p.Filter(isVowel).Filter(Not(isVowel)).Gen(10).Do(),
		)
	})

	t.Run("empty cycle", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		p := Range(0, 10, 1)
		res := p.Do()
		require.Equal(t, 10, len(res))
		for i := 0; i < 10; i++ {
			require.Equal(t, i, res[i])
		}
	})
	t.Run("single_step_owerflow", func(t *testing.T) {
		t.Parallel()
		p := Range(1, 10, 50)
		res := p.Do()
		require.Equal(t, 1, len(res))
		require.Equal(t, 1, res[0])
	})
	t.Run("finish_is_less_than_start", func(t *testing.T) {
		t.Parallel()
		p := Range(100, 10, 50)
		res := p.Do()
		require.Equal(t, 0, len(res))
	})
	t.Run("step_is_0", func(t *testing.T) {
		t.Parallel()
		p := Range(1, 10, 0)
		res := p.Do()
		require.Equal(t, 0, len(res))
	})
	t.Run("start_is_finish", func(t *testing.T) {
		t.Parallel()
		p := Range(1, 1, 1)
		res := p.Do()
		require.Equal(t, 0, len(res))
	})
	t.Run("start_is_finish_negative", func(t *testing.T) {
		t.Parallel()
		p := Range(1, 1, -1)
		res := p.Do()
		require.Equal(t, 0, len(res))
	})
	t.Run("finish_is_less_than_start_and_step_is_negative", func(t *testing.T) {
		t.Parallel()
		p := Range(100, 10, -50)
		res := p.Do()
		require.Equal(t, 2, len(res))
		require.Equal(t, 100, res[0])
		require.Equal(t, 50, res[1])
	})
}

func Test_Range_FloatLen(t *testing.T) {
	t.Parallel()

	// Float64 division can round just past an integer, so a math.Ceil-based Len
	// was off by one in either direction: it could add a phantom element or,
	// worse, drop the real last one.
	cases := []struct {
		start, finish, step float64
		want                int
	}{
		{0, 0.9, 0.3, 4},
		{0, 1.8, 0.3, 7},
		{0, 2.1, 0.3, 7},
		{0, 3.6, 1.2, 4},
		{0, 8.4, 0.6, 14},
		{0, 0.2, 0.1, 2},
		{0, 1, 0.1, 10},
		{0, 1, 0.25, 4},
		{1, 0, -0.1, 10},
		{0, -1, -0.1, 10},
	}
	for _, c := range cases {
		c := c
		p := Range(c.start, c.finish, c.step)

		yielded := 0
		var tmp float64
		for i := 0; ; i++ {
			if p.Fn(i, &tmp) {
				break
			}
			yielded++
		}

		require.Equal(t, c.want, yielded, "yielded of Range(%v, %v, %v)", c.start, c.finish, c.step)
		require.Equal(t, c.want, p.Len, "Len of Range(%v, %v, %v)", c.start, c.finish, c.step)
		require.Equal(t, c.want, len(p.Do()), "Do() len of Range(%v, %v, %v)", c.start, c.finish, c.step)
	}
}

func Test_Repeat(t *testing.T) {
	t.Parallel()

	t.Run("happy", func(t *testing.T) {
		t.Parallel()

		p := Repeat("hello", 5).Map(strings.ToUpper).Do()
		require.Equal(t, []string{"HELLO", "HELLO", "HELLO", "HELLO", "HELLO"}, p)
	})

	t.Run("n==0", func(t *testing.T) {
		t.Parallel()

		p := Repeat("hello", 0).Map(strings.ToUpper).Do()
		require.Equal(t, []string{}, p)
	})
}

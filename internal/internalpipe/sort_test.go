package internalpipe

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Sort(t *testing.T) {
	exp := make([]int, 100_000)
	for i := 0; i < 100_000; i++ {
		exp[i] = i
	}
	a := make([]int, 100_000)
	copy(a, exp)
	rand.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})

	neq := func(exp, a []int) bool {
		for i := range exp {
			if exp[i] != a[i] {
				return true
			}
		}
		return false
	}

	t.Run("single thread", func(t *testing.T) {
		require.True(t, neq(exp, a))
		p := Func(func(i int) (int, bool) {
			return a[i], true
		}).
			Take(100_000).
			Sort(func(x, y *int) bool { return *x < *y }).
			Do()
		require.Equal(t, len(exp), len(p))
		for i := range p {
			require.Equal(t, exp[i], p[i])
		}
	})

	t.Run("seven thread", func(t *testing.T) {
		require.True(t, neq(exp, a))
		p := Func(func(i int) (int, bool) {
			return a[i], true
		}).
			Take(100_000).
			Sort(func(x, y *int) bool { return *x < *y }).
			Parallel(7).
			Do()
		require.Equal(t, len(exp), len(p))
		for i := range p {
			require.Equal(t, exp[i], p[i])
		}
	})

	t.Run("single thread empty", func(t *testing.T) {
		require.True(t, neq(exp, a))
		p := Func(func(i int) (int, bool) {
			return a[i], false
		}).
			Gen(10).
			Sort(func(x, y *int) bool { return *x < *y }).
			Do()
		require.Equal(t, []int{}, p)
	})

	t.Run("seven thread empty", func(t *testing.T) {
		require.True(t, neq(exp, a))
		p := Func(func(i int) (int, bool) {
			return a[i], false
		}).
			Gen(10).
			Sort(func(x, y *int) bool { return *x < *y }).
			Parallel(7).
			Do()
		require.Equal(t, []int{}, p)
	})
}

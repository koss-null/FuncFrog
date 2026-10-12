package internalpipe

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// This file targets the non-trivial parts of the new destination-pointer
// generator contract: the parallel do() gather/compaction, chunk boundaries,
// worker-scratch reuse and element order. The expectations are always computed
// with a straight-line loop, never by re-running the code under test with a
// different goroutine count, so a bug in the shared evaluation path cannot
// mask itself.

// plainChain is an independent implementation of the pipeline built in
// deepChain, used to derive the expected result.
func plainChain(data []int) []int {
	out := make([]int, 0, len(data))
	for _, x := range data {
		x++ // Map(x + 1)
		if x%3 == 0 {
			continue // Filter(x%3 != 0)
		}
		x *= 2 // MapFilter(x*2, keep when != 0)
		if x == 0 {
			continue
		}
		if x%5 == 2 {
			continue // Filter(x%5 != 2)
		}
		x -= 1000 // Map(x - 1000)
		out = append(out, x)
	}
	return out
}

func deepChain(data []int, goroutines int) []int {
	return Slice(data).
		Parallel(uint16(goroutines)).
		Map(func(x int) int { return x + 1 }).
		Filter(func(x *int) bool { return *x%3 != 0 }).
		MapFilter(func(x int) (int, bool) { return x * 2, x*2 != 0 }).
		Filter(func(x *int) bool { return *x%5 != 2 }).
		Map(func(x int) int { return x - 1000 }).
		Do()
}

func Test_DeepChain_OrderAndSkips(t *testing.T) {
	t.Parallel()

	data := make([]int, 2000)
	for i := range data {
		data[i] = i*7 - 3
	}
	want := plainChain(data)
	require.NotEmpty(t, want)

	for _, g := range []int{1, 2, 3, 4, 5, 7, 16, 103} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()
			require.Equal(t, want, deepChain(data, g))
		})
	}
}

// Test_ParallelDo_SkipBoundaries is built so that dropped positions land on
// both ends, inside a chunk and (for the chosen goroutine counts) exactly on
// chunk borders - the in-place compaction has to survive all of them.
func Test_ParallelDo_SkipBoundaries(t *testing.T) {
	t.Parallel()

	const n = 103 // deliberately not divisible by most of the goroutine counts
	data := make([]int, n)
	for i := range data {
		data[i] = i
	}
	keep := func(i int) bool { return i != 0 && i != n-1 && i%7 != 0 }

	want := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if keep(i) {
			want = append(want, i)
		}
	}
	require.NotEmpty(t, want)

	for _, g := range []int{1, 2, 3, 4, 5, 7, 8, 13, 103} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()
			got := Slice(data).
				Parallel(uint16(g)).
				Filter(func(x *int) bool { return keep(*x) }).
				Do()
			require.Equal(t, want, got)
		})
	}
}

func Test_ParallelDo_Edges(t *testing.T) {
	t.Parallel()

	t.Run("empty pipe", func(t *testing.T) {
		t.Parallel()
		got := Slice([]int{}).Parallel(10).Map(func(x int) int { return x + 1 }).Do()
		require.Equal(t, []int{}, got)
	})

	t.Run("more workers than elements", func(t *testing.T) {
		t.Parallel()
		got := Slice([]int{1, 2, 3}).Parallel(10).Map(func(x int) int { return x * x }).Do()
		require.Equal(t, []int{1, 4, 9}, got)
	})

	t.Run("every element skipped", func(t *testing.T) {
		t.Parallel()
		data := make([]int, 1000)
		for i := range data {
			data[i] = i
		}
		got := Slice(data).
			Parallel(6).
			MapFilter(func(x int) (int, bool) { return x, false }).
			Do()
		require.Equal(t, []int{}, got)
	})

	t.Run("only first and last survive", func(t *testing.T) {
		t.Parallel()
		data := make([]int, 1000)
		for i := range data {
			data[i] = i
		}
		got := Slice(data).
			Parallel(6).
			Filter(func(x *int) bool { return *x == 0 || *x == 999 }).
			Do()
		require.Equal(t, []int{0, 999}, got)
	})
}

func Test_TakeWithSkips(t *testing.T) {
	t.Parallel()

	for _, g := range []int{1, 3, 8} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()
			got := Func(func(i int) (int, bool) { return i, true }).
				Parallel(uint16(g)).
				MapFilter(func(x int) (int, bool) { return x, x%2 == 0 }).
				Take(5).
				Do()
			require.Equal(t, []int{0, 2, 4, 6, 8}, got)
		})
	}
}

// Test_Reduce_OrderSensitive fails loudly if the parallel gather ever reorders
// elements: the parallel Reduce folds over Do(), so a shuffled result would
// change a non-commutative reduction.
func Test_Reduce_OrderSensitive(t *testing.T) {
	t.Parallel()

	data := []int{1, 2, 3, 4, 5, 6}
	sub := func(a, b *int) int { return *a - *b }

	want := data[0]
	for _, v := range data[1:] {
		want -= v // -19
	}

	t.Run("single thread", func(t *testing.T) {
		t.Parallel()
		got := Slice(data).Reduce(sub)
		require.NotNil(t, got)
		require.Equal(t, want, *got)
	})

	t.Run("parallel preserves order", func(t *testing.T) {
		t.Parallel()
		got := Slice(data).Parallel(4).Reduce(sub)
		require.NotNil(t, got)
		require.Equal(t, want, *got)
	})

	t.Run("parallel with skips", func(t *testing.T) {
		t.Parallel()
		got := Slice(data).
			Parallel(4).
			Filter(func(x *int) bool { return *x%2 == 0 }). // 2,4,6
			Reduce(sub)
		require.NotNil(t, got)
		require.Equal(t, 2-4-6, *got) // -8
	})
}

func Test_Count_ParallelWithSkips(t *testing.T) {
	t.Parallel()

	data := make([]int, 5000)
	for i := range data {
		data[i] = i
	}

	for _, g := range []int{1, 4, 9} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()
			got := Slice(data).
				Parallel(uint16(g)).
				Filter(func(x *int) bool { return *x%3 == 0 }).
				Count()
			require.Equal(t, 1667, got) // multiples of 3 in [0, 5000)
		})
	}
}

func Test_Sum_ParallelWithSkips(t *testing.T) {
	t.Parallel()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	plus := func(a, b *int) int { return *a + *b }

	want := 0
	for _, v := range data {
		if v%2 == 0 {
			want += v
		}
	}
	require.Equal(t, 249500, want)

	for _, g := range []int{1, 4, 7} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()
			got := Slice(data).
				Parallel(uint16(g)).
				Filter(func(x *int) bool { return *x%2 == 0 }).
				Sum(plus)
			require.Equal(t, want, got)
		})
	}
}

func Test_Any_ParallelEdges(t *testing.T) {
	t.Parallel()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	last := len(data) - 1

	t.Run("only the last element matches", func(t *testing.T) {
		t.Parallel()
		got := Slice(data).
			Parallel(7).
			Filter(func(x *int) bool { return *x == last }).
			Any()
		require.NotNil(t, got)
		require.Equal(t, last, *got)
	})

	t.Run("nothing matches -> nil", func(t *testing.T) {
		t.Parallel()
		got := Slice(data).
			Parallel(7).
			Filter(func(*int) bool { return false }).
			Any()
		require.Nil(t, got)
	})

	t.Run("single thread only the last element matches", func(t *testing.T) {
		t.Parallel()
		got := Slice(data).
			Filter(func(x *int) bool { return *x == last }).
			Any()
		require.NotNil(t, got)
		require.Equal(t, last, *got)
	})
}

// Test_First_ParallelLowestIndex pins the ordering guarantee: with g=4 the
// chunk size is 250, so the cases below probe a border, just-after-a-border,
// the very last element and the first chunk.
func Test_First_ParallelLowestIndex(t *testing.T) {
	t.Parallel()

	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	cases := []struct {
		name string
		pred func(*int) bool
		want int
	}{
		{"match on a chunk border", func(x *int) bool { return *x >= 500 }, 500},
		{"match just after a border", func(x *int) bool { return *x >= 501 }, 501},
		{"only the last element", func(x *int) bool { return *x == 999 }, 999},
		{"match in the first chunk", func(x *int) bool { return *x >= 7 }, 7},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := Slice(data).Parallel(4).Filter(c.pred).First()
			require.NotNil(t, got)
			require.Equal(t, c.want, *got)
		})
	}
}

// Test_ParallelDo_Deterministic runs the same parallel pipeline many times and
// requires byte-identical output every time - a cheap guard against
// non-determinism and data races in the worker/gather machinery.
func Test_ParallelDo_Deterministic(t *testing.T) {
	t.Parallel()

	rnd := rand.New(rand.NewSource(1))
	data := make([]int, 6000)
	for i := range data {
		data[i] = rnd.Intn(1000)
	}
	pred := func(x *int) bool { return *x%3 != 1 }

	ref := make([]int, 0, len(data))
	for _, v := range Slice(data).Filter(pred).Do() {
		ref = append(ref, v*2)
	}

	for i := 0; i < 30; i++ {
		got := Slice(data).
			Parallel(5).
			Filter(pred).
			Map(func(x int) int { return x * 2 }).
			Do()
		require.Equal(t, ref, got)
	}
}

// Test_Erase_ParallelRoundTrip checks the erased pipe keeps a distinct pointer
// per surviving element, in order, even through the parallel path.
func Test_Erase_ParallelRoundTrip(t *testing.T) {
	t.Parallel()

	data := make([]int, 500)
	for i := range data {
		data[i] = i
	}

	got := Slice(data).
		Parallel(4).
		Filter(func(x *int) bool { return *x%2 == 0 }).
		Erase().
		Do()

	require.Equal(t, 250, len(got))
	for idx, v := range got {
		p, ok := v.(*int)
		require.True(t, ok, "element %d is not a *int", idx)
		require.Equal(t, idx*2, *p)
	}
}

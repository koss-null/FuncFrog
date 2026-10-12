package pipe_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/funcfrog/pkg/pipe"
)

// Complex end-to-end cases on the *public* API: they exercise the prefix
// helpers (pkg/pipe Map/Collect) and the parallel gather path with type
// changes, skipping and sorting - the places where the new internal
// destination-pointer generator is easiest to get subtly wrong.

// plainDeepChain mirrors the pipeline built in Test_Contract_DeepChainTypeChange.
func plainDeepChain(data []int) []string {
	out := make([]string, 0, len(data))
	for _, x := range data {
		if x%2 != 0 {
			continue // Filter(even)
		}
		s := fmt.Sprintf("#%d", x*10) // prefix Map(int -> string)
		if len(s) <= 1 {
			continue // Filter(len > 1)
		}
		out = append(out, s)
	}
	return out
}

func Test_Contract_DeepChainTypeChange(t *testing.T) {
	t.Parallel()

	data := make([]int, 1500)
	for i := range data {
		data[i] = i
	}
	want := plainDeepChain(data)
	require.NotEmpty(t, want)

	for _, g := range []int{1, 2, 4, 7, 32} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()

			evens := pipe.Slice(data).
				Filter(func(x *int) bool { return *x%2 == 0 }).
				Parallel(uint16(g))

			got := pipe.Map(evens, func(x int) string { return fmt.Sprintf("#%d", x*10) }).
				Filter(func(s *string) bool { return len(*s) > 1 }).
				Do()

			require.Equal(t, want, got)
		})
	}
}

func Test_Contract_CollectParallelRoundTrip(t *testing.T) {
	t.Parallel()

	data := make([]int, 800)
	for i := range data {
		data[i] = i
	}
	want := make([]int, 0, 400)
	for i := 0; i < len(data); i += 2 {
		want = append(want, i)
	}

	for _, g := range []int{1, 3, 5, 16} {
		g := g
		t.Run(fmt.Sprintf("goroutines=%d", g), func(t *testing.T) {
			t.Parallel()

			got := pipe.Collect[int](
				pipe.Slice(data).
					Parallel(uint16(g)).
					Filter(func(x *int) bool { return *x%2 == 0 }).
					Erase(),
			).Do()

			require.Equal(t, want, got)
		})
	}
}

// Test_Contract_FilterThenSort covers Sort running over a slice that Filter
// made shorter while the pipe still reports the original length - the tail
// positions must be skipped, not read out of range.
func Test_Contract_FilterThenSort(t *testing.T) {
	t.Parallel()

	data := []int{5, 3, 8, 1, 9, 2, 7, 4, 6, 0}

	got := pipe.Slice(data).
		Filter(func(x *int) bool { return *x%2 == 0 }).
		Sort(func(a, b *int) bool { return *a < *b }).
		Parallel(3).
		Do()

	require.Equal(t, []int{0, 2, 4, 6, 8}, got)
}

func Test_Contract_ReduceOrderSensitive(t *testing.T) {
	t.Parallel()

	sub := func(a, b *int) int { return *a - *b }

	t.Run("parallel keeps order", func(t *testing.T) {
		t.Parallel()
		got := pipe.Slice([]int{1, 2, 3, 4, 5, 6}).Parallel(3).Reduce(sub)
		require.NotNil(t, got)
		require.Equal(t, -19, *got) // ((1-2)-3)-4-5-6
	})

	t.Run("parallel with type change", func(t *testing.T) {
		t.Parallel()
		// int -> string, then Reduce over the mapped values in order.
		got := pipe.Map(
			pipe.Slice([]string{"a", "b", "c"}).Parallel(3),
			func(s string) string { return s + "!" },
		).Reduce(func(a, b *string) string { return *a + *b })
		require.NotNil(t, got)
		require.Equal(t, "a!b!c!", *got)
	})
}

func Test_Contract_ParallelDeterministic(t *testing.T) {
	t.Parallel()

	rnd := rand.New(rand.NewSource(2))
	data := make([]int, 5000)
	for i := range data {
		data[i] = rnd.Intn(500)
	}
	keep := func(x *int) bool { return *x%4 != 2 }
	dbl := func(x int) int { return x * 2 }

	want := pipe.Slice(data).Filter(keep).Map(dbl).Do()
	require.NotEmpty(t, want)

	for i := 0; i < 20; i++ {
		got := pipe.Slice(data).Parallel(6).Filter(keep).Map(dbl).Do()
		require.Equal(t, want, got)
	}
}

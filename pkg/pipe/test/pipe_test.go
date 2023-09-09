package pipe_test

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/funcfrog/internal/primitive/pointer"
	"github.com/koss-null/funcfrog/pkg/pipe"
	"github.com/koss-null/funcfrog/pkg/pipies"
)

var (
	once sync.Once
	a1kk []float64
)

func initA1kk() {
	once.Do(func() {
		a1kk = make([]float64, 1_000_000)
		for i := range a1kk {
			a1kk[i] = float64(i)
		}
	})
}

func TestSlice_ok(t *testing.T) {
	t.Parallel()
	initA1kk()

	// Slice() function
	t.Run("Slice() happy 100k slice", func(t *testing.T) {
		a := make([]struct {
			i int
			f float64
		}, 100_000)
		s := pipe.Slice(a).Do()
		for i := range a {
			require.Equal(t, a[i], s[i])
		}
	})

	t.Run("Slice() happy 1kk slice", func(t *testing.T) {
		s := pipe.Slice(a1kk).Do()
		for i := range a1kk {
			require.Equal(t, a1kk[i], s[i])
		}
	})

	t.Run("Slice() happy struct slice", func(t *testing.T) {
		a := make([]struct {
			i int
			f float64
		}, 100)
		s := pipe.Slice(a).Do()
		for i := range a {
			require.Equal(t, a[i], s[i])
		}
	})

	// Map() function
	t.Run("Map().Map() 1kk", func(t *testing.T) {
		s := pipe.Slice(a1kk).
			Parallel(4).
			Map(func(x float64) float64 { return x * x * x }).
			Map(math.Sqrt).
			Do()

		for i := range a1kk {
			aaa := float64(a1kk[i] * a1kk[i] * a1kk[i])
			require.Equal(t, math.Sqrt(aaa), s[i])
		}
	})

	// First() function
	t.Run("Slice().First() happy", func(t *testing.T) {
		s := pipe.Slice(a1kk).
			Filter(func(x *float64) bool { return *x > 100_000 }).
			First()
		require.NotNil(t, s)
		require.Equal(t, float64(100_001), *s)
	})

	t.Run("Func().First() happy", func(t *testing.T) {
		s := pipe.Func(func(i int) (float64, bool) { return float64(i), true }).
			Filter(func(x *float64) bool { return *x > 100_000 }).
			Take(200_000).
			First()
		require.NotNil(t, s)
		require.Equal(t, float64(100_001), *s)
	})

	t.Run("Func().First() happy no limit", func(t *testing.T) {
		s := pipe.Fn(func(i int) float64 { return float64(i) }).
			Filter(func(x *float64) bool { return *x > 100_000 }).
			First()
		require.NotNil(t, s)
		require.Equal(t, float64(100_001), *s)
	})

	// Filter() function
	t.Run("Func().Filter() happy remove nils", func(t *testing.T) {
	})
}

func TestFilter_NotNull_ok(t *testing.T) {
	genFunc := func(i int) (*float64, bool) {
		if i%10 == 0 {
			return nil, true
		}
		return pointer.To(float64(i)), true
	}

	s := pipe.Map(
		pipe.Func(genFunc).
			Filter(pipies.NotNil[*float64]).
			Take(10_000),
		pointer.From[float64],
	).
		Sum(pipies.Sum[float64])
	require.NotNil(t, s)

	sm := 0
	for i := 1; i < 10000; i++ {
		if i%10 != 0 {
			sm += i
		}
	}
	require.Equal(t, float64(sm), s)
}

// Sort() function

func TestSort_ok_parallel1(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	rnd.Seed(11)
	s := pipe.Func(func(i int) (float32, bool) {
		return rnd.Float32(), true
	}).
		Parallel(1).
		Take(100_000).
		Sort(pipies.Less[float32]).
		Do()

	require.NotNil(t, s)
	prevItem := s[0]
	for _, item := range s {
		require.GreaterOrEqual(t, item, prevItem)
	}
}

func TestSort_ok_parallel_default(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	rnd.Seed(11)
	s := pipe.Func(func(i int) (float32, bool) {
		return rnd.Float32(), true
	}).
		Take(100_000).
		Sort(pipies.Less[float32]).
		Do()

	require.NotNil(t, s)
	prevItem := s[0]
	for _, item := range s {
		require.GreaterOrEqual(t, item, prevItem)
	}
}

func TestSort_ok_parallel_slice(t *testing.T) {
	a := make([]int, 6000)
	rnd := rand.New(rand.NewSource(42))
	rnd.Seed(22)
	for i := range a {
		a[i] = int(rnd.Float32() * 100_000.0)
	}

	s := pipe.Slice(a).
		Sort(pipies.Less[int]).
		Do()

	require.NotNil(t, s)
	sort.Ints(a)
	for i := range a {
		require.Equal(t, s[i], a[i])
	}
}

func TestSort_ok_parallel12(t *testing.T) {
	s := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Parallel(12).
		Take(100_000).
		Sort(pipies.Less[float32]).
		Do()

	require.NotNil(t, s)
	prevItem := s[0]
	for _, item := range s {
		require.GreaterOrEqual(t, item, prevItem)
	}
}

func TestSort_ok_parallel_large(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	s := pipe.Func(func(i int) (float32, bool) {
		return rnd.Float32(), true
	}).
		Parallel(12).
		Take(10_000_000).
		Sort(pipies.Less[float32]).
		Do()

	require.NotNil(t, s)
	prevItem := s[0]
	for _, item := range s {
		require.GreaterOrEqual(t, item, prevItem)
	}
}

func TestReduce(t *testing.T) {
	res := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(6000).
		Reduce(func(a, b *int) int { return *a + *b })

	expected := 0
	for i := 1; i < 6000; i++ {
		expected += i
	}
	require.Equal(t, expected, *res)
}

func TestSum(t *testing.T) {
	res := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(6000).
		Sum(func(a, b *int) int { return *a + *b })

	expected := 0
	for i := 1; i < 6000; i++ {
		expected += i
	}
	require.Equal(t, expected, res)
}

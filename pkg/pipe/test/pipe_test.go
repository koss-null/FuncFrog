package pipe_test

import (
	"crypto/rand"
	"math"
	"math/big"
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

var (
	randSrc []int
	randMx  sync.Mutex
)

func randN(n int) []int {
	randMx.Lock()
	defer randMx.Unlock()

	if len(randSrc) != 0 {
		dest := make([]int, n)
		copy(dest, randSrc[:n])
		return dest
	}

	result := make([]int, 10e8)
	for i := 0; i < n; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(10e8))
		if err != nil {
			panic(err)
		}
		result[i] = int(nBig.Int64())
	}

	randSrc = result
	dest := make([]int, n)
	copy(dest, randSrc[:n])
	return dest
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
	rnd := randN(100_000)
	s := pipe.Func(func(i int) (float32, bool) {
		return float32(rnd[i]), true
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
	rnd := randN(100_000)
	s := pipe.Func(func(i int) (float32, bool) {
		return float32(rnd[i]), true
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
	rnd := randN(100_000)
	for i := range a {
		a[i] = int(rnd[i] * 100_000)
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
	rnd := randN(100_000)
	s := pipe.Func(func(i int) (float32, bool) {
		return float32(rnd[i]), true
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
	largeNumber := 6_000_000
	rnd := randN(largeNumber)
	s := pipe.Func(func(i int) (float32, bool) {
		return float32(rnd[i]), true
	}).
		Parallel(12).
		Take(largeNumber).
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

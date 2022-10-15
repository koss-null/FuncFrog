package pipe_test

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/pkg/pipe"
)

var once sync.Once
var a10kk []float64

func initA10kk() {
	once.Do(func() {
		a10kk = make([]float64, 10_000_000)
		for i := range a10kk {
			a10kk[i] = float64(i)
		}
	})
}

// Slice() function

func TestSlice_ok(t *testing.T) {
	a := make([]struct {
		i int
		f float64
	}, 100_000)
	s := pipe.Slice(a).Do()
	for i := range a {
		require.Equal(t, a[i], s[i])
	}
}

func TestSlice_ok_test2(t *testing.T) {
	initA10kk()
	s := pipe.Slice(a10kk).Do()
	for i := range a10kk {
		require.Equal(t, a10kk[i], s[i])
	}
}

func TestSlice_ok_test3(t *testing.T) {
	a := make([]struct {
		i int
		f float64
	}, 100)
	s := pipe.Slice(a).Do()
	for i := range a {
		require.Equal(t, a[i], s[i])
	}
}

func TestSlice_ok_test4(t *testing.T) {
	initA10kk()

	s := pipe.Slice(a10kk).Do()
	for i := range a10kk {
		require.Equal(t, a10kk[i], s[i])
	}
}

// Map() function

func TestMap_ok(t *testing.T) {
	initA10kk()

	s := pipe.Slice(a10kk).
		Parallel(12).
		Map(func(x float64) float64 { return x * x * x }).
		Map(math.Sqrt).
		Do()

	for i := range a10kk {
		aaa := float64(a10kk[i] * a10kk[i] * a10kk[i])
		require.Equal(t, math.Sqrt(aaa), s[i])
	}
}

// First() function

func TestFirst_ok_slice(t *testing.T) {
	initA10kk()

	s := pipe.Slice(a10kk).
		Filter(func(x float64) bool { return x > 100_000 }).
		First()
	require.NotNil(t, s)
	require.Equal(t, *s, float64(100_001))
}

func TestFirst_ok_func(t *testing.T) {
	initA10kk()

	s := pipe.Func(func(i int) (float64, bool) { return float64(i), true }).
		Filter(func(x float64) bool { return x > 100_000 }).
		Take(200_000).
		First()
	require.NotNil(t, s)
	require.Equal(t, *s, float64(100_001))
}

func TestFirst_ok_func_bigint_nogen_notake(t *testing.T) {
	initA10kk()

	s := pipe.Func(func(i int) (float64, bool) { return float64(i), true }).
		Filter(func(x float64) bool { return x > 100_000 }).
		Take(math.MaxInt64).
		First()
	require.NotNil(t, s)
	require.Equal(t, *s, float64(100_001))
}

func TestSum_ok_slice(t *testing.T) {
	initA10kk()

	s := pipe.Slice(a10kk).
		Filter(func(x float64) bool { return x > 100_000 }).
		Sum(pipe.Sum[float64])
	require.NotNil(t, s)
	require.Equal(t, *s, float64(49994994950000))
}

func TestSum_ok_func_gen(t *testing.T) {
	s := pipe.Func(func(i int) (float64, bool) { return float64(i), true }).
		Filter(func(x float64) bool { return x > 100_000 }).
		Gen(10_000_000).
		Sum(pipe.Sum[float64])
	require.NotNil(t, s)
	require.Equal(t, *s, float64(49994994950000))
}

func TestSum_ok_func_take(t *testing.T) {
	s := pipe.Func(func(i int) (float64, bool) { return float64(i), true }).
		Filter(func(x float64) bool { return x > 100_000 }).
		Take(10_000_000).
		Sum(pipe.Sum[float64])
	require.NotNil(t, s)
	require.Equal(t, *s, float64(51000005000000))
}

package pipe_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/pkg/pipe"
)

// Slice() function

func Slice_ok_test(t *testing.T) {
	a := make([]struct {
		i int
		f float64
	}, 100_000)
	s := pipe.Slice(a).Do()
	for i := range a {
		require.Equal(t, a[i], s[i])
	}
}

func Slice_ok_test2(t *testing.T) {
	a := make([]struct {
		i int
		f float64
	}, 100_000_000)
	s := pipe.Slice(a).Do()
	for i := range a {
		require.Equal(t, a[i], s[i])
	}
}

func Slice_ok_test3(t *testing.T) {
	a := make([]struct {
		i int
		f float64
	}, 100)
	s := pipe.Slice(a).Do()
	for i := range a {
		require.Equal(t, a[i], s[i])
	}
}

func Slice_ok_test4(t *testing.T) {
	a := make([]int, 100_000_000)
	for i := range a {
		a[i] = i
	}
	s := pipe.Slice(a).Do()
	for i := range a {
		require.Equal(t, a[i], s[i])
	}
}

// Map() function

func Map_ok_test(t *testing.T) {
	a := make([]float64, 100_000_000)
	for i := range a {
		a[i] = float64(i)
	}

	s := pipe.Slice(a).
		Map(math.Sqrt).
		Do()

	for i := range a {
		aaa := float64(a[i] * a[i] * a[i])
		require.Equal(t, int(math.Sqrt(aaa)), s[i])
	}
}

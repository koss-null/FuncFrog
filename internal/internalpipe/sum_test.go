package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/internal/primitive/pointer"
)

func TestSumOk1thread(t *testing.T) {
	initA100k()

	s := Sum(
		1,
		len(a100k),
		func(x, y *float64) float64 {
			return *x + *y
		},
		func(i int) (*float64, bool) {
			return &a100k[i], false
		},
	)

	require.NotNil(t, s)
	require.Equal(t, 4999950000.0, s)
}

func TestSumOk4thread(t *testing.T) {
	initA100k()

	s := Sum(
		4,
		len(a100k),
		func(x, y *float64) float64 {
			return *x + *y
		},
		func(i int) (*float64, bool) {
			return &a100k[i], false
		},
	)

	require.NotNil(t, s)
	require.Equal(t, 4999950000.0, s)
}

func TestSumOk1threadEmpty(t *testing.T) {
	s := Sum(
		1,
		0,
		func(x, y *float64) float64 {
			return *x + *y
		},
		func(i int) (*float64, bool) {
			return pointer.To(1.), false
		},
	)

	require.NotNil(t, s)
	require.Equal(t, 0.0, s)
}

func TestSumOk4threadEmpty(t *testing.T) {
	s := Sum(
		4,
		0,
		func(x, y *float64) float64 {
			return *x + *y
		},
		func(i int) (*float64, bool) {
			return pointer.To(1.), false
		},
	)

	require.NotNil(t, s)
	require.True(t, s == 0)
}

func TestSumOk1threadSingle(t *testing.T) {
	s := Sum(
		1,
		1,
		func(x, y *float64) float64 {
			return *x + *y
		},
		func(i int) (*float64, bool) {
			return pointer.To(100500.), i != 0
		},
	)

	require.NotNil(t, s)
	require.True(t, s == 100500.)
}

func TestSumOk4threadSingle(t *testing.T) {
	s := Sum(
		4,
		1,
		func(x, y *float64) float64 {
			return *x + *y
		},
		func(i int) (*float64, bool) {
			return pointer.To(100500.), i != 0
		},
	)

	require.NotNil(t, s)
	require.True(t, s == 100500.)
}

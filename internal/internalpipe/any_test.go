package internalpipe

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/internal/primitive/pointer"
)

var (
	once  sync.Once
	a100k []float64
)

func initA100k() {
	once.Do(func() {
		a100k = make([]float64, 100_000)
		for i := range a100k {
			a100k[i] = float64(i)
		}
	})
}

func TestAnyOkNolimit1Thread(t *testing.T) {
	initA100k()

	s := Any(false, -1, 1, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), a100k[i] <= 90_000.0
	})

	require.NotNil(t, s)
	require.Greater(t, pointer.From(s), 90_000.0)
}

func TestAnyOkNolimit4thread(t *testing.T) {
	initA100k()

	s := Any(false, -1, 4, func(i int) (*float64, bool) {
		if i > len(a100k) {
			return nil, true
		}
		return pointer.To(a100k[i]), a100k[i] <= 90_000.0
	})

	require.NotNil(t, s)
	require.Greater(t, pointer.From(s), 90_000.0)
}

func TestAnyOkLimit1thread(t *testing.T) {
	initA100k()

	s := Any(true, len(a100k), 1, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), a100k[i] <= 90_000.0
	})

	require.NotNil(t, s)
	require.True(t, pointer.From(s) > 90_000.0)
}

func TestAnyOkLimit4thread(t *testing.T) {
	initA100k()

	s := Any(true, len(a100k), 4, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), a100k[i] <= 90_000.0
	})

	require.NotNil(t, s)
	require.True(t, pointer.From(s) > 90_000.0)
}

func TestAnyNFLimit1thread(t *testing.T) {
	initA100k()

	s := Any(true, len(a100k), 1, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), true
	})

	require.Nil(t, s)
}

func TestAnyNFLimit4thread(t *testing.T) {
	initA100k()

	s := Any(true, len(a100k), 4, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), true
	})

	require.Nil(t, s)
}

func TestAnyOkSingleElement1threadFinite(t *testing.T) {
	initA100k()

	s := Any(true, len(a100k), 1, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), !(a100k[i] > 90_000.0 && a100k[i] < 90_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(90_001), *s)
}

func TestAnyOkSingleElement4threadFinite(t *testing.T) {
	initA100k()

	s := Any(true, len(a100k), 4, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), !(a100k[i] > 90_000.0 && a100k[i] < 90_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(90_001), *s)
}

func TestAnyOkSingleElement1threadInfinite(t *testing.T) {
	initA100k()

	s := Any(false, -1, 1, func(i int) (*float64, bool) {
		return pointer.To(a100k[i]), !(a100k[i] > 90_000.0 && a100k[i] < 90_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(90_001), *s)
}

func TestAnyOkSingleElement4threadInfinite(t *testing.T) {
	initA100k()

	s := Any(false, -1, 4, func(i int) (*float64, bool) {
		if i >= len(a100k) {
			return pointer.To(100_000.0), true
		}
		return pointer.To(a100k[i]), !(a100k[i] > 90_000.0 && a100k[i] < 90_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(90_001), *s)
}

package internalpipe

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/internal/primitive/pointer"
)

var (
	once  sync.Once
	a10kk []float64
)

func initA10kk() {
	once.Do(func() {
		a10kk = make([]float64, 10_000_000)
		for i := range a10kk {
			a10kk[i] = float64(i)
		}
	})
}

func TestAnyOkNolimit1Thread(t *testing.T) {
	initA10kk()

	s := Any(false, -1, 1, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), a10kk[i] <= 100_000.0
	})

	require.NotNil(t, s)
	require.Greater(t, pointer.From(s), 100_000.0)
}

func TestAnyOkNolimit4thread(t *testing.T) {
	initA10kk()

	s := Any(false, -1, 4, func(i int) (*float64, bool) {
		if i > len(a10kk) {
			return nil, true
		}
		return pointer.To(a10kk[i]), a10kk[i] <= 100_000.0
	})

	require.NotNil(t, s)
	require.Greater(t, pointer.From(s), 100_000.0)
}

func TestAnyOkLimit1thread(t *testing.T) {
	initA10kk()

	s := Any(true, len(a10kk), 1, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), a10kk[i] <= 100_000.0
	})

	require.NotNil(t, s)
	require.True(t, pointer.From(s) > 100_000.0)
}

func TestAnyOkLimit4thread(t *testing.T) {
	initA10kk()

	s := Any(true, len(a10kk), 4, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), a10kk[i] <= 100_000.0
	})

	require.NotNil(t, s)
	require.True(t, pointer.From(s) > 100_000.0)
}

func TestAnyNFLimit1thread(t *testing.T) {
	initA10kk()

	s := Any(true, len(a10kk), 1, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), true
	})

	require.Nil(t, s)
}

func TestAnyNFLimit4thread(t *testing.T) {
	initA10kk()

	s := Any(true, len(a10kk), 4, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), true
	})

	require.Nil(t, s)
}

func TestAnyOkSingleElement1threadFinite(t *testing.T) {
	initA10kk()

	s := Any(true, len(a10kk), 1, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), !(a10kk[i] > 100_000.0 && a10kk[i] < 100_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(100_001), *s)
}

func TestAnyOkSingleElement4threadFinite(t *testing.T) {
	initA10kk()

	s := Any(true, len(a10kk), 4, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), !(a10kk[i] > 100_000.0 && a10kk[i] < 100_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(100_001), *s)
}

func TestAnyOkSingleElement1threadInfinite(t *testing.T) {
	initA10kk()

	s := Any(false, -1, 1, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), !(a10kk[i] > 100_000.0 && a10kk[i] < 100_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(100_001), *s)
}

func TestAnyOkSingleElement4threadInfinite(t *testing.T) {
	initA10kk()

	s := Any(false, -1, 4, func(i int) (*float64, bool) {
		return pointer.To(a10kk[i]), !(a10kk[i] > 100_000.0 && a10kk[i] < 100_002.0)
	})

	require.NotNil(t, s)
	require.Equal(t, float64(100_001), *s)
}

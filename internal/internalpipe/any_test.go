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

func TestAny(t *testing.T) {
	initA100k()
	t.Parallel()

	t.Run("Single thread no limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			return a100k[i], a100k[i] <= 90_000.0
		})
		s := p.Any()
		require.NotNil(t, s)
		require.Greater(t, 90_000.0, *s)
	})

	t.Run("Seven thread no limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			if i >= len(a100k) {
				return 0., false
			}
			return a100k[i], a100k[i] <= 90_000.0
		}).Parallel(7)
		s := p.Any()
		require.NotNil(t, s)
		require.Greater(t, 90_000.0, *s)
	})

	t.Run("Single thread limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			return a100k[i], a100k[i] <= 90_000.0
		}).Gen(len(a100k))
		s := p.Any()
		require.NotNil(t, s)
		require.Greater(t, 90_000.0, pointer.From(s))
	})

	t.Run("Seven thread limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			if i >= len(a100k) {
				return 0., false
			}
			return a100k[i], a100k[i] <= 90_000.0
		}).Gen(len(a100k)).Parallel(7)
		s := p.Any()
		require.NotNil(t, s)
		require.Greater(t, 90_000.0, pointer.From(s))
	})

	t.Run("Single thread NF limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			return a100k[i], false
		}).Gen(len(a100k))
		s := p.Any()
		require.Nil(t, s)
	})

	t.Run("Seven thread NF limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			if i >= len(a100k) {
				return 0., false
			}
			return a100k[i], false
		}).Gen(len(a100k)).Parallel(7)
		s := p.Any()
		require.Nil(t, s)
	})

	t.Run("Single thread bounded limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			return a100k[i], false
		}).Gen(len(a100k))
		s := p.Any()
		require.Nil(t, s)
	})

	t.Run("Seven thread bounded limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			if i >= len(a100k) {
				return 0., false
			}
			return a100k[i], a100k[i] > 90_000.0 && a100k[i] < 90_002.0
		}).Gen(len(a100k)).Parallel(7)
		s := p.Any()
		require.NotNil(t, s)
		require.Equal(t, 90_001., *s)
	})

	t.Run("Single thread bounded no limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			if i >= len(a100k) {
				return 0., false
			}
			return a100k[i], a100k[i] > 90_000.0 && a100k[i] < 90_002.0
		})
		s := p.Any()
		require.NotNil(t, s)
		require.Equal(t, 90_001., *s)
	})

	t.Run("Seven thread bounded no limit", func(t *testing.T) {
		p := Func(func(i int) (float64, bool) {
			if i >= len(a100k) {
				return 0., false
			}
			return a100k[i], a100k[i] > 90_000.0 && a100k[i] < 90_002.0
		}).Parallel(7)
		s := p.Any()
		require.NotNil(t, s)
		require.Equal(t, 90_001., *s)
	})
}

package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromices(t *testing.T) {
	t.Parallel()

	a := Func(func(i int) (int, bool) {
		return i, true
	}).Gen(100)

	proms := a.Promices()
	for i := 0; i < 100; i++ {
		res, ok := proms[i]()
		require.True(t, ok)
		require.Equal(t, i, res)
	}
}

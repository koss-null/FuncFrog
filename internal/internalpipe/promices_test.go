package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromices(t *testing.T) {
	t.Parallel()

	a := Func(func(i int) (int, bool) {
		return i, true
	}).Filter(func(x *int) bool { return *x != 100 }).Take(101)

	proms := a.Promices()
	for i := 0; i < 100; i++ {
		res, ok := proms[i]()
		require.True(t, ok)
		if i == 100 {
			require.Equal(t, i+1, res)
			continue
		}
		require.Equal(t, i, res)
	}
	_, ok := proms[100]()
	require.False(t, ok)
}

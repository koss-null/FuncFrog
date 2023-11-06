package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErase(t *testing.T) {
	t.Parallel()

	a := Func(func(i int) (int, bool) {
		return i, true
	}).Gen(100)

	er := a.Erase()
	res := er.Do()
	for i := 0; i < 100; i++ {
		require.Equal(t, i, *(res[i].(*int)))
	}
}

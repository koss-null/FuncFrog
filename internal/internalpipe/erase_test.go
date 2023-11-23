package internalpipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErase(t *testing.T) {
	t.Parallel()

	a := Func(func(i int) (int, bool) {
		if i%5 == 0 {
			return 10, false
		}
		return i, true
	}).Gen(100)

	er := a.Erase()
	res := er.Do()
	idx := 0
	for i := 0; i < 100; i++ {
		if i%5 != 0 {
			require.Equal(t, i, *(res[idx].(*int)))
			idx++
		}
	}
}

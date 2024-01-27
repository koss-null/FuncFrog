package pointer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Pointer(t *testing.T) {
	t.Parallel()

	t.Run("To", func(t *testing.T) {
		require.Equal(t, 10, *Ref(10))
		require.Equal(t, 10., *Ref(10.))
		s := struct{ a, b int }{1, 2}
		require.Equal(t, s, *Ref(s))
	})

	t.Run("From", func(t *testing.T) {
		require.Equal(t, 10, Deref(Ref(10)))
		require.Equal(t, 10., Deref(Ref(10.)))
		require.Equal(t, 0, Deref[int](nil))
	})
}

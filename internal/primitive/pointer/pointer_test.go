package pointer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Pointer(t *testing.T) {
	t.Parallel()

	t.Run("To", func(t *testing.T) {
		require.Equal(t, 10, *To(10))
		require.Equal(t, 10., *To(10.))
		s := struct{ a, b int }{1, 2}
		require.Equal(t, s, *To(s))
	})

	t.Run("From", func(t *testing.T) {
		require.Equal(t, 10, From(To(10)))
		require.Equal(t, 10., From(To(10.)))
		require.Equal(t, 0, From[int](nil))
	})
}

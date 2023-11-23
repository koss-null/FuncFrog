package internalpipe

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipe_Snag(t *testing.T) {
	t.Parallel()

	t.Run("empty yeti", func(t *testing.T) {
		t.Parallel()

		handlerCalled := false
		_ = Func(func(i int) (int, bool) {
			return i, true
		}).Take(1000).Snag(func(_ error) { handlerCalled = true }).Do()
		require.False(t, handlerCalled)
	})

	t.Run("happy yeti snag", func(t *testing.T) {
		t.Parallel()

		yeti := NewYeti()
		handlerCalled := false
		_ = Func(func(i int) (int, bool) {
			if i == 10 {
				yeti.Yeet(fmt.Errorf("failed on %d", i))
				return i, false
			}
			return i, true
		}).Yeti(yeti).Snag(func(_ error) { handlerCalled = true }).Take(1000).Do()
		require.True(t, handlerCalled)
	})

	t.Run("happy double yeti snag", func(t *testing.T) {
		t.Parallel()

		yeti, yeti2 := NewYeti(), NewYeti()
		handlerCalled := false
		_ = Func(func(i int) (int, bool) {
			if i == 10 {
				yeti.Yeet(fmt.Errorf("failed on %d", i))
				return i, false
			}
			return i, true
		}).Yeti(yeti2).Yeti(yeti).Snag(func(_ error) { handlerCalled = true }).Take(1000).Do()
		require.True(t, handlerCalled)
	})

	t.Run("yeti not set no snag handled", func(t *testing.T) {
		t.Parallel()

		yeti := NewYeti()
		handlerCalled := false
		_ = Func(func(i int) (int, bool) {
			if i == 10 {
				yeti.Yeet(fmt.Errorf("failed on %d", i))
				return i, false
			}
			return i, true
		}).Snag(func(_ error) { handlerCalled = true }).Take(1000).Do()
		require.False(t, handlerCalled)
	})
}

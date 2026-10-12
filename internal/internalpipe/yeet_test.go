package internalpipe

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestYeti_Yeet(t *testing.T) {
	t.Parallel()

	yeti := NewYeti()

	err := errors.New("sample error")
	yeti.Yeet(err)

	require.Contains(t, yeti.errs, err, "Error not added to Yeti's errors")

	err2 := errors.New("another error")
	yeti.Yeet(err2)

	require.Contains(t, yeti.errs, err2, "Error not added to Yeti's errors")
}

func TestYeti_Snag(t *testing.T) {
	t.Parallel()

	yeti := NewYeti()

	handler := func(err error) {}
	yeti.Snag(handler)

	require.Equal(t, 1, len(yeti.handlers), "Error handler not added to Yeti's handlers")
}

func TestYeti_Handle(t *testing.T) {
	t.Parallel()

	someErr := errors.New("some error")
	yeti := NewYeti()

	handlerCalled := false
	handler := func(err error) {
		require.ErrorIs(t, someErr, err)
		handlerCalled = true
	}

	yeti.Snag(handler)
	yeti.Yeet(someErr)
	yeti.Handle()

	require.True(t, handlerCalled, "Error handler not called for error")
}

func TestYeti_AddYeti(t *testing.T) {
	t.Parallel()

	yeti := NewYeti()
	yeti2 := NewYeti()
	yeti.AddYeti(yeti2)

	require.Contains(t, yeti.yetis, yeti2, "Yeti not added to Yeti's yetties")
}

func TestYeti_AddYetiHandle(t *testing.T) {
	t.Parallel()

	yeti := NewYeti()
	yeti2 := NewYeti()
	yeti.AddYeti(yeti2)

	someErr := errors.New("some error")
	someErr2 := errors.New("some error 2")

	handlerCalled := 0
	handler := func(err error) {
		require.ErrorIs(t, someErr, err)
		handlerCalled++
	}
	handler2 := func(err error) {
		require.ErrorIs(t, someErr2, err)
		handlerCalled++
	}

	yeti.Snag(handler)
	yeti2.Snag(handler2)
	yeti.Yeet(someErr)
	yeti.Yeet(someErr)
	yeti2.Yeet(someErr2)
	yeti.Handle()

	require.Equal(t, 3, handlerCalled, "Error handler is not called 3 times on 2 yeti")
}

func TestYeti_HandleReentrant(t *testing.T) {
	t.Parallel()

	yeti := NewYeti()

	handlerCalls := 0
	yeti.Snag(func(err error) {
		handlerCalls++
		yeti.Yeet(errors.New("reentrant"))
		yeti.Snag(func(error) {})
	})
	yeti.Yeet(errors.New("some error"))

	done := make(chan struct{})
	go func() {
		yeti.Handle()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Yeti.Handle deadlocked: a handler re-entered the same Yeti")
	}

	require.Equal(t, 1, handlerCalls)
}

func TestYeti_HandleConcurrent(t *testing.T) {
	t.Parallel()

	yeti := NewYeti()
	yeti.Snag(func(error) {})
	for i := 0; i < 10; i++ {
		yeti.Yeet(errors.New("seed"))
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 5_000; i++ {
			yeti.Yeet(errors.New("concurrent"))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			yeti.Handle()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			yeti.Snag(func(error) {})
		}
	}()
	wg.Wait()
}

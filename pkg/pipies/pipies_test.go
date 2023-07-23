package pipies

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/internal/primitive/pointer"
)

type testT string

func Test_Predicates(t *testing.T) {
	t.Parallel()

	t.Run("NotNull", func(t *testing.T) {
		require.True(t, NotNull(pointer.To(1)))
		require.False(t, NotNull[int](nil))
		require.False(t, NotNull[int](nil))
		require.False(t, NotNull[testT](nil))
		var empty *testT
		require.False(t, NotNull(empty))
		empty = nil
		require.False(t, NotNull(empty))
		var e any = empty
		require.True(t, NotNull[any](&e))
	})
	t.Run("IsNull", func(t *testing.T) {
		require.False(t, IsNull(pointer.To(1)))
		require.True(t, IsNull[int](nil))
		require.True(t, IsNull[int](nil))
		require.True(t, IsNull[testT](nil))
		var empty *testT
		require.True(t, IsNull(empty))
		empty = nil
		require.True(t, IsNull(empty))
		var e any = empty
		require.False(t, IsNull[any](&e))
	})
	t.Run("NotZero", func(t *testing.T) {
		require.True(t, NotZero(pointer.To(1)))
		require.False(t, NotZero(pointer.To(0)))
		require.False(t, NotZero(pointer.To[float32](0.0)))
		require.False(t, NotZero(&struct{ a, b, c int }{}))
		require.True(t, NotZero(&struct{ a, b, c int }{1, 0, 0}))
	})
}

func Test_PredicateBuilders(t *testing.T) {
	t.Parallel()

	t.Run("Eq", func(t *testing.T) {
		eq5 := Eq(5)
		require.True(t, eq5(pointer.To(5)))
		require.False(t, eq5(pointer.To(4)))
		eqS := Eq("test")
		require.True(t, eqS(pointer.To("test")))
		require.False(t, eqS(pointer.To("sett")))
	})

	t.Run("NotEq", func(t *testing.T) {
		neq5 := NotEq(5)
		require.False(t, neq5(pointer.To(5)))
		require.True(t, neq5(pointer.To(4)))
		neqS := NotEq("test")
		require.False(t, neqS(pointer.To("test")))
		require.True(t, neqS(pointer.To("sett")))
	})

	t.Run("LessThan", func(t *testing.T) {
		lt5 := LessThan(5)
		require.True(t, lt5(pointer.To(4)))
		require.False(t, lt5(pointer.To(5)))
		require.False(t, lt5(pointer.To(6)))
		ltf5 := LessThan(5.0)
		require.True(t, ltf5(pointer.To(4.999)))
		require.False(t, ltf5(pointer.To(float64(int(5)))))
		require.False(t, ltf5(pointer.To(5.01)))
	})
}

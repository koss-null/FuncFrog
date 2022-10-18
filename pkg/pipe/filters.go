// filters.go is a collection of useful generic filter functions to use in Filter()
package pipe

import "golang.org/x/exp/constraints"

// NotNull filters only NotNull objects
func NotNull[T comparable](x T) bool {
	var zero T
	return x != zero
}

func Eq[T comparable](x T) func(T) bool {
	return func(y T) bool {
		return x == y
	}
}

func NotEq[T comparable](x T) func(T) bool {
	return func(y T) bool {
		return x != y
	}
}

func LessThan[T constraints.Ordered](x T) func(T) bool {
	return func(y T) bool {
		return y < x
	}
}

package pipe

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

// The list of filter functions to use in Filter().

// NotNull filters out values that are equal to nil.
// It uses reflection, so if you don't store any pointers, better use NotZero.
func NotNull[T any](x T) bool {
	return reflect.ValueOf(x).IsNil()
}

// NotZero filters out values wich are equal to a default zero value of type T.
func NotZero[T comparable](x T) bool {
	var zero T
	return x != zero
}

// Eq returns a predicate wich is true when the argument is equal to x.
func Eq[T comparable](x T) func(T) bool {
	return func(y T) bool {
		return x == y
	}
}

// NotEq returns a predicate wich is true when the argument is NOT equal to x.
func NotEq[T comparable](x T) func(T) bool {
	return func(y T) bool {
		return x != y
	}
}

// LessThan returns a predicate wich is true when the argument is less than x.
func LessThan[T constraints.Ordered](x T) func(T) bool {
	return func(y T) bool {
		return y < x
	}
}

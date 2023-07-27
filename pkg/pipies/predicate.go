package pipies

import (
	"reflect"

	"golang.org/x/exp/constraints"

	"github.com/koss-null/lambda/pkg/pipe"
)

// The list of filter functions to use in Filter().

// Predicates

// NotNil returns true is x underlying value is not nil.
// It uses reflection, so if you don't store any pointers, better use NotZero.
func NotNil[T any](x *T) bool {
	return x != nil && !reflect.ValueOf(x).IsNil()
}

// IsNil returns true if x underlying value is nil.
func IsNil[T any](x *T) bool {
	return x == nil || reflect.ValueOf(x).IsNil()
}

// NotZero returns true is x equals to a default zero value of the type T.
func NotZero[T comparable](x *T) bool {
	var zero T
	return *x != zero
}

// Predicate builders

// Eq returns a predicate wich is true when the argument is equal to x.
func Eq[T comparable](x T) pipe.Predicate[T] {
	return func(y *T) bool {
		return x == *y
	}
}

// NotEq returns a predicate wich is true when the argument is NOT equal to x.
func NotEq[T comparable](x T) pipe.Predicate[T] {
	return func(y *T) bool {
		return x != *y
	}
}

// LessThan returns a predicate wich is true when the argument is less than x.
func LessThan[T constraints.Ordered](x T) pipe.Predicate[T] {
	return func(y *T) bool {
		return *y < x
	}
}

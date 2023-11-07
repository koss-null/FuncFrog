// Function set to use with Filter
package pipies

import (
	"reflect"
	"sync"

	"golang.org/x/exp/constraints"

	"github.com/koss-null/funcfrog/pkg/pipe"
)

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

// Eq returns a predicate which is true when the argument is equal to x.
func Eq[T comparable](x T) pipe.Predicate[T] {
	return func(y *T) bool {
		return x == *y
	}
}

// NotEq returns a predicate which is true when the argument is NOT equal to x.
func NotEq[T comparable](x T) pipe.Predicate[T] {
	return func(y *T) bool {
		return x != *y
	}
}

// LessThan returns a predicate which is true when the argument is less than x.
func LessThan[T constraints.Ordered](x T) pipe.Predicate[T] {
	return func(y *T) bool {
		return *y < x
	}
}

// Distinct returns a predicate with filters out the same elements compated by the output of getKey function.
// getKey function should receive an argument of a Pipe value type.
// The result function is rather slow since it takes a lock on each element.
// You should use Pipe.Distinct() to get better performance.
func Distinct[T any, C comparable](getKey func(x *T) C) pipe.Predicate[T] {
	set := make(map[C]struct{})
	var mx sync.Mutex

	return func(y *T) bool {
		if y == nil {
			return false
		}
		key := getKey(y)

		mx.Lock()
		defer mx.Unlock()

		if _, ok := set[key]; ok {
			return false
		}
		set[key] = struct{}{}
		return true
	}
}

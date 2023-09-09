package pipe

// Accum is a standard type for reduce function.
// It is guaranteed that arguments are not nil, so you can dereference them with no check.
type Accum[T any] func(*T, *T) T

// Acc creates an Accum from a function of a different signature.
// Using it allows inner function `fn` not to use dereference,
// but also it leads to a larger memory allocation.
func Acc[T any](fn func(T, T) T) Accum[T] {
	return func(x, y *T) T {
		res := fn(*x, *y)
		return res
	}
}

// Predicate is a standard type for filtering.
// It is guaranteed that argument is not nil, so you can dereference it with no check.
type Predicate[T any] func(*T) bool

// Pred creates a Predicate from a function of a different signature.
// Using it allows inner function `fn` not to use dereference,
// but also it leads to a larger memory allocation.
func Pred[T any](fn func(T) bool) Predicate[T] {
	return func(x *T) bool {
		return fn(*x)
	}
}

// Comparator is a standard type for sorting comparisons.
// It is guaranteed that argument is not nil, so you can dereference it with no check.
type Comparator[T any] func(*T, *T) bool

// Comp creates a Comparator from a function of a different signature.
// Using it allows inner function `cmp` not to use dereference,
// but also it leads to a larger memory allocation.
func Comp[T any](cmp func(T, T) bool) Comparator[T] {
	return func(x, y *T) bool {
		return cmp(*x, *y)
	}
}

// Promice returns two values: the evaluated value and if it is not skipped.
// It should be checked as: if p, notSkipped := promice(); notSkipped { appendToAns(p) }
type Promice[T any] func() (T, bool)

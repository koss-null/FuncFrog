package pipe

import "golang.org/x/exp/constraints"

func Eq[T comparable](x, y T) bool {
	return x == y
}

func Less[T constraints.Ordered](x, y T) bool {
	return x < y
}

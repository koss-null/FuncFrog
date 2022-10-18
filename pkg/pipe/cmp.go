package pipe

import "golang.org/x/exp/constraints"

func Less[T constraints.Ordered](x, y T) bool {
	return x < y
}

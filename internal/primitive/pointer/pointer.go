package pointer

import "golang.org/x/exp/constraints"

func To[T constraints.Ordered](x T) *T {
	return &x
}

func From[T constraints.Ordered](x *T) (res T) {
	if x == nil {
		return
	}
	return *x
}

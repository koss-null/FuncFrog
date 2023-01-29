package piper

import (
	"math"

	"golang.org/x/exp/constraints"
)

func divUp(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}

func min[T constraints.Ordered](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func max[T constraints.Ordered](a, b T) T {
	if a < b {
		return b
	}
	return a
}

package batch

import "math"

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Do[T any](a []T, batchSize int) [][]T {
	batchCnt := int(math.Ceil(float64(len(a)) / float64(batchSize)))
	res := make([][]T, batchCnt)
	j := 0
	for i := 0; i < len(a); i += batchSize {
		res[j] = a[i:min(i+batchSize, len(a))]
		j++
	}
	return res
}

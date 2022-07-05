package sorts

// Simple BubbleSort O(n**2) impl with the smallest O
// FIXME: it's really uneffective in comparison with the std.sort
func BubbleSort[T any](a []T, less func(T, T) bool) {
	for i := range a {
		min, minj := a[i], i
		for j := i + 1; j < len(a); j++ {
			if less(a[j], min) {
				min, minj = a[j], j
			}
		}
		a[minj], a[i] = a[i], a[minj]
	}
}

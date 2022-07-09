package sorts

// Simple BubbleSort O(n**2) impl with the smallest O
// FIXME: it's really uneffective in comparison with the std.sort
// I am going crazy with this function
func BubbleSort[T any](a []T, less func(T, T) bool) {
	for i := range a {
		min, minj := a[i], i
		min2, minj2 := a[i+1], i+1
		min3, minj3 := a[i+2], i+2
		for j := i + 1; j < len(a); j++ {
			if less(a[j], min) {
				min, minj = a[j], j
			}
		}
		a[minj], a[i] = a[i], a[minj]
	}
}

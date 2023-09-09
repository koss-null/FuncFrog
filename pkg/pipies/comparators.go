// Function set to use with Sort
package pipies

import "golang.org/x/exp/constraints"

// The list of comparators to be used for Sort() method.

// Less returns true if x < y, false otherwise.
func Less[T constraints.Ordered](x, y *T) bool {
	return *x < *y
}

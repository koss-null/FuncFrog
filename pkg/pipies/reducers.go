// Function set to use with Reduce
package pipies

import (
	cns "golang.org/x/exp/constraints"
)

// Sum is a reduce function for summing up types that support "+".
func Sum[T cns.Float | cns.Integer | cns.Complex | ~string](a, b *T) T {
	res := *a + *b
	return res
}

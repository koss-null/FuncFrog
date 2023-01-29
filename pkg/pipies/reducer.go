package pipies

import cns "golang.org/x/exp/constraints"

// Sum is a reduce function for summing up types that support "+".
func Sum[T cns.Float | cns.Integer | cns.Complex](a, b T) T {
	return a + b
}

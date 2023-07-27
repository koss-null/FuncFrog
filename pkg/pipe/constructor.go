package pipe

import (
	"golang.org/x/exp/constraints"

	"github.com/koss-null/lambda/internal/internalpipe"
)

// Slice creates a Pipe from a slice
func Slice[T any](dt []T) Piper[T] {
	return &Pipe[T]{internalpipe.Slice(dt)}
}

// Func creates a lazy sequence by applying the provided function 'fn'
// to each index 'i', placing the resulting object into the sequence at index 'i'.
//
// The 'fn' function must return an object of type T, along with a boolean indicating
// if it exists.
//
// Use the 'Take' or 'Gen' functions to set the number of output values to generate,
// or use the 'Until' function to enforce a limit based on a predicate function.
func Func[T any](fn func(i int) (T, bool)) PiperNoLen[T] {
	return &PipeNL[T]{internalpipe.Func(fn)}
}

// Fu creates a lazy sequence by applying the provided function 'fn' to each index 'i',
// and setting the result to the sequence at index 'i'.
//
// 'fn' is a shortened version of the 'Func' function, where the second argument is true by default.
//
// Use the 'Take' or 'Gen' functions to set the number of output values to generate,
// or use the 'Until' function to enforce a limit based on a predicate function.
func Fn[T any](fn func(i int) T) PiperNoLen[T] {
	return Func(func(i int) (T, bool) {
		return fn(i), true
	})
}

// FuncP is the same as Func but allows to return pointers to the values.
// It creates a lazy sequence by applying the provided function 'fn' to each index 'i', and
// setting the result to the sequence at index 'i' as a pointer to the object of type 'T'.
//
// 'fn' should return a pointer to an object of type 'T', along with a boolean indicating
// whether the object exists.
//
// Use the 'Take' or 'Gen' functions to set the number of output values to generate,
// or use the 'Until' function to enforce a limit based on a predicate function.
func FuncP[T any](fn func(i int) (*T, bool)) PiperNoLen[T] {
	return &PipeNL[T]{internalpipe.FuncP(fn)}
}

// Cycle creates a lazy sequence that cycles through the elements of the provided slice 'a'.
// To use the resulting sequence, the 'Cycle' function returns a 'PiperNoLen' object.
//
// Use the 'Take' or 'Gen' functions to set the number of output values to generate,
// or use the 'Until' function to enforce a limit based on a predicate function.
func Cycle[T any](a []T) PiperNoLen[T] {
	return Fn(func(i int) T {
		return a[i%len(a)]
	})
}

// Range creates a lazy sequence of type 'T', which consists of values
// starting from 'start', incrementing by 'step' and stopping just before 'finish', i.e. [start..finish).
// The type 'T' can be either an integer or a float.
func Range[T constraints.Integer | constraints.Float](start, finish, step T) Piper[T] {
	return &Pipe[T]{internalpipe.Range(start, finish, step)}
}

package pipies

// This madness is the only way to implement Not function for a function with arbitary amount of arguments.
// The amount of t at the end of Not(t) function equals to the amount of arguments the initial function takes

// Not returns a new function that negates the result of the input function.
func Not[T any](fn func(x T) bool) func(T) bool {
	return func(x T) bool {
		return !fn(x)
	}
}

// Nott returns a new function that negates the result of the input function.
func Nott[T1, T2 any](fn func(x T1, y T2) bool) func(T1, T2) bool {
	return func(x T1, y T2) bool {
		return !fn(x, y)
	}
}

// Nottt returns a new function that negates the result of the input function.
func Nottt[T1, T2, T3 any](fn func(x T1, y T2, z T3) bool) func(T1, T2, T3) bool {
	return func(x T1, y T2, z T3) bool {
		return !fn(x, y, z)
	}
}

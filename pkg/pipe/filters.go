// filters.go is a collection of useful generic filter functions to use in Filter()
package pipe

// NotNull filters only NotNull objects
func NotNull[T comparable](x T) bool {
	var zero T
	return x != zero
}

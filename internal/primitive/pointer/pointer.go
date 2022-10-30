package pointer

func To[T comparable](x T) *T {
	return &x
}

func From[T comparable](x *T) (res T) {
	if x == nil {
		return
	}
	return *x
}

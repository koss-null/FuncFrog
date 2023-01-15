package pointer

func To[T any](x T) *T {
	return &x
}

func From[T any](x *T) (res T) {
	if x == nil {
		return
	}
	return *x
}

package pointer

func Ref[T any](x T) *T {
	return &x
}

func Deref[T any](x *T) (res T) {
	if x == nil {
		return
	}
	return *x
}

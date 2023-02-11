package internalpipe

type AccumFn[T any] func(*T, *T) *T

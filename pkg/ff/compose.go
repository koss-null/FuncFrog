package ff

func Compose[T1, T2, T3 any](fn1 func(T1) T2, fn2 func(T2) T3) func(T1) T3 {
	return func(x T1) T3 {
		return fn2(fn1(x))
	}
}

// FIXME: signarure about to be changed
func ComposeErr[T1, T2, T3 any](fn1 func(T1) (T2, error), fn2 func(T2) (T3, error), errHandler func(error)) func(T1) T3 {
	return func(x T1) T3 {
		res, err := fn1(x)
		if err != nil {
			errHandler(err)
		}
		res2, err := fn2(res)
		if err != nil {
			errHandler(err)
		}
		return res2
	}
}

package pipe

// Piper interface contains all methods of a pipe with determened length.
type Piper[T any] interface {
	doer[T]

	mapper[T, Piper[T]]
	filterer[T, Piper[T]]
	sorter[T, Piper[T]]

	paralleller[T, Piper[T]]

	firster[T]
	anier[T]
	reducer[T]
	summer[T]
	counter

	promicer[T]
	eraser[Piper[any]]
	snagger[Piper[T]]
}

// PiperNoLen represents methods available to a Pipe type with no length determened.
type PiperNoLen[T any] interface {
	taker[Piper[T]]
	genner[Piper[T]]

	mapper[T, PiperNoLen[T]]
	filterer[T, PiperNoLen[T]]

	paralleller[T, PiperNoLen[T]]

	firster[T]
	anier[T]

	eraser[PiperNoLen[any]]
	snagger[PiperNoLen[T]]
}

type paralleller[T, PiperT any] interface {
	Parallel(uint16) PiperT
}

type mapper[T, PiperT any] interface {
	Map(func(T) T) PiperT
}

type filterer[T, PiperT any] interface {
	Filter(Predicate[T]) PiperT
}

type sorter[T, PiperT any] interface {
	Sort(Comparator[T]) PiperT
}

type reducer[T any] interface {
	Reduce(Accum[T]) *T
}

type summer[T any] interface {
	Sum(Accum[T]) T
}

type taker[T any] interface {
	Take(int) T
}

type genner[T any] interface {
	Gen(int) T
}

type doer[T any] interface {
	Do() []T
}

type firster[T any] interface {
	First() *T
}

type anier[T any] interface {
	Any() *T
}

type counter interface {
	Count() int
}

type eraser[PiperT any] interface {
	Erase() PiperT
}

type promicer[T any] interface {
	Promices() []func() (T, bool)
}

type snagger[PiperT any] interface {
	Snag(func(error)) PiperT
}

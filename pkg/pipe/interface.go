package pipe

// Piper interface contains all methods of a pipe with determened length.
type Piper[T any] interface {
	changer[T]
	getter[T]
	configger[T, Piper[T]]
}

// PiperNoLen represents methods available to a Pipe type with no length determened.
type PiperNoLen[T any] interface {
	taker[Piper[T]]
	genner[Piper[T]]
	mapper[T, PiperNoLen[T]]
	filterer[T, PiperNoLen[T]]
}

type configger[T, PiperT any] interface {
	Parallel(uint16) PiperT
}

type changer[T any] interface {
	mapper[T, Piper[T]]
	filterer[T, Piper[T]]
	sorter[T, Piper[T]]
}

type getter[T any] interface {
	reducer[T]
	summer[T]
	doer[T]
	firster[T]
	anier[T]
	counter
}

type mapper[T, PiperT any] interface {
	Map(func(T) T) PiperT
}

type filterer[T, PiperT any] interface {
	Filter(func(T) bool) PiperT
}

type sorter[T, PiperT any] interface {
	Sort(func(T, T) bool) PiperT
}

type reducer[T any] interface {
	Reduce(func(T, T) T) *T
}

type summer[T any] interface {
	Sum(func(T, T) T) *T
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

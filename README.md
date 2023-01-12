# FuncFrog

[![Go Report Card](https://goreportcard.com/badge/github.com/koss-null/lambda)](https://goreportcard.com/report/github.com/koss-null/lambda)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

![FuncFrog icon](https://github.com/koss-null/lambda/blob/0.9.0/FuncFrogIco.jpg?raw=true)

FuncFrog is a library for performing parallel, lazy `map`, `reduce`, and `filter` operations on slices in one pipeline. The slice can be set by a generating function, and parallel execution is supported. It is expected that all function arguments will be **pure functions** (functions with no side effects that can be cached by their arguments). It is capable of handling large amounts of data with minimal overhead, and its parallel execution capabilities allow for even faster processing times. Additionally, the library is easy to use and has a clean, intuitive API. [Here](https://macias.info/entry/202212020000_go_streams.md) is some performance review.

# Repository Renamed

This repository has been renamed from github.com/koss-null/lambda to github.com/koss-null/funcfrog.
Please update any local clones or remotes to reflect this change. 

To update a local clone of the repository, you can use the following command:
```bash
$ git remote set-url origin https://github.com/koss-null/funcfrog
```
This command will update the URL of the "origin" reomte to the new repository URL. 

To update the import path of the repository in go code, you can use the following import statement:
```go
import "github.com/koss-null/funcfrog"
```
then in your project use:
```bash
go get github.com/koss-null/funcfrog
```

## Getting Started

To install FuncFrog, run the following command:

```
go get github.com/koss-null/funcfrog
```

Then, import the library into your Go code (basically you need the pipe package):

```go
import "github.com/koss-null/funcfrog/pkg/pipe"
```

You can then use the `pipe` package to create a pipeline of operations on a slice:
```go
res := pipe.Slice(a).
    Map(func(x int) int { return x * x }).
    Map(func(x int) int { return x + 1 }).
    Filter(func(x int) bool { return x > 100 }).
    Filter(func(x int) bool { return x < 1000 }).
    Parallel(12).
    Do()
```

To see more examples of how to use FuncFrog, check out the `examples/main.go` file. You can also run this file with `go run examples/main.go`.

## Basic information

The `Pipe` type is an interface that represents a lazy, potentially infinite sequence of data. The `Pipe` interface provides a set of methods that can be used to transform and filter the data in the sequence.

The following functions can be used to create a new `Pipe`:
- :frog: `Slice([]T) *Pipe`: creates a `Pipe` of a given type `T` from a slice.
- :frog: `Func(func(i int) (T, bool)) *Pipe`: creates a `Pipe` of type `T` from a function. The function should return the value of the element at the `i`th position in the `Pipe`, as well as a boolean indicating whether the element should be included (`true`) or skipped (`false`).
- :frog: `Take(n int) *Pipe`: if it's a `Func`-made `Pipe`, expects `n` values to be eventually returned.
- :frog: `Gen(n int) *Pipe`: if it's a `Func`-made `Pipe`, generates a sequence from `[0, n)` and applies the function to it.
- :frog: `Copy() *Pipe`: returns a copy of a current `Pipe` with the copyied underlying array.
- :seedling: TBD: `Cycle(data []T) *Pipe`: creates a new `Pipe` that cycles through the elements of the provided slice indefinitely.
- :seedling: TBD: `Range(start, end, step T) *Pipe`: creates a new `Pipe` that generates a sequence of values of type `T` from `start` to `end` (exclusive) with a fixed `step` value between each element. `T` can be any numeric type, such as `int`, `float32`, or `float64`.

The following functions can be used to transform and filter the data in the `Pipe`:
- :frog: `Map(fn func(x T) T) *Pipe`: applies the function `fn` to every element of the `Pipe` and returns a new `Pipe` with the transformed data.
- :frog: `Filter(fn func(x T) bool) *Pipe`: applies the predicate function `fn` to every element of the `Pipe` and returns a new `Pipe` with only the elements that satisfy the predicate.
- :frog: `Reduce(fn func(x, y T) T) T`: applies the binary function `fn` to the elements of the `Pipe` and returns a single value that is the result of the reduction.
- :frog: `Sum(sum func(x, y) T) T`: makes parallel reduce with associative function `sum`.
- :frog: `Sort(less func(x, y T) bool) *Pipe`: sorts the elements of the `Pipe` using the provided `less` function as the comparison function.

The following functions can be used to retrieve a single element or perform a boolean check on the `Pipe` without executing the entire pipeline:
- :frog: `Any(fn func(x T) bool) bool`: returns `true` if any element of the `Pipe` satisfies the predicate `fn`, and `false` otherwise.
- :frog: `First() T`: returns the first element of the `Pipe`, or `nil` if the `Pipe` is empty.
- :frog: `Count() int`: returns the number of elements in the `Pipe`. It does not execute the entire pipeline, but instead simply returns the number of elements in the `Pipe`.
- :seedling: TBD: `IsAny() bool`: returns `true` if the `Pipe` contains any elements, and `false` otherwise.
- :seedling: TBD: `MoreThan(n int) bool`: returns `true` if the `Pipe` contains more than `n` elements, and `false` otherwise.

The :frog: `Parallel(n int) *Pipe` function can be used to specify the level of parallelism in the pipeline, by setting the number of goroutines to be executed on (4 by default).

Finally, the :frog: `Do() []T` function is used to execute the pipeline and return the resulting slice of data. This function should be called at the end of the pipeline to retrieve the final result.

In addition to the functions described above, the `pipe` package also provides several utility functions that can be used to create common types of `Pipe`s, such as `Range`, `Repeat`, and `Cycle`. These functions can be useful for creating `Pipe`s of data that follow a certain pattern or sequence.```

## Examples

### Basic example:

```go
res := pipe.Slice(a).
	Map(func(x int) int { return x * x }).
	Map(func(x int) int { return x + 1 }).
	Filter(func(x int) bool { return x > 100 }).
	Filter(func(x int) bool { return x < 1000 }).
	Parallel(12).
	Do()
```

### Example using `Func` and `Take`:

```go
p := pipe.Func(func(i int) (int, bool) {
	if i < 10 {
		return i * i, true
	}
	return 0, false
}).Take(5).Do()
// p will be [0, 1, 4, 9, 16]
```

### Example using `Filter` and `Map`:

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
	Filter(func(x int) bool { return x % 2 == 0 }).
	Map(func(x int) string { return strconv.Itoa(x) }).
	Do()
// p will be ["2", "4", "6", "8", "10"]
```

### Example using `Map` and `Reduce` :

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5}).
	Map(func(x int) int { return x * x }).
	Reduce(func(x, y int) string { 
		return strconv.Itoa(x) + "-" + strconv.Itoa(y) 
	})
// p will be "1-4-9-16-25"
```

### Example of `Map` and `Reduce` with the underlying array type change:

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
strP := pipe.Map(p, func(x int) string { return strconv.Itoa(x) })
result := pipe.Reduce(strP, func(x, y string) int { return len(x) + len(y) }).Do()
// result will be 45
```

### Example using `Sort`:

```go
p := pipe.Func(func(i int) (float32, bool) {
	return float32(i) * 0.9, true
}).
	Map(func(x float32) float32 { return x * x }).
	Gen(100500).
	Sort(pipe.Less[float32]).
	Parallel(12).
	Do()
// p will contain the elements of p sorted in ascending order
```

### Example of infine sequence generation:

Here is an example of generating an infinite sequence of random `float32` values greater than `0.5`:

```go
p := pipe.Func(func(i int) (float32, bool) {
	rnd := rand.New(rand.NewSource(42))
	rnd.Seed(int64(i))
	return rnd.Float32(), true
}).
	Filter(func(x float32) bool { return x > 0.5 })
```

To generate a specific number of values, you can use the `Take` method:

```go
p = p.Take(65000)
```

To accumulate the elements of the `Pipe`, you can use the `Reduce` method:

```go
sum := p.Sum(pipe.Sum[float32])
//also you can: sum := p.Reduce(func(x, y float32) float32 { return x + y})
// sum will be the sum of the first 65000 random float32 values greater than 0.5
```

### Example using `Range` (not implemented yet) and `Map`:

```go
p := pipe.Range(10, 20, 2).Map(func(x int) int { return x * x }).Do()
// p will be [100, 144, 196, 256, 324]
```

### Example using `Repeat` (not implemented yet) and `Map`:

```go
p := pipe.Repeat("hello", 5).Map(func(s string) int { return len(s) }).Do()
// p will be [5, 5, 5, 5, 5]
```

### Example using `Cycle` (not implemented yet) and `Filter`:

```go
p := pipe.Cycle([]int{1, 2, 3}).Filter(func(x int) bool { return x % 2 == 0 }).Take(4).Do()
// p will be [2, 2, 2, 2]
```

## Is this package stable?

Yes it funally is since v0.9.0! All listed functionality is fully covered by unit-tests (`pkg/pipe/test/coverage_test.go`). Functionality marked as TBD
will be implemented as it described in the README and supplied covered by unit-tests to be delivered stable. 
If there will be any method signature changes, the major version will be incremented. 

## Contributions

I will accept any pr's with the functionality marked as TBD. 
Also I will accept any sane unit-tests. 
Bugfixes. 
You are welcome to create any issues. 

## What's next?

I hope to provide some roadmap of the project soon. 
Feel free to fork, inspire and use! 

## Supported functions list

- :frog: `Slice([]T) *Pipe`: creates a `Pipe` of a given type `T` from a slice.
- :frog: `Func(func(i int) (T, bool)) *Pipe`: creates a `Pipe` of type `T` from a function. The function should return the value of the element at the `i`th position in the `Pipe`, as well as a boolean indicating whether the element should be included (`true`) or skipped (`false`). This function can be used to generate elements on demand, rather than creating a slice beforehand.
- :frog: `Take(n int) *Pipe`: if it's a `Func`-made `Pipe`, expects `n` values to be eventually returned. This function can be used to limit the number of elements generated by the function.
- :frog: `Gen(n int) *Pipe`: if it's a `Func`-made `Pipe`, generates a sequence from `[0, n)` and applies the function to it. This function can be used to generate a predetermined number of elements using the function.
- :frog: `Parallel(n int) *Pipe`: sets the number of goroutines to be executed on (1 by default). This function can be used to specify the level of parallelism in the pipeline.
- :frog: `Map(fn func(x T) T) *Pipe`: applies the function `fn` to every element of the `Pipe` and returns a new `Pipe` with the transformed data. This function can be used to apply a transformation to each element in the `Pipe`.
- :frog: `Filter(fn func(x T) bool) *Pipe`: applies the predicate function `fn` to every element of the `Pipe` and returns a new `Pipe` with only the elements that satisfy the predicate. This function can be used to select a subset of elements from the `Pipe`.
- :frog: `Reduce(fn func(x, y T) T) T`: applies the binary function `fn` to the elements of the `Pipe` and returns a single value that is the result of the reduction. This function can be used to combine the elements of the `Pipe` into a single value.
- :frog: `Sum(sum func(x, y) T) T`: makes parallel reduce with associative function `sum`.
- :frog: `Do() []T`: executes the `Pipe` and returns the resulting slice of data.
- :frog: `First() T`: returns the first element of the `Pipe`, or `nil` if the `Pipe` is empty. This function can be used to retrieve the first element of the `Pipe` without executing the entire pipeline.
- :frog: `Any(fn func(x T) bool) bool`: returns `true` if any element of the `Pipe` satisfies the predicate `fn`, and `false` otherwise. This function can be used to check if any element in the `Pipe` satisfies a given condition.
- :frog: `Count() int`: returns the number of elements in the `Pipe`. It does not execute the entire pipeline, but instead simply returns the number of elements in the `Pipe`.
- :frog: `Sort(less func(x, y T) bool) *Pipe`: sorts the elements of the `Pipe` using the provided `less` function as the comparison function
- :seedling: TBD: `IsAny() bool`: returns `true` if the `Pipe` contains any elements, and `false` otherwise. This function can be used to check if the `Pipe` is empty.
- :seedling: TBD: `MoreThan(n int) bool`: returns `true` if the `Pipe` contains more than `n` elements, and `false` otherwise.
- :seedling: TBD: `Reverse() *Pipe`: reverses the underlying slice.

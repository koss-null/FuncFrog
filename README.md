# FuncFrog

[![Go Report Card](https://goreportcard.com/badge/github.com/koss-null/funcfrog)](https://goreportcard.com/report/github.com/koss-null/funcfrog)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://raw.githubusercontent.com/koss-null/funcfrog/master/coverage_badge.png?raw=true)](coverage)

![FuncFrog icon](https://github.com/koss-null/funcfrog/blob/master/FuncFrogIco.jpg?raw=true)

*FuncFrog* is a library for performing **efficient**, **parallel**, **lazy** `map`, `reduce`, `filter` and [many other](#supported-functions-list) operations on slices and other data sequences in a pipeline. The sequence can be set by a variety of [generating functions](#constructors). Everything is supported to be executed in parallel with **minimal overhead** on copying and locks. There is a built-in support of [error handling](#error-handling) with Yeet/Snag methods  
The library is easy to use and has a clean, intuitive API.  
You can measure performance comparing to vanilla `for` loop on your machine using `cd perf/; make` (spoiler: FuncFrog
is better when multithreading).  

## Table of Contents
- [Getting Started](#getting-started)
- [Basic information](#basic-information)
- [Supported functions list](#supported-functions-list)
  - [Constructors](#constructors)
  - [Set Pipe length](#set-pipe-length)
  - [Split evaluation into *n* goroutines](#split-evaluation-into-n-goroutines)
  - [Transform data](#transform-data)
  - [Retrieve a single element or perform a boolean check](#retrieve-a-single-element-or-perform-a-boolean-check)
  -  [Evaluate the pipeline](#evaluate-the-pipeline)
  - [Transform Pipe *from one type to another*](#transform-pipe-from-one-type-to-another)
  - [Easy type conversion for Pipe[any]]( #easy-type-conversion-for-pipe[any])
  - [Error handling](#error-handling)
  - [To be done](#to-be-done)
- [Using prefix `Pipe` to transform `Pipe` type](#using-prefix-pipe-to-transform-pipe-type)
- [Using `ff` package to write shortened pipes](#using-ff-package-to-write-shortened-pipes)
- [Look for useful functions in `Pipies` package](#look-for-useful-functions-in-pipies-package)
- [Examples](#examples)
  - [Basic example](#basic-example)
  - [Example using `Func` and `Take`](#example-using-func-and-take)
  - [Example using `Func` and `Gen`](#example-using-func-and-gen)
  - [Example difference between `Take` and `Gen`](#example-difference-between-take-and-gen)
  - [Example using `Filter` and `Map`](#example-using-filter-and-map)
  - [Example using `Map` and `Reduce`](#example-using-map-and-reduce)
  - [Example of `Map` and `Reduce` with the underlying array type change](#example-of-map-and-reduce-with-the-underlying-array-type-change)
  - [Example using `Sort`](#example-using-sort)
  - [Example of infine sequence generation](#example-of-infine-sequence-generation)
  - [Example using `Range` and `Map`](#example-using-range-not-implemented-yet-and-map)
  - [Example using `Repeat` and `Map`](#example-using-repeat-not-implemented-yet-and-map)
  - [Example using `Cycle` and `Filter`](#example-using-cycle-not-implemented-yet-and-filter)
  - [Example using `Erase` and `Collect`](#example-using-erase-and-collect)
  - [Example of simple error handling](#example-of-simple-error-handling)
  - [Example of multiple error handling](example-of-multiple-error-handling)
- [Is this package stable?](#is-this-package-stable)
- [Contributions](#contributions)
- [What's next?](#whats-next)

## Getting Started

To use FuncFrog in your project, run the following command:

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
    Filter(func(x *int) bool { return *x > 100 }).
    Filter(func(x *int) bool { return *x < 1000 }).
    Parallel(12).
    Do()
```

All operations are carefully fenced with interfaces, so feel free to use anything, autosuggestion suggests you.

To see some code snippets, check out the `examples/main.go` file. You can also run it with `go run examples/main.go`.

## Basic information

The `Piper` (or `PiperNoLen` for pipes with undetermined lengths) is an interface that represents a *lazy-evaluated sequence of data*. The `Piper` interface provides a set of methods that can be used to transform, filter, collect and analyze data in the sequence. 
Every pipe can be conveniently copied at every moment just by equating it to a variable.  Some methods (as `Take` or `Gen`) lead from `PiperNoLen` to `Piper` interface making wider method range available. 

## Supported functions list

The following functions can be used to create a new `Pipe` (this is how I call the inner representation of a sequence ofelements and a sequence operations on them): 
#### Constructors
- :frog: `Slice([]T) Piper`: creates a `Pipe` of a given type `T` from a slice, *the length is known*.
- :frog: `Func(func(i int) (T, bool)) PiperNL`: creates a `Pipe` of type `T` from a function. The function returns an element which is considered to be at `i`th position in the `Pipe`, as well as a boolean indicating whether the element should be included (`true`) or skipped (`false`), *the length is unknown*.
- :frog: `Fn(func(i int) (T)) PiperNL`: creates a `Pipe` of type `T` from a function. The function should return the value of the element at the `i`th position in the `Pipe`; to be able to skip values use `Func`.
- :frog: `FuncP(func(i int) (*T, bool)) PiperNL`: creates a `Pipe` of type `T` from a function. The function returns a pointer to an element which is considered to be at `i`th position in the `Pipe`, as well as a boolean indicating whether the element should be included (`true`) or skipped (`false`), *the length is unknown*.
- :frog: `Cycle(data []T) PiperNL`: creates a new `Pipe` that cycles through the elements of the provided slice indefinitely. *The length is unknown.*
- :frog: `Range(start, end, step T) Piper`: creates a new `Pipe` that generates a sequence of values of type `T` from `start` to `end` (exclusive) with a fixed `step` value between each element. `T` can be any numeric type, such as `int`, `float32`, or `float64`. *The length is known.*
- :frog: `Repeat(x T, n int) Piper`: creates a new `Pipe` that generates a sequence of values of type `T` and value x with the length of n. *The length is known.*

#### Set Pipe length
- :frog: `Take(n int) Piper`: if it's a `Func`-made `Pipe`, expects `n` values to be eventually returned. *Transforms unknown length to known.*
- :frog: `Gen(n int) Piper`: if it's a `Func`-made `Pipe`, generates a sequence from `[0, n)` and applies the function to it. *Transforms unknown length to known.*


#### Split evaluation into *n* goroutines
- :frog: `Parallel(n int) Pipe`: sets the number of goroutines to be executed on (1 by default). This function can be used to specify the level of parallelism in the pipeline. *Availabble for unknown length.*

#### Transform data
- :frog: `Map(fn func(x T) T) Pipe`: applies the function `fn` to every element of the `Pipe` and returns a new `Pipe` with the transformed data. *Available for unknown length.*
- :frog: `Filter(fn func(x *T) bool) Pipe`: applies the predicate function `fn` to every element of the `Pipe` and returns a new `Pipe` with only the elements that satisfy the predicate. *Available for unknown length.*
- :frog: `MapFilter(fn func(T) (T, bool)) Piper[T]`: applies given function to each element of the underlying slice. If the second returning value of `fn` is *false*, the element is skipped (may be **useful for error handling**).
- :frog: `Reduce(fn func(x, y *T) T) *T`: applies the binary function `fn` to the elements of the `Pipe` and returns a single value that is the result of the reduction. Returns `nil` if the `Pipe` was empty before reduction.
- :frog: `Sum(plus func(x, y *T) T) T`: makes parallel reduce with associative function `plus`.
- :frog: `Sort(less func(x, y *T) bool) Pipe`: sorts the elements of the `Pipe` using the provided `less` function as the comparison function.

#### Retrieve a single element or perform a boolean check
- :frog: `Any() T`: returns a random element existing in the pipe. *Available for unknown length.*
- :frog: `First() T`: returns the first element of the `Pipe`, or `nil` if the `Pipe` is empty. *Available for unknown length.*
- :frog: `Count() int`: returns the number of elements in the `Pipe`. It does not allocate memory for the elements, but instead simply returns the number of elements in the `Pipe`.

#### Evaluate the pipeline
- :frog: `Do() []T` function is used to **execute** the pipeline and **return the resulting slice of data**. This function should be called at the end of the pipeline to retrieve the final result.

#### Transform Pipe *from one type to another*
- :frog: `Erase() Pipe[any]`: returns a pipe where all objects are the objects from the initial `Pipe` but with erased type. Basically for each `x` it returns `any(&x)`. Use `pipe.Collect[T](Piper[any]) PiperT` to collect it back into some type (or `pipe.CollectNL` for slices with length not set yet).

#### Easy type conversion for Pipe[any]
- :frog: `pipe.Collect[T](Piper[any]) PiperNoLen[T]`
- :frog: `pipe.CollectNL[T](PiperNoLen[any]) PiperNoLen[T]`
This functions takes a Pipe of erased `interface{}` type (which is pretty useful if you have a lot of type conversions along your pipeline and can be achieved by calling `Erase()` on a `Pipe`). Basically, for each element `x` in a sequence `Collect` returns `*(x.(*T))` element.

#### Error handling
- :frog:  `Yeti(yeti) Pipe[T]`:set a `yeti` - an object that will collect errors thrown with `yeti.Yeet(error)`  and will be used to handle them.
- :frog: `Snag(func(error)) Pipe[T]`: set a function that will handle all errors which have been sent with `yeti.Yeet(error)` to the **last** `yeti` object that was set through `Pipe[T].Yeti(yeti) Pipe[T]` method. 
Error handling may look pretty uncommon at a first glance. To get better intuition about it you may like to check out [examples](#example-of-simple-error-handling) section.

#### To be done
- :seedling: TBD: `Until(fn func(*T) bool)`: if it's a `Func`-made `Pipe`, it evaluates one-by-one until fn return false. *This feature may require some new `Pipe` interfaces, since it is applicable only in a context of a single thread*
- :seedling: *TBD*: `IsAny() bool`: returns `true` if the `Pipe` contains any elements, and `false` otherwise. *Available for unknown length.*
- :seedling: *TBD*: `MoreThan(n int) bool`: returns `true` if the `Pipe` contains more than `n` elements, and `false` otherwise. *Available for unknown length.*
- :seedling: *TBD*: `Reverse() *Pipe`: reverses the underlying slice.

In addition to the functions described above, the `pipe` package also provides several utility functions that can be used to create common types of `Pipe`s, such as `Range`, `Repeat`, and `Cycle`. These functions can be useful for creating `Pipe`s of data that follow a certain pattern or sequence.

Also it is highly recommended to get familiarize with the `pipies` package, containing some useful *predecates*, *comparators* and *accumulators*.

### Using prefix `Pipe` to transform `Pipe` type

You may found that using `Erase()` is not so convenient as it makes you to do some pointer conversions. Fortunately there is another way to convert a pipe type: use functions from `pipe/prefixpipe.go`. These functions takes `Piper` or `PiperNoLen` as a first parameter and function to apply as the second and returns a resulting pipe (or the result itself) of a destination type.

#### Prefix pipe functinos

- :frog: `pipe.Map(Piper[SrcT], func(x SrcT) DstT) Piper[DstT] ` - applies *map* from one type to another for the `Pipe` with **known** length.
- :frog: `pipe.MapNL(PiperNoLen[SrcT], func(x SrcT) DstT) PiperNoLen[DstT] ` - applies *map* from one type to another for the `Pipe` with **unknown** length.
- :frog: `Reduce(Piper[SrcT], func(*DstT, *SrcT) DstT, initVal ...DstT)` - applies *reduce* operation on `Pipe` of type `SrcT` and returns result of type `DstT`. `initVal` is an optional parameter to **initialize** a value that should be used on the **first steps** of reduce.

### Using `ff` package to write shortened pipes

Sometimes you need just to apply a function. Creating a pipe using `pipe.Slice` and then call `Map` looks a little bit verbose, especially when you need to call `Map` or `Reduce` from one type to another. The solution for it is `funcfrog/pkg/ff` package. It contains shortened `Map` and `Reduce` functions which can be called directly with a slice as a first parameter.

- :frog: `Map([]SrcT, func(SrcT) DstT) pipe.Piper[DstT]` - applies sent function to a slice, returns a `Pipe` of resulting type
- :frog: `Reduce([]SrcT, func(*DstT, *SrcT) DstT, initVal ...DstT) DstT` - applies *reduce* operation on a slice and returns the result of type `DstT`. `initVal` is an optional parameter to **initialize** a value that should be used on the **first steps** of reduce.

### Look for useful functions in `Pipies` package

Some of the functions that are sent to `Map`, `Filter` or `Reduce` (or other `Pipe` methods) are pretty common. Also there is a common comparator for any integers and floats for a `Sort` method. 

## Examples

### Basic example:

```go
res := pipe.Slice(a).
	Map(func(x int) int { return x * x }).
	Map(func(x int) int { return x + 1 }).
	Filter(func(x *int) bool { return *x > 100 }).
	Filter(func(x *int) bool { return *x < 1000 }).
	Parallel(12).
	Do()
```

### Example using `Func` and `Take`:

```go
p := pipe.Func(func(i int) (v int, b bool) {
	if i < 10 {
		return i * i, true
	}; return
}).Take(5).Do()
// p will be [0, 1, 4, 9, 16]
```

### Example using `Func` and `Gen`:

```go
p := pipe.Func(func(i int) (v int, b bool) {
	if i < 10 {
		return i * i, true
	}; return
}).Gen(5).Do()
// p will be [0, 1, 4, 9, 16]
```

### Example difference between `Take` and `Gen`:

Gen(n) generates the sequence of n elements and applies all pipeline afterwards.
```go

p := pipe.Func(func(i int) (v int, b bool) {
        return i, true
    }).
    Filter(func(x *int) bool { return (*x) % 2 == 0})
    Gen(10).
    Do()
// p will be [0, 2, 4]
```

Take(n) expects the result to be of n length.
```go
p := pipe.Func(func(i int) (v int, b bool) {
        return i, true
    }).
    Filter(func(x *int) bool { return (*x) % 2 == 0})
    Take(10).
    Do()
// p will be [0, 2, 4, 6, 8, 10, 12, 14, 16, 18]
```

Watch out, if Take value is set uncarefully, it may jam the whole pipenile.
```go
// DO NOT DO THIS, IT WILL JAM
p := pipe.Func(func(i int) (v int, b bool) {
        return i, i < 10 // only 10 first values are not skipped
    }).
    Take(11). // we can't get any 11th value ever
    Parallel(4). // why not
    Do()
// Do() will try to evaluate the 11th value in 4 goroutines until it reaches maximum int value
```

### Example using `Filter` and `Map`:

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
	Filter(func(x *int) bool { return *x % 2 == 0 }).
	Map(func(x int) int { return len(strconv.Itoa(x)) }).
	Do()
// p will be [1, 1, 1, 1, 2]
```

### Example using `Map` and `Reduce`:

In this example Reduce is used in it's prefix form to be able to convert ints to string. 
```go
p := pipe.Reduce(
	pipe.Slice([]int{1, 2, 3, 4, 5}).
		Map(func(x int) int { return x * x }),
	func(x, y *int) string { 
		return strconv.Itoa(*x) + "-" + strconv.Itoa(y)
	},
)
// p will be "1-4-9-16-25"
```

In this example Reduce is used as usual in it's postfix form.
```go
p := pipe.Slice([]stirng{"Hello", "darkness", "my", "old", "friend"}).
	Map(strings.Title).
	Reduce(func(x, y *string) string { 
		return *x + " " + *y
	})
)
// p will be "Hello Darkness My Old Friend"
```

### Example of `Map` and `Reduce` with the underlying array type change:

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
strP := pipe.Map(p, func(x int) string { return strconv.Itoa(x) })
result := pipe.Reduce(strP, func(x, y *string) int { return len(*x) + len(*y) }).Do()
// result will be 45
```

### Example using `Sort`:

```go
p := pipe.Func(func(i int) (float32, bool) {
	return 100500-float32(i) * 0.9, true
}).
	Map(func(x float32) float32 { return x * x * 0.1 }).
	Gen(100500). // Sort is only availavle on pipes with known length
	Sort(pipies.Less[float32]). // pipies.Less(x, y *T) bool is available to all comparables
    // check out pipies package to find more usefull things
	Parallel(12).
	Do()
// p will contain the elements sorted in ascending order
```

### Example of infine sequence generation:

Here is an example of generating an infinite sequence of Fibonacci: 

```go
var fib []chan int
p := pipe.Func(func(i int) (int, bool) {
	if i < 2 {
		fib[i] <- i
		return i, true
	}
	p1 := <-fib[i-1]; fib[i-1] <- p1
	p2 := <-fib[i-2]; fib[i-2] <- p2
	
	fib[i] <- p1 + p2
	return p1 + p2, true
}).Parallel(20)
```

To generate a specific number of values, you can use the `Take` or `Gen` method: 

```go
// fill the array first
fib = make([]chan int, 60)
for i := range fib { fib[i] = make(chan int, 1) }
// do the Take
p = p.Take(60)
```

To accumulate the elements of the `Pipe`, you can use the `Reduce` or `Sum` method:

```go
sum := p.Sum(pipe.Sum[float32])
//also you can: sum := p.Reduce(func(x, y *float32) float32 { return *x + *y}) 
// sum will be the sum of the first 65000 random float32 values greater than 0.5
```

### Example using `Range` and `Map`:

```go
p := pipe.Range(10, 20, 2).Map(func(x int) int { return x * x }).Do()
// p will be [100, 144, 196, 256, 324]
```

### Example using `Repeat` and `Map`:

```go
p := pipe.Repeat("hello", 5).Map(strings.ToUpper).Do()
// p will be ["HELLO", "HELLO", "HELLO", "HELLO", "HELLO"]
```
Here is an example how you can handle multiple function returning error call this way:

```go
func foo() error {
    // <...>
    return nil
}

errs := pipe.Map(
            pipe.Repeat(foo, 50),
            func(f func() error) error { return f() },
        ).Do()

for _, e := range errs {
    if e != nil {
        log.Err(e)
    }
}
```

### Example using `Cycle` and `Filter`

```go
p := pipe.Cycle([]int{1, 2, 3}).Filter(func(x *int) bool { return *x % 2 == 0 }).Take(4).Do()
// p will be [2, 2, 2, 2]
```

### Example using `Erase` and `Collect`

```go
p := pipe.Slice([]int{1, 2, 3}).
Erase().
Map(func(x any) any {
    i := *(x.(*int))
    return &MyStruct{Weight: i}
}).Filter(x *any) bool {
    return (*x).(*MyStruct).Weight > 10
}
ms := pipe.Collect[MyStruct](p).Parallel(10).Do()
```



### Example of simple error handling

```go
y := pipe.NewYeti()
p := pipe.Range[int](-10, 10, 1).
	Yeti(y). // it's important to set yeti before yeeting, or the handle process will not be called
	MapFilter(func(x int) (int, bool) {
		if x == 0 {
			y.Yeet(errors.New("zero devision")) // yeet the error
			return 0, false                     // use MapFilter to filter out this value
		}
		return int(256.0 / float64(x)), true
	}).Snag(func(err error) {
	fmt.Println("oopsie-doopsie: ", err)
}).Do()

fmt.Println("p is: ", p)
/////////// output is:
// oopsie-doopsie:  zero devision
// p is:  [-25 -28 -32 -36 -42 -51 -64 -85 -128 -256 256 128 85 64 51 42 36 32 28]
```

This example demonstrates generating a set of values 256/i, where i ranges from -10 to 9 (excluding 10) with a step of 1. To handle division by zero, the library provides an error handling mechanism.

To begin, you need to create an error handler using the `pipe.NewYeti()` function. Then, register the error handler by calling the `Yeti(yeti)` method on your `pipe` object. This registered `yeti` will be the **last** error handler used in the `pipe` chain.

To **yeet** an error, you can use `y.Yeet(error)` from the registered `yeti` object.

To **handle** the yeeted error, use the `Snag(func(error))` method, which sets up an error handling function. You can set up multiple `Snag` functions, but all of them will consider the last `yeti` object set with the `Yeti(yeti)` method.

This is a simple example of how to handle basic errors. Below, you will find a more realistic example of error handling in a real-life scenario.

### Example of multiple error handling

```go
y1, y2 := pipe.NewYeti(), pipe.NewYeti()
users := pipe.Func(func(i int) (*domain.DomObj, bool) {
	domObj, err := uc.GetUser(i)
	if err != nil {
		y1.Yeet(err)
		return nil, false
	}
	return domObj, true
}).
	Yeti(y1).Snag(handleGetUserErr). // suppose we have some pre-defined handler
	MapFilter(func(do *domain.DomObj) (*domain.DomObj, bool) {
		enriched, err := uc.EnrichUser(do)
		if err != nil {
			return nil, false
		}
		return enriched, true
    }).Yeti(y2).Snag(handleEnrichUserErr).
	Do()
```

The full working code with samples of handlers and implementations of usecase functions can be found at: https://go.dev/play/p/YGtM-OeMWqu.


This example demonstrates how multiple error handling functions can be set up at different stages of the data processing pipeline to handle errors specific to each stage. 

Lets break down what is happening here. 

In this code fragment, there are two instances of `pipe.Yeti` created: `y1` and `y2`. These `Yeti` instances are used to handle errors at different stages of the data processing pipeline.

Within the `pipe.Func` operation, there are error-handling statements. When calling `uc.GetUser(i)`, if an error occurs, it is *yeeted* using `y1.Yeet(err)`, and the function returns `nil` and `false` to indicate the failure.

The `Yeti(y1).Snag(handleGetUserErr)` statement sets up an error handling function `handleGetUserErr` to handle the error thrown by `uc.GetUser(i)`. This function is defined elsewhere and specifies how to handle the error.

After that, the `MapFilter` operation is performed on the resulting `*domain.DomObj`. If the `uc.EnrichUser(do)` operation encounters an error, it returns `nil` and `false` to filter out the value.

The `Yeti(y2).Snag(handleEnrichUserErr)` statement sets up another error handling function `handleEnrichUserErr` to handle the error thrown by `uc.EnrichUser(do)`.

Finally, the `Do()` method executes the entire pipeline and assigns the result to the `users` variable.


## Is this package stable?

Yes it finally is **stable since v1.0.0**! All listed functionality is **fully covered** by unit-tests.
Functionality marked as TBD will be implemented as it described in the README and covered by unit-tests to be delivered stable. 

If there will be any method **signature changes**, the **major version** will be incremented. 

## Contributions

I will accept any pr's with the functionality marked as TBD. 

Also I will accept any sane unit-tests. 

Bugfixes. 

You are welcome to create any issues and connect to me via email. 

## What's next?

I hope to provide some roadmap of the project soon. 

Feel free to fork, inspire and use! 

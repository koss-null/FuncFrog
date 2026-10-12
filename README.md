# FuncFrog

[![Go Report Card](https://goreportcard.com/badge/github.com/koss-null/funcfrog)](https://goreportcard.com/report/github.com/koss-null/funcfrog)
[![Go Reference](https://pkg.go.dev/badge/github.com/koss-null/funcfrog.svg)](https://pkg.go.dev/github.com/koss-null/funcfrog)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://raw.githubusercontent.com/koss-null/funcfrog/master/coverage.svg?raw=true)](coverage.svg)

![FuncFrog icon](https://github.com/koss-null/funcfrog/blob/master/FuncFrogIco.jpg?raw=true)

*FuncFrog* is a tiny library for doing **fast**, **parallel**, **lazy** `Map`, `Filter`, `Reduce` and [a bunch of other things](#api-reference) over slices and other sequences — all chained into a single pipeline. You feed it a source (a slice, a `func`, a range...), stack a few operations on top, and pull the trigger with `Do()`.

Everything runs in parallel with **minimal copying and locking**, and there's built-in [error handling](#error-handling-yeti--snag) that is honestly a bit weird but very handy once it clicks (hello `Yeet`/`Snag`).

I tried hard to keep the API friendly and predictable. If you want to see how it performs against a plain `for` loop on your machine, run `cd perf/; make` (spoiler: a single goroutine is roughly on par, and FuncFrog pulls ahead the moment you go parallel).

## Table of Contents

- [Getting Started](#getting-started)
- [How it works](#how-it-works)
- [API reference](#api-reference)
  - [Constructors](#constructors)
  - [Set the pipe length: Take & Gen](#set-the-pipe-length-take--gen)
  - [Go parallel](#go-parallel)
  - [Transforming data](#transforming-data)
  - [Grabbing a single value](#grabbing-a-single-value)
  - [Running the pipe: Do & Promices](#running-the-pipe-do--promices)
  - [Erase & Collect](#erase--collect)
  - [Error handling: Yeti & Snag](#error-handling-yeti--snag)
  - [Ready-made helpers: pipies](#ready-made-helpers-pipies)
  - [Prefix helpers](#prefix-helpers)
  - [The ff shortcuts](#the-ff-shortcuts)
  - [Still cooking (TBD)](#still-cooking-tbd)
- [Examples](#examples)
  - [Basic](#basic)
  - [Func + Take](#func--take)
  - [Func + Gen](#func--gen)
  - [Take vs Gen](#take-vs-gen)
  - [Filter + Map](#filter--map)
  - [Map + Reduce](#map--reduce)
  - [Map + Reduce with a type change](#map--reduce-with-a-type-change)
  - [Sort](#sort)
  - [An infinite sequence](#an-infinite-sequence)
  - [Range + Map](#range--map)
  - [Repeat + Map](#repeat--map)
  - [Cycle + Filter](#cycle--filter)
  - [Erase + Collect](#erase--collect)
  - [Simple error handling](#simple-error-handling)
  - [Multiple error handling](#multiple-error-handling)
- [Performance](#performance)
- [Is this stable?](#is-this-stable)
- [Contributing](#contributing)
- [What's next?](#whats-next)

## Getting Started

Grab it with:

```
go get github.com/koss-null/funcfrog
```

Then import the `pipe` package (that's the main one you need):

```go
import "github.com/koss-null/funcfrog/pkg/pipe"
```

...and start chaining:

```go
res := pipe.Slice(a).
	Map(func(x int) int { return x * x }).
	Filter(func(x *int) bool { return *x > 100 }).
	Parallel(12).
	Do()
```

If you just want to fire a single function over a slice and skip building a pipe by hand, the `ff` package is your friend:

```go
import "github.com/koss-null/funcfrog/pkg/ff"

res := ff.Map(strArr, strings.ToUpper).Do()
```

Want to poke at real code? Jump straight to the [Examples](#examples).

## How it works

The whole library revolves around two interfaces:

- **`Piper[T]`** — a *lazy* sequence whose **length is known** (a slice, a `Range`, a `Repeat`...). It can do everything.
- **`PiperNoLen[T]`** — a *lazy* sequence whose **length is unknown yet** (a `Func`, a `Cycle`, ...). It can do almost everything, plus it can *become* a `Piper` via [`Take` or `Gen`](#set-the-pipe-length-take--gen).

Both are lazy: nothing actually runs until you call a terminal method (`Do`, `Reduce`, `Sum`, `Count`, `First`, `Any`...). Until then you're just describing the pipeline.

Because a pipe is a value, you can copy it around freely — assign it to a variable, reuse it, branch it:

```go
base := pipe.Slice(a).Map(square)
evens := base.Filter(isEven)
odds  := base.Filter(isOdd) // `base` is untouched
```

Operations can change the element type too, but two flavors exist:
- `Pipe.Map` keeps the **same** type (`func(T) T`) — cheap and easy.
- `pipe.Map` (prefix) / `ff.Map` change the **type** (`func(SrcT) DstT`) — see [Prefix helpers](#prefix-helpers).

## API reference

Every bullet below is a real signature — copy/paste safe.

### Constructors

- :frog: `Slice([]T) Piper[T]` — wraps a slice. *Length known.*
- :frog: `Func(fn func(i int) (T, bool)) PiperNoLen[T]` — builds a lazy sequence from a function of the index `i`. The bool says whether the value exists (`true`) or should be skipped (`false`). *Length unknown.*
- :frog: `Fn(fn func(i int) T) PiperNoLen[T]` — like `Func`, but nothing is ever skipped. *Length unknown.*
- :frog: `FuncP(fn func(i int) (*T, bool)) PiperNoLen[T]` — like `Func`, but your function hands back pointers (handy for big structs). *Length unknown.*
- :frog: `Cycle(a []T) PiperNoLen[T]` — loops over the slice forever. *Length unknown.*
- :frog: `Range[T Integer|Float](start, finish, step T) Piper[T]` — generates `start, start+step, ...` and stops **before** `finish` (i.e. `[start, finish)`). Works for any integer or float. *Length known.*
- :frog: `Repeat(x T, n int) Piper[T]` — `x` repeated `n` times. A non-positive `n` gives an empty pipe. *Length known.*

### Set the pipe length: Take & Gen

These two only exist on `PiperNoLen` and are how you turn an unknown-length pipe into a `Piper`. Only the **first** `Take`/`Gen` in a chain counts.

- :frog: `Take(n int) Piper[T]` — "I want `n` values in the end". FuncFrog keeps evaluating until it has collected `n` *surviving* values. *Length becomes known.*
- :frog: `Gen(n int) Piper[T]` — "evaluate the first `n` positions". Note the difference: it evaluates positions `[0, n)` and your `Filter` can still drop some, so the result may be shorter than `n`. *Length becomes known.*

### Go parallel

- :frog: `Parallel(n uint16) Piper[T]` (and `PiperNoLen[T]`) — split the work across `n` goroutines. Default is `1`. Only the **first** `Parallel` in a chain is applied.

### Transforming data

- :frog: `Map(fn func(T) T) Piper[T]` — apply `fn` to every element (same type in, same type out).
- :frog: `Filter(fn func(*T) bool) Piper[T]` — keep only the elements for which `fn` returns `true`. (**Tip:** the argument is a `*T`, so remember to dereference — `*x`.)
- :frog: `MapFilter(fn func(T) (T, bool)) Piper[T]` — map and filter in one pass. If the returned bool is `false`, the value is dropped. Great for "skip this one" cases, [including error handling](#simple-error-handling).
- :frog: `Sort(less func(*T, *T) bool) Piper[T]` — sort the current values. **Requires a known length**, so `Gen`/`Slice`/`Range`/`Repeat` first.
- :frog: `Reduce(fn func(*T, *T) T) *T` — fold the whole pipe into a single value, left to right. Returns `nil` for an empty pipe. (*Postfix* reduce: type stays the same.)
- :frog: `Sum(plus func(*T, *T) T) T` — like `Reduce` but **parallel-friendly**, so it wants an associative function (`+`, `*`, `min`, `max`...).

### Grabbing a single value

- :frog: `First() *T` — the first element, or `nil` if there is nothing (yes, it's a pointer).
- :frog: `Any() *T` — *some* (effectively random under parallelism) element, or `nil` if empty.
- :frog: `Count() int` — how many elements the pipe would produce. It still runs the pipeline, but skips building the result slice — so it's the memory-light way to ask "how many?".

### Running the pipe: Do & Promices

- :frog: `Do() []T` — **execute** everything and return the resulting slice. This is the usual "end of pipeline" call.
- :frog: `Promices() []func() (T, bool)` — instead of running now, get one lazy thunk per position. Each thunk returns `(value, ok)`; `ok == true` means "this value exists, append it". Handy if you want to schedule the work yourself. (The funny spelling is intentional — they're "promices" :)

### Erase & Collect

Sometimes your pipeline needs to lug around different types at different steps. `Erase` gives up the compile-time type so you can work with `any`, and `Collect` puts a type back on.

- :frog: `Erase() Piper[any]` (or `PiperNoLen[any]`) — every element becomes `any` (backed by a pointer to the original value).
- :frog: `pipe.Collect[DstT](Piper[any]) Piper[DstT]` — pull a `Piper[any]` back into a concrete `Piper[DstT]`.
- :frog: `pipe.CollectNL[DstT](PiperNoLen[any]) PiperNoLen[DstT]` — the same for a pipe with no known length.

A gotcha worth knowing: `Collect[DstT]` asserts each stored `any` to `*DstT`. So whatever you put into an erased pipe must be a **pointer** to the type you're collecting back — an `Erase()` of `Piper[int]` hands you `any` holding a `*int`, so you'd `Collect[int]`.

### Error handling: Yeti & Snag

This part is a little unusual, so read slowly.

- :frog: `Yeti(y internalpipe.YeetSnag) Piper[T]` — register a *yeti* on the pipe. A yeti is a little error collector you create with `pipe.NewYeti()`.
- :frog: `NewYeti() internalpipe.YeetSnag` — makes a yeti. It has `y.Yeet(err)` to throw an error into it.
- :frog: `Snag(h func(error)) Piper[T]` — attach a handler that receives whatever got `Yeet`-ed into the **last** yeti that was set via `Yeti`.

In practice: create a yeti, register it with `Yeti(y)` **before** you start yeeting at it, yeet errors from inside your functions, and let `Snag` deal with them at the end. See the [examples](#simple-error-handling) — they make it click way faster than this paragraph.

### Ready-made helpers: pipies

Package `pipies` has the little functions you'd otherwise rewrite every time:

- :frog: `Sum[T]` — a reducer: `func(a, b *T) T { return *a + *b }`. Works for numbers and `~string`.
- :frog: `Less[T]` — the everyday comparator for `Sort`.
- :frog: **Predicates:** `NotNil`, `IsNil`, `NotZero`.
- :frog: **Predicate builders:** `Eq(x)`, `NotEq(x)`, `LessThan(x)`, `Distinct(getKey)`.
- :frog: **Negation:** `Not(fn)`, `Nott(fn)`, `Nottt(fn)` — flip a boolean-returning function with 1, 2, or 3 args (hence the extra t's).

### Prefix helpers

`Erase`/`Collect` work, but all those pointer conversions are annoying. The prefix functions in `pkg/pipe` are the pleasant way to change a pipe's type: pass the pipe first, the function second.

- :frog: `pipe.Map(Piper[SrcT], func(SrcT) DstT) Piper[DstT]` — map + change type, known length.
- :frog: `pipe.MapNL(PiperNoLen[SrcT], func(SrcT) DstT) PiperNoLen[DstT]` — same, unknown length.
- :frog: `pipe.MapFilter(Piper[SrcT], func(SrcT) (DstT, bool)) Piper[DstT]` — map + filter + change type, known length.
- :frog: `pipe.MapFilterNL(PiperNoLen[SrcT], func(SrcT) (DstT, bool)) PiperNoLen[DstT]` — same, unknown length.
- :frog: `pipe.Reduce(Piper[SrcT], func(*DstT, *SrcT) DstT, initVal ...DstT) DstT` — prefix reduce that can also change type. Mind the argument order: the accumulator (`*DstT`) comes first, then the current element (`*SrcT`). `initVal` is optional and seeds the fold.

### The ff shortcuts

When a whole pipe is overkill and you just want to run one function over a slice, `pkg/ff` has one-liners:

- :frog: `ff.Map([]SrcT, func(SrcT) DstT) pipe.Piper[DstT]`
- :frog: `ff.MapFilter([]SrcT, func(SrcT) (DstT, bool)) pipe.Piper[DstT]`
- :frog: `ff.Filter([]SrcT, func(*SrcT) bool) pipe.Piper[SrcT]`
- :frog: `ff.Reduce([]SrcT, func(*DstT, *SrcT) DstT, initVal ...DstT) DstT`
- :frog: `ff.Compose(fn1 func(T1) T2, fn2 func(T2) T3) func(T1) T3` — glue two functions into one.

### Still cooking (TBD)

These aren't in yet — they're on the wishlist:

- :seedling: `Until(fn func(*T) bool)` — evaluate one-by-one until `fn` says stop. (Single-threaded only by nature.)
- :seedling: `IsAny() bool` — does the pipe have *any* element?
- :seedling: `MoreThan(n int) bool` — does it have *more than* `n`?
- :seedling: `Reverse() Pipe` — reverse the underlying slice.

## Examples

### Basic

```go
res := pipe.Slice(a).
	Map(func(x int) int { return x * x }).
	Map(func(x int) int { return x + 1 }).
	Filter(func(x *int) bool { return *x > 100 }).
	Filter(func(x *int) bool { return *x < 1000 }).
	Parallel(12).
	Do()
```

### Func + Take

```go
p := pipe.Func(func(i int) (v int, b bool) {
	if i < 10 {
		return i * i, true
	}; return
}).Take(5).Do()
// p will be [0, 1, 4, 9, 16]
```

`Take(5)` asks for 5 values, and the first five the function yields are `0,1,4,9,16`.

### Func + Gen

```go
p := pipe.Func(func(i int) (v int, b bool) {
	if i < 10 {
		return i * i, true
	}; return
}).Gen(5).Do()
// p will be [0, 1, 4, 9, 16]
```

`Gen(5)` evaluates positions `0..4`, which gives the same result here.

### Take vs Gen

**`Gen(n)`** evaluates the first `n` positions, then the filters get their say — so you can end up with **fewer than `n`** elements:

```go
p := pipe.Func(func(i int) (v int, b bool) {
	return i, true
}).
	Filter(func(x *int) bool { return *x%2 == 0 }).
	Gen(10).
	Do()
// p will be [0, 2, 4, 6, 8]
```

All ten positions `0..9` are evaluated, but only the evens survive — 5 values, not 10.

**`Take(n)`** keeps going until `n` values actually make it through the filter:

```go
p := pipe.Func(func(i int) (v int, b bool) {
	return i, true
}).
	Filter(func(x *int) bool { return *x%2 == 0 }).
	Take(10).
	Do()
// p will be [0, 2, 4, 6, 8, 10, 12, 14, 16, 18]
```

And here's the trap: if `Take` asks for more values than your source can ever produce, it will spin all the way up to `math.MaxInt` and then panic.

```go
// DO NOT DO THIS — it will grind until math.MaxInt and then panic
p := pipe.Func(func(i int) (v int, b bool) {
	return i, i < 10 // only the first 10 values survive
}).
	Take(11). // ...but we ask for an 11th, which never comes
	Parallel(4).
	Do()
// panic: the limit have been exceeded, but the result is not calculated
```

### Filter + Map

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
	Filter(func(x *int) bool { return *x%2 == 0 }).
	Map(func(x int) int { return len(strconv.Itoa(x)) }).
	Do()
// p will be [1, 1, 1, 1, 2]
```

### Map + Reduce

Here `Reduce` is used in its **prefix** form, which lets us fold `int`s down into a `string`:

```go
p := pipe.Reduce(
	pipe.Slice([]int{1, 2, 3, 4, 5}).
		Map(func(x int) int { return x * x }),
	func(x *string, y *int) string {
		if *x == "" { // first step: the accumulator is still the zero value
			return strconv.Itoa(*y)
		}
		return *x + "-" + strconv.Itoa(*y)
	},
)
// p will be "1-4-9-16-25"
```

And the same thing in the usual **postfix** form:

```go
capitalize := func(s string) string { return strings.ToUpper(s[:1]) + s[1:] }

p := pipe.Slice([]string{"Hello", "darkness", "my", "old", "friend"}).
	Map(capitalize).
	Reduce(func(x, y *string) string {
		if *x == "" {
			return *y
		}
		return *x + " " + *y
	})
// *p will be "Hello Darkness My Old Friend"  (Reduce returns a *string)
```

### Map + Reduce with a type change

Changing the element type *and* folding, using the prefix `Map` and `Reduce`:

```go
p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})     // Piper[int]
strP := pipe.Map(p, strconv.Itoa)                       // Piper[string]
result := pipe.Reduce(strP, func(x *int, y *string) int { return *x + len(*y) })
// result will be 45 (nine single-digit strings, each contributing 1)
```

### Sort

```go
p := pipe.Func(func(i int) (float32, bool) {
	return 100500 - float32(i)*0.9, true
}).
	Map(func(x float32) float32 { return x * x * 0.1 }).
	Gen(100500).                // Sort needs a known length
	Sort(pipies.Less[float32]). // pipies.Less is a ready-made comparator
	Parallel(12).
	Do()
// p holds the elements sorted in ascending order
```

Check the `pipies` package for more handy predicates and comparators.

### An infinite sequence

`Func` can be stateful, so you can describe a sequence that never ends — here's a lazy Fibonacci generator. Each value is stashed in a one-slot channel so the next call can find it. To keep the order sane, let it run on a single goroutine (the default) and cap it with `Take`:

```go
fib := make([]chan int, 60)
for i := range fib {
	fib[i] = make(chan int, 1)
}

p := pipe.Func(func(i int) (int, bool) {
	if i < 2 {
		fib[i] <- i
		return i, true
	}
	p1, p2 := <-fib[i-1], <-fib[i-2]
	fib[i-1], fib[i-2] = p1, p2 // put them back for the next call
	fib[i] <- p1 + p2
	return p1 + p2, true
}).Take(60) // grab the first 60 numbers
```

Once you have a `Piper`, accumulate away:

```go
sum := p.Sum(pipies.Sum[int])
// sum is the sum of the first 60 Fibonacci numbers
```

### Range + Map

```go
p := pipe.Range(10, 20, 2).Map(func(x int) int { return x * x }).Do()
// p will be [100, 144, 196, 256, 324]
```

### Repeat + Map

```go
p := pipe.Repeat("hello", 5).Map(strings.ToUpper).Do()
// p will be ["HELLO", "HELLO", "HELLO", "HELLO", "HELLO"]
```

Handy for calling a bunch of error-returning functions and collecting the errors:

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
		log.Println(e)
	}
}
```

### Cycle + Filter

```go
p := pipe.Cycle([]int{1, 2, 3}).Filter(func(x *int) bool { return *x%2 == 0 }).Take(4).Do()
// p will be [2, 2, 2, 2]
```

### Erase + Collect

Wiping the type to `any`, turning ints into a custom struct, then collecting back:

```go
type MyStruct struct {
	Weight int
}

p := pipe.Slice([]int{1, 2, 3}).
	Erase(). // every element is now `any` (backed by a *int)
	Map(func(x any) any {
		return &MyStruct{Weight: *(x.(*int)) * 100} // store a *MyStruct
	}).
	Filter(func(x *any) bool {
		return (*x).(*MyStruct).Weight > 100
	})

ms := pipe.Collect[MyStruct](p).Parallel(10).Do()
// ms will be []MyStruct{{Weight: 200}, {Weight: 300}}
```

### Simple error handling

```go
y := pipe.NewYeti()

p := pipe.Range[int](-10, 10, 1).
	Yeti(y). // register the yeti before anything gets yeeted at it
	MapFilter(func(x int) (int, bool) {
		if x == 0 {
			y.Yeet(errors.New("zero division")) // throw the error
			return 0, false                     // and skip this value
		}
		return int(256.0 / float64(x)), true
	}).
	Snag(func(err error) {
		fmt.Println("oopsie-doopsie:", err)
	}).
	Do()

fmt.Println("p is:", p)
// oopsie-doopsie: zero division
// p is: [-25 -28 -32 -36 -42 -51 -64 -85 -128 -256 256 128 85 64 51 42 36 32 28]
```

This builds `256/i` for `i` from `-10` to `9`. The `0` would explode, so we `Yeet` the error and let `MapFilter` drop that value.

The flow is:

1. `pipe.NewYeti()` — make an error collector.
2. `Yeti(y)` — register it on the pipe. **Do this before yeeting**, or nothing will catch your error.
3. `y.Yeet(err)` — throw an error from inside your function.
4. `Snag(handler)` — attach the function that deals with whatever was yeeted. You can stack several `Snag`s; they all go to the **last** yeti registered with `Yeti`.

### Multiple error handling

Different stages of a pipeline can have their own yetis and handlers — nice for a real user-enrichment flow:

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
	Yeti(y1).Snag(handleGetUserErr). // some pre-defined handler
	MapFilter(func(do *domain.DomObj) (*domain.DomObj, bool) {
		enriched, err := uc.EnrichUser(do)
		if err != nil {
			y2.Yeet(err)
			return nil, false
		}
		return enriched, true
	}).
	Yeti(y2).Snag(handleEnrichUserErr).
	Gen(userCount). // cap it, since Func is endless on its own
	Do()
```

There are two yetis here: `y1` guards `GetUser`, `y2` guards `EnrichUser`. Errors from each stage are yeeted into their own yeti and handed to their own `Snag`. A runnable version (with stub handlers and usecase functions) lives at https://go.dev/play/p/YGtM-OeMWqu.

## Performance

There's a ready-made benchmark suite in `perf/`. To run it:

```
cd perf/
make
```

It compares FuncFrog against plain `for` loops for `Map`, `Filter`, `Reduce`, `Any` and friends. On a single goroutine you're basically at `for`-loop speed; more goroutines = more win. Numbers depend on your machine, so run it yourself.

## Is this stable?

Yep — **stable since v1.0.0**. Everything listed above is covered by unit tests. The stuff under [Still cooking](#still-cooking-tbd) will land as described and get tests before it's called stable.

If I ever change a method signature, the **major version goes up**.

## Contributing

PRs are welcome, especially for the [TBD](#still-cooking-tbd) features. Same goes for sane unit tests and good bug reports. Open an issue anytime, or email me.

## What's next?

A proper roadmap is coming. Until then — fork it, play with it, and enjoy.

Happy piping! :frog:

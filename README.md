# Lambda 

![Lambda gopher picture](https://github.com/koss-null/lambda/blob/master/lambda_favicon.png?raw=true) 

Is a library to provide `map`, `reduce` and `filter` operations on slices in one pipeline.  
The slice can be set by generating function. Also the parallel execution is supported.  
It's strongly expected all function arguments to be **pure functions** (functions with no side effects, can be cached by
args).  

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
 
To play around with simple examples check out: `go run examples/main.go`.  
Also feel free to run it with `go run examples/main.go`.  
 
### Supported functions:
 
```go
Slice([]T) // creates a pipe of a given type T from a slice.
```
```go
Func(func(i int) (T, bool)) // creates a pipe of type T from a function.
```
The bool parameter is wether it's taken (true if it's taken, false - skipped), `i int` is for the i'th element of initial pipe
```go
Take(n int) // if it's a Func made pipe, it expects n values to be eventually returned
```
```go
Gen(n int) // if it's a Func made pipe, it generates a sequence from [0, n) and applies other function to it
```
Only the first of `Gen`, `Take` or `First` functions in the pipe (from left to right) is applied.  
You can't have `Func` made pipe and not set either `Gen` or `Take`, it's set to 0 by default.  
```go
Parallel(n int) // set the number of goroutines to be executed on (4 by default)
```
```go
Map(fn func(x T) T) // applies function fn to every element of the pipe and gets a new pipe thus
```
```go
Filter(fn func(x T) bool) // only leaves the element in a pipe if fn for this element is true
```
```go
Sort(less func(x, y T) bool) // sotrs the array in parallel, 
// you may use pipe.Less function for constraints.Ordered types
```
```go
Do() []T // executes all the pipe and returns the resulting slice
```
```go
Reduce(func(x, y T) T) T // executes all the pipe and returns the resulting value
```
```go
Sum(func(x, y) T) T // is pretty similar to Reduce, but works in parallel
```
you about to send any function here where f(a, b) = f(b, a) and f(f(a, b), c) == f((c, b), a) == f(f(a,c), b)  
```go
First() T // returns the first found value in the result slice
// Works with Func(...) like magic
```
there is also a function that works with `Func()` but without both `Take()` or `Gen()`: Any!  
```go
Any() T // returns the first found T instance (ont in order)
```
 
### But what about Map and Reduce to another type?
 
It is possible with `prefixpipe.go`: it provides `Map(T1) T2` and `Reduce(T1, T2) T1`  
It may look a little bit more ugly, but it is what it is for now and untill v1.0.0 for sure.  
I am concidering to create some another Pipe implementations in future to be able to write beautiful oneline
convertions.  
 
### Does it stable?
 
In short: not yet. But(!) for each release I do manual testing of everything I have touched since the previous release
and also I have a nice pack of unit-tests. So I beleve it is stable enough to use in your pet projects.  
I will provide some more convincing quality guarantees and benchmarks with `v1.0.0` release.  
 
### To be done functions (the names are not settled yet):
 
```go
IsAny() bool // returns true if there is at least 1 element in the result slice
MoreThan(n int) bool // returns if there is at least n elements in the result slice
Reverse() // reverses the underlying slice
```
 
### Important note:
 
```
pipelines initiated with Func(...) are not parallelize yet
```


### Quick usage review
 
Here are some more examples (pretty mutch same as in the example package):  
First of all let's make a simple slice:  

```go
a := make([]int, 100)
for i := range a {
	a[i] = i
}
```

Let's wobble it a little bit:  
```go
pipe.Slice(a).
	Map(func(x int) int { return x * x }).
	Map(func(x int) int { return x + 1 }).
	Filter(func(x int) bool { return x > 100 }).
	Filter(func(x int) bool { return x < 1000 }).
	Do()
```
 
Here are some fun facts: 
* it's executed in *4 threads* by default. (I am still hesitating about it, so the value may still change [most likley to 1]) 
* if there is less than *5k* items in your slice, it will be executed in a single thread anyway. 
 
Do I have some more tricks? Shure!  
```go
pipe.Func(func(i int) (float32, bool) {
	rnd := rand.New(rand.NewSource(42))
	rnd.Seed(int64(i))
	return rnd.Float32(), true
}).
	Filter(func(x float32) bool { return x > 0.5 }).
```
This one generates an infinite random sequence greater than 0.5  
Use `Gen(n)` to generate exactly **n** items and filter them after,  
Or just `Take(n)` exactly **n** items evaluating them while you can (until the biggest `int` is sent to a function).  
 
Let's assemble it all in one:  
```go
pipe.Func(func(i int) (float32, bool) {
	rnd := rand.New(rand.NewSource(42))
	rnd.Seed(int64(i))
	return rnd.Float32(), true
}).
	Filter(func(x float32) bool { return x > 0.5 }).
	// Take 65000 values, don't stop generate before get all of them
	Take(65000).
	// There is no Sum() yet but no worries
	Reduce(func(x, y float32) float32 { return x + y})
```
Did you notice the `Reduce`?  
Here is what reduce does:  
Let's say there is **zero** element for type **T** where any `f(zero, T) = T`  
No we can add this **zero** at the beginning of the pipe and apply our function to the first two elements: `p[0]`[**zero**] and `p[1]`  
Result we will put into `p[0]` and move the whole pipe to the left by 1   
Due to what we've said before about **zero** `p[0] = p[1]` now   
Now we are doing the same operation for `p[0]` and `p[1]` while we have `p[1]`   
Eventually `p[0]` will be the result.  
 
Anything else?  
You can  sort faster than the wind using all power of your core2duo:  
```go
pipe.Func(func(i int) (float32, bool) {
	return float32(i) * 0.9, true
}).
	Map(func(x float32) float32 { return x * x }).
	Gen(100500).
	Sort(pipe.Less[float32]).
	// This is how you can sort in parallel (it's rly faster!)
	Parallel(12).
	Do()
```
Also if you don't whant to carry the result array on your shoulders and only worry about the amount, you better use
`Count()` instead of `Do()`.  
 
### Contribution
 
For now I will accept any sane tests. Feel free to use any frameworks you would like.  
Any bugfixes are also welcome. I am going to do some refactor and maybe some decisions will be changed, so I will not
accept any new features PR's for now.  
 
###What's next?  
 
I hope to provide some roadmap of the project soon.   
Also I am going to craft some unit-tests and may be set up github pipelines eventually.   
Feel free to fork, inspire and use! I will try to supply all version tags by some manual testing and quality
control at least.   

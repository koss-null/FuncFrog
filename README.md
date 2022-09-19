# Lambda 

![Lambda gopher picture](https://github.com/koss-null/lambda/blob/master/lambda_favicon.png?raw=true) 

Is a library to provide `map`, `reduce` and `filter` operations on slices in one pipeline. 
The slice can be set by generating function. Also the parallel execution is supported. 
It's strongly expected all function arguments to be **clear functions** (functions with no side effects). 

###Basic example:

```go
res := pipe.Slice(a).
	Map(func(x int) int { return x * x }).
	Map(func(x int) int { return x + 1 }).
	Filter(func(x int) bool { return x > 100 }).
	Filter(func(x int) bool { return x < 1000 }).
	Parallel(12).
	Do()
```

Usage examples you may found in: `pkg/example/` 

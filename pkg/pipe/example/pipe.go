package example

import (
	"fmt"
	"math/rand"

	"github.com/koss-null/lambda/pkg/pipe"
)

// This are examples of pipe usage
func main() {
	a := make([]int, 100)
	for i := range a {
		a[i] = i
	}

	// just a simple pipeline
	res := pipe.Slice(a).
		Map(func(x int) int { return x * x }).
		Map(func(x int) int { return x + 1 }).
		Filter(func(x int) bool { return x > 100 }).
		Filter(func(x int) bool { return x < 1000 }).
		Do()
	fmt.Println("result1: ", res)

	// walues can be achieved from a function
	res2 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		// This one is not working since x is float32 now
		// Filter(func(x int) bool { return x < 1000 }).
		// We need to provide another function:
		Filter(func(x float32) bool { return x > 0.5 }).
		// Generate 100 values as an initial pipeline
		Gen(100).
		Do()
	fmt.Println("result2: ", res2)

	// walues can be achieved from a function
	res21 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.5 }).
		// Get 100 values, don't stop generate before get all of them
		Get(100).
		Do()
	fmt.Println("result2.1: ", res21)

	// walues can be achieved from a function
	res22 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.5 }).
		// if there is no Get and no Gen, the len of result slice is 0
		Do()
	fmt.Println("result2.2: ", res22)

	res23 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		// This one is not working since x is float32 now
		// Filter(func(x int) bool { return x < 1000 }).
		Filter(func(x float32) bool { return x > 0.5 }).
		// if there is no Get and no Gen, the Count of result is -1
		Count()
	fmt.Println("result2.3: ", res23)

	// you can just count values:
	res3 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.6 }).
		Gen(100).
		Count()
	fmt.Println("result3: ", res3)

	// trying to count values with Get(n int) will return n:
	res31 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.6 }).
		Get(100).
		Count()
	fmt.Println("result3.1: ", res31)

	// you can set the amount of goroutines using Parallel(n int)
	// the defalut value is 4
	res4 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Gen(1_000_000_0).
		Parallel(12).
		Count()
	fmt.Println("result4: ", res4)

	// if you need another type on map's output there is only an ugly prefix solution
	// since go does not support method's type parameters (which seems to make struct signature known at compile)
	res5 := pipe.Map(pipe.
		Func(func(i int) (float32, bool) {
			return float32(i) * 0.9, true
		}).
		Gen(100),
		func(x float32) int {
			if x-float32(int(x)) > 0.7 {
				return int(x) + 1
			}
			return int(x)
		},
	).Do()
	fmt.Println("result5: ", res5)

	// and even Sort them
	// TO BE IMPLEMENTED
	// res4 := pipe.Func(func(i int) (float32, bool) {
	// rnd.Seed(int64(i))
	// return rnd.Float32(), true
	// }).
	// use default pile.Less if the slice data type is comparable with "<"
	// Sort(pipe.Less[float32]).
	// Do()
	// fmt.Println("result4: ", res4)

	// you can also have reduce
	// TODO: add reduce example
}

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/pkg/profile"

	"github.com/koss-null/lambda/pkg/pipe"
)

// This are examples of pipe usage
func main() {
	defer profile.Start(profile.ProfilePath(".")).Stop()

	a := make([]int, 100)
	for i := range a {
		a[i] = i
	}

	// just a simple pipeline
	res1 := pipe.Slice(a).
		Map(func(x int) int { return x * x }).
		Map(func(x int) int { return x + 1 }).
		Filter(func(x int) bool { return x > 100 }).
		Filter(func(x int) bool { return x < 1000 }).
		Do()

	fmt.Println("1: simple pipeline result")
	fmt.Println(`
	pipe.Slice(a).
		Map(func(x int) int { return x * x }).
		Map(func(x int) int { return x + 1 }).
		Filter(func(x int) bool { return x > 100 }).
		Filter(func(x int) bool { return x < 1000 }).
		Do()`)
	fmt.Println(res1)

	// walues can be achieved from a function
	res2 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		// the line below is not working since x is float32 now
		// Filter(func(x int) bool { return x < 1000 }).
		// We need to provide another function:
		Filter(func(x float32) bool { return x > 0.5 }).
		// Generate 100 values as an initial pipeline
		Gen(100).
		Do()

	fmt.Println("2: creating pipe from a function")
	fmt.Println(`
	pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
	Filter(func(x float32) bool { return x > 0.5 }).
	Gen(100).
	Do()`)
	fmt.Println(res2)

	// We can Take(n) the exact amount of values - it will be executed until the length of the result slice will be n
	res21 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.5 }).
		// Take 100 values, don't stop generate before get all of them
		Take(100).
		Do()

	fmt.Println("2.1: creating pipe from a function but using Take(n) to gen exactly n values")
	fmt.Println(`
	pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.5 }).
		// Take 100 values, don't stop generate before get all of them
		Take(100).
		Do()
	`)
	fmt.Println(res21)

	// walues can be achieved from a function, but it have no sence if there is no Take() or Gen()
	res22 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.5 }).
		// if there is no Take and no Gen, the len of result slice is 0
		Do()

	fmt.Println("2.2: if there is no Take and no Gen, the len of result slice is 0")
	fmt.Println(`
	pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.5 }).
		Do()
	`)
	fmt.Println(res22)

	res23 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		// This one is not working since x is float32 now
		// Filter(func(x int) bool { return x < 1000 }).
		Filter(func(x float32) bool { return x > 0.5 }).
		// if there is no Take and no Gen, the Count of result is -1
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

	// trying to count values with Take(n int) will return n:
	res31 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.6 }).
		Take(100).
		Count()
	fmt.Println("result3.1: ", res31)

	runtime.GC()
	start := time.Now()
	arr4 := make([]float32, 0, 1_000_000_0)
	cnt := 0
	for i := 0; i < 1_000_000_0; i++ {
		arr4 = append(arr4, float32(i)*0.9)
	}
	for i := 0; i < 1_000_000; i++ {
		arr4[i] = arr4[i] * arr4[i]
		if arr4[i] > 5000.6 {
			cnt++
		}
	}
	fmt.Println("result4 (by-hand): ", cnt, " time ", time.Now().Sub(start))
	runtime.GC()

	runtime.GC()
	// you can set the amount of goroutines using Parallel(n int)
	start = time.Now()

	res4 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Gen(1_000_000).
		Parallel(12).
		Count()
	fmt.Println("result4: (parallel 12)", res4, "; eval took ", time.Now().Sub(start))
	runtime.GC()

	// single thread
	start = time.Now()
	res41 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Gen(1_000_000).
		Parallel(1).
		Count()
	fmt.Println("result41: (parallel 1)", res41, "; eval took ", time.Now().Sub(start))
	runtime.GC()

	// the defalut value is 4
	start = time.Now()
	res42 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Gen(1_000_000).
		Count()
	fmt.Println("result42 (parallel 4): ", res42, "; eval took ", time.Now().Sub(start))
	runtime.GC()

	// many goroutines are OK
	start = time.Now()
	res43 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Gen(1_000_000).
		Parallel(150).
		Count()
	fmt.Println("result43: (parallel 150)", res43, "; eval took ", time.Now().Sub(start))
	runtime.GC()

	// but the best value shoul be about the amount of your CPUs (do not set it very high)
	// the threshold is 256
	start = time.Now()
	res44 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Gen(1_000_000).
		Parallel(100500).
		Count()
	fmt.Println("result44: (parallel 1240)", res44, "; eval took ", time.Now().Sub(start))
	runtime.GC()

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
	// 	rnd.Seed(int64(i))
	// 	return rnd.Float32(), true
	// }).
	// 	// use default pile.Less if the slice data type is comparable with "<"
	// 	Sort(pipe.Less[float32]).
	// 	Do()
	// fmt.Println("result4: ", res4)

	// you can also have reduce
	// TODO: add reduce example
}

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/pkg/profile"

	"github.com/koss-null/lambda/pkg/pipe"
	"github.com/koss-null/lambda/pkg/pipies"
)

// User is an example struct
type User struct {
	FirstName, LastName string
	Age                 int
	OtherInfo           any
}

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
	fmt.Println(`pipe.Slice(a).
		Map(func(x int) int { return x * x }).
		Map(func(x int) int { return x + 1 }).
		Filter(func(x int) bool { return x > 100 }).
		Filter(func(x int) bool { return x < 1000 }).
		Do()
		`)
	fmt.Println(res1)

	// walues can be achieved from a function
	res2 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Gen(100).
		// the line below is not working since x is float32 now
		// Filter(func(x int) bool { return x < 1000 }).
		// We need to provide another function:
		Filter(func(x float32) bool { return x > 0.5 }).
		// Generate 100 values as an initial pipeline
		Do()

	fmt.Println("2: creating pipe from a function")
	fmt.Println(`pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
	Filter(func(x float32) bool { return x > 0.5 }).
	Gen(100).
	Do()
		`)
	fmt.Println(res2)

	// We can Take(n) the exact amount of values - it will be executed until the length of the result slice will be n
	res21 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Take(100).
		Filter(func(x float32) bool { return x > 0.5 }).
		// Take 100 values, don't stop generate before get all of them
		Do()

	fmt.Println("2.1: creating pipe from a function but using Take(n) to gen exactly n values")
	fmt.Println(`pipe.Func(func(i int) (float32, bool) {
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
	// you can just count values:

	res3 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Gen(100).
		Filter(func(x float32) bool { return x > 0.6 }).
		Count()
	fmt.Println("3: counting Gen(100) items values")
	fmt.Println(`pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.6 }).
		Gen(100).
		Count()
		`)
	fmt.Println(res3)

	// trying to count values with Take(n int) will return n:
	res31 := pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Take(100).
		Filter(func(x float32) bool { return x > 0.6 }).
		Count()
	fmt.Println("result3.1: trying to count values with Take(n int) will return n")
	fmt.Println(`pipe.Func(func(i int) (float32, bool) {
		rnd := rand.New(rand.NewSource(42))
		rnd.Seed(int64(i))
		return rnd.Float32(), true
	}).
		Filter(func(x float32) bool { return x > 0.6 }).
		Take(100).
		Count()
		`)
	fmt.Println(res31)

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
	fmt.Println("4 iterating and doing ops over 10^7 elems (by-hand): ", cnt, " time ", time.Since(start))
	runtime.GC()

	runtime.GC()
	// you can set the amount of goroutines using Parallel(n int)
	start = time.Now()

	res4 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Gen(1_000_000).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Parallel(12).
		Count()
	fmt.Println("4 (using pipe): (parallel 12)", res4, "; eval took ", time.Since(start))
	runtime.GC()

	// single thread
	start = time.Now()
	res41 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Gen(1_000_000).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Parallel(1).
		Count()
	fmt.Println("4.1: (parallel 1)", res41, "; eval took ", time.Since(start))
	runtime.GC()

	// the defalut value is 4
	start = time.Now()
	res42 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Gen(1_000_000).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Count()
	fmt.Println("4.2: (parallel 4): ", res42, "; eval took ", time.Since(start))
	runtime.GC()

	// many goroutines are OK
	start = time.Now()
	res43 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Gen(1_000_000).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Parallel(150).
		Count()
	fmt.Println("4.3: many goroutines are OK (parallel 150)", res43, "; eval took ", time.Since(start))
	runtime.GC()

	// but the best value shoul be about the amount of your CPUs (do not set it very high)
	// the threshold is 256
	start = time.Now()
	res44 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Gen(1_000_000).
		Map(func(x float32) float32 { return x * x }).
		Filter(func(x float32) bool { return x > 5000.6 }).
		Parallel(65000).
		Count()
	fmt.Println("result44: but too many is not great (parallel 100500 > 256)", res44, "; eval took ", time.Since(start))
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
	fmt.Println("5: another map output type example ", res5)

	// and even Sort them
	res6 := pipe.Func(func(i int) (float32, bool) {
		return float32(i) * 0.9, true
	}).
		Gen(10000).
		Map(func(x float32) float32 { return x * x }).
		// This is how you can sort in parallel (it's rly faster!)
		Parallel(12).
		Sort(pipies.Less[float32]).
		Do()
	fmt.Println("6: Soring in parallel! (first, middle and last item) ", res6[0], res6[len(res6)/2], res6[len(res6)-1])

	res7 := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(10000).
		// This is how you can sort in parallel (it's rly faster!)
		Parallel(12).
		Sum(pipies.Sum[int])
	fmt.Println("7: Sum (all nums from 0 to 10^5-1):", int64(*res7))
	sm := 0
	for i := 0; i < 10000; i++ {
		sm += i
	}
	fmt.Println("7: Sum should be", sm)

	res8 := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(10000).
		Reduce(func(a, b int) int { return a + b })
	fmt.Println("8: Sum with reduce", *res8)

	res81 := pipe.Reduce(
		pipe.Func(func(i int) (User, bool) {
			return User{"Slim", fmt.Sprintf("Shady %d", i), i % 45, nil}, true
		}).
			Gen(6000).
			Filter(func(x User) bool { return x.Age > 43 }),
		func(superName string, u User) string {
			return superName + u.LastName + " "
		},
		"My name is: ",
	)
	fmt.Println("8.1: reduced super name is:", res81)

	res9 := pipe.Slice(a).Map(func(x int) int { return x*x - 2 }).Parallel(100).First()
	fmt.Println("9: ", *res9)

	res91 := pipe.Fn(func(i int) int { return i }).
		Map(func(x int) int { return (x+2)*(x+2) - 2 }).
		Filter(func(x int) bool { return x > 100500 }).
		Parallel(100).
		First()
	fmt.Println("91: ", *res91)
}

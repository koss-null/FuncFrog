package perf_test

import (
	"runtime"
	"testing"

	"github.com/koss-null/funcfrog/pkg/pipe"
)

func fib(_ int) int {
	// about 100 operations to get 91th fib number
	n := 100 // 100 iters
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}

// Define the predicate function for the filter operation
func filterFunc(x *int) bool {
	// Perform a complex condition check
	return fib(*x%2) == 0
}

// Define the binary function for the reduce operation
func reduceFunc(x, y *int) int {
	// Perform a complex calculation or aggregation
	return fib(*x)%1024 + fib(*y)%1024
}

func BenchmarkMap(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input)
		result := pipe.Map(fib).Do()
		_ = result
	}
}

func BenchmarkMapParallel(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input).Parallel(uint16(runtime.NumCPU()))
		result := pipe.Map(fib).Do()
		_ = result
	}
}

func BenchmarkMapFor(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		result := make([]int, 0, len(input))
		for i := range input {
			result = append(result, fib(input[i]))
		}
		_ = result
	}
}

func BenchmarkFilter(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input)
		result := pipe.Filter(filterFunc).Do()
		_ = result
	}
}

func BenchmarkFilterParallel(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input).Parallel(uint16(runtime.NumCPU()))
		result := pipe.Filter(filterFunc).Do()
		_ = result
	}
}

func BenchmarkFilterFor(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		result := make([]int, 0)
		for i := range input {
			if filterFunc(&input[i]) {
				result = append(result, input[i])
			}
		}
		_ = result
	}
}

func BenchmarkReduce(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input)
		result := pipe.Reduce(reduceFunc)
		_ = result
	}
}

func BenchmarkSumParallel(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input).Parallel(uint16(runtime.NumCPU()))
		result := pipe.Sum(reduceFunc)
		_ = result
	}
}

// Benchmark the reduce with for-loop
func BenchmarkReduceFor(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		result := 0
		for i := range input {
			result = reduceFunc(&result, &input[i])
		}
		_ = result
	}
}

func BenchmarkAny(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input).Filter(func(x *int) bool { return *x > 5_000_00 })
		result := pipe.Any()
		_ = result
	}
}

func BenchmarkAnyParallel(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		pipe := pipe.Slice(input).
			Parallel(uint16(runtime.NumCPU())).
			Filter(func(x *int) bool { return *x > 5_000_00 })
		result := pipe.Any()
		_ = result
	}
}

func BenchmarkAnyFor(b *testing.B) {
	b.StopTimer()
	input := make([]int, 1_000_000)
	for i := 0; i < len(input); i++ {
		input[i] = i
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		for i := 0; i < len(input); i++ {
			if input[i] > 5_000_00 {
				_ = input[i]
				break
			}
		}
	}
}

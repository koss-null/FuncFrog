package perf_test

import (
	"runtime"
	"testing"

	"github.com/koss-null/funcfrog/pkg/pipe"
)

func fib(n int) int {
	n = n % 91
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

		// Create a Pipe from the input slice
		pipe := pipe.Slice(input)

		// Apply the map operation to the Pipe
		result := pipe.Map(fib).Do()

		// Perform any necessary assertions on the result
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

		// Create a Pipe from the input slice
		pipe := pipe.Slice(input).Parallel(uint16(runtime.NumCPU()))

		// Apply the map operation to the Pipe
		result := pipe.Map(fib).Do()

		// Perform any necessary assertions on the result
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

		// Perform any necessary assertions on the result
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
		// Create a Pipe from the input slice
		pipe := pipe.Slice(input)

		// Apply the filter operation to the Pipe
		result := pipe.Filter(filterFunc).Do()

		// Perform any necessary assertions on the result
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
		// Create a Pipe from the input slice
		pipe := pipe.Slice(input).Parallel(uint16(runtime.NumCPU()))

		// Apply the filter operation to the Pipe
		result := pipe.Filter(filterFunc).Do()

		// Perform any necessary assertions on the result
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

		// Apply the filter operation to the Pipe
		result := make([]int, 0)
		for i := range input {
			if filterFunc(&input[i]) {
				result = append(result, input[i])
			}
		}

		// Perform any necessary assertions on the result
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

		// Apply the reduce operation to the Pipe
		result := pipe.Reduce(reduceFunc)

		// Perform any necessary assertions on the result
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
		// Create a Pipe from the input slice
		pipe := pipe.Slice(input).Parallel(uint16(runtime.NumCPU()))

		// Apply the reduce operation to the Pipe
		result := pipe.Sum(reduceFunc)

		// Perform any necessary assertions on the result
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
		// Apply the reduce operation to the Pipe
		result := 0
		for i := range input {
			result = reduceFunc(&result, &input[i])
		}

		// Perform any necessary assertions on the result
		_ = result
	}
}

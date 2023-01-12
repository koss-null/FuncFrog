package pipe_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/pkg/pipe"
)

func wrap[T any](x T) func() T {
	return func() T {
		return x
	}
}

var a []int
var mx1 sync.Mutex

func largeSlice() []int {
	mx1.Lock()
	defer mx1.Unlock()

	if a == nil {
		a = make([]int, 1_000_000)
		for i := range a {
			a[i] = i * i
		}
	}
	return a
}

func TestScice(t *testing.T) {
	cases := []struct {
		name     string
		testCase func() []int
		expected func() []int
	}{
		{
			name: "simple",
			testCase: func() []int {
				return pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Do()
			},
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "large",
			testCase: func() []int {
				return pipe.Slice(largeSlice()).Do()
			},
			expected: wrap(largeSlice()),
		},
		{
			name: "empty",
			testCase: func() []int {
				return pipe.Slice([]int{}).Do()
			},
			expected: wrap([]int{}),
		},
		{
			name: "single element",
			testCase: func() []int {
				return pipe.Slice([]int{}).Do()
			},
			expected: wrap([]int{}),
		},
		{
			name: "simple parallel 12",
			testCase: func() []int {
				return pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Parallel(12).Do()
			},
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "large parallel 12",
			testCase: func() []int {
				return pipe.Slice(largeSlice()).Parallel(12).Do()
			},
			expected: wrap(largeSlice()),
		},
		{
			name: "empty parallel 12",
			testCase: func() []int {
				return pipe.Slice([]int{}).Parallel(12).Do()
			},
			expected: wrap([]int{}),
		},
		{
			name: "single element parallel 12",
			testCase: func() []int {
				return pipe.Slice([]int{}).Parallel(12).Do()
			},
			expected: wrap([]int{}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.expected(), c.testCase())
		})
	}
}

func TestFunc(t *testing.T) {
	cases := []struct {
		name     string
		testCase func() *pipe.Pipe[int]
		expected func() []int
	}{
		{
			name: "simple gen",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[i], true
				}).Gen(10)
			},
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "large gen",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return largeSlice()[i], true
				}).Gen(len(largeSlice()))
			},
			expected: wrap(largeSlice()),
		},
		{
			name: "empty gen",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return 0, false
				}).Gen(0)
			},
			expected: wrap([]int{}),
		},
		{
			name: "single element gen",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return 1, true
				}).Gen(1)
			},
			expected: wrap([]int{1}),
		},

		{
			name: "simple take",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[i], true
				}).Take(10)
			},
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "large take",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return largeSlice()[i], true
				}).Take(len(largeSlice()))
			},
			expected: wrap(largeSlice()),
		},
		{
			name: "empty take",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return 0, false
				}).Take(0)
			},
			expected: wrap([]int{}),
		},
		{
			name: "single element take",
			testCase: func() *pipe.Pipe[int] {
				return pipe.Func(func(i int) (int, bool) {
					return 1, true
				}).Take(1)
			},
			expected: wrap([]int{1}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := c.testCase()
			require.Equal(t, c.expected(), p.Do())
		})
		t.Run(c.name+" parallel 12", func(t *testing.T) {
			p := c.testCase()
			require.Equal(t, c.expected(), p.Do())
		})
	}
}

func TestMap(t *testing.T) {
	cases := []struct {
		name     string
		pipe     func() *pipe.Pipe[int]
		expected func() []int
	}{
		{
			name: "slice small",
			pipe: func() *pipe.Pipe[int] {
				return pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
					Map(func(x int) int { return x * 2 })
			},
			expected: wrap([]int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}),
		},
		{
			name: "slice large",
			pipe: func() *pipe.Pipe[int] {
				return pipe.Slice(largeSlice()).
					Map(func(x int) int { return x * 2 })
			},
			expected: func() []int {
				b := make([]int, len(largeSlice()))
				ls := largeSlice()
				for i := range ls {
					b[i] = ls[i] * 2
				}
				return b
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := c.pipe
			require.Equal(t, c.expected(), p().Do())
		})
		t.Run(c.name+" parallel 12", func(t *testing.T) {
			p := c.pipe
			require.Equal(t, c.expected(), p().Do())
		})
	}
}

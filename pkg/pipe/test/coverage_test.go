package pipe_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/internal/primitive/pointer"
	"github.com/koss-null/lambda/pkg/pipe"
)

func wrap[T any](x T) func() T {
	return func() T {
		return x
	}
}

var (
	a   []int
	mx1 sync.Mutex
)

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

var (
	pls *pipe.Pipe[int]
	mx2 sync.Mutex
)

func pipeLargeSlice() pipe.Pipe[int] {
	mx2.Lock()
	defer mx2.Unlock()
	if pls == nil {
		pls = pointer.To(pipe.Slice(largeSlice()))
	}
	return *pls
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
			expected: wrap([]int(nil)),
		},
		{
			name: "single element",
			testCase: func() []int {
				return pipe.Slice([]int{1}).Do()
			},
			expected: wrap([]int{1}),
		},
		{
			name: "simple parallel 12",
			testCase: func() []int {
				return pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Parallel(10).Do()
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
			expected: wrap([]int(nil)),
		},
		{
			name: "single element parallel 12",
			testCase: func() []int {
				return pipe.Slice([]int{1}).Parallel(12).Do()
			},
			expected: wrap([]int{1}),
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
		testCase pipe.Pipe[int]
		expected func() []int
	}{
		{
			name: "simple gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[i], true
			}).Gen(10),
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "large gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return largeSlice()[i], true
			}).Gen(len(largeSlice())),
			expected: wrap(largeSlice()),
		},
		{
			name: "empty gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 0, false
			}).Gen(0),
			expected: wrap([]int(nil)),
		},
		{
			name: "single element gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 1, true
			}).Gen(1),
			expected: wrap([]int{1}),
		},

		{
			name: "simple take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[i], true
			}).Take(10),
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "large take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return largeSlice()[i], true
			}).Take(len(largeSlice())),
			expected: wrap(largeSlice()),
		},
		{
			name: "empty take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 0, false
			}).Take(0),
			expected: wrap([]int(nil)),
		},
		{
			name: "single element take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 1, true
			}).Take(1),
			expected: wrap([]int{1}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.expected(), c.testCase.Do())
		})
		t.Run(c.name+" parallel 12", func(t *testing.T) {
			require.Equal(t, c.expected(), c.testCase.Do())
		})
	}
}

func TestMap(t *testing.T) {
	cases := []struct {
		name  string
		input pipe.Pipe[int]
		f     func(int) int
		want  []int
	}{
		{
			name:  "map double",
			input: pipe.Slice([]int{1, 2, 3}),
			f:     func(i int) int { return i * 2 },
			want:  []int{2, 4, 6},
		},
		{
			name:  "map empty",
			input: pipe.Slice([]int{}),
			f:     func(i int) int { return i },
			want:  []int(nil),
		},
		{
			name:  "map single element",
			input: pipe.Slice([]int{1}),
			f:     func(i int) int { return i },
			want:  []int{1},
		},
		{
			name:  "map many diffeerent elements",
			input: pipe.Slice(largeSlice()),
			f:     func(x int) int { return x * 2 },
			want: func() []int {
				b := make([]int, len(largeSlice()))
				ls := largeSlice()
				for i := range ls {
					b[i] = ls[i] * 2
				}
				return b
			}(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := c.input.Map(c.f).Do()
			require.Equal(t, c.want, res)
		})
		t.Run(c.name+" parallel 12", func(t *testing.T) {
			res := c.input.Map(c.f).Do()
			require.Equal(t, c.want, res)
		})
	}
}

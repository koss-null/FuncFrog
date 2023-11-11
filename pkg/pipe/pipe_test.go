package pipe_test

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/funcfrog/internal/primitive/pointer"
	"github.com/koss-null/funcfrog/pkg/ff"
	"github.com/koss-null/funcfrog/pkg/pipe"
	"github.com/koss-null/funcfrog/pkg/pipies"
)

func TestSlice(t *testing.T) {
	t.Parallel()

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
			expected: wrap([]int{}),
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
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.expected(), c.testCase())
		})
	}
}

func TestFunc(t *testing.T) {
	t.Parallel()
	ls := largeSlice()

	cases := []struct {
		name     string
		testCase pipe.Piper[int]
		expected func() []int
	}{
		{
			name: "Simple_gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[i], true
			}).Gen(10),
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "Large_gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return ls[i], true
			}).Gen(len(ls)),
			expected: wrap(largeSlice()),
		},
		{
			name: "Empty_gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 0, false
			}).Gen(0),
			expected: wrap([]int{}),
		},
		{
			name: "Single_element_gen",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 1, true
			}).Gen(1),
			expected: wrap([]int{1}),
		},
		/////////TAKE
		{
			name: "Simple_take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[i], true
			}).Take(10),
			expected: wrap([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		},
		{
			name: "Large_take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return ls[i], true
			}).Take(len(ls)),
			expected: wrap(largeSlice()),
		},
		{
			name: "Empty_take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 0, false
			}).Take(0),
			expected: wrap([]int{}),
		},
		{
			name: "Single_element_take",
			testCase: pipe.Func(func(i int) (int, bool) {
				return 1, true
			}).Take(1),
			expected: wrap([]int{1}),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.expected(), c.testCase.Do())
		})
		t.Run(c.name+" parallel 12", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.expected(), c.testCase.Do())
		})
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input pipe.Piper[int]
		f     func(int) int
		want  []int
	}{
		{
			name:  "Map_double",
			input: pipe.Slice([]int{1, 2, 3}),
			f:     func(i int) int { return i * 2 },
			want:  []int{2, 4, 6},
		},
		{
			name:  "Map_empty",
			input: pipe.Slice([]int{}),
			f:     func(i int) int { return i },
			want:  []int{},
		},
		{
			name:  "Map_single_element",
			input: pipe.Slice([]int{1}),
			f:     func(i int) int { return i },
			want:  []int{1},
		},
		{
			name:  "Map_many_diffeerent_elements",
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
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			res := c.input.Map(c.f).Do()
			require.Equal(t, c.want, res)
		})
		t.Run(c.name+" parallel 12", func(t *testing.T) {
			t.Parallel()
			res := c.input.Map(c.f).Do()
			require.Equal(t, c.want, res)
		})
	}
}

func TestFirst(t *testing.T) {
	t.Parallel()

	const limit = 100_000

	testCases := []struct {
		name          string
		function      func() *float64
		expectedFirst float64
	}{
		{
			name: "Slice_First_Filtered",
			function: func() *float64 {
				return ff.Map(largeSlice(), func(x int) float64 {
					return float64(x)
				}).
					Filter(func(x *float64) bool { return *x > limit }).
					First()
			},
			expectedFirst: float64(100489),
		},
		{
			name: "Func_First_Limited",
			function: func() *float64 {
				return pipe.Func(func(i int) (float64, bool) { return float64(i), true }).
					Filter(func(x *float64) bool { return *x > 10_000 }).
					Take(limit).
					First()
			},
			expectedFirst: 10001,
		},
		{
			name: "Func_First_No_Limit",
			function: func() *float64 {
				return pipe.Fn(func(i int) float64 { return float64(i) }).
					Filter(func(x *float64) bool { return *x > limit }).
					First()
			},
			expectedFirst: float64(limit) + 1.0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := tc.function()

			require.NotNil(t, s)
			require.Equal(t, tc.expectedFirst, *s)
		})
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()

	genFunc := func(i int) (*float64, bool) {
		if i%10 == 0 {
			return nil, true
		}
		return pointer.To(float64(i)), true
	}

	s := pipe.Map(
		pipe.Func(genFunc).
			Filter(pipies.NotNil[*float64]).
			Take(10_000),
		pointer.From[float64],
	).
		Sum(pipies.Sum[float64])
	require.NotNil(t, s)

	sm := 0
	for i := 1; i < 10000; i++ {
		if i%10 != 0 {
			sm += i
		}
	}
	require.Equal(t, float64(sm), s)
}

func TestReduce(t *testing.T) {
	t.Parallel()

	res := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(6000).
		Reduce(func(a, b *int) int { return *a + *b })

	expected := 0
	for i := 1; i < 6000; i++ {
		expected += i
	}
	require.Equal(t, expected, *res)
}

func TestSum(t *testing.T) {
	res := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(6000).
		Sum(func(a, b *int) int { return *a + *b })

	expected := 0
	for i := 1; i < 6000; i++ {
		expected += i
	}
	require.Equal(t, expected, res)
}

func TestSort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		size     int
		parallel uint16
	}{
		{
			name:     "Simple",
			size:     50_000,
			parallel: 4,
		},
		{
			name:     "Single_thread",
			size:     400_000,
			parallel: 1,
		},
		{
			name:     "Single_thread_tiny",
			size:     3,
			parallel: 1,
		},
		{
			name:     "Single_thread_empty",
			size:     0,
			parallel: 1,
		},
		{
			name:     "MultiThread",
			size:     400_000,
			parallel: 8,
		},
		{
			name:     "tiny",
			size:     3,
			parallel: 8,
		},
		{
			name:     "one",
			size:     1,
			parallel: 8,
		},
		{
			name:     "empty",
			size:     0,
			parallel: 8,
		},
		{
			name:     "smallArray",
			size:     6000,
			parallel: 8,
		},
		{
			name:     "Many_threads",
			size:     6000,
			parallel: 8000,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a, err := readTestData()
			require.Nil(t, err)
			a = a[:tc.size]
			res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(tc.parallel).Do()
			for i := 0; i < len(res)-1; i++ {
				require.LessOrEqual(t, res[i], res[i+1])
			}
		})
	}
}

// helping functions

func wrap[T any](x T) func() T {
	return func() T {
		return x
	}
}

var (
	a   []int
	mx1 sync.Mutex

	pls pipe.Piper[int]
	mx2 sync.Mutex

	testSlice   []int
	mxTestSlice sync.Mutex
)

func largeSlice() []int {
	const largeSize = 400_000
	mx1.Lock()
	defer mx1.Unlock()

	if a == nil {
		a = make([]int, largeSize)
		for i := range a {
			a[i] = i * i
		}
	}
	return a
}

func pipeLargeSlice() pipe.Piper[int] {
	mx2.Lock()
	defer mx2.Unlock()

	if pls == nil {
		pls = pipe.Slice(largeSlice())
	}

	return pls
}

func readTestData() ([]int, error) {
	mxTestSlice.Lock()
	defer mxTestSlice.Unlock()
	if len(testSlice) != 0 {
		return testSlice, nil
	}
	raw, err := os.ReadFile("../../.test_data/test1.txt")
	if err != nil {
		return nil, err
	}
	a := make([]int, 0)
	for _, val := range strings.Split(string(raw), " ") {
		val = strings.Trim(val, "][,")
		ival, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		a = append(a, ival)
	}

	testSlice = a
	return a, nil
}

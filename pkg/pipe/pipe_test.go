package pipe_test

import (
	"errors"
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

// testing collectors

func TestCollect(t *testing.T) {
	t.Parallel()

	p := pipe.Func(func(i int) (int, bool) {
		if i == 4 {
			return 0, false
		}
		return i, true
	}).Take(5).Erase()
	c := pipe.Collect[int](p).Do()
	require.Equal(t, []int{0, 1, 2, 3, 5}, c)
}

func TestCollectNL(t *testing.T) {
	t.Parallel()

	p := pipe.Func(func(i int) (int, bool) {
		if i == 4 {
			return 0, false
		}
		return i, true
	}).Erase()
	c := pipe.CollectNL[int](p).Take(5).Do()
	require.Equal(t, []int{0, 1, 2, 3, 5}, c)
}

// testing constructions

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

func TestFuncP(t *testing.T) {
	t.Parallel()

	const testSize = 10_000
	p := pipe.FuncP(func(i int) (*int, bool) { return &i, true }).Gen(testSize).Do()
	for i := 0; i < testSize; i++ {
		require.Equal(t, i, p[i])
	}
}

func TestCycle(t *testing.T) {
	t.Parallel()

	const testSize = 10_000
	c := pipe.Cycle([]int{0, 1, 2, 3, 4}).Take(testSize).Do()
	for i := 0; i < testSize; i++ {
		require.Equal(t, i%5, c[i])
	}
}

func TestRange(t *testing.T) {
	t.Parallel()

	t.Run("asc", func(t *testing.T) {
		t.Parallel()

		r := pipe.Range(0, 10_000, 5).Do()
		idx := 0
		for i := 0; i < 10_000; i += 5 {
			require.Equal(t, i, r[idx])
			idx++
		}
	})

	t.Run("desc", func(t *testing.T) {
		t.Parallel()

		r := pipe.Range(10_000, 0, -5).Do()
		idx := 0
		for i := 10_000; i > 0; i -= 5 {
			require.Equal(t, i, r[idx])
			idx++
		}
	})

	t.Run("desc2", func(t *testing.T) {
		t.Parallel()

		r := pipe.Range(0, -10_000, -5).Do()
		idx := 0
		for i := 0; i > -10_000; i -= 5 {
			require.Equal(t, i, r[idx])
			idx++
		}
	})
}

func TestRepeat(t *testing.T) {
	t.Parallel()

	const r = 1

	exp := [][]int{{}, {r, r, r, r, r}, {}}
	for i, n := range []int{0, 5, -1} {
		require.Equal(t, exp[i], pipe.Repeat(r, n).Do())
	}
}

// testing pipe and pipeNL functions

func TestMap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   pipe.Piper[int]
		inputNL pipe.PiperNoLen[int]
		f       func(int) int
		want    []int
	}{
		{
			name:  "double",
			input: pipe.Slice([]int{1, 2, 3}),
			f:     func(i int) int { return i * 2 },
			want:  []int{2, 4, 6},
		},
		{
			name:  "empty",
			input: pipe.Slice([]int{}),
			f:     func(i int) int { return i },
			want:  []int{},
		},
		{
			name:  "single_element",
			input: pipe.Slice([]int{1}),
			f:     func(i int) int { return i },
			want:  []int{1},
		},
		{
			name:  "many_diffeerent_elements",
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
		{
			name: "many_diffeerent_elements_fn",
			inputNL: pipe.Func(func(i int) (int, bool) {
				return largeSlice()[i], i < len(largeSlice())
			}),
			f: func(x int) int { return x * 2 },
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

			if c.inputNL != nil {
				res := c.inputNL.Map(c.f).Gen(len(largeSlice())).Do()
				require.Equal(t, c.want, res)
			} else {
				res := c.input.Map(c.f).Do()
				require.Equal(t, c.want, res)
			}
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()

			if c.inputNL != nil {
				res := c.inputNL.Map(c.f).Gen(len(largeSlice())).Parallel(7).Do()
				require.Equal(t, c.want, res)
			} else {
				res := c.input.Map(c.f).Parallel(7).Do()
				require.Equal(t, c.want, res)
			}
		})
	}
}

// FIXME: this test is ok but ugly, need to be refactored
func TestFilter(t *testing.T) {
	t.Parallel()

	genFunc := func(i int) (*float64, bool) {
		if i%10 == 0 {
			return nil, true
		}
		return pointer.To(float64(i)), true
	}

	s := pipe.MapNL(
		pipe.Func(genFunc).
			Filter(pipies.NotNil[*float64]),
		pointer.From[float64],
	).Take(10_000).Sum(pipies.Sum[float64])
	require.NotNil(t, s)

	sm := 0
	a := make([]*float64, 0)
	for i := 0; i < 10000; i++ {
		f := float64(i)
		a = append(a, &f)
		if i%10 != 0 {
			sm += i
		}
	}
	require.Equal(t, float64(sm), s)

	ss := pipe.Map(
		pipe.Slice(a).Map(func(x *float64) *float64 {
			if int(*x)%10 != 0 {
				return x
			}
			return nil
		}).Filter(pipies.NotNil[*float64]),
		pointer.From[float64],
	).Sum(pipies.Sum[float64])
	require.Equal(t, float64(sm), ss)
}

func TestMapFilter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     []int
		gen        int
		funcSource func(int) (int, bool)
		apply      func(pipe.Piper[int]) pipe.Piper[int]
		applyNL    func(pipe.PiperNoLen[int]) pipe.PiperNoLen[int]
		expect     []int
	}{
		{
			name:   "simple",
			source: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p.MapFilter(func(x int) (int, bool) {
					return x * 3, x%2 == 0
				})
			},
			expect: []int{6, 12, 18, 24, 0},
		},
		{
			name: "simple_fn",
			funcSource: func(i int) (int, bool) {
				if i < 10 {
					return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true
				}
				return 0, true
			},
			gen: 10,
			applyNL: func(p pipe.PiperNoLen[int]) pipe.PiperNoLen[int] {
				return p.MapFilter(func(x int) (int, bool) {
					return x * 3, x%2 == 0
				})
			},
			expect: []int{6, 12, 18, 24, 0},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var p pipe.Piper[int]
			if c.funcSource != nil {
				p = c.applyNL(pipe.Func(c.funcSource)).Gen(c.gen)
			} else {
				p = c.apply(pipe.Slice(c.source))
			}

			res := p.Do()
			require.Equal(t, c.expect, res)
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()

			var p pipe.Piper[int]
			if c.funcSource != nil {
				p = c.applyNL(pipe.Func(c.funcSource)).Parallel(7).Gen(c.gen)
			} else {
				p = c.apply(pipe.Slice(c.source)).Parallel(7)
			}

			res := p.Do()
			require.Equal(t, c.expect, res)
		})
	}
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

func TestReduce(t *testing.T) {
	t.Parallel()

	res := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(6000).
		Reduce(func(a, b *int) int { return *a + *b })

	expected := 0
	a := make([]int, 0)
	for i := 0; i < 6000; i++ {
		expected += i
		a = append(a, i)
	}
	require.Equal(t, expected, *res)

	res = pipe.Slice(a).Reduce(func(a, b *int) int { return *a + *b })
	require.Equal(t, expected, *res)
}

func TestSum(t *testing.T) {
	t.Parallel()

	res := pipe.Func(func(i int) (int, bool) {
		return i, true
	}).
		Gen(6000).
		Parallel(6000).
		Sum(func(a, b *int) int { return *a + *b })

	expected := 0
	a := make([]int, 0)
	for i := 0; i < 6000; i++ {
		expected += i
		a = append(a, i)
	}

	require.Equal(t, expected, res)

	res = pipe.Slice(a).Sum(func(a, b *int) int { return *a + *b })
	require.Equal(t, expected, res)
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

func TestAny(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     []int
		take       int
		funcSource func(int) (int, bool)
		apply      func(pipe.Piper[int]) pipe.Piper[int]
		applyNL    func(pipe.PiperNoLen[int]) pipe.PiperNoLen[int]
		expect     []int
	}{
		{
			name:   "simple",
			source: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
		{
			name: "simple_fn",
			funcSource: func(i int) (int, bool) {
				if i < 10 {
					return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true
				}
				return 0, false
			},
			take: 10,
			applyNL: func(p pipe.PiperNoLen[int]) pipe.PiperNoLen[int] {
				return p
			},
			expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			var p pipe.Piper[int]
			if c.funcSource != nil {
				rs := c.applyNL(pipe.Func(c.funcSource)).Any()
				require.Contains(t, c.expect, *rs)
				return
			}

			p = c.apply(pipe.Slice(c.source))
			res := p.Any()
			require.Contains(t, c.expect, *res)
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()

			var p pipe.Piper[int]
			if c.funcSource != nil {
				rs := c.applyNL(pipe.Func(c.funcSource)).Parallel(7).Any()
				require.Contains(t, c.expect, *rs)
				return
			}

			p = c.apply(pipe.Slice(c.source))
			res := p.Parallel(7).Any()
			require.Contains(t, c.expect, *res)
		})
	}
}

func TestCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     []int
		take       int
		funcSource func(int) (int, bool)
		apply      func(pipe.Piper[int]) pipe.Piper[int]
		expect     int
	}{
		{
			name:   "simple",
			source: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			expect: 10,
		},
		{
			name:   "zero",
			source: []int{},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			expect: 0,
		},
		{
			name:       "simple_fn",
			funcSource: func(i int) (int, bool) { return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true },
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			take:   10,
			expect: 10,
		},
		{
			name:       "zero_fn",
			funcSource: func(i int) (int, bool) { return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true },
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			take:   0,
			expect: 0,
		},
	}

	for _, c := range cases {
		c := c

		var p pipe.Piper[int]
		if c.funcSource != nil {
			p = c.apply(pipe.Func(c.funcSource).Take(c.take))
		} else {
			p = c.apply(pipe.Slice(c.source))
		}

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			res := p.Count()
			require.Equal(t, c.expect, res)
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()
			res := p.Parallel(7).Count()
			require.Equal(t, c.expect, res)
		})
	}
}

func TestPromices(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     []int
		take       int
		funcSource func(int) (int, bool)
		apply      func(pipe.Piper[int]) pipe.Piper[int]
		expect     []int
	}{
		{
			name:   "simple",
			source: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
		{
			name:   "zero",
			source: []int{},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			expect: []int{},
		},
		{
			name:       "simple_fn",
			funcSource: func(i int) (int, bool) { return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true },
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				return p
			},
			take:   10,
			expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
	}

	for _, c := range cases {
		c := c

		var p pipe.Piper[int]
		if c.funcSource != nil {
			p = c.apply(pipe.Func(c.funcSource).Take(c.take))
		} else {
			p = c.apply(pipe.Slice(c.source))
		}

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			res := p.Promices()
			resAr := make([]int, len(res))
			for i := range res {
				x, _ := res[i]()
				resAr[i] = x
			}
			require.Equal(t, c.expect, resAr)
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()

			res := p.Parallel(7).Promices()
			resAr := make([]int, len(res))
			for i := range res {
				x, _ := res[i]()
				resAr[i] = x
			}
			require.Equal(t, c.expect, resAr)
		})
	}
}

func TestErase(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     []int
		take       int
		funcSource func(int) (int, bool)
		apply      func(pipe.Piper[any]) pipe.Piper[any]
		expect     []int
	}{
		{
			name:   "simple",
			source: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			apply: func(p pipe.Piper[any]) pipe.Piper[any] {
				return p
			},
			expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
		{
			name:   "zero",
			source: []int{},
			apply: func(p pipe.Piper[any]) pipe.Piper[any] {
				return p
			},
			expect: []int{},
		},
		{
			name:       "simple_fn",
			funcSource: func(i int) (int, bool) { return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true },
			take:       10,
			apply: func(p pipe.Piper[any]) pipe.Piper[any] {
				return p
			},
			expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
		},
	}

	for _, c := range cases {
		c := c

		var p pipe.Piper[any]
		if c.funcSource != nil {
			p = c.apply(pipe.Func(c.funcSource).Erase().Take(c.take))
		} else {
			p = c.apply(pipe.Slice(c.source).Erase())
		}

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			res := p.Do()
			resAr := make([]int, len(res))
			for i := range res {
				resAr[i] = *(res[i].(*int))
			}
			require.Equal(t, c.expect, resAr)
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()

			res := p.Parallel(7).Do()
			resAr := make([]int, len(res))
			for i := range res {
				resAr[i] = *(res[i].(*int))
			}
			require.Equal(t, c.expect, resAr)
		})
	}
}

func TestYetiSnag(t *testing.T) {
	t.Parallel()

	randErr := errors.New("test err")
	simpleTestErr := errors.New("simple")
	simpleFnTestErr := errors.New("simpleFn")

	mx := sync.Mutex{}
	sharedCnt := 0
	shared := make(map[int]any)

	cases := []struct {
		name       string
		source     []int
		take       int
		funcSource func(int) (int, bool)
		apply      func(pipe.Piper[int]) pipe.Piper[int]
		applyNL    func(pipe.PiperNoLen[int]) pipe.PiperNoLen[int]
		expect     func(*testing.T)
	}{
		{
			name:   "simple",
			source: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			apply: func(p pipe.Piper[int]) pipe.Piper[int] {
				y := pipe.NewYeti()
				return p.
					Yeti(y).
					Map(func(i int) int {
						if i == 5 {
							y.Yeet(errors.Join(simpleTestErr, randErr))
						}
						return i
					}).
					Snag(func(e error) {
						mx.Lock()
						sharedCnt++
						shared[sharedCnt] = e
						mx.Unlock()
					})
			},
			expect: func(t *testing.T) {
				mx.Lock()
				defer mx.Unlock()
				found := false
				for _, v := range shared {
					er := v.(error)
					if errors.Is(er, simpleTestErr) {
						found = true
						break
					}
				}
				require.True(t, found, "simple test case error was not found")
			},
		},
		{
			name:       "simple_fn",
			funcSource: func(i int) (int, bool) { return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}[i], true },
			take:       10,
			applyNL: func(p pipe.PiperNoLen[int]) pipe.PiperNoLen[int] {
				y := pipe.NewYeti()
				return p.
					Yeti(y).
					Map(func(i int) int {
						if i == 5 {
							y.Yeet(errors.Join(simpleFnTestErr, randErr))
						}
						return i
					}).
					Snag(func(e error) {
						mx.Lock()
						sharedCnt++
						shared[sharedCnt] = e
						mx.Unlock()
					})
			},
			expect: func(t *testing.T) {
				mx.Lock()
				defer mx.Unlock()
				found := false
				for _, v := range shared {
					er := v.(error)
					if errors.Is(er, simpleFnTestErr) {
						found = true
						break
					}
				}
				require.True(t, found, "simple test case error was not found")
			},
		},
	}

	for _, c := range cases {
		c := c

		var p pipe.Piper[int]
		if c.funcSource != nil {
			p = c.applyNL(pipe.Func(c.funcSource)).Take(c.take)
		} else {
			p = c.apply(pipe.Slice(c.source))
		}

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			_ = p.Do()
			c.expect(t)
		})
		t.Run(c.name+"_parallel", func(t *testing.T) {
			t.Parallel()

			_ = p.Parallel(7).Do()
			c.expect(t)
		})
	}
}

// prefixpipe

func TestPrefixMap(t *testing.T) {
	res := pipe.Map(
		pipe.Slice([]int{1, 2, 3, 4}).
			Filter(func(x *int) bool {
				return *x != 2
			}),
		func(x int) string {
			return strconv.Itoa(x)
		},
	).Do()
	require.Equal(t, []string{"1", "3", "4"}, res)
}

func TestPrefixMapNL(t *testing.T) {
	t.Parallel()

	res := pipe.MapNL(
		pipe.Func(func(i int) (int, bool) {
			return []int{1, 2, 3, 4}[i], true
		}).Filter(func(x *int) bool {
			return *x != 2
		}),
		func(x int) string {
			return strconv.Itoa(x)
		},
	).Take(3).Do()
	require.Equal(t, []string{"1", "3", "4"}, res)
}

func TestPrefixReduce(t *testing.T) {
	t.Parallel()

	t.Run("common", func(t *testing.T) {
		t.Parallel()
		res := pipe.Reduce(
			pipe.Slice([]int{1, 2, 3, 4, 5}),
			func(s *string, n *int) string {
				return *s + strconv.Itoa(*n)
			},
			"the result string is: ",
		)
		require.Equal(t, "the result string is: 12345", res)
	})

	t.Run("zero_res", func(t *testing.T) {
		t.Parallel()
		res := pipe.Reduce(
			pipe.Slice([]int{1, 2, 3, 4, 5}).
				Filter(func(x *int) bool { return *x > 100 }),
			func(s *string, n *int) string {
				return *s + strconv.Itoa(*n)
			},
			"the result string is: ",
		)
		require.Equal(t, "the result string is: ", res)
	})

	t.Run("single_res", func(t *testing.T) {
		t.Parallel()
		res := pipe.Reduce(
			pipe.Slice([]int{1, 2, 3, 4, 5}).
				Filter(func(x *int) bool { return *x == 5 }),
			func(s *string, n *int) string {
				return *s + strconv.Itoa(*n)
			},
			"the result string is: ",
		)
		require.Equal(t, "the result string is: 5", res)
	})
}

// functype test

func TestAcc(t *testing.T) {
	t.Parallel()

	res := pipe.Slice([]int{1, 2, 3, 4, 5}).Reduce(
		pipe.Acc(func(x int, y int) int {
			return x + y
		}),
	)
	require.Equal(t, 15, *res)
}

func TestPred(t *testing.T) {
	t.Parallel()

	res := pipe.Slice([]int{1, 2, 3, 4, 5}).Filter(
		pipe.Pred(func(x int) bool {
			return x != 2
		}),
	).Do()
	require.Equal(t, []int{1, 3, 4, 5}, res)
}

func TestComp(t *testing.T) {
	t.Parallel()

	res := pipe.Slice([]int{5, 3, 4, 1, 3, 5, 2, 3, 6}).Sort(
		pipe.Comp(func(x, y int) bool { return x < y }),
	).Do()
	require.Equal(t, []int{1, 2, 3, 3, 3, 4, 5, 5, 6}, res)
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

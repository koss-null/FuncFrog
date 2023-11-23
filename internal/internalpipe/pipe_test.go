package internalpipe

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_min(t *testing.T) {
	t.Parallel()
	require.Equal(t, min(1, 2), 1)
	require.Equal(t, min(2, 1), 1)
	require.Equal(t, min(1.1, 1.11), 1.1)
	require.Equal(t, min(0, -1), -1)
	require.Equal(t, min(0.0, -0.1), -0.1)
	require.Equal(t, min(0.0, 0.0), 0.0)
}

func Test_max(t *testing.T) {
	t.Parallel()
	require.Equal(t, max(1, 2), 2)
	require.Equal(t, max(1.1, 1.11), 1.11)
	require.Equal(t, max(0, -1), 0)
	require.Equal(t, max(0.0, -0.1), 0.0)
	require.Equal(t, max(0.0, 0.0), 0.0)
}

func Test_divUp(t *testing.T) {
	t.Parallel()
	require.Equal(t, divUp(121, 1), 121)
	require.Equal(t, divUp(121, 2), 61)
	require.Equal(t, divUp(5, 5), 1)
	require.Equal(t, divUp(5, 4), 2)
	require.Equal(t, divUp(5, 100000000), 1)
	require.Equal(t, divUp(1, 1345), 1)
}

func Test_genTickets(t *testing.T) {
	t.Parallel()
	require.Equal(t, len(genTickets(100)), 100)
}

func Test_do(t *testing.T) {
	t.Parallel()

	p := Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           10000,
		ValLim:        -1,
		GoroutinesCnt: 5,
	}
	res := p.Do()
	for i := 0; i < 10000; i++ {
		require.Equal(t, i, res[i])
	}
}

func Test_Parallel(t *testing.T) {
	t.Parallel()

	p := Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           10000,
		ValLim:        -1,
		GoroutinesCnt: defaultParallelWrks,
	}
	p = p.Parallel(10)
	require.Equal(t, 10, p.GoroutinesCnt)
	p = p.Parallel(5) // second one should not be applied
	require.Equal(t, 10, p.GoroutinesCnt)

	p = Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           10000,
		ValLim:        -1,
		GoroutinesCnt: defaultParallelWrks,
	}
	p = p.Parallel(0)
	require.Equal(t, defaultParallelWrks, p.GoroutinesCnt)
	p = p.Parallel(4)
	require.Equal(t, 4, p.GoroutinesCnt)
}

func Test_Take(t *testing.T) {
	t.Parallel()

	p := Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           -1,
		ValLim:        -1,
		GoroutinesCnt: 5,
	}
	p = p.Take(10)
	require.Equal(t, 10, p.limit())
	p = p.Take(5)
	require.Equal(t, 10, p.limit())
	p = p.Gen(5)
	require.Equal(t, 10, p.limit())

	p = Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           -1,
		ValLim:        -1,
		GoroutinesCnt: 5,
	}
	p = p.Take(-1)
	require.Equal(t, math.MaxInt-1, p.limit())
	p = p.Take(0)
	require.Equal(t, 0, p.limit())
	p = p.Take(4)
	require.Equal(t, 0, p.limit())
}

func Test_Gen(t *testing.T) {
	t.Parallel()

	p := Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           -1,
		ValLim:        -1,
		GoroutinesCnt: 5,
	}
	p = p.Gen(10)
	require.Equal(t, 10, p.limit())
	p = p.Gen(5)
	require.Equal(t, 10, p.limit())
	p = p.Take(5)
	require.Equal(t, 10, p.limit())

	p = Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           -1,
		ValLim:        -1,
		GoroutinesCnt: 5,
	}
	p = p.Gen(-1)
	require.Equal(t, math.MaxInt-1, p.limit())
	p = p.Gen(0)
	require.Equal(t, 0, p.limit())
	p = p.Gen(4)
	require.Equal(t, 0, p.limit())
}

func Test_Count(t *testing.T) {
	t.Parallel()

	p := Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           10000,
		ValLim:        -1,
		GoroutinesCnt: 5,
	}
	require.Equal(t, 10000, p.Count())

	p = Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           10000,
		ValLim:        -1,
		GoroutinesCnt: 1,
	}
	require.Equal(t, 10000, p.Count())

	p = Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           -1,
		ValLim:        10000,
		GoroutinesCnt: 7,
	}
	require.Equal(t, 10000, p.Count())

	p = Pipe[int]{
		Fn: func(i int) (*int, bool) {
			return &i, false
		},
		Len:           -1,
		ValLim:        10000,
		GoroutinesCnt: 1,
	}
	require.Equal(t, 10000, p.Count())
}

func Test_limit(t *testing.T) {
	t.Parallel()

	p := Pipe[int]{
		Len:    10000,
		ValLim: -1,
	}
	require.Equal(t, 10000, p.limit())
	p = Pipe[int]{
		Len:    -1,
		ValLim: 10000,
	}
	require.Equal(t, 10000, p.limit())
	p = Pipe[int]{
		Len:    -1,
		ValLim: -1,
	}
	require.Equal(t, math.MaxInt-1, p.limit())
}

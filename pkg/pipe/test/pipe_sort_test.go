package pipe_test

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/koss-null/lambda/pkg/pipe"
	"github.com/koss-null/lambda/pkg/pipies"
)

var (
	testSlice []int
	mx        sync.Mutex
)

func readTestData() ([]int, error) {
	mx.Lock()
	defer mx.Unlock()
	if len(testSlice) != 0 {
		return testSlice, nil
	}
	raw, err := os.ReadFile("../../../.test_data/test1.txt")
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

func Test_Sort(t *testing.T) {
	a, err := readTestData()
	require.Nil(t, err)
	require.Equal(t, len(a), 1000000)
	res := pipe.Slice(a).Sort(pipies.Less[int]).Do()
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1])
	}
}

func Test_Sort_singleThread(t *testing.T) {
	a, err := readTestData()
	require.Nil(t, err)
	require.Equal(t, len(a), 1000000)
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(1).Do()
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1])
	}
}

func Test_Sort_multiThread(t *testing.T) {
	a, err := readTestData()
	require.Nil(t, err)
	require.Equal(t, len(a), 1000000)
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(8).Do()
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1])
	}
}

func Test_Sort_minArray(t *testing.T) {
	a, err := readTestData()
	require.Nil(t, err)
	a = a[0:5001]
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(8).Do()
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1])
	}
}

func Test_Sort_tiny(t *testing.T) {
	a := []int{4, 2, 1}
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(8).Do()
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1])
	}
}

func Test_Sort_one(t *testing.T) {
	a := []int{1}
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(8).Do()
	require.Equal(t, len(res), 1)
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1]) // does not run
	}
}

func Test_Sort_empty(t *testing.T) {
	a := []int{}
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(8).Do()
	require.Equal(t, len(res), 0)
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1]) // does not run
	}
}

func Test_Sort_smallArray(t *testing.T) {
	a, err := readTestData()
	require.Nil(t, err)
	a = a[0:6000]
	res := pipe.Slice(a).Sort(pipies.Less[int]).Parallel(8).Do()
	for i := 0; i < len(res)-1; i++ {
		require.LessOrEqual(t, res[i], res[i+1])
	}
}

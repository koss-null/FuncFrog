package ff

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompose(t *testing.T) {
	fn1 := func(x int) int {
		return x + 1
	}

	fn2 := func(x int) string {
		return strconv.Itoa(x)
	}

	composedFn := Compose(fn1, fn2)

	result := composedFn(5)
	require.Equal(t, "6", result, "Unexpected result for Compose")

	result = composedFn(10)
	require.Equal(t, "11", result, "Unexpected result for Compose")
}

func TestMap(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}

	fn := func(x int) string {
		return strconv.Itoa(x)
	}

	piper := Map(a, fn).Do()

	// Iterate through the values and test the output
	expected := []string{"1", "2", "3", "4", "5"}
	for i, val := range piper {
		require.Equal(t, expected[i], val, "Unexpected result for Map")
	}
}

func TestMapFilter(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8}

	fn := func(x int) (string, bool) {
		return strconv.Itoa(x), x%2 == 0
	}

	piper := MapFilter(a, fn).Do()

	// Iterate through the values and test the output
	expected := []string{"2", "4", "6", "8"}
	for i, val := range piper {
		require.Equal(t, expected[i], val, "Unexpected result for MapFilter")
	}
}

func TestFilter(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}

	fn := func(x *int) bool {
		return *x%2 != 0
	}

	piper := Filter(a, fn).Do()

	// Iterate through the values and test the output
	expected := []int{1, 3, 5}
	for i, val := range piper {
		require.Equal(t, expected[i], val, "Unexpected result for Filter")
	}
}

func TestReduce(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}

	sum := func(result *int, x *int) int {
		return *result + *x
	}

	result := Reduce(a, sum)

	expected := 15
	require.Equal(t, expected, result, "Unexpected result for Reduce")
}

package batch_test

import (
	"testing"

	"github.com/koss-null/lambda/internal/algo/batch"
	"github.com/stretchr/testify/require"
)

func Test_BatchDo(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7}
	b := batch.Do(a, 3)
	require.Equal(t, len(b), 3)
	require.Equal(t, b[0], []int{1, 2, 3})
	require.Equal(t, b[1], []int{4, 5, 6})
	require.Equal(t, b[2], []int{7})
}

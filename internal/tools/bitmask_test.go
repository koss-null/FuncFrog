package tools_test

import (
	"testing"

	"github.com/koss-null/lambda-go/internal/tools"
	"github.com/stretchr/testify/assert"
)

func Test_PutLine(t *testing.T) {
	bm := tools.Bitmask{}
	bm.PutLine(0, 100500, true)
	assert.True(t, bm.Get(uint(1000)))
}

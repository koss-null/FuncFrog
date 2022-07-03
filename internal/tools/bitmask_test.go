package tools_test

import (
	"strconv"
	"testing"

	"github.com/koss-null/lambda-go/internal/tools"
	"github.com/stretchr/testify/require"
)

func toString(bm *tools.Bitmask) string {
	s := ""
	cnt, val := 0, false
	for {
		p, r := bm.Next()
		if p == -1 {
			break
		}
		if val == r {
			cnt++
			continue
		}
		val = r
		if cnt == 0 {
			cnt++
			continue
		}
		s += strconv.Itoa(cnt)
		cnt = 0
		if val {
			s += "f"
			continue
		}
		s += "t"
	}
	return s
}

func Test_PutLine(t *testing.T) {
	bm := tools.Bitmask{}
	bm.PutLine(0, 100500, true)
	require.Equal(t, "100500t", toString(&bm))
}

func Test_PutLine2(t *testing.T) {
	bm := tools.Bitmask{}
	bm.PutLine(0, 100500, true)
	bm.PutLine(100, 500, false)

	require.Equal(t, "100t399f99999t", toString(&bm))
}

func Test_PutLine3(t *testing.T) {
	bm := tools.Bitmask{}
	bm.PutLine(0, 100500, true)
	bm.PutLine(100, 500, false)
	bm.PutLine(1000, 2500, false)

	require.Equal(t, "100t399f499t1499f97999t", toString(&bm))
}

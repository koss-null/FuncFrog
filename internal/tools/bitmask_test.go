package tools

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func toString[T any](bm *Bitmask[T]) string {
	s := ""
	cnt, val := 0, false
	bm = bm.Copy(0, uint(bm.Len()))
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

func Test_getBit(t *testing.T) {
	bm := Bitmask[int]{}
	for i := uint(0); i < 12; i++ {
		require.Equal(t, bm.getBit(uint64(1)<<12, i), false)
	}
	require.Equal(t, bm.getBit(uint64(1)<<12, 12), true)
	require.Equal(t, bm.getBit(uint64(1)<<12, 13), false)
}

func Test_setTrue(t *testing.T) {
	bm := Bitmask[int]{mask: []uint64{4096}}
	bm.setTrue(0, 5)
	require.Equal(t, uint64(4096+32), bm.mask[0])
}

func Test_setTrue2(t *testing.T) {
	bm := Bitmask[int]{mask: []uint64{4096}}
	bm.setTrue(0, 5)
	bm.setTrue(0, 6)
	bm.setTrue(0, 7)
	require.Equal(t, uint64(4096+32+64+128), bm.mask[0])
}

func Test_setFalse(t *testing.T) {
	bm := Bitmask[int]{mask: []uint64{4095}}
	bm.setFalse(0, 5)
	require.Equal(t, uint64(4095-32), bm.mask[0])
}

func Test_setFalse2(t *testing.T) {
	bm := Bitmask[int]{mask: []uint64{4095}}
	bm.setFalse(0, 5)
	bm.setFalse(0, 6)
	bm.setFalse(0, 7)
	require.Equal(t, uint64(4095-32-64-128), bm.mask[0])
}

const benchLen = 100500

func initBench() *Bitmask[int] {
	a := make([]uint64, benchLen)
	for i := range a {
		a[i] = uint64(9061 + i)
	}
	return &Bitmask[int]{mask: a}
}

// func Test_settersBenchmark(b *testing.B) {
// 	bm := initBench()
// 	for i := 0; i < b.N; i++ {
// 		for j := uint(0); j < 100500; j++ {
// 			for k := uint(0); k < maskElemLen; k += 2 {
// 				bm.setTrue(j, k)
// 				bm.setFalse(j, k+1)
// 			}
// 		}
// 	}
// }

func Test_settersBenchmark(t *testing.T) {
	bm := initBench()
	const n = 50
	start := time.Now()
	for i := 0; i < n; i++ {
		for j := uint(0); j < benchLen; j++ {
			for k := uint(0); k < maskElemLen; k += 2 {
				bm.setTrue(j, k)
				bm.setFalse(j, k+1)
			}
		}
	}
	finish := time.Now()

	startNoFn := time.Now()
	for i := 0; i < n; i++ {
		for j := uint(0); j < benchLen; j++ {
			for k := uint(0); k < maskElemLen; k += 2 {
				bm.mask[j] |= 1 << k
				bm.mask[j] = (bm.mask[j] | (1 << ((k + 1) % maskElemLen))) - (1 << ((k + 1) % maskElemLen))
			}
		}
	}
	finishNoFn := time.Now()

	dif1 := finish.Sub(start).Nanoseconds()
	dif2 := finishNoFn.Sub(startNoFn).Nanoseconds()

	fmt.Println(dif1, dif2)
	// Time should not be different more than 8%
	require.Less(t, (dif1 - dif2), dif1>>3)
	// An actual result is a slow down about 2-3%
}

func Test_PutLine(t *testing.T) {
	bm := Bitmask[int]{}
	bm.PutLine(0, 100500, true)
	require.Equal(t, "100500t", toString(&bm))
}

func Test_PutLine2(t *testing.T) {
	bm := Bitmask[int]{}
	bm.PutLine(0, 100500, true)
	bm.PutLine(100, 500, false)

	require.Equal(t, "100t399f99999t", toString(&bm))
}

func Test_PutLine3(t *testing.T) {
	bm := Bitmask[int]{}
	bm.PutLine(0, 100500, true)
	bm.PutLine(100, 500, false)
	bm.PutLine(1000, 2500, false)

	require.Equal(t, "100t399f499t1499f97999t", toString(&bm))
}

func Test_CaSBorder(t *testing.T) {
	bm := Bitmask[int]{}
	bm.PutLine(0, 100500, true)
	require.Equal(t, "100500t", toString(&bm))

	bm.PutLine(100, 200, false)
	fmt.Printf("%b %b %b\n\n", bm.mask[1], bm.mask[2], bm.mask[3])
	require.Equal(t, "100t99f100299t", toString(&bm))

	require.True(t, bm.CaSBorder(0, false, 97))
	require.Equal(t, "197t2f100299t", toString(&bm))
}

// Backwards works only with true->false transition yet
func Test_CaSBorderBw(t *testing.T) {
	bm := Bitmask[int]{}
	bm.PutLine(0, 100500, true)
	require.Equal(t, "100500t", toString(&bm))

	bm.PutLine(100, 200, false)
	bm.PutLine(300, 401, false)
	require.Equal(t, "100t99f99t100f100098t", toString(&bm))

	require.True(t, bm.CaSBorderBw(97))
	require.Equal(t, "100t99f99t100f100001t", toString(&bm))
}

// Backwards works only with true->false transition yet
func Test_CaSBorderBw2(t *testing.T) {
	bm := Bitmask[int]{}
	bm.PutLine(0, 401, true)
	require.Equal(t, "401t", toString(&bm))

	bm.PutLine(100, 200, false)
	bm.PutLine(300, 401, false)
	require.Equal(t, "100t99f99t", toString(&bm))

	require.True(t, bm.CaSBorderBw(97))
	require.Equal(t, "100t99f2t", toString(&bm))
}

// TODO: uncomment
// func Test_Apply(t *testing.T) {
// 	bm := Bitmask[int]{}
// 	bm.PutLine(0, 401, true)
// 	require.Equal(t, "401t", toString(&bm))

// 	bm.PutLine(100, 200, false)
// 	bm.PutLine(300, 401, false)
// 	require.Equal(t, "100t99f99t", toString(&bm))

// 	a := make([]int, 401)
// 	for i := range a {
// 		a[i] = i
// 	}

// 	b := bm.Apply(a)
// 	require.EqualValues(t, append(a[:100], a[200:300]...), b)
// }

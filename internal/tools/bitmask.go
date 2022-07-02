package tools

import (
	"fmt"
	"math/bits"
	"sync"
)

const (
	maskElemLen = 64
	u0          = uint(0)
	u64max      = ^uint64(0) // all ones
)

type Bitmask struct {
	mask []uint64
	cur  uint
	cpMx sync.Mutex
}

func (bm *Bitmask) Put(place uint, val bool) {
	fmt.Println(place, val)
	if place/maskElemLen > uint(len(bm.mask)) {
		// it's funny, but it seems we can do like this
		// (don't insert 0 if no place for it yet)
		if val {
			for i := 0; i < len(bm.mask); i++ {
				bm.mask = append(bm.mask, 0)
			}
			bm.Put(place, val)
		}
		return
	}

	if val {
		bm.mask[place/maskElemLen] |= 1 << (place % maskElemLen)
		return
	}

	fmt.Printf("before %b\n", bm.mask[place/maskElemLen])
	// here we know that on place position there is 1 and val is false, so:
	bm.mask[place/maskElemLen] |= 1 << (place % maskElemLen)
	bm.mask[place/maskElemLen] -= 1 << (place % maskElemLen)
	fmt.Printf(" after %b\n", bm.mask[place/maskElemLen])
}

// PutLine does Put for the elements in range [lf, rg) with the val
func (bm *Bitmask) PutLine(lf, rg uint, val bool) {
	if lf > rg {
		return
	}
	if lf == rg {
		bm.Put(lf, val)
		return
	}

	lfBlock := lf / maskElemLen
	rgBlock := rg / maskElemLen

	for rgBlock >= uint(len(bm.mask)) {
		bm.mask = append(bm.mask, 0)
	}
	if val {
		if lfBlock == rgBlock {
			for place := lf; place < rg; place++ {
				bm.mask[lfBlock] |= 1 << (place % maskElemLen)
			}
			return
		}

		for place := lf; place < maskElemLen; place++ {
			bm.mask[lfBlock] |= 1 << place
		}
		for block := lfBlock + 1; block < rgBlock; block++ {
			bm.mask[block] = u64max
		}
		for place := u0; place < rg%maskElemLen; place++ {
			bm.mask[rgBlock] |= 1 << place
		}
		return
	}

	if lfBlock == rgBlock {
		for place := lf; place < rg; place++ {
			bm.mask[lfBlock] |= 1 << place
			bm.mask[lfBlock] -= 1 << place
		}
		return
	}

	// can do it faster
	for place := lf; place < maskElemLen; place++ {
		bm.mask[lfBlock] |= 1 << place
		bm.mask[lfBlock] -= 1 << place
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = 0
	}
	for place := u0; place < rg%maskElemLen; place++ {
		bm.mask[lfBlock] |= 1 << place
		bm.mask[rgBlock] -= 1 << place
	}
	return
}

func (bm *Bitmask) Get(place uint) bool {
	if place/maskElemLen > uint(len(bm.mask)) {
		return false
	}
	return (bm.mask[place/maskElemLen]>>(place%maskElemLen))&1 == 1
}

// CaS compare and swap, returns true in case of success
// despite the title, the operation is not atomic
// since we are dealing with the bool, dst = !src
func (bm *Bitmask) CaS(place uint, src bool) bool {
	if place/maskElemLen > uint(len(bm.mask)) {
		return false
	}
	if ((bm.mask[place/maskElemLen]>>(place%maskElemLen))&1 == 1) == src {
		if src {
			bm.mask[place/maskElemLen] -= 1 << (place % maskElemLen)
			return true
		}
		bm.mask[place/maskElemLen] |= 1 << (place % maskElemLen)
		return true
	}
	return false
}

// CaSLine compares all values on the interval with src and
// changes to !src if the val is equal
// returns the amount of changed instances
func (bm *Bitmask) CaSLine(lf, rg uint, src bool) int {
	if lf > rg {
		return 0
	}
	if lf == rg {
		if bm.CaS(lf, src) {
			return 1
		}
		return 0
	}

	lfBlock := lf / maskElemLen
	rgBlock := rg / maskElemLen
	if rgBlock >= uint(len(bm.mask)) {
		return 0
	}

	cnt := 0
	if src {
		if lfBlock == rgBlock {
			for place := lf; place < rg; place++ {
				maskBit := (bm.mask[lfBlock]>>(place%maskElemLen))&1 == 1
				if maskBit == src {
					bm.mask[lfBlock] |= 1 << (place % maskElemLen)
					cnt++
				}
			}
			return cnt
		}

		for place := lf; place < maskElemLen; place++ {
			maskBit := (bm.mask[lfBlock]>>place)&1 == 1
			if maskBit {
				bm.mask[lfBlock] |= 1 << place
				cnt++
			}
		}
		for block := lfBlock + 1; block < rgBlock; block++ {
			cnt += bits.OnesCount64(bm.mask[block])
			bm.mask[block] = u64max
		}
		for place := u0; place < rg%maskElemLen; place++ {
			maskBit := (bm.mask[lfBlock]>>place)&1 == 1
			if maskBit {
				bm.mask[rgBlock] |= 1 << place
				cnt++
			}
		}
		return cnt
	}

	if lfBlock == rgBlock {
		for place := lf; place < rg; place++ {
			bm.mask[lfBlock] |= 1 << place
			bm.mask[lfBlock] -= 1 << place
			cnt++
		}
		return cnt
	}

	// can do it faster
	for place := lf; place < maskElemLen; place++ {
		bm.mask[lfBlock] |= 1 << place
		bm.mask[lfBlock] -= 1 << place
		cnt++
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		cnt += bits.OnesCount64(bm.mask[block])
		bm.mask[block] = 0
	}
	for place := u0; place < rg%maskElemLen; place++ {
		bm.mask[lfBlock] |= 1 << place
		bm.mask[rgBlock] -= 1 << place
		cnt++
	}
	return cnt
}

// CaSBorder compares all values starting from lf with src and
// changes to !src if the val is equal. runs until th changes is done
// returns true if the threshold achieved
func (bm *Bitmask) CaSBorder(lf uint, src bool, th uint) bool {
	lfBlock := lf / maskElemLen
	if lfBlock >= uint(len(bm.mask)) {
		return false
	}

	var cnt uint
	left, right := lf, lf+th
	for uint(cnt) != th && right/maskElemLen < uint(len(bm.mask)) {
		cnt += uint(bm.CaSLine(left, right, src))
		left, right = right, right+th-cnt
	}
	return uint(cnt) == th
}

// Next makes available iteration over the bitmap
func (bm *Bitmask) Next() (int, bool) {
	if bm.cur/maskElemLen >= uint(len(bm.mask)) {
		return -1, false
	}
	// no need to call other functions with defer here
	bm.cur++
	return int(bm.cur - 1), bm.Get(bm.cur - 1)
}

// Copy creates a copy of a bitmask which is only valid on [lf, rg) interval
// it's meant to be used for get requests from different goroutines
// This is the only method which has inner mutex yet
func (bm *Bitmask) Copy(lf, rg uint) *Bitmask {
	if lf >= rg {
		return nil
	}
	mask := make([]uint64, rg/maskElemLen+1)

	bm.cpMx.Lock()
	for i := lf / maskElemLen; i < rg/maskElemLen+1; i++ {
		if i >= uint(len(bm.mask)) {
			break
		}
		mask[i] = bm.mask[i]
	}
	bm.cpMx.Unlock()

	bmCp := Bitmask{mask: mask}
	return &bmCp
}

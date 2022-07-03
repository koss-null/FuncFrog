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

func (bm *Bitmask) setTrue(block uint, place uint) {
	bm.mask[block] |= 1 << place
}

func (bm *Bitmask) setFalse(block uint, place uint) {
	bm.mask[block] = (bm.mask[block] | (1 << (place % maskElemLen))) - (1 << (place % maskElemLen))
}

func (bm *Bitmask) getBit(num uint64, place uint) bool {
	return (num>>place)&1 == 1
}

func (bm *Bitmask) Put(place uint, val bool) {
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
		bm.setTrue(place/maskElemLen, place%maskElemLen)
		return
	}

	bm.setFalse(place/maskElemLen, place%maskElemLen)
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
	lfm, rgm := lf%maskElemLen, rg%maskElemLen

	for rgBlock >= uint(len(bm.mask)) {
		bm.mask = append(bm.mask, 0)
	}
	if val {
		if lfBlock == rgBlock {
			for place := lfm; place < rgm; place++ {
				bm.setTrue(lfBlock, place)
			}
			return
		}

		for place := lfm; place < maskElemLen; place++ {
			bm.setTrue(lfBlock, place)
		}
		for block := lfBlock + 1; block < rgBlock; block++ {
			bm.mask[block] = u64max
		}
		for place := u0; place < rgm; place++ {
			bm.setTrue(rgBlock, place)
		}
		return
	}

	if lfBlock == rgBlock {
		for place := lfm; place < rgm; place++ {
			bm.setFalse(lfBlock, place)
		}
		return
	}

	// can do it faster
	for place := lfm; place < maskElemLen; place++ {
		bm.setFalse(lfBlock, place)
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = 0
	}
	for place := u0; place < rgm; place++ {
		bm.setFalse(rgBlock, place)
	}
	return
}

func (bm *Bitmask) Get(place uint) bool {
	if place/maskElemLen > uint(len(bm.mask)) {
		return false
	}
	return bm.getBit(bm.mask[place/maskElemLen], place%maskElemLen)
}

// CaS compare and swap, returns true in case of success
// despite the title, the operation is not atomic
// since we are dealing with the bool, dst = !src
func (bm *Bitmask) CaS(place uint, src bool) bool {
	block := place / maskElemLen
	blkPlace := place % maskElemLen
	if block > uint(len(bm.mask)) {
		return false
	}

	if bm.getBit(bm.mask[block], blkPlace) == src {
		if src {
			bm.setFalse(block, blkPlace)
			return true
		}
		bm.setTrue(block, blkPlace)
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
	lfm, rgm := lf%maskElemLen, rg%maskElemLen
	if rgBlock >= uint(len(bm.mask)) {
		return 0
	}

	cnt := 0
	if src {
		if lfBlock == rgBlock {
			for place := lfm; place < rgm; place++ {
				fmt.Printf("bitmask %b\n", bm.mask[lfBlock])
				if bm.getBit(bm.mask[lfBlock], place) {
					bm.setFalse(lfBlock, place)
					cnt++
				}
				return cnt
			}
		}

		for place := lfm; place < maskElemLen; place++ {
			if bm.getBit(bm.mask[lfBlock], place) {
				bm.setFalse(lfBlock, place)
				cnt++
			}
		}
		for block := lfBlock + 1; block < rgBlock; block++ {
			bm.mask[block] = 0
			cnt += bits.OnesCount64(bm.mask[block])
		}
		for place := u0; place < rgm; place++ {
			if bm.getBit(bm.mask[rgBlock], place) {
				bm.setFalse(lfBlock, place)
				cnt++
			}
		}
		return cnt
	}

	if lfBlock == rgBlock {
		for place := lfm; place < rgm; place++ {
			if !bm.getBit(bm.mask[lfBlock], place) {
				bm.setTrue(lfBlock, place)
				cnt++
			}
		}
		return cnt
	}

	// can do it faster
	for place := lfm; place < maskElemLen; place++ {
		if !bm.getBit(bm.mask[lfBlock], place) {
			bm.setTrue(lfBlock, place)
			cnt++
		}
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		cnt += (maskElemLen - bits.OnesCount64(bm.mask[block]))
		bm.mask[block] = u64max
	}
	for place := u0; place < rgm; place++ {
		if !bm.getBit(bm.mask[rgBlock], place) {
			bm.setTrue(rgBlock, place)
			cnt++
		}

	}
	return cnt
}

// CaSBorder compares all values starting from lf with src and
// changes to !src if the val is equal. runs until th changes is done
// returns true if the threshold achieved
func (bm *Bitmask) CaSBorder(lf uint, src bool, th uint) bool {
	if th == 0 {
		return false
	}

	lfBlock := lf / maskElemLen
	if lfBlock > uint(len(bm.mask)) {
		lfBlock = uint(len(bm.mask))
	}

	var cnt uint
	left, right := lf, lf+th
	// FIXME: step depends on th
	for uint(cnt) != th && right/maskElemLen < uint(len(bm.mask)) {
		fmt.Println("cas line", left, right, src)
		cnt += uint(bm.CaSLine(left, right, src))
		left, right = right, right+th-cnt+1
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

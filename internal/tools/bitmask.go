package tools

import (
	"math/bits"
	"sync"
)

const (
	maskElemLen = 64
	u0          = uint(0)
	u64max      = ^uint64(0) // all ones
)

// FIXME: in some places we use uint indexing for non-blocks instances. It need to be changed into uint64
type Bitmask[T any] struct {
	mask []uint64
	cur  uint
	cpMx sync.Mutex
}

// setTrue makes bit from block on place equal to 1
// Warning: due to performanse reasons no additional safty checks done
func (bm *Bitmask[T]) setTrue(block uint, place uint) {
	bm.mask[block] |= 1 << place
}

func (bm *Bitmask[T]) setFalse(block uint, place uint) {
	bm.mask[block] = (bm.mask[block] | (1 << (place % maskElemLen))) - (1 << (place % maskElemLen))
}

func (bm *Bitmask[T]) getBit(num uint64, place uint) bool {
	return (num>>place)&1 == 1
}

func (bm *Bitmask[T]) Put(place uint, val bool) {
	// it's funny, but it seems we can do like this
	// (don't insert 0 if no place for it yet)
	if val && place/maskElemLen > uint(len(bm.mask)) {
		for i := len(bm.mask) - 1; uint(i) < place/maskElemLen; i++ {
			bm.mask = append(bm.mask, 0)
		}
		// it is called only once from here
		defer bm.Put(place, val)
		return
	}

	if val {
		bm.setTrue(place/maskElemLen, place%maskElemLen)
		return
	}
	bm.setFalse(place/maskElemLen, place%maskElemLen)
}

// PutLine does Put for the elements in range [lf, rg) with the val
func (bm *Bitmask[T]) PutLine(lf, rg uint, val bool) {
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
	fn, blockVal := bm.setFalse, uint64(0)
	if val {
		fn, blockVal = bm.setTrue, u64max
	}

	if lfBlock == rgBlock {
		for place := lfm; place < rgm; place++ {
			fn(lfBlock, place)
		}
		return
	}

	for place := lfm; place < maskElemLen; place++ {
		fn(lfBlock, place)
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = blockVal
	}
	for place := u0; place < rgm; place++ {
		fn(rgBlock, place)
	}
}

func (bm *Bitmask[T]) Get(place uint) bool {
	if place/maskElemLen > uint(len(bm.mask)) {
		return false
	}
	return bm.getBit(bm.mask[place/maskElemLen], place%maskElemLen)
}

// CaS compare and swap, returns true in case of success
// despite the title, the operation is not atomic
// since we are dealing with the bool, dst = !src
func (bm *Bitmask[T]) CaS(place uint, src bool) bool {
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
func (bm *Bitmask[T]) CaSLine(lf, rg uint, src bool) int {
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

	cnt, blockVal := 0, u64max
	fn := func(bl uint, pl uint) {
		if bm.getBit(bm.mask[bl], pl) {
			bm.setFalse(bl, pl)
			cnt++
		}
	}
	if src {
		blockVal = uint64(0)
		fn = func(bl uint, pl uint) {
			if !bm.getBit(bm.mask[bl], pl) {
				bm.setTrue(bl, pl)
				cnt++
			}
		}
	}

	if lfBlock == rgBlock {
		for place := lfm; place < rgm; place++ {
			fn(lfBlock, place)
		}
		return cnt
	}

	for place := lfm; place < maskElemLen; place++ {
		fn(lfBlock, place)
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = blockVal
		if src {
			cnt += bits.OnesCount64(bm.mask[block])
			continue
		}
		cnt += maskElemLen - bits.OnesCount64(bm.mask[block])
	}
	for place := u0; place < rgm; place++ {
		fn(rgBlock, place)
	}
	return cnt
}

// FIXME: why does some CaS functions returns int when other one returns bool
// int is more useful

// CaSBorder compares all values starting from lf with src and
// changes to !src if the val is equal. runs until th changes is done
// returns true if the threshold achieved
// FIXME: this one may work rly slow for now
func (bm *Bitmask[T]) CaSBorder(lf uint, src bool, th uint) bool {
	lfBlock := lf / maskElemLen
	lfm := lf % maskElemLen

	cnt, blockVal := u0, u64max
	fn := func(bl uint, pl uint) {
		if !bm.getBit(bm.mask[bl], pl) {
			bm.setTrue(bl, pl)
			cnt++
		}
	}
	if src {
		blockVal = uint64(0)
		fn = func(bl uint, pl uint) {
			if bm.getBit(bm.mask[bl], pl) {
				bm.setFalse(bl, pl)
				cnt++
			}
		}
	}

	for place := lfm; place < maskElemLen && cnt != th; place++ {
		fn(lfBlock, place)
	}

	block := lfBlock + 1
	for cnt != th && block < uint(len(bm.mask)) {
		curCnt := bits.OnesCount64(bm.mask[block])
		if !src {
			curCnt = maskElemLen - curCnt
		}
		if cnt+uint(curCnt) > th {
			break
		}
		bm.mask[block] = blockVal
		cnt += uint(curCnt)

		block++
	}

	for place := u0; cnt < th && block < uint(len(bm.mask)); place++ {
		fn(block, place)
	}
	return cnt == th
}

// CaSBorderBw is the same as CaSBorder but it goes from the end of the array
// FIXME: it takes one argument since it only works for src = true now
// FIXME: this one may work rly slow for now
func (bm *Bitmask[T]) CaSBorderBw(th uint) bool {
	cnt := u0
	block := len(bm.mask) - 1

	for cnt != th && block > -1 {
		curCnt := bits.OnesCount64(bm.mask[block])
		if cnt+uint(curCnt) > th {
			break
		}
		bm.mask[block] = uint64(0)
		cnt += uint(curCnt)

		block--
	}

	for place := maskElemLen - 1; cnt != th && place > -1; place-- {
		if bm.getBit(bm.mask[block], uint(place)) {
			bm.setFalse(uint(block), uint(place))
			cnt++
		}
	}
	return cnt == th
}

// Next makes available iteration over the bitmap
func (bm *Bitmask[T]) Next() (int, bool) {
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
func (bm *Bitmask[T]) Copy(lf, rg uint) *Bitmask[T] {
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

	return &Bitmask[T]{mask: mask}
}

// ShallowCopy creates the copy of a bitmask, but it does not copy the underlying array
// So changes done on this bm fragment will be done in the parent bm
// It should be used carefully since it does not work with methods that changes the underlying array
func (bm *Bitmask[T]) ShallowCopy(lf, rg uint) *Bitmask[T] {
	if lf >= rg {
		return nil
	}
	if rg > uint(len(bm.mask)) {
		rg = uint(len(bm.mask))
	}
	return &Bitmask[T]{mask: bm.mask[lf:rg]}
}

func (bm *Bitmask[T]) CountOnes() (cnt uint64) {
	for i := range bm.mask {
		cnt += uint64(bits.OnesCount64(bm.mask[i]))
	}
	return
}

// Len returns the value which is not less than the length of a bitmask
// it may be slightly larger, but all overbound values will be false
// it returns uint64 to avoid an overflow
func (bm *Bitmask[T]) Len() uint64 {
	return uint64(len(bm.mask)) * maskElemLen
}

// Apply applies the bitmask to an array
func (bm *Bitmask[T]) Apply(a []T) []T {
	elemCnt := 0
	for i := range bm.mask {
		elemCnt += bits.OnesCount64(bm.mask[i])
	}

	res := make([]T, elemCnt)
	resI := 0
	for i := range bm.mask {
		curMask, place := bm.mask[i], 0
		for curMask != 0 {
			if curMask&1 == 1 {
				res[resI] = a[i*maskElemLen+place]
				resI++
			}
			curMask <<= 1
			place++
		}
	}
	return res
}

package tools

import (
	"math/bits"
	"sync"

	"github.com/koss-null/lambda/pkg/fnmod"
)

const (
	blockLen = 64
	u0       = uint(0)
	u64max   = ^uint64(0) // all ones
)

type block struct {
	fn fnmod.SelfFn[uint64]
	mx *sync.Mutex
}

// FIXME: in some places we use uint indexing for non-blocks instances. It need to be changed into uint64
type Bitmask[T any] struct {
	mask []block
	cur  uint
	cpMx sync.Mutex
}

func wrapUint64(val uint64) block {
	mx := &sync.Mutex{}
	return block{
		fn: func() uint64 {
			mx.Lock()
			defer mx.Unlock()
			return val
		},
		mx: mx,
	}
}

func (b *block) true(place uint) {
	b.mx.Lock()

	val := b.fn() | (uint64(1) << uint64(place%blockLen))
	*b = wrapUint64(val)

	// this functinos are expected to be fast, so don't use defer here
	b.mx.Unlock()
}

func (b *block) false(place uint) {
	b.mx.Lock()

	val := (b.fn() | (1 << (place % blockLen))) - (1 << (place % blockLen))
	*b = wrapUint64(val)

	b.mx.Unlock()
}

func (b *block) bit(place uint) bool {
	return (b.fn()>>place)&1 == 1
}

func (b *block) copy() block {
	return wrapUint64(b.fn())
}

func (b *block) onesCnt() int {
	return bits.OnesCount64(b.fn())
}

// setTrue makes bit from block on place equal to 1
// Warning: due to performanse reasons no additional safty checks done
func (bm *Bitmask[T]) setTrue(block uint, place uint) {
	bm.mask[block].true(place)
}

func (bm *Bitmask[T]) setFalse(block uint, place uint) {
	bm.mask[block].false(place)
}

func (bm *Bitmask[T]) Put(place uint, val bool) {
	// it's funny, but it seems we can do like this
	// (don't insert 0 if no place for it yet)
	if val && place/blockLen > uint(len(bm.mask)) {
		for i := len(bm.mask) - 1; uint(i) < place/blockLen; i++ {
			bm.mask = append(bm.mask, wrapUint64(0))
		}
		// it is called only once from here
		bm.Put(place, val)
		return
	}

	if val {
		bm.mask[place/blockLen].true(place)
		return
	}
	bm.mask[place/blockLen].false(place)
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

	lfBlock := lf / blockLen
	rgBlock := rg / blockLen
	lfm, rgm := lf%blockLen, rg%blockLen

	for rgBlock >= uint(len(bm.mask)) {
		bm.mask = append(bm.mask, wrapUint64(0))
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

	for place := lfm; place < blockLen; place++ {
		fn(lfBlock, place)
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = wrapUint64(blockVal)
	}
	for place := u0; place < rgm; place++ {
		fn(rgBlock, place)
	}
}

func (bm *Bitmask[T]) Get(place uint) bool {
	if place/blockLen >= uint(len(bm.mask)) {
		return false
	}
	return bm.mask[place/blockLen].bit(place % blockLen)
}

// CaS compare and swap, returns true in case of success
// despite the title, the operation is not atomic
// since we are dealing with the bool, dst = !src
func (bm *Bitmask[T]) CaS(place uint, src bool) bool {
	block := place / blockLen
	blkPlace := place % blockLen
	if block > uint(len(bm.mask)) {
		return false
	}

	if bm.mask[block].bit(blkPlace) == src {
		if src {
			bm.mask[block].false(blkPlace)
			return true
		}
		bm.mask[block].true(blkPlace)
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

	lfBlock := lf / blockLen
	rgBlock := rg / blockLen
	lfm, rgm := lf%blockLen, rg%blockLen
	if rgBlock >= uint(len(bm.mask)) {
		return 0
	}

	cnt, blockVal := 0, u64max
	fn := func(bl uint, pl uint) {
		if bm.mask[bl].bit(pl) {
			bm.mask[bl].false(pl)
			cnt++
		}
	}
	if src {
		blockVal = uint64(0)
		fn = func(bl uint, pl uint) {
			if !bm.mask[bl].bit(pl) {
				bm.mask[bl].true(pl)
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

	for place := lfm; place < blockLen; place++ {
		fn(lfBlock, place)
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = wrapUint64(blockVal)
		if src {
			cnt += bm.mask[block].onesCnt()
			continue
		}
		cnt += int(blockLen) - bm.mask[block].onesCnt()
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
	lfBlock := lf / blockLen
	lfm := lf % blockLen

	cnt, blockVal := u0, u64max
	fn := func(bl uint, pl uint) {
		if !bm.mask[bl].bit(pl) {
			bm.mask[bl].true(pl)
			cnt++
		}
	}
	if src {
		blockVal = uint64(0)
		fn = func(bl uint, pl uint) {
			if bm.mask[bl].bit(pl) {
				bm.mask[bl].false(pl)
				cnt++
			}
		}
	}

	for place := lfm; place < blockLen && cnt != th; place++ {
		fn(lfBlock, place)
	}

	block := lfBlock + 1
	for cnt != th && block < uint(len(bm.mask)) {
		curCnt := bm.mask[block].onesCnt()
		if !src {
			curCnt = int(blockLen) - curCnt
		}
		if cnt+uint(curCnt) > th {
			break
		}
		bm.mask[block] = wrapUint64(blockVal)
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
		curCnt := bm.mask[block].onesCnt()
		if cnt+uint(curCnt) > th {
			break
		}
		bm.mask[block] = wrapUint64(0)
		cnt += uint(curCnt)

		block--
	}

	for place := int(blockLen) - 1; cnt != th && place > -1; place-- {
		if bm.mask[block].bit(uint(place)) {
			bm.mask[block].false(uint(place))
			cnt++
		}
	}
	return cnt == th
}

// Next makes available iteration over the bitmap
func (bm *Bitmask[T]) Next() (int, bool) {
	if bm.cur/blockLen >= uint(len(bm.mask)) {
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
	mask := make([]block, (rg-lf)/blockLen+1)

	for i := lf / blockLen; i < rg/blockLen+1; i++ {
		if i >= uint(len(bm.mask)) {
			break
		}
		mask[i] = bm.mask[i].copy()
	}

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
		cnt += uint64(bm.mask[i].onesCnt())
	}
	return
}

// Len returns the value which is not less than the length of a bitmask
// it may be slightly larger, but all overbound values will be false
// it returns uint64 to avoid an overflow
func (bm *Bitmask[T]) Len() uint64 {
	return uint64(len(bm.mask)) * blockLen
}

// Apply applies the bitmask to an array
func (bm *Bitmask[T]) Apply(a []T) []T {
	elemCnt := 0
	for i := range bm.mask {
		elemCnt += bm.mask[i].onesCnt()
	}

	res := make([]T, elemCnt)
	resI := 0
	for i := range bm.mask {
		curMask, place := bm.mask[i].fn(), 0
		for curMask != 0 {
			if curMask&1 == 1 {
				res[resI] = a[i*int(blockLen)+place]
				resI++
			}
			curMask <<= 1
			place++
		}
	}
	return res
}

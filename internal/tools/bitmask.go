package tools

const maskElemLen = 64

type Bitmask struct {
	mask []uint64
	cur  uint
}

func (bm Bitmask) Put(place uint, val bool) {
	if place/maskElemLen > uint(len(bm.mask)) {
		for i := 0; i < len(bm.mask); i++ {
			bm.mask = append(bm.mask, 0)
		}
		bm.Put(place, val)
		return
	}

	if val {
		bm.mask[place/maskElemLen] |= 1 << place
		return
	}

	// here we know that on place there is 1 and val is false, so:
	bm.mask[place/maskElemLen] -= 1 << place
}

func (bm Bitmask) Get(place uint) bool {
	if place/maskElemLen > uint(len(bm.mask)) {
		return false
	}
	return (bm.mask[place/maskElemLen]>>(place%maskElemLen))&1 == 1
}

// Next makes available iteration over the bitmap
func (bm Bitmask) Next() (int, bool) {
	if bm.cur >= uint(len(bm.mask)) {
		return -1, false
	}
	// no need to call other functions with defer here
	bm.cur++
	return int(bm.cur - 1), bm.Get(bm.cur - 1)
}

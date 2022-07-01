package tools

const (
	maskElemLen = 64
	u0          = uint(0)
	u64max      = ^uint64(0) // all ones
)

type Bitmask struct {
	mask []uint64
	cur  uint
	len  uint
}

func (bm *Bitmask) Put(place uint, val bool) {
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

	// here we know that on place position there is 1 and val is false, so:
	bm.mask[place/maskElemLen] -= 1 << place
	if place > bm.len {
		place = bm.len
	}
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
				bm.mask[lfBlock] |= 1 << place
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
			bm.mask[lfBlock] -= 1 << place
		}
		return
	}

	// can do it faster
	for place := lf; place < maskElemLen; place++ {
		bm.mask[lfBlock] -= 1 << place
	}
	for block := lfBlock + 1; block < rgBlock; block++ {
		bm.mask[block] = 0
	}
	for place := u0; place < rg%maskElemLen; place++ {
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

// Next makes available iteration over the bitmap
func (bm *Bitmask) Next() (int, bool) {
	if bm.cur/maskElemLen >= uint(len(bm.mask)) {
		return -1, false
	}
	// no need to call other functions with defer here
	bm.cur++
	return int(bm.cur - 1), bm.Get(bm.cur - 1)
}

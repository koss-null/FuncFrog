package bitmap

import "sync"

type naiveBM struct {
	mx *sync.Mutex
	bm []bool
}

func NewNaive(i int) Bitmap {
	return &naiveBM{
		bm: make([]bool, i),
		mx: &sync.Mutex{},
	}
}

func (b *naiveBM) Get(i int) bool {
	b.mx.Lock()
	defer b.mx.Unlock()

	if i >= len(b.bm) || i < 0 {
		return false
	}
	return b.bm[i]
}

func (b *naiveBM) Set(lf int, rg int, val bool) {
	if lf < 0 || rg < 0 || lf > rg {
		return
	}

	b.mx.Lock()
	defer b.mx.Unlock()

	if rg >= len(b.bm) {
		larger := make([]bool, rg)
		cp := b.bm
		if lf < len(b.bm) {
			cp = b.bm[:lf]
		}
		for i := range cp {
			larger[i] = cp[i]
		}
		b.bm = larger
	}
	for i := lf; i < rg; i++ {
		b.bm[i] = val
	}
}

func (b *naiveBM) SetTrue(i int) {
	b.Set(i, i+1, true)
}

func (b *naiveBM) SetFalse(i int) {
	b.Set(i, i+1, false)
}

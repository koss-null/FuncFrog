package tools

import (
	"math"
	"sync"
)

// SqrtDecMx is an application of sqrt decomposition for mutexes on array.
// It helps to block only some small part of an array instead of blocking all
// the array.
// All locks are taken one by one under the common mutex.
// If some areas are overlapping, they are stacked (block on Lock()).
type SqrtDecMx struct {
	setMxLock sync.Mutex
	mxs       []sync.Mutex
}

type unlock func()

func NewSqrtDecMx(n uint) *SqrtDecMx {
	sdm := &SqrtDecMx{}
	sdm.rebuild(n)
	return sdm
}

func (sdm *SqrtDecMx) rebuild(n uint) {
	sdm.setMxLock.Lock()
	defer sdm.setMxLock.Unlock()

	if sdm.mxs != nil {
		for i := range sdm.mxs {
			sdm.mxs[i].Lock()
		}
		// TODO: check out do I need to do it [for nice GC I guess]
		mxs := sdm.mxs
		defer func() {
			for i := range mxs {
				mxs[i].Unlock()
			}
		}()
	}

	sdm.mxs = make([]sync.Mutex, uint(math.Sqrt(float64(n)))+1)
}

// Lock blocks until it'll take all the mutexes to use [lf, rg) underlying array
// returns Unlock function wich is need to be called to unlock
func (sdm *SqrtDecMx) Lock(lf, rg uint) unlock {
	unlocks := []unlock{}
	sdm.setMxLock.Lock()
	defer sdm.setMxLock.Unlock()

	for i := int(math.Sqrt(float64(lf))); i < int(math.Sqrt(float64(rg))); i++ {
		sdm.mxs[i].Lock()
		unlocks = append(unlocks, sdm.mxs[i].Unlock)
	}
	return func() {
		for i := range unlocks {
			unlocks[i]()
		}
	}
}

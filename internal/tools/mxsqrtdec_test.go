package tools

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_SqrtDecMx(t *testing.T) {
	rand.Seed(time.Now().Unix())
	wg := sync.WaitGroup{}

	st := time.Now()
	sdm := NewSqrtDecMx(100500)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			lf, rg := rand.Uint32()%100500, rand.Uint32()%100500
			if lf > rg {
				rg, lf = lf, rg
			}
			ul := sdm.Lock(uint(lf), uint(rg))
			time.Sleep(69 * time.Microsecond)
			ul()
			wg.Done()
		}()
	}
	wg.Wait()
	require.True(t, time.Since(st).Milliseconds() < (10*time.Minute).Milliseconds())
}

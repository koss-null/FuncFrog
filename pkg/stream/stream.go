package stream

import (
	"sort"
	"sync"

	"github.com/koss-null/lambda-go/internal/tools"
)

const (
	u0 = uint(0)
)

type task[T any] struct {
	op Operation
	dt []T
	bm *tools.Bitmask
	wg *sync.WaitGroup
}

type (
	stream[T any] struct {
		onceRun sync.Once

		tasks    chan task
		wrksCnt uint
		stopWrkrs chan struct{}
		
		// syncMx is used for Op.sync == true
		syncMx sync.Mutex
		done map[task[T]]struct{}
		doneMx sync.Mutex

		ops     []Operation
		fns     []func([]T, int)

		dt      []T
		dtMu    sync.Mutex

		bm      *tools.Bitmask[T]
		mbMx    sync.Mutex
	}

	// StreamI describes all functions available for the stream API
	StreamI[T any] interface {
		// Trimming the sequence
		// Skip removes first n elements from underlying slice
		Skip(n uint) StreamI[T]
		// Trim removes last n elements from underlying slice
		Trim(n uint) StreamI[T]

		// Actions on sequence
		// Map executes function on each element of an underlying slice and
		// changes the element to the result of the function
		Map(func(T) T) StreamI[T]
		// Reduce takes the result of the prev operation and applies it to the next item
		Reduce(func(T, T) T) StreamI[T]
		// Filteer
		Filter(func(T) bool) StreamI[T]
		// Sorted sotrs the underlying array
		Sorted(less func(a, b T) bool) StreamI[T]
		// Split splits initial slice into multiple
		Split(func(a T) []T) StreamI[T]

		// Config functions
		// Go splits execution into n goroutines, if multiple Go functions are along
		// the stream pipeline, it applies as soon as it's met in the sequence
		Go(n uint) StreamI[T]

		// Final functions (which does not return the stream)
		// Slice returns resulting slice
		Slice() []T
		// Any returns true if the underlying slice is not empty and at least one element is true
		Any() bool
		// None is not Any()
		None() bool
		// Count returns the length of the result array
		Count() int
		// Sum sums up the items in the array
		Sum(func(int64, T) int64) int64
		// Contains returns true if element is in the array
		Contains(item T, eq func(T, T) bool) bool
	}
)

// Stream creates new instance of stream
func Stream[T any](data []T) StreamI[T] {
	dtCopy := make([]T, len(data))
	copy(dtCopy, data)
	workers := make(chan struct{}, 1)
	workers <- struct{}{}
	bm := tools.Bitmask[T]{}
	bm.PutLine(0, uint(len(data)), true)
	return &stream[T]{wrks: workers, wrksCnt: 1, dt: dtCopy, bm: &bm}
}

// S is a shortened Stream
func S[T any](data []T) StreamI[T] {
	return Stream(data)
}

func (st *stream[T]) waitSync() {
	for i := u0; i < st.wrksCnt; i++ {
		<-st.wrks
	}
}

func (st *stream[T]) returnWrks() {
	for i := u0; i < st.wrksCnt; i++ {
		st.wrks <- struct{}{}
	}
}

func (st *stream[T]) Skip(n uint) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}

		<-st.wrks

		st.mbMx.Lock()
		_ = st.bm.CaSBorder(0, true, n)
		st.mbMx.Unlock()

		st.wrks <- struct{}{}
	})
	return st
}

func (st *stream[T]) Trim(n uint) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}

		<-st.wrks

		st.mbMx.Lock()
		_ = st.bm.CaSBorderBw(n)
		st.mbMx.Unlock()

		st.wrks <- struct{}{}
	})
	return st
}

func (st *stream[T]) Map(fn func(T) T) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, _ int) {
		<-st.wrks

		for i := range dt {
			dt[i] = fn(dt[i])
		}

		st.wrks <- struct{}{}
	})
	return st
}

func (st *stream[T]) Reduce(fn func(T, T) T) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset == 0 {
			<-st.wrks
			res := dt[0]
			for i := range dt[1:] {
				res = fn(res, dt[i])
			}
			st.wrks <- struct
			retrun
		}
		<-st.wrks

		st.wrks <- struct
	})
	return st
}

func (st *stream[T]) Filter(fn func(T) bool) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		<-st.wrks

		bm := st.bm.Copy(uint(offset), uint(offset+len(dt)))
		res := make([]bool, len(dt))
		for i := range dt {
			if bm.Get(uint(offset + i)) {
				res[i] = fn(dt[i])
			}
		}

		st.mbMx.Lock()
		for i := range res {
			// Filter does not add already removed items
			if !res[i] && bm.Get(uint(offset+i)) {
				st.bm.Put(uint(offset+i), false)
			}
		}
		st.mbMx.Unlock()

		st.wrks <- struct{}{}
	})
	return st
}

// Sorted adds two functions: first one sorts everything
// FIXME: Sorted does not apply the bitmask
func (st *stream[T]) Sorted(less func(a, b T) bool) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		<-st.wrks

		sort.Slice(dt, func(i, j int) bool { return less(dt[j], dt[j]) })

		st.wrks <- struct{}{}
	})
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}

		<-st.wrks

		sort.Slice(st.dt, func(i, j int) bool {
			return less(dt[j], dt[j])
		})

		st.wrks <- struct{}{}
	})

	return st
}

func (st *stream[T]) Go(n uint) StreamI[T] {
	st.fns = append(st.fns, func(dt []T, offset int) {
		for i := u0; i < st.wrksCnt; i++ {
			<-st.wrks
		}
		close(st.wrks)

		wrks := make(chan struct{}, n)
		for i := u0; i < n; i++ {
			wrks <- struct{}{}
		}
		st.wrksCnt = n
	})
	return st
}

func (st *stream[T]) Split(sp func(T) []T) StreamI[T] {
	// TODO: implement
	return st
}

func (st *stream[T]) nextOp() Operation {
	return st.ops[0]
}

// FIXME: don't need this
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (st *stream[T]) addTasks() {
	st.dt = st.bm.Apply(st.dt)
	st.bm.PutLine(0, len(st.dt), true)

	dataLen := st.bm.Len()
	blockSize := min(minSplitLen, dataLen)
	if dataLen / st.wrksCnt > blockSize {
		blockSize = dataLen / st.wrksCnt
	}

	lf, rg := 0, blockSize
	for lf < dataLen {
		st.tasks <- task{st.nextOp(), st.dt[lf:rg], bm.ShallowCopy(lf, rg)}
		lf = rg
		rg += blockSize
	}

	return 
}

// run start executing st.fns
func (st *stream[T]) run() {
	st.onceRun.Do(func() {
		st.addTasks()
		st.startWorkers()
	})
}

func (st *stream[T]) startWorkers() {
	done := make(chan struct{})
	st.stopWrkrs = done
	for i := 0; i < st.wrksCnt; i++ {
		go func() {
			for {
				select {
				case <-done: 
						return
				case task := <-st.tasks:
					if task.op.Sync() {
						st.syncMx.Lock()
						task.wg.Wait()

						st.doneMx.Lock()
						if _, ok := st.done[task]; !ok {
							continue
						}
						st.done[task] = struct{}{}
						st.doneMx.Unlock()

						task.op.Do(task.dt, task.bm)
						st.addTasks()

						st.syncMx.Unlock()
						continue
					}

					task.op.Do(task.dt, task.bm)
					task.wg.Done()
					no := nextOp()
					if !no.sync {
						task.wg.Add(1)
					}
					st.tasks <- task{no, task.dt, task.bm, wg}
				}
			}
		}
	}
}

func (st *stream[T]) Slice() []T {
	var data []T
	st.fns = append(st.fns, func(dt []T, offset int) {
		if offset != 0 {
			return
		}
		for i := u0; i < st.wrksCnt; i++ {
			<-st.wrks
		}

		st.dtMu.Lock()
		st.mbMx.Lock()
		defer st.dtMu.Unlock()
		defer st.mbMx.Unlock()

		data = make([]T, 0, len(st.dt))
		for _, dt := range st.dt {
			_, exist := st.bm.Next()
			if exist {
				data = append(data, dt)
			}
		}
	})

	st.run()
	return data
}

func (st *stream[T]) Any() bool {
	return !st.None()
}

func (st *stream[T]) None() bool {
	return st.Count() == 0
}

func (st *stream[T]) Count() int {
	st.run()

	st.mbMx.Lock()
	defer st.mbMx.Unlock()
	return int(st.bm.CountOnes())
}

func (st *stream[T]) Sum(sum func(int64, T) int64) int64 {
	if sum == nil {
		return 0
	}

	var s int64
	st.fns = append(st.fns, func(a []T, offset int) {
		if offset != 0 {
			return
		}
		st.waitSync()
		for _, dt := range st.bm.Apply(st.dt) {
			s = sum(s, dt)
		}
	})

	st.run()
	return s
}

// Contains returns if the underlying array contains an item
func (st *stream[T]) Contains(item T, eq func(a, b T) bool) bool {
	if st.eq == nil {
		return false
	}

	st.waitSync()
	for i := range st.dt {
		if st.eq(st.dt[i], item) {
			return true
		}
	}
	return false
}

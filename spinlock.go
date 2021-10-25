package ppool

import (
	"runtime"
	"sync/atomic"
)

type spinLock uint32

const maxBackOff = 64

func (s *spinLock) Lock() {
	backOff := 1
	// 指数退避
	for !atomic.CompareAndSwapUint32((*uint32)(s), 0, 1) {
		for i := 0; i < backOff; i++ {
			runtime.Gosched()
		}
		if backOff < maxBackOff {
			backOff <<= 1
		}
	}
}

func (s *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(s), 0)
}

func NewSpinLock() *spinLock {
	return new(spinLock)
}

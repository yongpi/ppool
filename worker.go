package ppool

import (
	"runtime"
	"time"
)

var workerChanSize int

func init() {
	if runtime.GOMAXPROCS(0) > 1 {
		workerChanSize = 1
	}
}

type worker struct {
	ch         chan Task
	expireTime time.Time
	pool       *PPool
}

func NewWorker(pool *PPool) *worker {
	return &worker{
		ch:   make(chan Task, workerChanSize),
		pool: pool,
	}
}

func (w *worker) Run() {
	w.pool.IncrRunning()

	go func() {
		defer func() {
			w.pool.DecrRunning()
			w.pool.workerCache.Put(w)

			if err := recover(); err != nil {
				if ph := w.pool.config.PanicHandler; ph != nil {
					ph(err)
				} else {
					w.pool.config.Logger.Errorf("[PPool:worker]: worker exist from panic = %v", err)
					var buf [4096]byte
					n := runtime.Stack(buf[:], false)
					w.pool.config.Logger.Errorf("[PPool:worker]: worker exist from panic = %v, stack = %s", err, string(buf[:n]))

				}
			}

			w.pool.cond.Signal()
		}()

		for task := range w.ch {
			if task == nil {
				return
			}
			task()
			if flag := w.pool.addWorker(w); !flag {
				return
			}
		}
	}()
}

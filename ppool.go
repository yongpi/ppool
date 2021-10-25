package ppool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yongpi/putil/plog"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type defaultLogger struct {
}

func (d *defaultLogger) Debugf(format string, args ...interface{}) {
	plog.Debugf(format, args...)
}

func (d *defaultLogger) Errorf(format string, args ...interface{}) {
	plog.Errorf(format, args...)
}

type Options struct {
	MaxBlockingNum      int32
	WorkerCleanDuration time.Duration
	PanicHandler        func(interface{})
	Logger              Logger
}

type PPool struct {
	core        int32
	max         int32
	state       int32
	running     int32
	lock        sync.Locker
	cond        *sync.Cond
	blocking    int32
	workerCache *sync.Pool
	workers     Workers
	config      *Options
}

func NewPPool(core, max uint32, options ...Option) *PPool {
	lock := NewSpinLock()
	if core > max {
		core = max
	}
	pool := &PPool{
		core:   int32(core),
		max:    int32(max),
		lock:   lock,
		cond:   sync.NewCond(lock),
		config: new(Options),
	}

	for _, option := range options {
		option(pool.config)
	}

	pool.workerCache = &sync.Pool{
		New: func() interface{} {
			return NewWorker(pool)
		},
	}

	pool.workers = NewWorkerStack()
	if pool.config.WorkerCleanDuration == 0 {
		pool.config.WorkerCleanDuration = defaultCleanTime
	}
	if pool.config.Logger == nil {
		pool.config.Logger = new(defaultLogger)
	}

	// 定时清理 workers
	go pool.cycleCleanUp()

	return pool

}

func (p *PPool) cycleCleanUp() {
	tick := time.NewTicker(p.config.WorkerCleanDuration)
	defer tick.Stop()

	for range tick.C {
		if p.IsClosed() {
			break
		}

		p.lock.Lock()
		list := p.workers.Expiry(time.Now())
		p.lock.Unlock()
		for _, worker := range list {
			worker.ch <- nil
			worker = nil
		}

		if p.Running() == 0 {
			p.cond.Broadcast()
		}
	}
}

func (p *PPool) State() State {
	return State(atomic.LoadInt32(&p.state))
}

func (p *PPool) Submit(task Task) error {
	if p.IsClosed() {
		return errors.New("pool is closed")
	}
	w := p.allocateWorker()
	if w == nil {
		return errors.New("pool is overload")
	}
	w.Run()
	w.ch <- task
	return nil
}

func (p *PPool) allocateWorker() *worker {
	if p.Cap() > p.Running() {
		p.lock.Lock()
		w := p.workers.Get()
		p.lock.Unlock()
		if w == nil {
			w = p.workerCache.Get().(*worker)
		}
		return w
	}

	blockNum := p.config.MaxBlockingNum
	for p.Blocking() < blockNum {
		p.lock.Lock()
		p.blocking++
		p.cond.Wait()
		p.blocking--

		var rn int32
		if rn = p.Running(); rn == 0 {
			p.lock.Unlock()

			if p.IsClosed() {
				return nil
			}
			return p.workerCache.Get().(*worker)
		}

		w := p.workers.Get()
		p.lock.Unlock()
		if w != nil {
			return w
		}
		if p.Cap() > rn {
			return p.workerCache.Get().(*worker)
		}
	}

	return nil
}

func (p *PPool) Close() {
	atomic.StoreInt32(&p.state, int32(Closed))
	p.lock.Lock()
	p.workers.Reset()
	p.lock.Unlock()

	p.cond.Broadcast()
}

func (p *PPool) Running() int32 {
	return atomic.LoadInt32(&p.running)
}

func (p *PPool) Blocking() int32 {
	return atomic.LoadInt32(&p.blocking)
}

func (p *PPool) IncrRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *PPool) DecrRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *PPool) IsClosed() bool {
	return p.State() == Closed
}

func (p *PPool) Cap() int32 {
	return atomic.LoadInt32(&p.max)
}

func (p *PPool) Core() int32 {
	return atomic.LoadInt32(&p.core)
}

func (p *PPool) addWorker(w *worker) bool {
	if p.IsClosed() || p.Running() > p.Core() {
		return false
	}

	w.expireTime = time.Now().Add(p.config.WorkerCleanDuration)
	p.lock.Lock()

	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}

	err := p.workers.Insert(w)
	if err != nil {
		p.lock.Unlock()
		return false
	}

	p.lock.Unlock()
	p.cond.Signal()
	return true
}

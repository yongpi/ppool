package ppool

import "time"

type Workers interface {
	Insert(worker *worker) error
	Get() *worker
	Expiry(t time.Time) []*worker
	Reset()
}

func NewWorkerStack() *workerStack {
	return new(workerStack)
}

type workerStack struct {
	items   []*worker
	expires []*worker
}

func (w *workerStack) Reset() {
	for _, worker := range w.items {
		worker.ch <- nil
		worker = nil
	}
	w.items = nil
	w.expires = nil
}

func (w *workerStack) Insert(worker *worker) error {
	w.items = append(w.items, worker)
	return nil
}

func (w *workerStack) Get() *worker {
	il := len(w.items)
	if il == 0 {
		return nil
	}

	worker := w.items[il-1]
	w.items = w.items[:il-1]
	return worker
}

func (w *workerStack) Expiry(t time.Time) []*worker {
	w.expires = nil

	il := len(w.items)
	if il == 0 {
		return nil
	}

	i := w.BinarySearch(0, il-1, t)
	w.expires = append(w.expires, w.items[:i]...)
	copy(w.items, w.items[i:])
	for n := i; n < il; n++ {
		w.items[n] = nil
	}
	w.items = w.items[:i]

	return w.expires
}

func (w *workerStack) BinarySearch(l, r int, t time.Time) int {
	li := w.items[l]
	ri := w.items[r]
	if li.expireTime.After(t) {
		return l
	}
	if ri.expireTime.Before(t) {
		return r
	}

	var mid int
	for r > l {
		mid = (r-l)/2 + l
		mi := w.items[mid]
		if mi.expireTime.Before(t) {
			l = mid + 1
		}
		if mi.expireTime.After(t) {
			r = mid - 1
		}
	}

	return l
}

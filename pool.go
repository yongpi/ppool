package ppool

import "time"

type Task func()

type Pool interface {
	Submit(task Task) error
	Close()
	Running() int32
	Blocking() int32
	State() State
	Cap() int32
	Core() int32
}

type State int32

const (
	Init State = iota
	Closed
)

var (
	defaultCleanTime = time.Second
)

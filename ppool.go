package ppool

import "time"

type Task func()
type Spaceship func() interface{}

type Pool interface {
	Submit(task Task) error
	Close()
	Running() int32
	Blocking() int32
	State() State
	Cap() int32
	Core() int32
	Travel(ship Spaceship) (Feature, error)
	TravelSafe(ship Spaceship) Feature
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type State int32

const (
	Init State = iota
	Closed
)

var (
	defaultCleanTime = time.Second
)

type Feature interface {
	Get() interface{}
}

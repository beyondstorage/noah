package progress

import (
	"errors"
	"sync"
	"time"
)

// ErrTaskNotExist returns when try to get a non-exist task in center
var ErrTaskNotExist = errors.New("task not exist")

// Namer is the interface for Name() call
type Namer interface {
	Name() string
}

const finishedTotal int64 = -1

var center sync.Map         // center is the private sync map for progress data
var dataChan chan *sync.Map // dataChan is the chan for center trans
var sigChan chan struct{}   // sigChan is the signal chan to end the center trans

var once sync.Once

type taskID = string

func init() {
	dataChan = make(chan *sync.Map)
	sigChan = make(chan struct{})
}

// State is for the progress of a task
type State struct {
	TaskName string
	Status   string
	Done     int64
	Total    int64
}

// Finished specify whether a state is finished
func (s State) Finished() bool {
	return s.Total == finishedTotal
}

// InitState init a progress state
func InitState(name string) State {
	return NewState(name, "initing", 0, 0)
}

// FinishedState new a finished state
func FinishedState(name string) State {
	return NewState(name, "finished", 0, finishedTotal)
}

// NewState is the conductor of state
func NewState(name, status string, done, total int64) State {
	return State{
		TaskName: name,
		Status:   status,
		Done:     done,
		Total:    total,
	}
}

// SetState set the task's state with specific ID
func SetState(id taskID, s State) {
	center.Store(id, s)
}

// GetState get a task's state with specific ID
// if the task not exists, return ErrTaskNotExist err
func GetState(id taskID) (State, error) {
	v, ok := center.Load(id)
	if !ok {
		return State{}, ErrTaskNotExist
	}
	return v.(State), nil
}

func Start(d time.Duration) <-chan *sync.Map {
	go func() {
		tc := time.NewTicker(d)
		for {
			select {
			case <-tc.C:
				dataChan <- &center
			case <-sigChan:
				return
			}
		}
	}()
	return dataChan
}

// End close the sigChan to
// 1. close the sigChan to stop the goroutine in Start()
// 2. close the dataChan for client
// use once to ensure only close once
func End() {
	once.Do(func() {
		close(sigChan)
		close(dataChan)
	})
}

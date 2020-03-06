package progress

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrTaskNotExist returns when try to get a non-exist task in center
var ErrTaskNotExist = errors.New("task not exist")

// Namer is the interface for Name() call
type Namer interface {
	Name() string
}

const finishedStatus string = "finished"

type progressCenter struct {
	sync.Mutex
	data map[taskID]State
}

var center *progressCenter // center is the private locked map for progress data
// var center sync.Map // center is the private sync map for progress data
var dataChan chan map[taskID]State // dataChan is the chan for center trans
var sigChan chan struct{}          // sigChan is the signal chan to end the center trans

var once sync.Once

type taskID = string

func init() {
	center = &progressCenter{data: make(map[taskID]State)}
	dataChan = make(chan map[taskID]State)
	sigChan = make(chan struct{})
}

// State is for the progress of a task
type State struct {
	TaskName string
	Status   string
	Done     int64
	Total    int64
}

func (s State) String() string {
	return fmt.Sprintf("name: %s, status: %s, %d/%d", s.TaskName, s.Status, s.Done, s.Total)
}

// Finished specify whether a state is finished
func (s State) Finished() bool {
	return s.Status == finishedStatus
}

// InitState init a progress state
func InitState(name string) State {
	return NewState(name, "initing", 0, 1)
}

// FinishedState new a finished state
func FinishedState(name string, total int64) State {
	return NewState(name, finishedStatus, total, total)
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
	center.Lock()
	defer center.Unlock()
	center.data[id] = s
}

// GetState get a task's state with specific ID
// if the task not exists, return ErrTaskNotExist err
func GetState(id taskID) (State, error) {
	center.Lock()
	defer center.Unlock()
	v, ok := center.data[id]
	if !ok {
		return State{}, ErrTaskNotExist
	}
	return v, nil
}

// UpdateState just update the done field with given id
func UpdateState(id taskID, done int64) {
	center.Lock()
	defer center.Unlock()
	state, ok := center.data[id]
	if !ok {
		return
	}
	state.Done = done
	center.data[id] = state
}

// Start create a channel to get center for client
func Start(d time.Duration) <-chan map[taskID]State {
	go func() {
		tc := time.NewTicker(d)
		for {
			select {
			case <-tc.C:
				dataChan <- center.GetData()
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

// GetData copy the data from pc
func (pc *progressCenter) GetData() map[taskID]State {
	pc.Lock()
	defer pc.Unlock()
	res := make(map[taskID]State, len(pc.data))
	for k, v := range pc.data {
		res[k] = v
	}
	return res
}

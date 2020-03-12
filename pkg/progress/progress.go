package progress

import (
	"errors"
	"fmt"
	"sync"
)

// ErrTaskNotExist returns when try to get a non-exist task in center
var ErrTaskNotExist = errors.New("task not exist")

const finishedStatus string = "finished"

// ListingStatus used to specify a state is listing
const ListingStatus string = "listing"

type progressCenter struct {
	sync.Mutex
	data map[taskID]State
}

var center *progressCenter // center is the private locked map for progress data

type taskID = string

func init() {
	center = &progressCenter{data: make(map[taskID]State)}
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
	return s.Total > 0 && s.Done >= s.Total
}

// IsListing specify whether a state is listing
func (s State) IsListing() bool {
	return s.Status == ListingStatus
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

// GetStateByID get a task's state with specific ID
// if the task not exists, return ErrTaskNotExist err
func GetStateByID(id taskID) (State, error) {
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

// GetStates copy and return the state data from center
func GetStates() map[taskID]State {
	center.Lock()
	defer center.Unlock()
	res := make(map[taskID]State, len(center.data))
	for k, v := range center.data {
		res[k] = v
	}
	return res
}

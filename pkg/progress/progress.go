package progress

import (
	"errors"
	"fmt"
	"sync"
)

// typeList is progress which will last for a long time, like list
// typeInc is progress which will increase by time, like copy
const (
	typeList = iota
	typeInc
)

// ErrTaskNotExist returns when try to get a non-exist task in center
var ErrTaskNotExist = errors.New("task not exist")

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
	Name   string
	Status string
	Type   int
	Done   int64
	Total  int64
}

// modStateFn is the func to modify state's field
type modStateFn func(s *State)

// newState is the conductor of state
func newState(ms ...modStateFn) State {
	s := State{}
	for _, fn := range ms {
		fn(&s)
	}
	return s
}

// withType returns modify func with type
func withType(typ int) modStateFn {
	return func(s *State) {
		s.Type = typ
	}
}

// withStatus returns modify func with status
func withStatus(status string) modStateFn {
	return func(s *State) {
		s.Status = status
	}
}

// withName returns modify func with name
func withName(name string) modStateFn {
	return func(s *State) {
		s.Name = name
	}
}

// withProgress returns modify func with done and total
func withProgress(done, total int64) modStateFn {
	return func(s *State) {
		s.Done = done
		s.Total = total
	}
}

// String implements Stringer
func (s State) String() string {
	return fmt.Sprintf("name: %s, status: %s, type: %d, %d/%d", s.Name, s.Status, s.Type, s.Done, s.Total)
}

// Finished specify whether a state is finished
func (s State) Finished() bool {
	return s.Total > 0 && s.Done >= s.Total
}

// IsListType specify whether a state is list type
func (s State) IsListType() bool {
	return s.Type == typeList
}

// InitListState init a list progress' state with default progress 0 / 1
func InitListState(name, status string) State {
	return newState(
		withName(name),
		withStatus(status),
		withType(typeList),
		withProgress(0, 1),
	)
}

// InitIncState init an inc progress' state with default progress 0/total
func InitIncState(name, status string, total int64) State {
	return newState(
		withName(name),
		withStatus(status),
		withType(typeInc),
		withProgress(0, total),
	)
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

// FinishState set the state with given ID to final
func FinishState(id taskID) {
	center.Lock()
	defer center.Unlock()
	state, ok := center.data[id]
	if !ok {
		return
	}
	state.Done = state.Total
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

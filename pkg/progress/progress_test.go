package progress

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFinishState(t *testing.T) {
	id, total := uuid.New().String(), time.Now().Unix()
	SetState(id, InitIncState("", "", total))
	tests := []struct {
		name  string
		id    taskID
		total int64
	}{
		{
			"normal",
			id,
			total,
		},
		{
			"not exist",
			"",
			total,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			FinishState(tt.id)
			s, err := GetStateByID(tt.id)
			if err != nil {
				assert.Equal(t, ErrTaskNotExist, err, tt.name)
				return
			}
			assert.Nil(t, err, tt.name)
			assert.Equal(t, tt.total, s.Total)
		})
	}
}

func TestGetStates(t *testing.T) {
	SetState(uuid.New().String(), InitIncState(uuid.New().String(), uuid.New().String(), time.Now().Unix()))
	SetState(uuid.New().String(), InitListState(uuid.New().String(), uuid.New().String()))

	t.Run("get_states", func(t *testing.T) {
		if got := GetStates(); !reflect.DeepEqual(got, center.data) {
			t.Errorf("GetStates() = %v, want %v", got, center.data)
		}
	})
}

func TestInitIncState(t *testing.T) {
	name := uuid.New().String()
	status := uuid.New().String()
	total := time.Now().Unix()
	type args struct {
		name   string
		status string
		total  int64
	}
	tests := []struct {
		name string
		args args
		want State
	}{
		{
			"normal",
			args{
				name:   name,
				status: status,
				total:  total,
			},
			State{
				Name:   name,
				Status: status,
				Type:   typeInc,
				Done:   0,
				Total:  total,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitIncState(tt.args.name, tt.args.status, tt.args.total); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitIncState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitListState(t *testing.T) {
	name := uuid.New().String()
	status := uuid.New().String()
	type args struct {
		name   string
		status string
	}
	tests := []struct {
		name string
		args args
		want State
	}{
		{
			"normal",
			args{
				name:   name,
				status: status,
			},
			State{
				Name:   name,
				Status: status,
				Type:   typeList,
				Done:   0,
				Total:  1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitListState(tt.args.name, tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitListState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetState(t *testing.T) {
	type args struct {
		id taskID
		s  State
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"normal",
			args{
				id: uuid.New().String(),
				s:  newState(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetState(tt.args.id, tt.args.s)
			s, err := GetStateByID(tt.args.id)
			assert.Nil(t, err, tt.name)
			assert.Equal(t, s, tt.args.s)
		})
	}
}

func TestState_Finished(t *testing.T) {
	type fields struct {
		Name   string
		Status string
		Type   int
		Done   int64
		Total  int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"true",
			fields{
				Done:  100,
				Total: 100,
			},
			true,
		},
		{
			"false",
			fields{
				Done:  10,
				Total: 100,
			},
			false,
		},
		{
			"not start",
			fields{
				Done:  0,
				Total: 0,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{
				Name:   tt.fields.Name,
				Status: tt.fields.Status,
				Type:   tt.fields.Type,
				Done:   tt.fields.Done,
				Total:  tt.fields.Total,
			}
			if got := s.Finished(); got != tt.want {
				t.Errorf("Finished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestState_IsListType(t *testing.T) {
	rand.Seed(time.Now().Unix())
	var r int
	for r == 0 {
		r = rand.Int()
	}
	type fields struct {
		Name   string
		Status string
		Type   int
		Done   int64
		Total  int64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"true",
			fields{
				Type: typeList,
			},
			true,
		},
		{
			"false",
			fields{
				Type: typeInc,
			},
			false,
		},
		{
			"rand not 0",
			fields{
				Type: r,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{
				Name:   tt.fields.Name,
				Status: tt.fields.Status,
				Type:   tt.fields.Type,
				Done:   tt.fields.Done,
				Total:  tt.fields.Total,
			}
			if got := s.IsListType(); got != tt.want {
				t.Errorf("IsListType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestState_String(t *testing.T) {
	rand.Seed(time.Now().Unix())
	name, status, typ, done, total :=
		uuid.New().String(),
		uuid.New().String(),
		rand.Int(),
		rand.Int63(),
		rand.Int63()

	type fields struct {
		Name   string
		Status string
		Type   int
		Done   int64
		Total  int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"blank",
			fields{},
			"name: , status: , type: 0, 0/0",
		},
		{
			"full",
			fields{
				Name:   name,
				Status: status,
				Type:   typ,
				Done:   done,
				Total:  total,
			},
			fmt.Sprintf("name: %s, status: %s, type: %d, %d/%d", name, status, typ, done, total),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{
				Name:   tt.fields.Name,
				Status: tt.fields.Status,
				Type:   tt.fields.Type,
				Done:   tt.fields.Done,
				Total:  tt.fields.Total,
			}
			if got := s.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateState(t *testing.T) {
	id, done := uuid.New().String(), time.Now().Unix()
	SetState(id, InitIncState("", "", done))
	type args struct {
		id   taskID
		done int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal",
			args: args{
				id,
				done,
			},
		},
		{
			"not exist",
			args{
				id:   uuid.New().String(),
				done: done,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateState(tt.args.id, tt.args.done)
			s, err := GetStateByID(tt.args.id)
			if err != nil {
				assert.Equal(t, ErrTaskNotExist, err)
				return
			}
			assert.Nil(t, err, tt.name)
			assert.Equal(t, s.Done, tt.args.done, tt.name)
			assert.True(t, s.Finished(), tt.name)
		})
	}
}

func TestGetStateByID(t *testing.T) {
	id, s := uuid.New().String(), newState()
	SetState(id, s)

	tests := []struct {
		name    string
		id      taskID
		want    State
		wantErr error
	}{
		{
			name:    "normal",
			id:      id,
			want:    s,
			wantErr: nil,
		},
		{
			name:    "err",
			id:      "",
			want:    newState(),
			wantErr: ErrTaskNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStateByID(tt.id)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("GetStateByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}

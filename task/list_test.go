package task

import (
	"context"
	"errors"
	"testing"

	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
	"github.com/qingstor/noah/pkg/types"
)

func TestListDirTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testErr := errors.New("test error")
	it := typ.NewObjectIterator(nil)

	cases := []struct {
		name     string
		wantIter *typ.ObjectIterator
		hasErr   bool
	}{
		{
			name:     "normal",
			wantIter: it,
			hasErr:   false,
		},
		{
			name:     "fault",
			wantIter: nil,
			hasErr:   true,
		},
	}

	for _, tt := range cases {
		ctx := context.Background()

		store := mock.NewMockDirLister(ctrl)
		testPath := uuid.New().String()

		task := &ListDirTask{}
		task.SetID(uuid.New().String())
		task.SetFault(fault.New())
		task.SetDirLister(store)
		task.SetPath(testPath)

		store.EXPECT().ListDirWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, path string, opts ...*typ.Pair) (*typ.ObjectIterator, error) {
				assert.Equal(t, testPath, path, tt.name)
				if tt.hasErr {
					return nil, testErr
				}
				return tt.wantIter, nil
			})

		err := task.run(ctx)
		if tt.hasErr {
			assert.NotNil(t, err, tt.name)
			ae := &types.Unhandled{}
			assert.True(t, errors.As(err, &ae), tt.name)
			assert.False(t, task.ValidateObjectIter(), tt.name)
		} else {
			assert.Nil(t, err, tt.name)
			assert.Equal(t, tt.wantIter, task.GetObjectIter(), tt.name)
		}
	}
}

func TestListPrefixTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testErr := errors.New("test error")
	it := typ.NewObjectIterator(nil)

	cases := []struct {
		name     string
		wantIter *typ.ObjectIterator
		hasErr   bool
	}{
		{
			name:     "normal",
			wantIter: it,
			hasErr:   false,
		},
		{
			name:     "fault",
			wantIter: nil,
			hasErr:   true,
		},
	}

	for _, tt := range cases {
		ctx := context.Background()

		store := mock.NewMockPrefixLister(ctrl)
		testPath := uuid.New().String()

		task := &ListPrefixTask{}
		task.SetID(uuid.New().String())
		task.SetFault(fault.New())
		task.SetPrefixLister(store)
		task.SetPath(testPath)

		store.EXPECT().ListPrefixWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, path string, opts ...*typ.Pair) (*typ.ObjectIterator, error) {
				assert.Equal(t, testPath, path, tt.name)
				if tt.hasErr {
					return nil, testErr
				}
				return tt.wantIter, nil
			})

		err := task.run(ctx)
		if tt.hasErr {
			assert.NotNil(t, err, tt.name)
			ae := &types.Unhandled{}
			assert.True(t, errors.As(err, &ae), tt.name)
			assert.False(t, task.ValidateObjectIter(), tt.name)
		} else {
			assert.Nil(t, err, tt.name)
			assert.Equal(t, tt.wantIter, task.GetObjectIter(), tt.name)
		}
	}
}

func TestListSegmentTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testErr := errors.New("test error")
	it := typ.NewSegmentIterator(nil)

	cases := []struct {
		name     string
		wantIter *typ.SegmentIterator
		hasErr   bool
	}{
		{
			name:     "normal",
			wantIter: it,
			hasErr:   false,
		},
		{
			name:     "fault",
			wantIter: nil,
			hasErr:   true,
		},
	}

	for _, tt := range cases {
		ctx := context.Background()

		segmenter := mock.NewMockPrefixSegmentsLister(ctrl)
		testPath := uuid.New().String()

		task := ListSegmentTask{}
		task.SetFault(fault.New())
		task.SetPrefixSegmentsLister(segmenter)
		task.SetPath(testPath)

		segmenter.EXPECT().ListPrefixSegmentsWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, path string, opts ...*typ.Pair) (*typ.SegmentIterator, error) {
				assert.Equal(t, testPath, path, tt.name)
				if tt.hasErr {
					return nil, testErr
				}
				return tt.wantIter, nil
			})

		err := task.run(ctx)
		if tt.hasErr {
			assert.NotNil(t, err, tt.name)
			ae := &types.Unhandled{}
			assert.True(t, errors.As(err, &ae), tt.name)
			assert.False(t, task.ValidateSegmentIter(), tt.name)
		} else {
			assert.Nil(t, err, tt.name)
			assert.Equal(t, tt.wantIter, task.GetSegmentIter(), tt.name)
		}
	}
}

func TestListStorageTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testErr := errors.New("test error")
	it := typ.NewStoragerIterator(nil)

	cases := []struct {
		name     string
		zone     string
		wantIter *typ.StoragerIterator
		hasErr   bool
	}{
		{
			name:     "normal without zone",
			wantIter: it,
			hasErr:   false,
		},
		{
			name:     "normal with zone",
			zone:     "zone1",
			wantIter: it,
			hasErr:   false,
		},
		{
			name:     "fault",
			wantIter: nil,
			hasErr:   true,
		},
	}

	for _, tt := range cases {
		ctx := context.Background()

		srv := mock.NewMockServicer(ctrl)

		task := ListStorageTask{}
		task.SetFault(fault.New())
		task.SetService(srv)
		if tt.zone != "" {
			task.SetZone(tt.zone)
		}

		srv.EXPECT().ListWithContext(gomock.Eq(ctx), gomock.Any()).
			DoAndReturn(func(ctx context.Context, pairs ...*typ.Pair) (*typ.StoragerIterator, error) {
				if tt.zone != "" {
					assert.Equal(t, tt.zone, pairs[0].Value.(string), tt.name)
				} else {
					assert.Equal(t, 0, len(pairs), tt.name)
				}

				if tt.hasErr {
					return nil, testErr
				}
				return tt.wantIter, nil
			})

		err := task.run(ctx)
		if tt.hasErr {
			assert.NotNil(t, err, tt.name)
			ae := &types.Unhandled{}
			assert.True(t, errors.As(err, &ae), tt.name)
			assert.False(t, task.ValidateStorageIter(), tt.name)
		} else {
			assert.Nil(t, err, tt.name)
			assert.Equal(t, tt.wantIter, task.GetStorageIter(), tt.name)
		}
	}

}

package task

import (
	"context"
	"testing"

	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestListDirTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	store := mock.NewMockDirLister(ctrl)
	testPath := uuid.New().String()

	task := &ListDirTask{}
	task.SetID(uuid.New().String())
	task.SetFault(fault.New())
	task.SetDirLister(store)
	task.SetPath(testPath)

	store.EXPECT().ListDirWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, path string, opts ...*typ.Pair) error {
			assert.Equal(t, testPath, path)
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestListPrefixTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	store := mock.NewMockPrefixLister(ctrl)
	testPath := uuid.New().String()

	task := &ListPrefixTask{}
	task.SetID(uuid.New().String())
	task.SetFault(fault.New())
	task.SetPrefixLister(store)
	task.SetPath(testPath)

	store.EXPECT().ListPrefixWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, path string, opts ...*typ.Pair) error {
			assert.Equal(t, testPath, path)
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestListSegmentTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	segmenter := mock.NewMockPrefixSegmentsLister(ctrl)
	testPath := uuid.New().String()

	task := ListSegmentTask{}
	task.SetFault(fault.New())
	task.SetPrefixSegmentsLister(segmenter)
	task.SetPath(testPath)

	segmenter.EXPECT().ListPrefixSegmentsWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, path string, opts ...*typ.Pair) error {
			assert.Equal(t, testPath, path)
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestListStorageTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	srv := mock.NewMockServicer(ctrl)
	zone := uuid.New().String()

	task := ListStorageTask{}
	task.SetFault(fault.New())
	task.SetService(srv)
	task.SetZone(zone)

	srv.EXPECT().ListWithContext(gomock.Eq(ctx), gomock.Any()).
		Do(func(ctx context.Context, pairs ...*typ.Pair) error {
			assert.Equal(t, zone, pairs[0].Value.(string))
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

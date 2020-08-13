package task

import (
	"context"
	"errors"
	"testing"

	"github.com/Xuanwo/storage"
	typ "github.com/Xuanwo/storage/types"
	"github.com/Xuanwo/storage/types/info"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
	"github.com/qingstor/noah/pkg/types"
)

func TestStatFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	store := mock.NewMockStorager(ctrl)
	expectedPath := uuid.New().String()

	task := StatFileTask{}
	task.SetFault(fault.New())
	task.SetStorage(store)
	task.SetPath(expectedPath)

	store.EXPECT().StatWithContext(gomock.Eq(ctx), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string) (o *typ.Object, err error) {
			assert.Equal(t, expectedPath, path)
			return &typ.Object{}, nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
	assert.NotNil(t, task.GetObject())
}

func TestStatStorageTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	store := mock.NewMockStorager(ctrl)
	store.EXPECT().String().DoAndReturn(func() string {
		return ""
	}).AnyTimes()

	t.Run("insufficient ability error return", func(t *testing.T) {
		task := StatStorageTask{}
		task.SetFault(fault.New())
		task.SetStorage(store)

		task.run(ctx)
		assert.NotEmpty(t, task.GetFault().Error())
		tarErr := &types.StorageInsufficientAbility{}
		assert.True(t, errors.As(task.GetFault(), &tarErr))
	})

	t.Run("normal return", func(t *testing.T) {
		statistician := mock.NewMockStatistician(ctrl)
		storeComb := struct {
			storage.Storager
			storage.Statistician
		}{
			store,
			statistician,
		}
		statistician.EXPECT().StatisticalWithContext(ctx).
			DoAndReturn(func(ctx context.Context) (info.StorageStatistic, error) {
				return info.NewStorageStatistic(), nil
			})

		task := StatStorageTask{}
		task.SetFault(fault.New())
		task.SetStorage(storeComb)

		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
		assert.NotNil(t, task.GetStorageInfo())
	})
}

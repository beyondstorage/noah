package task

import (
	"context"
	"testing"

	"github.com/aos-dev/go-storage/v2"
	"github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestCreateStorageTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	service := mock.NewMockServicer(ctrl)
	storageName := uuid.New().String()
	zone := uuid.New().String()

	task := CreateStorageTask{}
	task.SetFault(fault.New())
	task.SetService(service)
	task.SetStorageName(storageName)
	task.SetZone(zone)

	service.EXPECT().CreateWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).Do(func(ctx context.Context, name string, pairs ...*types.Pair) (storage.Storager, error) {
		assert.Equal(t, storageName, name)
		assert.Equal(t, zone, pairs[0].Value.(string))
		return nil, nil
	})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

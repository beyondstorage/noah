package task

import (
	"context"
	"testing"

	"github.com/Xuanwo/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestReachFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	reacher := mock.NewMockReacher(ctrl)
	reachPath := uuid.New().String()
	reachExpire := 1024
	reachedURL := uuid.New().String()

	task := ReachFileTask{}
	task.SetFault(fault.New())
	task.SetReacher(reacher)
	task.SetPath(reachPath)
	task.SetExpire(reachExpire)

	reacher.EXPECT().ReachWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, pairs ...*types.Pair) (url string, err error) {
			assert.Equal(t, reachPath, path)
			assert.Equal(t, reachExpire, pairs[0].Value.(int))
			return reachedURL, nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
	assert.Equal(t, reachedURL, task.GetURL())
}

package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/Xuanwo/navvy"
	"github.com/Xuanwo/storage"
	typ "github.com/Xuanwo/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestMoveDirTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	t.Run("normal", func(t *testing.T) {
		sche := mock.NewMockScheduler(ctrl)
		srcStore := mock.NewMockStorager(ctrl)
		srcLister := mock.NewMockDirLister(ctrl)
		src := struct {
			storage.Storager
			storage.DirLister
		}{
			srcStore, srcLister,
		}
		dstStore := mock.NewMockStorager(ctrl)

		task := MoveDirTask{}
		task.SetFault(fault.New())
		task.SetPool(navvy.NewPool(10))
		task.SetSourcePath("source")
		task.SetSourceStorage(src)
		task.SetDestinationPath("destination")
		task.SetDestinationStorage(dstStore)
		task.SetScheduler(sche)
		task.SetCheckMD5(false)

		obj := &typ.Object{Name: "obj-name"}

		sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
			switch v := task.(type) {
			case *ListDirTask:
				v.validateInput()
				v.GetFileFunc()(obj)
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		})

		sche.EXPECT().Async(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
			switch v := task.(type) {
			case *MoveFileTask:
				v.validateInput()
				assert.Equal(t, obj.Name, v.GetSourcePath())
				assert.Equal(t, obj.Name, v.GetDestinationPath())
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		})

		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
	})
}

func TestMoveFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	t.Run("normal", func(t *testing.T) {
		sche := mock.NewMockScheduler(ctrl)
		srcStore := mock.NewMockStorager(ctrl)
		dstStore := mock.NewMockStorager(ctrl)

		task := MoveFileTask{}
		task.SetFault(fault.New())
		task.SetPool(navvy.NewPool(10))
		task.SetSourcePath("source")
		task.SetSourceStorage(srcStore)
		task.SetDestinationPath("destination")
		task.SetDestinationStorage(dstStore)
		task.SetScheduler(sche)
		task.SetCheckMD5(false)

		sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
			switch v := task.(type) {
			case *CopyFileTask:
				v.validateInput()
			case *DeleteFileTask:
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()

		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
	})
}

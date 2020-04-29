package task

import (
	"fmt"
	"testing"

	"github.com/Xuanwo/navvy"
	"github.com/Xuanwo/storage"
	"github.com/Xuanwo/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestSyncTask_run(t *testing.T) {
	t.Run("without flag", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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
		sourcePath := uuid.New().String()
		dstPath := uuid.New().String()

		task := SyncTask{}
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetFault(fault.New())
		task.SetSourcePath(sourcePath)
		task.SetSourceStorage(src)
		task.SetDestinationStorage(dstStore)
		task.SetDestinationPath(dstPath)
		task.SetDryRun(false)
		task.SetDryRunFunc(nil)
		task.SetExisting(false)
		task.SetIgnoreExisting(false)
		task.SetRecursive(false)
		task.SetUpdate(false)

		srcStore.EXPECT().String().Do(func() {}).AnyTimes()
		dstStore.EXPECT().String().Do(func() {}).AnyTimes()
		sche.EXPECT().Sync(gomock.Any()).Do(func(task navvy.Task) {
			switch v := task.(type) {
			case *ListDirTask:
				v.SetDirFunc(nil)
				v.SetFileFunc(nil)
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("with all flags", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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
		sourcePath := uuid.New().String()
		dstPath := uuid.New().String()

		task := SyncTask{}
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetFault(fault.New())
		task.SetSourcePath(sourcePath)
		task.SetSourceStorage(src)
		task.SetDestinationStorage(dstStore)
		task.SetDestinationPath(dstPath)
		task.SetDryRun(false)
		task.SetDryRunFunc(nil)
		task.SetExisting(true)
		task.SetIgnoreExisting(true)
		task.SetRecursive(false)
		task.SetUpdate(true)

		sche.EXPECT().Sync(gomock.Any()).Do(func(task navvy.Task) {
			switch v := task.(type) {
			case *ListDirTask:
				v.SetFileFunc(nil)
				v.SetDirFunc(nil)
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("with dry-run", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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
		sourcePath := uuid.New().String()
		dstPath := uuid.New().String()

		task := SyncTask{}
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetFault(fault.New())
		task.SetSourcePath(sourcePath)
		task.SetSourceStorage(src)
		task.SetDestinationStorage(dstStore)
		task.SetDestinationPath(dstPath)
		task.SetDryRun(true)
		task.SetDryRunFunc(func(o *types.Object) {
			t.Log(o.Name)
		})
		task.SetExisting(false)
		task.SetIgnoreExisting(false)
		task.SetRecursive(false)
		task.SetUpdate(false)

		sche.EXPECT().Sync(gomock.Any()).Do(func(task navvy.Task) {
			switch v := task.(type) {
			case *ListDirTask:
				v.SetDirFunc(nil)
				v.SetFileFunc(nil)
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("with recursive", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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
		sourcePath := uuid.New().String()
		dstPath := uuid.New().String()

		task := SyncTask{}
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetFault(fault.New())
		task.SetSourcePath(sourcePath)
		task.SetSourceStorage(src)
		task.SetDestinationStorage(dstStore)
		task.SetDestinationPath(dstPath)
		task.SetDryRun(true)
		task.SetDryRunFunc(func(o *types.Object) {
			t.Log(o.Name)
		})
		task.SetExisting(false)
		task.SetIgnoreExisting(false)
		task.SetRecursive(true)
		task.SetUpdate(false)

		sche.EXPECT().Sync(gomock.Any()).Do(func(task navvy.Task) {
			switch v := task.(type) {
			case *ListDirTask:
				v.SetDirFunc(nil)
				v.SetFileFunc(nil)
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})
}

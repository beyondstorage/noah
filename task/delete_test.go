package task

import (
	"fmt"
	"testing"

	"github.com/Xuanwo/navvy"
	"github.com/Xuanwo/storage"
	"github.com/Xuanwo/storage/pkg/segment"
	"github.com/Xuanwo/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestDeleteFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockStorager(ctrl)
	testPath := uuid.New().String()

	task := DeleteFileTask{}
	task.SetFault(fault.New())
	task.SetStorage(store)
	task.SetPath(testPath)

	store.EXPECT().Delete(gomock.Any()).Do(func(name string) error {
		assert.Equal(t, testPath, name)
		return nil
	})

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

func TestDeleteDirTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mixer := struct {
		storage.Storager
		storage.DirLister
	}{
		mock.NewMockStorager(ctrl),
		mock.NewMockDirLister(ctrl),
	}

	sche := mock.NewMockScheduler(ctrl)
	path := uuid.New().String()

	task := DeleteDirTask{}
	task.SetFault(fault.New())
	task.SetPool(navvy.NewPool(10))
	task.SetScheduler(sche)
	task.SetPath(path)
	task.SetStorage(mixer)

	sche.EXPECT().Sync(gomock.Any()).Do(func(task navvy.Task) {
		switch v := task.(type) {
		case *ListDirTask:
			v.validateInput()
		case *DeleteFileTask:
			v.validateInput()
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

func TestDeletePrefixTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mixer := struct {
		storage.Storager
		storage.PrefixLister
	}{
		mock.NewMockStorager(ctrl),
		mock.NewMockPrefixLister(ctrl),
	}

	sche := mock.NewMockScheduler(ctrl)
	path := uuid.New().String()

	task := DeletePrefixTask{}
	task.SetFault(fault.New())
	task.SetPool(navvy.NewPool(10))
	task.SetScheduler(sche)
	task.SetPath(path)
	task.SetStorage(mixer)

	sche.EXPECT().Sync(gomock.Any()).Do(func(task navvy.Task) {
		switch v := task.(type) {
		case *ListPrefixTask:
			v.validateInput()
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

func TestDeleteSegmentTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	segmenter := mock.NewMockPrefixSegmentsLister(ctrl)
	seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

	task := DeleteSegmentTask{}
	task.SetFault(fault.New())
	task.SetPrefixSegmentsLister(segmenter)
	task.SetSegment(seg)

	segmenter.EXPECT().AbortSegment(gomock.Any()).Do(func(inputSeg segment.Segment) error {
		assert.Equal(t, seg, inputSeg)
		return nil
	})

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

func TestNewDeleteStorageTask(t *testing.T) {
	t.Run("delete without force", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := mock.NewMockServicer(ctrl)
		storageName := uuid.New().String()

		task := DeleteStorageTask{}
		task.SetFault(fault.New())
		task.SetService(srv)
		task.SetStorageName(storageName)
		task.SetForce(false)
		task.SetZone("")

		srv.EXPECT().Delete(gomock.Any()).Do(func(name string) error {
			assert.Equal(t, storageName, name)
			return nil
		})

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("delete with force", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sche := mock.NewMockScheduler(ctrl)
		srv := mock.NewMockServicer(ctrl)
		store := mock.NewMockStorager(ctrl)
		storageName := uuid.New().String()

		task := DeleteStorageTask{}
		task.SetFault(fault.New())
		task.SetService(srv)
		task.SetStorageName(storageName)
		task.SetForce(true)
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetZone("")

		srv.EXPECT().Delete(gomock.Any()).Do(func(name string) error {
			assert.Equal(t, storageName, name)
			return nil
		})
		srv.EXPECT().Get(gomock.Any()).DoAndReturn(func(name string, pairs ...*types.Pair) (storage.Storager, error) {
			assert.Equal(t, storageName, name)
			return store, nil
		})
		sche.EXPECT().Async(gomock.Any()).Do(func(task navvy.Task) {
			switch v := task.(type) {
			case *DeletePrefixTask:
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()
		sche.EXPECT().Wait()

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("delete with segmenter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		sche := mock.NewMockScheduler(ctrl)
		srv := mock.NewMockServicer(ctrl)
		store := mock.NewMockStorager(ctrl)
		segmenter := mock.NewMockPrefixSegmentsLister(ctrl)
		storageName := uuid.New().String()

		task := DeleteStorageTask{}
		task.SetFault(fault.New())
		task.SetService(srv)
		task.SetStorageName(storageName)
		task.SetForce(true)
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetZone("")

		srv.EXPECT().Delete(gomock.Any()).Do(func(name string) error {
			assert.Equal(t, storageName, name)
			return nil
		})
		srv.EXPECT().Get(gomock.Any()).DoAndReturn(func(name string, pairs ...*types.Pair) (storage.Storager, error) {
			assert.Equal(t, storageName, name)
			return struct {
				storage.Storager
				storage.PrefixSegmentsLister
			}{
				store,
				segmenter,
			}, nil
		})
		sche.EXPECT().Async(gomock.Any()).Do(func(task navvy.Task) {
			switch v := task.(type) {
			case *DeletePrefixTask:
				v.validateInput()
			case *ListSegmentTask:
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()
		sche.EXPECT().Wait()

		task.run()
		assert.Empty(t, task.GetFault().Error())
	})
}

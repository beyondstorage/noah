package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/Xuanwo/navvy"
	"github.com/aos-dev/go-storage/v2"
	"github.com/aos-dev/go-storage/v2/pkg/segment"
	"github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestDeleteFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	store := mock.NewMockStorager(ctrl)
	testPath := uuid.New().String()

	task := DeleteFileTask{}
	task.SetFault(fault.New())
	task.SetStorage(store)
	task.SetPath(testPath)

	store.EXPECT().DeleteWithContext(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, name string) error {
		assert.Equal(t, testPath, name)
		return nil
	})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestDeleteDirTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

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

	obj := &types.Object{Name: "obj-name"}
	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *ListDirTask:
			v.validateInput()
			v.GetFileFunc()(obj)
		case *DeleteFileTask:
			v.validateInput()
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	sche.EXPECT().Async(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *DeleteFileTask:
			v.validateInput()
			assert.Equal(t, obj.Name, v.GetPath())
		}
	})
	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestDeletePrefixTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

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

	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *ListPrefixTask:
			v.validateInput()
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestDeleteSegmentTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	segmenter := mock.NewMockPrefixSegmentsLister(ctrl)
	seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

	task := DeleteSegmentTask{}
	task.SetFault(fault.New())
	task.SetPrefixSegmentsLister(segmenter)
	task.SetSegment(seg)

	segmenter.EXPECT().AbortSegmentWithContext(gomock.Eq(ctx), gomock.Any()).
		Do(func(ctx context.Context, inputSeg segment.Segment) error {
			assert.Equal(t, seg, inputSeg)
			return nil
		})

	task.run(context.Background())
	assert.Empty(t, task.GetFault().Error())
}

func TestNewDeleteStorageTask(t *testing.T) {
	t.Run("delete without force", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()

		srv := mock.NewMockServicer(ctrl)
		storageName := uuid.New().String()

		task := DeleteStorageTask{}
		task.SetFault(fault.New())
		task.SetService(srv)
		task.SetStorageName(storageName)
		task.SetForce(false)
		task.SetZone("")

		srv.EXPECT().DeleteWithContext(gomock.Eq(ctx), gomock.Any()).
			Do(func(ctx context.Context, name string) error {
				assert.Equal(t, storageName, name)
				return nil
			})

		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("delete with force", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()

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

		srv.EXPECT().DeleteWithContext(gomock.Eq(ctx), gomock.Any()).
			Do(func(ctx context.Context, name string) error {
				assert.Equal(t, storageName, name)
				return nil
			})
		srv.EXPECT().GetWithContext(gomock.Eq(ctx), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string, pairs ...*types.Pair) (storage.Storager, error) {
				assert.Equal(t, storageName, name)
				return store, nil
			})
		sche.EXPECT().Async(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
			switch v := task.(type) {
			case *DeletePrefixTask:
				v.validateInput()
			default:
				panic(fmt.Errorf("unexpected task %v", v))
			}
		}).AnyTimes()
		sche.EXPECT().Wait()

		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
	})

	t.Run("delete with segmenter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()

		path, id := uuid.New().String(), uuid.New().String()
		sche := mock.NewMockScheduler(ctrl)
		srv := mock.NewMockServicer(ctrl)
		store := mock.NewMockStorager(ctrl)
		segmenter := mock.NewMockPrefixSegmentsLister(ctrl)
		seg := segment.NewIndexBasedSegment(path, id)
		storageName := uuid.New().String()

		task := DeleteStorageTask{}
		task.SetFault(fault.New())
		task.SetService(srv)
		task.SetStorageName(storageName)
		task.SetForce(true)
		task.SetPool(navvy.NewPool(10))
		task.SetScheduler(sche)
		task.SetZone("")

		srv.EXPECT().DeleteWithContext(gomock.Eq(ctx), gomock.Any()).
			Do(func(ctx context.Context, name string) error {
				assert.Equal(t, storageName, name)
				return nil
			})
		srv.EXPECT().GetWithContext(gomock.Eq(ctx), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string, pairs ...*types.Pair) (storage.Storager, error) {
				assert.Equal(t, storageName, name)
				return struct {
					storage.Storager
					storage.PrefixSegmentsLister
				}{
					store,
					segmenter,
				}, nil
			})
		sche.EXPECT().Async(gomock.Eq(ctx), gomock.Any()).
			Do(func(ctx context.Context, task navvy.Task) {
				switch v := task.(type) {
				case *DeletePrefixTask:
					v.validateInput()
				case *ListSegmentTask:
					v.validateInput()
					v.GetSegmentFunc()(seg)
				case *DeleteSegmentTask:
					v.validateInput()
					assert.Equal(t, seg, v.GetSegment())
					assert.Equal(t, struct {
						storage.Storager
						storage.PrefixSegmentsLister
					}{
						store,
						segmenter,
					}, v.GetPrefixSegmentsLister())
				default:
					panic(fmt.Errorf("unexpected task %v", v))
				}
			}).AnyTimes()
		sche.EXPECT().Wait()

		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
	})
}
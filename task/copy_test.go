package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/Xuanwo/navvy"
	"github.com/Xuanwo/storage"
	"github.com/Xuanwo/storage/pkg/segment"
	typ "github.com/Xuanwo/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/constants"
	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestCopyDirTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	t.Run("normal", func(t *testing.T) {
		sche := mock.NewMockScheduler(ctrl)
		store := mock.NewMockStorager(ctrl)
		lister := mock.NewMockDirLister(ctrl)
		srcStore := struct {
			storage.Storager
			storage.DirLister
		}{
			store, lister,
		}
		dstStore := mock.NewMockStorager(ctrl)

		task := CopyDirTask{}
		task.SetFault(fault.New())
		task.SetPool(navvy.NewPool(10))
		task.SetSourcePath("source")
		task.SetSourceStorage(srcStore)
		task.SetDestinationPath("destination")
		task.SetDestinationStorage(dstStore)
		task.SetScheduler(sche)
		task.SetCheckTasks(nil)

		sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
			_, ok := task.(*ListDirTask)
			assert.True(t, ok)
		})
		task.run(ctx)
		assert.Empty(t, task.GetFault().Error())
	})
}

func TestCopyFileTask_run(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()

		cases := []struct {
			name string
			size int64
		}{
			{
				"large file",
				constants.MaximumAutoMultipartSize + 1,
			},
			{
				"small file",
				constants.MaximumAutoMultipartSize - 1,
			},
		}

		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				sche := mock.NewMockScheduler(ctrl)
				srcStore := mock.NewMockStorager(ctrl)
				srcPath := uuid.New().String()
				// if dstStore not implement IndexSegmenter,
				// CopySmallFileTask will always be called
				dstStore := struct {
					storage.Storager
					storage.IndexSegmenter
				}{
					mock.NewMockStorager(ctrl),
					mock.NewMockIndexSegmenter(ctrl),
				}
				dstPath := uuid.New().String()

				task := &CopyFileTask{}
				task.SetFault(fault.New())
				task.SetPool(navvy.NewPool(10))
				task.SetScheduler(sche)
				task.SetCheckTasks([]func(t navvy.Task) navvy.Task{
					func(t navvy.Task) navvy.Task {
						res := &IsDestinationObjectExistTask{}
						res.SetResult(true)
						return res
					},
				})
				task.SetSourcePath(srcPath)
				task.SetSourceStorage(srcStore)
				task.SetDestinationPath(dstPath)
				task.SetDestinationStorage(dstStore)

				sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
					switch v := task.(type) {
					case *BetweenStorageCheckTask:
						v.SetSourceObject(&typ.Object{Name: srcPath, Size: tt.size})
						v.SetDestinationObject(&typ.Object{Name: dstPath})
					case *CopyLargeFileTask:
						assert.True(t, tt.size >= constants.MaximumAutoMultipartSize)
					case *CopySmallFileTask:
						assert.True(t, tt.size < constants.MaximumAutoMultipartSize)
					}
				}).AnyTimes()

				task.run(ctx)
				assert.Empty(t, task.GetFault().Error())
			})
		}
	})
}

func TestCopySmallFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	sche := mock.NewMockScheduler(ctrl)
	srcStore := mock.NewMockStorager(ctrl)
	srcPath := uuid.New().String()
	dstStore := mock.NewMockStorager(ctrl)
	dstPath := uuid.New().String()

	task := &CopySmallFileTask{}
	task.SetFault(fault.New())
	task.SetPool(navvy.NewPool(10))
	task.SetSourcePath(srcPath)
	task.SetSourceStorage(srcStore)
	task.SetDestinationPath(dstPath)
	task.SetDestinationStorage(dstStore)
	task.SetScheduler(sche)
	task.SetSize(1024)
	task.SetCheckMD5(true)

	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *MD5SumFileTask:
			assert.Equal(t, srcPath, v.GetPath())
			assert.Equal(t, int64(0), v.GetOffset())
			v.SetMD5Sum([]byte("string"))
		case *CopySingleFileTask:
			assert.Equal(t, []byte("string"), v.GetMD5Sum())
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestCopyLargeFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	sche := mock.NewMockScheduler(ctrl)
	srcStore := mock.NewMockStorager(ctrl)
	srcPath := uuid.New().String()
	dstStore := mock.NewMockStorager(ctrl)
	dstSegmenter := mock.NewMockIndexSegmenter(ctrl)
	dstPath := uuid.New().String()
	seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

	cpTask := &CopyLargeFileTask{}
	cpTask.SetID(uuid.New().String())
	cpTask.SetPool(navvy.NewPool(10))
	cpTask.SetSourcePath(srcPath)
	cpTask.SetSourceStorage(srcStore)
	cpTask.SetDestinationPath(dstPath)
	cpTask.SetDestinationStorage(struct {
		storage.Storager
		storage.IndexSegmenter
	}{
		dstStore,
		dstSegmenter,
	})
	cpTask.SetScheduler(sche)
	cpTask.SetFault(fault.New())
	// 50G
	cpTask.SetTotalSize(10 * constants.MaximumPartSize)

	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *SegmentInitTask:
			assert.Equal(t, dstPath, v.GetPath())
			v.SetSegment(seg)
		case *SegmentCompleteTask:
			assert.Equal(t, dstPath, v.GetPath())
			assert.Equal(t, seg, v.GetSegment())
		default:
			panic(fmt.Errorf("invalid task %v", v))
		}
	}).AnyTimes()

	copyPartialFileTaskCallTime := 0
	sche.EXPECT().Async(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *CopyPartialFileTask:
			part := copyPartialFileTaskCallTime
			assert.Equal(t, srcPath, v.GetSourcePath())
			assert.Equal(t, dstPath, v.GetDestinationPath())
			assert.Equal(t, seg, v.GetSegment())
			assert.Equal(t, part, v.GetIndex())
			assert.Equal(t, cpTask.GetPartSize(), v.GetSize())
			assert.Equal(t, int64(part)*cpTask.GetPartSize(), v.GetOffset())
			v.SetDone(func() bool {
				if copyPartialFileTaskCallTime == 0 {
					return false
				}
				return true
			}())
			copyPartialFileTaskCallTime++
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()
	sche.EXPECT().Wait().Do(func() {})

	cpTask.run(ctx)
	assert.Empty(t, cpTask.GetFault().Error())
}

func TestCopyPartialFileTask_new(t *testing.T) {
	cases := []struct {
		name       string
		totalsize  int64
		offset     int64
		partsize   int64
		expectSize int64
		expectDone bool
	}{
		{
			"middle part",
			1024,
			128,
			128,
			128,
			false,
		},
		{
			"last fulfilled part",
			1024,
			512,
			512,
			512,
			true,
		},
		{
			"last not fulfilled part",
			1024,
			768,
			512,
			256,
			true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			task := &CopyPartialFileTask{}
			task.SetTotalSize(tt.totalsize)
			task.SetOffset(tt.offset)
			task.SetPartSize(tt.partsize)

			task.new()

			assert.Equal(t, tt.expectSize, task.GetSize())
			assert.Equal(t, tt.expectDone, task.GetDone())
		})
	}
}

func TestCopyPartialFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	sche := mock.NewMockScheduler(ctrl)
	srcStore := mock.NewMockStorager(ctrl)
	srcPath := uuid.New().String()
	dstStore := mock.NewMockStorager(ctrl)
	dstSegmenter := mock.NewMockIndexSegmenter(ctrl)
	dstPath := uuid.New().String()
	seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

	task := &CopyPartialFileTask{}
	task.SetFault(fault.New())
	task.SetPool(navvy.NewPool(10))
	task.SetSourcePath(srcPath)
	task.SetSourceStorage(srcStore)
	task.SetDestinationPath(dstPath)
	task.SetDestinationStorage(struct {
		storage.Storager
		storage.IndexSegmenter
	}{
		dstStore,
		dstSegmenter,
	})
	task.SetScheduler(sche)
	task.SetSize(1024)
	task.SetOffset(512)
	task.SetIndex(1)
	task.SetSegment(seg)
	task.SetCheckMD5(true)

	srcStore.EXPECT().String().DoAndReturn(func() string { return "src" }).AnyTimes()
	dstStore.EXPECT().String().DoAndReturn(func() string { return "dst" }).AnyTimes()

	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		t.Logf("Got task %v", task)

		switch v := task.(type) {
		case *MD5SumFileTask:
			assert.Equal(t, srcPath, v.GetPath())
			assert.Equal(t, int64(512), v.GetOffset())
			v.SetMD5Sum([]byte("string"))
		case *SegmentFileCopyTask:
			assert.Equal(t, []byte("string"), v.GetMD5Sum())
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestCopyStreamTask_new(t *testing.T) {
	task := &CopyStreamTask{}
	task.new()
	assert.True(t, task.ValidateBytesPool())
}

func TestCopyStreamTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	sche := mock.NewMockScheduler(ctrl)
	srcStore := mock.NewMockStorager(ctrl)
	srcPath := uuid.New().String()
	dstStore := mock.NewMockStorager(ctrl)
	dstSegmenter := mock.NewMockIndexSegmenter(ctrl)
	dstPath := uuid.New().String()
	seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

	task := &CopyStreamTask{}
	task.new()
	task.SetID(uuid.New().String())
	task.SetPool(navvy.NewPool(10))
	task.SetSourcePath(srcPath)
	task.SetSourceStorage(srcStore)
	task.SetDestinationPath(dstPath)
	task.SetDestinationStorage(struct {
		storage.Storager
		storage.IndexSegmenter
	}{
		dstStore,
		dstSegmenter,
	})
	task.SetScheduler(sche)
	task.SetFault(fault.New())

	initSegmentStreamTaskCallTime := 0
	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *SegmentInitTask:
			assert.Equal(t, dstPath, v.GetPath())
			v.SetSegment(seg)
		case *SegmentCompleteTask:
			assert.Equal(t, dstPath, v.GetPath())
			assert.Equal(t, seg, v.GetSegment())
		case *InitSegmentStreamTask:
			v.SetSize(1024)
			v.SetContent(bytes.NewBuffer(nil))
			v.SetDone(func() bool {
				if initSegmentStreamTaskCallTime == 0 {
					return false
				}
				return true
			}())
			initSegmentStreamTaskCallTime++
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	copyPartialStreamTaskCallTime := 0
	sche.EXPECT().Async(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *CopyPartialStreamTask:
			index := copyPartialStreamTaskCallTime
			assert.Equal(t, dstPath, v.GetDestinationPath())
			assert.Equal(t, seg, v.GetSegment())
			assert.Equal(t, int64(index*1024), v.GetOffset())
			assert.Equal(t, index, v.GetIndex())
			copyPartialStreamTaskCallTime++
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()
	sche.EXPECT().Wait().Do(func() {})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestCopyPartialStreamTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	sche := mock.NewMockScheduler(ctrl)
	dstPath := uuid.New().String()
	dstStore := mock.NewMockStorager(ctrl)
	dstSegmenter := mock.NewMockIndexSegmenter(ctrl)
	seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

	task := CopyPartialStreamTask{}
	task.SetPool(navvy.NewPool(10))
	task.SetScheduler(sche)
	task.SetFault(fault.New())
	task.SetDestinationPath(dstPath)
	task.SetDestinationStorage(struct {
		storage.Storager
		storage.IndexSegmenter
	}{
		dstStore,
		dstSegmenter,
	})
	task.SetContent(bytes.NewBuffer(nil))
	task.SetOffset(0)
	task.SetSegment(seg)
	task.SetSize(0)
	task.SetIndex(1)
	task.SetCheckMD5(true)

	sche.EXPECT().Sync(gomock.Eq(ctx), gomock.Any()).Do(func(ctx context.Context, task navvy.Task) {
		switch v := task.(type) {
		case *MD5SumStreamTask:
			v.validateInput()
			v.SetMD5Sum([]byte("string"))
		case *SegmentStreamCopyTask:
			v.validateInput()
			assert.Equal(t, []byte("string"), v.GetMD5Sum())
		default:
			panic(fmt.Errorf("unexpected task %v", v))
		}
	}).AnyTimes()

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestCopySingleFileTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	srcReader := mock.NewMockReadCloser(ctrl)

	srcStore := mock.NewMockStorager(ctrl)
	srcPath := uuid.New().String()
	dstStore := mock.NewMockStorager(ctrl)
	dstPath := uuid.New().String()

	task := CopySingleFileTask{}
	task.SetID(uuid.New().String())
	task.SetFault(fault.New())
	task.SetSourcePath(srcPath)
	task.SetSourceStorage(srcStore)
	task.SetDestinationPath(dstPath)
	task.SetDestinationStorage(dstStore)
	task.SetSize(1024)

	srcReader.EXPECT().Close().Do(func() {})
	srcStore.EXPECT().ReadWithContext(gomock.Eq(ctx), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, pairs ...*typ.Pair) (r io.ReadCloser, err error) {
			assert.Equal(t, srcPath, path)
			return srcReader, nil
		})
	dstStore.EXPECT().WriteWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, r io.Reader, pairs ...*typ.Pair) (err error) {
			assert.Equal(t, dstPath, path)
			assert.Equal(t, int64(1024), pairs[0].Value.(int64))
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

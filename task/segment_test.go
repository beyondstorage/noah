package task

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/aos-dev/go-storage/v2/pkg/segment"
	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestSegmentInitTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	t.Run("normal", func(t *testing.T) {
		store := mock.NewMockIndexSegmenter(ctrl)
		path := uuid.New().String()
		seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

		task := SegmentInitTask{}
		task.SetIndexSegmenter(store)
		task.SetPath(path)
		task.SetPartSize(1000)

		store.EXPECT().InitIndexSegmentWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, inputPath string, pairs ...*typ.Pair) (segment.Segment, error) {
				assert.Equal(t, inputPath, path)
				return seg, nil
			},
		)

		task.run(ctx)

		assert.Equal(t, seg, task.GetSegment())
	})

	t.Run("init segment returned error", func(t *testing.T) {
		store := mock.NewMockIndexSegmenter(ctrl)
		path := uuid.New().String()

		task := SegmentInitTask{}
		task.SetFault(fault.New())
		task.SetIndexSegmenter(store)
		task.SetPath(path)
		task.SetPartSize(1000)

		store.EXPECT().InitIndexSegmentWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, inputPath string, pairs ...*typ.Pair) (string, error) {
				assert.Equal(t, inputPath, path)
				return "", errors.New("test")
			},
			)

		task.run(ctx)

		assert.False(t, task.ValidateSegment())
		assert.True(t, task.GetFault().HasError())
	})
}

func TestSegmentFileCopyTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	srcStore := mock.NewMockStorager(ctrl)
	srcPath := uuid.New().String()
	srcReader := mock.NewMockReadCloser(ctrl)
	srcSize := int64(1024)
	dstSegment := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())
	dstSegmenter := mock.NewMockIndexSegmenter(ctrl)
	dstPath := uuid.New().String()

	task := SegmentFileCopyTask{}
	task.SetID(uuid.New().String())
	task.SetFault(fault.New())
	task.SetSourcePath(srcPath)
	task.SetSourceStorage(srcStore)
	task.SetDestinationPath(dstPath)
	task.SetDestinationIndexSegmenter(dstSegmenter)
	task.SetSize(srcSize)
	task.SetOffset(0)
	task.SetSegment(dstSegment)
	task.SetIndex(1)

	srcReader.EXPECT().Close()
	srcStore.EXPECT().ReadWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, pairs ...*typ.Pair) (r io.ReadCloser, err error) {
			assert.Equal(t, srcPath, path)
			return srcReader, nil
		})
	dstSegmenter.EXPECT().WriteIndexSegmentWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, inputSeg segment.Segment, r io.Reader, index int, size int64, pair *typ.Pair) (err error) {
			assert.Equal(t, dstSegment, inputSeg)
			assert.Equal(t, 1, index)
			assert.Equal(t, srcSize, size)
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestSegmentStreamCopyTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	srcSize := int64(1024)
	dstSegment := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())
	dstPath := uuid.New().String()
	segmenter := mock.NewMockIndexSegmenter(ctrl)

	task := SegmentStreamCopyTask{}
	task.SetID(uuid.New().String())
	task.SetFault(fault.New())
	task.SetDestinationPath(dstPath)
	task.SetDestinationIndexSegmenter(segmenter)
	task.SetSize(srcSize)
	task.SetOffset(0)
	task.SetSegment(dstSegment)
	task.SetContent(&bytes.Buffer{})
	task.SetIndex(0)

	segmenter.EXPECT().WriteIndexSegmentWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, inputSeg segment.Segment, r io.Reader, index int, size int64, pair ...*typ.Pair) {
			assert.Equal(t, dstSegment, inputSeg)
			assert.Equal(t, 0, index)
			assert.Equal(t, srcSize, size)
			return
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestSegmentCompleteTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	dstSegment := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())
	segmenter := mock.NewMockIndexSegmenter(ctrl)

	task := SegmentCompleteTask{}
	task.SetFault(fault.New())
	task.SetIndexSegmenter(segmenter)
	task.SetSegment(dstSegment)

	segmenter.EXPECT().CompleteSegmentWithContext(gomock.Eq(context.Background()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, seg segment.Segment) (err error) {
			assert.Equal(t, dstSegment, seg)
			return nil
		})

	task.run(ctx)
	assert.Empty(t, task.GetFault().Error())
}

func TestInitSegmentStreamTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	srcPath := uuid.New().String()
	srcStore := mock.NewMockStorager(ctrl)
	srcReader := mock.NewMockReadCloser(ctrl)

	task := SegmentStreamInitTask{}
	task.SetPartSize(1024)
	task.SetSourceStorage(srcStore)
	task.SetSourcePath(srcPath)
	task.SetBytesPool(&sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	})

	srcStore.EXPECT().ReadWithContext(gomock.Eq(ctx), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, pairs ...*typ.Pair) (r io.ReadCloser, err error) {
			assert.Equal(t, srcPath, path)
			assert.Equal(t, int64(1024), pairs[0].Value.(int64))
			return srcReader, nil
		})
	srcReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (n int, err error) {
		return 768, io.EOF
	})

	task.run(ctx)

	assert.True(t, task.ValidateContent())
	assert.Equal(t, int64(768), task.GetSize())
	assert.Equal(t, true, task.GetDone())
}

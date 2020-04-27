package task

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/Xuanwo/storage/pkg/segment"
	typ "github.com/Xuanwo/storage/types"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/qingstor/noah/pkg/fault"
	"github.com/qingstor/noah/pkg/mock"
)

func TestSegmentInitTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("normal", func(t *testing.T) {
		store := mock.NewMockIndexSegmenter(ctrl)
		path := uuid.New().String()
		seg := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())

		task := SegmentInitTask{}
		task.SetIndexSegmenter(store)
		task.SetPath(path)
		task.SetPartSize(1000)

		store.EXPECT().InitIndexSegment(gomock.Any(), gomock.Any()).DoAndReturn(
			func(inputPath string, pairs ...*typ.Pair) (segment.Segment, error) {
				assert.Equal(t, inputPath, path)
				return seg, nil
			},
		)

		task.run()

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

		store.EXPECT().InitIndexSegment(gomock.Any(), gomock.Any()).DoAndReturn(
			func(inputPath string, pairs ...*typ.Pair) (string, error) {
				assert.Equal(t, inputPath, path)
				return "", errors.New("test")
			},
		)

		task.run()

		assert.False(t, task.ValidateSegment())
		assert.True(t, task.GetFault().HasError())
	})
}

func TestSegmentFileCopyTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	srcStore.EXPECT().Read(gomock.Any(), gomock.Any()).DoAndReturn(func(path string, pairs ...*typ.Pair) (r io.ReadCloser, err error) {
		assert.Equal(t, srcPath, path)
		return srcReader, nil
	})
	dstSegmenter.EXPECT().WriteIndexSegment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(inputSeg segment.Segment, r io.Reader, index int, size int64, pair *typ.Pair) (err error) {
			assert.Equal(t, dstSegment, inputSeg)
			assert.Equal(t, 1, index)
			assert.Equal(t, srcSize, size)
			return nil
		})

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

func TestSegmentStreamCopyTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	segmenter.EXPECT().WriteIndexSegment(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(inputSeg segment.Segment, r io.Reader, index int, size int64, pair ...*typ.Pair) {
			assert.Equal(t, dstSegment, inputSeg)
			assert.Equal(t, 0, index)
			assert.Equal(t, srcSize, size)
			return
		})

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

func TestSegmentCompleteTask_run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dstSegment := segment.NewIndexBasedSegment(uuid.New().String(), uuid.New().String())
	segmenter := mock.NewMockIndexSegmenter(ctrl)

	task := SegmentCompleteTask{}
	task.SetFault(fault.New())
	task.SetIndexSegmenter(segmenter)
	task.SetSegment(dstSegment)

	segmenter.EXPECT().CompleteSegment(gomock.Any()).DoAndReturn(func(seg segment.Segment) (err error) {
		assert.Equal(t, dstSegment, seg)
		return nil
	})

	task.run()
	assert.Empty(t, task.GetFault().Error())
}

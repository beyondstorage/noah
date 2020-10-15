package task

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/constants"
	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *CopyDirTask) new() {}
func (t *CopyDirTask) run(ctx context.Context) error {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		return err
	}

	if err := t.Sync(ctx, x); err != nil {
		return err
	}

	it := x.GetObjectIter()
	for {
		obj, err := it.Next()
		if err != nil {
			if errors.Is(err, typ.IterateDone) {
				break
			}
			return types.NewErrUnhandled(err)
		}

		switch obj.Type {
		case typ.ObjectTypeFile:
			sf := NewCopyFile(t)
			sf.SetSourcePath(obj.Name)
			sf.SetDestinationPath(obj.Name)
			if t.ValidateHandleObjCallbackFunc() {
				sf.SetCallbackFunc(func() {
					t.GetHandleObjCallbackFunc()(obj)
				})
			}
			t.Async(ctx, sf)
		case typ.ObjectTypeDir:
			sf := NewCopyDir(t)
			sf.SetSourcePath(obj.Name)
			sf.SetDestinationPath(obj.Name)
			t.Async(ctx, sf)
		default:
			return types.NewErrObjectTypeInvalid(nil, obj)
		}
	}

	return nil
}

func (t *CopyFileTask) new() {}
func (t *CopyFileTask) run(ctx context.Context) error {
	check := NewBetweenStorageCheck(t)
	err := t.Sync(ctx, check)
	if err != nil {
		return err
	}

	// Execute check tasks
	if t.ValidateCheckTasks() {
		for _, v := range t.GetCheckTasks() {
			ct := v(check)
			if err := t.Sync(ctx, ct); err != nil {
				return err
			}
			// If either check not pass, do not copy this file.
			if result := ct.(types.ResultGetter); !result.GetResult() {
				return nil
			}
			// If all check passed, we should continue do copy works.
		}
	}

	// if dry run func set, only dry run, not copy
	if t.ValidateDryRunFunc() {
		t.GetDryRunFunc()(check.GetSourceObject())
		return nil
	}

	srcSize, ok := check.GetSourceObject().GetSize()
	if !ok {
		return types.NewErrObjectMetaInvalid(nil, "size", check.GetSourceObject())
	}

	// if destination not support segmenter, we do not call multipart api
	_, ok = t.GetDestinationStorage().(typ.IndexSegmenter)
	if !ok {
		x := NewCopySmallFile(t)
		x.SetSize(srcSize)
		return t.Sync(ctx, x)
	}

	// if part threshold not set, use default value (1G) instead
	if !t.ValidatePartThreshold() {
		t.SetPartThreshold(constants.MaximumAutoMultipartSize)
	}

	if srcSize <= t.GetPartThreshold() {
		x := NewCopySmallFile(t)
		x.SetSize(srcSize)
		return t.Sync(ctx, x)
	}

	x := NewCopyLargeFile(t)
	x.SetTotalSize(srcSize)
	return t.Sync(ctx, x)
}

func (t *CopySmallFileTask) new() {}
func (t *CopySmallFileTask) run(ctx context.Context) error {
	fileCopyTask := NewCopySingleFile(t)

	if t.ValidateCheckMD5() && t.GetCheckMD5() {
		md5Task := NewMD5SumFile(t)
		utils.ChooseSourceStorage(md5Task, t)
		md5Task.SetOffset(0)
		if err := t.Sync(ctx, md5Task); err != nil {
			return err
		}
		fileCopyTask.SetMD5Sum(md5Task.GetMD5Sum())
	}

	return t.Sync(ctx, fileCopyTask)
}

func (t *CopyLargeFileTask) new() {}
func (t *CopyLargeFileTask) run(ctx context.Context) error {
	// if part size set, use it directly
	// TODO: we need to check validation of part size
	if !t.ValidatePartSize() {
		// otherwise, calculate part size.
		partSize, err := utils.CalculatePartSize(t.GetTotalSize())
		if err != nil {
			return types.NewErrUnhandled(err)
		}
		t.SetPartSize(partSize)
	}

	initTask := NewSegmentInit(t)
	err := utils.ChooseDestinationStorageAsIndexSegmenter(initTask, t)
	if err != nil {
		return err
	}

	if err := t.Sync(ctx, initTask); err != nil {
		return err
	}
	t.SetSegment(initTask.GetSegment())

	offset, part := int64(0), 0
	for {
		t.SetOffset(offset)

		x := NewCopyPartialFile(t)
		x.SetIndex(part)
		t.Async(ctx, x)
		// While GetDone is true, this must be the last part.
		if x.GetDone() {
			break
		}

		offset += x.GetSize()
		part++
	}

	// Make sure all segment upload finished.
	t.Await()
	// if meet error, not continue run complete task
	if t.GetFault().HasError() {
		return nil
	}
	return t.Sync(ctx, NewSegmentCompleteTask(initTask))
}

func (t *CopyPartialFileTask) new() {
	totalSize := t.GetTotalSize()
	partSize := t.GetPartSize()
	offset := t.GetOffset()

	if totalSize <= offset+partSize {
		t.SetSize(totalSize - offset)
		t.SetDone(true)
	} else {
		t.SetSize(partSize)
		t.SetDone(false)
	}
}
func (t *CopyPartialFileTask) run(ctx context.Context) error {
	fileCopyTask := NewSegmentFileCopy(t)

	if t.ValidateCheckMD5() && t.GetCheckMD5() {
		md5Task := NewMD5SumFile(t)
		utils.ChooseSourceStorage(md5Task, t)
		if err := t.Sync(ctx, md5Task); err != nil {
			return err
		}
		fileCopyTask.SetMD5Sum(md5Task.GetMD5Sum())
	}

	err := utils.ChooseDestinationStorageAsDestinationIndexSegmenter(fileCopyTask, t)
	if err != nil {
		return err
	}
	return t.Sync(ctx, fileCopyTask)
}

func (t *CopyStreamTask) new() {
	bytesPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, t.GetPartSize()))
		},
	}
	t.SetBytesPool(bytesPool)
}
func (t *CopyStreamTask) run(ctx context.Context) error {
	initTask := NewSegmentInit(t)
	err := utils.ChooseDestinationStorageAsIndexSegmenter(initTask, t)
	if err != nil {
		return err
	}

	// TODO: we will use expect size to calculate part size later.
	// if part size not set, use default value, otherwise load part size automatically
	if !t.ValidatePartSize() {
		t.SetPartSize(constants.DefaultPartSize)
	}

	if err := t.Sync(ctx, initTask); err != nil {
		return err
	}
	t.SetSegment(initTask.GetSegment())

	offset, part := int64(0), 0
	for {
		it := NewSegmentStreamInit(t)
		if err := t.Sync(ctx, it); err != nil {
			return err
		}

		x := NewCopyPartialStream(t)
		x.SetSize(it.GetSize())
		x.SetContent(it.GetContent())
		x.SetOffset(offset)
		x.SetIndex(part)
		t.Async(ctx, x)

		if it.GetDone() {
			break
		}

		offset += it.GetSize()
		part++
	}

	t.Await()
	// if meet error, not continue run complete task
	if t.GetFault().HasError() {
		return nil
	}
	return t.Sync(ctx, NewSegmentCompleteTask(initTask))
}

func (t *CopyPartialStreamTask) new() {}
func (t *CopyPartialStreamTask) run(ctx context.Context) error {
	copyTask := NewSegmentStreamCopy(t)
	if t.GetCheckMD5() {
		md5sumTask := NewMD5SumStream(t)
		if err := t.Sync(ctx, md5sumTask); err != nil {
			return err
		}
		copyTask.SetMD5Sum(md5sumTask.GetMD5Sum())
	}

	err := utils.ChooseDestinationStorageAsDestinationIndexSegmenter(copyTask, t)
	if err != nil {
		return err
	}

	return t.Sync(ctx, copyTask)
}

func (t *CopySingleFileTask) new() {}
func (t *CopySingleFileTask) run(ctx context.Context) error {
	r, w := io.Pipe()

	rst := NewReadFile(t)
	utils.ChooseSourceStorage(rst, t)
	rst.SetWriteCloser(w)
	t.Async(ctx, rst)

	wst := NewWriteFile(t)
	utils.ChooseDestinationStorage(wst, t)
	wst.SetReadCloser(r)

	// improve progress bar's performance, do not set state for small files less than 32M
	if t.GetSize() > constants.DefaultPartSize/4 {
		writeDone := 0
		progress.SetState(t.GetID(), progress.InitIncState(t.GetSourcePath(), "copy:", t.GetSize()))
		wst.SetReadCallBackFunc(func(b []byte) {
			writeDone += len(b)
			progress.UpdateState(t.GetID(), int64(writeDone))
		})
	}

	t.Async(ctx, wst)
	return nil
}

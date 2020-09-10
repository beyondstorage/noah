package task

import (
	"bytes"
	"context"
	"sync"

	"github.com/aos-dev/go-storage/v2"
	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/aos-dev/go-storage/v2/types/pairs"

	"github.com/qingstor/noah/constants"
	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *CopyDirTask) new() {}
func (t *CopyDirTask) run(ctx context.Context) {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	x.SetFileFunc(func(o *typ.Object) {
		sf := NewCopyFile(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		if t.ValidateHandleObjCallback() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallback()(o)
			})
		}
		if t.ValidatePartSize() {
			sf.SetPartSize(t.GetPartSize())
		}
		t.GetScheduler().Async(ctx, sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewCopyDir(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		if t.ValidateHandleObjCallback() {
			sf.SetHandleObjCallback(t.GetHandleObjCallback())
		}
		if t.ValidatePartSize() {
			sf.SetPartSize(t.GetPartSize())
		}
		t.GetScheduler().Sync(ctx, sf)
	})
	t.GetScheduler().Sync(ctx, x)
}

func (t *CopyFileTask) new() {}
func (t *CopyFileTask) run(ctx context.Context) {
	check := NewBetweenStorageCheck(t)
	t.GetScheduler().Sync(ctx, check)
	if t.GetFault().HasError() {
		return
	}

	// Execute check tasks
	for _, v := range t.GetCheckTasks() {
		ct := v(check)
		t.GetScheduler().Sync(ctx, ct)
		if t.GetFault().HasError() {
			return
		}
		// If either check not pass, do not copy this file.
		if result := ct.(types.ResultGetter); !result.GetResult() {
			return
		}
		// If all check passed, we should continue do copy works.
	}

	srcSize := check.GetSourceObject().Size

	// if destination not support segmenter, we do not call multipart api
	_, ok := t.GetDestinationStorage().(storage.IndexSegmenter)
	if !ok {
		x := NewCopySmallFile(t)
		x.SetSize(srcSize)
		t.GetScheduler().Sync(ctx, x)
		return
	}

	if srcSize <= t.GetPartThreshold() {
		x := NewCopySmallFile(t)
		x.SetSize(srcSize)
		t.GetScheduler().Sync(ctx, x)
	} else {
		x := NewCopyLargeFile(t)
		x.SetTotalSize(srcSize)
		if t.ValidatePartSize() {
			x.SetPartSize(t.GetPartSize())
		}
		t.GetScheduler().Sync(ctx, x)
	}
}

func (t *CopySmallFileTask) new() {}
func (t *CopySmallFileTask) run(ctx context.Context) {
	fileCopyTask := NewCopySingleFile(t)

	if t.GetCheckMD5() {
		md5Task := NewMD5SumFile(t)
		utils.ChooseSourceStorage(md5Task, t)
		md5Task.SetOffset(0)
		t.GetScheduler().Sync(ctx, md5Task)
		if t.GetFault().HasError() {
			return
		}
		fileCopyTask.SetMD5Sum(md5Task.GetMD5Sum())
	} else {
		fileCopyTask.SetMD5Sum(nil)
	}

	t.GetScheduler().Sync(ctx, fileCopyTask)
}

func (t *CopyLargeFileTask) new() {}
func (t *CopyLargeFileTask) run(ctx context.Context) {
	// if part size set, use it directly
	// TODO: we need to check validation of part size
	if t.ValidatePartSize() {
		t.SetPartSize(t.GetPartSize())
	} else {
		// otherwise, calculate part size.
		partSize, err := utils.CalculatePartSize(t.GetTotalSize())
		if err != nil {
			t.TriggerFault(types.NewErrUnhandled(err))
			return
		}
		t.SetPartSize(partSize)
	}

	initTask := NewSegmentInit(t)
	err := utils.ChooseDestinationStorageAsIndexSegmenter(initTask, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	t.GetScheduler().Sync(ctx, initTask)
	if t.GetFault().HasError() {
		return
	}
	t.SetSegment(initTask.GetSegment())

	offset, part := int64(0), 0
	for {
		t.SetOffset(offset)

		x := NewCopyPartialFile(t)
		x.SetIndex(part)
		t.GetScheduler().Async(ctx, x)
		// While GetDone is true, this must be the last part.
		if x.GetDone() {
			break
		}

		offset += x.GetSize()
		part++
	}

	// Make sure all segment upload finished.
	t.GetScheduler().Wait()
	if t.GetFault().HasError() {
		return
	}
	t.GetScheduler().Sync(ctx, NewSegmentCompleteTask(initTask))
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
func (t *CopyPartialFileTask) run(ctx context.Context) {
	fileCopyTask := NewSegmentFileCopy(t)

	if t.GetCheckMD5() {
		md5Task := NewMD5SumFile(t)
		utils.ChooseSourceStorage(md5Task, t)
		t.GetScheduler().Sync(ctx, md5Task)
		if t.GetFault().HasError() {
			return
		}
		fileCopyTask.SetMD5Sum(md5Task.GetMD5Sum())
	} else {
		fileCopyTask.SetMD5Sum(nil)
	}

	err := utils.ChooseDestinationIndexSegmenter(fileCopyTask, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}
	t.GetScheduler().Sync(ctx, fileCopyTask)
}

func (t *CopyStreamTask) new() {
	bytesPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, t.GetPartSize()))
		},
	}
	t.SetBytesPool(bytesPool)
}
func (t *CopyStreamTask) run(ctx context.Context) {
	initTask := NewSegmentInit(t)
	err := utils.ChooseDestinationStorageAsIndexSegmenter(initTask, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	// TODO: we will use expect size to calculate part size later.
	// if part size was set, use it directly
	if !t.ValidatePartSize() {
		t.SetPartSize(constants.DefaultPartSize)
	}

	t.GetScheduler().Sync(ctx, initTask)
	if t.GetFault().HasError() {
		return
	}
	t.SetSegment(initTask.GetSegment())

	offset, part := int64(0), 0
	for {
		it := NewInitSegmentStream(t)
		t.GetScheduler().Sync(ctx, it)
		if t.GetFault().HasError() {
			return
		}

		x := NewCopyPartialStream(t)
		x.SetSize(it.GetSize())
		x.SetContent(it.GetContent())
		x.SetOffset(offset)
		x.SetIndex(part)
		t.GetScheduler().Async(ctx, x)

		if it.GetDone() {
			break
		}

		offset += it.GetSize()
		part++
	}

	t.GetScheduler().Wait()
	if t.GetFault().HasError() {
		return
	}
	t.GetScheduler().Sync(ctx, NewSegmentCompleteTask(initTask))
}

func (t *CopyPartialStreamTask) new() {}
func (t *CopyPartialStreamTask) run(ctx context.Context) {
	copyTask := NewSegmentStreamCopy(t)
	if t.GetCheckMD5() {
		md5sumTask := NewMD5SumStream(t)
		t.GetScheduler().Sync(ctx, md5sumTask)
		if t.GetFault().HasError() {
			return
		}
		copyTask.SetMD5Sum(md5sumTask.GetMD5Sum())
	} else {
		copyTask.SetMD5Sum(nil)
	}

	err := utils.ChooseDestinationIndexSegmenter(copyTask, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	t.GetScheduler().Sync(ctx, copyTask)
}

func (t *CopySingleFileTask) new() {}
func (t *CopySingleFileTask) run(ctx context.Context) {
	r, err := t.GetSourceStorage().ReadWithContext(ctx, t.GetSourcePath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	defer r.Close()

	// improve progress bar's performance, do not set state for small files less than 32M
	if t.GetSize() > constants.DefaultPartSize/4 {
		progress.SetState(t.GetID(), progress.InitIncState(t.GetSourcePath(), "copy:", t.GetSize()))
	}
	// TODO: add checksum support
	writeDone := 0
	err = t.GetDestinationStorage().WriteWithContext(ctx, t.GetDestinationPath(), r, pairs.WithSize(t.GetSize()),
		pairs.WithReadCallbackFunc(func(b []byte) {
			writeDone += len(b)
			progress.UpdateState(t.GetID(), int64(writeDone))
		}))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

package task

import (
	"bytes"
	"io"
	"sync"

	typ "github.com/Xuanwo/storage/types"
	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/constants"
	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *CopyDirTask) new() {}
func (t *CopyDirTask) run() {
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
		t.GetScheduler().Async(sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewCopyDir(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		t.GetScheduler().Sync(sf)
	})
	t.GetScheduler().Sync(x)
}

func (t *CopyFileTask) new() {}
func (t *CopyFileTask) run() {
	check := NewBetweenStorageCheck(t)
	t.GetScheduler().Sync(check)
	if t.GetFault().HasError() {
		return
	}

	// Execute check tasks
	for _, v := range t.GetCheckTasks() {
		ct := v(check)
		t.GetScheduler().Sync(ct)
		// If either check not pass, do not copy this file.
		if result := ct.(types.ResultGetter); !result.GetResult() {
			return
		}
		// If all check passed, we should continue do copy works.
	}

	srcSize := check.GetSourceObject().Size
	if srcSize >= constants.MaximumAutoMultipartSize {
		x := NewCopyLargeFile(t)
		x.SetTotalSize(srcSize)
		t.GetScheduler().Sync(x)
	} else {
		x := NewCopySmallFile(t)
		x.SetSize(srcSize)
		t.GetScheduler().Sync(x)
	}
}

func (t *CopySmallFileTask) new() {}
func (t *CopySmallFileTask) run() {
	md5Task := NewMD5SumFile(t)
	utils.ChooseSourceStorage(md5Task, t)
	md5Task.SetOffset(0)
	t.GetScheduler().Sync(md5Task)

	fileCopyTask := NewCopySingleFile(t)
	fileCopyTask.SetMD5Sum(md5Task.GetMD5Sum())
	t.GetScheduler().Sync(fileCopyTask)
}

func (t *CopyLargeFileTask) new() {}
func (t *CopyLargeFileTask) run() {
	// Set segment part size.
	partSize, err := utils.CalculatePartSize(t.GetTotalSize())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetPartSize(partSize)

	initTask := NewSegmentInit(t)
	err = utils.ChooseDestinationStorageAsIndexSegmenter(initTask, t)
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}

	t.GetScheduler().Sync(initTask)
	t.SetSegment(initTask.GetSegment())

	offset, part, doneCount := int64(0), 0, int64(0)
	for {
		t.SetOffset(offset)

		x := NewCopyPartialFile(t)
		x.SetCallbackFunc(func() {
			doneCount++
			progress.UpdateState(t.GetID(), doneCount)
		})
		x.SetIndex(part)
		t.GetScheduler().Async(x)
		// While GetDone is true, this must be the last part.
		if x.GetDone() {
			break
		}

		offset += x.GetSize()
		part++
	}
	progress.SetState(t.GetID(), progress.InitIncState(t.GetSourcePath(), "copy part:", int64(part)))

	// Make sure all segment upload finished.
	t.GetScheduler().Wait()
	t.GetScheduler().Sync(NewSegmentCompleteTask(initTask))
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
func (t *CopyPartialFileTask) run() {
	md5Task := NewMD5SumFile(t)
	utils.ChooseSourceStorage(md5Task, t)
	t.GetScheduler().Sync(md5Task)

	fileCopyTask := NewSegmentFileCopy(t)
	fileCopyTask.SetMD5Sum(md5Task.GetMD5Sum())
	err := utils.ChooseDestinationIndexSegmenter(fileCopyTask, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}
	t.GetScheduler().Sync(fileCopyTask)
}

func (t *CopyStreamTask) new() {
	bytesPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, t.GetPartSize()))
		},
	}
	t.SetBytesPool(bytesPool)
}
func (t *CopyStreamTask) run() {
	initTask := NewSegmentInit(t)
	err := utils.ChooseDestinationStorageAsIndexSegmenter(initTask, t)
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}

	// TODO: we will use expect size to calculate part size later.
	partSize := int64(constants.DefaultPartSize)
	t.SetPartSize(partSize)

	t.GetScheduler().Sync(initTask)
	t.SetSegment(initTask.GetSegment())

	offset, part, doneCount := int64(0), 0, int64(0)
	for {
		x := NewCopyPartialStream(t)
		x.SetOffset(offset)
		x.SetCallbackFunc(func() {
			doneCount++
			progress.UpdateState(t.GetID(), doneCount)
		})
		x.SetIndex(part)
		t.GetScheduler().Async(x)

		if x.GetDone() {
			break
		}

		offset += x.GetSize()
		part++
	}
	progress.SetState(t.GetID(), progress.InitIncState(t.GetSourcePath(), "copy stream part:", int64(part)))

	t.GetScheduler().Wait()
	t.GetScheduler().Sync(NewSegmentCompleteTask(initTask))
}

func (t *CopyPartialStreamTask) new() {
	// Set size and update offset.
	partSize := t.GetPartSize()

	r, err := t.GetSourceStorage().Read(t.GetSourcePath(), pairs.WithSize(partSize))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}

	b := t.GetBytesPool().Get().(*bytes.Buffer)
	n, err := io.Copy(b, r)
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}

	t.SetSize(n)
	t.SetContent(b)
	if n < partSize {
		t.SetDone(true)
	} else {
		t.SetDone(false)
	}
}
func (t *CopyPartialStreamTask) run() {
	md5sumTask := NewMD5SumStream(t)
	t.GetScheduler().Sync(md5sumTask)

	copyTask := NewSegmentStreamCopy(t)
	err := utils.ChooseDestinationIndexSegmenter(copyTask, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}
	copyTask.SetMD5Sum(md5sumTask.GetMD5Sum())
	t.GetScheduler().Sync(copyTask)
}

func (t *CopySingleFileTask) new() {}
func (t *CopySingleFileTask) run() {
	r, err := t.GetSourceStorage().Read(t.GetSourcePath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	defer r.Close()

	progress.SetState(t.GetID(), progress.InitIncState(t.GetSourcePath(), "copy:", t.GetSize()))
	// TODO: add checksum support
	writeDone := 0
	err = t.GetDestinationStorage().Write(t.GetDestinationPath(), r, pairs.WithSize(t.GetSize()),
		pairs.WithReadCallbackFunc(func(b []byte) {
			writeDone += len(b)
			progress.UpdateState(t.GetID(), int64(writeDone))
		}))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

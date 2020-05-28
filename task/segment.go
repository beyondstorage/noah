package task

import (
	"fmt"
	"io/ioutil"

	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
)

func (t *SegmentInitTask) new() {}
func (t *SegmentInitTask) run() {
	seg, err := t.GetIndexSegmenter().InitIndexSegment(t.GetPath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetSegment(seg)
}

func (t *SegmentFileCopyTask) new() {}
func (t *SegmentFileCopyTask) run() {
	r, err := t.GetSourceStorage().Read(t.GetSourcePath(), pairs.WithSize(t.GetSize()), pairs.WithOffset(t.GetOffset()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	defer r.Close()

	progress.SetState(t.GetID(), progress.InitIncState(t.GetDestinationPath(), fmt.Sprintf("copy file part: %d", t.GetIndex()), t.GetSize()))
	// TODO: Add checksum support.
	writeDone := 0
	seg := t.GetSegment()
	err = t.GetDestinationIndexSegmenter().WriteIndexSegment(seg, r, t.GetIndex(), t.GetSize(),
		pairs.WithReadCallbackFunc(func(b []byte) {
			writeDone += len(b)
			progress.UpdateState(t.GetID(), int64(writeDone))
		}),
	)
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *SegmentStreamCopyTask) new() {}
func (t *SegmentStreamCopyTask) run() {
	progress.SetState(t.GetID(), progress.InitIncState(t.GetDestinationPath(), fmt.Sprintf("copy stream part: %d", t.GetIndex()), t.GetSize()))
	// TODO: Add checksum support
	writeDone := 0
	err := t.GetDestinationIndexSegmenter().WriteIndexSegment(t.GetSegment(), ioutil.NopCloser(t.GetContent()),
		t.GetIndex(), t.GetSize(),
		pairs.WithReadCallbackFunc(func(b []byte) {
			writeDone += len(b)
			progress.UpdateState(t.GetID(), int64(writeDone))
		}))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *SegmentCompleteTask) new() {}
func (t *SegmentCompleteTask) run() {
	err := t.GetIndexSegmenter().CompleteSegment(t.GetSegment())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aos-dev/go-storage/v2/types/pairs"

	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
)

func (t *SegmentInitTask) new() {}
func (t *SegmentInitTask) run(ctx context.Context) {
	seg, err := t.GetIndexSegmenter().InitIndexSegmentWithContext(ctx, t.GetPath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetSegment(seg)
}

func (t *SegmentFileCopyTask) new() {}
func (t *SegmentFileCopyTask) run(ctx context.Context) {
	r, err := t.GetSourceStorage().ReadWithContext(ctx, t.GetSourcePath(), pairs.WithSize(t.GetSize()), pairs.WithOffset(t.GetOffset()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	defer r.Close()

	progress.SetState(t.GetID(), progress.InitIncState(t.GetDestinationPath(), fmt.Sprintf("copy file part: %d", t.GetIndex()), t.GetSize()))
	// TODO: Add checksum support.
	writeDone := 0
	seg := t.GetSegment()
	err = t.GetDestinationIndexSegmenter().WriteIndexSegmentWithContext(ctx, seg, r, t.GetIndex(), t.GetSize(),
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

func (t *InitSegmentStreamTask) new() {}
func (t *InitSegmentStreamTask) run(ctx context.Context) {
	// Set size and update offset.
	partSize := t.GetPartSize()

	r, err := t.GetSourceStorage().ReadWithContext(ctx, t.GetSourcePath(), pairs.WithSize(partSize))
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

func (t *SegmentStreamCopyTask) new() {}
func (t *SegmentStreamCopyTask) run(ctx context.Context) {
	progress.SetState(t.GetID(), progress.InitIncState(t.GetDestinationPath(), fmt.Sprintf("copy stream part: %d", t.GetIndex()), t.GetSize()))
	// TODO: Add checksum support
	writeDone := 0
	err := t.GetDestinationIndexSegmenter().WriteIndexSegmentWithContext(ctx, t.GetSegment(), ioutil.NopCloser(t.GetContent()),
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
func (t *SegmentCompleteTask) run(ctx context.Context) {
	err := t.GetIndexSegmenter().CompleteSegmentWithContext(ctx, t.GetSegment())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

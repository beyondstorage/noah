package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aos-dev/go-storage/v2/pairs"

	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *SegmentInitTask) new() {}
func (t *SegmentInitTask) run(ctx context.Context) error {
	seg, err := t.GetIndexSegmenter().InitIndexSegmentWithContext(ctx, t.GetPath())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetSegment(seg)
	return nil
}

func (t *SegmentFileCopyTask) new() {}
func (t *SegmentFileCopyTask) run(ctx context.Context) error {
	r, w := io.Pipe()

	rst := NewReadFile(t)
	utils.ChooseSourceStorage(rst, t)
	rst.SetWriteCloser(w)

	t.Async(ctx, rst)

	wst := NewWriteSegment(t)
	utils.ChooseDestinationIndexSegmenter(wst, t)
	wst.SetReadCloser(r)

	progress.SetState(t.GetID(), progress.InitIncState(t.GetDestinationPath(), fmt.Sprintf("copy file part: %d", t.GetIndex()), t.GetSize()))
	writeDone := 0
	wst.SetReadCallBackFunc(func(b []byte) {
		writeDone += len(b)
		progress.UpdateState(t.GetID(), int64(writeDone))
	})

	t.Async(ctx, wst)
	return nil
}

func (t *SegmentStreamInitTask) new() {}
func (t *SegmentStreamInitTask) run(ctx context.Context) error {
	// Set size and update offset.
	partSize := t.GetPartSize()
	b := t.GetBytesPool().Get().(*bytes.Buffer)

	n, err := t.GetSourceStorage().ReadWithContext(ctx, t.GetSourcePath(), b, pairs.WithSize(partSize))
	if err != nil {
		return types.NewErrUnhandled(err)
	}

	t.SetSize(n)
	t.SetContent(b)
	if n < partSize {
		t.SetDone(true)
	} else {
		t.SetDone(false)
	}
	return nil
}

func (t *SegmentStreamCopyTask) new() {}
func (t *SegmentStreamCopyTask) run(ctx context.Context) error {
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
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *SegmentCompleteTask) new() {}
func (t *SegmentCompleteTask) run(ctx context.Context) error {
	err := t.GetIndexSegmenter().CompleteSegmentWithContext(ctx, t.GetSegment())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

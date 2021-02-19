package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2/pairs"
	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/aos-dev/noah/pkg/types"
)

func (t *ReadFileTask) new() {}
func (t *ReadFileTask) run(ctx context.Context) error {
	ps := make([]typ.Pair, 0)
	if t.ValidateSize() {
		ps = append(ps, pairs.WithSize(t.GetSize()))
	}
	if t.ValidateOffset() {
		ps = append(ps, pairs.WithOffset(t.GetOffset()))
	}
	if t.ValidateReadCallBackFunc() {
		ps = append(ps, pairs.WithReadCallbackFunc(t.GetReadCallBackFunc()))
	}

	if _, err := t.GetStorage().ReadWithContext(ctx, t.GetPath(), t.GetWriteCloser(), ps...); err != nil {
		t.GetWriteCloser().Close()
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *WriteFileTask) new() {}
func (t *WriteFileTask) run(ctx context.Context) error {
	ps := make([]typ.Pair, 0)
	if t.ValidateSize() {
		ps = append(ps, pairs.WithSize(t.GetSize()))
	}
	if t.ValidateOffset() {
		ps = append(ps, pairs.WithOffset(t.GetOffset()))
	}
	if t.ValidateReadCallBackFunc() {
		ps = append(ps, pairs.WithReadCallbackFunc(t.GetReadCallBackFunc()))
	}
	if t.ValidateStorageClass() {
		ps = append(ps, pairs.WithStorageClass(t.GetStorageClass()))
	}

	if _, err := t.GetStorage().WriteWithContext(ctx, t.GetPath(), t.GetReadCloser(), ps...); err != nil {
		t.GetReadCloser().Close()
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *WriteSegmentTask) new() {}
func (t *WriteSegmentTask) run(ctx context.Context) error {
	ps := make([]typ.Pair, 0)
	if t.ValidateReadCallBackFunc() {
		ps = append(ps, pairs.WithReadCallbackFunc(t.GetReadCallBackFunc()))
	}

	if err := t.GetIndexSegmenter().WriteIndexSegmentWithContext(ctx, t.GetSegment(), t.GetReadCloser(),
		t.GetIndex(), t.GetSize(), ps...); err != nil {
		t.GetReadCloser().Close()
		return types.NewErrUnhandled(err)
	}
	return nil
}

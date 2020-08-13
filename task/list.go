package task

import (
	"context"

	"github.com/Xuanwo/storage"
	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *ListDirTask) new() {}
func (t *ListDirTask) run(ctx context.Context) {
	err := t.GetDirLister().ListDirWithContext(
		ctx, t.GetPath(), pairs.WithDirFunc(t.GetDirFunc()), pairs.WithFileFunc(t.GetFileFunc()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *ListPrefixTask) new() {}
func (t *ListPrefixTask) run(ctx context.Context) {
	err := t.GetPrefixLister().ListPrefixWithContext(
		ctx, t.GetPath(), pairs.WithObjectFunc(t.GetObjectFunc()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *ListSegmentTask) new() {}
func (t *ListSegmentTask) run(ctx context.Context) {
	err := t.GetPrefixSegmentsLister().ListPrefixSegmentsWithContext(
		ctx, t.GetPath(), pairs.WithSegmentFunc(t.GetSegmentFunc()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *ListStorageTask) new() {}
func (t *ListStorageTask) run(ctx context.Context) {
	err := t.GetService().ListWithContext(
		ctx, pairs.WithLocation(t.GetZone()), pairs.WithStoragerFunc(func(storager storage.Storager) {
			t.GetStoragerFunc()(storager)
		}))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

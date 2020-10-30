package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2"
	"github.com/aos-dev/go-storage/v2/types/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *ListDirTask) new() {}
func (t *ListDirTask) run(ctx context.Context) error {
	err := t.GetDirLister().ListDirWithContext(
		ctx, t.GetPath(), pairs.WithDirFunc(t.GetDirFunc()), pairs.WithFileFunc(t.GetFileFunc()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *ListPrefixTask) new() {}
func (t *ListPrefixTask) run(ctx context.Context) error {
	err := t.GetPrefixLister().ListPrefixWithContext(
		ctx, t.GetPath(), pairs.WithObjectFunc(t.GetObjectFunc()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *ListSegmentTask) new() {}
func (t *ListSegmentTask) run(ctx context.Context) error {
	err := t.GetPrefixSegmentsLister().ListPrefixSegmentsWithContext(
		ctx, t.GetPath(), pairs.WithSegmentFunc(t.GetSegmentFunc()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *ListStorageTask) new() {}
func (t *ListStorageTask) run(ctx context.Context) error {
	err := t.GetService().ListWithContext(
		ctx, pairs.WithLocation(t.GetZone()), pairs.WithStoragerFunc(func(storager storage.Storager) {
			t.GetStoragerFunc()(storager)
		}))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

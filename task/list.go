package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2/pairs"
	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/pkg/types"
)

func (t *ListDirTask) new() {}
func (t *ListDirTask) run(ctx context.Context) error {
	it, err := t.GetDirLister().ListDirWithContext(
		ctx, t.GetPath())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetObjectIter(it)
	return nil
}

func (t *ListPrefixTask) new() {}
func (t *ListPrefixTask) run(ctx context.Context) error {
	it, err := t.GetPrefixLister().ListPrefixWithContext(
		ctx, t.GetPath())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetObjectIter(it)
	return nil
}

func (t *ListSegmentTask) new() {}
func (t *ListSegmentTask) run(ctx context.Context) error {
	it, err := t.GetPrefixSegmentsLister().ListPrefixSegmentsWithContext(
		ctx, t.GetPath())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetSegmentIter(it)
	return nil
}

func (t *ListStorageTask) new() {}
func (t *ListStorageTask) run(ctx context.Context) error {
	ps := make([]typ.Pair, 0)
	if t.ValidateZone() {
		ps = append(ps, pairs.WithLocation(t.GetZone()))
	}

	it, err := t.GetService().ListWithContext(
		ctx, ps...)
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetStorageIter(it)
	return nil
}

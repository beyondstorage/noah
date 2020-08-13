package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2"
	"github.com/aos-dev/go-storage/v2/pkg/segment"
	typ "github.com/aos-dev/go-storage/v2/types"
	"github.com/aos-dev/go-storage/v2/types/pairs"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *DeleteFileTask) new() {}
func (t *DeleteFileTask) run(ctx context.Context) {
	if err := t.GetStorage().DeleteWithContext(ctx, t.GetPath()); err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *DeleteDirTask) new() {}
func (t *DeleteDirTask) run(ctx context.Context) {
	x := NewListDir(t)
	err := utils.ChooseStorageAsDirLister(x, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	x.SetFileFunc(func(o *typ.Object) {
		sf := NewDeleteFile(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallback() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallback()(o)
			})
		}
		t.GetScheduler().Async(ctx, sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewDeleteDir(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallback() {
			sf.SetHandleObjCallback(t.GetHandleObjCallback())
		}
		t.GetScheduler().Sync(ctx, sf)
	})
	t.GetScheduler().Sync(ctx, x)
	if t.GetFault().HasError() {
		return
	}

	// after delete all files in this dir, delete dir itself as a file, see issue #43
	dr := NewDeleteFile(t)
	if t.ValidateHandleObjCallback() {
		dr.SetHandleObjCallback(t.GetHandleObjCallback())
	}
	t.GetScheduler().Sync(ctx, dr)
}

func (t *DeletePrefixTask) new() {}
func (t *DeletePrefixTask) run(ctx context.Context) {
	x := NewListPrefix(t)
	err := utils.ChooseStorageAsPrefixLister(x, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	x.SetObjectFunc(func(o *typ.Object) {
		sf := NewDeleteFile(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallback() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallback()(o)
			})
		}
		t.GetScheduler().Async(ctx, sf)
	})

	t.GetScheduler().Sync(ctx, x)
}

func (t *DeleteSegmentTask) new() {}
func (t *DeleteSegmentTask) run(ctx context.Context) {
	if err := t.GetPrefixSegmentsLister().AbortSegmentWithContext(ctx, t.GetSegment()); err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *DeleteStorageTask) new() {}
func (t *DeleteStorageTask) run(ctx context.Context) {
	var ps []*typ.Pair
	if t.GetZone() != "" {
		ps = append(ps, pairs.WithLocation(t.GetZone()))
	}
	if t.GetForce() {
		store, err := t.GetService().GetWithContext(ctx, t.GetStorageName(), ps...)
		if err != nil {
			t.TriggerFault(types.NewErrUnhandled(err))
			return
		}

		deletePrefix := NewDeletePrefix(t)
		deletePrefix.SetPath("")
		deletePrefix.SetStorage(store)
		if t.ValidateHandleObjCallback() {
			deletePrefix.SetHandleObjCallback(t.GetHandleObjCallback())
		}

		t.GetScheduler().Async(ctx, deletePrefix)

		segmenter, ok := store.(storage.PrefixSegmentsLister)
		if ok {
			listSegments := NewListSegment(t)
			listSegments.SetPrefixSegmentsLister(segmenter)
			listSegments.SetPath("")
			listSegments.SetSegmentFunc(func(s segment.Segment) {
				sf := NewDeleteSegment(t)
				sf.SetPrefixSegmentsLister(segmenter)
				sf.SetSegment(s)
				if t.ValidateHandleSegmentCallback() {
					sf.SetCallbackFunc(func() {
						t.GetHandleSegmentCallback()(s)
					})
				}
				t.GetScheduler().Async(ctx, sf)
			})

			t.GetScheduler().Async(ctx, listSegments)
		}

		t.GetScheduler().Wait()
		if t.GetFault().HasError() {
			return
		}
	}

	if err := t.GetService().DeleteWithContext(ctx, t.GetStorageName(), ps...); err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

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
func (t *DeleteFileTask) run(ctx context.Context) error {
	if err := t.GetStorage().DeleteWithContext(ctx, t.GetPath()); err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *DeleteDirTask) new() {}
func (t *DeleteDirTask) run(ctx context.Context) error {
	x := NewListDir(t)
	err := utils.ChooseStorageAsDirLister(x, t)
	if err != nil {
		return err
	}

	x.SetFileFunc(func(o *typ.Object) {
		sf := NewDeleteFile(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallbackFunc() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallbackFunc()(o)
			})
		}
		t.Async(ctx, sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewDeleteDir(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallbackFunc() {
			sf.SetHandleObjCallbackFunc(t.GetHandleObjCallbackFunc())
		}
		t.Sync(ctx, sf)
	})
	if err := t.Sync(ctx, x); err != nil {
		return err
	}

	// after delete all files in this dir, delete dir itself as a file, see issue #43
	dr := NewDeleteFile(t)
	if t.ValidateHandleObjCallbackFunc() {
		dr.SetHandleObjCallbackFunc(t.GetHandleObjCallbackFunc())
	}
	return t.Sync(ctx, dr)
}

func (t *DeletePrefixTask) new() {}
func (t *DeletePrefixTask) run(ctx context.Context) error {
	x := NewListPrefix(t)
	err := utils.ChooseStorageAsPrefixLister(x, t)
	if err != nil {
		return err
	}

	x.SetObjectFunc(func(o *typ.Object) {
		sf := NewDeleteFile(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallbackFunc() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallbackFunc()(o)
			})
		}
		t.Async(ctx, sf)
	})

	return t.Sync(ctx, x)
}

func (t *DeleteSegmentTask) new() {}
func (t *DeleteSegmentTask) run(ctx context.Context) error {
	if err := t.GetPrefixSegmentsLister().AbortSegmentWithContext(ctx, t.GetSegment()); err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *DeleteStorageTask) new() {}
func (t *DeleteStorageTask) run(ctx context.Context) error {
	var ps []*typ.Pair
	if t.ValidateZone() {
		ps = append(ps, pairs.WithLocation(t.GetZone()))
	}
	if t.ValidateForce() && t.GetForce() {
		store, err := t.GetService().GetWithContext(ctx, t.GetStorageName(), ps...)
		if err != nil {
			return types.NewErrUnhandled(err)
		}

		deletePrefix := NewDeletePrefix(t)
		deletePrefix.SetPath("")
		deletePrefix.SetStorage(store)
		if t.ValidateHandleObjCallbackFunc() {
			deletePrefix.SetHandleObjCallbackFunc(t.GetHandleObjCallbackFunc())
		}

		t.Async(ctx, deletePrefix)

		segmenter, ok := store.(storage.PrefixSegmentsLister)
		if ok {
			listSegments := NewListSegment(t)
			listSegments.SetPrefixSegmentsLister(segmenter)
			listSegments.SetPath("")
			listSegments.SetSegmentFunc(func(s segment.Segment) {
				sf := NewDeleteSegment(t)
				sf.SetPrefixSegmentsLister(segmenter)
				sf.SetSegment(s)
				if t.ValidateHandleSegmentCallbackFunc() {
					sf.SetCallbackFunc(func() {
						t.GetHandleSegmentCallbackFunc()(s)
					})
				}
				t.Async(ctx, sf)
			})

			t.Async(ctx, listSegments)
		}

		if err := t.Await(); err != nil {
			return err
		}
	}

	if err := t.GetService().DeleteWithContext(ctx, t.GetStorageName(), ps...); err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

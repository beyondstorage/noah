// +build refactor

package task

import (
	"context"
	"errors"

	"github.com/aos-dev/go-storage/v2/pairs"
	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/aos-dev/noah/pkg/types"
	"github.com/aos-dev/noah/utils"
)

func (t *DeleteFileTask) new() {}
func (t *DeleteFileTask) run(ctx context.Context) error {
	if err := t.GetStorage().DeleteWithContext(ctx, t.GetPath()); err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

func (t *DeleteSegmentTask) new() {}
func (t *DeleteSegmentTask) run(ctx context.Context) error {
	if err := t.GetPrefixSegmentsLister().AbortSegmentWithContext(ctx, t.GetSegment()); err != nil {
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

	if err := t.Sync(ctx, x); err != nil {
		return err
	}

	it := x.GetObjectIter()
	for {
		obj, err := it.Next()
		if err != nil {
			if errors.Is(err, typ.IterateDone) {
				break
			}
			return types.NewErrUnhandled(err)
		}
		switch obj.Type {
		case typ.ObjectTypeFile:
			sf := NewDeleteFile(t)
			sf.SetPath(obj.Name)
			if t.ValidateHandleObjCallbackFunc() {
				sf.SetCallbackFunc(func() {
					t.GetHandleObjCallbackFunc()(obj)
				})
			}
			t.Async(ctx, sf)
		case typ.ObjectTypeDir:
			sf := NewDeleteDir(t)
			sf.SetPath(obj.Name)
			if err := t.Sync(ctx, sf); err != nil {
				return err
			}
		default:
			return types.NewErrObjectTypeInvalid(nil, obj)
		}
	}

	// Make sure all objects in current dir deleted.
	// if meet error, not continue run delete itself
	if err := t.Await(); err != nil {
		return err
	}
	// after delete all files in this dir, delete dir itself as a file, see issue #43
	dr := NewDeleteFile(t)
	return t.Sync(ctx, dr)
}

func (t *DeletePrefixTask) new() {}
func (t *DeletePrefixTask) run(ctx context.Context) error {
	x := NewListPrefix(t)
	err := utils.ChooseStorageAsPrefixLister(x, t)
	if err != nil {
		return err
	}

	if err := t.Sync(ctx, x); err != nil {
		return err
	}

	it := x.GetObjectIter()
	for {
		obj, err := it.Next()
		if err != nil {
			if errors.Is(err, typ.IterateDone) {
				break
			}
			return types.NewErrUnhandled(err)
		}

		sf := NewDeleteFile(t)
		sf.SetPath(obj.Name)
		if t.ValidateHandleObjCallbackFunc() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallbackFunc()(obj)
			})
		}
		t.Async(ctx, sf)
	}

	return nil
}

func (t *DeleteSegmentsByPrefixTask) new() {}
func (t *DeleteSegmentsByPrefixTask) run(ctx context.Context) error {
	listSegments := NewListSegment(t)
	listSegments.SetPath(t.GetPrefix())
	if err := t.Sync(ctx, listSegments); err != nil {
		return err
	}

	it := listSegments.GetSegmentIter()
	for {
		obj, err := it.Next()
		if err != nil {
			if errors.Is(err, typ.IterateDone) {
				break
			}
			return types.NewErrUnhandled(err)
		}

		sf := NewDeleteSegment(t)
		sf.SetSegment(obj)
		if t.ValidateHandleSegmentCallbackFunc() {
			sf.SetCallbackFunc(func() {
				t.GetHandleSegmentCallbackFunc()(obj)
			})
		}
		t.Async(ctx, sf)
	}
	return nil
}

func (t *DeleteStorageTask) new() {}
func (t *DeleteStorageTask) run(ctx context.Context) error {
	var ps []typ.Pair
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

		t.Async(ctx, deletePrefix)

		segmenter, ok := store.(typ.PrefixSegmentsLister)
		if ok {
			deleteSegs := NewDeleteSegmentsByPrefix(t)
			deleteSegs.SetPrefixSegmentsLister(segmenter)
			deleteSegs.SetPrefix("")

			t.Async(ctx, deleteSegs)
		}

		// if meet error, not continue to delete storage
		if err := t.Await(); err != nil {
			return err
		}
	}

	if err := t.GetService().DeleteWithContext(ctx, t.GetStorageName(), ps...); err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

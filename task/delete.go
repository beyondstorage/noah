package task

import (
	"github.com/Xuanwo/storage"
	"github.com/Xuanwo/storage/pkg/segment"
	typ "github.com/Xuanwo/storage/types"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *DeleteFileTask) new() {}
func (t *DeleteFileTask) run() {
	if err := t.GetStorage().Delete(t.GetPath()); err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *DeleteDirTask) new() {}
func (t *DeleteDirTask) run() {
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
		t.GetScheduler().Async(sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewDeleteDir(t)
		sf.SetPath(o.Name)
		if t.ValidateHandleObjCallback() {
			sf.SetHandleObjCallback(t.GetHandleObjCallback())
		}
		t.GetScheduler().Sync(sf)
	})
	t.GetScheduler().Sync(x)

	// after delete all files in this dir, delete dir itself as a file, see issue #43
	dr := NewDeleteFile(t)
	if t.ValidateHandleObjCallback() {
		dr.SetHandleObjCallback(t.GetHandleObjCallback())
	}
	t.GetScheduler().Sync(dr)
}

func (t *DeletePrefixTask) new() {}
func (t *DeletePrefixTask) run() {
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
		t.GetScheduler().Async(sf)
	})

	t.GetScheduler().Sync(x)
}

func (t *DeleteSegmentTask) new() {}
func (t *DeleteSegmentTask) run() {
	if err := t.GetPrefixSegmentsLister().AbortSegment(t.GetSegment()); err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

func (t *DeleteStorageTask) new() {}
func (t *DeleteStorageTask) run() {
	if t.GetForce() {
		store, err := t.GetService().Get(t.GetStorageName())
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

		t.GetScheduler().Async(deletePrefix)

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
				t.GetScheduler().Async(sf)
			})

			t.GetScheduler().Async(listSegments)
		}

		t.GetScheduler().Wait()
	}

	err := t.GetService().Delete(t.GetStorageName())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

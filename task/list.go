package task

import (
	"github.com/Xuanwo/storage"
	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/pkg/progress"
	"github.com/qingstor/noah/pkg/types"
)

func (t *ListDirTask) new() {
	t.SetCallbackFunc(func(types.IDGetter) {
		progress.FinishState(t.GetID())
	})
}
func (t *ListDirTask) run() {
	progress.SetState(t.GetID(), progress.InitListState(t.GetPath(), "listing:"))
	err := t.GetDirLister().ListDir(
		t.GetPath(), pairs.WithDirFunc(t.GetDirFunc()), pairs.WithFileFunc(t.GetFileFunc()))
	if err != nil {
		t.TriggerFault(err)
		return
	}
}

func (t *ListPrefixTask) new() {
	t.SetCallbackFunc(func(types.IDGetter) {
		progress.FinishState(t.GetID())
	})
}

func (t *ListPrefixTask) run() {
	progress.SetState(t.GetID(), progress.InitListState(t.GetPath(), "listing:"))
	err := t.GetPrefixLister().ListPrefix(
		t.GetPath(), pairs.WithObjectFunc(t.GetObjectFunc()))
	if err != nil {
		t.TriggerFault(err)
		return
	}
}

func (t *ListSegmentTask) new() {}
func (t *ListSegmentTask) run() {
	err := t.GetPrefixSegmentsLister().ListPrefixSegments(t.GetPath(),
		pairs.WithSegmentFunc(t.GetSegmentFunc()))
	if err != nil {
		t.TriggerFault(err)
		return
	}
}

func (t *ListStorageTask) new() {}
func (t *ListStorageTask) run() {
	err := t.GetService().List(pairs.WithLocation(t.GetZone()),
		pairs.WithStoragerFunc(func(storager storage.Storager) {
			t.GetStoragerFunc()(storager)
		}))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

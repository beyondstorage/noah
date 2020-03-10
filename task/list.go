package task

import (
	"fmt"

	"github.com/Xuanwo/storage"
	typ "github.com/Xuanwo/storage/types"
	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *ListDirTask) new() {}
func (t *ListDirTask) run() {
	ps := make([]*typ.Pair, 0)
	if t.ValidateDirFunc() {
		ps = append(ps, pairs.WithDirFunc(t.GetDirFunc()))
	}
	if t.ValidateFileFunc() {
		ps = append(ps, pairs.WithFileFunc(t.GetFileFunc()))
	}
	if t.ValidateObjectFunc() {
		ps = append(ps, pairs.WithObjectFunc(t.GetObjectFunc()))
	}
	fmt.Println("listing:", t.GetPath())
	err := t.GetStorage().List(
		t.GetPath(), ps...,
	)
	if err != nil {
		t.TriggerFault(err)
		return
	}
}

func (t *ListSegmentTask) new() {}
func (t *ListSegmentTask) run() {
	err := t.GetSegmenter().ListSegments(t.GetPath(),
		pairs.WithSegmentFunc(t.GetSegmentFunc()))
	if err != nil {
		t.TriggerFault(err)
		return
	}
}

func (t *ListStorageTask) new() {}
func (t *ListStorageTask) run() {
	err := t.GetService().List(pairs.WithLocation(t.GetZone()), pairs.WithStoragerFunc(func(storager storage.Storager) {
		t.GetStoragerFunc()(storager)
	}))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

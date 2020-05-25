package task

import (
	"github.com/Xuanwo/storage"

	"github.com/qingstor/noah/pkg/types"
)

func (t *StatFileTask) new() {}
func (t *StatFileTask) run() {
	om, err := t.GetStorage().Stat(t.GetPath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetObject(om)
}

func (t *StatStorageTask) new() {}
func (t *StatStorageTask) run() {
	s, ok := t.GetStorage().(storage.Statistician)
	if !ok {
		t.TriggerFault(types.NewErrStorageInsufficientAbility(nil))
		return
	}
	info, err := s.Statistical()
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetStorageInfo(info)
}

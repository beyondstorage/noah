package task

import (
	"context"

	"github.com/Xuanwo/storage"

	"github.com/qingstor/noah/pkg/types"
)

func (t *StatFileTask) new() {}
func (t *StatFileTask) run(ctx context.Context) {
	om, err := t.GetStorage().StatWithContext(ctx, t.GetPath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetObject(om)
}

func (t *StatStorageTask) new() {}
func (t *StatStorageTask) run(ctx context.Context) {
	s, ok := t.GetStorage().(storage.Statistician)
	if !ok {
		t.TriggerFault(types.NewErrStorageInsufficientAbility(nil))
		return
	}
	info, err := s.StatisticalWithContext(ctx)
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetStorageInfo(info)
}

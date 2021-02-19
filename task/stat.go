package task

import (
	"context"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/aos-dev/noah/pkg/types"
)

func (t *StatFileTask) new() {}
func (t *StatFileTask) run(ctx context.Context) error {
	om, err := t.GetStorage().StatWithContext(ctx, t.GetPath())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetObject(om)
	return nil
}

func (t *StatStorageTask) new() {}
func (t *StatStorageTask) run(ctx context.Context) error {
	s, ok := t.GetStorage().(typ.Statistician)
	if !ok {
		return types.NewErrStorageInsufficientAbility(nil)
	}
	info, err := s.StatisticalWithContext(ctx)
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetStorageInfo(info)
	return nil
}

package task

import (
	"context"

	"github.com/Xuanwo/storage/types/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *CreateStorageTask) new() {}
func (t *CreateStorageTask) run(ctx context.Context) {
	_, err := t.GetService().CreateWithContext(ctx, t.GetStorageName(), pairs.WithLocation(t.GetZone()))
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
}

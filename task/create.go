package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2/types/pairs"

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

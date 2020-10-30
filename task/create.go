package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2/pairs"

	"github.com/qingstor/noah/pkg/types"
)

func (t *CreateStorageTask) new() {}
func (t *CreateStorageTask) run(ctx context.Context) error {
	_, err := t.GetService().CreateWithContext(ctx, t.GetStorageName(), pairs.WithLocation(t.GetZone()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

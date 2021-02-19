package task

import (
	"context"

	"github.com/aos-dev/go-storage/v2/pairs"

	"github.com/aos-dev/noah/pkg/types"
)

func (t *CreateStorageTask) new() {}
func (t *CreateStorageTask) run(ctx context.Context) error {
	_, err := t.GetService().CreateWithContext(ctx, t.GetStorageName(), pairs.WithLocation(t.GetZone()))
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	return nil
}

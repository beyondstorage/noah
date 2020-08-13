package task

import (
	"context"
	"errors"

	"github.com/aos-dev/go-storage/v2/services"

	"github.com/qingstor/noah/pkg/types"
)

func (t *BetweenStorageCheckTask) new() {}
func (t *BetweenStorageCheckTask) run(ctx context.Context) {
	// Source Object must be exist.
	src, err := t.GetSourceStorage().StatWithContext(ctx, t.GetSourcePath())
	if err != nil {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetSourceObject(src)

	// If Destination Object not exist, we will set DestinationObject to nil.
	// So we can check its existences later.
	dst, err := t.GetDestinationStorage().StatWithContext(ctx, t.GetDestinationPath())
	if err != nil && !errors.Is(err, services.ErrObjectNotExist) {
		t.TriggerFault(types.NewErrUnhandled(err))
		return
	}
	t.SetDestinationObject(dst)
}

func (t *IsDestinationObjectExistTask) new() {}
func (t *IsDestinationObjectExistTask) run(_ context.Context) {
	t.SetResult(t.GetDestinationObject() != nil)
}

func (t *IsDestinationObjectNotExistTask) new() {}
func (t *IsDestinationObjectNotExistTask) run(_ context.Context) {
	t.SetResult(t.GetDestinationObject() == nil)
}

func (t *IsSizeEqualTask) new() {}
func (t *IsSizeEqualTask) run(_ context.Context) {
	t.SetResult(t.GetSourceObject().Size == t.GetDestinationObject().Size)
}

func (t *IsUpdateAtGreaterTask) new() {}
func (t *IsUpdateAtGreaterTask) run(_ context.Context) {
	// if destination object not exist, always consider src is newer
	if t.GetDestinationObject() == nil {
		t.SetResult(true)
	} else {
		t.SetResult(t.GetSourceObject().UpdatedAt.After(t.GetDestinationObject().UpdatedAt))
	}
}

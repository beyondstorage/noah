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

func (t *IsSourcePathExcludeIncludeTask) new() {}
func (t *IsSourcePathExcludeIncludeTask) run(_ context.Context) {
	// source path is rel path based on work-dir, we check exclude and include here:
	// 0. if exclude not set, copy
	// 1. if exclude set but not match, copy
	// 2. if exclude match and include not set, do not copy
	// 3. if exclude match and include set but not match, do not copy
	// 4. if exclude match and include set and match, copy
	// TODO: move exclude and include check into list (in go-storage)
	if t.GetExcludeRegx() == nil {
		t.SetResult(true)
		return
	}

	exMatch := t.GetExcludeRegx().MatchString(t.GetSourcePath())
	if !exMatch {
		t.SetResult(true)
		return
	}

	if t.GetIncludeRegx() == nil {
		t.SetResult(false)
		return
	}

	inMatch := t.GetIncludeRegx().MatchString(t.GetSourcePath())
	t.SetResult(inMatch)
}

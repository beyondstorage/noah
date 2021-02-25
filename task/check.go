// +build refactor

package task

import (
	"context"
	"errors"

	"github.com/aos-dev/go-storage/v2/services"

	"github.com/aos-dev/noah/pkg/types"
)

func (t *BetweenStorageCheckTask) new() {}
func (t *BetweenStorageCheckTask) run(ctx context.Context) error {
	// Source Object must be exist.
	src, err := t.GetSourceStorage().StatWithContext(ctx, t.GetSourcePath())
	if err != nil {
		return types.NewErrUnhandled(err)
	}
	t.SetSourceObject(src)

	// If Destination Object not exist, we will set DestinationObject to nil.
	// So we can check its existences later.
	dst, err := t.GetDestinationStorage().StatWithContext(ctx, t.GetDestinationPath())
	if err != nil && !errors.Is(err, services.ErrObjectNotExist) {
		return types.NewErrUnhandled(err)
	}
	t.SetDestinationObject(dst)

	return nil
}

func (t *IsDestinationObjectExistTask) new() {}
func (t *IsDestinationObjectExistTask) run(_ context.Context) error {
	t.SetResult(t.GetDestinationObject() != nil)
	return nil
}

func (t *IsDestinationObjectNotExistTask) new() {}
func (t *IsDestinationObjectNotExistTask) run(_ context.Context) error {
	t.SetResult(t.GetDestinationObject() == nil)
	return nil
}

func (t *IsSizeEqualTask) new() {}
func (t *IsSizeEqualTask) run(_ context.Context) error {
	srcSize, ok := t.GetSourceObject().GetSize()
	if !ok {
		return types.NewErrObjectMetaInvalid(nil, "size", t.GetSourceObject())
	}

	dstSize, ok := t.GetDestinationObject().GetSize()
	if !ok {
		return types.NewErrObjectMetaInvalid(nil, "size", t.GetDestinationObject())
	}
	t.SetResult(srcSize == dstSize)
	return nil
}

func (t *IsUpdateAtGreaterTask) new() {}
func (t *IsUpdateAtGreaterTask) run(_ context.Context) error {
	// if destination object not exist, always consider src is newer
	if t.GetDestinationObject() == nil {
		t.SetResult(true)
		return nil
	}

	srcUpdate, ok := t.GetSourceObject().GetUpdatedAt()
	if !ok {
		return types.NewErrObjectMetaInvalid(nil, "update_at", t.GetSourceObject())
	}

	dstUpdate, ok := t.GetDestinationObject().GetUpdatedAt()
	if !ok {
		return types.NewErrObjectMetaInvalid(nil, "update_at", t.GetDestinationObject())
	}
	t.SetResult(srcUpdate.After(dstUpdate))
	return nil
}

func (t *IsSourcePathExcludeIncludeTask) new() {}
func (t *IsSourcePathExcludeIncludeTask) run(_ context.Context) error {
	// source path is rel path based on work-dir, we check exclude and include here:
	// 0. if exclude not set, copy
	// 1. if exclude set but not match, copy
	// 2. if exclude match and include not set, do not copy
	// 3. if exclude match and include set but not match, do not copy
	// 4. if exclude match and include set and match, copy
	// TODO: move exclude and include check into list (in go-storage)
	if t.GetExcludeRegexp() == nil {
		t.SetResult(true)
		return nil
	}

	exMatch := t.GetExcludeRegexp().MatchString(t.GetSourcePath())
	if !exMatch {
		t.SetResult(true)
		return nil
	}

	if t.GetIncludeRegexp() == nil {
		t.SetResult(false)
		return nil
	}

	inMatch := t.GetIncludeRegexp().MatchString(t.GetSourcePath())
	t.SetResult(inMatch)
	return nil
}

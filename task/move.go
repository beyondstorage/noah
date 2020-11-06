package task

import (
	"context"
	"errors"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *MoveDirTask) new() {}
func (t *MoveDirTask) run(ctx context.Context) error {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		return err
	}

	if err := t.Sync(ctx, x); err != nil {
		return err
	}

	it := x.GetObjectIter()
	for {
		obj, err := it.Next()
		if err != nil {
			if errors.Is(err, typ.IterateDone) {
				break
			}
			return types.NewErrUnhandled(err)
		}
		switch obj.Type {
		case typ.ObjectTypeFile:
			sf := NewMoveFile(t)
			sf.SetSourcePath(obj.Name)
			sf.SetDestinationPath(obj.Name)
			if t.ValidateHandleObjCallbackFunc() {
				sf.SetCallbackFunc(func() {
					t.GetHandleObjCallbackFunc()(obj)
				})
			}
			t.Async(ctx, sf)
		case typ.ObjectTypeDir:
			sf := NewMoveDir(t)
			sf.SetSourcePath(obj.Name)
			sf.SetDestinationPath(obj.Name)
			if err := t.Sync(ctx, sf); err != nil {
				return err
			}
		default:
			return types.NewErrObjectTypeInvalid(nil, obj)
		}
	}

	return nil
}

func (t *MoveFileTask) new() {}
func (t *MoveFileTask) run(ctx context.Context) error {
	ct := NewCopyFile(t)
	if err := t.Sync(ctx, ct); err != nil {
		return err
	}

	dt := NewDeleteFile(t)
	utils.ChooseSourceStorage(dt, t)
	return t.Sync(ctx, dt)
}

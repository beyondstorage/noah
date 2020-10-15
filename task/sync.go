package task

import (
	"context"
	"errors"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *SyncTask) new() {}
func (t *SyncTask) run(ctx context.Context) error {
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
			sf := NewCopyFile(t)
			sf.SetSourcePath(obj.Name)
			sf.SetDestinationPath(obj.Name)
			if t.ValidateHandleObjCallbackFunc() {
				sf.SetCallbackFunc(func() {
					t.GetHandleObjCallbackFunc()(obj)
				})
			}
			t.Async(ctx, sf)
		case typ.ObjectTypeDir:
			if t.ValidateRecursive() && t.GetRecursive() {
				sf := NewSync(t)
				sf.SetSourcePath(obj.Name)
				sf.SetDestinationPath(obj.Name)
				t.Async(ctx, sf)
			}
		default:
			return types.NewErrObjectTypeInvalid(nil, obj)
		}
	}

	return nil
}

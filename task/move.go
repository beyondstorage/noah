package task

import (
	"context"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/utils"
)

func (t *MoveDirTask) new() {}
func (t *MoveDirTask) run(ctx context.Context) error {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		return err
	}

	x.SetFileFunc(func(o *typ.Object) {
		sf := NewMoveFile(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		if t.ValidateHandleObjCallbackFunc() {
			sf.SetCallbackFunc(func() {
				t.GetHandleObjCallbackFunc()(o)
			})
		}
		t.Async(ctx, sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewMoveDir(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		t.Sync(ctx, sf)
	})
	return t.Sync(ctx, x)
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

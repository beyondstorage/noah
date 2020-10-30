package task

import (
	"context"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/utils"
)

func (t *MoveDirTask) new() {}
func (t *MoveDirTask) run(ctx context.Context) {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		t.TriggerFault(err)
		return
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
		if t.ValidatePartSize() {
			sf.SetPartSize(t.GetPartSize())
		}
		t.GetScheduler().Async(ctx, sf)
	})
	x.SetDirFunc(func(o *typ.Object) {
		sf := NewMoveDir(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		if t.ValidateHandleObjCallbackFunc() {
			sf.SetHandleObjCallbackFunc(t.GetHandleObjCallbackFunc())
		}
		if t.ValidatePartSize() {
			sf.SetPartSize(t.GetPartSize())
		}
		t.GetScheduler().Sync(ctx, sf)
	})
	t.GetScheduler().Sync(ctx, x)
}

func (t *MoveFileTask) new() {}
func (t *MoveFileTask) run(ctx context.Context) {
	ct := NewCopyFile(t)
	if t.ValidatePartSize() {
		ct.SetPartSize(t.GetPartSize())
	}
	t.GetScheduler().Sync(ctx, ct)
	if t.GetFault().HasError() {
		return
	}

	dt := NewDeleteFile(t)
	utils.ChooseSourceStorage(dt, t)
	t.GetScheduler().Sync(ctx, dt)
}

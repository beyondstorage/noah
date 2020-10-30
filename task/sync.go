package task

import (
	"context"

	typ "github.com/aos-dev/go-storage/v2/types"

	"github.com/qingstor/noah/pkg/types"
	"github.com/qingstor/noah/utils"
)

func (t *SyncTask) new() {}
func (t *SyncTask) run(ctx context.Context) {
	x := NewListDir(t)
	err := utils.ChooseSourceStorageAsDirLister(x, t)
	if err != nil {
		t.TriggerFault(err)
		return
	}

	if t.GetRecursive() {
		x.SetDirFunc(func(o *typ.Object) {
			sf := NewSync(t)
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
	} else {
		// if not recursive, do nothing with dir
		x.SetDirFunc(func(_ *typ.Object) {})
	}

	x.SetFileFunc(func(o *typ.Object) {
		sf := NewCopyFile(t)
		sf.SetSourcePath(o.Name)
		sf.SetDestinationPath(o.Name)
		// set check task to nil, to skip check in copy file, because we check here, below
		sf.SetCheckTasks(nil)

		// put check tasks outside of copy, to make sure flags' priority is higher than dry-run
		check := NewBetweenStorageCheck(sf)
		sf.GetScheduler().Sync(ctx, check)
		if sf.GetFault().HasError() {
			return
		}
		for _, v := range t.GetCheckTasks() {
			ct := v(check)
			sf.GetScheduler().Sync(ctx, ct)
			if sf.GetFault().HasError() {
				return
			}
			// If any of checks not pass, do not copy this file.
			if result := ct.(types.ResultGetter); !result.GetResult() {
				return
			}
			// If all check passed, we should continue do copy works.
		}

		// if dry-run, only check, and call dry-run func if check passed
		if t.GetDryRunFunc() != nil {
			t.GetDryRunFunc()(o)
			return
		}

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
	t.GetScheduler().Sync(ctx, x)
}
